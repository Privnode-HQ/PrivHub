package model

import (
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminServiceAccountTestDB(t *testing.T) {
	t.Helper()

	originalDB := DB
	originalCryptoSecret := common.CryptoSecret
	originalJWTSecret := common.AdminServiceAccountJWTSecret
	t.Cleanup(func() {
		DB = originalDB
		common.CryptoSecret = originalCryptoSecret
		common.AdminServiceAccountJWTSecret = originalJWTSecret
	})

	common.CryptoSecret = "test-asa-hmac-secret"
	common.AdminServiceAccountJWTSecret = "test-asa-jwt-secret"

	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&User{}, &AdminServiceAccount{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	DB = db
}

func TestAdminServiceAccountJWTContainsAdminClaimsAndValidates(t *testing.T) {
	setupAdminServiceAccountTestDB(t)

	admin := &User{
		Username:    "admin",
		DisplayName: "Admin",
		Password:    "irrelevant",
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	if err := DB.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	account := &AdminServiceAccount{
		Name:              "ci-bot",
		Description:       "CI automation",
		UserID:            admin.Id,
		Username:          admin.Username,
		UserCAHID:         admin.CAHID,
		UserRole:          admin.Role,
		CreatedByID:       admin.Id,
		CreatedByUsername: admin.Username,
		CreatedByCAHID:    admin.CAHID,
		ExpiresAt:         time.Now().Add(24 * time.Hour).Unix(),
	}
	credential, err := CreateAdminServiceAccount(account, admin)
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}
	if credential == "" {
		t.Fatal("expected JWT credential")
	}
	if account.CredentialHash == "" || strings.Contains(account.CredentialHash, credential) {
		t.Fatalf("expected stored credential hash, got %q", account.CredentialHash)
	}

	user, validatedAccount, claims, err := ValidateAdminServiceAccountJWT(credential, time.Now())
	if err != nil {
		t.Fatalf("validate service account: %v", err)
	}
	if user.Id != admin.Id {
		t.Fatalf("expected user %d, got %d", admin.Id, user.Id)
	}
	if validatedAccount.ServiceAccountID != account.ServiceAccountID {
		t.Fatalf("expected service account %s, got %s", account.ServiceAccountID, validatedAccount.ServiceAccountID)
	}
	if claims.TokenUse != AdminServiceAccountTokenUse {
		t.Fatalf("expected token_use %s, got %s", AdminServiceAccountTokenUse, claims.TokenUse)
	}
	if claims.UserID != admin.Id || claims.UserCAHID != admin.CAHID || claims.UserRole != common.RoleAdminUser {
		t.Fatalf("claims missing admin identity: %+v", claims)
	}
	if claims.Issuer != AdminServiceAccountIssuer || len(claims.Audience) == 0 || claims.Audience[0] != AdminServiceAccountAudience {
		t.Fatalf("claims missing issuer/audience: %+v", claims.RegisteredClaims)
	}
}

func TestAdminServiceAccountJWTRejectsDisabledAccount(t *testing.T) {
	setupAdminServiceAccountTestDB(t)

	admin := &User{
		Username:    "disabled-admin",
		DisplayName: "Disabled Admin",
		Password:    "irrelevant",
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	if err := DB.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	account := &AdminServiceAccount{
		Name:              "disabled-asa",
		UserID:            admin.Id,
		Username:          admin.Username,
		UserCAHID:         admin.CAHID,
		UserRole:          admin.Role,
		CreatedByID:       admin.Id,
		CreatedByUsername: admin.Username,
		CreatedByCAHID:    admin.CAHID,
		ExpiresAt:         time.Now().Add(24 * time.Hour).Unix(),
	}
	credential, err := CreateAdminServiceAccount(account, admin)
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}
	account.Status = AdminServiceAccountStatusDisabled
	if err = account.UpdateMetadata(); err != nil {
		t.Fatalf("disable service account: %v", err)
	}

	if _, _, _, err = ValidateAdminServiceAccountJWT(credential, time.Now()); err == nil {
		t.Fatal("expected disabled service account JWT to be rejected")
	}
}

func TestAdminServiceAccountRotationInvalidatesPreviousJWT(t *testing.T) {
	setupAdminServiceAccountTestDB(t)

	admin := &User{
		Username:    "rotation-admin",
		DisplayName: "Rotation Admin",
		Password:    "irrelevant",
		Role:        common.RoleAdminUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}
	if err := DB.Create(admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	account := &AdminServiceAccount{
		Name:              "rotation-asa",
		UserID:            admin.Id,
		Username:          admin.Username,
		UserCAHID:         admin.CAHID,
		UserRole:          admin.Role,
		CreatedByID:       admin.Id,
		CreatedByUsername: admin.Username,
		CreatedByCAHID:    admin.CAHID,
		ExpiresAt:         time.Now().Add(24 * time.Hour).Unix(),
	}
	oldCredential, err := CreateAdminServiceAccount(account, admin)
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}

	newCredential, err := RotateAdminServiceAccountCredential(account, admin, time.Now().Add(48*time.Hour).Unix())
	if err != nil {
		t.Fatalf("rotate service account: %v", err)
	}
	if newCredential == "" || newCredential == oldCredential {
		t.Fatal("expected rotation to issue a new JWT")
	}
	if _, _, _, err = ValidateAdminServiceAccountJWT(oldCredential, time.Now()); err == nil {
		t.Fatal("expected previous JWT to be rejected after rotation")
	}
	if _, _, _, err = ValidateAdminServiceAccountJWT(newCredential, time.Now()); err != nil {
		t.Fatalf("expected rotated JWT to validate: %v", err)
	}
}
