package model

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	quotaPerUnitRawDataMigrationKey = "QuotaPerUnitRawDataMigratedTo1000000"
	quotaPerUnitLogMigrationKey     = "QuotaPerUnitLogDataMigratedTo1000000"
	quotaPerUnitMigrationSourceKey  = "QuotaPerUnitRawDataMigrationSource"
)

func quotaPerUnitMigrationMultiplier() (int64, error) {
	ratio := decimal.NewFromFloat(common.DefaultQuotaPerUnit).Div(decimal.NewFromFloat(common.LegacyDefaultQuotaPerUnit))
	rounded := ratio.Round(0)
	if !ratio.Equal(rounded) || rounded.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("unsupported quota per unit migration ratio %s", ratio.String())
	}
	return rounded.IntPart(), nil
}

func getOptionValueTx(tx *gorm.DB, key string) (string, bool, error) {
	var option Option
	err := tx.Where("key = ?", key).First(&option).Error
	if err == nil {
		return option.Value, true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return "", false, nil
	}
	return "", false, err
}

func upsertOptionTx(tx *gorm.DB, key string, value string) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&Option{Key: key, Value: value}).Error
}

func optionIsTrueTx(tx *gorm.DB, key string) (bool, error) {
	value, found, err := getOptionValueTx(tx, key)
	if err != nil || !found {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(value), "true"), nil
}

func shouldMigrateLegacyDefaultQuotaPerUnitTx(tx *gorm.DB, markerKey string) (bool, error) {
	migrated, err := optionIsTrueTx(tx, markerKey)
	if err != nil || migrated {
		return false, err
	}

	value, found, err := getOptionValueTx(tx, "QuotaPerUnit")
	if err != nil {
		return false, err
	}
	if !found {
		return true, nil
	}

	quotaPerUnit, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || quotaPerUnit <= 0 {
		return true, nil
	}
	return math.Abs(quotaPerUnit-common.LegacyDefaultQuotaPerUnit) < 0.000001, nil
}

func scaleColumnsTx(tx *gorm.DB, modelValue any, multiplier int64, columns ...string) error {
	updates := make(map[string]interface{}, len(columns))
	for _, column := range columns {
		updates[column] = gorm.Expr(column+" * ?", multiplier)
	}
	return tx.Unscoped().Model(modelValue).Where("1 = 1").Updates(updates).Error
}

func scaleTopUpRawQuotaAmountsTx(tx *gorm.DB, multiplier int64) error {
	return tx.Model(&TopUp{}).
		Where("payment_method IN ?", []string{"creem", ""}).
		Update("amount", gorm.Expr("amount * ?", multiplier)).
		Error
}

func scaleIntegerOptionTx(tx *gorm.DB, key string, multiplier int64) error {
	value, found, err := getOptionValueTx(tx, key)
	if err != nil || !found {
		return err
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil
	}
	return upsertOptionTx(tx, key, strconv.FormatInt(parsed*multiplier, 10))
}

func scaleJSONNumberValue(value any, multiplier int64) (any, bool) {
	switch typed := value.(type) {
	case float64:
		return typed * float64(multiplier), true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return value, false
		}
		if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
			scaled := parsed * float64(multiplier)
			if math.Trunc(scaled) == scaled {
				return strconv.FormatInt(int64(scaled), 10), true
			}
			return strconv.FormatFloat(scaled, 'f', -1, 64), true
		}
		return value, false
	default:
		return value, false
	}
}

func scaleJSONNumberField(object map[string]any, field string, multiplier int64) bool {
	value, ok := object[field]
	if !ok {
		return false
	}
	scaled, changed := scaleJSONNumberValue(value, multiplier)
	if changed {
		object[field] = scaled
	}
	return changed
}

func scaleUserJSONQuotaFieldsTx(tx *gorm.DB, multiplier int64) error {
	type userQuotaJSONRow struct {
		Id               int
		Setting          string
		SubscriptionData string
	}

	var users []userQuotaJSONRow
	if err := tx.Model(&User{}).
		Unscoped().
		Select("id", "setting", "subscription_data").
		Where("setting <> '' OR subscription_data <> ''").
		Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		updates := map[string]interface{}{}

		if strings.TrimSpace(user.Setting) != "" {
			var setting map[string]any
			if err := json.Unmarshal([]byte(user.Setting), &setting); err == nil {
				if scaleJSONNumberField(setting, "quota_warning_threshold", multiplier) {
					if encoded, err := json.Marshal(setting); err == nil {
						updates["setting"] = string(encoded)
					}
				}
			}
		}

		if strings.TrimSpace(user.SubscriptionData) != "" {
			var data SubscriptionData
			if err := json.Unmarshal([]byte(user.SubscriptionData), &data); err == nil {
				for i := range data.Items {
					data.Items[i].Limit5H.Total *= multiplier
					data.Items[i].Limit5H.Available *= multiplier
					data.Items[i].Limit7D.Total *= multiplier
					data.Items[i].Limit7D.Available *= multiplier
				}
				if encoded, err := json.Marshal(data); err == nil {
					updates["subscription_data"] = string(encoded)
				}
			}
		}

		if len(updates) > 0 {
			if err := tx.Unscoped().Model(&User{}).Where("id = ?", user.Id).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func scaleCreemProductQuotaOptionTx(tx *gorm.DB, multiplier int64) error {
	value, found, err := getOptionValueTx(tx, "CreemProducts")
	if err != nil || !found || strings.TrimSpace(value) == "" {
		return err
	}

	var products []map[string]any
	if err := json.Unmarshal([]byte(value), &products); err != nil {
		return nil
	}

	changed := false
	for _, product := range products {
		if scaleJSONNumberField(product, "quota", multiplier) {
			changed = true
		}
	}
	if !changed {
		return nil
	}

	encoded, err := json.Marshal(products)
	if err != nil {
		return err
	}
	return upsertOptionTx(tx, "CreemProducts", string(encoded))
}

func scaleQuotaOptionValuesTx(tx *gorm.DB, multiplier int64) error {
	for _, key := range []string{
		"QuotaForNewUser",
		"QuotaForInviter",
		"QuotaForInvitee",
		"QuotaRemindThreshold",
		"PreConsumedQuota",
	} {
		if err := scaleIntegerOptionTx(tx, key, multiplier); err != nil {
			return err
		}
	}
	return scaleCreemProductQuotaOptionTx(tx, multiplier)
}

func migrateLegacyDefaultQuotaPerUnitData(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	multiplier, err := quotaPerUnitMigrationMultiplier()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		shouldMigrate, err := shouldMigrateLegacyDefaultQuotaPerUnitTx(tx, quotaPerUnitRawDataMigrationKey)
		if err != nil || !shouldMigrate {
			return err
		}

		if err := scaleColumnsTx(tx, &User{}, multiplier, "quota", "used_quota", "aff_quota", "aff_history"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &Token{}, multiplier, "remain_quota", "used_quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &Redemption{}, multiplier, "quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &Channel{}, multiplier, "used_quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &Midjourney{}, multiplier, "quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &Task{}, multiplier, "quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &QuotaData{}, multiplier, "quota"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &UserUsageWindow{}, multiplier, "budget_used", "budget_reserved"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &UserUsageReservation{}, multiplier, "reserved_budget"); err != nil {
			return err
		}
		if err := scaleColumnsTx(tx, &AffRebateLog{}, multiplier, "reward_quota"); err != nil {
			return err
		}
		if err := scaleTopUpRawQuotaAmountsTx(tx, multiplier); err != nil {
			return err
		}
		if err := scaleUserJSONQuotaFieldsTx(tx, multiplier); err != nil {
			return err
		}
		if err := scaleQuotaOptionValuesTx(tx, multiplier); err != nil {
			return err
		}

		if err := upsertOptionTx(tx, quotaPerUnitMigrationSourceKey, strconv.FormatFloat(common.LegacyDefaultQuotaPerUnit, 'f', -1, 64)); err != nil {
			return err
		}
		if err := upsertOptionTx(tx, "QuotaPerUnit", strconv.FormatFloat(common.DefaultQuotaPerUnit, 'f', -1, 64)); err != nil {
			return err
		}
		if err := upsertOptionTx(tx, quotaPerUnitRawDataMigrationKey, "true"); err != nil {
			return err
		}

		common.SysLog(fmt.Sprintf("migrated legacy quota data from quotaPerUnit %.0f to %.0f", common.LegacyDefaultQuotaPerUnit, common.DefaultQuotaPerUnit))
		return nil
	})
}

func migrateLegacyDefaultQuotaPerUnitLogData(logDB *gorm.DB) error {
	if DB == nil || logDB == nil {
		return nil
	}
	multiplier, err := quotaPerUnitMigrationMultiplier()
	if err != nil {
		return err
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		source, found, err := getOptionValueTx(tx, quotaPerUnitMigrationSourceKey)
		if err != nil || !found {
			return err
		}
		sourceQuotaPerUnit, err := strconv.ParseFloat(strings.TrimSpace(source), 64)
		if err != nil || math.Abs(sourceQuotaPerUnit-common.LegacyDefaultQuotaPerUnit) >= 0.000001 {
			return nil
		}
		migrated, err := optionIsTrueTx(tx, quotaPerUnitLogMigrationKey)
		if err != nil || migrated {
			return err
		}

		targetLogDB := logDB
		if logDB == DB {
			targetLogDB = tx
		}
		if err := targetLogDB.Unscoped().
			Model(&Log{}).
			Where("1 = 1").
			Update("quota", gorm.Expr("quota * ?", multiplier)).
			Error; err != nil {
			return err
		}
		if err := upsertOptionTx(tx, quotaPerUnitLogMigrationKey, "true"); err != nil {
			return err
		}

		common.SysLog(fmt.Sprintf("migrated legacy log quota data from quotaPerUnit %.0f to %.0f", common.LegacyDefaultQuotaPerUnit, common.DefaultQuotaPerUnit))
		return nil
	})
}
