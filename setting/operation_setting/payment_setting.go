package operation_setting

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/QuantumNous/new-api/setting/config"
)

type AmountDiscountRule struct {
	DiscountAmount float64
	CouponID       string
}

func (r AmountDiscountRule) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{r.DiscountAmount, r.CouponID})
}

func (r *AmountDiscountRule) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*r = AmountDiscountRule{}
		return nil
	}

	var legacyAmount float64
	if err := json.Unmarshal(data, &legacyAmount); err == nil {
		r.DiscountAmount = legacyAmount
		r.CouponID = ""
		return nil
	}

	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return fmt.Errorf("invalid amount discount rule: %w", err)
	}
	if len(rawItems) == 0 {
		*r = AmountDiscountRule{}
		return nil
	}

	if err := json.Unmarshal(rawItems[0], &r.DiscountAmount); err != nil {
		return fmt.Errorf("invalid discount amount: %w", err)
	}
	r.CouponID = ""
	if len(rawItems) > 1 && !bytes.Equal(bytes.TrimSpace(rawItems[1]), []byte("null")) {
		if err := json.Unmarshal(rawItems[1], &r.CouponID); err != nil {
			return fmt.Errorf("invalid coupon id: %w", err)
		}
	}
	return nil
}

type PaymentSetting struct {
	AmountOptions  []int                      `json:"amount_options"`
	AmountDiscount map[int]AmountDiscountRule `json:"amount_discount"` // 充值金额对应的优惠配置，例如 100 元 [5, "coupon_id"] 表示充值 100 元减免 5 元，并在 Stripe 预设 coupon
}

// 默认配置
var paymentSetting = PaymentSetting{
	AmountOptions:  []int{10, 20, 50, 100, 200, 500},
	AmountDiscount: map[int]AmountDiscountRule{},
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}

func (p *PaymentSetting) GetAmountDiscountValues() map[int]float64 {
	result := make(map[int]float64, len(p.AmountDiscount))
	for amount, rule := range p.AmountDiscount {
		result[amount] = rule.DiscountAmount
	}
	return result
}
