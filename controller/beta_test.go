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

func setupAdminPaidQuotaControllerTestDB(t *testing.T) {
	t.Helper()

	originalDB := model.DB
	originalQuotaPerUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		model.DB = originalDB
		common.QuotaPerUnit = originalQuotaPerUnit
	})

	common.QuotaPerUnit = common.DefaultQuotaPerUnit
	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(&model.User{}, &model.TopUp{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	model.DB = db
}

func TestGetAdminUserRemainActualPaidAmountByCAHIDWithRate(t *testing.T) {
	setupAdminPaidQuotaControllerTestDB(t)
	gin.SetMode(gin.TestMode)

	user := &model.User{
		Username:  "paid-controller",
		Password:  "irrelevant",
		Quota:     11_000_000,
		UsedQuota: 4_000_000,
	}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	topUps := []model.TopUp{
		{
			UserId:        user.Id,
			TradeNo:       "discounted-epay-controller",
			PaymentMethod: "alipay",
			Status:        common.TopUpStatusSuccess,
			Amount:        10,
			OriginalMoney: 10,
			PayMoney:      7.35,
			ProcessingFee: 0.35,
		},
		{
			UserId:        user.Id,
			TradeNo:       "creem-full-controller",
			PaymentMethod: "creem",
			Status:        common.TopUpStatusSuccess,
			Amount:        2_000_000,
			OriginalMoney: 2,
			PayMoney:      2,
		},
	}
	if err := model.DB.Create(&topUps).Error; err != nil {
		t.Fatalf("create topups: %v", err)
	}

	router := gin.New()
	router.GET("/admin/users/:cah_id/remain_actual_paid_amount", GetAdminUserRemainActualPaidAmountByCAHID)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin/users/"+user.CAHID+"/remain_actual_paid_amount?rate=0.9", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			UserTotalQuota              int64  `json:"user_total_quota"`
			UserRemainQuota             int64  `json:"user_remain_quota"`
			UserRemainPaidQuota         int64  `json:"user_remain_paid_quota"`
			UserRemainNonPaidQuota      int64  `json:"user_remain_non_paid_quota"`
			UserRemainPaidQuotaAdjusted int64  `json:"user_remain_paid_quota_adjusted"`
			Rate                        string `json:"rate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response: %s", recorder.Body.String())
	}
	if response.Data.UserTotalQuota != 15_000_000 ||
		response.Data.UserRemainQuota != 11_000_000 ||
		response.Data.UserRemainPaidQuota != 5_000_000 ||
		response.Data.UserRemainNonPaidQuota != 6_000_000 ||
		response.Data.UserRemainPaidQuotaAdjusted != 4_500_000 ||
		response.Data.Rate != "0.9" {
		t.Fatalf("unexpected response data: %+v", response.Data)
	}
}
