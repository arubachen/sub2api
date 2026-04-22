//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPBackendRateLimitResetter_ResetRateLimit_PostsCoolingReset(t *testing.T) {
	var gotAuth string
	var gotPath string
	var gotPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resetter := NewHTTPBackendRateLimitResetter(server.Client())
	account := &Account{
		ID: 99,
		Credentials: map[string]any{
			clearRateLimitSyncCredential: true,
			"base_url":                   server.URL + "/v1",
			"api_key":                    "sync-token",
		},
	}

	err := resetter.ResetRateLimit(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, "Bearer sync-token", gotAuth)
	require.Equal(t, "/v1/admin/tokens/reset", gotPath)
	require.Equal(t, true, gotPayload["cooling_only"])
	require.Equal(t, "sub2api_clear_rate_limit", gotPayload["source"])
}

func TestHTTPBackendRateLimitResetter_ResetRateLimit_SkipsWhenSyncDisabled(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resetter := NewHTTPBackendRateLimitResetter(server.Client())
	account := &Account{
		ID: 100,
		Credentials: map[string]any{
			"base_url": server.URL + "/v1",
			"api_key":  "sync-token",
		},
	}

	err := resetter.ResetRateLimit(context.Background(), account)
	require.NoError(t, err)
	require.False(t, called)
}

func TestHTTPBackendRateLimitResetter_ResetRateLimit_UsesOverrideURLAndKey(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	resetter := NewHTTPBackendRateLimitResetter(server.Client())
	account := &Account{
		ID: 101,
		Credentials: map[string]any{
			clearRateLimitSyncCredential:    "true",
			clearRateLimitSyncURLCredential: server.URL + "/custom/reset",
			clearRateLimitSyncKeyCredential: "override-token",
			"base_url":                      "http://ignored/v1",
			"api_key":                       "sync-token",
		},
	}

	err := resetter.ResetRateLimit(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, "Bearer override-token", gotAuth)
}

func TestHTTPBackendRateLimitResetter_ResetRateLimit_ReturnsRemoteFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer server.Close()

	resetter := NewHTTPBackendRateLimitResetter(server.Client())
	account := &Account{
		ID: 102,
		Credentials: map[string]any{
			clearRateLimitSyncCredential: true,
			"base_url":                   server.URL + "/v1",
			"api_key":                    "sync-token",
		},
	}

	err := resetter.ResetRateLimit(context.Background(), account)
	require.Error(t, err)
	require.Contains(t, err.Error(), "502")
}
