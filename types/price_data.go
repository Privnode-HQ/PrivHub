package types

import "fmt"

const tokensPerMillion = 1000 * 1000.0

// ModelRatioInputPricePerMillionTokens converts the stored model ratio to the
// displayed input price. The UI documents model ratio as half of the per-million
// input-token price, and completion/cache/audio ratios are applied on top of
// that input price.
func ModelRatioInputPricePerMillionTokens(modelRatio float64) float64 {
	return modelRatio * 2
}

// ModelRatioTokenQuotaRatio converts a model ratio into raw quota units charged
// per token. It keeps quota-unit changes aligned with the per-million-token
// price shown in billing details.
func ModelRatioTokenQuotaRatio(modelRatio, quotaPerUnit float64) float64 {
	return ModelRatioInputPricePerMillionTokens(modelRatio) * quotaPerUnit / tokensPerMillion
}

// ModelRatioTokenPrice converts a model ratio into the configured currency
// amount charged per token.
func ModelRatioTokenPrice(modelRatio float64) float64 {
	return ModelRatioInputPricePerMillionTokens(modelRatio) / tokensPerMillion
}

type GroupRatioInfo struct {
	GroupRatio        float64
	GroupSpecialRatio float64
	HasSpecialRatio   bool
}

type PriceData struct {
	FreeModel            bool
	ModelPrice           float64
	ModelRatio           float64
	CompletionRatio      float64
	CacheRatio           float64
	CacheCreationRatio   float64
	CacheCreation5mRatio float64
	CacheCreation1hRatio float64
	ImageRatio           float64
	AudioRatio           float64
	AudioCompletionRatio float64
	OtherRatios          map[string]float64
	UsePrice             bool
	QuotaToPreConsume    int // 预消耗额度
	GroupRatioInfo       GroupRatioInfo
}

type PerCallPriceData struct {
	ModelPrice     float64
	Quota          int
	GroupRatioInfo GroupRatioInfo
}

func (p PriceData) ToSetting() string {
	return fmt.Sprintf("ModelPrice: %f, ModelRatio: %f, CompletionRatio: %f, CacheRatio: %f, GroupRatio: %f, UsePrice: %t, CacheCreationRatio: %f, CacheCreation5mRatio: %f, CacheCreation1hRatio: %f, QuotaToPreConsume: %d, ImageRatio: %f, AudioRatio: %f, AudioCompletionRatio: %f", p.ModelPrice, p.ModelRatio, p.CompletionRatio, p.CacheRatio, p.GroupRatioInfo.GroupRatio, p.UsePrice, p.CacheCreationRatio, p.CacheCreation5mRatio, p.CacheCreation1hRatio, p.QuotaToPreConsume, p.ImageRatio, p.AudioRatio, p.AudioCompletionRatio)
}
