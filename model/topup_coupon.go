package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"gorm.io/gorm"
)

const topUpCouponReservationTTLSeconds = int64((24 * time.Hour) / time.Second)

type TopUpCoupon struct {
	Id               int     `json:"id"`
	Name             string  `json:"name" gorm:"type:varchar(64);index"`
	BoundUserId      int     `json:"bound_user_id" gorm:"index"`
	BoundUsername    string  `json:"bound_username" gorm:"-"`
	DeductionAmount  float64 `json:"deduction_amount"`
	CurrencyCode     string  `json:"currency_code" gorm:"type:varchar(16);index"`
	Status           string  `json:"status" gorm:"type:varchar(16);index"`
	EffectiveStatus  string  `json:"effective_status" gorm:"-"`
	ValidFrom        int64   `json:"valid_from" gorm:"index"`
	ExpiresAt        int64   `json:"expires_at" gorm:"index"`
	IssuedByAdminId  int     `json:"issued_by_admin_id" gorm:"index"`
	IssuedAt         int64   `json:"issued_at"`
	ReservedTopUpId  int     `json:"reserved_top_up_id" gorm:"index"`
	ReservedAt       int64   `json:"reserved_at"`
	UsedTopUpId      int     `json:"used_top_up_id" gorm:"index"`
	UsedAt           int64   `json:"used_at"`
	RevokedAt        int64   `json:"revoked_at"`
	RevokedByAdminId int     `json:"revoked_by_admin_id" gorm:"index"`
	RevokeReason     string  `json:"revoke_reason" gorm:"type:varchar(255)"`
	CreatedTime      int64   `json:"created_time"`
	UpdatedTime      int64   `json:"updated_time"`
}

type TopUpCouponFilter struct {
	Keyword string
	Status  string
	UserId  int
}

type UserTopUpCouponSummary struct {
	HasAvailableCoupon      bool    `json:"has_available_coupon"`
	AvailableCount          int     `json:"available_count"`
	StrongestDiscountAmount float64 `json:"strongest_discount_amount"`
	StrongestCurrencyCode   string  `json:"strongest_currency_code,omitempty"`
	HasMixedCurrency        bool    `json:"has_mixed_currency"`
	BannerMessage           string  `json:"banner_message"`
}

func NormalizeTopUpCouponCurrencyCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func DefaultTopUpCouponCurrencyCode() string {
	switch operation_setting.GetQuotaDisplayType() {
	case operation_setting.QuotaDisplayTypeCNY:
		return "CNY"
	case operation_setting.QuotaDisplayTypeCustom:
		return "CUSTOM"
	default:
		return "USD"
	}
}

func isValidTopUpCouponCurrencyCode(code string) bool {
	if len(code) < 2 || len(code) > 16 {
		return false
	}
	for _, r := range code {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func (coupon *TopUpCoupon) GetCurrencyCode() string {
	if coupon == nil {
		return ""
	}
	return NormalizeTopUpCouponCurrencyCode(coupon.CurrencyCode)
}

func (coupon *TopUpCoupon) GetDisplayCurrencyCode(fallback string) string {
	couponCurrency := coupon.GetCurrencyCode()
	if couponCurrency != "" {
		return couponCurrency
	}
	return NormalizeTopUpCouponCurrencyCode(fallback)
}

func (coupon *TopUpCoupon) IsCurrencyCompatible(currencyCode string) bool {
	couponCurrency := coupon.GetCurrencyCode()
	paymentCurrency := NormalizeTopUpCouponCurrencyCode(currencyCode)
	if couponCurrency == "" || paymentCurrency == "" {
		return true
	}
	return couponCurrency == paymentCurrency
}

func (coupon *TopUpCoupon) Insert() error {
	now := common.GetTimestamp()
	if coupon.ValidFrom == 0 {
		coupon.ValidFrom = now
	}
	if coupon.IssuedAt == 0 {
		coupon.IssuedAt = now
	}
	if coupon.CreatedTime == 0 {
		coupon.CreatedTime = now
	}
	coupon.UpdatedTime = now
	if coupon.Status == "" {
		coupon.Status = common.TopUpCouponStatusAvailable
	}
	coupon.CurrencyCode = NormalizeTopUpCouponCurrencyCode(coupon.CurrencyCode)
	return DB.Create(coupon).Error
}

func (coupon *TopUpCoupon) Update() error {
	coupon.UpdatedTime = common.GetTimestamp()
	coupon.CurrencyCode = NormalizeTopUpCouponCurrencyCode(coupon.CurrencyCode)
	return DB.Save(coupon).Error
}

func (coupon *TopUpCoupon) GetEffectiveStatus() string {
	if coupon == nil {
		return ""
	}
	if coupon.EffectiveStatus != "" {
		return coupon.EffectiveStatus
	}
	return coupon.Status
}

func GetTopUpCouponById(id int) (*TopUpCoupon, error) {
	if id == 0 {
		return nil, errors.New("缺少优惠券 ID")
	}

	coupon := &TopUpCoupon{}
	if err := DB.Where("id = ?", id).First(coupon).Error; err != nil {
		return nil, err
	}
	if err := fillCouponViewData([]*TopUpCoupon{coupon}); err != nil {
		return nil, err
	}
	return coupon, nil
}

func GetAllTopUpCoupons(pageInfo *common.PageInfo, filter TopUpCouponFilter) ([]*TopUpCoupon, int64, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := applyTopUpCouponFilter(tx.Model(&TopUpCoupon{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	var coupons []*TopUpCoupon
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&coupons).Error; err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}
	if err := fillCouponViewData(coupons); err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

func SearchTopUpCoupons(keyword string, pageInfo *common.PageInfo) ([]*TopUpCoupon, int64, error) {
	return GetAllTopUpCoupons(pageInfo, TopUpCouponFilter{Keyword: keyword})
}

func GetUserAvailableTopUpCoupons(userId int) ([]*TopUpCoupon, error) {
	if userId == 0 {
		return nil, nil
	}

	now := common.GetTimestamp()
	var coupons []*TopUpCoupon
	if err := DB.Where(
		"bound_user_id = ? AND status IN ? AND valid_from <= ? AND (expires_at = 0 OR expires_at >= ?)",
		userId,
		[]string{common.TopUpCouponStatusAvailable, common.TopUpCouponStatusReserved},
		now,
		now,
	).Order("deduction_amount desc, id desc").Find(&coupons).Error; err != nil {
		return nil, err
	}
	if err := fillCouponEffectiveStatuses(coupons); err != nil {
		return nil, err
	}

	availableCoupons := make([]*TopUpCoupon, 0, len(coupons))
	for _, coupon := range coupons {
		if coupon.GetEffectiveStatus() == common.TopUpCouponStatusAvailable {
			availableCoupons = append(availableCoupons, coupon)
		}
	}
	return availableCoupons, nil
}

func GetUserTopUpCouponSummary(userId int) (*UserTopUpCouponSummary, error) {
	coupons, err := GetUserAvailableTopUpCoupons(userId)
	if err != nil {
		return nil, err
	}

	summary := &UserTopUpCouponSummary{
		HasAvailableCoupon: len(coupons) > 0,
		AvailableCount:     len(coupons),
		BannerMessage:      "您有可用于充值的优惠券",
	}
	currencySet := make(map[string]struct{})
	for _, coupon := range coupons {
		if coupon.DeductionAmount > summary.StrongestDiscountAmount {
			summary.StrongestDiscountAmount = coupon.DeductionAmount
			summary.StrongestCurrencyCode = coupon.GetDisplayCurrencyCode(DefaultTopUpCouponCurrencyCode())
		}
		currencyCode := coupon.GetDisplayCurrencyCode(DefaultTopUpCouponCurrencyCode())
		if currencyCode != "" {
			currencySet[currencyCode] = struct{}{}
		}
	}
	if len(currencySet) > 1 {
		summary.HasMixedCurrency = true
	}
	return summary, nil
}

func ReserveTopUpCouponTx(tx *gorm.DB, couponId int, userId int, topUp *TopUp) (*TopUpCoupon, error) {
	if couponId == 0 || topUp == nil {
		return nil, nil
	}

	coupon := &TopUpCoupon{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", couponId).First(coupon).Error; err != nil {
		return nil, errors.New("优惠券不存在")
	}

	now := common.GetTimestamp()
	if err := ensureCouponReservableTx(tx, coupon, userId, now); err != nil {
		return nil, err
	}

	coupon.Status = common.TopUpCouponStatusReserved
	coupon.ReservedTopUpId = topUp.Id
	coupon.ReservedAt = now
	coupon.UpdatedTime = now

	topUp.CouponId = coupon.Id
	topUp.CouponName = coupon.Name

	if err := tx.Save(coupon).Error; err != nil {
		return nil, err
	}
	if err := tx.Save(topUp).Error; err != nil {
		return nil, err
	}
	return coupon, nil
}

func MarkTopUpCouponUsedTx(tx *gorm.DB, topUp *TopUp) error {
	if topUp == nil || topUp.CouponId == 0 {
		return nil
	}

	coupon := &TopUpCoupon{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", topUp.CouponId).First(coupon).Error; err != nil {
		return errors.New("优惠券不存在")
	}
	if coupon.Status != common.TopUpCouponStatusReserved || coupon.ReservedTopUpId != topUp.Id {
		return errors.New("优惠券状态错误")
	}

	now := common.GetTimestamp()
	coupon.Status = common.TopUpCouponStatusUsed
	coupon.UsedTopUpId = topUp.Id
	coupon.UsedAt = now
	coupon.ReservedTopUpId = 0
	coupon.ReservedAt = 0
	coupon.UpdatedTime = now
	return tx.Save(coupon).Error
}

func ReleaseTopUpCouponReservationTx(tx *gorm.DB, topUp *TopUp) error {
	if topUp == nil || topUp.CouponId == 0 {
		return nil
	}

	coupon := &TopUpCoupon{}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", topUp.CouponId).First(coupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if coupon.ReservedTopUpId != topUp.Id || coupon.Status != common.TopUpCouponStatusReserved {
		return nil
	}

	now := common.GetTimestamp()
	coupon.ReservedTopUpId = 0
	coupon.ReservedAt = 0
	coupon.UpdatedTime = now
	if coupon.ExpiresAt != 0 && coupon.ExpiresAt < now {
		coupon.Status = common.TopUpCouponStatusExpired
	} else {
		coupon.Status = common.TopUpCouponStatusAvailable
	}
	return tx.Save(coupon).Error
}

func RevokeTopUpCoupon(id int, adminId int, reason string) (*TopUpCoupon, error) {
	if id == 0 {
		return nil, errors.New("缺少优惠券 ID")
	}

	var result *TopUpCoupon
	err := DB.Transaction(func(tx *gorm.DB) error {
		coupon := &TopUpCoupon{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(coupon).Error; err != nil {
			return err
		}
		now := common.GetTimestamp()
		if err := normalizeTopUpCouponStateTx(tx, coupon, now); err != nil {
			return err
		}
		if coupon.Status == common.TopUpCouponStatusUsed {
			return errors.New("已使用的优惠券不能撤销")
		}
		if coupon.Status == common.TopUpCouponStatusReserved {
			return errors.New("支付中的优惠券不能撤销")
		}

		coupon.Status = common.TopUpCouponStatusRevoked
		coupon.RevokedAt = now
		coupon.RevokedByAdminId = adminId
		coupon.RevokeReason = strings.TrimSpace(reason)
		coupon.ReservedTopUpId = 0
		coupon.ReservedAt = 0
		coupon.UpdatedTime = now
		if err := tx.Save(coupon).Error; err != nil {
			return err
		}
		result = coupon
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := fillCouponViewData([]*TopUpCoupon{result}); err != nil {
		return nil, err
	}
	return result, nil
}

func RefreshTopUpCouponState(id int) (*TopUpCoupon, error) {
	if id == 0 {
		return nil, errors.New("缺少优惠券 ID")
	}

	var result *TopUpCoupon
	err := DB.Transaction(func(tx *gorm.DB) error {
		coupon := &TopUpCoupon{}
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(coupon).Error; err != nil {
			return err
		}
		if err := normalizeTopUpCouponStateTx(tx, coupon, common.GetTimestamp()); err != nil {
			return err
		}
		result = coupon
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := fillCouponViewData([]*TopUpCoupon{result}); err != nil {
		return nil, err
	}
	return result, nil
}

func CleanupTopUpCouponStates() error {
	now := common.GetTimestamp()
	return DB.Transaction(func(tx *gorm.DB) error {
		var availableCoupons []*TopUpCoupon
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("status = ? AND expires_at != 0 AND expires_at < ?", common.TopUpCouponStatusAvailable, now).
			Find(&availableCoupons).Error; err != nil {
			return err
		}
		for _, coupon := range availableCoupons {
			if err := normalizeTopUpCouponStateTx(tx, coupon, now); err != nil {
				return err
			}
		}

		var reservedCoupons []*TopUpCoupon
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("status = ?", common.TopUpCouponStatusReserved).
			Find(&reservedCoupons).Error; err != nil {
			return err
		}
		for _, coupon := range reservedCoupons {
			if err := normalizeTopUpCouponStateTx(tx, coupon, now); err != nil {
				return err
			}
		}

		return nil
	})
}

func applyTopUpCouponFilter(query *gorm.DB, filter TopUpCouponFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UserId > 0 {
		query = query.Where("bound_user_id = ?", filter.UserId)
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword == "" {
		return query
	}

	like := "%" + keyword + "%"
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR bound_user_id = ? OR name LIKE ?", id, id, like)
	} else {
		query = query.Where("name LIKE ?", like)
	}
	return query
}

func ensureCouponReservableTx(tx *gorm.DB, coupon *TopUpCoupon, userId int, now int64) error {
	if coupon.BoundUserId != userId {
		return errors.New("优惠券不属于当前用户")
	}
	if coupon.ValidFrom > now {
		return errors.New("优惠券暂未生效")
	}
	if err := normalizeTopUpCouponStateTx(tx, coupon, now); err != nil {
		return err
	}
	if coupon.Status == common.TopUpCouponStatusRevoked {
		return errors.New("优惠券已撤销")
	}
	if coupon.Status == common.TopUpCouponStatusUsed {
		return errors.New("优惠券已使用")
	}
	if coupon.Status == common.TopUpCouponStatusExpired {
		return errors.New("优惠券已过期")
	}
	if coupon.Status == common.TopUpCouponStatusReserved {
		return errors.New("优惠券已被占用")
	}
	if coupon.Status != common.TopUpCouponStatusAvailable {
		return errors.New("优惠券不可用")
	}
	return nil
}

func normalizeTopUpCouponStateTx(tx *gorm.DB, coupon *TopUpCoupon, now int64) error {
	if coupon == nil {
		return nil
	}

	switch coupon.Status {
	case common.TopUpCouponStatusAvailable:
		if coupon.ExpiresAt == 0 || coupon.ExpiresAt >= now {
			return nil
		}
		coupon.Status = common.TopUpCouponStatusExpired
		coupon.UpdatedTime = now
		return tx.Save(coupon).Error
	case common.TopUpCouponStatusReserved:
		stale, shouldExpireTopUp, err := isCouponReservationStaleTx(tx, coupon, now)
		if err != nil {
			return err
		}
		if !stale {
			return nil
		}

		if shouldExpireTopUp && coupon.ReservedTopUpId != 0 {
			topUp := &TopUp{}
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", coupon.ReservedTopUpId).First(topUp).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			} else if topUp.Status == common.TopUpStatusPending {
				topUp.Status = common.TopUpStatusExpired
				if err := tx.Save(topUp).Error; err != nil {
					return err
				}
			}
		}

		coupon.ReservedTopUpId = 0
		coupon.ReservedAt = 0
		coupon.UpdatedTime = now
		if coupon.ExpiresAt != 0 && coupon.ExpiresAt < now {
			coupon.Status = common.TopUpCouponStatusExpired
		} else {
			coupon.Status = common.TopUpCouponStatusAvailable
		}
		return tx.Save(coupon).Error
	}

	return nil
}

func isCouponReservationStaleTx(tx *gorm.DB, coupon *TopUpCoupon, now int64) (bool, bool, error) {
	if coupon.ReservedTopUpId == 0 {
		return true, false, nil
	}

	topUp := &TopUp{}
	if err := tx.Where("id = ?", coupon.ReservedTopUpId).First(topUp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, false, nil
		}
		return false, false, err
	}

	if topUp.Status == common.TopUpStatusExpired {
		return true, false, nil
	}
	if topUp.Status != common.TopUpStatusPending {
		return false, false, nil
	}

	return now-topUp.CreateTime >= topUpCouponReservationTTLSeconds, true, nil
}

func getCouponEffectiveStatusTx(tx *gorm.DB, coupon *TopUpCoupon, now int64) (string, error) {
	if coupon == nil {
		return "", nil
	}

	switch coupon.Status {
	case common.TopUpCouponStatusAvailable:
		if coupon.ExpiresAt != 0 && coupon.ExpiresAt < now {
			return common.TopUpCouponStatusExpired, nil
		}
	case common.TopUpCouponStatusReserved:
		stale, _, err := isCouponReservationStaleTx(tx, coupon, now)
		if err != nil {
			return "", err
		}
		if stale {
			if coupon.ExpiresAt != 0 && coupon.ExpiresAt < now {
				return common.TopUpCouponStatusExpired, nil
			}
			return common.TopUpCouponStatusAvailable, nil
		}
	}

	return coupon.Status, nil
}

func fillCouponUsernames(coupons []*TopUpCoupon) error {
	if len(coupons) == 0 {
		return nil
	}

	ids := make([]int, 0, len(coupons))
	seen := make(map[int]struct{})
	for _, coupon := range coupons {
		if coupon == nil || coupon.BoundUserId == 0 {
			continue
		}
		if _, ok := seen[coupon.BoundUserId]; ok {
			continue
		}
		seen[coupon.BoundUserId] = struct{}{}
		ids = append(ids, coupon.BoundUserId)
	}
	if len(ids) == 0 {
		return nil
	}

	type userNameRow struct {
		Id       int    `gorm:"column:id"`
		Username string `gorm:"column:username"`
	}
	var users []userNameRow
	if err := DB.Model(&User{}).Select("id, username").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return err
	}

	nameMap := make(map[int]string, len(users))
	for _, user := range users {
		nameMap[user.Id] = user.Username
	}
	for _, coupon := range coupons {
		coupon.BoundUsername = nameMap[coupon.BoundUserId]
	}
	return nil
}

func fillCouponEffectiveStatuses(coupons []*TopUpCoupon) error {
	if len(coupons) == 0 {
		return nil
	}

	now := common.GetTimestamp()
	return DB.Transaction(func(tx *gorm.DB) error {
		for _, coupon := range coupons {
			effectiveStatus, err := getCouponEffectiveStatusTx(tx, coupon, now)
			if err != nil {
				return err
			}
			coupon.EffectiveStatus = effectiveStatus
		}
		return nil
	})
}

func fillCouponViewData(coupons []*TopUpCoupon) error {
	if err := fillCouponUsernames(coupons); err != nil {
		return err
	}
	return fillCouponEffectiveStatuses(coupons)
}

func (coupon *TopUpCoupon) Validate() error {
	name := strings.TrimSpace(coupon.Name)
	if name == "" {
		return errors.New("优惠券名称不能为空")
	}
	if len([]rune(name)) > 50 {
		return errors.New("优惠券名称长度不能超过 50")
	}
	if coupon.BoundUserId == 0 {
		return errors.New("请选择用户")
	}
	if coupon.DeductionAmount <= 0 {
		return errors.New("优惠金额必须大于 0")
	}
	coupon.CurrencyCode = NormalizeTopUpCouponCurrencyCode(coupon.CurrencyCode)
	if coupon.CurrencyCode != "" && !isValidTopUpCouponCurrencyCode(coupon.CurrencyCode) {
		return errors.New("优惠货币格式不正确")
	}
	if coupon.ExpiresAt != 0 && coupon.ValidFrom != 0 && coupon.ExpiresAt <= coupon.ValidFrom {
		return errors.New("过期时间必须晚于生效时间")
	}
	if coupon.ExpiresAt != 0 && coupon.ExpiresAt < common.GetTimestamp() {
		return errors.New("过期时间不能早于当前时间")
	}
	coupon.Name = name
	return nil
}

func (coupon *TopUpCoupon) StatusText() string {
	return fmt.Sprintf("%s", coupon.Status)
}
