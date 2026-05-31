package model

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const promotionCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type TopUpPromotionRule struct {
	MinAmount     float64 `json:"min_amount"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
}

type TopUpPromotionCampaign struct {
	Id                    int                  `json:"id"`
	Name                  string               `json:"name" gorm:"type:varchar(80);index"`
	Description           string               `json:"description" gorm:"type:varchar(255)"`
	CurrencyCode          string               `json:"currency_code" gorm:"type:varchar(16);index"`
	DiscountRules         string               `json:"-" gorm:"type:text"`
	Rules                 []TopUpPromotionRule `json:"rules" gorm:"-"`
	AllowedGroups         string               `json:"-" gorm:"type:text"`
	AllowedGroupNames     []string             `json:"allowed_groups" gorm:"-"`
	Status                string               `json:"status" gorm:"type:varchar(16);index"`
	EffectiveStatus       string               `json:"effective_status" gorm:"-"`
	ValidFrom             int64                `json:"valid_from" gorm:"index"`
	ExpiresAt             int64                `json:"expires_at" gorm:"index"`
	MaxRedemptions        int                  `json:"max_redemptions"`
	MaxRedemptionsPerUser int                  `json:"max_redemptions_per_user"`
	CreatedByAdminId      int                  `json:"created_by_admin_id" gorm:"index"`
	CreatedTime           int64                `json:"created_time"`
	UpdatedTime           int64                `json:"updated_time"`
	RevokedAt             int64                `json:"revoked_at"`
	RevokedByAdminId      int                  `json:"revoked_by_admin_id" gorm:"index"`
	RevokeReason          string               `json:"revoke_reason" gorm:"type:varchar(255)"`
	CodeCount             int64                `json:"code_count" gorm:"-"`
	ReservedCount         int64                `json:"reserved_count" gorm:"-"`
	UsedCount             int64                `json:"used_count" gorm:"-"`
}

type TopUpPromotionCode struct {
	Id                    int    `json:"id"`
	CampaignId            int    `json:"campaign_id" gorm:"index"`
	CampaignName          string `json:"campaign_name" gorm:"-"`
	Code                  string `json:"code" gorm:"type:varchar(64);uniqueIndex"`
	Status                string `json:"status" gorm:"type:varchar(16);index"`
	EffectiveStatus       string `json:"effective_status" gorm:"-"`
	ValidFrom             int64  `json:"valid_from" gorm:"index"`
	ExpiresAt             int64  `json:"expires_at" gorm:"index"`
	MaxRedemptions        int    `json:"max_redemptions"`
	MaxRedemptionsPerUser int    `json:"max_redemptions_per_user"`
	CreatedByAdminId      int    `json:"created_by_admin_id" gorm:"index"`
	CreatedTime           int64  `json:"created_time"`
	UpdatedTime           int64  `json:"updated_time"`
	RevokedAt             int64  `json:"revoked_at"`
	RevokedByAdminId      int    `json:"revoked_by_admin_id" gorm:"index"`
	RevokeReason          string `json:"revoke_reason" gorm:"type:varchar(255)"`
	ReservedCount         int64  `json:"reserved_count" gorm:"-"`
	UsedCount             int64  `json:"used_count" gorm:"-"`
}

type TopUpPromotionRedemption struct {
	Id                 int     `json:"id"`
	CampaignId         int     `json:"campaign_id" gorm:"index"`
	CampaignName       string  `json:"campaign_name" gorm:"-"`
	CodeId             int     `json:"code_id" gorm:"index"`
	Code               string  `json:"code" gorm:"type:varchar(64);index"`
	UserId             int     `json:"user_id" gorm:"index"`
	Username           string  `json:"username" gorm:"-"`
	TopUpId            int     `json:"top_up_id" gorm:"index"`
	TradeNo            string  `json:"trade_no" gorm:"type:varchar(255);index"`
	PaymentMethod      string  `json:"payment_method" gorm:"type:varchar(50);index"`
	Amount             int64   `json:"amount"`
	OriginalAmount     float64 `json:"original_amount"`
	DiscountAmount     float64 `json:"discount_amount"`
	FinalPayableAmount float64 `json:"final_payable_amount"`
	CurrencyCode       string  `json:"currency_code" gorm:"type:varchar(16);index"`
	Status             string  `json:"status" gorm:"type:varchar(16);index"`
	ReservedAt         int64   `json:"reserved_at"`
	UsedAt             int64   `json:"used_at"`
	ExpiredAt          int64   `json:"expired_at"`
	CreatedTime        int64   `json:"created_time"`
	UpdatedTime        int64   `json:"updated_time"`
}

type TopUpPromotionFilter struct {
	Keyword string
	Status  string
}

type TopUpPromotionCodeFilter struct {
	CampaignId int
	Keyword    string
	Status     string
}

type TopUpPromotionRedemptionFilter struct {
	CampaignId int
	CodeId     int
	UserId     int
	Keyword    string
	Status     string
}

type TopUpPromotionPreview struct {
	CampaignId     int
	CampaignName   string
	CodeId         int
	Code           string
	CurrencyCode   string
	DiscountAmount decimal.Decimal
	MatchedRule    TopUpPromotionRule
}

func NormalizeTopUpPromotionCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func isValidTopUpPromotionCode(code string) bool {
	if len(code) < 3 || len(code) > 64 {
		return false
	}
	for _, r := range code {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func normalizePromotionRules(rules []TopUpPromotionRule) []TopUpPromotionRule {
	normalized := make([]TopUpPromotionRule, 0, len(rules))
	for _, rule := range rules {
		rule.DiscountType = strings.ToLower(strings.TrimSpace(rule.DiscountType))
		normalized = append(normalized, rule)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].MinAmount < normalized[j].MinAmount
	})
	return normalized
}

func encodePromotionRules(rules []TopUpPromotionRule) (string, error) {
	rules = normalizePromotionRules(rules)
	bytes, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func decodePromotionRules(raw string) []TopUpPromotionRule {
	var rules []TopUpPromotionRule
	if strings.TrimSpace(raw) == "" {
		return rules
	}
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return []TopUpPromotionRule{}
	}
	return normalizePromotionRules(rules)
}

func encodePromotionGroups(groups []string) (string, []string, error) {
	seen := make(map[string]struct{}, len(groups))
	normalized := make([]string, 0, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		normalized = append(normalized, group)
	}
	sort.Strings(normalized)
	bytes, err := json.Marshal(normalized)
	if err != nil {
		return "", nil, err
	}
	return string(bytes), normalized, nil
}

func decodePromotionGroups(raw string) []string {
	var groups []string
	if strings.TrimSpace(raw) == "" {
		return groups
	}
	if err := json.Unmarshal([]byte(raw), &groups); err != nil {
		return []string{}
	}
	_, normalized, err := encodePromotionGroups(groups)
	if err != nil {
		return []string{}
	}
	return normalized
}

func (campaign *TopUpPromotionCampaign) GetRules() []TopUpPromotionRule {
	if campaign == nil {
		return nil
	}
	if len(campaign.Rules) > 0 {
		return normalizePromotionRules(campaign.Rules)
	}
	return decodePromotionRules(campaign.DiscountRules)
}

func (campaign *TopUpPromotionCampaign) SetRules(rules []TopUpPromotionRule) error {
	encoded, err := encodePromotionRules(rules)
	if err != nil {
		return err
	}
	campaign.Rules = normalizePromotionRules(rules)
	campaign.DiscountRules = encoded
	return nil
}

func (campaign *TopUpPromotionCampaign) GetAllowedGroups() []string {
	if campaign == nil {
		return nil
	}
	if len(campaign.AllowedGroupNames) > 0 {
		return campaign.AllowedGroupNames
	}
	return decodePromotionGroups(campaign.AllowedGroups)
}

func (campaign *TopUpPromotionCampaign) SetAllowedGroups(groups []string) error {
	encoded, normalized, err := encodePromotionGroups(groups)
	if err != nil {
		return err
	}
	campaign.AllowedGroups = encoded
	campaign.AllowedGroupNames = normalized
	return nil
}

func (campaign *TopUpPromotionCampaign) IsGroupAllowed(group string) bool {
	allowedGroups := campaign.GetAllowedGroups()
	if len(allowedGroups) == 0 {
		return true
	}
	for _, allowedGroup := range allowedGroups {
		if allowedGroup == group {
			return true
		}
	}
	return false
}

func (campaign *TopUpPromotionCampaign) GetEffectiveStatus() string {
	if campaign == nil {
		return ""
	}
	if campaign.EffectiveStatus != "" {
		return campaign.EffectiveStatus
	}
	return getTimedPromotionStatus(campaign.Status, campaign.ValidFrom, campaign.ExpiresAt, common.GetTimestamp())
}

func (code *TopUpPromotionCode) GetEffectiveStatus() string {
	if code == nil {
		return ""
	}
	if code.EffectiveStatus != "" {
		return code.EffectiveStatus
	}
	return getTimedPromotionStatus(code.Status, code.ValidFrom, code.ExpiresAt, common.GetTimestamp())
}

func getTimedPromotionStatus(status string, validFrom, expiresAt, now int64) string {
	switch status {
	case common.TopUpPromotionStatusRevoked, common.TopUpPromotionStatusPaused:
		return status
	}
	if validFrom > now {
		return common.TopUpPromotionStatusScheduled
	}
	if expiresAt != 0 && expiresAt < now {
		return common.TopUpPromotionStatusExpired
	}
	if status == "" {
		return common.TopUpPromotionStatusActive
	}
	return status
}

func CalculateTopUpPromotionDiscount(rules []TopUpPromotionRule, basePayable decimal.Decimal) (decimal.Decimal, TopUpPromotionRule, bool) {
	rules = normalizePromotionRules(rules)
	if len(rules) == 0 || !basePayable.GreaterThan(decimal.Zero) {
		return decimal.Zero, TopUpPromotionRule{}, false
	}

	selectedIndex := -1
	for i := range rules {
		minAmount := decimal.NewFromFloat(rules[i].MinAmount)
		if basePayable.GreaterThanOrEqual(minAmount) {
			selectedIndex = i
		}
	}
	if selectedIndex == -1 {
		return decimal.Zero, TopUpPromotionRule{}, false
	}

	rule := rules[selectedIndex]
	var discount decimal.Decimal
	switch rule.DiscountType {
	case common.TopUpPromotionDiscountTypePercent:
		discount = basePayable.Mul(decimal.NewFromFloat(rule.DiscountValue)).Div(decimal.NewFromInt(100))
	case common.TopUpPromotionDiscountTypeFixed:
		discount = decimal.NewFromFloat(rule.DiscountValue)
	default:
		return decimal.Zero, TopUpPromotionRule{}, false
	}
	if discount.LessThan(decimal.Zero) {
		return decimal.Zero, TopUpPromotionRule{}, false
	}
	if discount.GreaterThan(basePayable) {
		discount = basePayable
	}
	discount = discount.Round(2)
	return discount, rule, discount.GreaterThan(decimal.Zero)
}

func (campaign *TopUpPromotionCampaign) Validate() error {
	if campaign == nil {
		return errors.New("促销活动不存在")
	}
	name := strings.TrimSpace(campaign.Name)
	if name == "" {
		return errors.New("促销活动名称不能为空")
	}
	if utf8.RuneCountInString(name) > 80 {
		return errors.New("促销活动名称长度不能超过 80")
	}
	campaign.Name = name
	campaign.Description = strings.TrimSpace(campaign.Description)
	if utf8.RuneCountInString(campaign.Description) > 255 {
		return errors.New("促销活动说明长度不能超过 255")
	}
	campaign.CurrencyCode = NormalizeTopUpCouponCurrencyCode(campaign.CurrencyCode)
	if campaign.CurrencyCode == "" {
		return errors.New("请选择促销货币")
	}
	if !IsValidTopUpCouponCurrencyCode(campaign.CurrencyCode) {
		return errors.New("促销货币格式不正确")
	}
	if campaign.ValidFrom == 0 {
		campaign.ValidFrom = common.GetTimestamp()
	}
	if campaign.ExpiresAt != 0 && campaign.ExpiresAt <= campaign.ValidFrom {
		return errors.New("过期时间必须晚于生效时间")
	}
	if campaign.MaxRedemptions < 0 || campaign.MaxRedemptionsPerUser < 0 {
		return errors.New("兑换次数限制不能小于 0")
	}
	rules := campaign.GetRules()
	if len(rules) == 0 {
		return errors.New("请至少配置一档促销规则")
	}
	for _, rule := range rules {
		if rule.MinAmount < 0 {
			return errors.New("促销门槛不能小于 0")
		}
		switch rule.DiscountType {
		case common.TopUpPromotionDiscountTypeFixed:
			if rule.DiscountValue <= 0 {
				return errors.New("固定优惠金额必须大于 0")
			}
		case common.TopUpPromotionDiscountTypePercent:
			if rule.DiscountValue <= 0 || rule.DiscountValue > 100 {
				return errors.New("百分比优惠必须在 0-100 之间")
			}
		default:
			return errors.New("促销类型必须为固定金额或百分比")
		}
	}
	if err := campaign.SetRules(rules); err != nil {
		return err
	}
	if err := campaign.SetAllowedGroups(campaign.GetAllowedGroups()); err != nil {
		return err
	}
	if campaign.Status == "" {
		campaign.Status = common.TopUpPromotionStatusActive
	}
	return nil
}

func (code *TopUpPromotionCode) Validate() error {
	if code == nil {
		return errors.New("促销码不存在")
	}
	code.Code = NormalizeTopUpPromotionCode(code.Code)
	if !isValidTopUpPromotionCode(code.Code) {
		return errors.New("促销码仅支持 3-64 位大写字母、数字、下划线或短横线")
	}
	if code.CampaignId == 0 {
		return errors.New("缺少促销活动")
	}
	if code.ValidFrom == 0 {
		code.ValidFrom = common.GetTimestamp()
	}
	if code.ExpiresAt != 0 && code.ExpiresAt <= code.ValidFrom {
		return errors.New("促销码过期时间必须晚于生效时间")
	}
	if code.MaxRedemptions < 0 || code.MaxRedemptionsPerUser < 0 {
		return errors.New("兑换次数限制不能小于 0")
	}
	if code.Status == "" {
		code.Status = common.TopUpPromotionStatusActive
	}
	return nil
}

func (campaign *TopUpPromotionCampaign) Insert() error {
	now := common.GetTimestamp()
	if campaign.CreatedTime == 0 {
		campaign.CreatedTime = now
	}
	campaign.UpdatedTime = now
	if err := campaign.Validate(); err != nil {
		return err
	}
	return DB.Create(campaign).Error
}

func (campaign *TopUpPromotionCampaign) Update() error {
	campaign.UpdatedTime = common.GetTimestamp()
	if err := campaign.Validate(); err != nil {
		return err
	}
	return DB.Save(campaign).Error
}

func (code *TopUpPromotionCode) Insert() error {
	now := common.GetTimestamp()
	if code.CreatedTime == 0 {
		code.CreatedTime = now
	}
	code.UpdatedTime = now
	if err := code.Validate(); err != nil {
		return err
	}
	return DB.Create(code).Error
}

func GetTopUpPromotionCampaignById(id int) (*TopUpPromotionCampaign, error) {
	if id == 0 {
		return nil, errors.New("缺少促销活动 ID")
	}
	campaign := &TopUpPromotionCampaign{}
	if err := DB.Where("id = ?", id).First(campaign).Error; err != nil {
		return nil, err
	}
	if err := fillPromotionCampaignViewData([]*TopUpPromotionCampaign{campaign}); err != nil {
		return nil, err
	}
	return campaign, nil
}

func GetAllTopUpPromotionCampaigns(pageInfo *common.PageInfo, filter TopUpPromotionFilter) ([]*TopUpPromotionCampaign, int64, error) {
	query := applyTopUpPromotionFilter(DB.Model(&TopUpPromotionCampaign{}), filter)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var campaigns []*TopUpPromotionCampaign
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}
	if err := fillPromotionCampaignViewData(campaigns); err != nil {
		return nil, 0, err
	}
	return campaigns, total, nil
}

func GetTopUpPromotionCodes(pageInfo *common.PageInfo, filter TopUpPromotionCodeFilter) ([]*TopUpPromotionCode, int64, error) {
	query := DB.Model(&TopUpPromotionCode{})
	if filter.CampaignId > 0 {
		query = query.Where("campaign_id = ?", filter.CampaignId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + NormalizeTopUpPromotionCode(keyword) + "%"
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR campaign_id = ? OR code LIKE ?", id, id, like)
		} else {
			query = query.Where("code LIKE ?", like)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var codes []*TopUpPromotionCode
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&codes).Error; err != nil {
		return nil, 0, err
	}
	if err := fillPromotionCodeViewData(codes); err != nil {
		return nil, 0, err
	}
	return codes, total, nil
}

func GetTopUpPromotionRedemptions(pageInfo *common.PageInfo, filter TopUpPromotionRedemptionFilter) ([]*TopUpPromotionRedemption, int64, error) {
	query := DB.Model(&TopUpPromotionRedemption{})
	if filter.CampaignId > 0 {
		query = query.Where("campaign_id = ?", filter.CampaignId)
	}
	if filter.CodeId > 0 {
		query = query.Where("code_id = ?", filter.CodeId)
	}
	if filter.UserId > 0 {
		query = query.Where("user_id = ?", filter.UserId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		normalizedCodeLike := "%" + NormalizeTopUpPromotionCode(keyword) + "%"
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR user_id = ? OR top_up_id = ? OR code LIKE ? OR trade_no LIKE ?", id, id, id, normalizedCodeLike, like)
		} else {
			query = query.Where("code LIKE ? OR trade_no LIKE ?", normalizedCodeLike, like)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var redemptions []*TopUpPromotionRedemption
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&redemptions).Error; err != nil {
		return nil, 0, err
	}
	if err := fillPromotionRedemptionViewData(redemptions); err != nil {
		return nil, 0, err
	}
	return redemptions, total, nil
}

func applyTopUpPromotionFilter(query *gorm.DB, filter TopUpPromotionFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword == "" {
		return query
	}
	like := "%" + keyword + "%"
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, like)
	} else {
		query = query.Where("name LIKE ?", like)
	}
	return query
}

func fillPromotionCampaignViewData(campaigns []*TopUpPromotionCampaign) error {
	if len(campaigns) == 0 {
		return nil
	}
	now := common.GetTimestamp()
	statuses := []string{common.TopUpPromotionRedemptionStatusReserved, common.TopUpPromotionRedemptionStatusUsed}
	for _, campaign := range campaigns {
		if campaign == nil {
			continue
		}
		campaign.Rules = campaign.GetRules()
		campaign.AllowedGroupNames = campaign.GetAllowedGroups()
		campaign.EffectiveStatus = getTimedPromotionStatus(campaign.Status, campaign.ValidFrom, campaign.ExpiresAt, now)
		_ = DB.Model(&TopUpPromotionCode{}).Where("campaign_id = ?", campaign.Id).Count(&campaign.CodeCount).Error
		_ = DB.Model(&TopUpPromotionRedemption{}).Where("campaign_id = ? AND status = ?", campaign.Id, common.TopUpPromotionRedemptionStatusReserved).Count(&campaign.ReservedCount).Error
		_ = DB.Model(&TopUpPromotionRedemption{}).Where("campaign_id = ? AND status IN ?", campaign.Id, statuses[1:]).Count(&campaign.UsedCount).Error
	}
	return nil
}

func fillPromotionCodeViewData(codes []*TopUpPromotionCode) error {
	if len(codes) == 0 {
		return nil
	}
	now := common.GetTimestamp()
	campaignIds := make([]int, 0, len(codes))
	seen := map[int]struct{}{}
	for _, code := range codes {
		if code == nil {
			continue
		}
		code.EffectiveStatus = getTimedPromotionStatus(code.Status, code.ValidFrom, code.ExpiresAt, now)
		_ = DB.Model(&TopUpPromotionRedemption{}).Where("code_id = ? AND status = ?", code.Id, common.TopUpPromotionRedemptionStatusReserved).Count(&code.ReservedCount).Error
		_ = DB.Model(&TopUpPromotionRedemption{}).Where("code_id = ? AND status = ?", code.Id, common.TopUpPromotionRedemptionStatusUsed).Count(&code.UsedCount).Error
		if code.CampaignId != 0 {
			if _, ok := seen[code.CampaignId]; !ok {
				seen[code.CampaignId] = struct{}{}
				campaignIds = append(campaignIds, code.CampaignId)
			}
		}
	}
	if len(campaignIds) == 0 {
		return nil
	}
	type campaignNameRow struct {
		Id   int
		Name string
	}
	var campaigns []campaignNameRow
	if err := DB.Model(&TopUpPromotionCampaign{}).Select("id, name").Where("id IN ?", campaignIds).Find(&campaigns).Error; err != nil {
		return err
	}
	nameMap := make(map[int]string, len(campaigns))
	for _, campaign := range campaigns {
		nameMap[campaign.Id] = campaign.Name
	}
	for _, code := range codes {
		code.CampaignName = nameMap[code.CampaignId]
	}
	return nil
}

func fillPromotionRedemptionViewData(redemptions []*TopUpPromotionRedemption) error {
	if len(redemptions) == 0 {
		return nil
	}
	campaignIds := make([]int, 0, len(redemptions))
	userIds := make([]int, 0, len(redemptions))
	seenCampaigns := map[int]struct{}{}
	seenUsers := map[int]struct{}{}
	for _, redemption := range redemptions {
		if redemption == nil {
			continue
		}
		if redemption.CampaignId != 0 {
			if _, ok := seenCampaigns[redemption.CampaignId]; !ok {
				seenCampaigns[redemption.CampaignId] = struct{}{}
				campaignIds = append(campaignIds, redemption.CampaignId)
			}
		}
		if redemption.UserId != 0 {
			if _, ok := seenUsers[redemption.UserId]; !ok {
				seenUsers[redemption.UserId] = struct{}{}
				userIds = append(userIds, redemption.UserId)
			}
		}
	}
	campaignNames := map[int]string{}
	if len(campaignIds) > 0 {
		type campaignNameRow struct {
			Id   int
			Name string
		}
		var campaigns []campaignNameRow
		if err := DB.Model(&TopUpPromotionCampaign{}).Select("id, name").Where("id IN ?", campaignIds).Find(&campaigns).Error; err != nil {
			return err
		}
		for _, campaign := range campaigns {
			campaignNames[campaign.Id] = campaign.Name
		}
	}
	userNames := map[int]string{}
	if len(userIds) > 0 {
		type userNameRow struct {
			Id       int
			Username string
		}
		var users []userNameRow
		if err := DB.Model(&User{}).Select("id, username").Where("id IN ?", userIds).Find(&users).Error; err != nil {
			return err
		}
		for _, user := range users {
			userNames[user.Id] = user.Username
		}
	}
	for _, redemption := range redemptions {
		redemption.CampaignName = campaignNames[redemption.CampaignId]
		redemption.Username = userNames[redemption.UserId]
	}
	return nil
}

func GenerateTopUpPromotionCode(prefix string) (string, error) {
	prefix = NormalizeTopUpPromotionCode(prefix)
	builder := strings.Builder{}
	if prefix != "" {
		builder.WriteString(prefix)
		builder.WriteString("-")
	}
	for i := 0; i < 10; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(promotionCodeAlphabet))))
		if err != nil {
			return "", err
		}
		builder.WriteByte(promotionCodeAlphabet[index.Int64()])
	}
	return builder.String(), nil
}

func CreateTopUpPromotionCodes(campaignId int, adminId int, rawCodes []string, autoCount int, prefix string, validFrom, expiresAt int64, maxRedemptions, maxRedemptionsPerUser int) ([]*TopUpPromotionCode, error) {
	if campaignId == 0 {
		return nil, errors.New("缺少促销活动 ID")
	}
	if autoCount < 0 || autoCount > 500 {
		return nil, errors.New("自动生成数量必须在 0-500 之间")
	}

	now := common.GetTimestamp()
	codes := make([]string, 0, len(rawCodes)+autoCount)
	seen := map[string]struct{}{}
	for _, rawCode := range rawCodes {
		code := NormalizeTopUpPromotionCode(rawCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	targetCount := len(codes) + autoCount
	for len(codes) < targetCount {
		code, err := GenerateTopUpPromotionCode(prefix)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[code]; ok {
			continue
		}
		var count int64
		if err := DB.Model(&TopUpPromotionCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, errors.New("请至少提供一个促销码")
	}

	createdCodes := make([]*TopUpPromotionCode, 0, len(codes))
	err := DB.Transaction(func(tx *gorm.DB) error {
		campaign := &TopUpPromotionCampaign{}
		if err := tx.Where("id = ?", campaignId).First(campaign).Error; err != nil {
			return err
		}
		for _, codeText := range codes {
			code := &TopUpPromotionCode{
				CampaignId:            campaignId,
				Code:                  codeText,
				Status:                common.TopUpPromotionStatusActive,
				ValidFrom:             validFrom,
				ExpiresAt:             expiresAt,
				MaxRedemptions:        maxRedemptions,
				MaxRedemptionsPerUser: maxRedemptionsPerUser,
				CreatedByAdminId:      adminId,
				CreatedTime:           now,
				UpdatedTime:           now,
			}
			if code.ValidFrom == 0 {
				code.ValidFrom = campaign.ValidFrom
			}
			if code.ExpiresAt == 0 {
				code.ExpiresAt = campaign.ExpiresAt
			}
			if err := code.Validate(); err != nil {
				return err
			}
			if err := tx.Create(code).Error; err != nil {
				return err
			}
			createdCodes = append(createdCodes, code)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := fillPromotionCodeViewData(createdCodes); err != nil {
		return nil, err
	}
	return createdCodes, nil
}

func RevokeTopUpPromotionCampaign(id int, adminId int, reason string) (*TopUpPromotionCampaign, error) {
	var result *TopUpPromotionCampaign
	err := DB.Transaction(func(tx *gorm.DB) error {
		campaign := &TopUpPromotionCampaign{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(campaign).Error; err != nil {
			return err
		}
		if campaign.Status == common.TopUpPromotionStatusRevoked {
			result = campaign
			return nil
		}
		now := common.GetTimestamp()
		campaign.Status = common.TopUpPromotionStatusRevoked
		campaign.RevokedAt = now
		campaign.RevokedByAdminId = adminId
		campaign.RevokeReason = strings.TrimSpace(reason)
		campaign.UpdatedTime = now
		if err := tx.Save(campaign).Error; err != nil {
			return err
		}
		result = campaign
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := fillPromotionCampaignViewData([]*TopUpPromotionCampaign{result}); err != nil {
		return nil, err
	}
	return result, nil
}

func RevokeTopUpPromotionCode(id int, adminId int, reason string) (*TopUpPromotionCode, error) {
	var result *TopUpPromotionCode
	err := DB.Transaction(func(tx *gorm.DB) error {
		code := &TopUpPromotionCode{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(code).Error; err != nil {
			return err
		}
		if code.Status == common.TopUpPromotionStatusRevoked {
			result = code
			return nil
		}
		now := common.GetTimestamp()
		code.Status = common.TopUpPromotionStatusRevoked
		code.RevokedAt = now
		code.RevokedByAdminId = adminId
		code.RevokeReason = strings.TrimSpace(reason)
		code.UpdatedTime = now
		if err := tx.Save(code).Error; err != nil {
			return err
		}
		result = code
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := fillPromotionCodeViewData([]*TopUpPromotionCode{result}); err != nil {
		return nil, err
	}
	return result, nil
}

func GetTopUpPromotionPreview(rawCode string, user *User, currencyCode string, originalPayable, minThreshold decimal.Decimal) (*TopUpPromotionPreview, error) {
	codeText := NormalizeTopUpPromotionCode(rawCode)
	if codeText == "" {
		return nil, nil
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	var preview *TopUpPromotionPreview
	err := DB.Transaction(func(tx *gorm.DB) error {
		code, campaign, err := findPromotionCodeAndCampaignTx(tx, codeText, 0)
		if err != nil {
			return err
		}
		if err := ensurePromotionUsableTx(tx, campaign, code, user, currencyCode); err != nil {
			return err
		}
		discount, rule, ok := CalculateTopUpPromotionDiscount(campaign.GetRules(), originalPayable)
		if !ok {
			return errors.New("未满足促销活动使用门槛")
		}
		finalPayable := originalPayable.Sub(discount)
		if finalPayable.LessThanOrEqual(minThreshold) {
			return errors.New("使用促销码后金额必须高于最低充值金额")
		}
		preview = &TopUpPromotionPreview{
			CampaignId:     campaign.Id,
			CampaignName:   campaign.Name,
			CodeId:         code.Id,
			Code:           code.Code,
			CurrencyCode:   campaign.CurrencyCode,
			DiscountAmount: discount,
			MatchedRule:    rule,
		}
		return nil
	})
	return preview, err
}

func findPromotionCodeAndCampaignTx(tx *gorm.DB, codeText string, codeId int) (*TopUpPromotionCode, *TopUpPromotionCampaign, error) {
	code := &TopUpPromotionCode{}
	query := tx.Set("gorm:query_option", "FOR UPDATE")
	if codeId > 0 {
		query = query.Where("id = ?", codeId)
	} else {
		query = query.Where("code = ?", codeText)
	}
	if err := query.First(code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("促销码不存在或已停用")
		}
		return nil, nil, err
	}
	if codeText != "" && code.Code != NormalizeTopUpPromotionCode(codeText) {
		return nil, nil, errors.New("促销码不存在或已停用")
	}
	campaign := &TopUpPromotionCampaign{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", code.CampaignId).First(campaign).Error; err != nil {
		return nil, nil, err
	}
	campaign.Rules = campaign.GetRules()
	campaign.AllowedGroupNames = campaign.GetAllowedGroups()
	return code, campaign, nil
}

func ensurePromotionUsableTx(tx *gorm.DB, campaign *TopUpPromotionCampaign, code *TopUpPromotionCode, user *User, currencyCode string) error {
	now := common.GetTimestamp()
	if campaign == nil || code == nil {
		return errors.New("促销码不存在或已停用")
	}
	if campaign.GetEffectiveStatus() == common.TopUpPromotionStatusScheduled || campaign.ValidFrom > now {
		return errors.New("促销活动尚未生效")
	}
	if campaign.GetEffectiveStatus() != common.TopUpPromotionStatusActive {
		return errors.New("促销活动不可用")
	}
	if code.GetEffectiveStatus() == common.TopUpPromotionStatusScheduled || code.ValidFrom > now {
		return errors.New("促销码尚未生效")
	}
	if code.GetEffectiveStatus() != common.TopUpPromotionStatusActive {
		return errors.New("促销码不可用")
	}
	if !campaign.IsGroupAllowed(user.Group) {
		return errors.New("当前用户分组不可使用该促销码")
	}
	normalizedCurrencyCode := NormalizeTopUpCouponCurrencyCode(currencyCode)
	if campaign.CurrencyCode != "" && normalizedCurrencyCode != "" && campaign.CurrencyCode != normalizedCurrencyCode {
		return errors.New("促销码币种与当前支付方式不匹配")
	}

	statuses := []string{common.TopUpPromotionRedemptionStatusReserved, common.TopUpPromotionRedemptionStatusUsed}
	if campaign.MaxRedemptions > 0 {
		count, err := countPromotionRedemptionsTx(tx, campaign.Id, 0, 0, statuses)
		if err != nil {
			return err
		}
		if count >= int64(campaign.MaxRedemptions) {
			return errors.New("促销活动兑换次数已用完")
		}
	}
	if campaign.MaxRedemptionsPerUser > 0 {
		count, err := countPromotionRedemptionsTx(tx, campaign.Id, 0, user.Id, statuses)
		if err != nil {
			return err
		}
		if count >= int64(campaign.MaxRedemptionsPerUser) {
			return errors.New("当前用户已达到促销活动兑换次数限制")
		}
	}
	if code.MaxRedemptions > 0 {
		count, err := countPromotionRedemptionsTx(tx, 0, code.Id, 0, statuses)
		if err != nil {
			return err
		}
		if count >= int64(code.MaxRedemptions) {
			return errors.New("促销码兑换次数已用完")
		}
	}
	if code.MaxRedemptionsPerUser > 0 {
		count, err := countPromotionRedemptionsTx(tx, 0, code.Id, user.Id, statuses)
		if err != nil {
			return err
		}
		if count >= int64(code.MaxRedemptionsPerUser) {
			return errors.New("当前用户已达到促销码兑换次数限制")
		}
	}
	return nil
}

func countPromotionRedemptionsTx(tx *gorm.DB, campaignId, codeId, userId int, statuses []string) (int64, error) {
	query := tx.Model(&TopUpPromotionRedemption{})
	if campaignId > 0 {
		query = query.Where("campaign_id = ?", campaignId)
	}
	if codeId > 0 {
		query = query.Where("code_id = ?", codeId)
	}
	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func ReserveTopUpPromotionCodeTx(tx *gorm.DB, topUp *TopUp, currencyCode string) (*TopUpPromotionRedemption, error) {
	if topUp == nil || topUp.PromotionCodeId == 0 {
		return nil, nil
	}
	user := &User{}
	if err := tx.Where("id = ?", topUp.UserId).First(user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	code, campaign, err := findPromotionCodeAndCampaignTx(tx, topUp.PromotionCode, topUp.PromotionCodeId)
	if err != nil {
		return nil, err
	}
	if err := ensurePromotionUsableTx(tx, campaign, code, user, currencyCode); err != nil {
		return nil, err
	}
	discount, _, ok := CalculateTopUpPromotionDiscount(campaign.GetRules(), decimal.NewFromFloat(topUp.OriginalMoney))
	if !ok {
		return nil, errors.New("未满足促销活动使用门槛")
	}
	if !discount.Equal(decimal.NewFromFloat(topUp.PromotionDiscount).Round(2)) {
		return nil, errors.New("促销活动规则已变化，请重新确认订单")
	}
	now := common.GetTimestamp()
	redemption := &TopUpPromotionRedemption{
		CampaignId:         campaign.Id,
		CodeId:             code.Id,
		Code:               code.Code,
		UserId:             topUp.UserId,
		TopUpId:            topUp.Id,
		TradeNo:            topUp.TradeNo,
		PaymentMethod:      topUp.PaymentMethod,
		Amount:             topUp.Amount,
		OriginalAmount:     topUp.OriginalMoney,
		DiscountAmount:     topUp.PromotionDiscount,
		FinalPayableAmount: topUp.PayMoney,
		CurrencyCode:       NormalizeTopUpCouponCurrencyCode(currencyCode),
		Status:             common.TopUpPromotionRedemptionStatusReserved,
		ReservedAt:         now,
		CreatedTime:        now,
		UpdatedTime:        now,
	}
	if err := tx.Create(redemption).Error; err != nil {
		return nil, err
	}
	topUp.PromotionCampaignId = campaign.Id
	topUp.PromotionCodeId = code.Id
	topUp.PromotionCode = code.Code
	topUp.PromotionRedemptionId = redemption.Id
	return redemption, tx.Save(topUp).Error
}

func MarkTopUpPromotionUsedTx(tx *gorm.DB, topUp *TopUp) error {
	if topUp == nil || topUp.PromotionRedemptionId == 0 {
		return nil
	}
	redemption := &TopUpPromotionRedemption{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", topUp.PromotionRedemptionId).First(redemption).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if redemption.Status != common.TopUpPromotionRedemptionStatusReserved {
		return nil
	}
	now := common.GetTimestamp()
	redemption.Status = common.TopUpPromotionRedemptionStatusUsed
	redemption.UsedAt = now
	redemption.UpdatedTime = now
	return tx.Save(redemption).Error
}

func ReleaseTopUpPromotionReservationTx(tx *gorm.DB, topUp *TopUp) error {
	if topUp == nil || topUp.PromotionRedemptionId == 0 {
		return nil
	}
	redemption := &TopUpPromotionRedemption{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", topUp.PromotionRedemptionId).First(redemption).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if redemption.Status != common.TopUpPromotionRedemptionStatusReserved {
		return nil
	}
	now := common.GetTimestamp()
	redemption.Status = common.TopUpPromotionRedemptionStatusExpired
	redemption.ExpiredAt = now
	redemption.UpdatedTime = now
	return tx.Save(redemption).Error
}

func CleanupTopUpPromotionStates() error {
	now := common.GetTimestamp()
	return DB.Transaction(func(tx *gorm.DB) error {
		var redemptions []*TopUpPromotionRedemption
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("status = ?", common.TopUpPromotionRedemptionStatusReserved).
			Find(&redemptions).Error; err != nil {
			return err
		}
		for _, redemption := range redemptions {
			if redemption.TopUpId == 0 {
				redemption.Status = common.TopUpPromotionRedemptionStatusExpired
				redemption.ExpiredAt = now
				redemption.UpdatedTime = now
				if err := tx.Save(redemption).Error; err != nil {
					return err
				}
				continue
			}
			topUp := &TopUp{}
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", redemption.TopUpId).First(topUp).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				redemption.Status = common.TopUpPromotionRedemptionStatusExpired
				redemption.ExpiredAt = now
				redemption.UpdatedTime = now
				if err := tx.Save(redemption).Error; err != nil {
					return err
				}
				continue
			}
			if topUp.Status == common.TopUpStatusSuccess {
				redemption.Status = common.TopUpPromotionRedemptionStatusUsed
				redemption.UsedAt = topUp.CompleteTime
				redemption.UpdatedTime = now
				if err := tx.Save(redemption).Error; err != nil {
					return err
				}
				continue
			}
			if topUp.Status == common.TopUpStatusExpired || now-topUp.CreateTime >= topUpCouponReservationTTLSeconds {
				topUp.Status = common.TopUpStatusExpired
				if err := tx.Save(topUp).Error; err != nil {
					return err
				}
				redemption.Status = common.TopUpPromotionRedemptionStatusExpired
				redemption.ExpiredAt = now
				redemption.UpdatedTime = now
				if err := tx.Save(redemption).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func FormatTopUpPromotionRule(rule TopUpPromotionRule) string {
	switch rule.DiscountType {
	case common.TopUpPromotionDiscountTypePercent:
		return fmt.Sprintf("满 %.2f 减 %.2f%%", rule.MinAmount, rule.DiscountValue)
	default:
		return fmt.Sprintf("满 %.2f 减 %.2f", rule.MinAmount, rule.DiscountValue)
	}
}
