package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	usageWindowMinute = "minute"
	usageWindowHour   = "hour"
	usageWindowDay    = "day"
	usageWindowWeek   = "week"
	usageWindowMonth  = "month"

	usageReservationReserved = "reserved"
	usageReservationSettled  = "settled"
	usageReservationReleased = "released"
	usageReservationExpired  = "expired"
)

var usageReservationTTL = 15 * time.Minute

type usageWindowBound struct {
	Kind  string
	Start time.Time
	End   time.Time
}

type usageReservationWindows struct {
	Minute *model.UserUsageWindow
	Hour   *model.UserUsageWindow
	Day    *model.UserUsageWindow
	Week   *model.UserUsageWindow
	Month  *model.UserUsageWindow
}

type UsageMetricSummary struct {
	Unit               string     `json:"unit"`
	Limit              *int64     `json:"limit"`
	Used               *int64     `json:"used"`
	Pending            *int64     `json:"pending"`
	Remaining          *int64     `json:"remaining"`
	ResetAt            *time.Time `json:"reset_at"`
	Status             string     `json:"status"`
	HideDetails        bool       `json:"hide_details"`
	ConsumptionPercent *float64   `json:"consumption_percent"`
}

type UsageMetricMap struct {
	RPM     UsageMetricSummary `json:"rpm"`
	RPD     UsageMetricSummary `json:"rpd"`
	TPM     UsageMetricSummary `json:"tpm"`
	TPD     UsageMetricSummary `json:"tpd"`
	Hourly  UsageMetricSummary `json:"hourly"`
	Daily   UsageMetricSummary `json:"daily"`
	Weekly  UsageMetricSummary `json:"weekly"`
	Monthly UsageMetricSummary `json:"monthly"`
}

type UserUsageLimitSnapshot struct {
	UserID                       int            `json:"user_id"`
	Group                        string         `json:"group"`
	BillingUnit                  string         `json:"billing_unit"`
	GeneratedAt                  time.Time      `json:"generated_at"`
	LegacyGroupRateLimitReplaced bool           `json:"legacy_group_rate_limit_replaced"`
	NoLimitsConfigured           bool           `json:"no_limits_configured"`
	Metrics                      UsageMetricMap `json:"metrics"`
}

type usageLimitExceededError struct {
	Metric  string
	Limit   int64
	ResetAt time.Time
}

func IsUsageLimitExceededError(err error) bool {
	var limitErr *usageLimitExceededError
	return errors.As(err, &limitErr)
}

func (e *usageLimitExceededError) Error() string {
	switch e.Metric {
	case "rpm":
		return fmt.Sprintf("您已达到 RPM 限制：每分钟最多 %d 次请求，将于 %s 重置", e.Limit, e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "rpd":
		return fmt.Sprintf("您已达到 RPD 限制：每日最多 %d 次请求，将于 %s 重置", e.Limit, e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "tpm":
		return fmt.Sprintf("您已达到 TPM 限制：每分钟最多消耗 %d tokens，将于 %s 重置", e.Limit, e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "tpd":
		return fmt.Sprintf("您已达到 TPD 限制：每日最多消耗 %d tokens，将于 %s 重置", e.Limit, e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "hourly":
		return fmt.Sprintf("您已达到每小时预算限制：每小时最多可使用 %s，将于 %s 重置", logger.FormatQuota(int(e.Limit)), e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "daily":
		return fmt.Sprintf("您已达到每日预算限制：每日最多可使用 %s，将于 %s 重置", logger.FormatQuota(int(e.Limit)), e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "weekly":
		return fmt.Sprintf("您已达到每周预算限制：本周最多可使用 %s，将于 %s 重置", logger.FormatQuota(int(e.Limit)), e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	case "monthly":
		return fmt.Sprintf("您已达到月度预算限制：本月最多可使用 %s，将于 %s 重置", logger.FormatQuota(int(e.Limit)), e.ResetAt.In(time.Local).Format("2006-01-02 15:04:05"))
	default:
		return "已达到使用限制"
	}
}

func usagePolicyHasAnyConfiguredLimit(policy setting.GroupUsageLimitPolicy) bool {
	return policy.RPM != nil || policy.RPD != nil || policy.TPM != nil || policy.TPD != nil || policy.Hourly != nil || policy.Daily != nil || policy.Weekly != nil || policy.Monthly != nil
}

func startOfWeek(now time.Time) time.Time {
	localNow := now.In(time.Local)
	dayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local)
	weekday := int(dayStart.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return dayStart.AddDate(0, 0, -(weekday - 1))
}

func getUsageWindowBounds(now time.Time) map[string]usageWindowBound {
	localNow := now.In(time.Local)
	minuteStart := localNow.Truncate(time.Minute)
	hourStart := localNow.Truncate(time.Hour)
	dayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.Local)
	weekStart := startOfWeek(localNow)
	monthStart := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, time.Local)

	return map[string]usageWindowBound{
		usageWindowMinute: {
			Kind:  usageWindowMinute,
			Start: minuteStart,
			End:   minuteStart.Add(time.Minute),
		},
		usageWindowHour: {
			Kind:  usageWindowHour,
			Start: hourStart,
			End:   hourStart.Add(time.Hour),
		},
		usageWindowDay: {
			Kind:  usageWindowDay,
			Start: dayStart,
			End:   dayStart.Add(24 * time.Hour),
		},
		usageWindowWeek: {
			Kind:  usageWindowWeek,
			Start: weekStart,
			End:   weekStart.AddDate(0, 0, 7),
		},
		usageWindowMonth: {
			Kind:  usageWindowMonth,
			Start: monthStart,
			End:   monthStart.AddDate(0, 1, 0),
		},
	}
}

func getReservationWindowBounds(reservation *model.UserUsageReservation) map[string]usageWindowBound {
	minuteStart := time.Unix(reservation.MinuteWindowStart, 0).In(time.Local)
	hourStart := minuteStart.Truncate(time.Hour)
	if reservation.HourWindowStart > 0 {
		hourStart = time.Unix(reservation.HourWindowStart, 0).In(time.Local)
	}
	dayStart := time.Unix(reservation.DayWindowStart, 0).In(time.Local)
	weekStart := startOfWeek(minuteStart)
	if reservation.WeekWindowStart > 0 {
		weekStart = time.Unix(reservation.WeekWindowStart, 0).In(time.Local)
	}
	monthStart := time.Unix(reservation.MonthWindowStart, 0).In(time.Local)

	return map[string]usageWindowBound{
		usageWindowMinute: {
			Kind:  usageWindowMinute,
			Start: minuteStart,
			End:   minuteStart.Add(time.Minute),
		},
		usageWindowHour: {
			Kind:  usageWindowHour,
			Start: hourStart,
			End:   hourStart.Add(time.Hour),
		},
		usageWindowDay: {
			Kind:  usageWindowDay,
			Start: dayStart,
			End:   dayStart.Add(24 * time.Hour),
		},
		usageWindowWeek: {
			Kind:  usageWindowWeek,
			Start: weekStart,
			End:   weekStart.AddDate(0, 0, 7),
		},
		usageWindowMonth: {
			Kind:  usageWindowMonth,
			Start: monthStart,
			End:   monthStart.AddDate(0, 1, 0),
		},
	}
}

func getOrCreateUsageWindowTx(tx *gorm.DB, userID int, bound usageWindowBound) (*model.UserUsageWindow, error) {
	window := &model.UserUsageWindow{
		UserID:      userID,
		WindowKind:  bound.Kind,
		WindowStart: bound.Start.Unix(),
	}
	createAttrs := model.UserUsageWindow{
		UserID:      userID,
		WindowKind:  bound.Kind,
		WindowStart: bound.Start.Unix(),
		WindowEnd:   bound.End.Unix(),
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, bound.Kind, bound.Start.Unix()).
		Attrs(createAttrs).
		FirstOrCreate(window).Error; err != nil {
		return nil, err
	}
	if window.WindowEnd != bound.End.Unix() {
		window.WindowEnd = bound.End.Unix()
		if err := tx.Save(window).Error; err != nil {
			return nil, err
		}
	}
	return window, nil
}

func getUsageWindowTx(tx *gorm.DB, userID int, bound usageWindowBound) (*model.UserUsageWindow, error) {
	var window model.UserUsageWindow
	if err := tx.Where("user_id = ? AND window_kind = ? AND window_start = ?", userID, bound.Kind, bound.Start.Unix()).First(&window).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.UserUsageWindow{
				UserID:      userID,
				WindowKind:  bound.Kind,
				WindowStart: bound.Start.Unix(),
				WindowEnd:   bound.End.Unix(),
			}, nil
		}
		return nil, err
	}
	return &window, nil
}

func loadReservationWindowsTx(tx *gorm.DB, reservation *model.UserUsageReservation) (*usageReservationWindows, error) {
	bounds := getReservationWindowBounds(reservation)

	minuteWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, bounds[usageWindowMinute])
	if err != nil {
		return nil, err
	}
	hourWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, bounds[usageWindowHour])
	if err != nil {
		return nil, err
	}
	dayWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, bounds[usageWindowDay])
	if err != nil {
		return nil, err
	}
	weekWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, bounds[usageWindowWeek])
	if err != nil {
		return nil, err
	}
	monthWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, bounds[usageWindowMonth])
	if err != nil {
		return nil, err
	}

	return &usageReservationWindows{
		Minute: minuteWindow,
		Hour:   hourWindow,
		Day:    dayWindow,
		Week:   weekWindow,
		Month:  monthWindow,
	}, nil
}

func saveUsageWindowTx(tx *gorm.DB, window *model.UserUsageWindow) error {
	if window.RequestReserved < 0 {
		window.RequestReserved = 0
	}
	if window.TokenReserved < 0 {
		window.TokenReserved = 0
	}
	if window.BudgetReserved < 0 {
		window.BudgetReserved = 0
	}
	if window.RequestUsed < 0 {
		window.RequestUsed = 0
	}
	if window.TokenUsed < 0 {
		window.TokenUsed = 0
	}
	if window.BudgetUsed < 0 {
		window.BudgetUsed = 0
	}
	return tx.Save(window).Error
}

func expireUserReservationsTx(tx *gorm.DB, userID int, now time.Time) error {
	var reservations []model.UserUsageReservation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND status = ? AND expires_at <= ?", userID, usageReservationReserved, now.Unix()).
		Find(&reservations).Error; err != nil {
		return err
	}
	if len(reservations) == 0 {
		return nil
	}

	for i := range reservations {
		reservation := &reservations[i]
		windows, err := loadReservationWindowsTx(tx, reservation)
		if err != nil {
			return err
		}

		windows.Minute.RequestReserved -= reservation.ReservedRequests
		windows.Minute.RequestUsed += reservation.ReservedRequests
		windows.Minute.TokenReserved -= reservation.ReservedTokens
		windows.Hour.BudgetReserved -= reservation.ReservedBudget
		windows.Day.RequestReserved -= reservation.ReservedRequests
		windows.Day.RequestUsed += reservation.ReservedRequests
		windows.Day.TokenReserved -= reservation.ReservedTokens
		windows.Day.BudgetReserved -= reservation.ReservedBudget
		windows.Week.BudgetReserved -= reservation.ReservedBudget
		windows.Month.BudgetReserved -= reservation.ReservedBudget

		if err := saveUsageWindowTx(tx, windows.Minute); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Hour); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Day); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Week); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Month); err != nil {
			return err
		}

		reservation.Status = usageReservationExpired
		if err := tx.Save(reservation).Error; err != nil {
			return err
		}
	}
	return nil
}

func estimateUsageTokens(promptTokens, maxTokens int) int64 {
	outputEstimate := common.PreConsumedQuota
	if maxTokens > outputEstimate {
		outputEstimate = maxTokens
	}
	return int64(promptTokens + outputEstimate)
}

func computeRemaining(limit *int64, used, pending int64) *int64 {
	if limit == nil {
		return nil
	}
	remaining := *limit - used - pending
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}

func cloneNullableInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func int64Pointer(value int64) *int64 {
	cloned := value
	return &cloned
}

func computeConsumptionPercent(limit *int64, used, pending int64) *float64 {
	if limit == nil {
		return nil
	}
	if *limit <= 0 {
		percent := 100.0
		return &percent
	}
	percent := float64(used+pending) * 100 / float64(*limit)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return &percent
}

func computeMetricStatus(limit *int64, used, pending int64) string {
	if limit == nil {
		return "unlimited"
	}
	if used+pending >= *limit {
		return "blocked"
	}
	return "available"
}

func buildUsageMetricSummary(unit string, limit *int64, used, pending int64, resetAt *time.Time, hideDetails bool) UsageMetricSummary {
	summary := UsageMetricSummary{
		Unit:               unit,
		ResetAt:            resetAt,
		Status:             computeMetricStatus(limit, used, pending),
		HideDetails:        hideDetails,
		ConsumptionPercent: computeConsumptionPercent(limit, used, pending),
	}
	if hideDetails {
		return summary
	}
	summary.Limit = cloneNullableInt64(limit)
	summary.Used = int64Pointer(used)
	summary.Pending = int64Pointer(pending)
	summary.Remaining = computeRemaining(limit, used, pending)
	return summary
}

func ReserveUsageRequest(c *gin.Context) error {
	userID := c.GetInt("id")
	if userID == 0 {
		return nil
	}
	group := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	if group == "" {
		return nil
	}

	policy, found := setting.GetUserGroupUsageLimit(group)
	if !found {
		return nil
	}

	now := time.Now()
	bounds := getUsageWindowBounds(now)
	reservationID := uuid.NewString()

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := expireUserReservationsTx(tx, userID, now); err != nil {
			return err
		}

		minuteWindow, err := getOrCreateUsageWindowTx(tx, userID, bounds[usageWindowMinute])
		if err != nil {
			return err
		}
		dayWindow, err := getOrCreateUsageWindowTx(tx, userID, bounds[usageWindowDay])
		if err != nil {
			return err
		}

		if policy.RPM != nil && minuteWindow.RequestUsed+minuteWindow.RequestReserved+1 > *policy.RPM {
			return &usageLimitExceededError{Metric: "rpm", Limit: *policy.RPM, ResetAt: bounds[usageWindowMinute].End}
		}
		if policy.RPD != nil && dayWindow.RequestUsed+dayWindow.RequestReserved+1 > *policy.RPD {
			return &usageLimitExceededError{Metric: "rpd", Limit: *policy.RPD, ResetAt: bounds[usageWindowDay].End}
		}

		minuteWindow.RequestReserved += 1
		dayWindow.RequestReserved += 1
		if err := saveUsageWindowTx(tx, minuteWindow); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, dayWindow); err != nil {
			return err
		}

		reservation := &model.UserUsageReservation{
			ReservationID:     reservationID,
			UserID:            userID,
			GroupName:         group,
			MinuteWindowStart: bounds[usageWindowMinute].Start.Unix(),
			HourWindowStart:   bounds[usageWindowHour].Start.Unix(),
			DayWindowStart:    bounds[usageWindowDay].Start.Unix(),
			WeekWindowStart:   bounds[usageWindowWeek].Start.Unix(),
			MonthWindowStart:  bounds[usageWindowMonth].Start.Unix(),
			ReservedRequests:  1,
			Status:            usageReservationReserved,
			ExpiresAt:         now.Add(usageReservationTTL).Unix(),
		}
		return tx.Create(reservation).Error
	})
	if err != nil {
		return err
	}

	common.SetContextKey(c, constant.ContextKeyUsageReservationID, reservationID)
	return nil
}

func ReserveUsageEstimate(c *gin.Context, relayInfo *relaycommon.RelayInfo, meta *types.TokenCountMeta, estimatedBudget int) *types.NewAPIError {
	if relayInfo == nil || relayInfo.UsageReservationID == "" {
		return nil
	}

	policy, found := setting.GetUserGroupUsageLimit(relayInfo.UserGroup)
	if !found {
		return nil
	}

	now := time.Now()
	currentBounds := getUsageWindowBounds(now)
	maxTokens := 0
	if meta != nil {
		maxTokens = meta.MaxTokens
	}
	estimatedTokens := estimateUsageTokens(relayInfo.PromptTokens, maxTokens)

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := expireUserReservationsTx(tx, relayInfo.UserId, now); err != nil {
			return err
		}

		var reservation model.UserUsageReservation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("reservation_id = ?", relayInfo.UsageReservationID).
			First(&reservation).Error; err != nil {
			return err
		}
		if reservation.Status != usageReservationReserved {
			return nil
		}

		reservationBounds := getReservationWindowBounds(&reservation)

		minuteWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, reservationBounds[usageWindowMinute])
		if err != nil {
			return err
		}
		hourWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, reservationBounds[usageWindowHour])
		if err != nil {
			return err
		}
		dayWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, reservationBounds[usageWindowDay])
		if err != nil {
			return err
		}
		weekWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, reservationBounds[usageWindowWeek])
		if err != nil {
			return err
		}
		monthWindow, err := getOrCreateUsageWindowTx(tx, reservation.UserID, reservationBounds[usageWindowMonth])
		if err != nil {
			return err
		}

		tokenDelta := estimatedTokens - reservation.ReservedTokens
		budgetDelta := int64(estimatedBudget) - reservation.ReservedBudget

		if tokenDelta > 0 {
			if policy.TPM != nil && reservationBounds[usageWindowMinute].Start.Equal(currentBounds[usageWindowMinute].Start) &&
				minuteWindow.TokenUsed+minuteWindow.TokenReserved+tokenDelta > *policy.TPM {
				return &usageLimitExceededError{Metric: "tpm", Limit: *policy.TPM, ResetAt: reservationBounds[usageWindowMinute].End}
			}

			dayCheckWindow := dayWindow
			dayResetAt := reservationBounds[usageWindowDay].End
			if !reservationBounds[usageWindowDay].Start.Equal(currentBounds[usageWindowDay].Start) {
				dayCheckWindow, err = getUsageWindowTx(tx, reservation.UserID, currentBounds[usageWindowDay])
				if err != nil {
					return err
				}
				dayResetAt = currentBounds[usageWindowDay].End
			}
			if policy.TPD != nil && dayCheckWindow.TokenUsed+dayCheckWindow.TokenReserved+tokenDelta > *policy.TPD {
				return &usageLimitExceededError{Metric: "tpd", Limit: *policy.TPD, ResetAt: dayResetAt}
			}
		}
		if budgetDelta > 0 {
			if policy.Hourly != nil {
				hourCheckWindow := hourWindow
				hourResetAt := reservationBounds[usageWindowHour].End
				if !reservationBounds[usageWindowHour].Start.Equal(currentBounds[usageWindowHour].Start) {
					hourCheckWindow, err = getUsageWindowTx(tx, reservation.UserID, currentBounds[usageWindowHour])
					if err != nil {
						return err
					}
					hourResetAt = currentBounds[usageWindowHour].End
				}
				if hourCheckWindow.BudgetUsed+hourCheckWindow.BudgetReserved+budgetDelta > *policy.Hourly {
					return &usageLimitExceededError{Metric: "hourly", Limit: *policy.Hourly, ResetAt: hourResetAt}
				}
			}
			if policy.Daily != nil {
				dayBudgetCheckWindow := dayWindow
				dayBudgetResetAt := reservationBounds[usageWindowDay].End
				if !reservationBounds[usageWindowDay].Start.Equal(currentBounds[usageWindowDay].Start) {
					dayBudgetCheckWindow, err = getUsageWindowTx(tx, reservation.UserID, currentBounds[usageWindowDay])
					if err != nil {
						return err
					}
					dayBudgetResetAt = currentBounds[usageWindowDay].End
				}
				if dayBudgetCheckWindow.BudgetUsed+dayBudgetCheckWindow.BudgetReserved+budgetDelta > *policy.Daily {
					return &usageLimitExceededError{Metric: "daily", Limit: *policy.Daily, ResetAt: dayBudgetResetAt}
				}
			}
			if policy.Weekly != nil {
				weekCheckWindow := weekWindow
				weekResetAt := reservationBounds[usageWindowWeek].End
				if !reservationBounds[usageWindowWeek].Start.Equal(currentBounds[usageWindowWeek].Start) {
					weekCheckWindow, err = getUsageWindowTx(tx, reservation.UserID, currentBounds[usageWindowWeek])
					if err != nil {
						return err
					}
					weekResetAt = currentBounds[usageWindowWeek].End
				}
				if weekCheckWindow.BudgetUsed+weekCheckWindow.BudgetReserved+budgetDelta > *policy.Weekly {
					return &usageLimitExceededError{Metric: "weekly", Limit: *policy.Weekly, ResetAt: weekResetAt}
				}
			}
			if policy.Monthly != nil {
				monthCheckWindow := monthWindow
				monthResetAt := reservationBounds[usageWindowMonth].End
				if !reservationBounds[usageWindowMonth].Start.Equal(currentBounds[usageWindowMonth].Start) {
					monthCheckWindow, err = getUsageWindowTx(tx, reservation.UserID, currentBounds[usageWindowMonth])
					if err != nil {
						return err
					}
					monthResetAt = currentBounds[usageWindowMonth].End
				}
				if monthCheckWindow.BudgetUsed+monthCheckWindow.BudgetReserved+budgetDelta > *policy.Monthly {
					return &usageLimitExceededError{Metric: "monthly", Limit: *policy.Monthly, ResetAt: monthResetAt}
				}
			}
		}

		minuteWindow.TokenReserved += tokenDelta
		hourWindow.BudgetReserved += budgetDelta
		dayWindow.TokenReserved += tokenDelta
		dayWindow.BudgetReserved += budgetDelta
		weekWindow.BudgetReserved += budgetDelta
		monthWindow.BudgetReserved += budgetDelta

		if err := saveUsageWindowTx(tx, minuteWindow); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, hourWindow); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, dayWindow); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, weekWindow); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, monthWindow); err != nil {
			return err
		}

		reservation.ReservedTokens = estimatedTokens
		reservation.ReservedBudget = int64(estimatedBudget)
		reservation.ExpiresAt = now.Add(usageReservationTTL).Unix()
		return tx.Save(&reservation).Error
	})
	if err == nil {
		return nil
	}

	if limitErr, ok := err.(*usageLimitExceededError); ok {
		return types.NewErrorWithStatusCode(limitErr, types.ErrorCodeGroupUsageLimitExceeded, 429, types.ErrOptionWithSkipRetry(), types.ErrOptionWithNoRecordErrorLog())
	}
	return types.NewError(err, types.ErrorCodeUpdateDataError, types.ErrOptionWithSkipRetry())
}

func ReleaseUsageReservation(relayInfo *relaycommon.RelayInfo) error {
	if relayInfo == nil || relayInfo.UsageReservationID == "" {
		return nil
	}
	return releaseUsageReservationByID(relayInfo.UserId, relayInfo.UsageReservationID)
}

func releaseUsageReservationByID(userID int, reservationID string) error {
	now := time.Now()
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := expireUserReservationsTx(tx, userID, now); err != nil {
			return err
		}

		var reservation model.UserUsageReservation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("reservation_id = ?", reservationID).
			First(&reservation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if reservation.Status != usageReservationReserved {
			return nil
		}

		windows, err := loadReservationWindowsTx(tx, &reservation)
		if err != nil {
			return err
		}

		// Failed requests still consume RPM/RPD, but token and budget reservations
		// are returned because no successful model usage was finalized.
		windows.Minute.RequestReserved -= reservation.ReservedRequests
		windows.Minute.RequestUsed += reservation.ReservedRequests
		windows.Minute.TokenReserved -= reservation.ReservedTokens
		windows.Hour.BudgetReserved -= reservation.ReservedBudget
		windows.Day.RequestReserved -= reservation.ReservedRequests
		windows.Day.RequestUsed += reservation.ReservedRequests
		windows.Day.TokenReserved -= reservation.ReservedTokens
		windows.Day.BudgetReserved -= reservation.ReservedBudget
		windows.Week.BudgetReserved -= reservation.ReservedBudget
		windows.Month.BudgetReserved -= reservation.ReservedBudget

		if err := saveUsageWindowTx(tx, windows.Minute); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Hour); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Day); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Week); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Month); err != nil {
			return err
		}

		reservation.Status = usageReservationReleased
		return tx.Save(&reservation).Error
	})
}

func SettleUsageReservation(relayInfo *relaycommon.RelayInfo, actualTokens int, actualBudget int) error {
	if relayInfo == nil || relayInfo.UsageReservationID == "" {
		return nil
	}

	now := time.Now()
	return model.DB.Transaction(func(tx *gorm.DB) error {
		if err := expireUserReservationsTx(tx, relayInfo.UserId, now); err != nil {
			return err
		}

		var reservation model.UserUsageReservation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("reservation_id = ?", relayInfo.UsageReservationID).
			First(&reservation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if reservation.Status != usageReservationReserved {
			return nil
		}

		windows, err := loadReservationWindowsTx(tx, &reservation)
		if err != nil {
			return err
		}

		windows.Minute.RequestReserved -= reservation.ReservedRequests
		windows.Minute.RequestUsed += 1
		windows.Minute.TokenReserved -= reservation.ReservedTokens
		windows.Minute.TokenUsed += int64(actualTokens)

		windows.Hour.BudgetReserved -= reservation.ReservedBudget
		windows.Hour.BudgetUsed += int64(actualBudget)

		windows.Day.RequestReserved -= reservation.ReservedRequests
		windows.Day.RequestUsed += 1
		windows.Day.TokenReserved -= reservation.ReservedTokens
		windows.Day.TokenUsed += int64(actualTokens)
		windows.Day.BudgetReserved -= reservation.ReservedBudget
		windows.Day.BudgetUsed += int64(actualBudget)

		windows.Week.BudgetReserved -= reservation.ReservedBudget
		windows.Week.BudgetUsed += int64(actualBudget)

		windows.Month.BudgetReserved -= reservation.ReservedBudget
		windows.Month.BudgetUsed += int64(actualBudget)

		if err := saveUsageWindowTx(tx, windows.Minute); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Hour); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Day); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Week); err != nil {
			return err
		}
		if err := saveUsageWindowTx(tx, windows.Month); err != nil {
			return err
		}

		reservation.Status = usageReservationSettled
		return tx.Save(&reservation).Error
	})
}

func GetUserUsageLimitSnapshot(userID int, group string) (*UserUsageLimitSnapshot, error) {
	now := time.Now()
	bounds := getUsageWindowBounds(now)
	policy, found := setting.GetUserGroupUsageLimit(group)

	var minuteWindow *model.UserUsageWindow
	var hourWindow *model.UserUsageWindow
	var dayWindow *model.UserUsageWindow
	var weekWindow *model.UserUsageWindow
	var monthWindow *model.UserUsageWindow

	if err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := expireUserReservationsTx(tx, userID, now); err != nil {
			return err
		}

		var err error
		minuteWindow, err = getUsageWindowTx(tx, userID, bounds[usageWindowMinute])
		if err != nil {
			return err
		}
		hourWindow, err = getUsageWindowTx(tx, userID, bounds[usageWindowHour])
		if err != nil {
			return err
		}
		dayWindow, err = getUsageWindowTx(tx, userID, bounds[usageWindowDay])
		if err != nil {
			return err
		}
		weekWindow, err = getUsageWindowTx(tx, userID, bounds[usageWindowWeek])
		if err != nil {
			return err
		}
		monthWindow, err = getUsageWindowTx(tx, userID, bounds[usageWindowMonth])
		return err
	}); err != nil {
		return nil, err
	}

	noLimitsConfigured := !found || !usagePolicyHasAnyConfiguredLimit(policy)

	metricResetTime := func(bound usageWindowBound) *time.Time {
		resetAt := bound.End
		return &resetAt
	}

	snapshot := &UserUsageLimitSnapshot{
		UserID:                       userID,
		Group:                        group,
		BillingUnit:                  operation_setting.GetQuotaDisplayType(),
		GeneratedAt:                  now,
		LegacyGroupRateLimitReplaced: true,
		NoLimitsConfigured:           noLimitsConfigured,
		Metrics: UsageMetricMap{
			RPM:     buildUsageMetricSummary("requests", policy.RPM, minuteWindow.RequestUsed, minuteWindow.RequestReserved, metricResetTime(bounds[usageWindowMinute]), policy.ShouldHideMetricDetails("rpm")),
			RPD:     buildUsageMetricSummary("requests", policy.RPD, dayWindow.RequestUsed, dayWindow.RequestReserved, metricResetTime(bounds[usageWindowDay]), policy.ShouldHideMetricDetails("rpd")),
			TPM:     buildUsageMetricSummary("tokens", policy.TPM, minuteWindow.TokenUsed, minuteWindow.TokenReserved, metricResetTime(bounds[usageWindowMinute]), policy.ShouldHideMetricDetails("tpm")),
			TPD:     buildUsageMetricSummary("tokens", policy.TPD, dayWindow.TokenUsed, dayWindow.TokenReserved, metricResetTime(bounds[usageWindowDay]), policy.ShouldHideMetricDetails("tpd")),
			Hourly:  buildUsageMetricSummary("quota", policy.Hourly, hourWindow.BudgetUsed, hourWindow.BudgetReserved, metricResetTime(bounds[usageWindowHour]), policy.ShouldHideMetricDetails("hourly")),
			Daily:   buildUsageMetricSummary("quota", policy.Daily, dayWindow.BudgetUsed, dayWindow.BudgetReserved, metricResetTime(bounds[usageWindowDay]), policy.ShouldHideMetricDetails("daily")),
			Weekly:  buildUsageMetricSummary("quota", policy.Weekly, weekWindow.BudgetUsed, weekWindow.BudgetReserved, metricResetTime(bounds[usageWindowWeek]), policy.ShouldHideMetricDetails("weekly")),
			Monthly: buildUsageMetricSummary("quota", policy.Monthly, monthWindow.BudgetUsed, monthWindow.BudgetReserved, metricResetTime(bounds[usageWindowMonth]), policy.ShouldHideMetricDetails("monthly")),
		},
	}
	return snapshot, nil
}
