package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func GetSelfUsageLimits(c *gin.Context) {
	userID := c.GetInt("id")
	userCache, err := model.GetUserCache(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	snapshot, err := service.GetUserUsageLimitSnapshot(userID, userCache.Group)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    snapshot,
	})
}

type ResetUsageLimitsRequest struct {
	Scope      string   `json:"scope"`
	GroupNames []string `json:"group_names"`
	UserIDs    []int    `json:"user_ids"`
}

func ResetUsageLimits(c *gin.Context) {
	var req ResetUsageLimitsRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	req.Scope = strings.TrimSpace(strings.ToLower(req.Scope))
	result, err := service.ResetUserUsageLimits(req.Scope, req.GroupNames, req.UserIDs)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, result)
}
