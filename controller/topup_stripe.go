package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	stripecoupon "github.com/stripe/stripe-go/v81/coupon"
	stripeprice "github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/thanhpk/randstr"
)

const (
	PaymentMethodStripe = "stripe"
)

var stripeAdaptor = &StripeAdaptor{}

type StripePayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	CouponId      int    `json:"coupon_id"`
}

type StripeAdaptor struct {
}

type stripeUserCouponContext struct {
	UserID         int
	Username       string
	UserCouponID   int
	DiscountAmount decimal.Decimal
}

func (*StripeAdaptor) RequestAmount(c *gin.Context, req *StripePayRequest) {
	if req.Amount < getStripeMinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getStripeMinTopup())})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	originalPayMoney := getStripeOriginalPayMoney(req.Amount, group)
	minThreshold := decimal.NewFromFloat(getStripePayMoney(getStripeMinTopup(), group))
	hasUserCouponPriority, err := hasEligibleTopUpCoupon(id, originalPayMoney, minThreshold)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取优惠券信息失败"})
		return
	}
	payMoney := applyTopupDiscount(originalPayMoney, req.Amount).InexactFloat64()
	if hasUserCouponPriority {
		payMoney = originalPayMoney.InexactFloat64()
	}
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

func (*StripeAdaptor) RequestPay(c *gin.Context, req *StripePayRequest) {
	if req.PaymentMethod != PaymentMethodStripe {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}
	if req.Amount < getStripeMinTopup() {
		c.JSON(200, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", getStripeMinTopup()), "data": 10})
		return
	}
	if req.Amount > 10000 {
		c.JSON(200, gin.H{"message": "充值数量不能大于 10000", "data": 10})
		return
	}

	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)
	if user == nil {
		c.JSON(200, gin.H{"message": "error", "data": "用户不存在"})
		return
	}
	quote, err := buildTopUpQuote(user, TopUpQuoteRequest{
		PaymentMethod: PaymentMethodStripe,
		Amount:        req.Amount,
		CouponId:      req.CouponId,
	})
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}
	if err := validateSelectedCoupon(req.CouponId, quote); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}
	chargedMoney := GetChargedAmount(float64(req.Amount), *user)
	originalPayMoney := getStripeOriginalPayMoney(req.Amount, user.Group)
	payMoney := decimal.NewFromFloat(quote.FinalPayableAmount)
	if payMoney.InexactFloat64() <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	discountRule, _ := getTopupDiscountRule(req.Amount)
	if quote.PlatformDiscountAmount <= 0 {
		discountRule = operation_setting.AmountDiscountRule{}
	}
	var userCoupon *stripeUserCouponContext
	if req.CouponId != 0 {
		userCoupon = &stripeUserCouponContext{
			UserID:         user.Id,
			Username:       user.Username,
			UserCouponID:   req.CouponId,
			DiscountAmount: decimal.NewFromFloat(quote.CouponDiscountAmount),
		}
	}

	reference := fmt.Sprintf("new-api-ref-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(4))
	referenceId := "ref_" + common.Sha1([]byte(reference))

	payLink, stripeCouponId, err := genStripeLink(
		referenceId,
		user.StripeCustomer,
		user.Email,
		req.Amount,
		originalPayMoney,
		decimal.NewFromFloat(quote.BasePayableAmount),
		payMoney,
		discountRule,
		userCoupon,
	)
	if err != nil {
		log.Println("获取Stripe Checkout支付链接失败", err)
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:           id,
		Amount:           req.Amount,
		Money:            chargedMoney,
		TradeNo:          referenceId,
		PaymentMethod:    PaymentMethodStripe,
		CreateTime:       time.Now().Unix(),
		Status:           common.TopUpStatusPending,
		CouponId:         req.CouponId,
		OriginalMoney:    quote.OriginalAmount,
		PlatformDiscount: quote.PlatformDiscountAmount,
		CouponDiscount:   quote.CouponDiscountAmount,
		PayMoney:         quote.FinalPayableAmount,
		StripeCouponId:   stripeCouponId,
	}
	err = createTopUpOrder(topUp)
	if err != nil {
		if cleanupErr := cleanupStripeCouponByID(stripeCouponId); cleanupErr != nil {
			log.Printf("清理 Stripe 临时优惠券失败: %v", cleanupErr)
		}
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}
	if req.CouponId != 0 {
		model.RecordLog(id, model.LogTypeTopup, fmt.Sprintf("创建 Stripe 充值订单并使用优惠券 %s，抵扣 %.2f", topUp.CouponName, topUp.CouponDiscount))
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_link": payLink,
		},
	})
}

func RequestStripeAmount(c *gin.Context) {
	var req StripePayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	stripeAdaptor.RequestAmount(c, &req)
}

func RequestStripePay(c *gin.Context) {
	var req StripePayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	stripeAdaptor.RequestPay(c, &req)
}

func StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("解析Stripe Webhook参数失败: %v\n", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	endpointSecret := setting.StripeWebhookSecret
	event, err := webhook.ConstructEventWithOptions(payload, signature, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})

	if err != nil {
		log.Printf("Stripe Webhook验签失败: %v\n", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		sessionCompleted(event)
	case stripe.EventTypeCheckoutSessionAsyncPaymentFailed:
		sessionAsyncPaymentFailed(event)
	case stripe.EventTypeCheckoutSessionExpired:
		sessionExpired(event)
	default:
		log.Printf("不支持的Stripe Webhook事件类型: %s\n", event.Type)
	}

	c.Status(http.StatusOK)
}

func sessionCompleted(event stripe.Event) {
	customerId := event.GetObjectValue("customer")
	referenceId := event.GetObjectValue("client_reference_id")
	status := event.GetObjectValue("status")
	if "complete" != status {
		log.Println("错误的Stripe Checkout完成状态:", status, ",", referenceId)
		return
	}

	err := model.Recharge(referenceId, customerId)
	if err != nil {
		log.Println(err.Error(), referenceId)
		return
	}
	if err := cleanupStripeCheckoutCouponByTradeNo(referenceId); err != nil {
		log.Println("清理 Stripe 临时优惠券失败", referenceId, ", err:", err.Error())
	}

	total, _ := strconv.ParseFloat(event.GetObjectValue("amount_total"), 64)
	currency := strings.ToUpper(event.GetObjectValue("currency"))
	log.Printf("收到款项：%s, %.2f(%s)", referenceId, total/100, currency)
}

func sessionAsyncPaymentFailed(event stripe.Event) {
	referenceId := event.GetObjectValue("client_reference_id")
	if len(referenceId) == 0 {
		log.Println("未提供支付单号")
		return
	}

	if err := expireTopUpOrderByTradeNo(referenceId); err != nil {
		log.Println("异步支付失败处理订单失败", referenceId, ", err:", err.Error())
	}
	if err := cleanupStripeCheckoutCouponByTradeNo(referenceId); err != nil {
		log.Println("清理 Stripe 临时优惠券失败", referenceId, ", err:", err.Error())
	}
}

func sessionExpired(event stripe.Event) {
	referenceId := event.GetObjectValue("client_reference_id")
	status := event.GetObjectValue("status")
	if "expired" != status {
		log.Println("错误的Stripe Checkout过期状态:", status, ",", referenceId)
		return
	}

	if len(referenceId) == 0 {
		log.Println("未提供支付单号")
		return
	}

	topUp := model.GetTopUpByTradeNo(referenceId)
	if topUp == nil {
		log.Println("充值订单不存在", referenceId)
		return
	}

	if topUp.Status != common.TopUpStatusPending {
		log.Println("充值订单状态错误", referenceId)
	}

	err := expireTopUpOrderByTradeNo(referenceId)
	if err != nil {
		log.Println("过期充值订单失败", referenceId, ", err:", err.Error())
		return
	}
	if err := cleanupStripeCheckoutCouponByTradeNo(referenceId); err != nil {
		log.Println("清理 Stripe 临时优惠券失败", referenceId, ", err:", err.Error())
	}

	log.Println("充值订单已过期", referenceId)
}

func genStripeLink(
	referenceId string,
	customerId string,
	email string,
	amount int64,
	originalPayMoney decimal.Decimal,
	basePayableMoney decimal.Decimal,
	finalPayMoney decimal.Decimal,
	discountRule operation_setting.AmountDiscountRule,
	userCoupon *stripeUserCouponContext,
) (string, string, error) {
	if !strings.HasPrefix(setting.StripeApiSecret, "sk_") && !strings.HasPrefix(setting.StripeApiSecret, "rk_") {
		return "", "", fmt.Errorf("无效的Stripe API密钥")
	}

	stripe.Key = setting.StripeApiSecret

	priceInfo, err := stripeprice.Get(setting.StripePriceId, nil)
	if err != nil {
		return "", "", fmt.Errorf("获取Stripe价格配置失败: %w", err)
	}

	lineAmount, usePresetCoupon := resolveStripeCheckoutAmount(
		originalPayMoney,
		basePayableMoney,
		finalPayMoney,
		discountRule,
		userCoupon != nil,
	)
	appliedCouponID := ""
	temporaryCouponID := ""
	if userCoupon != nil {
		createdCoupon, createErr := createStripeUserDiscountCoupon(priceInfo, userCoupon)
		if createErr != nil {
			return "", "", createErr
		}
		appliedCouponID = createdCoupon.ID
		temporaryCouponID = createdCoupon.ID
	} else if usePresetCoupon {
		appliedCouponID = discountRule.CouponID
	}

	minorAmount := getStripeMinorUnitAmount(lineAmount, setting.StripeUnitPrice, priceInfo.UnitAmount)
	if minorAmount <= 0 {
		if cleanupErr := cleanupStripeCouponByID(temporaryCouponID); cleanupErr != nil {
			log.Printf("清理 Stripe 临时优惠券失败: %v", cleanupErr)
		}
		return "", "", fmt.Errorf("无效的Stripe支付金额")
	}

	lineItem := &stripe.CheckoutSessionLineItemParams{
		Quantity: stripe.Int64(1),
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency:   stripe.String(string(priceInfo.Currency)),
			UnitAmount: stripe.Int64(minorAmount),
		},
	}

	if priceInfo.Product != nil && priceInfo.Product.ID != "" {
		lineItem.PriceData.Product = stripe.String(priceInfo.Product.ID)
	} else {
		lineItem.PriceData.ProductData = &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name: stripe.String(fmt.Sprintf("Top up %d", amount)),
		}
	}

	if priceInfo.TaxBehavior != "" && priceInfo.TaxBehavior != stripe.PriceTaxBehaviorUnspecified {
		lineItem.PriceData.TaxBehavior = stripe.String(string(priceInfo.TaxBehavior))
	}

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(referenceId),
		SuccessURL:        stripe.String(system_setting.ServerAddress + "/console/log"),
		CancelURL:         stripe.String(system_setting.ServerAddress + "/console/topup"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			lineItem,
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}
	if appliedCouponID != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{
				Coupon: stripe.String(appliedCouponID),
			},
		}
	} else {
		params.AllowPromotionCodes = stripe.Bool(setting.StripePromotionCodesEnabled)
	}

	if "" == customerId {
		if "" != email {
			params.CustomerEmail = stripe.String(email)
		}

		params.CustomerCreation = stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways))
	} else {
		params.Customer = stripe.String(customerId)
	}

	result, err := session.New(params)
	if err != nil {
		if cleanupErr := cleanupStripeCouponByID(temporaryCouponID); cleanupErr != nil {
			log.Printf("清理 Stripe 临时优惠券失败: %v", cleanupErr)
		}
		return "", "", err
	}

	return result.URL, temporaryCouponID, nil
}

func resolveStripeCheckoutAmount(
	originalPayMoney decimal.Decimal,
	basePayableMoney decimal.Decimal,
	finalPayMoney decimal.Decimal,
	discountRule operation_setting.AmountDiscountRule,
	hasUserCoupon bool,
) (decimal.Decimal, bool) {
	usePresetCoupon := discountRule.CouponID != "" && discountRule.DiscountAmount > 0

	switch {
	case hasUserCoupon:
		// Checkout supports only one coupon, so platform discounts stay baked
		// into the line item and the user coupon is represented as the Stripe coupon.
		return basePayableMoney, false
	case usePresetCoupon:
		// Platform discount is delegated to Stripe via the preset coupon, so
		// Stripe must receive the original price to avoid double-discounting.
		return originalPayMoney, true
	default:
		// Without a Stripe-side coupon, the line item itself carries the
		// discounted payable amount.
		return finalPayMoney, false
	}
}

func createStripeUserDiscountCoupon(priceInfo *stripe.Price, userCoupon *stripeUserCouponContext) (*stripe.Coupon, error) {
	if priceInfo == nil || userCoupon == nil {
		return nil, fmt.Errorf("缺少 Stripe 用户优惠券上下文")
	}

	amountOff := getStripeMinorUnitAmount(userCoupon.DiscountAmount, setting.StripeUnitPrice, priceInfo.UnitAmount)
	if amountOff <= 0 {
		return nil, fmt.Errorf("无效的 Stripe 用户优惠券金额")
	}

	params := &stripe.CouponParams{
		AmountOff:      stripe.Int64(amountOff),
		Currency:       stripe.String(string(priceInfo.Currency)),
		Duration:       stripe.String(string(stripe.CouponDurationOnce)),
		MaxRedemptions: stripe.Int64(1),
		Name:           stripe.String("User Discount"),
	}
	params.AddMetadata("UserID", strconv.Itoa(userCoupon.UserID))
	params.AddMetadata("Username", userCoupon.Username)
	params.AddMetadata("UserCouponId", strconv.Itoa(userCoupon.UserCouponID))

	return stripecoupon.New(params)
}

func cleanupStripeCheckoutCouponByTradeNo(tradeNo string) error {
	if tradeNo == "" {
		return nil
	}

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil || topUp.StripeCouponId == "" {
		return nil
	}

	if err := cleanupStripeCouponByID(topUp.StripeCouponId); err != nil {
		return err
	}

	topUp.StripeCouponId = ""
	return model.DB.Model(&model.TopUp{}).Where("id = ?", topUp.Id).Update("stripe_coupon_id", "").Error
}

func cleanupStripeCouponByID(couponID string) error {
	if couponID == "" {
		return nil
	}

	stripe.Key = setting.StripeApiSecret
	_, err := stripecoupon.Del(couponID, nil)
	if err == nil {
		return nil
	}

	var stripeErr *stripe.Error
	if errors.As(err, &stripeErr) && stripeErr.Code == stripe.ErrorCodeResourceMissing {
		return nil
	}
	return err
}

func GetChargedAmount(count float64, user model.User) float64 {
	topUpGroupRatio := common.GetTopupGroupRatio(user.Group)
	if topUpGroupRatio == 0 {
		topUpGroupRatio = 1
	}

	return count * topUpGroupRatio
}

func getStripePayMoney(amount int64, group string) float64 {
	payMoney := getStripeOriginalPayMoney(amount, group)
	payMoney = applyTopupDiscount(payMoney, amount)
	return payMoney.InexactFloat64()
}

func getStripeOriginalPayMoney(amount int64, group string) decimal.Decimal {
	dAmount := decimal.NewFromInt(amount)
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount = dAmount.Div(decimal.NewFromFloat(common.QuotaPerUnit))
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dUnitPrice := decimal.NewFromFloat(setting.StripeUnitPrice)
	return dAmount.Mul(dUnitPrice).Mul(dTopupGroupRatio)
}

func getStripeMinTopup() int64 {
	minTopup := setting.StripeMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}
