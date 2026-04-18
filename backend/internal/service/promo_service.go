package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrPromoCodeNotFound              = infraerrors.NotFound("PROMO_CODE_NOT_FOUND", "promo code not found")
	ErrPromoCodeExpired               = infraerrors.BadRequest("PROMO_CODE_EXPIRED", "promo code has expired")
	ErrPromoCodeDisabled              = infraerrors.BadRequest("PROMO_CODE_DISABLED", "promo code is disabled")
	ErrPromoCodeMaxUsed               = infraerrors.BadRequest("PROMO_CODE_MAX_USED", "promo code has reached maximum uses")
	ErrPromoCodeAlreadyUsed           = infraerrors.Conflict("PROMO_CODE_ALREADY_USED", "you have already used this promo code")
	ErrPromoCodeEmailSuffixNotAllowed = infraerrors.BadRequest("PROMO_CODE_EMAIL_SUFFIX_NOT_ALLOWED", "promo code is not allowed for this email domain")
	ErrPromoCodeInvalid               = infraerrors.BadRequest("PROMO_CODE_INVALID", "invalid promo code")
)

// PromoService 优惠码服务
type PromoService struct {
	promoRepo            PromoCodeRepository
	userRepo             UserRepository
	billingCacheService  *BillingCacheService
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
}

// NewPromoService 创建优惠码服务实例
func NewPromoService(
	promoRepo PromoCodeRepository,
	userRepo UserRepository,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
) *PromoService {
	return &PromoService{
		promoRepo:            promoRepo,
		userRepo:             userRepo,
		billingCacheService:  billingCacheService,
		entClient:            entClient,
		authCacheInvalidator: authCacheInvalidator,
	}
}

// ValidatePromoCode 验证优惠码（注册前调用）
// 返回 nil, nil 表示空码（不报错）
func (s *PromoService) ValidatePromoCode(ctx context.Context, code string) (*PromoCode, error) {
	return s.ValidatePromoCodeForEmail(ctx, code, "")
}

// ValidatePromoCodeForEmail 验证优惠码（注册前调用，可选校验邮箱域名）
// 返回 nil, nil 表示空码（不报错）
func (s *PromoService) ValidatePromoCodeForEmail(ctx context.Context, code, email string) (*PromoCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil // 空码不报错，直接返回
	}

	promoCode, err := s.promoRepo.GetByCode(ctx, code)
	if err != nil {
		// 保留原始错误类型，不要统一映射为 NotFound
		return nil, err
	}

	if err := s.validatePromoCodeStatus(promoCode); err != nil {
		return nil, err
	}
	if err := s.validatePromoCodeEmailPolicy(promoCode, email); err != nil {
		return nil, err
	}

	return promoCode, nil
}

// validatePromoCodeStatus 验证优惠码状态
func (s *PromoService) validatePromoCodeStatus(promoCode *PromoCode) error {
	if !promoCode.CanUse() {
		if promoCode.IsExpired() {
			return ErrPromoCodeExpired
		}
		if promoCode.Status == PromoCodeStatusDisabled {
			return ErrPromoCodeDisabled
		}
		if promoCode.MaxUses > 0 && promoCode.UsedCount >= promoCode.MaxUses {
			return ErrPromoCodeMaxUsed
		}
		return ErrPromoCodeInvalid
	}
	return nil
}

func (s *PromoService) validatePromoCodeEmailPolicy(promoCode *PromoCode, email string) error {
	allowed := normalizePromoAllowedEmailSuffixes(promoCode.AllowedEmailSuffixes)
	if len(allowed) == 0 || strings.TrimSpace(email) == "" {
		return nil
	}
	if IsRegistrationEmailSuffixAllowed(email, allowed) {
		return nil
	}
	return buildPromoCodeEmailSuffixNotAllowedError(allowed)
}

func buildPromoCodeEmailSuffixNotAllowedError(whitelist []string) error {
	if len(whitelist) == 0 {
		return ErrPromoCodeEmailSuffixNotAllowed
	}
	allowed := strings.Join(whitelist, ", ")
	return infraerrors.BadRequest(
		"PROMO_CODE_EMAIL_SUFFIX_NOT_ALLOWED",
		fmt.Sprintf("promo code is not allowed for this email domain, allowed suffixes: %s", allowed),
	).WithMetadata(map[string]string{
		"allowed_suffixes":     strings.Join(whitelist, ","),
		"allowed_suffix_count": fmt.Sprintf("%d", len(whitelist)),
	})
}

// ApplyPromoCode 应用优惠码（注册成功后调用）
// 使用事务和行锁确保并发安全
func (s *PromoService) ApplyPromoCode(ctx context.Context, userID int64, code string) error {
	return s.ApplyPromoCodeForEmail(ctx, userID, "", code)
}

// ApplyPromoCodeForEmail 应用优惠码（注册成功后调用）
// 使用事务和行锁确保并发安全
func (s *PromoService) ApplyPromoCodeForEmail(ctx context.Context, userID int64, email, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}

	// 开启事务
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)

	// 在事务中获取并锁定优惠码记录（FOR UPDATE）
	promoCode, err := s.promoRepo.GetByCodeForUpdate(txCtx, code)
	if err != nil {
		return err
	}

	if _, err := s.applyPromoCodeTx(txCtx, userID, email, promoCode); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	s.postApplyPromoCode(ctx, userID, promoCode.BonusAmount)
	return nil
}

func (s *PromoService) applyPromoCodeTx(ctx context.Context, userID int64, email string, promoCode *PromoCode) (*PromoCodeUsage, error) {
	if promoCode == nil {
		return nil, ErrPromoCodeInvalid
	}
	if err := s.validatePromoCodeStatus(promoCode); err != nil {
		return nil, err
	}
	if err := s.validatePromoCodeEmailPolicy(promoCode, email); err != nil {
		return nil, err
	}

	existing, err := s.promoRepo.GetUsageByPromoCodeAndUser(ctx, promoCode.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("check existing usage: %w", err)
	}
	if existing != nil {
		return nil, ErrPromoCodeAlreadyUsed
	}

	if err := s.userRepo.UpdateBalance(ctx, userID, promoCode.BonusAmount); err != nil {
		return nil, fmt.Errorf("update user balance: %w", err)
	}

	usage := &PromoCodeUsage{
		PromoCodeID: promoCode.ID,
		UserID:      userID,
		BonusAmount: promoCode.BonusAmount,
		UsedAt:      time.Now(),
	}
	if err := s.promoRepo.CreateUsage(ctx, usage); err != nil {
		return nil, fmt.Errorf("create usage record: %w", err)
	}

	if err := s.promoRepo.IncrementUsedCount(ctx, promoCode.ID); err != nil {
		return nil, fmt.Errorf("increment used count: %w", err)
	}
	return usage, nil
}

func (s *PromoService) invalidatePromoCaches(ctx context.Context, userID int64, bonusAmount float64) {
	if bonusAmount == 0 || s.authCacheInvalidator == nil {
		return
	}
	s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
}

func (s *PromoService) postApplyPromoCode(ctx context.Context, userID int64, bonusAmount float64) {
	s.invalidatePromoCaches(ctx, userID, bonusAmount)
	if s.billingCacheService != nil {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
		}()
	}
}

// GenerateRandomCode 生成随机优惠码
func (s *PromoService) GenerateRandomCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// Create 创建优惠码
func (s *PromoService) Create(ctx context.Context, input *CreatePromoCodeInput) (*PromoCode, error) {
	code := strings.TrimSpace(input.Code)
	if code == "" {
		// 自动生成
		var err error
		code, err = s.GenerateRandomCode()
		if err != nil {
			return nil, err
		}
	}

	promoCode := &PromoCode{
		Code:        strings.ToUpper(code),
		BonusAmount: input.BonusAmount,
		MaxUses:     input.MaxUses,
		UsedCount:   0,
		Status:      PromoCodeStatusActive,
		ExpiresAt:   input.ExpiresAt,
		Notes:       input.Notes,
	}
	allowedEmailSuffixes, err := normalizePromoAllowedEmailSuffixesStrict(input.AllowedEmailSuffixes)
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_PROMO_ALLOWED_EMAIL_SUFFIXES", err.Error())
	}
	promoCode.AllowedEmailSuffixes = allowedEmailSuffixes

	if err := s.promoRepo.Create(ctx, promoCode); err != nil {
		return nil, fmt.Errorf("create promo code: %w", err)
	}

	return promoCode, nil
}

// GetByID 根据ID获取优惠码
func (s *PromoService) GetByID(ctx context.Context, id int64) (*PromoCode, error) {
	code, err := s.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return code, nil
}

// Update 更新优惠码
func (s *PromoService) Update(ctx context.Context, id int64, input *UpdatePromoCodeInput) (*PromoCode, error) {
	promoCode, err := s.promoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil {
		promoCode.Code = strings.ToUpper(strings.TrimSpace(*input.Code))
	}
	if input.AllowedEmailSuffixes != nil {
		allowedEmailSuffixes, err := normalizePromoAllowedEmailSuffixesStrict(*input.AllowedEmailSuffixes)
		if err != nil {
			return nil, infraerrors.BadRequest("INVALID_PROMO_ALLOWED_EMAIL_SUFFIXES", err.Error())
		}
		promoCode.AllowedEmailSuffixes = allowedEmailSuffixes
	}
	if input.BonusAmount != nil {
		promoCode.BonusAmount = *input.BonusAmount
	}
	if input.MaxUses != nil {
		promoCode.MaxUses = *input.MaxUses
	}
	if input.Status != nil {
		promoCode.Status = *input.Status
	}
	if input.ExpiresAt != nil {
		promoCode.ExpiresAt = input.ExpiresAt
	}
	if input.Notes != nil {
		promoCode.Notes = *input.Notes
	}

	if err := s.promoRepo.Update(ctx, promoCode); err != nil {
		return nil, fmt.Errorf("update promo code: %w", err)
	}

	return promoCode, nil
}

func normalizePromoAllowedEmailSuffixes(raw []string) []string {
	normalized, err := NormalizeRegistrationEmailSuffixWhitelist(raw)
	if err != nil || len(normalized) == 0 {
		return []string{}
	}
	return normalized
}

func normalizePromoAllowedEmailSuffixesStrict(raw []string) ([]string, error) {
	normalized, err := NormalizeRegistrationEmailSuffixWhitelist(raw)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return []string{}, nil
	}
	return normalized, nil
}

// Delete 删除优惠码
func (s *PromoService) Delete(ctx context.Context, id int64) error {
	if err := s.promoRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete promo code: %w", err)
	}
	return nil
}

// List 获取优惠码列表
func (s *PromoService) List(ctx context.Context, params pagination.PaginationParams, status, search, allowedEmailSuffix string) ([]PromoCode, *pagination.PaginationResult, error) {
	return s.promoRepo.ListWithFilters(ctx, params, status, search, allowedEmailSuffix)
}

// ListUsages 获取使用记录
func (s *PromoService) ListUsages(ctx context.Context, promoCodeID int64, params pagination.PaginationParams) ([]PromoCodeUsage, *pagination.PaginationResult, error) {
	return s.promoRepo.ListUsagesByPromoCode(ctx, promoCodeID, params)
}
