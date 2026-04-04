package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-contrib/sessions"
	"gorm.io/gorm"
)

const (
	ImpersonationStartStatePendingApproval = "pending_approval"
	ImpersonationStartStatePendingExisting = "pending_existing"
	ImpersonationStartStateActivated       = "activated"
	ImpersonationStartStateBreakGlass      = "break_glass"
)

type StartImpersonationResult struct {
	State string
	Grant *model.ImpersonationGrant
}

func fillImpersonationGrantSnapshot(grant *model.ImpersonationGrant, operator *model.User, target *model.User) {
	if grant == nil {
		return
	}
	if operator != nil {
		grant.OperatorId = operator.Id
		grant.OperatorUsername = strings.TrimSpace(operator.Username)
		grant.OperatorCAHID = strings.TrimSpace(operator.CAHID)
	}
	if target != nil {
		grant.TargetUserId = target.Id
		grant.TargetUsername = strings.TrimSpace(target.Username)
		grant.TargetDisplayName = strings.TrimSpace(target.DisplayName)
		grant.TargetCAHID = strings.TrimSpace(target.CAHID)
	}
}

func defaultServerAddress() string {
	serverAddress := strings.TrimSpace(system_setting.ServerAddress)
	if serverAddress == "" {
		serverAddress = "https://privnode.com"
	}
	return strings.TrimRight(serverAddress, "/")
}

func buildImpersonationApprovalURL(token string) string {
	redirectTarget := fmt.Sprintf("/console/personal?support_access_token=%s", url.QueryEscape(strings.TrimSpace(token)))
	return common.BuildURL(defaultServerAddress(), "/login?redirect="+url.QueryEscape(redirectTarget))
}

func buildPersonalSettingsURL() string {
	return common.BuildURL(defaultServerAddress(), "/console/personal")
}

func impersonationAccessLabel(readOnly bool) string {
	if readOnly {
		return "只读访问"
	}
	return "标准访问"
}

func notifyUserOfImpersonationRequest(target *model.User, grant *model.ImpersonationGrant) error {
	if target == nil || grant == nil {
		return nil
	}
	link := buildImpersonationApprovalURL(grant.ApprovalToken)
	content := fmt.Sprintf(
		"客服人员 **%s** 请求以 **%s** 的方式访问您的账户。\n\n- 访问级别：%s\n- 使用限制：批准后仅可激活一次，需在 24 小时内开始使用\n- 会话时效：激活后 24 小时自动失效\n- 二次验证：如果您已启用 2FA 或 Passkey，批准前需要再次验证\n\n[查看并处理该请求](%s)",
		grant.OperatorUsername,
		impersonationAccessLabel(grant.RequestedReadOnly),
		func() string {
			if grant.RequestedReadOnly {
				return "只读会话，客服仅可查看数据，不能代您执行操作"
			}
			return "标准会话，客服可以像您本人一样查看并执行操作"
		}(),
		link,
	)

	return NotifyUser(target.Id, target.Email, target.GetSetting(), dto.NewNotify(
		"support_access_request",
		fmt.Sprintf("%s 请求访问您的账户", grant.OperatorUsername),
		content,
		nil,
	))
}

func notifyOperatorOfGrantDecision(operator *model.User, grant *model.ImpersonationGrant, approved bool) error {
	if operator == nil || grant == nil {
		return nil
	}
	actionText := "已拒绝"
	title := fmt.Sprintf("%s 未批准账户访问请求", grant.TargetUsername)
	if approved {
		actionText = "已批准"
		title = fmt.Sprintf("%s 已批准账户访问请求", grant.TargetUsername)
	}
	content := fmt.Sprintf(
		"用户 **%s** %s您对其账户的访问请求。\n\n- 访问级别：%s\n- 下一步：请重新点击用户管理中的仿冒入口以开始会话\n- 使用限制：批准后仅可激活一次，需在 24 小时内开始使用",
		grant.TargetUsername,
		actionText,
		impersonationAccessLabel(grant.RequestedReadOnly),
	)
	return NotifyUser(operator.Id, operator.Email, operator.GetSetting(), dto.NewNotify(
		"support_access_decision",
		title,
		content,
		nil,
	))
}

func notifyUserOfImpersonationActivation(target *model.User, grant *model.ImpersonationGrant) error {
	if target == nil || grant == nil {
		return nil
	}
	content := fmt.Sprintf(
		"客服人员 **%s** 已开始以 **%s** 的方式访问您的账户。\n\n- 会话级别：%s\n- 会话失效时间：%s\n- 入口：您可以在个人设置中查看相关访问记录",
		grant.OperatorUsername,
		impersonationAccessLabel(grant.ActiveReadOnly),
		func() string {
			if grant.ActiveReadOnly {
				return "只读会话，客服仅可查看数据"
			}
			return "标准会话，客服可代表您执行操作"
		}(),
		func() string {
			if grant.SessionExpiresAt == nil {
				return "手动结束"
			}
			return grant.SessionExpiresAt.Format("2006-01-02 15:04:05")
		}(),
	)
	return NotifyUser(target.Id, target.Email, target.GetSetting(), dto.NewNotify(
		"support_access_started",
		fmt.Sprintf("%s 已开始访问您的账户", grant.OperatorUsername),
		content,
		nil,
	))
}

func notifyUserOfBreakGlassStarted(target *model.User, grant *model.ImpersonationGrant) error {
	if target == nil || grant == nil {
		return nil
	}
	content := fmt.Sprintf(
		"客服人员 **%s** 已通过 **打破玻璃** 方式强制访问您的账户。\n\n- 当前状态：访问进行中\n- 限制说明：本次访问不受一次性或 24 小时限制\n- 查看入口：%s\n\n您现在会在界面顶部看到相关安全横幅，并可查看访问期间的操作记录。",
		grant.OperatorUsername,
		buildPersonalSettingsURL(),
	)
	return NotifyUser(target.Id, target.Email, target.GetSetting(), dto.NewNotify(
		"support_break_glass_started",
		fmt.Sprintf("%s 通过打破玻璃访问了您的账户", grant.OperatorUsername),
		content,
		nil,
	))
}

func notifyUserOfBreakGlassEnded(target *model.User, grant *model.ImpersonationGrant) error {
	if target == nil || grant == nil {
		return nil
	}
	content := fmt.Sprintf(
		"客服人员 **%s** 的打破玻璃访问已结束。\n\n- 开始时间：%s\n- 结束时间：%s\n- 查看入口：%s\n\n您可以在个人设置中查看本次访问期间的操作记录。",
		grant.OperatorUsername,
		func() string {
			if grant.ActivatedAt == nil {
				return grant.RequestedAt.Format("2006-01-02 15:04:05")
			}
			return grant.ActivatedAt.Format("2006-01-02 15:04:05")
		}(),
		func() string {
			if grant.EndedAt == nil {
				return time.Now().Format("2006-01-02 15:04:05")
			}
			return grant.EndedAt.Format("2006-01-02 15:04:05")
		}(),
		buildPersonalSettingsURL(),
	)
	return NotifyUser(target.Id, target.Email, target.GetSetting(), dto.NewNotify(
		"support_break_glass_ended",
		fmt.Sprintf("%s 的打破玻璃访问已结束", grant.OperatorUsername),
		content,
		nil,
	))
}

func UserNeedsSupportAccessVerification(userID int) bool {
	if userID == 0 {
		return false
	}
	if model.IsTwoFAEnabled(userID) {
		return true
	}
	passkey, err := model.GetPasskeyByUserID(userID)
	return err == nil && passkey != nil
}

func RequestOrStartImpersonation(session sessions.Session, operator *model.User, target *model.User, requestReadOnly bool, breakGlass bool) (*StartImpersonationResult, error) {
	if session == nil || operator == nil || target == nil {
		return nil, errors.New("缺少必要的仿冒参数")
	}
	if GetImpersonationSessionState(session).Active {
		return nil, errors.New("请先结束当前仿冒会话")
	}
	if operator.Id == target.Id {
		return nil, errors.New("不能仿冒当前登录账户")
	}

	now := time.Now()
	if breakGlass {
		grant := &model.ImpersonationGrant{
			Source:           model.ImpersonationSourceAdminRequest,
			Mode:             model.ImpersonationModeBreakGlass,
			State:            model.ImpersonationStateApproved,
			RequestedReadOnly: false,
			RequestedAt:      now,
			ApprovedAt:       &now,
			ActivatedAt:      &now,
			ActiveReadOnly:   false,
		}
		fillImpersonationGrantSnapshot(grant, operator, target)
		if err := model.CreateImpersonationGrant(grant); err != nil {
			return nil, err
		}
		if err := model.ActivateApprovedImpersonationGrant(grant.Id, operator.Id, operator.Username, operator.CAHID, false, now, nil); err != nil {
			return nil, err
		}
		grant.State = model.ImpersonationStateActive
		grant.ActivatedAt = &now
		if err := BeginImpersonationSession(session, operator, target, grant, false); err != nil {
			_ = model.DB.Model(&model.ImpersonationGrant{}).Where("id = ?", grant.Id).
				Updates(map[string]interface{}{
					"state":        model.ImpersonationStateCompleted,
					"ended_at":     now,
				}).Error
			return nil, err
		}
		_ = notifyUserOfBreakGlassStarted(target, grant)
		return &StartImpersonationResult{State: ImpersonationStartStateBreakGlass, Grant: grant}, nil
	}

	if activatable, err := model.FindLatestActivatableImpersonationGrant(operator.Id, target.Id, requestReadOnly); err == nil && activatable != nil {
		if model.ExpireImpersonationGrantIfNeeded(activatable, now) {
			_ = model.SaveImpersonationGrant(activatable)
		}
		if activatable.State == model.ImpersonationStateApproved {
			sessionExpiresAt := now.Add(ImpersonationSessionWindow)
			effectiveReadOnly := requestReadOnly || activatable.RequestedReadOnly
			if err := model.ActivateApprovedImpersonationGrant(activatable.Id, operator.Id, operator.Username, operator.CAHID, effectiveReadOnly, now, &sessionExpiresAt); err != nil {
				return nil, err
			}
			activatable.State = model.ImpersonationStateActive
			activatable.ActiveReadOnly = effectiveReadOnly
			activatable.OperatorId = operator.Id
			activatable.OperatorUsername = operator.Username
			activatable.OperatorCAHID = operator.CAHID
			activatable.ActivatedAt = &now
			activatable.SessionExpiresAt = &sessionExpiresAt
			if err := BeginImpersonationSession(session, operator, target, activatable, effectiveReadOnly); err != nil {
				_ = model.DB.Model(&model.ImpersonationGrant{}).Where("id = ?", activatable.Id).
					Updates(map[string]interface{}{
						"state":              model.ImpersonationStateApproved,
						"activated_at":       nil,
						"session_expires_at": nil,
						"active_read_only":   false,
					}).Error
				return nil, err
			}
			_ = notifyUserOfImpersonationActivation(target, activatable)
			return &StartImpersonationResult{State: ImpersonationStartStateActivated, Grant: activatable}, nil
		}
	}

	if pending, err := model.FindLatestPendingImpersonationGrant(operator.Id, target.Id, requestReadOnly); err == nil && pending != nil {
		return &StartImpersonationResult{State: ImpersonationStartStatePendingExisting, Grant: pending}, nil
	}

	token, err := common.GenerateRandomKey(48)
	if err != nil {
		return nil, err
	}
	grant := &model.ImpersonationGrant{
		ApprovalToken:     token,
		Source:            model.ImpersonationSourceAdminRequest,
		Mode:              model.ImpersonationModeStandard,
		State:             model.ImpersonationStatePending,
		RequestedReadOnly: requestReadOnly,
		RequestedAt:       now,
	}
	fillImpersonationGrantSnapshot(grant, operator, target)
	if err = model.CreateImpersonationGrant(grant); err != nil {
		return nil, err
	}
	if err = notifyUserOfImpersonationRequest(target, grant); err != nil {
		return nil, err
	}
	return &StartImpersonationResult{State: ImpersonationStartStatePendingApproval, Grant: grant}, nil
}

func ApproveImpersonationGrant(grant *model.ImpersonationGrant, user *model.User, method string) error {
	if grant == nil || user == nil {
		return errors.New("无效的批准请求")
	}
	if grant.TargetUserId != user.Id {
		return errors.New("无权批准此请求")
	}
	now := time.Now()
	if model.ExpireImpersonationGrantIfNeeded(grant, now) {
		if err := model.SaveImpersonationGrant(grant); err != nil {
			return err
		}
	}
	if grant.State != model.ImpersonationStatePending {
		return errors.New("该请求已无法处理")
	}
	expiresAt := now.Add(ImpersonationGrantWindow)
	if err := model.ApprovePendingImpersonationGrant(grant.Id, user.Id, strings.TrimSpace(method), now, expiresAt); err != nil {
		return err
	}
	grant.State = model.ImpersonationStateApproved
	grant.ApprovedByUserId = user.Id
	grant.ApprovedByMethod = strings.TrimSpace(method)
	grant.ApprovedAt = &now
	grant.GrantedExpiresAt = &expiresAt
	if grant.OperatorId != 0 {
		if operator, err := model.GetUserById(grant.OperatorId, true); err == nil && operator != nil {
			_ = notifyOperatorOfGrantDecision(operator, grant, true)
		}
	}
	return nil
}

func RejectImpersonationGrant(grant *model.ImpersonationGrant, user *model.User, method string) error {
	if grant == nil || user == nil {
		return errors.New("无效的拒绝请求")
	}
	if grant.TargetUserId != user.Id {
		return errors.New("无权拒绝此请求")
	}
	if grant.State != model.ImpersonationStatePending {
		return errors.New("该请求已无法处理")
	}
	now := time.Now()
	if err := model.RejectPendingImpersonationGrant(grant.Id, user.Id, strings.TrimSpace(method), now); err != nil {
		return err
	}
	grant.State = model.ImpersonationStateRejected
	grant.ApprovedByUserId = user.Id
	grant.ApprovedByMethod = strings.TrimSpace(method)
	grant.ApprovedAt = &now
	if grant.OperatorId != 0 {
		if operator, err := model.GetUserById(grant.OperatorId, true); err == nil && operator != nil {
			_ = notifyOperatorOfGrantDecision(operator, grant, false)
		}
	}
	return nil
}

func OpenSelfServiceSupportAccess(user *model.User) (*model.ImpersonationGrant, error) {
	if user == nil {
		return nil, errors.New("用户不存在")
	}
	if existing, err := model.FindLatestOpenSelfServiceGrant(user.Id); err == nil && existing != nil {
		return existing, nil
	}

	now := time.Now()
	expiresAt := now.Add(ImpersonationGrantWindow)
	grant := &model.ImpersonationGrant{
		Source:           model.ImpersonationSourceSelfService,
		Mode:             model.ImpersonationModeStandard,
		State:            model.ImpersonationStateApproved,
		RequestedReadOnly: false,
		RequestedAt:      now,
		ApprovedAt:       &now,
		GrantedExpiresAt: &expiresAt,
	}
	fillImpersonationGrantSnapshot(grant, nil, user)
	grant.ApprovedByUserId = user.Id
	grant.ApprovedByMethod = "self_service"
	if err := model.CreateImpersonationGrant(grant); err != nil {
		return nil, err
	}
	return grant, nil
}

func CloseSelfServiceSupportAccess(userID int) error {
	grant, err := model.FindLatestOpenSelfServiceGrant(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	grant.State = model.ImpersonationStateCancelled
	return model.SaveImpersonationGrant(grant)
}

func StopCurrentImpersonation(session sessions.Session, preserveHeaderAlias bool) (*model.ImpersonationGrant, error) {
	if session == nil {
		return nil, nil
	}
	state := GetImpersonationSessionState(session)
	if !state.Active {
		return nil, nil
	}

	var grant *model.ImpersonationGrant
	if state.GrantID != 0 {
		if loadedGrant, err := model.GetImpersonationGrantByID(state.GrantID); err == nil {
			grant = loadedGrant
		}
	}

	now := time.Now()
	if grant != nil && grant.State == model.ImpersonationStateActive {
		grant.State = model.ImpersonationStateCompleted
		grant.EndedAt = &now
		_ = model.CompleteImpersonationGrant(grant.Id, now)
	}

	if _, err := RestoreOriginalSession(session, preserveHeaderAlias); err != nil {
		return grant, err
	}

	if grant != nil && grant.IsBreakGlass() {
		if target, err := model.GetUserById(grant.TargetUserId, true); err == nil && target != nil {
			grant.EndedAt = &now
			_ = notifyUserOfBreakGlassEnded(target, grant)
		}
	}

	return grant, nil
}

func RecordBreakGlassAction(grantID uint, operatorID int, operatorUsername string, targetUserID int, method string, path string, route string, statusCode int) error {
	if grantID == 0 || targetUserID == 0 {
		return nil
	}
	return model.CreateImpersonationActionLog(&model.ImpersonationActionLog{
		GrantId:          grantID,
		OperatorId:       operatorID,
		OperatorUsername: strings.TrimSpace(operatorUsername),
		TargetUserId:     targetUserID,
		Method:           strings.ToUpper(strings.TrimSpace(method)),
		Path:             strings.TrimSpace(path),
		Route:            strings.TrimSpace(route),
		StatusCode:       statusCode,
		Success:          statusCode >= 200 && statusCode < 400,
	})
}
