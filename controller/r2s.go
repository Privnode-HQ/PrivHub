package controller

import (
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

type r2sSettingsRequest struct {
	ReceiptRequired     bool   `json:"receipt_required"`
	DefaultCurrencyCode string `json:"default_currency_code"`
	BalanceReminderDays int    `json:"balance_reminder_days"`
}

type r2sSupplierRequest struct {
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	Status              string  `json:"status"`
	DefaultCurrencyCode string  `json:"default_currency_code"`
	DefaultExchangeRate float64 `json:"default_exchange_rate"`
	BalanceAmount       float64 `json:"balance_amount"`
	BalanceCurrencyCode string  `json:"balance_currency_code"`
	BalanceReminderDays int     `json:"balance_reminder_days"`
}

type r2sChannelBindingRequest struct {
	SupplierId          int     `json:"supplier_id"`
	ChannelId           int     `json:"channel_id"`
	ChannelNameSnapshot string  `json:"channel_name_snapshot"`
	UpstreamGroupName   string  `json:"upstream_group_name"`
	GroupMultiplier     float64 `json:"group_multiplier"`
	Status              string  `json:"status"`
}

type r2sPaymentRequest struct {
	SupplierId   int      `json:"supplier_id"`
	PaymentType  string   `json:"payment_type"`
	Amount       float64  `json:"amount"`
	CurrencyCode string   `json:"currency_code"`
	ExchangeRate float64  `json:"exchange_rate"`
	BalanceAfter *float64 `json:"balance_after"`
	ReceiptURL   string   `json:"receipt_url"`
	Note         string   `json:"note"`
	PaidAt       int64    `json:"paid_at"`
}

type r2sBalanceUpdateRequest struct {
	SupplierId          int     `json:"supplier_id"`
	BalanceAfter        float64 `json:"balance_after"`
	CurrencyCode        string  `json:"currency_code"`
	ExchangeRate        float64 `json:"exchange_rate"`
	BalanceReminderDays *int    `json:"balance_reminder_days"`
	Note                string  `json:"note"`
}

type r2sRecognitionRecordRequest struct {
	SourceType              string  `json:"source_type"`
	SourceReference         string  `json:"source_reference"`
	SupplierId              int     `json:"supplier_id"`
	ChannelId               int     `json:"channel_id"`
	ChannelBindingId        int     `json:"channel_binding_id"`
	PromotionCampaignId     int     `json:"promotion_campaign_id"`
	CurrencyCode            string  `json:"currency_code"`
	ExchangeRate            float64 `json:"exchange_rate"`
	RevenueAmount           float64 `json:"revenue_amount"`
	CostAmount              float64 `json:"cost_amount"`
	GroupMultiplierSnapshot float64 `json:"group_multiplier_snapshot"`
	PeriodStart             int64   `json:"period_start"`
	PeriodEnd               int64   `json:"period_end"`
	Note                    string  `json:"note"`
}

type r2sRecognitionSyncRequest struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

func GetR2SSettings(c *gin.Context) {
	common.ApiSuccess(c, model.GetR2SSettings())
}

func UpdateR2SSettings(c *gin.Context) {
	req := r2sSettingsRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := model.UpdateR2SSettings(req.ReceiptRequired, req.DefaultCurrencyCode, req.BalanceReminderDays); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, "管理员更新 R2S 系统设置")
	common.ApiSuccess(c, model.GetR2SSettings())
}

func GetR2SSuppliers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	suppliers, total, err := model.GetR2SSuppliers(pageInfo, model.R2SListFilter{
		Keyword: c.Query("keyword"),
		Status:  c.Query("status"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(suppliers)
	common.ApiSuccess(c, pageInfo)
}

func GetR2SSupplier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier, err := model.GetR2SSupplierByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, supplier)
}

func CreateR2SSupplier(c *gin.Context) {
	req := r2sSupplierRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	supplier := &model.R2SSupplier{
		Name:                req.Name,
		Description:         req.Description,
		Status:              req.Status,
		DefaultCurrencyCode: req.DefaultCurrencyCode,
		DefaultExchangeRate: req.DefaultExchangeRate,
		BalanceAmount:       req.BalanceAmount,
		BalanceCurrencyCode: req.BalanceCurrencyCode,
		BalanceReminderDays: req.BalanceReminderDays,
	}
	if err := supplier.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员创建 R2S 供应商 #%d %s", supplier.Id, supplier.Name))
	common.ApiSuccess(c, supplier)
}

func UpdateR2SSupplier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req := r2sSupplierRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	supplier, err := model.GetR2SSupplierByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier.Name = req.Name
	supplier.Description = req.Description
	supplier.Status = req.Status
	supplier.DefaultCurrencyCode = req.DefaultCurrencyCode
	supplier.DefaultExchangeRate = req.DefaultExchangeRate
	supplier.BalanceReminderDays = req.BalanceReminderDays
	if err := supplier.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员更新 R2S 供应商 #%d %s", supplier.Id, supplier.Name))
	common.ApiSuccess(c, supplier)
}

func DisableR2SSupplier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier, err := model.GetR2SSupplierByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier.Status = model.R2SStatusDisabled
	if err := supplier.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员停用 R2S 供应商 #%d %s", supplier.Id, supplier.Name))
	common.ApiSuccess(c, supplier)
}

func EnableR2SSupplier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier, err := model.GetR2SSupplierByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	supplier.Status = model.R2SStatusActive
	if err := supplier.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员启用 R2S 供应商 #%d %s", supplier.Id, supplier.Name))
	common.ApiSuccess(c, supplier)
}

func DeleteR2SSupplier(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.DeleteR2SSupplier(id); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员删除 R2S 供应商 #%d", id))
	common.ApiSuccess(c, nil)
}

func GetR2SChannelBindings(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	supplierId, _ := strconv.Atoi(c.Query("supplier_id"))
	channelId, _ := strconv.Atoi(c.Query("channel_id"))
	bindings, total, err := model.GetR2SChannelBindings(pageInfo, model.R2SListFilter{
		Keyword:    c.Query("keyword"),
		Status:     c.Query("status"),
		SupplierId: supplierId,
		ChannelId:  channelId,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(bindings)
	common.ApiSuccess(c, pageInfo)
}

func CreateR2SChannelBinding(c *gin.Context) {
	req := r2sChannelBindingRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	binding := &model.R2SChannelBinding{
		SupplierId:          req.SupplierId,
		ChannelId:           req.ChannelId,
		ChannelNameSnapshot: req.ChannelNameSnapshot,
		UpstreamGroupName:   req.UpstreamGroupName,
		GroupMultiplier:     req.GroupMultiplier,
		Status:              req.Status,
	}
	if err := binding.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员创建 R2S 渠道绑定 #%d", binding.Id))
	common.ApiSuccess(c, binding)
}

func UpdateR2SChannelBinding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req := r2sChannelBindingRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	binding, err := model.GetR2SChannelBindingByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	binding.SupplierId = req.SupplierId
	binding.ChannelId = req.ChannelId
	binding.ChannelNameSnapshot = req.ChannelNameSnapshot
	binding.UpstreamGroupName = req.UpstreamGroupName
	binding.GroupMultiplier = req.GroupMultiplier
	binding.Status = req.Status
	if err := binding.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员更新 R2S 渠道绑定 #%d", binding.Id))
	common.ApiSuccess(c, binding)
}

func DisableR2SChannelBinding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	binding, err := model.GetR2SChannelBindingByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	binding.Status = model.R2SStatusDisabled
	if err := binding.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员停用 R2S 渠道绑定 #%d", binding.Id))
	common.ApiSuccess(c, binding)
}

func GetR2SPayments(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	supplierId, _ := strconv.Atoi(c.Query("supplier_id"))
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	payments, total, err := model.GetR2SPayments(pageInfo, model.R2SListFilter{
		Keyword:    c.Query("keyword"),
		SupplierId: supplierId,
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(payments)
	common.ApiSuccess(c, pageInfo)
}

func CreateR2SPayment(c *gin.Context) {
	req := r2sPaymentRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	payment := &model.R2SPayment{
		SupplierId:   req.SupplierId,
		PaymentType:  req.PaymentType,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		ExchangeRate: req.ExchangeRate,
		ReceiptURL:   req.ReceiptURL,
		Note:         req.Note,
		PaidAt:       req.PaidAt,
	}
	payment, balanceUpdate, err := model.CreateR2SPaymentWithBalance(payment, req.BalanceAfter, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员创建 R2S 付款记录 #%d", payment.Id))
	common.ApiSuccess(c, gin.H{
		"payment":        payment,
		"balance_update": balanceUpdate,
	})
}

func GetR2SBalanceUpdates(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	supplierId, _ := strconv.Atoi(c.Query("supplier_id"))
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	updates, total, err := model.GetR2SBalanceUpdates(pageInfo, model.R2SListFilter{
		SupplierId: supplierId,
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(updates)
	common.ApiSuccess(c, pageInfo)
}

func CreateR2SBalanceUpdate(c *gin.Context) {
	req := r2sBalanceUpdateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	update, err := model.CreateR2SManualBalanceUpdate(
		req.SupplierId,
		req.BalanceAfter,
		req.CurrencyCode,
		req.ExchangeRate,
		req.BalanceReminderDays,
		req.Note,
		c.GetInt("id"),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员手动更新 R2S 供应商 #%d 余额", req.SupplierId))
	common.ApiSuccess(c, update)
}

func GetR2SRecognitionRecords(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	supplierId, _ := strconv.Atoi(c.Query("supplier_id"))
	channelId, _ := strconv.Atoi(c.Query("channel_id"))
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	records, total, err := model.GetR2SRecognitionRecords(pageInfo, model.R2SListFilter{
		Keyword:    c.Query("keyword"),
		SupplierId: supplierId,
		ChannelId:  channelId,
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func CreateR2SRecognitionRecord(c *gin.Context) {
	req := r2sRecognitionRecordRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	record := &model.R2SRecognitionRecord{
		SourceType:              req.SourceType,
		SourceReference:         req.SourceReference,
		SupplierId:              req.SupplierId,
		ChannelId:               req.ChannelId,
		ChannelBindingId:        req.ChannelBindingId,
		PromotionCampaignId:     req.PromotionCampaignId,
		CurrencyCode:            req.CurrencyCode,
		ExchangeRate:            req.ExchangeRate,
		RevenueAmount:           req.RevenueAmount,
		CostAmount:              req.CostAmount,
		GroupMultiplierSnapshot: req.GroupMultiplierSnapshot,
		PeriodStart:             req.PeriodStart,
		PeriodEnd:               req.PeriodEnd,
		Note:                    req.Note,
		CreatedByAdminId:        c.GetInt("id"),
	}
	if record.SourceType == "" && record.PromotionCampaignId > 0 {
		record.SourceType = model.R2SRecognitionSourcePromotion
	}
	if err := record.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员创建 R2S 收入识别记录 #%d", record.Id))
	common.ApiSuccess(c, record)
}

func SyncR2SRecognitionRecords(c *gin.Context) {
	req := r2sRecognitionSyncRequest{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ApiErrorMsg(c, "参数错误")
			return
		}
	}
	result, err := model.SyncR2SRecognitionFromUsageLogs(
		req.StartTime,
		req.EndTime,
		c.GetInt("id"),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(
		c.GetInt("id"),
		model.LogTypeSystem,
		fmt.Sprintf(
			"管理员同步 R2S 历史收入识别，新增 %d，更新 %d，跳过 %d",
			result.CreatedCount,
			result.UpdatedCount,
			result.SkippedCount,
		),
	)
	common.ApiSuccess(c, result)
}

func GetR2SSummary(c *gin.Context) {
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	summary, err := model.GetR2SSummary(startTime, endTime)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func GetR2STrend(c *gin.Context) {
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	rows, err := model.GetR2STrend(startTime, endTime)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

func GetR2SPromotionProfitability(c *gin.Context) {
	campaignId, _ := strconv.Atoi(c.Query("campaign_id"))
	startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	rows, err := model.GetR2SPromotionProfitability(campaignId, startTime, endTime)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}
