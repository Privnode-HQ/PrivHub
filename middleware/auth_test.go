package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminServiceAccountAuthTestDB(t *testing.T) {
	t.Helper()

	originalDB := model.DB
	originalCryptoSecret := common.CryptoSecret
	originalJWTSecret := common.AdminServiceAccountJWTSecret
	t.Cleanup(func() {
		model.DB = originalDB
		common.CryptoSecret = originalCryptoSecret
		common.AdminServiceAccountJWTSecret = originalJWTSecret
	})

	common.CryptoSecret = "test-asa-auth-hmac-secret"
	common.AdminServiceAccountJWTSecret = "test-asa-auth-jwt-secret"

	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.User{}, &model.AdminServiceAccount{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	model.DB = db
}

func TestAdminServiceAccountBearerSkipsNewAPIUserHeader(t *testing.T) {
	setupAdminServiceAccountAuthTestDB(t)
	gin.SetMode(gin.TestMode)

	admin := &model.User{
		Username:    "asa-admin",
		DisplayName: "ASA Admin",
		Password:    "irrelevant",
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	if err := model.DB.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	account := &model.AdminServiceAccount{
		Name:              "asa-auth-test",
		UserID:            admin.Id,
		Username:          admin.Username,
		UserCAHID:         admin.CAHID,
		UserRole:          admin.Role,
		CreatedByID:       admin.Id,
		CreatedByUsername: admin.Username,
		CreatedByCAHID:    admin.CAHID,
		ExpiresAt:         time.Now().Add(24 * time.Hour).Unix(),
	}
	credential, err := model.CreateAdminServiceAccount(account, admin)
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}

	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("test-session-secret"))))
	router.GET("/admin-only", AdminAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":                        c.GetInt("id"),
			"username":                  c.GetString("username"),
			"use_admin_service_account": c.GetBool("use_admin_service_account"),
			"service_account_id":        c.GetString("admin_service_account_id"),
		})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	request.Header.Set("Authorization", "Bearer "+credential)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"use_admin_service_account":true`) {
		t.Fatalf("expected ASA auth context, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), account.ServiceAccountID) {
		t.Fatalf("expected service account id in context, got %s", recorder.Body.String())
	}
}
