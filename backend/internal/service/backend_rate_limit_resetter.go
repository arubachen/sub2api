package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	clearRateLimitSyncCredential    = "clear_rate_limit_sync"
	clearRateLimitSyncURLCredential = "clear_rate_limit_sync_url"
	clearRateLimitSyncKeyCredential = "clear_rate_limit_sync_key"
	clearRateLimitSyncTimeout       = 5 * time.Second
)

// BackendRateLimitResetter 用于在本地清限流前，同步清理外部后端运行时状态。
type BackendRateLimitResetter interface {
	ResetRateLimit(ctx context.Context, account *Account) error
}

// HTTPBackendRateLimitResetter 通过 HTTP 调用外部后端 reset 接口。
type HTTPBackendRateLimitResetter struct {
	client *http.Client
}

func NewHTTPBackendRateLimitResetter(client *http.Client) *HTTPBackendRateLimitResetter {
	if client == nil {
		client = &http.Client{Timeout: clearRateLimitSyncTimeout}
	}
	return &HTTPBackendRateLimitResetter{client: client}
}

func (r *HTTPBackendRateLimitResetter) ResetRateLimit(ctx context.Context, account *Account) error {
	if r == nil || account == nil || !clearRateLimitSyncEnabled(account) {
		return nil
	}

	endpoint, err := resolveClearRateLimitSyncURL(account)
	if err != nil {
		return err
	}

	authToken := strings.TrimSpace(account.GetCredential(clearRateLimitSyncKeyCredential))
	if authToken == "" {
		authToken = strings.TrimSpace(account.GetCredential("api_key"))
	}
	if authToken == "" {
		return fmt.Errorf("clear rate limit sync auth token is empty for account %d", account.ID)
	}

	payloadBytes, err := json.Marshal(map[string]any{
		"cooling_only": true,
		"source":       "sub2api_clear_rate_limit",
		"account_id":   account.ID,
	})
	if err != nil {
		return fmt.Errorf("marshal clear rate limit sync payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("build clear rate limit sync request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("call clear rate limit sync endpoint: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("clear rate limit sync endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func clearRateLimitSyncEnabled(account *Account) bool {
	if account == nil || account.Credentials == nil {
		return false
	}
	raw, ok := account.Credentials[clearRateLimitSyncCredential]
	if !ok || raw == nil {
		return false
	}

	switch v := raw.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		default:
			return false
		}
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

func resolveClearRateLimitSyncURL(account *Account) (string, error) {
	if account == nil {
		return "", fmt.Errorf("account is nil")
	}
	if override := strings.TrimSpace(account.GetCredential(clearRateLimitSyncURLCredential)); override != "" {
		return override, nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(account.GetCredential("base_url")), "/")
	if baseURL == "" {
		return "", fmt.Errorf("clear rate limit sync base_url is empty for account %d", account.ID)
	}
	if strings.HasSuffix(strings.ToLower(baseURL), "/v1") {
		return baseURL + "/admin/tokens/reset", nil
	}
	return baseURL + "/v1/admin/tokens/reset", nil
}
