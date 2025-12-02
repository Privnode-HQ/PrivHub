package model

import (
	"github.com/QuantumNous/new-api/common"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const affRebateMinMoney = 30.0

var affRebateTiers = []struct {
	maxPurchase int
	rate        float64
}{
	{maxPurchase: 1, rate: 0.05},
	{maxPurchase: 11, rate: 0.02},
	{maxPurchase: 21, rate: 0.01},
}

const affDefaultRebateRate = 0.01

func getAffRebateRate(purchaseCount int) float64 {
	for _, tier := range affRebateTiers {
		if purchaseCount <= tier.maxPurchase {
			return tier.rate
		}
	}
	return affDefaultRebateRate
}

// AffRebateLog captures each rebate credit for auditability.
type AffRebateLog struct {
	Id            int     `json:"id"`
	InviterId     int     `json:"inviter_id" gorm:"index"`
	InviteeId     int     `json:"invitee_id" gorm:"index"`
	TopUpId       int     `json:"top_up_id" gorm:"uniqueIndex"`
	RewardQuota   int64   `json:"reward_quota"`
	RebateRate    float64 `json:"rebate_rate"`
	PurchaseCount int     `json:"purchase_count"`
	PayMoney      float64 `json:"pay_money"`
	CreatedAt     int64   `json:"created_at" gorm:"autoCreateTime:false"`
}

// AffRebateResult lets callers log inviter notifications after the transaction commits.
type AffRebateResult struct {
	InviterId       int
	RewardQuota     int64
	InviteeUsername string
}

// ApplyAffRebateTx adds inviter rewards when an invitee completes an eligible top-up.
// The logic needs to run within the caller's transaction to avoid crediting rebates when top-ups fail.
func ApplyAffRebateTx(tx *gorm.DB, inviteeId int, topUpId int, quotaAdded int64, payMoney float64) (*AffRebateResult, error) {
	if quotaAdded <= 0 || payMoney < affRebateMinMoney {
		return nil, nil
	}
	if tx == nil {
		tx = DB
	}
	if topUpId > 0 {
		var existing int64
		if err := tx.Model(&AffRebateLog{}).Where("top_up_id = ?", topUpId).Count(&existing).Error; err != nil {
			return nil, err
		}
		if existing > 0 {
			return nil, nil
		}
	}

	var invitee User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", inviteeId).First(&invitee).Error; err != nil {
		return nil, err
	}
	if invitee.InviterId == 0 {
		return nil, nil
	}

	var purchaseCount int64
	if err := tx.Model(&TopUp{}).
		Where("user_id = ? AND status = ? AND money >= ?", invitee.Id, common.TopUpStatusSuccess, affRebateMinMoney).
		Count(&purchaseCount).Error; err != nil {
		return nil, err
	}
	// purchaseCount is post-increment: the current successful top-up is already persisted before this query.
	rate := getAffRebateRate(int(purchaseCount))
	if rate <= 0 {
		return nil, nil
	}

	reward := decimal.NewFromInt(quotaAdded).Mul(decimal.NewFromFloat(rate)).IntPart()
	if reward <= 0 {
		return nil, nil
	}

	logEntry := &AffRebateLog{
		InviterId:     invitee.InviterId,
		InviteeId:     invitee.Id,
		TopUpId:       topUpId,
		RewardQuota:   reward,
		RebateRate:    rate,
		PurchaseCount: int(purchaseCount),
		PayMoney:      payMoney,
		CreatedAt:     common.GetTimestamp(),
	}
	if err := tx.Create(logEntry).Error; err != nil {
		return nil, err
	}

	updateErr := tx.Model(&User{}).Where("id = ?", invitee.InviterId).Updates(map[string]interface{}{
		"aff_quota":   gorm.Expr("aff_quota + ?", reward),
		"aff_history": gorm.Expr("aff_history + ?", reward),
	}).Error
	if updateErr != nil {
		return nil, updateErr
	}

	return &AffRebateResult{
		InviterId:       invitee.InviterId,
		RewardQuota:     reward,
		InviteeUsername: invitee.Username,
	}, nil
}
