package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// GetRemainActualPaidAmount returns the remaining paid quota for the current user.
//
// remaining = floor(P * R / T)
// - T: total quota (remain + used), including promotional quota
// - R: remaining quota
// - P: total paid quota, calculated from top_ups
func GetRemainActualPaidAmount(c *gin.Context) {
	userId := c.GetInt("id")

	remainQuota, err := model.GetUserQuota(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	usedQuota, err := model.GetUserUsedQuota(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	R := int64(remainQuota)
	T := int64(remainQuota + usedQuota)
	if R <= 0 || T <= 0 {
		common.ApiSuccess(c, gin.H{"remain_actual_paid_amount": int64(0)})
		return
	}

	P, err := model.GetUserTotalPaidQuota(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if P <= 0 {
		common.ApiSuccess(c, gin.H{"remain_actual_paid_amount": int64(0)})
		return
	}

	remainActualPaid := decimal.NewFromInt(P).
		Mul(decimal.NewFromInt(R)).
		Div(decimal.NewFromInt(T)).
		IntPart()
	if remainActualPaid < 0 {
		remainActualPaid = 0
	}

	common.ApiSuccess(c, gin.H{"remain_actual_paid_amount": remainActualPaid})
}

type selfQuotaTakeAwayRequest struct {
	Money float64 `json:"money"`
}

// SelfQuotaTakeAway decreases the current user's remaining quota by money*QuotaPerUnit.
// It does NOT increase used_quota.
func SelfQuotaTakeAway(c *gin.Context) {
	var req selfQuotaTakeAwayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.Money <= 0 {
		common.ApiErrorMsg(c, "money 必须大于 0")
		return
	}

	quotaToTake := decimal.NewFromFloat(req.Money).
		Mul(decimal.NewFromFloat(common.QuotaPerUnit)).
		IntPart()
	if quotaToTake <= 0 {
		common.ApiErrorMsg(c, "扣减额度过小")
		return
	}
	maxInt := int64(^uint(0) >> 1)
	if quotaToTake > maxInt {
		common.ApiErrorMsg(c, "扣减额度过大")
		return
	}

	userId := c.GetInt("id")
	beforeQuota, err := model.GetUserQuota(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if int64(beforeQuota) < quotaToTake {
		common.ApiErrorMsg(c, "额度不足")
		return
	}

	if err := model.DecreaseUserQuota(userId, int(quotaToTake)); err != nil {
		common.ApiError(c, err)
		return
	}

	afterQuota := int64(beforeQuota) - quotaToTake
	model.RecordLog(
		userId,
		model.LogTypeManage,
		fmt.Sprintf(
			"用户自助扣除额度 %s（money=%.6f）余额：%s -> %s",
			logger.LogQuota(quotaToTake),
			req.Money,
			logger.LogQuota(int64(beforeQuota)),
			logger.LogQuota(afterQuota),
		),
	)

	common.ApiSuccess(c, gin.H{
		"quota_taken": quotaToTake,
		"quota_after": afterQuota,
	})
}
