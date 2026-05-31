package controller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TopUpPromotionRequest struct {
	Id                        int                        `json:"id"`
	Name                      string                     `json:"name"`
	Description               string                     `json:"description"`
	CurrencyCode              string                     `json:"currency_code"`
	Rules                     []model.TopUpPromotionRule `json:"rules"`
	AllowedGroups             []string                   `json:"allowed_groups"`
	ValidFrom                 int64                      `json:"valid_from"`
	ExpiresAt                 int64                      `json:"expires_at"`
	MaxRedemptions            int                        `json:"max_redemptions"`
	MaxRedemptionsPerUser     int                        `json:"max_redemptions_per_user"`
	Codes                     []string                   `json:"codes"`
	AutoCodeCount             int                        `json:"auto_code_count"`
	CodePrefix                string                     `json:"code_prefix"`
	CodeValidFrom             int64                      `json:"code_valid_from"`
	CodeExpiresAt             int64                      `json:"code_expires_at"`
	CodeMaxRedemptions        int                        `json:"code_max_redemptions"`
	CodeMaxRedemptionsPerUser int                        `json:"code_max_redemptions_per_user"`
	Action                    string                     `json:"action"`
	RevokeReason              string                     `json:"revoke_reason"`
}

type TopUpPromotionCodeRequest struct {
	Id                    int      `json:"id"`
	CampaignId            int      `json:"campaign_id"`
	Code                  string   `json:"code"`
	Codes                 []string `json:"codes"`
	AutoCodeCount         int      `json:"auto_code_count"`
	CodePrefix            string   `json:"code_prefix"`
	ValidFrom             int64    `json:"valid_from"`
	ExpiresAt             int64    `json:"expires_at"`
	MaxRedemptions        int      `json:"max_redemptions"`
	MaxRedemptionsPerUser int      `json:"max_redemptions_per_user"`
	Action                string   `json:"action"`
	RevokeReason          string   `json:"revoke_reason"`
}

func GetAllTopUpPromotions(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	campaigns, total, err := model.GetAllTopUpPromotionCampaigns(pageInfo, model.TopUpPromotionFilter{
		Keyword: c.Query("keyword"),
		Status:  c.Query("status"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(campaigns)
	common.ApiSuccess(c, pageInfo)
}

func GetTopUpPromotion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	campaign, err := model.GetTopUpPromotionCampaignById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, campaign)
}

func AddTopUpPromotion(c *gin.Context) {
	req := TopUpPromotionRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	campaign := &model.TopUpPromotionCampaign{
		Name:                  req.Name,
		Description:           req.Description,
		CurrencyCode:          req.CurrencyCode,
		Status:                common.TopUpPromotionStatusActive,
		ValidFrom:             req.ValidFrom,
		ExpiresAt:             req.ExpiresAt,
		MaxRedemptions:        req.MaxRedemptions,
		MaxRedemptionsPerUser: req.MaxRedemptionsPerUser,
		CreatedByAdminId:      c.GetInt("id"),
		CreatedTime:           common.GetTimestamp(),
		UpdatedTime:           common.GetTimestamp(),
	}
	if err := campaign.SetRules(req.Rules); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := campaign.SetAllowedGroups(req.AllowedGroups); err != nil {
		common.ApiError(c, err)
		return
	}

	var result *model.TopUpPromotionCampaign
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := campaign.Validate(); err != nil {
			return err
		}
		if err := tx.Create(campaign).Error; err != nil {
			return err
		}
		codes, err := buildPromotionCodeTextsTx(tx, req.Codes, req.AutoCodeCount, req.CodePrefix)
		if err != nil {
			return err
		}
		for _, codeText := range codes {
			code := &model.TopUpPromotionCode{
				CampaignId:            campaign.Id,
				Code:                  codeText,
				Status:                common.TopUpPromotionStatusActive,
				ValidFrom:             req.CodeValidFrom,
				ExpiresAt:             req.CodeExpiresAt,
				MaxRedemptions:        req.CodeMaxRedemptions,
				MaxRedemptionsPerUser: req.CodeMaxRedemptionsPerUser,
				CreatedByAdminId:      c.GetInt("id"),
				CreatedTime:           common.GetTimestamp(),
				UpdatedTime:           common.GetTimestamp(),
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
		}
		result = campaign
		return nil
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员创建促销活动 #%d %s", result.Id, result.Name))
	result, err = model.GetTopUpPromotionCampaignById(result.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func UpdateTopUpPromotion(c *gin.Context) {
	req := TopUpPromotionRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if strings.TrimSpace(req.Action) == "revoke" {
		campaign, err := model.RevokeTopUpPromotionCampaign(req.Id, c.GetInt("id"), req.RevokeReason)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员撤销促销活动 #%d %s", campaign.Id, campaign.Name))
		common.ApiSuccess(c, campaign)
		return
	}

	campaign, err := model.GetTopUpPromotionCampaignById(req.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if campaign.Status == common.TopUpPromotionStatusRevoked {
		common.ApiErrorMsg(c, "已撤销的促销活动不能编辑")
		return
	}
	if req.Name != "" {
		campaign.Name = req.Name
	}
	campaign.Description = req.Description
	if req.CurrencyCode != "" {
		campaign.CurrencyCode = req.CurrencyCode
	}
	if len(req.Rules) > 0 {
		if err := campaign.SetRules(req.Rules); err != nil {
			common.ApiError(c, err)
			return
		}
	}
	if err := campaign.SetAllowedGroups(req.AllowedGroups); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.ValidFrom > 0 {
		campaign.ValidFrom = req.ValidFrom
	}
	campaign.ExpiresAt = req.ExpiresAt
	campaign.MaxRedemptions = req.MaxRedemptions
	campaign.MaxRedemptionsPerUser = req.MaxRedemptionsPerUser
	if err := campaign.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员更新促销活动 #%d %s", campaign.Id, campaign.Name))
	result, err := model.GetTopUpPromotionCampaignById(campaign.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func GetTopUpPromotionCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	campaignId, _ := strconv.Atoi(c.Query("campaign_id"))
	codes, total, err := model.GetTopUpPromotionCodes(pageInfo, model.TopUpPromotionCodeFilter{
		CampaignId: campaignId,
		Keyword:    c.Query("keyword"),
		Status:     c.Query("status"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func AddTopUpPromotionCodes(c *gin.Context) {
	req := TopUpPromotionCodeRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	rawCodes := append([]string{}, req.Codes...)
	if strings.TrimSpace(req.Code) != "" {
		rawCodes = append(rawCodes, req.Code)
	}
	codes, err := model.CreateTopUpPromotionCodes(
		req.CampaignId,
		c.GetInt("id"),
		rawCodes,
		req.AutoCodeCount,
		req.CodePrefix,
		req.ValidFrom,
		req.ExpiresAt,
		req.MaxRedemptions,
		req.MaxRedemptionsPerUser,
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员为促销活动 #%d 新增 %d 个促销码", req.CampaignId, len(codes)))
	common.ApiSuccess(c, codes)
}

func UpdateTopUpPromotionCode(c *gin.Context) {
	req := TopUpPromotionCodeRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if strings.TrimSpace(req.Action) == "revoke" {
		code, err := model.RevokeTopUpPromotionCode(req.Id, c.GetInt("id"), req.RevokeReason)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员撤销促销码 #%d %s", code.Id, code.Code))
		common.ApiSuccess(c, code)
		return
	}
	var result *model.TopUpPromotionCode
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		code := &model.TopUpPromotionCode{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", req.Id).First(code).Error; err != nil {
			return err
		}
		if code.Status == common.TopUpPromotionStatusRevoked {
			return errors.New("已撤销的促销码不能编辑")
		}
		if strings.TrimSpace(req.Code) != "" {
			code.Code = req.Code
		}
		if req.ValidFrom > 0 {
			code.ValidFrom = req.ValidFrom
		}
		code.ExpiresAt = req.ExpiresAt
		code.MaxRedemptions = req.MaxRedemptions
		code.MaxRedemptionsPerUser = req.MaxRedemptionsPerUser
		code.UpdatedTime = common.GetTimestamp()
		if err := code.Validate(); err != nil {
			return err
		}
		if err := tx.Save(code).Error; err != nil {
			return err
		}
		result = code
		return nil
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := fillOnePromotionCode(result); err != nil {
		common.ApiError(c, err)
		return
	}
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("管理员更新促销码 #%d %s", result.Id, result.Code))
	common.ApiSuccess(c, result)
}

func GetTopUpPromotionRedemptions(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	campaignId, _ := strconv.Atoi(c.Query("campaign_id"))
	codeId, _ := strconv.Atoi(c.Query("code_id"))
	userId, _ := strconv.Atoi(c.Query("user_id"))
	redemptions, total, err := model.GetTopUpPromotionRedemptions(pageInfo, model.TopUpPromotionRedemptionFilter{
		CampaignId: campaignId,
		CodeId:     codeId,
		UserId:     userId,
		Keyword:    c.Query("keyword"),
		Status:     c.Query("status"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(redemptions)
	common.ApiSuccess(c, pageInfo)
}

func buildPromotionCodeTextsTx(tx *gorm.DB, rawCodes []string, autoCount int, prefix string) ([]string, error) {
	if autoCount == 0 && len(rawCodes) == 0 {
		autoCount = 1
	}
	if autoCount < 0 || autoCount > 500 {
		return nil, errors.New("自动生成数量必须在 0-500 之间")
	}
	codes := make([]string, 0, len(rawCodes)+autoCount)
	seen := map[string]struct{}{}
	for _, rawCode := range rawCodes {
		code := model.NormalizeTopUpPromotionCode(rawCode)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		if err := ensurePromotionCodeNotExistsTx(tx, code); err != nil {
			return nil, err
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	targetCount := len(codes) + autoCount
	for len(codes) < targetCount {
		code, err := model.GenerateTopUpPromotionCode(prefix)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[code]; ok {
			continue
		}
		exists, err := promotionCodeExistsTx(tx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, errors.New("请至少提供一个促销码")
	}
	return codes, nil
}

func ensurePromotionCodeNotExistsTx(tx *gorm.DB, code string) error {
	exists, err := promotionCodeExistsTx(tx, code)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("促销码 %s 已存在", code)
	}
	return nil
}

func promotionCodeExistsTx(tx *gorm.DB, code string) (bool, error) {
	var count int64
	if err := tx.Model(&model.TopUpPromotionCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func fillOnePromotionCode(code *model.TopUpPromotionCode) error {
	if code == nil {
		return nil
	}
	codes, _, err := model.GetTopUpPromotionCodes(&common.PageInfo{Page: 1, PageSize: 100}, model.TopUpPromotionCodeFilter{CampaignId: code.CampaignId})
	if err != nil {
		return err
	}
	for _, candidate := range codes {
		if candidate != nil && candidate.Id == code.Id {
			*code = *candidate
			return nil
		}
	}
	return nil
}
