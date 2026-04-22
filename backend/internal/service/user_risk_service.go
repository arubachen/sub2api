package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	userRiskWindow = 24 * time.Hour

	userRiskIPInfoLiteProvider = "ipinfo_lite"
	userRiskIPInfoLiteDocsURL  = "https://ipinfo.io/developers/lite-api"

	settingKeyUserRiskIPIntelEnabled    = "user_risk_ip_intel_enabled"
	settingKeyUserRiskIPIntelProvider   = "user_risk_ip_intel_provider"
	settingKeyUserRiskIPIntelToken      = "user_risk_ip_intel_token"
	settingKeyUserRiskReviewThreshold   = "user_risk_review_threshold"
	settingKeyUserRiskThrottleThreshold = "user_risk_throttle_threshold"
	settingKeyUserRiskFreezeThreshold   = "user_risk_freeze_threshold"
	settingKeyUserRiskAutoEnabled       = "user_risk_auto_enabled"
	settingKeyUserRiskAutoThrottle      = "user_risk_auto_throttle"
	settingKeyUserRiskAutoFreeze        = "user_risk_auto_freeze"
	settingKeyUserRiskAutoThrottleCap   = "user_risk_auto_throttle_concurrency_cap"

	userRiskAutomationCacheTTL = 3 * time.Minute
)

var (
	userRiskErrDBUnavailable = infraerrors.InternalServer("USER_RISK_DB_UNAVAILABLE", "user risk database unavailable")
)

// UserRiskService computes rolling user risk details and exposes module settings.
type UserRiskService struct {
	db          *sql.DB
	userRepo    UserRepository
	settingRepo SettingRepository
	httpClient  *http.Client
	autoCache   sync.Map
}

func NewUserRiskService(db *sql.DB, userRepo UserRepository, settingRepo SettingRepository) *UserRiskService {
	return &UserRiskService{
		db:          db,
		userRepo:    userRepo,
		settingRepo: settingRepo,
		httpClient:  &http.Client{Timeout: 12 * time.Second},
	}
}

type UserRiskDetail struct {
	User      UserRiskUser       `json:"user"`
	Window    UserRiskWindow     `json:"window"`
	Summary   UserRiskSummary    `json:"summary"`
	Metrics   UserRiskMetrics    `json:"metrics"`
	RuleHits  []UserRiskRuleHit  `json:"rule_hits"`
	IPDetails []UserRiskIPDetail `json:"ip_details"`
	UADetails []UserRiskUADetail `json:"ua_details"`
}

type UserRiskUser struct {
	ID       int64   `json:"id"`
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
	Status   string  `json:"status"`
}

type UserRiskWindow struct {
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Timezone string    `json:"timezone"`
}

type UserRiskSummary struct {
	RiskScore     int       `json:"risk_score"`
	Decision      string    `json:"decision"`
	DecisionLabel string    `json:"decision_label"`
	ComputedAt    time.Time `json:"computed_at"`
}

type UserRiskMetrics struct {
	RequestCount24h               int64   `json:"request_count_24h"`
	ActualCost24h                 float64 `json:"actual_cost_24h"`
	FirstIP                       string  `json:"first_ip,omitempty"`
	HistoricalIPCount             int     `json:"historical_ip_count"`
	UA24hCount                    int     `json:"ua_24h_count"`
	ActiveHoursCount              int     `json:"active_hours_count"`
	ActiveHours                   []int   `json:"active_hours"`
	LongestSilenceHours           float64 `json:"longest_silence_hours"`
	AllDayActive                  bool    `json:"all_day_active"`
	HourConcentration             float64 `json:"hour_concentration"`
	KeyCount                      int     `json:"key_count"`
	ConcurrentMultiIPUAMinutes24h int     `json:"concurrent_multi_ip_ua_minutes_24h"`
}

type UserRiskRuleHit struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Score       int    `json:"score"`
}

type UserRiskIPDetail struct {
	IPAddress    string `json:"ip_address"`
	Requests     int64  `json:"requests"`
	IPType       string `json:"ip_type"`
	Label        string `json:"label"`
	ASN          string `json:"asn,omitempty"`
	Organization string `json:"organization,omitempty"`
	Domain       string `json:"domain,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	Country      string `json:"country,omitempty"`
	Continent    string `json:"continent,omitempty"`
}

type UserRiskUADetail struct {
	UserAgent     string `json:"user_agent"`
	Requests      int64  `json:"requests"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	BaseScore     int    `json:"base_score"`
	ConfigStatus  string `json:"config_status"`
	HitRule       string `json:"hit_rule,omitempty"`
	Programmatic  bool   `json:"programmatic"`
	NormalAllowed bool   `json:"normal_allowed"`
}

type UserRiskSummaryItem struct {
	UserID       int64           `json:"user_id"`
	Summary      UserRiskSummary `json:"summary"`
	Metrics      UserRiskMetrics `json:"metrics"`
	RuleHitCount int             `json:"rule_hit_count"`
	TopUserAgent string          `json:"top_user_agent,omitempty"`
}

type UserRiskSettings struct {
	IPIntelEnabled         bool   `json:"ip_intel_enabled"`
	IPIntelProvider        string `json:"ip_intel_provider"`
	IPIntelTokenConfigured bool   `json:"ip_intel_token_configured"`
	IPIntelDocsURL         string `json:"ip_intel_docs_url"`
	ReviewThreshold        int    `json:"review_threshold"`
	ThrottleThreshold      int    `json:"throttle_threshold"`
	FreezeThreshold        int    `json:"freeze_threshold"`
	AutoEnabled            bool   `json:"auto_enabled"`
	AutoThrottle           bool   `json:"auto_throttle"`
	AutoFreeze             bool   `json:"auto_freeze"`
	AutoThrottleCap        int    `json:"auto_throttle_concurrency_cap"`
}

type UserRiskSettingsUpdate struct {
	IPIntelEnabled    *bool   `json:"ip_intel_enabled"`
	IPIntelProvider   *string `json:"ip_intel_provider"`
	IPIntelToken      *string `json:"ip_intel_token"`
	ClearIPIntelToken *bool   `json:"clear_ip_intel_token"`
	ReviewThreshold   *int    `json:"review_threshold"`
	ThrottleThreshold *int    `json:"throttle_threshold"`
	FreezeThreshold   *int    `json:"freeze_threshold"`
	AutoEnabled       *bool   `json:"auto_enabled"`
	AutoThrottle      *bool   `json:"auto_throttle"`
	AutoFreeze        *bool   `json:"auto_freeze"`
	AutoThrottleCap   *int    `json:"auto_throttle_concurrency_cap"`
}

type UserRiskAutomationDecision struct {
	Enabled                 bool
	Blocked                 bool
	EffectiveConcurrencyCap int
	Decision                string
	Message                 string
	Score                   int
}

type userRiskAutomationCacheEntry struct {
	Decision  UserRiskAutomationDecision
	ExpiresAt time.Time
}

type userRiskUsageRow struct {
	CreatedAt  time.Time
	ActualCost float64
	IPAddress  string
	UserAgent  string
}

type userRiskHistoricalSummary struct {
	FirstIP           string
	HistoricalIPCount int
}

type userRiskUASummary struct {
	TopDetail         UserRiskUADetail
	TopShare          float64
	AbnormalCount     int
	AbnormalUserAgent []string
}

type userRiskMinuteBucket struct {
	IPs map[string]struct{}
	UAs map[string]struct{}
}

type userAgentProfile struct {
	Category      string
	Description   string
	BaseScore     int
	ConfigStatus  string
	HitRule       string
	Programmatic  bool
	NormalAllowed bool
}

type userRiskSettingsInternal struct {
	Settings     UserRiskSettings
	IPIntelToken string
}

type userRiskIPIntelRecord struct {
	IP            string `json:"ip"`
	ASN           string `json:"asn"`
	ASName        string `json:"as_name"`
	ASDomain      string `json:"as_domain"`
	CountryCode   string `json:"country_code"`
	Country       string `json:"country"`
	ContinentCode string `json:"continent_code"`
	Continent     string `json:"continent"`
}

var userRiskNormalUAPrefixes = []string{
	"codex_cli_rs/",
	"codex_vscode/",
	"codex desktop/",
	"codex_app/",
	"codex_chatgpt_desktop/",
	"codex_atlas/",
	"codex_exec/",
	"codex_sdk_ts/",
}

var userRiskAbnormalUAExactScores = map[string]int{
	"go-http-client/2.0": 15,
}

func defaultUserRiskSettings() UserRiskSettings {
	return UserRiskSettings{
		IPIntelEnabled:         false,
		IPIntelProvider:        userRiskIPInfoLiteProvider,
		IPIntelTokenConfigured: false,
		IPIntelDocsURL:         userRiskIPInfoLiteDocsURL,
		ReviewThreshold:        50,
		ThrottleThreshold:      80,
		FreezeThreshold:        120,
		AutoEnabled:            false,
		AutoThrottle:           false,
		AutoFreeze:             false,
		AutoThrottleCap:        1,
	}
}

func (s *UserRiskService) GetSettings(ctx context.Context) (*UserRiskSettings, error) {
	settings, err := s.loadSettingsInternal(ctx)
	if err != nil {
		return nil, err
	}
	copy := settings.Settings
	return &copy, nil
}

func (s *UserRiskService) UpdateSettings(ctx context.Context, update *UserRiskSettingsUpdate) (*UserRiskSettings, error) {
	if update == nil {
		return s.GetSettings(ctx)
	}
	current, err := s.loadSettingsInternal(ctx)
	if err != nil {
		return nil, err
	}
	settings := current.Settings
	secret := current.IPIntelToken

	if update.IPIntelEnabled != nil {
		settings.IPIntelEnabled = *update.IPIntelEnabled
	}
	if update.IPIntelProvider != nil {
		settings.IPIntelProvider = strings.TrimSpace(*update.IPIntelProvider)
	}
	if update.ReviewThreshold != nil {
		settings.ReviewThreshold = *update.ReviewThreshold
	}
	if update.ThrottleThreshold != nil {
		settings.ThrottleThreshold = *update.ThrottleThreshold
	}
	if update.FreezeThreshold != nil {
		settings.FreezeThreshold = *update.FreezeThreshold
	}
	if update.AutoEnabled != nil {
		settings.AutoEnabled = *update.AutoEnabled
	}
	if update.AutoThrottle != nil {
		settings.AutoThrottle = *update.AutoThrottle
	}
	if update.AutoFreeze != nil {
		settings.AutoFreeze = *update.AutoFreeze
	}
	if update.AutoThrottleCap != nil {
		settings.AutoThrottleCap = *update.AutoThrottleCap
	}
	if update.ClearIPIntelToken != nil && *update.ClearIPIntelToken {
		secret = ""
	}
	if update.IPIntelToken != nil {
		secret = strings.TrimSpace(*update.IPIntelToken)
	}
	settings.IPIntelTokenConfigured = strings.TrimSpace(secret) != ""

	if err := validateUserRiskSettings(settings); err != nil {
		return nil, err
	}
	if s.settingRepo == nil {
		return nil, infraerrors.InternalServer("USER_RISK_SETTINGS_UNAVAILABLE", "user risk settings repository unavailable")
	}

	values := map[string]string{
		settingKeyUserRiskIPIntelEnabled:    userRiskFormatBool(settings.IPIntelEnabled),
		settingKeyUserRiskIPIntelProvider:   settings.IPIntelProvider,
		settingKeyUserRiskIPIntelToken:      secret,
		settingKeyUserRiskReviewThreshold:   userRiskFormatInt(settings.ReviewThreshold),
		settingKeyUserRiskThrottleThreshold: userRiskFormatInt(settings.ThrottleThreshold),
		settingKeyUserRiskFreezeThreshold:   userRiskFormatInt(settings.FreezeThreshold),
		settingKeyUserRiskAutoEnabled:       userRiskFormatBool(settings.AutoEnabled),
		settingKeyUserRiskAutoThrottle:      userRiskFormatBool(settings.AutoThrottle),
		settingKeyUserRiskAutoFreeze:        userRiskFormatBool(settings.AutoFreeze),
		settingKeyUserRiskAutoThrottleCap:   userRiskFormatInt(settings.AutoThrottleCap),
	}
	if err := s.settingRepo.SetMultiple(ctx, values); err != nil {
		return nil, fmt.Errorf("set user risk settings: %w", err)
	}
	s.autoCache.Range(func(key, _ any) bool {
		s.autoCache.Delete(key)
		return true
	})
	return s.GetSettings(ctx)
}

func (s *UserRiskService) GetUserRiskSummaries(ctx context.Context, userIDs []int64, timezone string) ([]UserRiskSummaryItem, error) {
	settings, err := s.loadSettingsInternal(ctx)
	if err != nil {
		return nil, err
	}
	ids := normalizePositiveUserRiskIDs(userIDs)
	if len(ids) == 0 {
		return []UserRiskSummaryItem{}, nil
	}

	items := make([]UserRiskSummaryItem, len(ids))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for idx, userID := range ids {
		idx := idx
		userID := userID
		g.Go(func() error {
			detail, err := s.getUserRiskDetailWithSettings(gctx, userID, timezone, settings, false)
			if err != nil {
				return err
			}
			topUA := ""
			if len(detail.UADetails) > 0 {
				topUA = detail.UADetails[0].UserAgent
			}
			items[idx] = UserRiskSummaryItem{
				UserID:       userID,
				Summary:      detail.Summary,
				Metrics:      detail.Metrics,
				RuleHitCount: len(detail.RuleHits),
				TopUserAgent: topUA,
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *UserRiskService) GetUserRiskDetail(ctx context.Context, userID int64, timezone string) (*UserRiskDetail, error) {
	settings, err := s.loadSettingsInternal(ctx)
	if err != nil {
		return nil, err
	}
	return s.getUserRiskDetailWithSettings(ctx, userID, timezone, settings, true)
}

func (s *UserRiskService) EvaluateAutomation(ctx context.Context, userID int64) (*UserRiskAutomationDecision, error) {
	if s == nil {
		return &UserRiskAutomationDecision{}, nil
	}
	if cached, ok := s.autoCache.Load(userID); ok {
		if entry, ok := cached.(userRiskAutomationCacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
			decision := entry.Decision
			return &decision, nil
		}
	}

	settings, err := s.loadSettingsInternal(ctx)
	if err != nil {
		return nil, err
	}
	decision := UserRiskAutomationDecision{
		Enabled:                 settings.Settings.AutoEnabled,
		EffectiveConcurrencyCap: 0,
	}
	if !settings.Settings.AutoEnabled {
		s.autoCache.Store(userID, userRiskAutomationCacheEntry{Decision: decision, ExpiresAt: time.Now().Add(userRiskAutomationCacheTTL)})
		return &decision, nil
	}

	detail, err := s.getUserRiskDetailWithSettings(ctx, userID, "", settings, false)
	if err != nil {
		return nil, err
	}
	score := detail.Summary.RiskScore
	decision.Score = score
	decision.Decision = detail.Summary.Decision
	if settings.Settings.AutoFreeze && score >= settings.Settings.FreezeThreshold {
		decision.Blocked = true
		decision.Message = "用户触发自动冻结审查，请联系管理员"
	} else if settings.Settings.AutoThrottle && score >= settings.Settings.ThrottleThreshold {
		cap := settings.Settings.AutoThrottleCap
		if cap < 1 {
			cap = 1
		}
		decision.EffectiveConcurrencyCap = cap
		decision.Message = "用户触发自动限流观察，系统已下调并发上限"
	}
	s.autoCache.Store(userID, userRiskAutomationCacheEntry{Decision: decision, ExpiresAt: time.Now().Add(userRiskAutomationCacheTTL)})
	return &decision, nil
}

func (s *UserRiskService) getUserRiskDetailWithSettings(ctx context.Context, userID int64, timezone string, settings *userRiskSettingsInternal, includeIPIntel bool) (*UserRiskDetail, error) {
	if s == nil || s.db == nil {
		return nil, userRiskErrDBUnavailable
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	location, tzName := loadUserRiskLocation(timezone)
	windowEnd := time.Now().UTC()
	windowStart := windowEnd.Add(-userRiskWindow)

	recentRows, err := s.loadRecentUsageRows(ctx, userID, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("load recent user risk usage rows: %w", err)
	}

	historySummary, err := s.loadHistoricalIPSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user risk historical summary: %w", err)
	}

	keyCount, err := s.loadKeyCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user risk key count: %w", err)
	}

	ipIntel := map[string]userRiskIPIntelRecord{}
	if includeIPIntel {
		ipIntel, err = s.loadIPIntelForRows(ctx, recentRows, settings)
		if err != nil {
			// IP intel is an optional enhancement; keep core risk detail available.
			ipIntel = map[string]userRiskIPIntelRecord{}
		}
	}

	return buildUserRiskDetail(user, recentRows, historySummary, keyCount, windowStart, windowEnd, location, tzName, settings.Settings, ipIntel), nil
}

func (s *UserRiskService) loadRecentUsageRows(ctx context.Context, userID int64, startAt, endAt time.Time) ([]userRiskUsageRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT created_at, COALESCE(actual_cost, 0), COALESCE(ip_address, ''), COALESCE(user_agent, '')
		FROM usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at ASC
	`, userID, startAt, endAt)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make([]userRiskUsageRow, 0, 256)
	for rows.Next() {
		var row userRiskUsageRow
		if err := rows.Scan(&row.CreatedAt, &row.ActualCost, &row.IPAddress, &row.UserAgent); err != nil {
			return nil, err
		}
		row.IPAddress = strings.TrimSpace(row.IPAddress)
		row.UserAgent = strings.TrimSpace(row.UserAgent)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *UserRiskService) loadHistoricalIPSummary(ctx context.Context, userID int64) (userRiskHistoricalSummary, error) {
	var summary userRiskHistoricalSummary

	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(ip_address, '')
		FROM usage_logs
		WHERE user_id = $1 AND ip_address IS NOT NULL AND TRIM(ip_address) <> ''
		ORDER BY created_at ASC
		LIMIT 1
	`, userID).Scan(&summary.FirstIP); err != nil {
		if err != sql.ErrNoRows {
			return summary, err
		}
	}
	summary.FirstIP = strings.TrimSpace(summary.FirstIP)

	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT DISTINCT ip_address
			FROM usage_logs
			WHERE user_id = $1 AND ip_address IS NOT NULL AND TRIM(ip_address) <> ''
		) AS distinct_ips
	`, userID).Scan(&summary.HistoricalIPCount); err != nil {
		return summary, err
	}

	return summary, nil
}

func (s *UserRiskService) loadKeyCount(ctx context.Context, userID int64) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM api_keys
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *UserRiskService) loadSettingsInternal(ctx context.Context) (*userRiskSettingsInternal, error) {
	settings := defaultUserRiskSettings()
	internal := &userRiskSettingsInternal{Settings: settings}
	if s.settingRepo == nil {
		return internal, nil
	}
	values, err := s.settingRepo.GetMultiple(ctx, []string{
		settingKeyUserRiskIPIntelEnabled,
		settingKeyUserRiskIPIntelProvider,
		settingKeyUserRiskIPIntelToken,
		settingKeyUserRiskReviewThreshold,
		settingKeyUserRiskThrottleThreshold,
		settingKeyUserRiskFreezeThreshold,
		settingKeyUserRiskAutoEnabled,
		settingKeyUserRiskAutoThrottle,
		settingKeyUserRiskAutoFreeze,
		settingKeyUserRiskAutoThrottleCap,
	})
	if err != nil {
		return nil, fmt.Errorf("get user risk settings: %w", err)
	}

	if raw := strings.TrimSpace(values[settingKeyUserRiskIPIntelEnabled]); raw != "" {
		settings.IPIntelEnabled = raw == "true"
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskIPIntelProvider]); raw != "" {
		settings.IPIntelProvider = raw
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskReviewThreshold]); raw != "" {
		if v, err := userRiskParseInt(raw); err == nil {
			settings.ReviewThreshold = v
		}
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskThrottleThreshold]); raw != "" {
		if v, err := userRiskParseInt(raw); err == nil {
			settings.ThrottleThreshold = v
		}
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskFreezeThreshold]); raw != "" {
		if v, err := userRiskParseInt(raw); err == nil {
			settings.FreezeThreshold = v
		}
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskAutoEnabled]); raw != "" {
		settings.AutoEnabled = raw == "true"
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskAutoThrottle]); raw != "" {
		settings.AutoThrottle = raw == "true"
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskAutoFreeze]); raw != "" {
		settings.AutoFreeze = raw == "true"
	}
	if raw := strings.TrimSpace(values[settingKeyUserRiskAutoThrottleCap]); raw != "" {
		if v, err := userRiskParseInt(raw); err == nil {
			settings.AutoThrottleCap = v
		}
	}
	internal.IPIntelToken = strings.TrimSpace(values[settingKeyUserRiskIPIntelToken])
	settings.IPIntelTokenConfigured = internal.IPIntelToken != ""
	settings.IPIntelDocsURL = userRiskIPInfoLiteDocsURL
	if err := validateUserRiskSettings(settings); err != nil {
		return nil, err
	}
	internal.Settings = settings
	return internal, nil
}

func validateUserRiskSettings(settings UserRiskSettings) error {
	provider := strings.TrimSpace(settings.IPIntelProvider)
	if provider == "" {
		provider = userRiskIPInfoLiteProvider
	}
	if provider != userRiskIPInfoLiteProvider {
		return infraerrors.BadRequest("USER_RISK_PROVIDER_INVALID", "unsupported IP intelligence provider")
	}
	if settings.ReviewThreshold < 1 || settings.ThrottleThreshold < 1 || settings.FreezeThreshold < 1 {
		return infraerrors.BadRequest("USER_RISK_THRESHOLD_INVALID", "risk thresholds must be positive integers")
	}
	if settings.ReviewThreshold >= settings.ThrottleThreshold || settings.ThrottleThreshold >= settings.FreezeThreshold {
		return infraerrors.BadRequest("USER_RISK_THRESHOLD_INVALID", "review/throttle/freeze thresholds must be strictly increasing")
	}
	if settings.AutoThrottleCap < 1 {
		return infraerrors.BadRequest("USER_RISK_AUTOMATION_INVALID", "auto throttle concurrency cap must be at least 1")
	}
	return nil
}

func (s *UserRiskService) loadIPIntelForRows(ctx context.Context, rows []userRiskUsageRow, settings *userRiskSettingsInternal) (map[string]userRiskIPIntelRecord, error) {
	if settings == nil || !settings.Settings.IPIntelEnabled || strings.TrimSpace(settings.IPIntelToken) == "" {
		return map[string]userRiskIPIntelRecord{}, nil
	}
	uniqueIPs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, row := range rows {
		ipAddr := strings.TrimSpace(row.IPAddress)
		if ipAddr == "" {
			continue
		}
		if net.ParseIP(ipAddr) == nil {
			continue
		}
		if _, ok := seen[ipAddr]; ok {
			continue
		}
		seen[ipAddr] = struct{}{}
		uniqueIPs = append(uniqueIPs, ipAddr)
	}
	if len(uniqueIPs) == 0 {
		return map[string]userRiskIPIntelRecord{}, nil
	}
	switch settings.Settings.IPIntelProvider {
	case userRiskIPInfoLiteProvider:
		return s.loadIPInfoLiteBatch(ctx, uniqueIPs, settings.IPIntelToken)
	default:
		return map[string]userRiskIPIntelRecord{}, nil
	}
}

func (s *UserRiskService) loadIPInfoLiteBatch(ctx context.Context, ips []string, token string) (map[string]userRiskIPIntelRecord, error) {
	result := make(map[string]userRiskIPIntelRecord, len(ips))
	if len(ips) == 0 {
		return result, nil
	}
	for start := 0; start < len(ips); start += 1000 {
		end := start + 1000
		if end > len(ips) {
			end = len(ips)
		}
		chunk := ips[start:end]
		body, err := json.Marshal(chunk)
		if err != nil {
			return nil, fmt.Errorf("marshal ip intel request: %w", err)
		}
		endpoint := "https://api.ipinfo.io/batch/lite?token=" + url.QueryEscape(token)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build ip intel request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request ip intel: %w", err)
		}
		func() {
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
				err = fmt.Errorf("ip intel request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(payload)))
				return
			}
			decoded := make(map[string]userRiskIPIntelRecord)
			if decodeErr := json.NewDecoder(resp.Body).Decode(&decoded); decodeErr != nil {
				err = fmt.Errorf("decode ip intel response: %w", decodeErr)
				return
			}
			for ipAddr, record := range decoded {
				if record.IP == "" {
					record.IP = ipAddr
				}
				result[ipAddr] = record
			}
		}()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func buildUserRiskDetail(
	user *User,
	recentRows []userRiskUsageRow,
	historySummary userRiskHistoricalSummary,
	keyCount int,
	windowStart time.Time,
	windowEnd time.Time,
	location *time.Location,
	timezone string,
	settings UserRiskSettings,
	ipIntel map[string]userRiskIPIntelRecord,
) *UserRiskDetail {
	requestCount24h := int64(len(recentRows))
	actualCost24h := 0.0
	ipCounts := make(map[string]int64)
	uaCounts := make(map[string]int64)
	hourCounts := make(map[int]int64)
	minuteBuckets := make(map[time.Time]*userRiskMinuteBucket)
	timestamps := make([]time.Time, 0, len(recentRows))

	for _, row := range recentRows {
		actualCost24h += row.ActualCost
		timestamps = append(timestamps, row.CreatedAt)

		if row.IPAddress != "" {
			ipCounts[row.IPAddress]++
		}
		if row.UserAgent != "" {
			uaCounts[row.UserAgent]++
		}

		localTime := row.CreatedAt.In(location)
		hourCounts[localTime.Hour()]++

		bucketAt := row.CreatedAt.UTC().Truncate(time.Minute)
		bucket := minuteBuckets[bucketAt]
		if bucket == nil {
			bucket = &userRiskMinuteBucket{IPs: make(map[string]struct{}), UAs: make(map[string]struct{})}
			minuteBuckets[bucketAt] = bucket
		}
		if row.IPAddress != "" {
			bucket.IPs[row.IPAddress] = struct{}{}
		}
		if row.UserAgent != "" {
			bucket.UAs[row.UserAgent] = struct{}{}
		}
	}

	activeHours := make([]int, 0, len(hourCounts))
	for hour := range hourCounts {
		activeHours = append(activeHours, hour)
	}
	sort.Ints(activeHours)

	concurrentMultiIPUAMinutes := 0
	for _, bucket := range minuteBuckets {
		if len(bucket.IPs) >= 2 && len(bucket.UAs) >= 2 {
			concurrentMultiIPUAMinutes++
		}
	}

	hourConcentration := calculateHourConcentration(hourCounts, requestCount24h)
	longestSilenceHours := calculateLongestSilenceHours(windowStart, windowEnd, timestamps)
	uaDetails, uaSummary := buildUserRiskUADetails(uaCounts)
	ipDetails := buildUserRiskIPDetails(ipCounts, ipIntel)

	metrics := UserRiskMetrics{
		RequestCount24h:               requestCount24h,
		ActualCost24h:                 roundFloat(actualCost24h, 4),
		FirstIP:                       historySummary.FirstIP,
		HistoricalIPCount:             historySummary.HistoricalIPCount,
		UA24hCount:                    len(uaCounts),
		ActiveHoursCount:              len(activeHours),
		ActiveHours:                   activeHours,
		LongestSilenceHours:           roundFloat(longestSilenceHours, 2),
		AllDayActive:                  len(activeHours) == 24,
		HourConcentration:             roundFloat(hourConcentration, 4),
		KeyCount:                      keyCount,
		ConcurrentMultiIPUAMinutes24h: concurrentMultiIPUAMinutes,
	}

	ruleHits := buildUserRiskRuleHits(metrics, uaSummary)
	riskScore := 0
	for _, hit := range ruleHits {
		riskScore += hit.Score
	}
	decision, decisionLabel := scoreUserRiskDecision(riskScore, settings)

	return &UserRiskDetail{
		User: UserRiskUser{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Balance:  user.Balance,
			Status:   user.Status,
		},
		Window: UserRiskWindow{
			StartAt:  windowStart,
			EndAt:    windowEnd,
			Timezone: timezone,
		},
		Summary: UserRiskSummary{
			RiskScore:     riskScore,
			Decision:      decision,
			DecisionLabel: decisionLabel,
			ComputedAt:    time.Now().UTC(),
		},
		Metrics:   metrics,
		RuleHits:  ruleHits,
		IPDetails: ipDetails,
		UADetails: uaDetails,
	}
}

func buildUserRiskRuleHits(metrics UserRiskMetrics, uaSummary userRiskUASummary) []UserRiskRuleHit {
	ruleHits := make([]UserRiskRuleHit, 0, 10)
	appendRule := func(code, label, description string, score int) {
		ruleHits = append(ruleHits, UserRiskRuleHit{Code: code, Label: label, Description: description, Score: score})
	}

	switch {
	case metrics.ActualCost24h > 200:
		appendRule("R_COST_200", "24 小时消耗异常高", "近 24 小时实际消耗超过 200 美元。", 35)
	case metrics.ActualCost24h > 100:
		appendRule("R_COST_100", "24 小时消耗偏高", "近 24 小时实际消耗超过 100 美元。", 20)
	case metrics.ActualCost24h > 50:
		appendRule("R_COST_50", "24 小时消耗抬升", "近 24 小时实际消耗超过 50 美元。", 10)
	case metrics.ActualCost24h > 20:
		appendRule("R_COST_20", "24 小时消耗明显", "近 24 小时实际消耗超过 20 美元。", 5)
	}

	switch {
	case metrics.HistoricalIPCount > 20:
		appendRule("R_IP_HISTORY_20", "历史 IP 很多", "历史去重 IP 数超过 20。", 25)
	case metrics.HistoricalIPCount > 10:
		appendRule("R_IP_HISTORY_10", "历史 IP 偏多", "历史去重 IP 数超过 10。", 15)
	case metrics.HistoricalIPCount > 5:
		appendRule("R_IP_HISTORY_5", "历史 IP 增多", "历史去重 IP 数超过 5。", 8)
	}

	switch {
	case metrics.UA24hCount > 4:
		appendRule("R_UA_24H_4", "24 小时 UA 很多", "近 24 小时去重 UA 数超过 4。", 20)
	case metrics.UA24hCount > 2:
		appendRule("R_UA_24H_2", "24 小时多个 UA", "近 24 小时去重 UA 数超过 2。", 10)
	}

	switch {
	case metrics.ActiveHoursCount >= 20:
		appendRule("R_ACTIVE_HOURS_20", "接近全天活跃", "近 24 小时覆盖至少 20 个本地小时桶。", 12)
	case metrics.ActiveHoursCount >= 16:
		appendRule("R_ACTIVE_HOURS_16", "活跃时段很宽", "近 24 小时覆盖至少 16 个本地小时桶。", 8)
	}

	switch {
	case metrics.LongestSilenceHours < 1:
		appendRule("R_SILENCE_LT_1", "几乎无静默", "最长静默时长小于 1 小时。", 20)
	case metrics.LongestSilenceHours < 2:
		appendRule("R_SILENCE_LT_2", "静默很短", "最长静默时长小于 2 小时。", 15)
	case metrics.LongestSilenceHours < 4:
		appendRule("R_SILENCE_LT_4", "静默偏短", "最长静默时长小于 4 小时。", 10)
	}

	if metrics.AllDayActive {
		appendRule("R_ALL_DAY", "全天活跃", "近 24 小时所有本地小时桶都有请求。", 10)
	}

	switch {
	case metrics.ConcurrentMultiIPUAMinutes24h >= 3:
		appendRule("R_CONCURRENT_MINUTES_3", "反复出现多 IP 多 UA 并发", "至少 3 个分钟桶同时出现多个 IP 和多个 UA。", 15)
	case metrics.ConcurrentMultiIPUAMinutes24h >= 1:
		appendRule("R_CONCURRENT_MINUTES_1", "出现多 IP 多 UA 并发", "至少 1 个分钟桶同时出现多个 IP 和多个 UA。", 8)
	}

	switch {
	case metrics.KeyCount > 3:
		appendRule("R_KEYS_3", "API Key 偏多", "该用户拥有超过 3 个有效 API Key。", 8)
	case metrics.KeyCount > 1:
		appendRule("R_KEYS_1", "存在多个 API Key", "该用户拥有超过 1 个有效 API Key。", 3)
	}

	if uaSummary.TopDetail.Programmatic && !uaSummary.TopDetail.NormalAllowed && uaSummary.TopShare >= 0.5 && uaSummary.TopDetail.BaseScore >= 8 {
		appendRule("R_DOMINANT_PROGRAMMATIC_UA", "主力 UA 为程序客户端", fmt.Sprintf("主力 UA 分类为 %s，且占近 24 小时流量超过一半。", uaSummary.TopDetail.Category), 10)
	}

	if uaSummary.AbnormalCount > 0 {
		description := "发现命中的异常 UA 配置。"
		if len(uaSummary.AbnormalUserAgent) > 0 {
			description = fmt.Sprintf("发现命中的异常 UA：%s。", strings.Join(uaSummary.AbnormalUserAgent, "、"))
		}
		appendRule("R_ABNORMAL_UA", "命中异常 UA", description, 15)
	}

	return ruleHits
}

func buildUserRiskIPDetails(ipCounts map[string]int64, intel map[string]userRiskIPIntelRecord) []UserRiskIPDetail {
	if len(ipCounts) == 0 {
		return []UserRiskIPDetail{}
	}
	items := make([]UserRiskIPDetail, 0, len(ipCounts))
	for ipAddr, count := range ipCounts {
		record := intel[ipAddr]
		label := "未知"
		if record.ASName != "" {
			label = record.ASName
		} else if record.CountryCode != "" {
			label = record.CountryCode
		}
		items = append(items, UserRiskIPDetail{
			IPAddress:    ipAddr,
			Requests:     count,
			IPType:       classifyIPType(ipAddr),
			Label:        label,
			ASN:          record.ASN,
			Organization: record.ASName,
			Domain:       record.ASDomain,
			CountryCode:  record.CountryCode,
			Country:      record.Country,
			Continent:    record.Continent,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Requests != items[j].Requests {
			return items[i].Requests > items[j].Requests
		}
		return items[i].IPAddress < items[j].IPAddress
	})
	return items
}

func buildUserRiskUADetails(uaCounts map[string]int64) ([]UserRiskUADetail, userRiskUASummary) {
	if len(uaCounts) == 0 {
		return []UserRiskUADetail{}, userRiskUASummary{}
	}

	totalRequests := int64(0)
	for _, count := range uaCounts {
		totalRequests += count
	}

	items := make([]UserRiskUADetail, 0, len(uaCounts))
	summary := userRiskUASummary{}
	for ua, count := range uaCounts {
		profile := classifyRiskUserAgent(ua)
		item := UserRiskUADetail{
			UserAgent:     ua,
			Requests:      count,
			Category:      profile.Category,
			Description:   profile.Description,
			BaseScore:     profile.BaseScore,
			ConfigStatus:  profile.ConfigStatus,
			HitRule:       profile.HitRule,
			Programmatic:  profile.Programmatic,
			NormalAllowed: profile.NormalAllowed,
		}
		items = append(items, item)

		share := 0.0
		if totalRequests > 0 {
			share = float64(count) / float64(totalRequests)
		}
		if count > summary.TopDetail.Requests || (count == summary.TopDetail.Requests && ua < summary.TopDetail.UserAgent) {
			summary.TopDetail = item
			summary.TopShare = share
		}
		if profile.ConfigStatus == "abnormal" {
			summary.AbnormalCount++
			summary.AbnormalUserAgent = append(summary.AbnormalUserAgent, ua)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Requests != items[j].Requests {
			return items[i].Requests > items[j].Requests
		}
		return items[i].UserAgent < items[j].UserAgent
	})
	sort.Strings(summary.AbnormalUserAgent)

	return items, summary
}

func classifyRiskUserAgent(userAgent string) userAgentProfile {
	normalized := strings.ToLower(strings.TrimSpace(userAgent))
	if normalized == "" {
		return userAgentProfile{
			Category:     "未知",
			Description:  "缺少 User-Agent。",
			BaseScore:    0,
			ConfigStatus: "unconfigured",
		}
	}

	for _, prefix := range userRiskNormalUAPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return userAgentProfile{
				Category:      classifyKnownClientCategory(normalized),
				Description:   fmt.Sprintf("命中正常 UA 前缀：%s", prefix),
				BaseScore:     0,
				ConfigStatus:  "normal",
				HitRule:       fmt.Sprintf("前缀：%s", prefix),
				Programmatic:  strings.Contains(normalized, "codex_sdk") || strings.Contains(normalized, "codex_exec"),
				NormalAllowed: true,
			}
		}
	}

	if abnormalScore, ok := userRiskAbnormalUAExactScores[normalized]; ok {
		return userAgentProfile{
			Category:     classifyKnownClientCategory(normalized),
			Description:  fmt.Sprintf("已标记为异常 UA（+%d 分）。", abnormalScore),
			BaseScore:    abnormalScore,
			ConfigStatus: "abnormal",
			HitRule:      fmt.Sprintf("精确：%s", userAgent),
			Programmatic: true,
		}
	}

	switch {
	case strings.HasPrefix(normalized, "go-http-client/"):
		return userAgentProfile{Category: "Go HTTP客户端", Description: "可能是中转程序或自动化请求。", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "curl/"):
		return userAgentProfile{Category: "curl", Description: "命令行 HTTP 客户端。", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.Contains(normalized, "python-requests/"):
		return userAgentProfile{Category: "Python Requests", Description: "Python 自动化脚本请求。", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.Contains(normalized, "okhttp/"):
		return userAgentProfile{Category: "OkHttp", Description: "移动端或 JVM SDK 请求。", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "openai/") || strings.Contains(normalized, "openai/js") || strings.Contains(normalized, "openai sdk"):
		return userAgentProfile{Category: "OpenAI SDK", Description: "SDK / 程序集成请求。", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "postmanruntime/"):
		return userAgentProfile{Category: "Postman", Description: "API 调试工具请求。", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "codex-tui/") || strings.HasPrefix(normalized, "codex_tui/"):
		return userAgentProfile{Category: "Codex TUI", Description: "终端交互客户端。", BaseScore: 4, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "codex desktop/"):
		return userAgentProfile{Category: "Codex Desktop", Description: "桌面客户端请求。", BaseScore: 0, ConfigStatus: "normal", HitRule: "前缀：codex desktop/", NormalAllowed: true}
	case strings.Contains(normalized, "mozilla/") || strings.Contains(normalized, "chrome/") || strings.Contains(normalized, "safari/") || strings.Contains(normalized, "firefox/") || strings.Contains(normalized, "electron/"):
		return userAgentProfile{Category: "浏览器", Description: "浏览器样式请求。", BaseScore: 0, ConfigStatus: "unconfigured", Programmatic: false}
	default:
		return userAgentProfile{Category: "其他", Description: "未知客户端指纹。", BaseScore: 4, ConfigStatus: "unconfigured", Programmatic: false}
	}
}

func classifyKnownClientCategory(normalizedUA string) string {
	switch {
	case strings.HasPrefix(normalizedUA, "codex_cli_rs/"):
		return "Codex CLI"
	case strings.HasPrefix(normalizedUA, "codex_vscode/"):
		return "Codex VS Code"
	case strings.HasPrefix(normalizedUA, "codex desktop/"):
		return "Codex Desktop"
	case strings.HasPrefix(normalizedUA, "codex_app/") || strings.HasPrefix(normalizedUA, "codex_chatgpt_desktop/"):
		return "Codex Desktop"
	case strings.HasPrefix(normalizedUA, "codex_sdk_ts/"):
		return "Codex SDK"
	case strings.HasPrefix(normalizedUA, "go-http-client/"):
		return "Go HTTP客户端"
	default:
		return "已知客户端"
	}
}

func calculateHourConcentration(hourCounts map[int]int64, totalRequests int64) float64 {
	if totalRequests <= 0 {
		return 0
	}
	hhi := 0.0
	for _, count := range hourCounts {
		share := float64(count) / float64(totalRequests)
		hhi += share * share
	}
	return hhi
}

func calculateLongestSilenceHours(windowStart, windowEnd time.Time, timestamps []time.Time) float64 {
	if len(timestamps) == 0 {
		return windowEnd.Sub(windowStart).Hours()
	}
	longest := timestamps[0].Sub(windowStart)
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		if gap > longest {
			longest = gap
		}
	}
	endGap := windowEnd.Sub(timestamps[len(timestamps)-1])
	if endGap > longest {
		longest = endGap
	}
	if longest < 0 {
		return 0
	}
	return longest.Hours()
}

func loadUserRiskLocation(timezone string) (*time.Location, string) {
	tz := strings.TrimSpace(timezone)
	if tz == "" {
		return time.UTC, "UTC"
	}
	location, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC, "UTC"
	}
	return location, tz
}

func scoreUserRiskDecision(score int, settings UserRiskSettings) (string, string) {
	observeThreshold := settings.ReviewThreshold / 2
	if observeThreshold < 25 {
		observeThreshold = 25
	}
	switch {
	case score >= settings.FreezeThreshold:
		return "freeze_review", "建议冻结审查"
	case score >= settings.ThrottleThreshold:
		return "throttle", "建议限流观察"
	case score >= settings.ReviewThreshold:
		return "review", "建议人工审查"
	case score >= observeThreshold:
		return "observe", "建议观察"
	default:
		return "normal", "正常"
	}
}

func roundFloat(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}

func classifyIPType(ipAddress string) string {
	parsed := net.ParseIP(strings.TrimSpace(ipAddress))
	if parsed == nil {
		return "unknown"
	}
	if parsed.To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

func normalizePositiveUserRiskIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func userRiskFormatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func userRiskFormatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func userRiskParseInt(raw string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
