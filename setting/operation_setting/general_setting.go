package operation_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// 额度展示类型
const (
	QuotaDisplayTypeUSD = "USD"
	QuotaDisplayTypeCNY = "CNY"
)

type GeneralSetting struct {
	DocsLink            string `json:"docs_link"`
	PingIntervalEnabled bool   `json:"ping_interval_enabled"`
	PingIntervalSeconds int    `json:"ping_interval_seconds"`
	// 当前站点额度展示类型：USD / CNY。币种只影响 quotaPerUnit 的展示符号，不做汇率换算。
	QuotaDisplayType string `json:"quota_display_type"`
}

// 默认配置
var generalSetting = GeneralSetting{
	DocsLink:            "https://docs.newapi.pro",
	PingIntervalEnabled: false,
	PingIntervalSeconds: 60,
	QuotaDisplayType:    QuotaDisplayTypeUSD,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("general_setting", &generalSetting)
}

func GetGeneralSetting() *GeneralSetting {
	return &generalSetting
}

func NormalizeQuotaDisplayType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case QuotaDisplayTypeCNY:
		return QuotaDisplayTypeCNY
	default:
		return QuotaDisplayTypeUSD
	}
}

func SetQuotaDisplayType(value string) {
	generalSetting.QuotaDisplayType = NormalizeQuotaDisplayType(value)
}

// IsCurrencyDisplay 是否以货币形式展示。当前额度展示始终使用全局货币。
func IsCurrencyDisplay() bool {
	return true
}

// IsCNYDisplay 是否以人民币展示
func IsCNYDisplay() bool {
	return GetQuotaDisplayType() == QuotaDisplayTypeCNY
}

// GetQuotaDisplayType 返回额度展示类型
func GetQuotaDisplayType() string {
	return NormalizeQuotaDisplayType(generalSetting.QuotaDisplayType)
}

// GetCurrencySymbol 返回当前展示类型对应符号
func GetCurrencySymbol() string {
	switch GetQuotaDisplayType() {
	case QuotaDisplayTypeUSD:
		return "$"
	case QuotaDisplayTypeCNY:
		return "¥"
	default:
		return "$"
	}
}
