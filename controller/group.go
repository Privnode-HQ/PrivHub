package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range ratio_setting.GetGroupRatioCopy() {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}

func GetUserGroups(c *gin.Context) {
	usableGroups := make(map[string]map[string]interface{})
	userGroup := ""
	userId := c.GetInt("id")
	userGroup, _ = model.GetUserGroup(userId, false)
	userSetting := dto.UserSetting{}
	if userId != 0 {
		if settingMap, err := model.GetUserSetting(userId, false); err == nil {
			userSetting = settingMap
		}
	}
	userUsableGroups := service.GetUserUsableGroups(userGroup)
	for groupName, _ := range ratio_setting.GetGroupRatioCopy() {
		// UserUsableGroups contains the groups that the user can use
		if desc, ok := userUsableGroups[groupName]; ok {
			captureRate := service.GetGroupCaptureRateForSelection(userGroup, groupName)
			usableGroups[groupName] = map[string]interface{}{
				"ratio":                          service.GetUserGroupRatio(userGroup, groupName),
				"desc":                           desc,
				"capture_rate":                   captureRate,
				"requires_training_data_consent": captureRate > 0,
				"training_data_allowed":          userSetting.AllowTrainingDataGroups,
			}
		}
	}
	if _, ok := userUsableGroups["auto"]; ok {
		captureRate := service.GetGroupCaptureRateForSelection(userGroup, "auto")
		usableGroups["auto"] = map[string]interface{}{
			"ratio":                          "自动",
			"desc":                           setting.GetUsableGroupDescription("auto"),
			"capture_rate":                   captureRate,
			"requires_training_data_consent": captureRate > 0,
			"training_data_allowed":          userSetting.AllowTrainingDataGroups,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    usableGroups,
	})
}
