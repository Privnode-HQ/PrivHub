package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TopUpCouponRequest struct {
	Id              int     `json:"id"`
	Name            string  `json:"name"`
	BoundUserId     int     `json:"bound_user_id"`
	DeductionAmount float64 `json:"deduction_amount"`
	CurrencyCode    string  `json:"currency_code"`
	ValidFrom       int64   `json:"valid_from"`
	ExpiresAt       int64   `json:"expires_at"`
	Action          string  `json:"action"`
	RevokeReason    string  `json:"revoke_reason"`
}

type TopUpQuoteRequest struct {
	PaymentMethod string `json:"payment_method"`
	Amount        int64  `json:"amount"`
	ProductId     string `json:"product_id"`
	CurrencyCode  string `json:"currency_code"`
	CouponId      int    `json:"coupon_id"`
}

type TopUpQuoteCoupon struct {
	Id              int     `json:"id"`
	Name            string  `json:"name"`
	DeductionAmount float64 `json:"deduction_amount"`
	CurrencyCode    string  `json:"currency_code,omitempty"`
	Status          string  `json:"status"`
	ExpiresAt       int64   `json:"expires_at"`
}

type TopUpQuoteData struct {
	PaymentMethod          string             `json:"payment_method"`
	Amount                 int64              `json:"amount,omitempty"`
	ProductId              string             `json:"product_id,omitempty"`
	ProductName            string             `json:"product_name,omitempty"`
	ProductQuota           int64              `json:"product_quota,omitempty"`
	CurrencyCode           string             `json:"currency_code,omitempty"`
	SupportedCurrencyCodes []string           `json:"supported_currency_codes,omitempty"`
	OriginalAmount         float64            `json:"original_amount"`
	BasePayableAmount      float64            `json:"base_payable_amount"`
	PlatformDiscountAmount float64            `json:"platform_discount_amount"`
	CouponDiscountAmount   float64            `json:"coupon_discount_amount"`
	FinalPayableAmount     float64            `json:"final_payable_amount"`
	MinPayableThreshold    float64            `json:"min_payable_threshold"`
	SelectedCouponId       int                `json:"selected_coupon_id,omitempty"`
	IneligibleReason       string             `json:"ineligible_reason,omitempty"`
	AvailableCoupons       []TopUpQuoteCoupon `json:"available_coupons"`
}

func GetAllTopUpCoupons(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	userId, _ := strconv.Atoi(c.Query("user_id"))
	coupons, total, err := model.GetAllTopUpCoupons(pageInfo, model.TopUpCouponFilter{
		Keyword: c.Query("keyword"),
		Status:  c.Query("status"),
		UserId:  userId,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(coupons)
	common.ApiSuccess(c, pageInfo)
}

func SearchTopUpCoupons(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	coupons, total, err := model.SearchTopUpCoupons(c.Query("keyword"), pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(coupons)
	common.ApiSuccess(c, pageInfo)
}

func GetTopUpCoupon(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	coupon, err := model.GetTopUpCouponById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, coupon)
}

func AddTopUpCoupon(c *gin.Context) {
	req := TopUpCouponRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if utf8.RuneCountInString(strings.TrimSpace(req.Name)) == 0 || utf8.RuneCountInString(strings.TrimSpace(req.Name)) > 50 {
		common.ApiErrorMsg(c, "优惠券名称长度必须在 1-50 之间")
		return
	}
	if _, err := model.GetUserById(req.BoundUserId, false); err != nil {
		common.ApiErrorMsg(c, "目标用户不存在")
		return
	}
	currencyCode := model.NormalizeTopUpCouponCurrencyCode(req.CurrencyCode)
	if currencyCode == "" {
		common.ApiErrorMsg(c, "请选择优惠货币")
		return
	}

	now := common.GetTimestamp()
	coupon := &model.TopUpCoupon{
		Name:            req.Name,
		BoundUserId:     req.BoundUserId,
		DeductionAmount: req.DeductionAmount,
		CurrencyCode:    currencyCode,
		ValidFrom:       req.ValidFrom,
		ExpiresAt:       req.ExpiresAt,
		IssuedByAdminId: c.GetInt("id"),
		IssuedAt:        now,
		CreatedTime:     now,
		UpdatedTime:     now,
		Status:          common.TopUpCouponStatusAvailable,
	}
	if coupon.ValidFrom == 0 {
		coupon.ValidFrom = now
	}
	if err := coupon.Validate(); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := coupon.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员向用户 #%d 发放优惠券 %s，优惠金额 %.2f %s", coupon.BoundUserId, coupon.Name, coupon.DeductionAmount, coupon.GetDisplayCurrencyCode(model.DefaultTopUpCouponCurrencyCode())))
	result, err := model.GetTopUpCouponById(coupon.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func UpdateTopUpCoupon(c *gin.Context) {
	req := TopUpCouponRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	action := strings.TrimSpace(req.Action)
	if action == "revoke" {
		coupon, err := model.RevokeTopUpCoupon(req.Id, c.GetInt("id"), req.RevokeReason)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员撤销优惠券 #%d %s", coupon.Id, coupon.Name))
		result, err := model.GetTopUpCouponById(coupon.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		common.ApiSuccess(c, result)
		return
	}

	coupon, err := model.RefreshTopUpCouponState(req.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if coupon.GetEffectiveStatus() == common.TopUpCouponStatusUsed {
		common.ApiErrorMsg(c, "已使用的优惠券不能编辑")
		return
	}
	if coupon.GetEffectiveStatus() == common.TopUpCouponStatusReserved {
		common.ApiErrorMsg(c, "支付中的优惠券不能编辑")
		return
	}

	if req.Name != "" {
		coupon.Name = req.Name
	}
	if req.BoundUserId > 0 && req.BoundUserId != coupon.BoundUserId {
		if _, err := model.GetUserById(req.BoundUserId, false); err != nil {
			common.ApiErrorMsg(c, "目标用户不存在")
			return
		}
		coupon.BoundUserId = req.BoundUserId
	}
	if req.DeductionAmount > 0 {
		coupon.DeductionAmount = req.DeductionAmount
	}
	if normalizedCurrencyCode := model.NormalizeTopUpCouponCurrencyCode(req.CurrencyCode); normalizedCurrencyCode != "" {
		coupon.CurrencyCode = normalizedCurrencyCode
	}
	if req.ValidFrom > 0 {
		coupon.ValidFrom = req.ValidFrom
	}
	coupon.ExpiresAt = req.ExpiresAt
	if coupon.Status == common.TopUpCouponStatusExpired && (coupon.ExpiresAt == 0 || coupon.ExpiresAt >= common.GetTimestamp()) {
		coupon.Status = common.TopUpCouponStatusAvailable
	}
	if err := coupon.Validate(); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := coupon.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员更新优惠券 #%d %s", coupon.Id, coupon.Name))
	result, err := model.GetTopUpCouponById(coupon.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func RequestTopUpQuote(c *gin.Context) {
	req := TopUpQuoteRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	user, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil || user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}

	quote, err := buildTopUpQuote(user, req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, quote)
}

func buildTopUpQuote(user *model.User, req TopUpQuoteRequest) (*TopUpQuoteData, error) {
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	if strings.TrimSpace(req.PaymentMethod) == "" {
		return nil, errors.New("请选择支付方式")
	}

	quote := &TopUpQuoteData{
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		ProductId:     req.ProductId,
	}

	var (
		originalAmount           decimal.Decimal
		discountedBasePayable    decimal.Decimal
		basePayable              decimal.Decimal
		platformDiscount         decimal.Decimal
		resolvedPlatformDiscount decimal.Decimal
		minThreshold             decimal.Decimal
		allCoupons               []*model.TopUpCoupon
		err                      error
	)

	switch req.PaymentMethod {
	case PaymentMethodStripe:
		if req.Amount < getStripeMinTopup() {
			return nil, fmt.Errorf("充值数量不能小于 %d", getStripeMinTopup())
		}
		stripePriceInfo, priceErr := getStripePriceInfo(req.CurrencyCode)
		if priceErr != nil {
			return nil, priceErr
		}
		originalAmount, _, err = getStripeOriginalPayMoneyWithPrice(stripePriceInfo, req.Amount)
		if err != nil {
			return nil, err
		}
		quote.CurrencyCode = model.NormalizeTopUpCouponCurrencyCode(string(stripePriceInfo.Currency))
		quote.SupportedCurrencyCodes = stripePriceInfo.SupportedCurrencyCodes
		platformDiscount = getTopupDiscountAmount(req.Amount)
		discountedBasePayable = applyTopupDiscount(originalAmount, req.Amount)
		stripeMinPayMoney, minErr := getStripePayMoneyWithPrice(stripePriceInfo, getStripeMinTopup())
		if minErr != nil {
			return nil, minErr
		}
		minThreshold = decimal.NewFromFloat(stripeMinPayMoney)
		allCoupons, err = model.GetUserAvailableTopUpCoupons(user.Id)
	case PaymentMethodCreem:
		product, productErr := findCreemProductById(req.ProductId)
		if productErr != nil {
			return nil, productErr
		}
		quote.ProductName = product.Name
		quote.ProductQuota = product.Quota
		quote.CurrencyCode = model.NormalizeTopUpCouponCurrencyCode(product.Currency)
		originalAmount = decimal.NewFromFloat(product.Price)
		discountedBasePayable = originalAmount
		minThreshold = decimal.NewFromFloat(getCreemMinimumPayable())
		if req.CouponId != 0 {
			quote.IneligibleReason = "当前支付方式暂不支持优惠券"
		}
		allCoupons = []*model.TopUpCoupon{}
	default:
		if !operation_setting.ContainsPayMethod(req.PaymentMethod) {
			return nil, errors.New("支付方式不存在")
		}
		if req.Amount < getMinTopup() {
			return nil, fmt.Errorf("充值数量不能小于 %d", getMinTopup())
		}
		originalAmount = getOriginalPayMoney(req.Amount, user.Group)
		platformDiscount = getTopupDiscountAmount(req.Amount)
		discountedBasePayable = applyTopupDiscount(originalAmount, req.Amount)
		minThreshold = decimal.NewFromFloat(getPayMoney(getMinTopup(), user.Group))
		quote.CurrencyCode = model.DefaultTopUpCouponCurrencyCode()
		allCoupons, err = model.GetUserAvailableTopUpCoupons(user.Id)
	}
	if err != nil {
		return nil, err
	}

	eligibleCoupons, selectedCoupon, ineligibleReason := buildEligibleTopUpQuoteCoupons(allCoupons, originalAmount, minThreshold, req.CouponId, quote.CurrencyCode)
	basePayable, resolvedPlatformDiscount = resolveTopUpBasePayable(
		originalAmount,
		discountedBasePayable,
		platformDiscount,
		selectedCoupon != nil,
	)

	quote.OriginalAmount = roundMoney(originalAmount)
	quote.BasePayableAmount = roundMoney(basePayable)
	quote.PlatformDiscountAmount = roundMoney(resolvedPlatformDiscount)
	quote.MinPayableThreshold = roundMoney(minThreshold)
	if ineligibleReason != "" {
		quote.IneligibleReason = ineligibleReason
	}

	if req.CouponId != 0 && selectedCoupon == nil && quote.IneligibleReason == "" {
		quote.IneligibleReason = "优惠券不可用"
	}

	quote.AvailableCoupons = eligibleCoupons
	if selectedCoupon != nil {
		quote.SelectedCouponId = selectedCoupon.Id
		quote.CouponDiscountAmount = roundMoney(decimal.NewFromFloat(selectedCoupon.DeductionAmount))
		quote.FinalPayableAmount = roundMoney(basePayable.Sub(decimal.NewFromFloat(selectedCoupon.DeductionAmount)))
	} else {
		quote.FinalPayableAmount = roundMoney(basePayable)
	}

	if quote.FinalPayableAmount <= 0 {
		return nil, errors.New("充值金额过低")
	}
	return quote, nil
}

func buildEligibleTopUpQuoteCoupons(allCoupons []*model.TopUpCoupon, basePayable, minThreshold decimal.Decimal, requestedCouponId int, quoteCurrencyCode string) ([]TopUpQuoteCoupon, *model.TopUpCoupon, string) {
	eligibleCoupons := make([]TopUpQuoteCoupon, 0, len(allCoupons))
	selectedCoupon := (*model.TopUpCoupon)(nil)
	ineligibleReason := ""
	normalizedQuoteCurrencyCode := model.NormalizeTopUpCouponCurrencyCode(quoteCurrencyCode)

	for _, coupon := range allCoupons {
		if !coupon.IsCurrencyCompatible(normalizedQuoteCurrencyCode) {
			if requestedCouponId == coupon.Id {
				ineligibleReason = "优惠券币种与当前支付方式不匹配"
			}
			continue
		}

		finalPayable := basePayable.Sub(decimal.NewFromFloat(coupon.DeductionAmount))
		if finalPayable.LessThanOrEqual(minThreshold) {
			if requestedCouponId == coupon.Id {
				ineligibleReason = "使用优惠券后金额必须高于最低充值金额"
			}
			continue
		}

		eligibleCoupons = append(eligibleCoupons, TopUpQuoteCoupon{
			Id:              coupon.Id,
			Name:            coupon.Name,
			DeductionAmount: roundMoney(decimal.NewFromFloat(coupon.DeductionAmount)),
			CurrencyCode:    coupon.GetDisplayCurrencyCode(normalizedQuoteCurrencyCode),
			Status:          coupon.GetEffectiveStatus(),
			ExpiresAt:       coupon.ExpiresAt,
		})
		if requestedCouponId == coupon.Id {
			selectedCoupon = coupon
		}
	}

	return eligibleCoupons, selectedCoupon, ineligibleReason
}

func hasEligibleTopUpCoupon(userId int, basePayable, minThreshold decimal.Decimal, quoteCurrencyCode string) (bool, error) {
	allCoupons, err := model.GetUserAvailableTopUpCoupons(userId)
	if err != nil {
		return false, err
	}
	eligibleCoupons, _, _ := buildEligibleTopUpQuoteCoupons(allCoupons, basePayable, minThreshold, 0, quoteCurrencyCode)
	return len(eligibleCoupons) > 0, nil
}

func validateSelectedCoupon(requestedCouponId int, quote *TopUpQuoteData) error {
	if requestedCouponId == 0 {
		return nil
	}
	if quote != nil && quote.SelectedCouponId == requestedCouponId {
		return nil
	}
	if quote != nil && quote.IneligibleReason != "" {
		return errors.New(quote.IneligibleReason)
	}
	return errors.New("优惠券不可用")
}

func createTopUpOrder(topUp *model.TopUp) error {
	if topUp == nil {
		return errors.New("订单不存在")
	}

	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(topUp).Error; err != nil {
			return err
		}
		if topUp.CouponId != 0 {
			if _, err := model.ReserveTopUpCouponTx(tx, topUp.CouponId, topUp.UserId, topUp); err != nil {
				return err
			}
		}
		return nil
	})
}

func expireTopUpOrderByTradeNo(tradeNo string) error {
	if tradeNo == "" {
		return errors.New("未提供支付单号")
	}

	return model.DB.Transaction(func(tx *gorm.DB) error {
		topUp := &model.TopUp{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("trade_no = ?", tradeNo).First(topUp).Error; err != nil {
			return err
		}
		if topUp.Status != common.TopUpStatusPending {
			return nil
		}
		topUp.Status = common.TopUpStatusExpired
		if err := tx.Save(topUp).Error; err != nil {
			return err
		}
		return model.ReleaseTopUpCouponReservationTx(tx, topUp)
	})
}

func findCreemProductById(productId string) (*CreemProduct, error) {
	var products []CreemProduct
	if err := json.Unmarshal([]byte(setting.CreemProducts), &products); err != nil {
		return nil, errors.New("产品配置错误")
	}
	for _, product := range products {
		if product.ProductId == productId {
			return &product, nil
		}
	}
	return nil, errors.New("产品不存在")
}

func getCreemMinimumPayable() float64 {
	var products []CreemProduct
	if err := json.Unmarshal([]byte(setting.CreemProducts), &products); err != nil || len(products) == 0 {
		return 0.01
	}

	minPrice := math.MaxFloat64
	for _, product := range products {
		if product.Price > 0 && product.Price < minPrice {
			minPrice = product.Price
		}
	}
	if minPrice == math.MaxFloat64 {
		return 0.01
	}
	return minPrice
}

func roundMoney(value decimal.Decimal) float64 {
	return value.Round(2).InexactFloat64()
}
