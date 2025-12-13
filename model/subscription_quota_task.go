package model

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const subscriptionQuotaResetInterval = 5 * time.Minute

var subscriptionQuotaResetOnce sync.Once

func StartSubscriptionQuotaResetLoop() {
	subscriptionQuotaResetOnce.Do(func() {
		ticker := time.NewTicker(subscriptionQuotaResetInterval)
		go func() {
			if err := ResetSubscriptionQuotaForAllUsers(time.Now().Unix()); err != nil {
				common.SysLog("failed to reset subscription quota: " + err.Error())
			}
			for range ticker.C {
				err := ResetSubscriptionQuotaForAllUsers(time.Now().Unix())
				if err != nil {
					common.SysLog("failed to reset subscription quota: " + err.Error())
				}
			}
		}()
	})
}
