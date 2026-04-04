package controller

import (
	"reflect"
	"testing"

	"github.com/stripe/stripe-go/v81"
)

func TestResolveStripePriceContextUsesRequestedCurrencyOption(t *testing.T) {
	priceInfo := &stripe.Price{
		ID:         "price_test",
		Currency:   stripe.CurrencyUSD,
		UnitAmount: 500,
		CurrencyOptions: map[string]*stripe.PriceCurrencyOptions{
			"eur": {UnitAmount: 450},
			"cny": {UnitAmount: 3600},
		},
	}

	resolved, err := resolveStripePriceContext(priceInfo, "eur")
	if err != nil {
		t.Fatalf("expected resolveStripePriceContext to succeed, got error: %v", err)
	}
	if resolved.Currency != stripe.CurrencyEUR {
		t.Fatalf("expected resolved currency EUR, got %s", resolved.Currency)
	}
	if resolved.DefaultCurrency != stripe.CurrencyUSD {
		t.Fatalf("expected default currency USD, got %s", resolved.DefaultCurrency)
	}
	if resolved.UnitAmount != 450 {
		t.Fatalf("expected resolved unit amount 450, got %d", resolved.UnitAmount)
	}

	expectedSupportedCurrencies := []string{"USD", "CNY", "EUR"}
	if !reflect.DeepEqual(resolved.SupportedCurrencyCodes, expectedSupportedCurrencies) {
		t.Fatalf("expected supported currencies %v, got %v", expectedSupportedCurrencies, resolved.SupportedCurrencyCodes)
	}
}

func TestResolveStripePriceContextDefaultsToPrimaryCurrency(t *testing.T) {
	priceInfo := &stripe.Price{
		ID:         "price_test",
		Currency:   stripe.CurrencyUSD,
		UnitAmount: 500,
		CurrencyOptions: map[string]*stripe.PriceCurrencyOptions{
			"eur": {UnitAmount: 450},
		},
	}

	resolved, err := resolveStripePriceContext(priceInfo, "")
	if err != nil {
		t.Fatalf("expected resolveStripePriceContext to succeed, got error: %v", err)
	}
	if resolved.Currency != stripe.CurrencyUSD {
		t.Fatalf("expected resolved currency USD, got %s", resolved.Currency)
	}
	if resolved.UnitAmount != 500 {
		t.Fatalf("expected resolved unit amount 500, got %d", resolved.UnitAmount)
	}
}

func TestResolveStripePriceContextRejectsUnsupportedCurrency(t *testing.T) {
	priceInfo := &stripe.Price{
		ID:         "price_test",
		Currency:   stripe.CurrencyUSD,
		UnitAmount: 500,
		CurrencyOptions: map[string]*stripe.PriceCurrencyOptions{
			"eur": {UnitAmount: 450},
		},
	}

	if _, err := resolveStripePriceContext(priceInfo, "jpy"); err == nil {
		t.Fatal("expected resolveStripePriceContext to reject unsupported currency")
	}
}
