package controller

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

const (
	adminUserVerificationCodeLength           = 8
	adminUserVerificationCodePurposeMaxLength = 120
)

var adminUserVerificationCodeForbiddenPurposeTerms = []string{
	"登录",
	"登陆",
	"login",
	"log in",
	"sign in",
	"signin",
	"sign-in",
	"密码",
	"password",
	"reset",
	"注册",
	"register",
	"邮箱绑定",
	"绑定邮箱",
	"邮箱验证",
	"email bind",
	"email binding",
	"email verification",
	"2fa",
	"two-factor",
	"two factor",
	"两步验证",
	"二步验证",
}

type adminUserVerificationCodeRequest struct {
	Purpose string `json:"purpose"`
}

func SendAdminUserVerificationCode(c *gin.Context) {
	var req adminUserVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "无效的参数")
		return
	}

	purpose, err := normalizeAdminUserVerificationCodePurpose(req.Purpose)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	target, err := model.GetUserByCAHID(c.Param("cah_id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if c.GetInt("role") <= target.Role && c.GetInt("role") != common.RoleRootUser {
		common.ApiErrorMsg(c, "无权向同级或更高权限用户发送验证码")
		return
	}

	email := strings.TrimSpace(target.Email)
	if email == "" {
		common.ApiErrorMsg(c, "目标用户未绑定邮箱")
		return
	}
	if err := common.Validate.Var(email, "email"); err != nil {
		common.ApiErrorMsg(c, "目标用户邮箱格式无效")
		return
	}

	model.SetAdminAuditMeta(c, model.AdminAuditMeta{
		Resource:   "user",
		Action:     "send_verification_code",
		TargetType: "user",
		TargetId:   target.Id,
		TargetName: target.Username,
		Content:    "管理员已向用户邮箱发送人工核验验证码",
		Details: map[string]interface{}{
			"cah_id":  target.CAHID,
			"email":   email,
			"purpose": purpose,
		},
	})

	code, err := common.GenerateHexVerificationCode(adminUserVerificationCodeLength)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	systemName := resolveAdminUserVerificationCodeSystemName(common.SystemName)
	subject := fmt.Sprintf("%s人工核验验证码", systemName)
	content := buildAdminUserVerificationCodeEmailContent(systemName, purpose, code)
	err = common.SendEmailWithIdempotencyKeyAndContext(
		subject,
		email,
		content,
		common.GenerateEmailIdempotencyKey("admin-user-verification-code", target.CAHID, purpose, code),
		common.EmailRecipientContext{
			Username:      target.Username,
			RecipientName: target.DisplayName,
			CAHID:         target.CAHID,
			Email:         email,
		},
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送",
		"data": gin.H{
			"cah_id":  target.CAHID,
			"email":   email,
			"purpose": purpose,
			"code":    code,
		},
	})
}

func normalizeAdminUserVerificationCodePurpose(raw string) (string, error) {
	purpose := strings.TrimSpace(raw)
	if purpose == "" {
		return "", errors.New("验证码用途不能为空")
	}
	if utf8.RuneCountInString(purpose) > adminUserVerificationCodePurposeMaxLength {
		return "", fmt.Errorf("验证码用途不能超过 %d 个字符", adminUserVerificationCodePurposeMaxLength)
	}
	for _, r := range purpose {
		if r == '\n' || r == '\r' || r == '\t' || unicode.IsControl(r) {
			return "", errors.New("验证码用途不能包含换行、制表符或控制字符")
		}
	}

	lowerPurpose := strings.ToLower(purpose)
	for _, term := range adminUserVerificationCodeForbiddenPurposeTerms {
		if strings.Contains(lowerPurpose, strings.ToLower(term)) {
			return "", errors.New("验证码用途不能是登录、注册、密码重置、邮箱绑定或两步验证等身份认证用途")
		}
	}

	return purpose, nil
}

func buildAdminUserVerificationCodeEmailContent(systemName string, purpose string, code string) string {
	systemName = resolveAdminUserVerificationCodeSystemName(systemName)

	return fmt.Sprintf(
		"感谢您选择 Privnode，您的 %s 一次性验证代码是：\n\n ## `%s` \n\n 请勿与他人分享此验证代码。",
		html.EscapeString(strings.TrimSpace(purpose)),
		strings.ToUpper(strings.TrimSpace(code)),
	)
}

func resolveAdminUserVerificationCodeSystemName(systemName string) string {
	systemName = strings.TrimSpace(systemName)
	if systemName == "" {
		return "PrivHub"
	}
	return systemName
}
