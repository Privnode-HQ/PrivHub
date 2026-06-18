package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	R2SStatusActive   = "active"
	R2SStatusDisabled = "disabled"

	R2SPaymentTypePrepaid    = "prepaid"
	R2SPaymentTypePostpaid   = "postpaid"
	R2SPaymentTypeGrant      = "grant"
	R2SPaymentTypeRefund     = "refund"
	R2SPaymentTypeAdjustment = "adjustment"

	R2SBalanceUpdateTypeManual = "manual"

	R2SRecognitionSourceManual    = "manual"
	R2SRecognitionSourcePromotion = "promotion"

	R2SReceiptRequiredOptionKey     = "R2SReceiptRequired"
	R2SDefaultCurrencyCodeOptionKey = "R2SDefaultCurrencyCode"
	R2SBalanceReminderDaysOptionKey = "R2SBalanceReminderDays"
)

type R2SSettings struct {
	ReceiptRequired     bool   `json:"receipt_required"`
	DefaultCurrencyCode string `json:"default_currency_code"`
	BalanceReminderDays int    `json:"balance_reminder_days"`
	SystemCurrencyCode  string `json:"system_currency_code"`
}

type R2SSupplier struct {
	Id                    int     `json:"id"`
	Name                  string  `json:"name" gorm:"type:varchar(100);index"`
	Description           string  `json:"description" gorm:"type:varchar(255)"`
	Status                string  `json:"status" gorm:"type:varchar(16);index"`
	DefaultCurrencyCode   string  `json:"default_currency_code" gorm:"type:varchar(16);index"`
	DefaultExchangeRate   float64 `json:"default_exchange_rate"`
	BalanceAmount         float64 `json:"balance_amount"`
	BalanceCurrencyCode   string  `json:"balance_currency_code" gorm:"type:varchar(16);index"`
	SystemBalanceAmount   float64 `json:"system_balance_amount"`
	BalanceUpdatedTime    int64   `json:"balance_updated_time" gorm:"index"`
	BalanceReminderDays   int     `json:"balance_reminder_days"`
	NextBalanceReminderAt int64   `json:"next_balance_reminder_at" gorm:"index"`
	CreatedTime           int64   `json:"created_time" gorm:"index"`
	UpdatedTime           int64   `json:"updated_time"`

	ChannelCount    int64 `json:"channel_count" gorm:"-"`
	LastPaymentTime int64 `json:"last_payment_time" gorm:"-"`
}

type R2SChannelBinding struct {
	Id                  int     `json:"id"`
	SupplierId          int     `json:"supplier_id" gorm:"index"`
	SupplierName        string  `json:"supplier_name" gorm:"-"`
	ChannelId           int     `json:"channel_id" gorm:"index"`
	ChannelNameSnapshot string  `json:"channel_name_snapshot" gorm:"type:varchar(255)"`
	UpstreamGroupName   string  `json:"upstream_group_name" gorm:"type:varchar(100);index"`
	GroupMultiplier     float64 `json:"group_multiplier"`
	Status              string  `json:"status" gorm:"type:varchar(16);index"`
	CreatedTime         int64   `json:"created_time" gorm:"index"`
	UpdatedTime         int64   `json:"updated_time"`
}

type R2SPayment struct {
	Id                   int     `json:"id"`
	SupplierId           int     `json:"supplier_id" gorm:"index"`
	SupplierNameSnapshot string  `json:"supplier_name_snapshot" gorm:"type:varchar(100);index"`
	PaymentType          string  `json:"payment_type" gorm:"type:varchar(32);index"`
	Amount               float64 `json:"amount"`
	CurrencyCode         string  `json:"currency_code" gorm:"type:varchar(16);index"`
	ExchangeRate         float64 `json:"exchange_rate"`
	SystemAmount         float64 `json:"system_amount"`
	BalanceBefore        float64 `json:"balance_before"`
	BalanceAfter         float64 `json:"balance_after"`
	ReceiptURL           string  `json:"receipt_url" gorm:"type:varchar(1024)"`
	ReceiptRequired      bool    `json:"receipt_required"`
	Note                 string  `json:"note" gorm:"type:varchar(500)"`
	PaidAt               int64   `json:"paid_at" gorm:"index"`
	CreatedByAdminId     int     `json:"created_by_admin_id" gorm:"index"`
	CreatedTime          int64   `json:"created_time" gorm:"index"`
}

type R2SBalanceUpdate struct {
	Id                   int     `json:"id"`
	SupplierId           int     `json:"supplier_id" gorm:"index"`
	SupplierNameSnapshot string  `json:"supplier_name_snapshot" gorm:"type:varchar(100);index"`
	UpdateType           string  `json:"update_type" gorm:"type:varchar(32);index"`
	BalanceBefore        float64 `json:"balance_before"`
	BalanceAfter         float64 `json:"balance_after"`
	DeltaAmount          float64 `json:"delta_amount"`
	CurrencyCode         string  `json:"currency_code" gorm:"type:varchar(16);index"`
	ExchangeRate         float64 `json:"exchange_rate"`
	SystemDeltaAmount    float64 `json:"system_delta_amount"`
	ReminderDaysSnapshot int     `json:"reminder_days_snapshot"`
	NextReminderAt       int64   `json:"next_reminder_at" gorm:"index"`
	Note                 string  `json:"note" gorm:"type:varchar(500)"`
	CreatedByAdminId     int     `json:"created_by_admin_id" gorm:"index"`
	CreatedTime          int64   `json:"created_time" gorm:"index"`
}

type R2SRecognitionRecord struct {
	Id                        int     `json:"id"`
	SourceType                string  `json:"source_type" gorm:"type:varchar(32);index"`
	SourceReference           string  `json:"source_reference" gorm:"type:varchar(100);index"`
	SupplierId                int     `json:"supplier_id" gorm:"index"`
	SupplierNameSnapshot      string  `json:"supplier_name_snapshot" gorm:"type:varchar(100);index"`
	ChannelId                 int     `json:"channel_id" gorm:"index"`
	ChannelBindingId          int     `json:"channel_binding_id" gorm:"index"`
	ChannelNameSnapshot       string  `json:"channel_name_snapshot" gorm:"type:varchar(255)"`
	UpstreamGroupNameSnapshot string  `json:"upstream_group_name_snapshot" gorm:"type:varchar(100);index"`
	GroupMultiplierSnapshot   float64 `json:"group_multiplier_snapshot"`
	PromotionCampaignId       int     `json:"promotion_campaign_id" gorm:"index"`
	PromotionCampaignName     string  `json:"promotion_campaign_name" gorm:"type:varchar(80);index"`
	CurrencyCode              string  `json:"currency_code" gorm:"type:varchar(16);index"`
	ExchangeRate              float64 `json:"exchange_rate"`
	RevenueAmount             float64 `json:"revenue_amount"`
	CostAmount                float64 `json:"cost_amount"`
	SystemRevenueAmount       float64 `json:"system_revenue_amount"`
	SystemCostAmount          float64 `json:"system_cost_amount"`
	SystemProfitAmount        float64 `json:"system_profit_amount"`
	ProfitMargin              float64 `json:"profit_margin"`
	PeriodStart               int64   `json:"period_start" gorm:"index"`
	PeriodEnd                 int64   `json:"period_end" gorm:"index"`
	Note                      string  `json:"note" gorm:"type:varchar(500)"`
	CreatedByAdminId          int     `json:"created_by_admin_id" gorm:"index"`
	CreatedTime               int64   `json:"created_time" gorm:"index"`
}

type R2SListFilter struct {
	Keyword    string
	Status     string
	SupplierId int
	ChannelId  int
	StartTime  int64
	EndTime    int64
}

type R2SSummary struct {
	SystemCurrencyCode      string  `json:"system_currency_code"`
	RecognizedRevenueAmount float64 `json:"recognized_revenue_amount"`
	RecognizedCostAmount    float64 `json:"recognized_cost_amount"`
	RecognizedProfitAmount  float64 `json:"recognized_profit_amount"`
	ProfitMargin            float64 `json:"profit_margin"`
	PaymentSystemAmount     float64 `json:"payment_system_amount"`
	SupplierBalanceAmount   float64 `json:"supplier_balance_amount"`
	SupplierCount           int64   `json:"supplier_count"`
	ActiveSupplierCount     int64   `json:"active_supplier_count"`
	ChannelBindingCount     int64   `json:"channel_binding_count"`
	ReminderDueCount        int64   `json:"reminder_due_count"`
}

type R2SPromotionProfitability struct {
	CampaignId           int     `json:"campaign_id"`
	CampaignName         string  `json:"campaign_name"`
	TopUpCount           int64   `json:"top_up_count"`
	GrossRevenueAmount   float64 `json:"gross_revenue_amount"`
	DiscountAmount       float64 `json:"discount_amount"`
	NetRevenueAmount     float64 `json:"net_revenue_amount"`
	RecognizedCostAmount float64 `json:"recognized_cost_amount"`
	ProfitAmount         float64 `json:"profit_amount"`
	ProfitMargin         float64 `json:"profit_margin"`
	CurrencyCode         string  `json:"currency_code"`
	SystemCurrencyCode   string  `json:"system_currency_code"`
	ProfitCalculated     bool    `json:"profit_calculated"`
}

func (R2SSupplier) TableName() string {
	return "r2s_suppliers"
}

func (R2SChannelBinding) TableName() string {
	return "r2s_channel_bindings"
}

func (R2SPayment) TableName() string {
	return "r2s_payments"
}

func (R2SBalanceUpdate) TableName() string {
	return "r2s_balance_updates"
}

func (R2SRecognitionRecord) TableName() string {
	return "r2s_recognition_records"
}

func GetR2SSettings() R2SSettings {
	defaultCurrency := getR2SOption(R2SDefaultCurrencyCodeOptionKey, "")
	if strings.TrimSpace(defaultCurrency) == "" {
		defaultCurrency = operation_setting.GetQuotaDisplayType()
	}
	reminderDays, err := strconv.Atoi(getR2SOption(R2SBalanceReminderDaysOptionKey, "30"))
	if err != nil || reminderDays < 0 {
		reminderDays = 30
	}
	receiptRequired, _ := strconv.ParseBool(getR2SOption(R2SReceiptRequiredOptionKey, "false"))
	return R2SSettings{
		ReceiptRequired:     receiptRequired,
		DefaultCurrencyCode: NormalizeR2SCurrencyCode(defaultCurrency),
		BalanceReminderDays: reminderDays,
		SystemCurrencyCode:  operation_setting.GetQuotaDisplayType(),
	}
}

func UpdateR2SSettings(receiptRequired bool, defaultCurrencyCode string, balanceReminderDays int) error {
	defaultCurrencyCode = NormalizeR2SCurrencyCode(defaultCurrencyCode)
	if defaultCurrencyCode == "" {
		defaultCurrencyCode = operation_setting.GetQuotaDisplayType()
	}
	if !IsValidR2SCurrencyCode(defaultCurrencyCode) {
		return errors.New("默认货币格式不正确")
	}
	if balanceReminderDays < 0 {
		return errors.New("余额提醒天数不能小于 0")
	}
	if err := UpdateOption(R2SReceiptRequiredOptionKey, strconv.FormatBool(receiptRequired)); err != nil {
		return err
	}
	if err := UpdateOption(R2SDefaultCurrencyCodeOptionKey, defaultCurrencyCode); err != nil {
		return err
	}
	return UpdateOption(R2SBalanceReminderDaysOptionKey, strconv.Itoa(balanceReminderDays))
}

func NormalizeR2SCurrencyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func IsValidR2SCurrencyCode(value string) bool {
	value = NormalizeR2SCurrencyCode(value)
	if len(value) < 2 || len(value) > 16 {
		return false
	}
	for _, r := range value {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func CalculateR2SSystemAmount(amount float64, exchangeRate float64) float64 {
	return roundR2SDecimal(decimal.NewFromFloat(amount).Mul(decimal.NewFromFloat(exchangeRate)))
}

func CalculateR2SProfitMargin(revenue float64, profit float64) float64 {
	revenueDecimal := decimal.NewFromFloat(revenue)
	if !revenueDecimal.GreaterThan(decimal.Zero) {
		return 0
	}
	return roundR2SDecimal(decimal.NewFromFloat(profit).Div(revenueDecimal).Mul(decimal.NewFromInt(100)))
}

func GetR2SSupplierByID(id int) (*R2SSupplier, error) {
	if id <= 0 {
		return nil, errors.New("缺少供应商 ID")
	}
	supplier := &R2SSupplier{}
	if err := DB.Where("id = ?", id).First(supplier).Error; err != nil {
		return nil, err
	}
	if err := fillR2SSupplierViewData([]*R2SSupplier{supplier}); err != nil {
		return nil, err
	}
	return supplier, nil
}

func GetR2SSuppliers(pageInfo *common.PageInfo, filter R2SListFilter) ([]*R2SSupplier, int64, error) {
	query := applyR2SSupplierFilter(DB.Model(&R2SSupplier{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var suppliers []*R2SSupplier
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}
	if err := fillR2SSupplierViewData(suppliers); err != nil {
		return nil, 0, err
	}
	return suppliers, total, nil
}

func (supplier *R2SSupplier) Insert() error {
	now := common.GetTimestamp()
	supplier.CreatedTime = now
	supplier.UpdatedTime = now
	if supplier.BalanceUpdatedTime == 0 {
		supplier.BalanceUpdatedTime = now
	}
	if err := supplier.Validate(); err != nil {
		return err
	}
	return DB.Create(supplier).Error
}

func (supplier *R2SSupplier) Update() error {
	supplier.UpdatedTime = common.GetTimestamp()
	if err := supplier.Validate(); err != nil {
		return err
	}
	return DB.Save(supplier).Error
}

func DeleteR2SSupplier(id int) error {
	if id <= 0 {
		return errors.New("缺少供应商 ID")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		supplier := &R2SSupplier{}
		if err := tx.Where("id = ?", id).First(supplier).Error; err != nil {
			return err
		}
		historyChecks := []struct {
			name  string
			model any
		}{
			{name: "付款历史", model: &R2SPayment{}},
			{name: "余额更新历史", model: &R2SBalanceUpdate{}},
			{name: "收入识别记录", model: &R2SRecognitionRecord{}},
		}
		for _, check := range historyChecks {
			var count int64
			if err := tx.Model(check.model).Where("supplier_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("供应商已有%s，不能删除，请停用供应商保留历史", check.name)
			}
		}
		if err := tx.Where("supplier_id = ?", id).Delete(&R2SChannelBinding{}).Error; err != nil {
			return err
		}
		return tx.Delete(supplier).Error
	})
}

func (supplier *R2SSupplier) Validate() error {
	if supplier == nil {
		return errors.New("供应商不存在")
	}
	supplier.Name = strings.TrimSpace(supplier.Name)
	if supplier.Name == "" {
		return errors.New("供应商名称不能为空")
	}
	if utf8.RuneCountInString(supplier.Name) > 100 {
		return errors.New("供应商名称长度不能超过 100")
	}
	supplier.Description = strings.TrimSpace(supplier.Description)
	if utf8.RuneCountInString(supplier.Description) > 255 {
		return errors.New("供应商说明长度不能超过 255")
	}
	if supplier.Status == "" {
		supplier.Status = R2SStatusActive
	}
	if supplier.Status != R2SStatusActive && supplier.Status != R2SStatusDisabled {
		return errors.New("供应商状态不正确")
	}
	settings := GetR2SSettings()
	supplier.DefaultCurrencyCode = NormalizeR2SCurrencyCode(supplier.DefaultCurrencyCode)
	if supplier.DefaultCurrencyCode == "" {
		supplier.DefaultCurrencyCode = settings.DefaultCurrencyCode
	}
	if !IsValidR2SCurrencyCode(supplier.DefaultCurrencyCode) {
		return errors.New("供应商默认货币格式不正确")
	}
	supplier.BalanceCurrencyCode = NormalizeR2SCurrencyCode(supplier.BalanceCurrencyCode)
	if supplier.BalanceCurrencyCode == "" {
		supplier.BalanceCurrencyCode = supplier.DefaultCurrencyCode
	}
	if !IsValidR2SCurrencyCode(supplier.BalanceCurrencyCode) {
		return errors.New("供应商余额货币格式不正确")
	}
	if supplier.DefaultExchangeRate == 0 {
		supplier.DefaultExchangeRate = 1
	}
	if supplier.DefaultExchangeRate <= 0 || math.IsNaN(supplier.DefaultExchangeRate) || math.IsInf(supplier.DefaultExchangeRate, 0) {
		return errors.New("供应商默认汇率必须大于 0")
	}
	if supplier.BalanceReminderDays < 0 {
		return errors.New("余额提醒天数不能小于 0")
	}
	if supplier.BalanceReminderDays == 0 && supplier.Id == 0 {
		supplier.BalanceReminderDays = settings.BalanceReminderDays
	}
	supplier.SystemBalanceAmount = CalculateR2SSystemAmount(supplier.BalanceAmount, supplier.DefaultExchangeRate)
	if supplier.BalanceReminderDays > 0 {
		supplier.NextBalanceReminderAt = supplier.BalanceUpdatedTime + int64(supplier.BalanceReminderDays)*86400
	} else {
		supplier.NextBalanceReminderAt = 0
	}
	return nil
}

func GetR2SChannelBindings(pageInfo *common.PageInfo, filter R2SListFilter) ([]*R2SChannelBinding, int64, error) {
	query := DB.Model(&R2SChannelBinding{})
	if filter.SupplierId > 0 {
		query = query.Where("supplier_id = ?", filter.SupplierId)
	}
	if filter.ChannelId > 0 {
		query = query.Where("channel_id = ?", filter.ChannelId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("channel_name_snapshot LIKE ? OR upstream_group_name LIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var bindings []*R2SChannelBinding
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&bindings).Error; err != nil {
		return nil, 0, err
	}
	if err := fillR2SChannelBindingViewData(bindings); err != nil {
		return nil, 0, err
	}
	return bindings, total, nil
}

func GetR2SChannelBindingByID(id int) (*R2SChannelBinding, error) {
	if id <= 0 {
		return nil, errors.New("缺少渠道绑定 ID")
	}
	binding := &R2SChannelBinding{}
	if err := DB.Where("id = ?", id).First(binding).Error; err != nil {
		return nil, err
	}
	if err := fillR2SChannelBindingViewData([]*R2SChannelBinding{binding}); err != nil {
		return nil, err
	}
	return binding, nil
}

func (binding *R2SChannelBinding) Validate() error {
	if binding == nil {
		return errors.New("渠道绑定不存在")
	}
	if binding.SupplierId <= 0 {
		return errors.New("缺少供应商 ID")
	}
	if binding.ChannelId <= 0 {
		return errors.New("缺少渠道 ID")
	}
	if binding.GroupMultiplier == 0 {
		binding.GroupMultiplier = 1
	}
	if binding.GroupMultiplier <= 0 || math.IsNaN(binding.GroupMultiplier) || math.IsInf(binding.GroupMultiplier, 0) {
		return errors.New("上游分组倍率必须大于 0")
	}
	if binding.Status == "" {
		binding.Status = R2SStatusActive
	}
	if binding.Status != R2SStatusActive && binding.Status != R2SStatusDisabled {
		return errors.New("渠道绑定状态不正确")
	}
	binding.UpstreamGroupName = strings.TrimSpace(binding.UpstreamGroupName)
	if utf8.RuneCountInString(binding.UpstreamGroupName) > 100 {
		return errors.New("上游分组名称长度不能超过 100")
	}
	supplier, err := GetR2SSupplierByID(binding.SupplierId)
	if err != nil {
		return err
	}
	binding.SupplierName = supplier.Name
	channel := &Channel{}
	if err := DB.Where("id = ?", binding.ChannelId).First(channel).Error; err != nil {
		return err
	}
	if strings.TrimSpace(binding.ChannelNameSnapshot) == "" {
		binding.ChannelNameSnapshot = channel.Name
	}
	if binding.UpstreamGroupName == "" {
		binding.UpstreamGroupName = firstR2SChannelGroup(channel.Group)
	}
	return nil
}

func (binding *R2SChannelBinding) Insert() error {
	now := common.GetTimestamp()
	binding.CreatedTime = now
	binding.UpdatedTime = now
	if err := binding.Validate(); err != nil {
		return err
	}
	if err := ensureR2SChannelBindingUnique(binding.ChannelId, binding.Id); err != nil {
		return err
	}
	return DB.Create(binding).Error
}

func (binding *R2SChannelBinding) Update() error {
	binding.UpdatedTime = common.GetTimestamp()
	if err := binding.Validate(); err != nil {
		return err
	}
	if err := ensureR2SChannelBindingUnique(binding.ChannelId, binding.Id); err != nil {
		return err
	}
	return DB.Save(binding).Error
}

func GetR2SPayments(pageInfo *common.PageInfo, filter R2SListFilter) ([]*R2SPayment, int64, error) {
	query := DB.Model(&R2SPayment{})
	if filter.SupplierId > 0 {
		query = query.Where("supplier_id = ?", filter.SupplierId)
	}
	if filter.StartTime > 0 {
		query = query.Where("paid_at >= ?", filter.StartTime)
	}
	if filter.EndTime > 0 {
		query = query.Where("paid_at <= ?", filter.EndTime)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("supplier_name_snapshot LIKE ? OR payment_type LIKE ? OR note LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var payments []*R2SPayment
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}

func (payment *R2SPayment) Validate() error {
	if payment == nil {
		return errors.New("付款记录不存在")
	}
	if payment.SupplierId <= 0 {
		return errors.New("缺少供应商 ID")
	}
	if !isValidR2SPaymentType(payment.PaymentType) {
		return errors.New("付款类型不正确")
	}
	if payment.Amount <= 0 || math.IsNaN(payment.Amount) || math.IsInf(payment.Amount, 0) {
		return errors.New("付款金额必须大于 0")
	}
	payment.CurrencyCode = NormalizeR2SCurrencyCode(payment.CurrencyCode)
	if payment.CurrencyCode == "" {
		payment.CurrencyCode = GetR2SSettings().DefaultCurrencyCode
	}
	if !IsValidR2SCurrencyCode(payment.CurrencyCode) {
		return errors.New("付款货币格式不正确")
	}
	if payment.ExchangeRate == 0 {
		payment.ExchangeRate = 1
	}
	if payment.ExchangeRate <= 0 || math.IsNaN(payment.ExchangeRate) || math.IsInf(payment.ExchangeRate, 0) {
		return errors.New("付款汇率必须大于 0")
	}
	if payment.ReceiptRequired && strings.TrimSpace(payment.ReceiptURL) == "" {
		return errors.New("当前设置要求上传收据或支付截图")
	}
	payment.ReceiptURL = strings.TrimSpace(payment.ReceiptURL)
	payment.Note = strings.TrimSpace(payment.Note)
	if utf8.RuneCountInString(payment.Note) > 500 {
		return errors.New("备注长度不能超过 500")
	}
	payment.SystemAmount = CalculateR2SSystemAmount(payment.Amount, payment.ExchangeRate)
	if payment.PaidAt == 0 {
		payment.PaidAt = common.GetTimestamp()
	}
	return nil
}

func GetR2SBalanceUpdates(pageInfo *common.PageInfo, filter R2SListFilter) ([]*R2SBalanceUpdate, int64, error) {
	query := DB.Model(&R2SBalanceUpdate{})
	if filter.SupplierId > 0 {
		query = query.Where("supplier_id = ?", filter.SupplierId)
	}
	if filter.StartTime > 0 {
		query = query.Where("created_time >= ?", filter.StartTime)
	}
	if filter.EndTime > 0 {
		query = query.Where("created_time <= ?", filter.EndTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var updates []*R2SBalanceUpdate
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&updates).Error; err != nil {
		return nil, 0, err
	}
	return updates, total, nil
}

func (update *R2SBalanceUpdate) Validate() error {
	if update == nil {
		return errors.New("余额更新记录不存在")
	}
	if update.SupplierId <= 0 {
		return errors.New("缺少供应商 ID")
	}
	if update.UpdateType == "" {
		update.UpdateType = R2SBalanceUpdateTypeManual
	}
	update.CurrencyCode = NormalizeR2SCurrencyCode(update.CurrencyCode)
	if update.CurrencyCode == "" {
		update.CurrencyCode = GetR2SSettings().DefaultCurrencyCode
	}
	if !IsValidR2SCurrencyCode(update.CurrencyCode) {
		return errors.New("余额货币格式不正确")
	}
	if update.ExchangeRate == 0 {
		update.ExchangeRate = 1
	}
	if update.ExchangeRate <= 0 || math.IsNaN(update.ExchangeRate) || math.IsInf(update.ExchangeRate, 0) {
		return errors.New("余额汇率必须大于 0")
	}
	update.DeltaAmount = roundR2SDecimal(decimal.NewFromFloat(update.BalanceAfter).Sub(decimal.NewFromFloat(update.BalanceBefore)))
	update.SystemDeltaAmount = CalculateR2SSystemAmount(update.DeltaAmount, update.ExchangeRate)
	update.Note = strings.TrimSpace(update.Note)
	if utf8.RuneCountInString(update.Note) > 500 {
		return errors.New("备注长度不能超过 500")
	}
	return nil
}

func GetR2SRecognitionRecords(pageInfo *common.PageInfo, filter R2SListFilter) ([]*R2SRecognitionRecord, int64, error) {
	query := DB.Model(&R2SRecognitionRecord{})
	if filter.SupplierId > 0 {
		query = query.Where("supplier_id = ?", filter.SupplierId)
	}
	if filter.ChannelId > 0 {
		query = query.Where("channel_id = ?", filter.ChannelId)
	}
	if filter.StartTime > 0 {
		query = query.Where("period_end >= ?", filter.StartTime)
	}
	if filter.EndTime > 0 {
		query = query.Where("period_start <= ?", filter.EndTime)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("supplier_name_snapshot LIKE ? OR channel_name_snapshot LIKE ? OR source_reference LIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []*R2SRecognitionRecord
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (record *R2SRecognitionRecord) ValidateAndSnapshot() error {
	if record == nil {
		return errors.New("收入识别记录不存在")
	}
	if record.SupplierId <= 0 {
		return errors.New("缺少供应商 ID")
	}
	if record.RevenueAmount < 0 || record.CostAmount < 0 {
		return errors.New("收入与成本不能小于 0")
	}
	if record.SourceType == "" {
		if record.PromotionCampaignId > 0 {
			record.SourceType = R2SRecognitionSourcePromotion
		} else {
			record.SourceType = R2SRecognitionSourceManual
		}
	}
	if record.SourceType != R2SRecognitionSourceManual && record.SourceType != R2SRecognitionSourcePromotion {
		return errors.New("收入识别来源不正确")
	}
	supplier, err := GetR2SSupplierByID(record.SupplierId)
	if err != nil {
		return err
	}
	record.SupplierNameSnapshot = supplier.Name
	if record.CurrencyCode == "" {
		record.CurrencyCode = supplier.DefaultCurrencyCode
	}
	record.CurrencyCode = NormalizeR2SCurrencyCode(record.CurrencyCode)
	if !IsValidR2SCurrencyCode(record.CurrencyCode) {
		return errors.New("收入识别货币格式不正确")
	}
	if record.ExchangeRate == 0 {
		record.ExchangeRate = supplier.DefaultExchangeRate
	}
	if record.ExchangeRate <= 0 || math.IsNaN(record.ExchangeRate) || math.IsInf(record.ExchangeRate, 0) {
		return errors.New("收入识别汇率必须大于 0")
	}
	if err := record.fillBindingSnapshot(); err != nil {
		return err
	}
	if record.PromotionCampaignId > 0 {
		campaign, err := GetTopUpPromotionCampaignById(record.PromotionCampaignId)
		if err != nil {
			return err
		}
		record.PromotionCampaignName = campaign.Name
	}
	if record.PeriodEnd != 0 && record.PeriodStart != 0 && record.PeriodEnd < record.PeriodStart {
		return errors.New("结束时间不能早于开始时间")
	}
	record.Note = strings.TrimSpace(record.Note)
	if utf8.RuneCountInString(record.Note) > 500 {
		return errors.New("备注长度不能超过 500")
	}
	record.SystemRevenueAmount = CalculateR2SSystemAmount(record.RevenueAmount, record.ExchangeRate)
	record.SystemCostAmount = CalculateR2SSystemAmount(record.CostAmount, record.ExchangeRate)
	record.SystemProfitAmount = roundR2SDecimal(decimal.NewFromFloat(record.SystemRevenueAmount).Sub(decimal.NewFromFloat(record.SystemCostAmount)))
	record.ProfitMargin = CalculateR2SProfitMargin(record.SystemRevenueAmount, record.SystemProfitAmount)
	return nil
}

func (record *R2SRecognitionRecord) Insert() error {
	record.CreatedTime = common.GetTimestamp()
	if err := record.ValidateAndSnapshot(); err != nil {
		return err
	}
	return DB.Create(record).Error
}

func GetR2SSummary(startTime int64, endTime int64) (*R2SSummary, error) {
	summary := &R2SSummary{
		SystemCurrencyCode: operation_setting.GetQuotaDisplayType(),
	}
	recognitionQuery := DB.Model(&R2SRecognitionRecord{})
	if startTime > 0 {
		recognitionQuery = recognitionQuery.Where("period_end >= ?", startTime)
	}
	if endTime > 0 {
		recognitionQuery = recognitionQuery.Where("period_start <= ?", endTime)
	}
	type recognitionSums struct {
		Revenue float64
		Cost    float64
		Profit  float64
	}
	var sums recognitionSums
	if err := recognitionQuery.Select(
		"COALESCE(SUM(system_revenue_amount), 0) AS revenue, COALESCE(SUM(system_cost_amount), 0) AS cost, COALESCE(SUM(system_profit_amount), 0) AS profit",
	).Scan(&sums).Error; err != nil {
		return nil, err
	}
	summary.RecognizedRevenueAmount = roundR2SDecimal(decimal.NewFromFloat(sums.Revenue))
	summary.RecognizedCostAmount = roundR2SDecimal(decimal.NewFromFloat(sums.Cost))
	summary.RecognizedProfitAmount = roundR2SDecimal(decimal.NewFromFloat(sums.Profit))
	summary.ProfitMargin = CalculateR2SProfitMargin(summary.RecognizedRevenueAmount, summary.RecognizedProfitAmount)

	paymentQuery := DB.Model(&R2SPayment{})
	if startTime > 0 {
		paymentQuery = paymentQuery.Where("paid_at >= ?", startTime)
	}
	if endTime > 0 {
		paymentQuery = paymentQuery.Where("paid_at <= ?", endTime)
	}
	if err := paymentQuery.Select("COALESCE(SUM(system_amount), 0)").Scan(&summary.PaymentSystemAmount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&R2SSupplier{}).
		Where("status = ?", R2SStatusActive).
		Select("COALESCE(SUM(system_balance_amount), 0)").
		Scan(&summary.SupplierBalanceAmount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&R2SSupplier{}).Count(&summary.SupplierCount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&R2SSupplier{}).Where("status = ?", R2SStatusActive).Count(&summary.ActiveSupplierCount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&R2SChannelBinding{}).Where("status = ?", R2SStatusActive).Count(&summary.ChannelBindingCount).Error; err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	if err := DB.Model(&R2SSupplier{}).
		Where("status = ?", R2SStatusActive).
		Where("next_balance_reminder_at > 0 AND next_balance_reminder_at <= ?", now).
		Count(&summary.ReminderDueCount).Error; err != nil {
		return nil, err
	}
	return summary, nil
}

func GetR2SPromotionProfitability(campaignId int, startTime int64, endTime int64) ([]R2SPromotionProfitability, error) {
	query := DB.Model(&TopUpPromotionRedemption{}).
		Joins(
			"LEFT JOIN top_up_promotion_campaigns "+
				"ON top_up_promotion_campaigns.id = top_up_promotion_redemptions.campaign_id",
		).
		Where("top_up_promotion_redemptions.status = ?", common.TopUpPromotionRedemptionStatusUsed)
	if campaignId > 0 {
		query = query.Where("top_up_promotion_redemptions.campaign_id = ?", campaignId)
	}
	if startTime > 0 {
		query = query.Where("top_up_promotion_redemptions.used_at >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("top_up_promotion_redemptions.used_at <= ?", endTime)
	}
	var rows []R2SPromotionProfitability
	if err := query.Select(
		"top_up_promotion_redemptions.campaign_id AS campaign_id, " +
			"COALESCE(top_up_promotion_campaigns.name, '') AS campaign_name, " +
			"COUNT(*) AS top_up_count, " +
			"COALESCE(SUM(top_up_promotion_redemptions.original_amount), 0) " +
			"AS gross_revenue_amount, " +
			"COALESCE(SUM(top_up_promotion_redemptions.discount_amount), 0) " +
			"AS discount_amount, " +
			"COALESCE(SUM(top_up_promotion_redemptions.final_payable_amount), 0) " +
			"AS net_revenue_amount, " +
			"top_up_promotion_redemptions.currency_code AS currency_code",
	).Group(
		"top_up_promotion_redemptions.campaign_id, " +
			"top_up_promotion_campaigns.name, " +
			"top_up_promotion_redemptions.currency_code",
	).
		Order("top_up_promotion_redemptions.campaign_id desc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	systemCurrencyCode := operation_setting.GetQuotaDisplayType()
	for i := range rows {
		rows[i].SystemCurrencyCode = systemCurrencyCode
		if rows[i].CurrencyCode == "" {
			rows[i].CurrencyCode = systemCurrencyCode
		}

		var cost float64
		costQuery := DB.Model(&R2SRecognitionRecord{}).Where("promotion_campaign_id = ?", rows[i].CampaignId)
		if startTime > 0 {
			costQuery = costQuery.Where("period_end >= ?", startTime)
		}
		if endTime > 0 {
			costQuery = costQuery.Where("period_start <= ?", endTime)
		}
		if err := costQuery.Select("COALESCE(SUM(system_cost_amount), 0)").Scan(&cost).Error; err != nil {
			return nil, err
		}
		rows[i].RecognizedCostAmount = roundR2SDecimal(decimal.NewFromFloat(cost))
		if NormalizeR2SCurrencyCode(rows[i].CurrencyCode) == NormalizeR2SCurrencyCode(systemCurrencyCode) {
			rows[i].ProfitAmount = roundR2SDecimal(
				decimal.NewFromFloat(rows[i].NetRevenueAmount).
					Sub(decimal.NewFromFloat(rows[i].RecognizedCostAmount)),
			)
			rows[i].ProfitMargin = CalculateR2SProfitMargin(
				rows[i].NetRevenueAmount,
				rows[i].ProfitAmount,
			)
			rows[i].ProfitCalculated = true
		}
	}
	return rows, nil
}

func getR2SOption(key string, fallback string) string {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	if common.OptionMap == nil {
		return fallback
	}
	value, ok := common.OptionMap[key]
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func applyR2SSupplierFilter(query *gorm.DB, filter R2SListFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR name LIKE ? OR description LIKE ?", id, like, like)
		} else {
			query = query.Where("name LIKE ? OR description LIKE ?", like, like)
		}
	}
	return query
}

func fillR2SSupplierViewData(suppliers []*R2SSupplier) error {
	for _, supplier := range suppliers {
		if supplier == nil {
			continue
		}
		if err := DB.Model(&R2SChannelBinding{}).Where("supplier_id = ? AND status = ?", supplier.Id, R2SStatusActive).Count(&supplier.ChannelCount).Error; err != nil {
			return err
		}
		if err := DB.Model(&R2SPayment{}).Where("supplier_id = ?", supplier.Id).Select("COALESCE(MAX(paid_at), 0)").Scan(&supplier.LastPaymentTime).Error; err != nil {
			return err
		}
	}
	return nil
}

func fillR2SChannelBindingViewData(bindings []*R2SChannelBinding) error {
	supplierIds := make([]int, 0, len(bindings))
	for _, binding := range bindings {
		if binding != nil && binding.SupplierId > 0 {
			supplierIds = append(supplierIds, binding.SupplierId)
		}
	}
	if len(supplierIds) == 0 {
		return nil
	}
	var suppliers []R2SSupplier
	if err := DB.Where("id IN ?", supplierIds).Find(&suppliers).Error; err != nil {
		return err
	}
	supplierNames := make(map[int]string, len(suppliers))
	for _, supplier := range suppliers {
		supplierNames[supplier.Id] = supplier.Name
	}
	for _, binding := range bindings {
		if binding != nil {
			binding.SupplierName = supplierNames[binding.SupplierId]
		}
	}
	return nil
}

func ensureR2SChannelBindingUnique(channelId int, currentId int) error {
	var count int64
	query := DB.Model(&R2SChannelBinding{}).Where("channel_id = ? AND status = ?", channelId, R2SStatusActive)
	if currentId > 0 {
		query = query.Where("id <> ?", currentId)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该渠道已绑定到启用中的 R2S 供应商")
	}
	return nil
}

func firstR2SChannelGroup(raw string) string {
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return "default"
}

func isValidR2SPaymentType(paymentType string) bool {
	switch paymentType {
	case R2SPaymentTypePrepaid, R2SPaymentTypePostpaid, R2SPaymentTypeGrant,
		R2SPaymentTypeRefund, R2SPaymentTypeAdjustment:
		return true
	default:
		return false
	}
}

func r2sPaymentBalanceDelta(paymentType string, amount float64) float64 {
	switch paymentType {
	case R2SPaymentTypePrepaid, R2SPaymentTypeGrant, R2SPaymentTypeAdjustment:
		return amount
	case R2SPaymentTypeRefund:
		return -amount
	default:
		return 0
	}
}

func (record *R2SRecognitionRecord) fillBindingSnapshot() error {
	if record.ChannelBindingId > 0 {
		binding, err := GetR2SChannelBindingByID(record.ChannelBindingId)
		if err != nil {
			return err
		}
		record.ChannelId = binding.ChannelId
		record.ChannelNameSnapshot = binding.ChannelNameSnapshot
		record.UpstreamGroupNameSnapshot = binding.UpstreamGroupName
		record.GroupMultiplierSnapshot = binding.GroupMultiplier
		return nil
	}
	if record.ChannelId == 0 {
		return nil
	}
	var binding R2SChannelBinding
	err := DB.Where("channel_id = ? AND supplier_id = ? AND status = ?", record.ChannelId, record.SupplierId, R2SStatusActive).First(&binding).Error
	if err == nil {
		record.ChannelBindingId = binding.Id
		record.ChannelNameSnapshot = binding.ChannelNameSnapshot
		record.UpstreamGroupNameSnapshot = binding.UpstreamGroupName
		record.GroupMultiplierSnapshot = binding.GroupMultiplier
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	channel := &Channel{}
	if err := DB.Where("id = ?", record.ChannelId).First(channel).Error; err != nil {
		return err
	}
	record.ChannelNameSnapshot = channel.Name
	record.UpstreamGroupNameSnapshot = firstR2SChannelGroup(channel.Group)
	if record.GroupMultiplierSnapshot == 0 {
		record.GroupMultiplierSnapshot = 1
	}
	return nil
}

func roundR2SDecimal(value decimal.Decimal) float64 {
	result, _ := value.Round(6).Float64()
	return result
}

func createR2SBalanceUpdateTx(tx *gorm.DB, supplier *R2SSupplier, balanceAfter float64, currencyCode string, exchangeRate float64, updateType string, note string, adminId int) (*R2SBalanceUpdate, error) {
	if supplier == nil {
		return nil, errors.New("供应商不存在")
	}
	now := common.GetTimestamp()
	if updateType == "" {
		updateType = R2SBalanceUpdateTypeManual
	}
	if currencyCode == "" {
		currencyCode = supplier.BalanceCurrencyCode
	}
	if exchangeRate == 0 {
		exchangeRate = supplier.DefaultExchangeRate
	}
	update := &R2SBalanceUpdate{
		SupplierId:           supplier.Id,
		SupplierNameSnapshot: supplier.Name,
		UpdateType:           updateType,
		BalanceBefore:        supplier.BalanceAmount,
		BalanceAfter:         balanceAfter,
		CurrencyCode:         currencyCode,
		ExchangeRate:         exchangeRate,
		ReminderDaysSnapshot: supplier.BalanceReminderDays,
		Note:                 note,
		CreatedByAdminId:     adminId,
		CreatedTime:          now,
	}
	if supplier.BalanceReminderDays > 0 {
		update.NextReminderAt = now + int64(supplier.BalanceReminderDays)*86400
	}
	if err := update.Validate(); err != nil {
		return nil, err
	}
	if err := tx.Create(update).Error; err != nil {
		return nil, err
	}
	supplier.BalanceAmount = update.BalanceAfter
	supplier.BalanceCurrencyCode = update.CurrencyCode
	supplier.DefaultExchangeRate = update.ExchangeRate
	supplier.SystemBalanceAmount = CalculateR2SSystemAmount(update.BalanceAfter, update.ExchangeRate)
	supplier.BalanceUpdatedTime = now
	supplier.NextBalanceReminderAt = update.NextReminderAt
	supplier.UpdatedTime = now
	if err := tx.Save(supplier).Error; err != nil {
		return nil, err
	}
	return update, nil
}

func CreateR2SPaymentWithBalance(payment *R2SPayment, balanceAfter *float64, adminId int) (*R2SPayment, *R2SBalanceUpdate, error) {
	var balanceUpdate *R2SBalanceUpdate
	err := DB.Transaction(func(tx *gorm.DB) error {
		supplier := &R2SSupplier{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", payment.SupplierId).First(supplier).Error; err != nil {
			return err
		}
		payment.SupplierNameSnapshot = supplier.Name
		if payment.CurrencyCode == "" {
			payment.CurrencyCode = supplier.DefaultCurrencyCode
		}
		if payment.ExchangeRate == 0 {
			payment.ExchangeRate = supplier.DefaultExchangeRate
		}
		payment.ReceiptRequired = GetR2SSettings().ReceiptRequired
		payment.BalanceBefore = supplier.BalanceAmount
		targetBalance := supplier.BalanceAmount + r2sPaymentBalanceDelta(payment.PaymentType, payment.Amount)
		if balanceAfter != nil {
			targetBalance = *balanceAfter
		}
		payment.BalanceAfter = targetBalance
		payment.CreatedByAdminId = adminId
		payment.CreatedTime = common.GetTimestamp()
		if err := payment.Validate(); err != nil {
			return err
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}
		if targetBalance != supplier.BalanceAmount || payment.CurrencyCode != supplier.BalanceCurrencyCode {
			update, err := createR2SBalanceUpdateTx(
				tx,
				supplier,
				targetBalance,
				payment.CurrencyCode,
				payment.ExchangeRate,
				payment.PaymentType,
				fmt.Sprintf("付款记录 #%d 自动更新余额", payment.Id),
				adminId,
			)
			if err != nil {
				return err
			}
			balanceUpdate = update
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return payment, balanceUpdate, nil
}

func CreateR2SManualBalanceUpdate(supplierId int, balanceAfter float64, currencyCode string, exchangeRate float64, reminderDays *int, note string, adminId int) (*R2SBalanceUpdate, error) {
	var update *R2SBalanceUpdate
	err := DB.Transaction(func(tx *gorm.DB) error {
		supplier := &R2SSupplier{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", supplierId).First(supplier).Error; err != nil {
			return err
		}
		if reminderDays != nil {
			if *reminderDays < 0 {
				return errors.New("余额提醒天数不能小于 0")
			}
			supplier.BalanceReminderDays = *reminderDays
		}
		var err error
		update, err = createR2SBalanceUpdateTx(
			tx,
			supplier,
			balanceAfter,
			currencyCode,
			exchangeRate,
			R2SBalanceUpdateTypeManual,
			note,
			adminId,
		)
		return err
	})
	return update, err
}
