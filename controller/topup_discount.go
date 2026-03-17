package controller

import (
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

func getTopupDiscountAmount(originalAmount int64) decimal.Decimal {
	discount, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(originalAmount)]
	if !ok || discount <= 0 {
		return decimal.Zero
	}
	return decimal.NewFromFloat(discount)
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

func getStripeMinorUnitAmount(payMoney decimal.Decimal, stripeUnitPrice float64, stripeUnitAmount int64) int64 {
	if !payMoney.GreaterThan(decimal.Zero) {
		return 0
	}

	multiplier := decimal.NewFromInt(100)
	if stripeUnitPrice > 0 && stripeUnitAmount > 0 {
		multiplier = decimal.NewFromInt(stripeUnitAmount).Div(decimal.NewFromFloat(stripeUnitPrice))
	}

	minorAmount := payMoney.Mul(multiplier).Round(0)
	if !minorAmount.GreaterThan(decimal.Zero) {
		return 0
	}
	return minorAmount.IntPart()
}
