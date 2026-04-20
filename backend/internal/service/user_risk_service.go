package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const userRiskWindow = 24 * time.Hour

var (
	userRiskErrDBUnavailable = infraerrors.InternalServer("USER_RISK_DB_UNAVAILABLE", "user risk database unavailable")
)

// UserRiskService computes rolling user risk details for admin review pages.
type UserRiskService struct {
	db       *sql.DB
	userRepo UserRepository
}

func NewUserRiskService(db *sql.DB, userRepo UserRepository) *UserRiskService {
	return &UserRiskService{db: db, userRepo: userRepo}
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
	Organization string `json:"organization,omitempty"`
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

func (s *UserRiskService) GetUserRiskDetail(ctx context.Context, userID int64, timezone string) (*UserRiskDetail, error) {
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

	return buildUserRiskDetail(user, recentRows, historySummary, keyCount, windowStart, windowEnd, location, tzName), nil
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
	defer rows.Close()

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

func buildUserRiskDetail(
	user *User,
	recentRows []userRiskUsageRow,
	historySummary userRiskHistoricalSummary,
	keyCount int,
	windowStart time.Time,
	windowEnd time.Time,
	location *time.Location,
	timezone string,
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
	ipDetails := buildUserRiskIPDetails(ipCounts)

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
	decision, decisionLabel := scoreUserRiskDecision(riskScore)

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
		appendRule("R_COST_200", "24h spend spike", "24h actual spend exceeds $200.", 35)
	case metrics.ActualCost24h > 100:
		appendRule("R_COST_100", "24h spend high", "24h actual spend exceeds $100.", 20)
	case metrics.ActualCost24h > 50:
		appendRule("R_COST_50", "24h spend elevated", "24h actual spend exceeds $50.", 10)
	case metrics.ActualCost24h > 20:
		appendRule("R_COST_20", "24h spend noticeable", "24h actual spend exceeds $20.", 5)
	}

	switch {
	case metrics.HistoricalIPCount > 20:
		appendRule("R_IP_HISTORY_20", "Many historical IPs", "Historical distinct IPs exceed 20.", 25)
	case metrics.HistoricalIPCount > 10:
		appendRule("R_IP_HISTORY_10", "Historical IPs high", "Historical distinct IPs exceed 10.", 15)
	case metrics.HistoricalIPCount > 5:
		appendRule("R_IP_HISTORY_5", "Historical IPs elevated", "Historical distinct IPs exceed 5.", 8)
	}

	switch {
	case metrics.UA24hCount > 4:
		appendRule("R_UA_24H_4", "Many user agents in 24h", "Distinct 24h user agents exceed 4.", 20)
	case metrics.UA24hCount > 2:
		appendRule("R_UA_24H_2", "Multiple user agents in 24h", "Distinct 24h user agents exceed 2.", 10)
	}

	switch {
	case metrics.ActiveHoursCount >= 20:
		appendRule("R_ACTIVE_HOURS_20", "Almost all-day activity", "Requests span at least 20 distinct local hours in 24h.", 12)
	case metrics.ActiveHoursCount >= 16:
		appendRule("R_ACTIVE_HOURS_16", "Wide activity window", "Requests span at least 16 distinct local hours in 24h.", 8)
	}

	switch {
	case metrics.LongestSilenceHours < 1:
		appendRule("R_SILENCE_LT_1", "No meaningful quiet period", "Longest quiet period is under 1 hour.", 20)
	case metrics.LongestSilenceHours < 2:
		appendRule("R_SILENCE_LT_2", "Quiet period very short", "Longest quiet period is under 2 hours.", 15)
	case metrics.LongestSilenceHours < 4:
		appendRule("R_SILENCE_LT_4", "Quiet period short", "Longest quiet period is under 4 hours.", 10)
	}

	if metrics.AllDayActive {
		appendRule("R_ALL_DAY", "Active across all 24 hours", "Requests appear in every local hour bucket in the last 24h.", 10)
	}

	switch {
	case metrics.ConcurrentMultiIPUAMinutes24h >= 3:
		appendRule("R_CONCURRENT_MINUTES_3", "Repeated multi-IP multi-UA bursts", "At least 3 minute buckets contain both multiple IPs and multiple user agents.", 15)
	case metrics.ConcurrentMultiIPUAMinutes24h >= 1:
		appendRule("R_CONCURRENT_MINUTES_1", "Multi-IP multi-UA burst", "At least 1 minute bucket contains both multiple IPs and multiple user agents.", 8)
	}

	switch {
	case metrics.KeyCount > 3:
		appendRule("R_KEYS_3", "Many API keys", "User owns more than 3 active API keys.", 8)
	case metrics.KeyCount > 1:
		appendRule("R_KEYS_1", "Multiple API keys", "User owns more than 1 active API key.", 3)
	}

	if uaSummary.TopDetail.Programmatic && !uaSummary.TopDetail.NormalAllowed && uaSummary.TopShare >= 0.5 && uaSummary.TopDetail.BaseScore >= 8 {
		appendRule("R_DOMINANT_PROGRAMMATIC_UA", "Dominant programmatic UA", fmt.Sprintf("Top user agent is %s and dominates 24h traffic.", uaSummary.TopDetail.Category), 10)
	}

	if uaSummary.AbnormalCount > 0 {
		description := "One or more configured abnormal user agents were observed."
		if len(uaSummary.AbnormalUserAgent) > 0 {
			description = fmt.Sprintf("Configured abnormal user agents observed: %s.", strings.Join(uaSummary.AbnormalUserAgent, ", "))
		}
		appendRule("R_ABNORMAL_UA", "Configured abnormal UA", description, 15)
	}

	return ruleHits
}

func buildUserRiskIPDetails(ipCounts map[string]int64) []UserRiskIPDetail {
	if len(ipCounts) == 0 {
		return []UserRiskIPDetail{}
	}
	items := make([]UserRiskIPDetail, 0, len(ipCounts))
	for ipAddr, count := range ipCounts {
		items = append(items, UserRiskIPDetail{
			IPAddress:    ipAddr,
			Requests:     count,
			IPType:       classifyIPType(ipAddr),
			Label:        "Unknown",
			Organization: "",
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
			Category:     "Unknown",
			Description:  "Missing user agent.",
			BaseScore:    0,
			ConfigStatus: "unconfigured",
		}
	}

	for _, prefix := range userRiskNormalUAPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return userAgentProfile{
				Category:      classifyKnownClientCategory(normalized),
				Description:   fmt.Sprintf("Matches normal UA prefix: %s", prefix),
				BaseScore:     0,
				ConfigStatus:  "normal",
				HitRule:       fmt.Sprintf("prefix: %s", prefix),
				Programmatic:  strings.Contains(normalized, "codex_sdk") || strings.Contains(normalized, "codex_exec"),
				NormalAllowed: true,
			}
		}
	}

	if abnormalScore, ok := userRiskAbnormalUAExactScores[normalized]; ok {
		return userAgentProfile{
			Category:     classifyKnownClientCategory(normalized),
			Description:  fmt.Sprintf("Configured abnormal user agent (+%d).", abnormalScore),
			BaseScore:    abnormalScore,
			ConfigStatus: "abnormal",
			HitRule:      fmt.Sprintf("exact: %s", userAgent),
			Programmatic: true,
		}
	}

	switch {
	case strings.HasPrefix(normalized, "go-http-client/"):
		return userAgentProfile{Category: "Go HTTP Client", Description: "Likely automated relay or direct programmatic traffic.", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "curl/"):
		return userAgentProfile{Category: "curl", Description: "Command-line HTTP client.", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.Contains(normalized, "python-requests/"):
		return userAgentProfile{Category: "Python Requests", Description: "Python automation / script traffic.", BaseScore: 10, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.Contains(normalized, "okhttp/"):
		return userAgentProfile{Category: "OkHttp", Description: "Mobile or JVM SDK traffic.", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "openai/") || strings.Contains(normalized, "openai/js") || strings.Contains(normalized, "openai sdk"):
		return userAgentProfile{Category: "OpenAI SDK", Description: "SDK / automation integration.", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "postmanruntime/"):
		return userAgentProfile{Category: "Postman", Description: "API tooling traffic.", BaseScore: 8, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "codex-tui/") || strings.HasPrefix(normalized, "codex_tui/"):
		return userAgentProfile{Category: "Codex TUI", Description: "Terminal client traffic.", BaseScore: 4, ConfigStatus: "unconfigured", Programmatic: true}
	case strings.HasPrefix(normalized, "codex desktop/"):
		return userAgentProfile{Category: "Codex Desktop", Description: "Desktop client traffic.", BaseScore: 0, ConfigStatus: "normal", HitRule: "prefix: codex desktop/", NormalAllowed: true}
	case strings.Contains(normalized, "mozilla/") || strings.Contains(normalized, "chrome/") || strings.Contains(normalized, "safari/") || strings.Contains(normalized, "firefox/") || strings.Contains(normalized, "electron/"):
		return userAgentProfile{Category: "Browser", Description: "Browser-like traffic.", BaseScore: 0, ConfigStatus: "unconfigured", Programmatic: false}
	default:
		return userAgentProfile{Category: "Other", Description: "Unknown client fingerprint.", BaseScore: 4, ConfigStatus: "unconfigured", Programmatic: false}
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
		return "Go HTTP Client"
	default:
		return "Known Client"
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

func scoreUserRiskDecision(score int) (string, string) {
	switch {
	case score >= 120:
		return "freeze_review", "Recommend freeze & review"
	case score >= 80:
		return "throttle", "Recommend limit / throttle"
	case score >= 50:
		return "review", "Recommend manual review"
	case score >= 25:
		return "observe", "Observe"
	default:
		return "normal", "Normal"
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
