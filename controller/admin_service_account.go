package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	defaultAdminServiceAccountDays = 90
	maxAdminServiceAccountDays     = 365
)

type adminServiceAccountRequest struct {
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	Target        string  `json:"target"`
	UserID        int     `json:"user_id"`
	ExpiresAt     int64   `json:"expires_at"`
	ExpiresInDays int     `json:"expires_in_days"`
	AllowIps      *string `json:"allow_ips"`
	Status        int     `json:"status"`
}

type adminServiceAccountView struct {
	*model.AdminServiceAccount
	Expired   bool        `json:"expired"`
	Target    interface{} `json:"target"`
	CreatedBy interface{} `json:"created_by"`
}

func GetAdminServiceAccounts(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	keyword := c.Query("keyword")
	maxRole := c.GetInt("role")
	if maxRole == common.RoleRootUser {
		maxRole = 0
	}

	accounts, err := model.GetAdminServiceAccounts(keyword, maxRole, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	total, err := model.CountAdminServiceAccounts(keyword, maxRole)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(buildAdminServiceAccountViews(accounts))
	common.ApiSuccess(c, pageInfo)
}

func CreateAdminServiceAccount(c *gin.Context) {
	req := adminServiceAccountRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}

	operator, target, expiresAt, allowIps, err := prepareAdminServiceAccountRequest(c, req, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	name, err := model.NormalizeAdminServiceAccountName(req.Name)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	descriptionInput := ""
	if req.Description != nil {
		descriptionInput = *req.Description
	}
	description, err := model.NormalizeAdminServiceAccountDescription(descriptionInput)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	account := &model.AdminServiceAccount{
		Name:              name,
		Description:       description,
		UserID:            target.Id,
		Username:          target.Username,
		UserCAHID:         target.CAHID,
		UserRole:          target.Role,
		CreatedByID:       operator.Id,
		CreatedByUsername: operator.Username,
		CreatedByCAHID:    operator.CAHID,
		Status:            model.AdminServiceAccountStatusEnabled,
		Scopes:            "admin:api",
		AllowIps:          allowIps,
		ExpiresAt:         expiresAt,
	}

	credential, err := model.CreateAdminServiceAccount(account, target)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "admin_service_account",
		Action:     "create",
		TargetType: "admin_service_account",
		TargetId:   account.Id,
		TargetName: account.Name,
		Details: map[string]interface{}{
			"service_account_id": account.ServiceAccountID,
			"target_user_id":     target.Id,
			"target_username":    target.Username,
			"expires_at":         account.ExpiresAt,
		},
	})

	common.ApiSuccess(c, gin.H{
		"account":         buildAdminServiceAccountView(account),
		"credential":      credential,
		"credential_type": "Bearer JWT",
	})
}

func UpdateAdminServiceAccount(c *gin.Context) {
	account, target, err := loadManageableAdminServiceAccount(c)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req := adminServiceAccountRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}
	if req.Name != "" {
		name, err := model.NormalizeAdminServiceAccountName(req.Name)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		account.Name = name
	}
	if req.Description != nil {
		description, err := model.NormalizeAdminServiceAccountDescription(*req.Description)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		account.Description = description
	}
	if req.Status != 0 {
		if !model.IsAdminServiceAccountStatusValid(req.Status) {
			common.ApiErrorMsg(c, "Service Account 状态无效")
			return
		}
		account.Status = req.Status
	}
	if req.AllowIps != nil {
		allowIps, err := model.NormalizeAdminServiceAccountAllowIps(req.AllowIps)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		account.AllowIps = allowIps
	}

	if err = account.UpdateMetadata(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "admin_service_account",
		Action:     "update",
		TargetType: "admin_service_account",
		TargetId:   account.Id,
		TargetName: account.Name,
		Details: map[string]interface{}{
			"service_account_id": account.ServiceAccountID,
			"target_user_id":     target.Id,
			"target_username":    target.Username,
			"status":             account.Status,
		},
	})
	common.ApiSuccess(c, buildAdminServiceAccountView(account))
}

func RotateAdminServiceAccountCredential(c *gin.Context) {
	account, target, err := loadManageableAdminServiceAccount(c)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req := adminServiceAccountRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}
	expiresAt, err := resolveAdminServiceAccountExpiresAt(req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	credential, err := model.RotateAdminServiceAccountCredential(account, target, expiresAt)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "admin_service_account",
		Action:     "rotate",
		TargetType: "admin_service_account",
		TargetId:   account.Id,
		TargetName: account.Name,
		Details: map[string]interface{}{
			"service_account_id": account.ServiceAccountID,
			"target_user_id":     target.Id,
			"target_username":    target.Username,
			"expires_at":         account.ExpiresAt,
		},
	})
	common.ApiSuccess(c, gin.H{
		"account":         buildAdminServiceAccountView(account),
		"credential":      credential,
		"credential_type": "Bearer JWT",
	})
}

func DeleteAdminServiceAccount(c *gin.Context) {
	account, target, err := loadManageableAdminServiceAccount(c)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err = account.Delete(); err != nil {
		common.ApiError(c, err)
		return
	}
	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "admin_service_account",
		Action:     "delete",
		TargetType: "admin_service_account",
		TargetId:   account.Id,
		TargetName: account.Name,
		Details: map[string]interface{}{
			"service_account_id": account.ServiceAccountID,
			"target_user_id":     target.Id,
			"target_username":    target.Username,
		},
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func prepareAdminServiceAccountRequest(c *gin.Context, req adminServiceAccountRequest, requireTarget bool) (*model.User, *model.User, int64, *string, error) {
	operator, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	target, err := resolveAdminServiceAccountTarget(operator, req, requireTarget)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	if err = ensureCanManageAdminServiceAccountForTarget(operator, target); err != nil {
		return nil, nil, 0, nil, err
	}
	expiresAt, err := resolveAdminServiceAccountExpiresAt(req)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	allowIps, err := model.NormalizeAdminServiceAccountAllowIps(req.AllowIps)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	return operator, target, expiresAt, allowIps, nil
}

func resolveAdminServiceAccountTarget(operator *model.User, req adminServiceAccountRequest, requireTarget bool) (*model.User, error) {
	if operator == nil {
		return nil, errors.New("当前管理员不存在")
	}
	if req.UserID > 0 {
		return model.GetUserById(req.UserID, false)
	}
	target := strings.TrimSpace(req.Target)
	if target == "" {
		if requireTarget {
			return operator, nil
		}
		return nil, errors.New("目标管理员不能为空")
	}
	if id, err := strconv.Atoi(target); err == nil && id > 0 {
		return model.GetUserById(id, false)
	}
	if model.IsValidCAHID(target) {
		return model.GetUserByCAHID(target, false)
	}
	user := &model.User{}
	err := model.DB.Omit("password").Where("username = ?", target).First(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("目标管理员不存在")
	}
	return user, err
}

func ensureCanManageAdminServiceAccountForTarget(operator *model.User, target *model.User) error {
	if operator == nil || target == nil {
		return errors.New("管理员信息无效")
	}
	if target.Status != common.UserStatusEnabled {
		return errors.New("只能为启用状态的管理员生成 Service Account")
	}
	if target.Role < common.RoleAdminUser {
		return errors.New("只能为管理员或超级管理员生成 Service Account")
	}
	if operator.Role != common.RoleRootUser && target.Role > operator.Role {
		return errors.New("无法为权限高于自己的管理员生成 Service Account")
	}
	if operator.Role != common.RoleRootUser && target.Role == common.RoleRootUser {
		return errors.New("只有超级管理员可以管理超级管理员的 Service Account")
	}
	return nil
}

func resolveAdminServiceAccountExpiresAt(req adminServiceAccountRequest) (int64, error) {
	now := common.GetTimestamp()
	maxExpiresAt := now + int64(maxAdminServiceAccountDays)*24*60*60
	if req.ExpiresAt > 0 {
		if req.ExpiresAt <= now+60 {
			return 0, errors.New("Service Account 过期时间必须晚于当前时间")
		}
		if req.ExpiresAt > maxExpiresAt {
			return 0, errors.New("Service Account 有效期不能超过 365 天")
		}
		return req.ExpiresAt, nil
	}

	days := req.ExpiresInDays
	if days == 0 {
		days = defaultAdminServiceAccountDays
	}
	if days < 1 || days > maxAdminServiceAccountDays {
		return 0, errors.New("Service Account 有效天数必须在 1 到 365 天之间")
	}
	return now + int64(days)*24*60*60, nil
}

func loadManageableAdminServiceAccount(c *gin.Context) (*model.AdminServiceAccount, *model.User, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return nil, nil, errors.New("Service Account ID 无效")
	}
	account, err := model.GetAdminServiceAccountByID(id)
	if err != nil {
		return nil, nil, err
	}
	target, err := model.GetUserById(account.UserID, false)
	if err != nil {
		return nil, nil, err
	}
	operator, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		return nil, nil, err
	}
	if err = ensureCanManageAdminServiceAccountForTarget(operator, target); err != nil {
		return nil, nil, err
	}
	return account, target, nil
}

func buildAdminServiceAccountViews(accounts []*model.AdminServiceAccount) []adminServiceAccountView {
	views := make([]adminServiceAccountView, 0, len(accounts))
	for _, account := range accounts {
		views = append(views, buildAdminServiceAccountView(account))
	}
	return views
}

func buildAdminServiceAccountView(account *model.AdminServiceAccount) adminServiceAccountView {
	view := adminServiceAccountView{
		AdminServiceAccount: account,
		Expired:             account != nil && account.ExpiresAt <= time.Now().Unix(),
	}
	if account == nil {
		return view
	}
	view.Target = gin.H{
		"id":       account.UserID,
		"username": account.Username,
		"cah_id":   account.UserCAHID,
		"role":     account.UserRole,
	}
	view.CreatedBy = gin.H{
		"id":       account.CreatedByID,
		"username": account.CreatedByUsername,
		"cah_id":   account.CreatedByCAHID,
	}
	return view
}
