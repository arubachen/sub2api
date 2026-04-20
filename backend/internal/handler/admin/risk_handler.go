package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type RiskHandler struct {
	userRiskService *service.UserRiskService
}

func NewRiskHandler(userRiskService *service.UserRiskService) *RiskHandler {
	return &RiskHandler{userRiskService: userRiskService}
}

type BatchRiskSummariesRequest struct {
	UserIDs  []int64 `json:"user_ids"`
	Timezone string  `json:"timezone"`
}

type UpdateRiskSettingsRequest struct {
	IPIntelEnabled    *bool   `json:"ip_intel_enabled"`
	IPIntelProvider   *string `json:"ip_intel_provider"`
	IPIntelToken      *string `json:"ip_intel_token"`
	ClearIPIntelToken *bool   `json:"clear_ip_intel_token"`
	ReviewThreshold   *int    `json:"review_threshold"`
	ThrottleThreshold *int    `json:"throttle_threshold"`
	FreezeThreshold   *int    `json:"freeze_threshold"`
}

func (h *RiskHandler) GetSettings(c *gin.Context) {
	settings, err := h.userRiskService.GetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *RiskHandler) UpdateSettings(c *gin.Context) {
	var req UpdateRiskSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	updated, err := h.userRiskService.UpdateSettings(c.Request.Context(), &service.UserRiskSettingsUpdate{
		IPIntelEnabled:    req.IPIntelEnabled,
		IPIntelProvider:   req.IPIntelProvider,
		IPIntelToken:      req.IPIntelToken,
		ClearIPIntelToken: req.ClearIPIntelToken,
		ReviewThreshold:   req.ReviewThreshold,
		ThrottleThreshold: req.ThrottleThreshold,
		FreezeThreshold:   req.FreezeThreshold,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, updated)
}

func (h *RiskHandler) BatchSummaries(c *gin.Context) {
	var req BatchRiskSummariesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	items, err := h.userRiskService.GetUserRiskSummaries(c.Request.Context(), req.UserIDs, strings.TrimSpace(req.Timezone))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *RiskHandler) GetUserDetail(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	detail, err := h.userRiskService.GetUserRiskDetail(c.Request.Context(), userID, strings.TrimSpace(c.Query("timezone")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}
