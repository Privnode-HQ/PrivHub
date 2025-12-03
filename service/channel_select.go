package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

const maxTemporaryDisableSelectionAttempts = 8

var errAllCandidateChannelsTemporarilyDisabled = errors.New("all candidate channels are temporarily disabled")

func CacheGetRandomSatisfiedChannel(c *gin.Context, group string, modelName string, retry int) (*model.Channel, string, error) {
	var channel *model.Channel
	var err error
	selectGroup := group
	userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	if group == "auto" {
		if len(setting.GetAutoGroups()) == 0 {
			return nil, selectGroup, errors.New("auto groups is not enabled")
		}
		var lastErr error
		for _, autoGroup := range GetUserAutoGroup(userGroup) {
			logger.LogDebug(c, "Auto selecting group:", autoGroup)
			channel, err = selectChannelWithTemporarySkip(c, autoGroup, modelName, retry)
			if err != nil {
				if errors.Is(err, errAllCandidateChannelsTemporarilyDisabled) {
					lastErr = err
					continue
				}
				return nil, autoGroup, err
			}
			if channel == nil {
				continue
			}
			c.Set("auto_group", autoGroup)
			selectGroup = autoGroup
			logger.LogDebug(c, "Auto selected group:", autoGroup)
			break
		}
		if channel == nil && lastErr != nil {
			return nil, selectGroup, lastErr
		}
	} else {
		channel, err = selectChannelWithTemporarySkip(c, group, modelName, retry)
		if err != nil {
			return nil, group, err
		}
	}
	return channel, selectGroup, nil
}

func selectChannelWithTemporarySkip(c *gin.Context, group string, modelName string, retry int) (*model.Channel, error) {
	attempts := 0
	var excluded map[int]struct{}
	for {
		channel, err := model.GetRandomSatisfiedChannel(group, modelName, retry, excluded)
		if err != nil {
			if errors.Is(err, model.ErrAllCandidateChannelsFiltered) {
				return nil, fmt.Errorf("%w: group=%s, model=%s", errAllCandidateChannelsTemporarilyDisabled, group, modelName)
			}
			return channel, err
		}
		if channel == nil {
			return nil, nil
		}
		expireAt, reason, ok := GetTemporaryDisabledChannelInfo(channel.Id)
		if !ok {
			return channel, nil
		}
		if excluded == nil {
			excluded = make(map[int]struct{})
		}
		excluded[channel.Id] = struct{}{}
		logger.LogWarn(c, fmt.Sprintf("channel #%d is temporarily disabled until %s: %s", channel.Id, expireAt.Format(time.RFC3339), reason))
		attempts++
		if attempts >= maxTemporaryDisableSelectionAttempts {
			return nil, fmt.Errorf("%w: group=%s, model=%s", errAllCandidateChannelsTemporarilyDisabled, group, modelName)
		}
	}
}
