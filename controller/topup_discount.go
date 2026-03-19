package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
)

func getTopupDiscountRule(originalAmount int64) (operation_setting.AmountDiscountRule, bool) {
	rule, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(originalAmount)]
	return rule, ok
}

func getTopupDiscountAmount(originalAmount int64) decimal.Decimal {
	rule, ok := getTopupDiscountRule(originalAmount)
	if !ok || rule.DiscountAmount <= 0 {
		return decimal.Zero
	}
	return decimal.NewFromFloat(rule.DiscountAmount)
}

func applyTopupDiscount(base decimal.Decimal, originalAmount int64) decimal.Decimal {
	discountAmount := getTopupDiscountAmount(originalAmount)
	if !discountAmount.GreaterThan(decimal.Zero) {
		return base
	}

	discounted := base.Sub(discountAmount)
	if discounted.IsNegative() {
		return decimal.Zero
	}
	return discounted
}

func resolveTopUpBasePayable(originalBase, discountedBase, platformDiscount decimal.Decimal, hasSelectedUserCoupon bool) (decimal.Decimal, decimal.Decimal) {
	if hasSelectedUserCoupon {
		return originalBase, decimal.Zero
	}
	return discountedBase, platformDiscount
}

func getStripeMinorUnitAmount(payMoney decimal.Decimal, currency stripe.Currency) int64 {
	if !payMoney.GreaterThan(decimal.Zero) {
		return 0
	}

	minorAmount := payMoney.Mul(getStripeCurrencyDivisor(currency)).Round(0)
	if !minorAmount.GreaterThan(decimal.Zero) {
		return 0
	}
	return minorAmount.IntPart()
}

func getStripeMajorUnitAmount(minorAmount int64, currency stripe.Currency) decimal.Decimal {
	if minorAmount <= 0 {
		return decimal.Zero
	}

	return decimal.NewFromInt(minorAmount).Div(getStripeCurrencyDivisor(currency))
}

func getStripeCheckoutQuantity(amount int64) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}
	if operation_setting.GetQuotaDisplayType() != operation_setting.QuotaDisplayTypeTokens {
		return amount, nil
	}

	quotaPerUnit := int64(decimal.NewFromFloat(common.QuotaPerUnit).IntPart())
	if quotaPerUnit <= 0 || amount%quotaPerUnit != 0 {
		return 0, invalidStripeQuantityStep(decimal.NewFromInt(quotaPerUnit).StringFixed(0))
	}
	return amount / quotaPerUnit, nil
}

func getStripeCurrencyDivisor(currency stripe.Currency) decimal.Decimal {
	switch currency {
	case stripe.CurrencyBIF,
		stripe.CurrencyCLP,
		stripe.CurrencyDJF,
		stripe.CurrencyGNF,
		stripe.CurrencyJPY,
		stripe.CurrencyKMF,
		stripe.CurrencyKRW,
		stripe.CurrencyMGA,
		stripe.CurrencyPYG,
		stripe.CurrencyRWF,
		stripe.CurrencyUGX,
		stripe.CurrencyVND,
		stripe.CurrencyVUV,
		stripe.CurrencyXAF,
		stripe.CurrencyXOF,
		stripe.CurrencyXPF:
		return decimal.NewFromInt(1)
	default:
		return decimal.NewFromInt(100)
	}
}

type invalidStripeQuantityStep string

func (s invalidStripeQuantityStep) Error() string {
	return "Stripe 充值数量必须为 " + string(s) + " 的整数倍"
}
