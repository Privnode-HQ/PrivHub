package model

import "time"

// UserUsageWindow stores aggregated counters for one user within a specific
// minute/hour/day/month window so live checks only touch the active rows.
type UserUsageWindow struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          int       `json:"user_id" gorm:"index:idx_user_usage_window,unique"`
	WindowKind      string    `json:"window_kind" gorm:"size:16;index:idx_user_usage_window,unique"`
	WindowStart     int64     `json:"window_start" gorm:"bigint;index:idx_user_usage_window,unique"`
	WindowEnd       int64     `json:"window_end" gorm:"bigint;index"`
	RequestUsed     int64     `json:"request_used" gorm:"default:0"`
	RequestReserved int64     `json:"request_reserved" gorm:"default:0"`
	TokenUsed       int64     `json:"token_used" gorm:"default:0"`
	TokenReserved   int64     `json:"token_reserved" gorm:"default:0"`
	BudgetUsed      int64     `json:"budget_used" gorm:"default:0"`
	BudgetReserved  int64     `json:"budget_reserved" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserUsageReservation tracks in-flight capacity that has been admitted but not
// yet finalized, which keeps request/token/budget windows consistent.
type UserUsageReservation struct {
	ReservationID     string    `json:"reservation_id" gorm:"primaryKey;size:64"`
	UserID            int       `json:"user_id" gorm:"index:idx_user_usage_reservation_status"`
	GroupName         string    `json:"group_name" gorm:"size:64"`
	MinuteWindowStart int64     `json:"minute_window_start" gorm:"bigint"`
	HourWindowStart   int64     `json:"hour_window_start" gorm:"bigint"`
	DayWindowStart    int64     `json:"day_window_start" gorm:"bigint"`
	MonthWindowStart  int64     `json:"month_window_start" gorm:"bigint"`
	ReservedRequests  int64     `json:"reserved_requests" gorm:"default:0"`
	ReservedTokens    int64     `json:"reserved_tokens" gorm:"default:0"`
	ReservedBudget    int64     `json:"reserved_budget" gorm:"default:0"`
	Status            string    `json:"status" gorm:"size:16;index:idx_user_usage_reservation_status"`
	ExpiresAt         int64     `json:"expires_at" gorm:"bigint;index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
