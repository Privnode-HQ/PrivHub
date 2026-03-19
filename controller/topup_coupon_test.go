package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/shopspring/decimal"
)

func TestBuildEligibleTopUpQuoteCouponsFiltersByCurrency(t *testing.T) {
	coupons := []*model.TopUpCoupon{
		{
			Id:              1,
			Name:            "usd-coupon",
			DeductionAmount: 10,
			CurrencyCode:    "usd",
			Status:          common.TopUpCouponStatusAvailable,
		},
		{
			Id:              2,
			Name:            "cny-coupon",
			DeductionAmount: 10,
			CurrencyCode:    "CNY",
			Status:          common.TopUpCouponStatusAvailable,
		},
		{
			Id:              3,
			Name:            "legacy-coupon",
			DeductionAmount: 5,
			Status:          common.TopUpCouponStatusAvailable,
		},
	}

	eligibleCoupons, selectedCoupon, ineligibleReason := buildEligibleTopUpQuoteCoupons(
		coupons,
		decimal.RequireFromString("100"),
		decimal.RequireFromString("1"),
		2,
		"USD",
	)

	if len(eligibleCoupons) != 2 {
		t.Fatalf("expected 2 eligible coupons, got %d", len(eligibleCoupons))
	}
	if eligibleCoupons[0].Id != 1 || eligibleCoupons[0].CurrencyCode != "USD" {
		t.Fatalf("expected first eligible coupon to be USD coupon, got %+v", eligibleCoupons[0])
	}
	if eligibleCoupons[1].Id != 3 || eligibleCoupons[1].CurrencyCode != "USD" {
		t.Fatalf("expected legacy coupon to inherit USD display currency, got %+v", eligibleCoupons[1])
	}
	if selectedCoupon != nil {
		t.Fatalf("expected no selected coupon, got %+v", selectedCoupon)
	}
	if ineligibleReason != "优惠券币种与当前支付方式不匹配" {
		t.Fatalf("expected currency mismatch reason, got %q", ineligibleReason)
	}
}

func TestBuildEligibleTopUpQuoteCouponsRejectsBelowMinimumAfterCurrencyCheck(t *testing.T) {
	coupons := []*model.TopUpCoupon{
		{
			Id:              1,
			Name:            "usd-coupon",
			DeductionAmount: 99,
			CurrencyCode:    "USD",
			Status:          common.TopUpCouponStatusAvailable,
		},
	}

	eligibleCoupons, selectedCoupon, ineligibleReason := buildEligibleTopUpQuoteCoupons(
		coupons,
		decimal.RequireFromString("100"),
		decimal.RequireFromString("1"),
		1,
		"USD",
	)

	if len(eligibleCoupons) != 0 {
		t.Fatalf("expected no eligible coupons, got %d", len(eligibleCoupons))
	}
	if selectedCoupon != nil {
		t.Fatalf("expected no selected coupon, got %+v", selectedCoupon)
	}
	if ineligibleReason != "使用优惠券后金额必须高于最低充值金额" {
		t.Fatalf("expected min threshold reason, got %q", ineligibleReason)
	}
}
