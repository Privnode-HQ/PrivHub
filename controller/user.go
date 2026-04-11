package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"

	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func buildAuthenticatedUserResponseData(user *model.User) gin.H {
	return gin.H{
		"cah_id":               user.CAHID,
		"username":             user.Username,
		"display_name":         user.DisplayName,
		"role":                 user.Role,
		"status":               user.Status,
		"group":                user.Group,
		"email":                user.Email,
		"force_password_reset": user.ForcePasswordReset,
		"force_email_bind":     user.ForceEmailBind,
		"required_actions":     user.GetRequiredActions(),
	}
}

func Login(c *gin.Context) {
	if !common.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了密码登录",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无效的参数",
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": "无效的参数",
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}

	// 检查是否启用2FA
	if model.IsTwoFAEnabled(user.Id) {
		// 设置pending session，等待2FA验证
		session := sessions.Default(c)
		session.Set("pending_username", user.Username)
		session.Set("pending_user_id", user.Id)
		err := session.Save()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "无法保存会话信息，请重试",
				"success": false,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "请输入两步验证码",
			"success": true,
			"data": map[string]interface{}{
				"require_2fa": true,
			},
		})
		return
	}

	setupLogin(&user, c)
}

// setup session & cookies and then return user info
func setupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	_, _ = service.CompleteCurrentAccessLinkGrant(session)
	service.ClearAccessLinkSession(session)
	service.ClearImpersonationSession(session)
	service.ClearImpersonationHeaderAlias(session)
	service.ApplyDefaultWebSessionOptions(session)
	service.SetAuthenticatedUserSession(session, user)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无法保存会话信息，请重试",
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    buildAuthenticatedUserResponseData(user),
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	_, _ = service.StopCurrentImpersonation(session, false)
	_, _ = service.CompleteCurrentAccessLinkGrant(session)
	session.Clear()
	service.ApplyDefaultWebSessionOptions(session)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

func Register(c *gin.Context) {
	if !common.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了新用户注册",
			"success": false,
		})
		return
	}
	if !common.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	if common.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员开启了邮箱验证，请输入邮箱地址和验证码",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "验证码错误或已过期",
			})
			return
		}
	}
	exist, err := model.CheckUserExistOrDeleted(user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "数据库错误，请稍后重试",
		})
		common.SysLog(fmt.Sprintf("CheckUserExistOrDeleted error: %v", err))
		return
	}
	if exist {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户名已存在，或已注销",
		})
		return
	}
	affCode := user.AffCode // this code is the inviter's code, not the user's own code
	inviterId, _ := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
		Role:        common.RoleCommonUser, // 明确设置角色为普通用户
	}
	if common.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	if err := cleanUser.Insert(inviterId); err != nil {
		common.ApiError(c, err)
		return
	}

	// 获取插入后的用户ID
	var insertedUser model.User
	if err := model.DB.Where("username = ?", cleanUser.Username).First(&insertedUser).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户注册失败或用户ID获取失败",
		})
		return
	}
	// 生成默认令牌
	if constant.GenerateDefaultToken {
		key, err := common.GenerateKey()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "生成默认令牌失败",
			})
			common.SysLog("failed to generate token key: " + err.Error())
			return
		}
		// 生成默认令牌
		token := model.Token{
			UserId:             insertedUser.Id, // 使用插入后的用户ID
			Name:               cleanUser.Username + "的初始令牌",
			Key:                key,
			CreatedTime:        common.GetTimestamp(),
			AccessedTime:       common.GetTimestamp(),
			ExpiredTime:        -1,     // 永不过期
			RemainQuota:        500000, // 示例额度
			UnlimitedQuota:     true,
			ModelLimitsEnabled: false,
		}
		if setting.DefaultUseAutoGroup {
			token.Group = "auto"
			token.Groups = model.TokenGroups{"auto"}
		}
		if err := token.Insert(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建默认令牌失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func GetAllUsers(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	users, total, err := model.GetAllUsers(pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(users)

	common.ApiSuccess(c, pageInfo)
	return
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	group := c.Query("group")
	pageInfo := common.GetPageQuery(c)
	users, total, err := model.SearchUsers(keyword, group, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(users)
	common.ApiSuccess(c, pageInfo)
	return
}

func SearchUserByAPIKey(c *gin.Context) {
	rawKey := c.Query("key")
	if strings.TrimSpace(rawKey) == "" {
		common.ApiErrorMsg(c, "请输入 API Key")
		return
	}

	user, token, err := model.FindUserByAPIKey(rawKey)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"user": user,
		"token": gin.H{
			"id":            token.Id,
			"user_id":       token.UserId,
			"name":          token.Name,
			"group":         token.Group,
			"status":        token.Status,
			"created_time":  token.CreatedTime,
			"accessed_time": token.AccessedTime,
			"expired_time":  token.ExpiredTime,
			"deleted":       token.DeletedAt.Valid,
		},
	})
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权获取同级或更高等级用户的信息",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
	return
}

func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	// get rand int 28-32
	randI := common.GetRandomInt(4)
	key, err := common.GenerateRandomKey(29 + randI)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "生成失败",
		})
		common.SysLog("failed to generate key: " + err.Error())
		return
	}
	user.SetAccessToken(key)

	if model.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请重试，系统生成的 UUID 竟然重复了！",
		})
		return
	}

	if err := user.Update(false); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AccessToken,
	})
	return
}

type TransferAffQuotaRequest struct {
	Quota int `json:"quota" binding:"required"`
}

func TransferAffQuota(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	tran := TransferAffQuotaRequest{}
	if err := c.ShouldBindJSON(&tran); err != nil {
		common.ApiError(c, err)
		return
	}
	err = user.TransferAffQuotaToQuota(tran.Quota)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "划转失败 " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "划转成功",
	})
}

func GetAffCode(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.AffCode == "" {
		user.AffCode = common.GetRandomString(4)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
	return
}

func buildSelfResponseData(user *model.User, userRole int) map[string]interface{} {
	permissions := calculateUserPermissions(userRole)
	userSetting := user.GetSetting()
	responseData := map[string]interface{}{
		"cah_id":                        user.CAHID,
		"username":                      user.Username,
		"display_name":                  user.DisplayName,
		"role":                          user.Role,
		"status":                        user.Status,
		"email":                         user.Email,
		"github_id":                     user.GitHubId,
		"discord_id":                    user.DiscordId,
		"oidc_id":                       user.OidcId,
		"wechat_id":                     user.WeChatId,
		"telegram_id":                   user.TelegramId,
		"group":                         user.Group,
		"quota":                         user.Quota,
		"used_quota":                    user.UsedQuota,
		"request_count":                 user.RequestCount,
		"aff_code":                      user.AffCode,
		"aff_count":                     user.AffCount,
		"aff_quota":                     user.AffQuota,
		"aff_history_quota":             user.AffHistoryQuota,
		"linux_do_id":                   user.LinuxDOId,
		"setting":                       user.Setting,
		"stripe_customer":               user.StripeCustomer,
		"sidebar_modules":               userSetting.SidebarModules, // 正确提取sidebar_modules字段
		"permissions":                   permissions,                // 新增权限字段
		"force_password_reset":          user.ForcePasswordReset,
		"force_email_bind":              user.ForceEmailBind,
		"required_actions":              user.GetRequiredActions(),
		"require_display_name_enabled":  common.RequireUserDisplayNameEnabled,
		"require_email_binding_enabled": common.RequireUserEmailBindingEnabled,
	}
	if user.InviterId != 0 {
		if inviterCAHID, inviterErr := model.GetUserCAHIDById(user.InviterId); inviterErr == nil {
			responseData["inviter_cah_id"] = inviterCAHID
		}
	}
	return responseData
}

func attachSessionStateToSelfResponse(session sessions.Session, user *model.User, responseData map[string]interface{}) {
	if session == nil || user == nil || responseData == nil {
		return
	}

	impersonationState := service.GetImpersonationSessionState(session)
	if impersonationState.Active {
		responseData["impersonation"] = gin.H{
			"active":            true,
			"grant_id":          impersonationState.GrantID,
			"read_only":         impersonationState.ReadOnly,
			"break_glass":       impersonationState.BreakGlass,
			"started_at":        impersonationState.StartedAt,
			"expires_at":        impersonationState.ExpiresAt,
			"operator_id":       impersonationState.OriginalID,
			"operator_username": impersonationState.OriginalUsername,
			"operator_cah_id":   impersonationState.OriginalCAHID,
		}
		return
	}

	accessLinkState := service.GetAccessLinkSessionState(session)
	if accessLinkState.Active {
		responseData["access_link_session"] = gin.H{
			"active":     true,
			"grant_id":   accessLinkState.GrantID,
			"expires_at": accessLinkState.ExpiresAt,
		}
	}

	grant, err := model.GetLatestBreakGlassAlertGrant(user.Id)
	if err != nil || grant == nil {
		return
	}
	now := time.Now()
	if model.ExpireImpersonationGrantIfNeeded(grant, now) {
		_ = model.SaveImpersonationGrant(grant)
	}
	responseData["break_glass_alert"] = gin.H{
		"grant_id":          grant.Id,
		"active":            grant.EndedAt == nil,
		"operator_id":       grant.OperatorId,
		"operator_username": grant.OperatorUsername,
		"started_at": func() *time.Time {
			if grant.ActivatedAt != nil {
				return grant.ActivatedAt
			}
			return &grant.RequestedAt
		}(),
		"ended_at": grant.EndedAt,
	}
}

func respondWithCurrentUser(c *gin.Context, userID int, userRole int) {
	user, err := model.GetUserById(userID, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user.Remark = ""
	responseData := buildSelfResponseData(user, userRole)
	session := sessions.Default(c)
	attachSessionStateToSelfResponse(session, user, responseData)
	sessionState := service.GetImpersonationSessionState(session)
	if !sessionState.Active && (sessionState.HeaderAliasID != 0 || sessionState.HeaderAliasCAHID != "") {
		service.ClearImpersonationHeaderAlias(session)
		_ = session.Save()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    responseData,
	})
}

func GetSelf(c *gin.Context) {
	respondWithCurrentUser(c, c.GetInt("id"), c.GetInt("role"))
	return
}

// 计算用户权限的辅助函数
func calculateUserPermissions(userRole int) map[string]interface{} {
	permissions := map[string]interface{}{}

	// 根据用户角色计算权限
	if userRole == common.RoleRootUser {
		// 超级管理员不需要边栏设置功能
		permissions["sidebar_settings"] = false
		permissions["sidebar_modules"] = map[string]interface{}{}
	} else if userRole == common.RoleSupportUser {
		// 支持人员拥有后台只读权限，但不能查看渠道和系统设置
		permissions["sidebar_settings"] = true
		permissions["sidebar_modules"] = map[string]interface{}{
			"admin": map[string]interface{}{
				"channel":        false,
				"message_manage": true,
				"models":         true,
				"redemption":     true,
				"topup_coupon":   true,
				"user":           true,
				"user_api_key":   true,
				"setting":        false,
			},
		}
	} else if userRole == common.RoleAdminUser {
		// 管理员可以设置边栏，但不包含系统设置功能
		permissions["sidebar_settings"] = true
		permissions["sidebar_modules"] = map[string]interface{}{
			"admin": map[string]interface{}{
				"user_api_key": true,
				"setting":      false, // 管理员不能访问系统设置
			},
		}
	} else {
		// 普通用户只能设置个人功能，不包含管理员区域
		permissions["sidebar_settings"] = true
		permissions["sidebar_modules"] = map[string]interface{}{
			"admin": false, // 普通用户不能访问管理员区域
		}
	}

	return permissions
}

// 根据用户角色生成默认的边栏配置
func generateDefaultSidebarConfig(userRole int) string {
	defaultConfig := map[string]interface{}{}

	// 聊天区域 - 所有用户都可以访问
	defaultConfig["chat"] = map[string]interface{}{
		"enabled":    true,
		"playground": true,
		"chat":       true,
	}

	// 控制台区域 - 所有用户都可以访问
	defaultConfig["console"] = map[string]interface{}{
		"enabled":    true,
		"detail":     true,
		"token":      true,
		"log":        true,
		"usage":      true,
		"midjourney": true,
		"task":       true,
	}

	// 个人中心区域 - 所有用户都可以访问
	defaultConfig["personal"] = map[string]interface{}{
		"enabled":  true,
		"topup":    true,
		"personal": true,
		"message":  true,
		"support":  true,
	}

	// 管理员区域 - 根据角色决定
	if userRole == common.RoleSupportUser {
		defaultConfig["admin"] = map[string]interface{}{
			"enabled":        true,
			"channel":        false,
			"message_manage": true,
			"models":         true,
			"redemption":     true,
			"topup_coupon":   true,
			"user":           true,
			"user_api_key":   true,
			"setting":        false,
		}
	} else if userRole == common.RoleAdminUser {
		// 管理员可以访问管理员区域，但不能访问系统设置
		defaultConfig["admin"] = map[string]interface{}{
			"enabled":        true,
			"channel":        true,
			"message_manage": true,
			"models":         true,
			"redemption":     true,
			"topup_coupon":   true,
			"user":           true,
			"user_api_key":   true,
			"setting":        false, // 管理员不能访问系统设置
		}
	} else if userRole == common.RoleRootUser {
		// 超级管理员可以访问所有功能
		defaultConfig["admin"] = map[string]interface{}{
			"enabled":        true,
			"channel":        true,
			"message_manage": true,
			"models":         true,
			"redemption":     true,
			"topup_coupon":   true,
			"user":           true,
			"user_api_key":   true,
			"setting":        true,
		}
	}
	// 普通用户不包含admin区域

	// 转换为JSON字符串
	configBytes, err := json.Marshal(defaultConfig)
	if err != nil {
		common.SysLog("生成默认边栏配置失败: " + err.Error())
		return ""
	}

	return string(configBytes)
}

func GetUserModels(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		id = c.GetInt("id")
	}
	user, err := model.GetUserCache(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	groups := service.GetUserUsableGroups(user.Group)
	var models []string
	for group := range groups {
		for _, g := range model.GetGroupEnabledModels(group) {
			if !common.StringsContains(models, g) {
				models = append(models, g)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
	return
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	if myRole <= updatedUser.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权将其他用户权限等级提升到大于等于自己的权限等级",
		})
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	if err := updatedUser.Edit(updatePassword); err != nil {
		common.ApiError(c, err)
		return
	}
	if updatePassword {
		originUser.WebSessionVersion++
		if err := originUser.UpdateSelected("WebSessionVersion"); err != nil {
			common.ApiError(c, err)
			return
		}
	}
	if originUser.Quota != updatedUser.Quota {
		model.RecordLog(originUser.Id, model.LogTypeManage, fmt.Sprintf("管理员将用户额度从 %s修改为 %s", logger.LogQuota(int64(originUser.Quota)), logger.LogQuota(int64(updatedUser.Quota))))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateSelf(c *gin.Context) {
	var requestData map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&requestData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// 检查是否是sidebar_modules更新请求
	if sidebarModules, exists := requestData["sidebar_modules"]; exists {
		userId := c.GetInt("id")
		user, err := model.GetUserById(userId, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}

		// 获取当前用户设置
		currentSetting := user.GetSetting()

		// 更新sidebar_modules字段
		if sidebarModulesStr, ok := sidebarModules.(string); ok {
			currentSetting.SidebarModules = sidebarModulesStr
		}

		// 保存更新后的设置
		user.SetSetting(currentSetting)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "更新设置失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "设置更新成功",
		})
		return
	}

	// 原有的用户信息更新逻辑
	var user model.User
	requestDataBytes, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	err = json.Unmarshal(requestDataBytes, &user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt("id"),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword, err := checkUpdatePassword(user.OriginalPassword, user.Password, cleanUser.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := cleanUser.Update(updatePassword); err != nil {
		common.ApiError(c, err)
		return
	}
	if updatePassword {
		currentUser, err := model.GetUserById(cleanUser.Id, false)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		currentUser.ForcePasswordReset = false
		currentUser.WebSessionVersion++
		if err := currentUser.UpdateSelected("ForcePasswordReset", "WebSessionVersion"); err != nil {
			common.ApiError(c, err)
			return
		}
		if !c.GetBool("use_access_token") {
			session := sessions.Default(c)
			session.Set("session_version", currentUser.WebSessionVersion)
			if err := session.Save(); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "无法保存会话信息，请重试",
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func BackToPayAsYouGo(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.Group != "subscription" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "当前账户不是订阅分组，无需操作",
		})
		return
	}

	user.Group = "default"
	if err := user.Update(false); err != nil {
		common.ApiError(c, err)
		return
	}

	session := sessions.Default(c)
	if session.Get("id") != nil {
		session.Set("group", user.Group)
		_ = session.Save()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func checkUpdatePassword(originalPassword string, newPassword string, userId int) (updatePassword bool, err error) {
	if newPassword == "" {
		return false, nil
	}
	var currentUser *model.User
	currentUser, err = model.GetUserById(userId, true)
	if err != nil {
		return
	}
	if !common.ValidatePasswordAndHash(originalPassword, currentUser.Password) {
		err = fmt.Errorf("原密码错误")
		return
	}
	if newPassword == "" {
		return
	}
	updatePassword = true
	return
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权删除同权限等级或更高权限等级的用户",
		})
		return
	}
	err = model.HardDeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
		return
	}
}

func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, _ := model.GetUserById(id, false)

	if user.Role == common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不能删除超级管理员账户",
		})
		return
	}

	err := model.DeleteUserById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func CreateUser(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	user.Username = strings.TrimSpace(user.Username)
	if err != nil || user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role == common.RoleGuestUser {
		user.Role = common.RoleCommonUser
	}
	if !common.IsValidateRole(user.Role) || user.Role == common.RoleGuestUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的用户角色",
		})
		return
	}
	if user.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法创建权限大于等于自己的用户",
		})
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
		Role:        user.Role, // 保持管理员设置的角色
	}
	if err := cleanUser.Insert(0); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type ManageRequest struct {
	Id        int    `json:"id"`
	Action    string `json:"action"`
	BanReason string `json:"ban_reason"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	user := model.User{
		Id: req.Id,
	}
	// Fill attributes
	model.DB.Unscoped().Where(&user).First(&user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != common.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	switch req.Action {
	case "disable":
		user.Status = common.UserStatusDisabled
		user.BanReason = strings.TrimSpace(req.BanReason)
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法禁用超级管理员用户",
			})
			return
		}
	case "enable":
		user.Status = common.UserStatusEnabled
	case "delete":
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法删除超级管理员用户",
			})
			return
		}
		if err := user.Delete(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		switch user.Role {
		case common.RoleCommonUser:
			user.Role = common.RoleSupportUser
		case common.RoleSupportUser:
			if myRole != common.RoleRootUser {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "普通管理员用户无法提升支持人员为管理员",
				})
				return
			}
			user.Role = common.RoleAdminUser
		default:
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是管理员",
			})
			return
		}
	case "demote":
		if user.Role == common.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法降级超级管理员用户",
			})
			return
		}
		if user.Role == common.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是普通用户",
			})
			return
		}
		if user.Role == common.RoleAdminUser {
			user.Role = common.RoleSupportUser
		} else {
			user.Role = common.RoleCommonUser
		}
	case "logout":
		user.WebSessionVersion++
	case "require_password_reset":
		user.ForcePasswordReset = true
	case "require_email_bind":
		user.ForceEmailBind = true
	}

	if req.Action == "disable" || req.Action == "enable" {
		if err := user.UpdateSelected("Status", "BanReason"); err != nil {
			common.ApiError(c, err)
			return
		}
	} else if req.Action == "logout" {
		if err := user.UpdateSelected("WebSessionVersion"); err != nil {
			common.ApiError(c, err)
			return
		}
	} else if req.Action == "require_password_reset" {
		if err := user.UpdateSelected("ForcePasswordReset"); err != nil {
			common.ApiError(c, err)
			return
		}
	} else if req.Action == "require_email_bind" {
		if err := user.UpdateSelected("ForceEmailBind"); err != nil {
			common.ApiError(c, err)
			return
		}
	} else if err := user.Update(false); err != nil {
		common.ApiError(c, err)
		return
	}
	clearUser := model.User{
		Role:               user.Role,
		Status:             user.Status,
		BanReason:          user.BanReason,
		ForcePasswordReset: user.ForcePasswordReset,
		ForceEmailBind:     user.ForceEmailBind,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
	return
}

func LogoutAllUsers(c *gin.Context) {
	nextVersion := common.GlobalWebSessionVersion + 1
	if err := model.UpdateOption("GlobalWebSessionVersion", strconv.Itoa(nextVersion)); err != nil {
		common.ApiError(c, err)
		return
	}

	if !c.GetBool("use_access_token") {
		session := sessions.Default(c)
		session.Set("global_session_version", common.GlobalWebSessionVersion)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法保存会话信息，请重试",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}
	id, err := getValidatedSessionUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		Id: id,
	}
	err = user.FillUserById()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user.Email = email
	user.ForceEmailBind = false
	err = user.UpdateSelected("Email", "ForceEmailBind")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type topUpRequest struct {
	Key string `json:"key"`
}

var topUpLocks sync.Map
var topUpCreateLock sync.Mutex

type topUpTryLock struct {
	ch chan struct{}
}

func newTopUpTryLock() *topUpTryLock {
	return &topUpTryLock{ch: make(chan struct{}, 1)}
}

func (l *topUpTryLock) TryLock() bool {
	select {
	case l.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *topUpTryLock) Unlock() {
	select {
	case <-l.ch:
	default:
	}
}

func getTopUpLock(userID int) *topUpTryLock {
	if v, ok := topUpLocks.Load(userID); ok {
		return v.(*topUpTryLock)
	}
	topUpCreateLock.Lock()
	defer topUpCreateLock.Unlock()
	if v, ok := topUpLocks.Load(userID); ok {
		return v.(*topUpTryLock)
	}
	l := newTopUpTryLock()
	topUpLocks.Store(userID, l)
	return l
}

func TopUp(c *gin.Context) {
	id := c.GetInt("id")
	lock := getTopUpLock(id)
	if !lock.TryLock() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "充值处理中，请稍后重试",
		})
		return
	}
	defer lock.Unlock()
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	quota, err := model.Redeem(req.Key, id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    quota,
	})
}

type UpdateUserSettingRequest struct {
	QuotaWarningType           string  `json:"notify_type"`
	QuotaWarningThreshold      float64 `json:"quota_warning_threshold"`
	WebhookUrl                 string  `json:"webhook_url,omitempty"`
	WebhookSecret              string  `json:"webhook_secret,omitempty"`
	NotificationEmail          string  `json:"notification_email,omitempty"`
	BarkUrl                    string  `json:"bark_url,omitempty"`
	GotifyUrl                  string  `json:"gotify_url,omitempty"`
	GotifyToken                string  `json:"gotify_token,omitempty"`
	GotifyPriority             int     `json:"gotify_priority,omitempty"`
	AcceptUnsetModelRatioModel bool    `json:"accept_unset_model_ratio_model"`
	RecordIpLog                bool    `json:"record_ip_log"`
}

func UpdateUserSetting(c *gin.Context) {
	var req UpdateUserSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	// 验证预警类型，仅保留邮件通知
	if req.QuotaWarningType != "" && req.QuotaWarningType != dto.NotifyTypeEmail {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "当前仅支持邮件通知",
		})
		return
	}

	// 验证预警阈值
	if req.QuotaWarningThreshold <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "预警阈值必须大于0",
		})
		return
	}

	// 验证邮箱地址
	if req.NotificationEmail != "" {
		// 验证邮箱格式
		if !strings.Contains(req.NotificationEmail, "@") {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无效的邮箱地址",
			})
			return
		}
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 构建设置
	settings := dto.UserSetting{
		NotifyType:            dto.NotifyTypeEmail,
		QuotaWarningThreshold: req.QuotaWarningThreshold,
		AcceptUnsetRatioModel: req.AcceptUnsetModelRatioModel,
		RecordIpLog:           req.RecordIpLog,
	}

	// 如果提供了通知邮箱，添加到设置中
	if req.NotificationEmail != "" {
		settings.NotificationEmail = req.NotificationEmail
	}

	// 更新用户设置
	user.SetSetting(settings)
	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "更新设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已更新",
	})
}
