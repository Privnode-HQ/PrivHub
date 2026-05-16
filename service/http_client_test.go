package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
)

func TestUpstreamHTTPClientTimeoutBumpsShortRelayTimeout(t *testing.T) {
	original := common.RelayTimeout
	t.Cleanup(func() {
		common.RelayTimeout = original
	})

	common.RelayTimeout = 30

	if got := upstreamHTTPClientTimeout(); got != DefaultLongUpstreamRequestTimeout {
		t.Fatalf("expected short relay timeout to be bumped to %s, got %s", DefaultLongUpstreamRequestTimeout, got)
	}
}

func TestUpstreamHTTPClientTimeoutPreservesUnlimitedRelayTimeout(t *testing.T) {
	original := common.RelayTimeout
	t.Cleanup(func() {
		common.RelayTimeout = original
	})

	common.RelayTimeout = 0

	if got := upstreamHTTPClientTimeout(); got != 0 {
		t.Fatalf("expected unlimited relay timeout to stay unlimited, got %s", got)
	}
}

func TestWithoutTotalTimeoutClonesConfiguredClient(t *testing.T) {
	client := &http.Client{Timeout: time.Second}

	streamClient := WithoutTotalTimeout(client)
	if streamClient == client {
		t.Fatal("expected a configured client to be cloned")
	}
	if streamClient.Timeout != 0 {
		t.Fatalf("expected cloned client total timeout to be disabled, got %s", streamClient.Timeout)
	}
	if client.Timeout != time.Second {
		t.Fatalf("expected original client timeout to remain unchanged, got %s", client.Timeout)
	}
}
