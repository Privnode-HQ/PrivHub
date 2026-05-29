package model

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	AdminServiceAccountStatusEnabled  = 1
	AdminServiceAccountStatusDisabled = 2

	AdminServiceAccountTokenUse = "admin_service_account"
	AdminServiceAccountIssuer   = "privhub"
	AdminServiceAccountAudience = "privhub-admin-api"
)

const (
	adminServiceAccountIDPrefix = "asa_"
	adminServiceAccountVersion  = "v1"
)

type AdminServiceAccount struct {
	Id                int            `json:"id"`
	ServiceAccountID  string         `json:"service_account_id" gorm:"size:64;uniqueIndex"`
	Name              string         `json:"name" gorm:"size:80;index"`
	Description       string         `json:"description" gorm:"size:255;default:''"`
	UserID            int            `json:"user_id" gorm:"column:user_id;index"`
	Username          string         `json:"username" gorm:"size:64;index"`
	UserCAHID         string         `json:"user_cah_id" gorm:"column:user_cah_id;size:16;index"`
	UserRole          int            `json:"user_role" gorm:"column:user_role;index"`
	CreatedByID       int            `json:"created_by_id" gorm:"column:created_by_id;index"`
	CreatedByUsername string         `json:"created_by_username" gorm:"size:64;index"`
	CreatedByCAHID    string         `json:"created_by_cah_id" gorm:"column:created_by_cah_id;size:16;index"`
	Status            int            `json:"status" gorm:"type:int;default:1;index"`
	JWTID             string         `json:"-" gorm:"column:jti;size:64;uniqueIndex"`
	CredentialHash    string         `json:"-" gorm:"column:credential_hash;size:64;uniqueIndex"`
	Scopes            string         `json:"scopes" gorm:"size:255;default:'admin:api'"`
	AllowIps          *string        `json:"allow_ips" gorm:"type:text"`
	CreatedTime       int64          `json:"created_time" gorm:"bigint;index"`
	AccessedTime      int64          `json:"accessed_time" gorm:"bigint;default:0;index"`
	ExpiresAt         int64          `json:"expires_at" gorm:"bigint;index"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type AdminServiceAccountClaims struct {
	TokenUse          string   `json:"token_use"`
	Version           string   `json:"ver"`
	ServiceAccountID  string   `json:"service_account_id"`
	ServiceName       string   `json:"service_account_name"`
	UserID            int      `json:"uid"`
	Username          string   `json:"username"`
	UserCAHID         string   `json:"cah_id"`
	UserRole          int      `json:"role"`
	CreatedByID       int      `json:"created_by_id"`
	CreatedByUsername string   `json:"created_by_username"`
	CreatedByCAHID    string   `json:"created_by_cah_id"`
	Scopes            []string `json:"scopes"`
	jwt.RegisteredClaims
}

func (account *AdminServiceAccount) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(account.ServiceAccountID) == "" {
		id, err := GenerateAdminServiceAccountPublicID()
		if err != nil {
			return err
		}
		account.ServiceAccountID = id
	}
	if strings.TrimSpace(account.JWTID) == "" {
		jti, err := GenerateAdminServiceAccountJWTID()
		if err != nil {
			return err
		}
		account.JWTID = jti
	}
	if account.Status == 0 {
		account.Status = AdminServiceAccountStatusEnabled
	}
	if strings.TrimSpace(account.Scopes) == "" {
		account.Scopes = "admin:api"
	}
	if account.CreatedTime == 0 {
		account.CreatedTime = common.GetTimestamp()
	}
	return nil
}

func GenerateAdminServiceAccountPublicID() (string, error) {
	token, err := common.GenerateURLSafeToken(18)
	if err != nil {
		return "", err
	}
	return adminServiceAccountIDPrefix + token, nil
}

func GenerateAdminServiceAccountJWTID() (string, error) {
	return common.GenerateURLSafeToken(24)
}

func NormalizeAdminServiceAccountName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("Service Account 名称不能为空")
	}
	if len([]rune(name)) > 80 {
		return "", errors.New("Service Account 名称不能超过 80 个字符")
	}
	return name, nil
}

func NormalizeAdminServiceAccountDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if len([]rune(description)) > 255 {
		return "", errors.New("Service Account 说明不能超过 255 个字符")
	}
	return description, nil
}

func NormalizeAdminServiceAccountAllowIps(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}

	parts := strings.FieldsFunc(*raw, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ';'
	})
	normalized := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if len(normalized) >= 100 {
			return nil, errors.New("IP 白名单最多支持 100 条")
		}
		if parsed := net.ParseIP(value); parsed != nil {
			value = parsed.String()
		} else {
			_, ipNet, err := net.ParseCIDR(value)
			if err != nil || ipNet == nil {
				return nil, fmt.Errorf("IP 白名单包含无效条目：%s", value)
			}
			value = ipNet.String()
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		empty := ""
		return &empty, nil
	}
	value := strings.Join(normalized, "\n")
	return &value, nil
}

func IsAdminServiceAccountStatusValid(status int) bool {
	return status == AdminServiceAccountStatusEnabled || status == AdminServiceAccountStatusDisabled
}

func AdminServiceAccountScopesFromString(scopes string) []string {
	result := make([]string, 0)
	for _, item := range strings.Split(scopes, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return []string{"admin:api"}
	}
	return result
}

func SignAdminServiceAccountJWT(account *AdminServiceAccount, target *User) (string, error) {
	if account == nil || target == nil {
		return "", errors.New("Service Account 或目标用户为空")
	}
	if common.AdminServiceAccountJWTSecret == "" {
		return "", errors.New("ADMIN_SERVICE_ACCOUNT_JWT_SECRET 未初始化")
	}
	if account.JWTID == "" {
		jti, err := GenerateAdminServiceAccountJWTID()
		if err != nil {
			return "", err
		}
		account.JWTID = jti
	}
	if account.ExpiresAt <= common.GetTimestamp() {
		return "", errors.New("Service Account 过期时间无效")
	}

	now := time.Now()
	claims := AdminServiceAccountClaims{
		TokenUse:          AdminServiceAccountTokenUse,
		Version:           adminServiceAccountVersion,
		ServiceAccountID:  account.ServiceAccountID,
		ServiceName:       account.Name,
		UserID:            target.Id,
		Username:          target.Username,
		UserCAHID:         target.CAHID,
		UserRole:          target.Role,
		CreatedByID:       account.CreatedByID,
		CreatedByUsername: account.CreatedByUsername,
		CreatedByCAHID:    account.CreatedByCAHID,
		Scopes:            AdminServiceAccountScopesFromString(account.Scopes),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        account.JWTID,
			Issuer:    AdminServiceAccountIssuer,
			Subject:   "admin:" + strconv.Itoa(target.Id),
			Audience:  jwt.ClaimStrings{AdminServiceAccountAudience},
			ExpiresAt: jwt.NewNumericDate(time.Unix(account.ExpiresAt, 0)),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = "JWT"
	token.Header["kid"] = "asa-v1"
	return token.SignedString([]byte(common.AdminServiceAccountJWTSecret))
}

func CreateAdminServiceAccount(account *AdminServiceAccount, target *User) (string, error) {
	if account == nil || target == nil {
		return "", errors.New("Service Account 或目标用户为空")
	}
	if err := account.BeforeCreate(nil); err != nil {
		return "", err
	}
	credential, err := SignAdminServiceAccountJWT(account, target)
	if err != nil {
		return "", err
	}
	account.CredentialHash = common.GenerateHMAC(credential)
	if err = DB.Create(account).Error; err != nil {
		return "", err
	}
	return credential, nil
}

func RotateAdminServiceAccountCredential(account *AdminServiceAccount, target *User, expiresAt int64) (string, error) {
	if account == nil || target == nil {
		return "", errors.New("Service Account 或目标用户为空")
	}
	jti, err := GenerateAdminServiceAccountJWTID()
	if err != nil {
		return "", err
	}
	account.JWTID = jti
	account.ExpiresAt = expiresAt
	account.Status = AdminServiceAccountStatusEnabled
	credential, err := SignAdminServiceAccountJWT(account, target)
	if err != nil {
		return "", err
	}
	account.CredentialHash = common.GenerateHMAC(credential)
	if err = DB.Model(account).Select("jti", "credential_hash", "expires_at", "status").Updates(account).Error; err != nil {
		return "", err
	}
	return credential, nil
}

func GetAdminServiceAccountByID(id int) (*AdminServiceAccount, error) {
	if id == 0 {
		return nil, errors.New("Service Account ID 为空")
	}
	account := &AdminServiceAccount{}
	err := DB.First(account, "id = ?", id).Error
	return account, err
}

func GetAdminServiceAccountByJWTID(jti string) (*AdminServiceAccount, error) {
	jti = strings.TrimSpace(jti)
	if jti == "" {
		return nil, errors.New("Service Account JWT ID 为空")
	}
	account := &AdminServiceAccount{}
	err := DB.First(account, "jti = ?", jti).Error
	return account, err
}

func CountAdminServiceAccounts(keyword string, maxRole int) (int64, error) {
	var total int64
	tx := DB.Model(&AdminServiceAccount{})
	tx = applyAdminServiceAccountSearch(tx, keyword)
	tx = applyAdminServiceAccountRoleFilter(tx, maxRole)
	err := tx.Count(&total).Error
	return total, err
}

func GetAdminServiceAccounts(keyword string, maxRole int, startIdx int, num int) ([]*AdminServiceAccount, error) {
	var accounts []*AdminServiceAccount
	tx := DB.Model(&AdminServiceAccount{})
	tx = applyAdminServiceAccountSearch(tx, keyword)
	tx = applyAdminServiceAccountRoleFilter(tx, maxRole)
	err := tx.Order("id desc").Limit(num).Offset(startIdx).Find(&accounts).Error
	return accounts, err
}

func applyAdminServiceAccountSearch(tx *gorm.DB, keyword string) *gorm.DB {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return tx
	}
	pattern := "%" + keyword + "%"
	return tx.Where(
		"name LIKE ? OR service_account_id LIKE ? OR username LIKE ? OR user_cah_id LIKE ? OR created_by_username LIKE ?",
		pattern,
		pattern,
		pattern,
		pattern,
		pattern,
	)
}

func applyAdminServiceAccountRoleFilter(tx *gorm.DB, maxRole int) *gorm.DB {
	if maxRole <= 0 {
		return tx
	}
	return tx.Where("user_role <= ?", maxRole)
}

func (account *AdminServiceAccount) UpdateMetadata() error {
	if account == nil || account.Id == 0 {
		return errors.New("Service Account ID 为空")
	}
	return DB.Model(account).Select("name", "description", "status", "allow_ips").Updates(account).Error
}

func (account *AdminServiceAccount) Delete() error {
	if account == nil || account.Id == 0 {
		return errors.New("Service Account ID 为空")
	}
	return DB.Delete(account).Error
}

func (account *AdminServiceAccount) TouchAccessedAt(now int64) error {
	if account == nil || account.Id == 0 {
		return nil
	}
	return DB.Model(account).Update("accessed_time", now).Error
}

func ValidateAdminServiceAccountJWT(rawToken string, now time.Time) (*User, *AdminServiceAccount, *AdminServiceAccountClaims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, nil, nil, errors.New("未提供 Admin Service Account JWT")
	}
	if common.AdminServiceAccountJWTSecret == "" {
		return nil, nil, nil, errors.New("ADMIN_SERVICE_ACCOUNT_JWT_SECRET 未初始化")
	}

	claims := &AdminServiceAccountClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(AdminServiceAccountIssuer),
		jwt.WithAudience(AdminServiceAccountAudience),
		jwt.WithLeeway(30*time.Second),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	parsed, err := parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(common.AdminServiceAccountJWTSecret), nil
	})
	if err != nil || parsed == nil || !parsed.Valid {
		if err == nil {
			err = errors.New("JWT 校验失败")
		}
		return nil, nil, nil, err
	}
	if claims.TokenUse != AdminServiceAccountTokenUse {
		return nil, nil, nil, errors.New("JWT 类型不是 Admin Service Account")
	}
	if claims.Version != adminServiceAccountVersion {
		return nil, nil, nil, errors.New("Admin Service Account JWT 版本不支持")
	}
	if strings.TrimSpace(claims.ID) == "" || strings.TrimSpace(claims.ServiceAccountID) == "" {
		return nil, nil, nil, errors.New("JWT 缺少 Service Account 标识")
	}

	account, err := GetAdminServiceAccountByJWTID(claims.ID)
	if err != nil {
		return nil, nil, nil, errors.New("Admin Service Account 不存在或已被撤销")
	}
	if account.ServiceAccountID != claims.ServiceAccountID {
		return nil, nil, nil, errors.New("JWT 与 Service Account 不匹配")
	}
	if account.Status != AdminServiceAccountStatusEnabled {
		return nil, nil, nil, errors.New("Admin Service Account 已停用")
	}
	if account.ExpiresAt <= now.Unix() {
		return nil, nil, nil, errors.New("Admin Service Account 已过期")
	}
	if account.CredentialHash == "" || subtle.ConstantTimeCompare([]byte(account.CredentialHash), []byte(common.GenerateHMAC(rawToken))) != 1 {
		return nil, nil, nil, errors.New("Admin Service Account 凭据已轮换")
	}

	user, err := GetUserById(account.UserID, false)
	if err != nil {
		return nil, nil, nil, errors.New("Admin Service Account 绑定的管理员不存在")
	}
	if user.Status != common.UserStatusEnabled {
		return nil, nil, nil, errors.New(common.UserBannedMessage(user.BanReason))
	}
	if user.Role < common.RoleAdminUser {
		return nil, nil, nil, errors.New("Admin Service Account 绑定用户不再是管理员")
	}
	if account.UserCAHID != "" && NormalizeCAHID(account.UserCAHID) != NormalizeCAHID(user.CAHID) {
		return nil, nil, nil, errors.New("Admin Service Account 绑定用户标识已变更")
	}

	return user, account, claims, nil
}

func IsClientIPAllowedByAdminServiceAccount(account *AdminServiceAccount, clientIP string) bool {
	if account == nil || account.AllowIps == nil || strings.TrimSpace(*account.AllowIps) == "" {
		return true
	}
	parsedClientIP := net.ParseIP(strings.TrimSpace(clientIP))
	if parsedClientIP == nil {
		return false
	}
	for _, item := range strings.Fields(*account.AllowIps) {
		if item == "" {
			continue
		}
		if ip := net.ParseIP(item); ip != nil {
			if ip.Equal(parsedClientIP) {
				return true
			}
			continue
		}
		_, ipNet, err := net.ParseCIDR(item)
		if err == nil && ipNet.Contains(parsedClientIP) {
			return true
		}
	}
	return false
}
