package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type startImpersonationRequest struct {
	ReadOnly   bool `json:"read_only"`
	BreakGlass bool `json:"break_glass"`
}

func writeVerificationRequired(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"success": false,
		"message": "需要安全验证",
		"code":    "VERIFICATION_REQUIRED",
	})
}

func buildCurrentUserEnvelope(c *gin.Context, user *model.User) gin.H {
	if user == nil {
		return gin.H{}
	}
	data := buildSelfResponseData(user, user.Role)
	session := sessions.Default(c)
	attachSessionStateToSelfResponse(session, user, data)
	sessionState := service.GetImpersonationSessionState(session)
	if !sessionState.Active && (sessionState.HeaderAliasID != 0 || sessionState.HeaderAliasCAHID != "") {
		service.ClearImpersonationHeaderAlias(session)
		_ = session.Save()
	}
	return data
}

func StartUserImpersonation(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil || targetID == 0 {
		common.ApiErrorMsg(c, "无效的用户 ID")
		return
	}

	var req startImpersonationRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "无效的请求参数")
		return
	}

	operator, err := model.GetUserById(c.GetInt("id"), true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	target, err := model.GetUserById(targetID, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if operator.Id == target.Id {
		common.ApiErrorMsg(c, "不能仿冒当前登录用户")
		return
	}
	if target.Status != common.UserStatusEnabled {
		common.ApiErrorMsg(c, "目标用户当前不可访问")
		return
	}
	if operator.Role <= target.Role && operator.Role != common.RoleRootUser {
		common.ApiErrorMsg(c, "无权仿冒同级或更高权限的用户")
		return
	}

	session := sessions.Default(c)
	result, err := service.RequestOrStartImpersonation(session, operator, target, req.ReadOnly, req.BreakGlass)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	action := "impersonation_request"
	content := "管理员已发起仿冒访问请求"
	if result.State == service.ImpersonationStartStateActivated {
		action = "impersonation_start"
		content = "管理员已开始仿冒用户会话"
	}
	if result.State == service.ImpersonationStartStateBreakGlass {
		action = "break_glass_start"
		content = "管理员已通过打破玻璃开始访问用户会话"
	}
	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "user",
		Action:     action,
		TargetType: "user",
		TargetId:   target.Id,
		TargetName: target.Username,
		Content:    content,
		Details: map[string]interface{}{
			"read_only":   req.ReadOnly,
			"break_glass": req.BreakGlass,
			"grant_id":    result.Grant.Id,
			"start_state": result.State,
		},
	})

	if result.State == service.ImpersonationStartStateActivated || result.State == service.ImpersonationStartStateBreakGlass {
		target, err = model.GetUserById(target.Id, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data": gin.H{
				"start_state": result.State,
				"grant_id":    result.Grant.Id,
				"user":        buildCurrentUserEnvelope(c, target),
			},
		})
		return
	}

	message := "访问请求已发送，等待用户批准"
	if result.State == service.ImpersonationStartStatePendingExisting {
		message = "该用户已有待处理的访问请求，请等待用户批准"
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"start_state": result.State,
			"grant_id":    result.Grant.Id,
		},
	})
}

func StopUserImpersonation(c *gin.Context) {
	session := sessions.Default(c)
	state := service.GetImpersonationSessionState(session)
	if !state.Active {
		respondWithCurrentUser(c, c.GetInt("id"), c.GetInt("role"))
		return
	}

	if _, err := service.StopCurrentImpersonation(session, false); err != nil {
		common.ApiError(c, err)
		return
	}

	user, err := model.GetUserById(state.OriginalID, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"user": buildCurrentUserEnvelope(c, user),
		},
	})
}

func GetImpersonationRequest(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	grant, err := model.GetImpersonationGrantByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.ApiErrorMsg(c, "访问请求不存在")
			return
		}
		common.ApiError(c, err)
		return
	}
	if grant.TargetUserId != c.GetInt("id") {
		common.ApiErrorMsg(c, "无权查看该访问请求")
		return
	}

	now := time.Now()
	if model.ExpireImpersonationGrantIfNeeded(grant, now) {
		_ = model.SaveImpersonationGrant(grant)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":                    grant.Id,
			"state":                 grant.State,
			"mode":                  grant.Mode,
			"requested_read_only":   grant.RequestedReadOnly,
			"operator_id":           grant.OperatorId,
			"operator_username":     grant.OperatorUsername,
			"operator_cah_id":       grant.OperatorCAHID,
			"target_user_id":        grant.TargetUserId,
			"target_username":       grant.TargetUsername,
			"requested_at":          grant.RequestedAt,
			"granted_expires_at":    grant.GrantedExpiresAt,
			"approved_at":           grant.ApprovedAt,
			"requires_verification": service.UserNeedsSupportAccessVerification(c.GetInt("id")),
		},
	})
}

func ApproveImpersonationRequest(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	grant, err := model.GetImpersonationGrantByToken(token)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if grant.TargetUserId != c.GetInt("id") {
		common.ApiErrorMsg(c, "无权批准该访问请求")
		return
	}
	if service.UserNeedsSupportAccessVerification(c.GetInt("id")) && !CheckSecureVerification(c) {
		writeVerificationRequired(c)
		return
	}

	user, err := model.GetUserById(c.GetInt("id"), true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	method := "email_link"
	if service.UserNeedsSupportAccessVerification(user.Id) {
		method = "secure_verification"
	}
	if err := service.ApproveImpersonationGrant(grant, user, method); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "访问请求已批准，客服需在 24 小时内开始使用",
	})
}

func RejectImpersonationRequest(c *gin.Context) {
	token := strings.TrimSpace(c.Param("token"))
	grant, err := model.GetImpersonationGrantByToken(token)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if grant.TargetUserId != c.GetInt("id") {
		common.ApiErrorMsg(c, "无权拒绝该访问请求")
		return
	}

	user, err := model.GetUserById(c.GetInt("id"), true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := service.RejectImpersonationGrant(grant, user, "user_rejected"); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "访问请求已拒绝",
	})
}

func OpenSelfServiceSupportAccess(c *gin.Context) {
	userID := c.GetInt("id")
	if service.UserNeedsSupportAccessVerification(userID) && !CheckSecureVerification(c) {
		writeVerificationRequired(c)
		return
	}

	user, err := model.GetUserById(userID, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	grant, err := service.OpenSelfServiceSupportAccess(user)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已开放一次客服访问，24 小时内可由管理员激活一次",
		"data": gin.H{
			"id":                grant.Id,
			"state":             grant.State,
			"granted_expires_at": grant.GrantedExpiresAt,
		},
	})
}

func CloseSelfServiceSupportAccess(c *gin.Context) {
	if err := service.CloseSelfServiceSupportAccess(c.GetInt("id")); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "未使用的客服访问已关闭",
	})
}

func GetImpersonationHistory(c *gin.Context) {
	userID := c.GetInt("id")
	incidents, err := model.LoadBreakGlassIncidentsByUserID(userID, 20)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var openGrant *model.ImpersonationGrant
	openGrant, err = model.FindLatestOpenSelfServiceGrant(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		common.ApiError(c, err)
		return
	}
	if openGrant != nil && model.ExpireImpersonationGrantIfNeeded(openGrant, time.Now()) {
		_ = model.SaveImpersonationGrant(openGrant)
		if openGrant.State != model.ImpersonationStateApproved {
			openGrant = nil
		}
	}

	items := make([]gin.H, 0, len(incidents))
	for _, incident := range incidents {
		if incident == nil || incident.Grant == nil {
			continue
		}
		actions := make([]gin.H, 0, len(incident.Actions))
		for _, action := range incident.Actions {
			if action == nil {
				continue
			}
			actions = append(actions, gin.H{
				"id":          action.Id,
				"created_at":  action.CreatedAt,
				"method":      action.Method,
				"path":        action.Path,
				"route":       action.Route,
				"status_code": action.StatusCode,
				"success":     action.Success,
			})
		}
		items = append(items, gin.H{
			"id":                incident.Grant.Id,
			"operator_id":       incident.Grant.OperatorId,
			"operator_username": incident.Grant.OperatorUsername,
			"started_at": func() *time.Time {
				if incident.Grant.ActivatedAt != nil {
					return incident.Grant.ActivatedAt
				}
				return &incident.Grant.RequestedAt
			}(),
			"ended_at":       incident.Grant.EndedAt,
			"active":         incident.Grant.EndedAt == nil,
			"action_count":   len(actions),
			"actions":        actions,
		})
	}

	responseData := gin.H{
		"break_glass_incidents": items,
	}
	if openGrant != nil {
		responseData["open_support_access"] = gin.H{
			"id":                 openGrant.Id,
			"granted_expires_at": openGrant.GrantedExpiresAt,
			"state":              openGrant.State,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    responseData,
	})
}
