package service

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func GetUserUsableGroups(userGroup string) map[string]string {
	groupsCopy := setting.GetUserUsableGroupsCopy()
	if userGroup != "" {
		specialSettings, b := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Get(userGroup)
		if b {
			// 处理特殊可用分组
			for specialGroup, desc := range specialSettings {
				if strings.HasPrefix(specialGroup, "-:") {
					// 移除分组
					groupToRemove := strings.TrimPrefix(specialGroup, "-:")
					delete(groupsCopy, groupToRemove)
				} else if strings.HasPrefix(specialGroup, "+:") {
					// 添加分组
					groupToAdd := strings.TrimPrefix(specialGroup, "+:")
					groupsCopy[groupToAdd] = desc
				} else {
					// 直接添加分组
					groupsCopy[specialGroup] = desc
				}
			}
		}
		// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
		if _, ok := groupsCopy[userGroup]; !ok {
			groupsCopy[userGroup] = "用户分组"
		}
	}
	return groupsCopy
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[groupName]
	return ok
}

func GetGroupCaptureRateForSelection(userGroup, groupName string) float64 {
	rate := ratio_setting.GetGroupCaptureRate(groupName)
	if groupName != "auto" {
		return rate
	}
	for _, autoGroup := range GetUserAutoGroup(userGroup) {
		if autoRate := ratio_setting.GetGroupCaptureRate(autoGroup); autoRate > rate {
			rate = autoRate
		}
	}
	return rate
}

func FirstGroupRequiringTrainingDataConsent(userGroup string, groups []string) (string, float64, bool) {
	if len(groups) == 0 {
		groups = []string{userGroup}
	}
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		rate := GetGroupCaptureRateForSelection(userGroup, group)
		if rate > 0 {
			return group, rate, true
		}
	}
	return "", 0, false
}

func ValidateTrainingDataGroupConsent(userGroup string, groups []string, userSetting dto.UserSetting) error {
	if userSetting.AllowTrainingDataGroups {
		return nil
	}
	group, rate, ok := FirstGroupRequiringTrainingDataConsent(userGroup, groups)
	if !ok {
		return nil
	}
	return fmt.Errorf("分组 %s 会按 %.2f%% 的采集率记录提示与补全，请先在用户设置 > 安全设置中允许使用数据采集分组", group, rate*100)
}

// GetUserAutoGroup 根据用户分组获取自动分组设置
func GetUserAutoGroup(userGroup string) []string {
	groups := GetUserUsableGroups(userGroup)
	autoGroups := make([]string, 0)
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; ok {
			autoGroups = append(autoGroups, group)
		}
	}
	return autoGroups
}

// GetUserGroupRatio 获取用户使用某个分组的倍率
// userGroup 用户分组
// group 需要获取倍率的分组
func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
