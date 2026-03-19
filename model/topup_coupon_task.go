package model

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const topUpCouponCleanupInterval = 5 * time.Minute

var topUpCouponCleanupOnce sync.Once

func StartTopUpCouponCleanupLoop() {
	topUpCouponCleanupOnce.Do(func() {
		ticker := time.NewTicker(topUpCouponCleanupInterval)
		go func() {
			if err := CleanupTopUpCouponStates(); err != nil {
				common.SysLog("failed to cleanup topup coupon states: " + err.Error())
			}
			for range ticker.C {
				if err := CleanupTopUpCouponStates(); err != nil {
					common.SysLog("failed to cleanup topup coupon states: " + err.Error())
				}
			}
		}()
	})
}
