package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminTopUpControllerTestDB(t *testing.T) {
	t.Helper()

	originalDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
	})

	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.User{}, &model.TopUp{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	model.DB = db
}

func TestGetAdminTopUpByTradeNoReturnsTopUpAndSafeUser(t *testing.T) {
	setupAdminTopUpControllerTestDB(t)
	gin.SetMode(gin.TestMode)

	accessToken := "sensitive-access-token"
	user := &model.User{
		Username:       "topup-user",
		Password:       "irrelevant",
		DisplayName:    "Top Up User",
		Email:          "topup@example.com",
		AccessToken:    &accessToken,
		Quota:          1000,
		UsedQuota:      200,
		Group:          "default",
		StripeCustomer: "cus_test",
	}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	topUp := &model.TopUp{
		UserId:        user.Id,
		Amount:        100,
		Money:         7.35,
		TradeNo:       "ref_admin_lookup",
		PaymentMethod: "stripe",
		Status:        common.TopUpStatusSuccess,
		PayMoney:      7.35,
	}
	if err := model.DB.Create(topUp).Error; err != nil {
		t.Fatalf("create topup: %v", err)
	}

	router := gin.New()
	router.GET("/admin/topups/:trade_no", GetAdminTopUpByTradeNo)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin/topups/ref_admin_lookup", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			TopUp struct {
				TradeNo       string  `json:"trade_no"`
				PaymentMethod string  `json:"payment_method"`
				PayMoney      float64 `json:"pay_money"`
			} `json:"topup"`
			User map[string]interface{} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response: %s", recorder.Body.String())
	}
	if response.Data.TopUp.TradeNo != topUp.TradeNo ||
		response.Data.TopUp.PaymentMethod != topUp.PaymentMethod ||
		response.Data.TopUp.PayMoney != topUp.PayMoney {
		t.Fatalf("unexpected topup response: %+v", response.Data.TopUp)
	}
	if response.Data.User["cah_id"] != user.CAHID ||
		response.Data.User["username"] != user.Username ||
		response.Data.User["email"] != user.Email {
		t.Fatalf("unexpected user response: %+v", response.Data.User)
	}
	if _, ok := response.Data.User["password"]; ok {
		t.Fatalf("password leaked in response: %+v", response.Data.User)
	}
	if _, ok := response.Data.User["access_token"]; ok {
		t.Fatalf("access_token leaked in response: %+v", response.Data.User)
	}
}

func TestGetAdminTopUpByTradeNoReturnsDeletedUserInfo(t *testing.T) {
	setupAdminTopUpControllerTestDB(t)
	gin.SetMode(gin.TestMode)

	user := &model.User{
		Username: "deleted-topup-user",
		Password: "irrelevant",
	}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	topUp := &model.TopUp{
		UserId:  user.Id,
		TradeNo: "ref_deleted_user_lookup",
		Status:  common.TopUpStatusPending,
	}
	if err := model.DB.Create(topUp).Error; err != nil {
		t.Fatalf("create topup: %v", err)
	}
	if err := model.DB.Delete(user).Error; err != nil {
		t.Fatalf("delete user: %v", err)
	}

	router := gin.New()
	router.GET("/admin/topups/:trade_no", GetAdminTopUpByTradeNo)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin/topups/ref_deleted_user_lookup", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			User map[string]interface{} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response: %s", recorder.Body.String())
	}
	if response.Data.User["deleted"] != true {
		t.Fatalf("expected deleted user marker, got %+v", response.Data.User)
	}
	if _, ok := response.Data.User["deleted_at"]; !ok {
		t.Fatalf("expected deleted_at in response: %+v", response.Data.User)
	}
}
