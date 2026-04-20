package service

import (
	"testing"
	"time"
)

func TestClassifyRiskUserAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ua            string
		wantCategory  string
		wantStatus    string
		wantBaseScore int
		wantNormal    bool
	}{
		{name: "normal codex cli", ua: "codex_cli_rs/0.1.0", wantCategory: "Codex CLI", wantStatus: "normal", wantBaseScore: 0, wantNormal: true},
		{name: "abnormal go http 2", ua: "Go-http-client/2.0", wantCategory: "Go HTTP客户端", wantStatus: "abnormal", wantBaseScore: 15},
		{name: "generic go http", ua: "Go-http-client/1.1", wantCategory: "Go HTTP客户端", wantStatus: "unconfigured", wantBaseScore: 10},
		{name: "browser", ua: "Mozilla/5.0 Chrome/122.0 Safari/537.36", wantCategory: "浏览器", wantStatus: "unconfigured", wantBaseScore: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyRiskUserAgent(tt.ua)
			if got.Category != tt.wantCategory {
				t.Fatalf("category = %q, want %q", got.Category, tt.wantCategory)
			}
			if got.ConfigStatus != tt.wantStatus {
				t.Fatalf("status = %q, want %q", got.ConfigStatus, tt.wantStatus)
			}
			if got.BaseScore != tt.wantBaseScore {
				t.Fatalf("base score = %d, want %d", got.BaseScore, tt.wantBaseScore)
			}
			if got.NormalAllowed != tt.wantNormal {
				t.Fatalf("normal allowed = %v, want %v", got.NormalAllowed, tt.wantNormal)
			}
		})
	}
}

func TestBuildUserRiskDetail(t *testing.T) {
	t.Parallel()

	user := &User{ID: 42, Email: "risk@example.com", Username: "risky", Balance: -0.33, Status: StatusActive}
	loc := time.UTC
	windowEnd := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	windowStart := windowEnd.Add(-24 * time.Hour)

	rows := []userRiskUsageRow{
		{CreatedAt: windowStart.Add(10 * time.Minute), ActualCost: 35, IPAddress: "35.238.205.178", UserAgent: "Go-http-client/2.0"},
		{CreatedAt: windowStart.Add(70 * time.Minute), ActualCost: 20, IPAddress: "35.238.205.178", UserAgent: "Go-http-client/2.0"},
		{CreatedAt: windowStart.Add(70 * time.Minute), ActualCost: 5, IPAddress: "34.173.72.158", UserAgent: "OpenAI/JS 6.10.0"},
		{CreatedAt: windowStart.Add(2*time.Hour + 5*time.Minute), ActualCost: 5, IPAddress: "34.173.72.158", UserAgent: "Go-http-client/2.0"},
	}

	detail := buildUserRiskDetail(
		user,
		rows,
		userRiskHistoricalSummary{FirstIP: "35.238.205.178", HistoricalIPCount: 12},
		3,
		windowStart,
		windowEnd,
		loc,
		"UTC",
		defaultUserRiskSettings(),
		nil,
	)

	if detail.Summary.RiskScore <= 0 {
		t.Fatalf("expected positive risk score, got %d", detail.Summary.RiskScore)
	}
	if detail.Metrics.RequestCount24h != 4 {
		t.Fatalf("request_count_24h = %d, want 4", detail.Metrics.RequestCount24h)
	}
	if detail.Metrics.HistoricalIPCount != 12 {
		t.Fatalf("historical_ip_count = %d, want 12", detail.Metrics.HistoricalIPCount)
	}
	if detail.Metrics.ConcurrentMultiIPUAMinutes24h != 1 {
		t.Fatalf("concurrent multi-ip/ua minutes = %d, want 1", detail.Metrics.ConcurrentMultiIPUAMinutes24h)
	}
	if len(detail.RuleHits) == 0 {
		t.Fatal("expected rule hits")
	}
	if got := detail.UADetails[0].UserAgent; got != "Go-http-client/2.0" {
		t.Fatalf("top ua = %q, want Go-http-client/2.0", got)
	}
}

func TestScoreUserRiskDecisionUsesChineseLabels(t *testing.T) {
	t.Parallel()

	settings := defaultUserRiskSettings()
	tests := []struct {
		score     int
		wantCode  string
		wantLabel string
	}{
		{score: 10, wantCode: "normal", wantLabel: "正常"},
		{score: 30, wantCode: "observe", wantLabel: "建议观察"},
		{score: 55, wantCode: "review", wantLabel: "建议人工审查"},
		{score: 85, wantCode: "throttle", wantLabel: "建议限流观察"},
		{score: 130, wantCode: "freeze_review", wantLabel: "建议冻结审查"},
	}

	for _, tt := range tests {
		code, label := scoreUserRiskDecision(tt.score, settings)
		if code != tt.wantCode || label != tt.wantLabel {
			t.Fatalf("score=%d => (%q, %q), want (%q, %q)", tt.score, code, label, tt.wantCode, tt.wantLabel)
		}
	}
}
