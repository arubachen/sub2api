//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type promoRepoUnitStub struct {
	code      *PromoCode
	created   *PromoCode
	updated   *PromoCode
	createErr error
	updateErr error
}

func (s *promoRepoUnitStub) Create(_ context.Context, code *PromoCode) error {
	if s.createErr != nil {
		return s.createErr
	}
	cloned := clonePromoCode(code)
	if cloned.ID == 0 {
		cloned.ID = 1
	}
	s.created = cloned
	s.code = clonePromoCode(cloned)
	code.ID = cloned.ID
	code.AllowedEmailSuffixes = append([]string{}, cloned.AllowedEmailSuffixes...)
	return nil
}

func (s *promoRepoUnitStub) GetByID(_ context.Context, id int64) (*PromoCode, error) {
	if s.code == nil || s.code.ID != id {
		return nil, ErrPromoCodeNotFound
	}
	return clonePromoCode(s.code), nil
}

func (s *promoRepoUnitStub) GetByCode(_ context.Context, code string) (*PromoCode, error) {
	if s.code == nil || !strings.EqualFold(s.code.Code, code) {
		return nil, ErrPromoCodeNotFound
	}
	return clonePromoCode(s.code), nil
}

func (s *promoRepoUnitStub) GetByCodeForUpdate(ctx context.Context, code string) (*PromoCode, error) {
	return s.GetByCode(ctx, code)
}

func (s *promoRepoUnitStub) Update(_ context.Context, code *PromoCode) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updated = clonePromoCode(code)
	s.code = clonePromoCode(code)
	return nil
}

func (s *promoRepoUnitStub) Delete(context.Context, int64) error {
	return nil
}

func (s *promoRepoUnitStub) List(context.Context, pagination.PaginationParams) ([]PromoCode, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *promoRepoUnitStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]PromoCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *promoRepoUnitStub) CreateUsage(context.Context, *PromoCodeUsage) error {
	panic("unexpected CreateUsage call")
}

func (s *promoRepoUnitStub) GetUsageByPromoCodeAndUser(context.Context, int64, int64) (*PromoCodeUsage, error) {
	panic("unexpected GetUsageByPromoCodeAndUser call")
}

func (s *promoRepoUnitStub) ListUsagesByPromoCode(context.Context, int64, pagination.PaginationParams) ([]PromoCodeUsage, *pagination.PaginationResult, error) {
	panic("unexpected ListUsagesByPromoCode call")
}

func (s *promoRepoUnitStub) IncrementUsedCount(context.Context, int64) error {
	panic("unexpected IncrementUsedCount call")
}

func clonePromoCode(code *PromoCode) *PromoCode {
	if code == nil {
		return nil
	}
	cloned := *code
	cloned.AllowedEmailSuffixes = append([]string{}, code.AllowedEmailSuffixes...)
	return &cloned
}

func TestPromoService_ValidatePromoCodeForEmail_AllowsMatchingSuffix(t *testing.T) {
	repo := &promoRepoUnitStub{
		code: &PromoCode{
			ID:                   1,
			Code:                 "TEAM30",
			AllowedEmailSuffixes: []string{"@company.com"},
			BonusAmount:          100,
			Status:               PromoCodeStatusActive,
		},
	}
	service := NewPromoService(repo, nil, nil, nil, nil)

	got, err := service.ValidatePromoCodeForEmail(context.Background(), "TEAM30", "member@company.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "@company.com", got.AllowedEmailSuffixes[0])
}

func TestPromoService_ValidatePromoCodeForEmail_RejectsNonMatchingSuffix(t *testing.T) {
	repo := &promoRepoUnitStub{
		code: &PromoCode{
			ID:                   1,
			Code:                 "TEAM30",
			AllowedEmailSuffixes: []string{"@company.com"},
			BonusAmount:          100,
			Status:               PromoCodeStatusActive,
		},
	}
	service := NewPromoService(repo, nil, nil, nil, nil)

	_, err := service.ValidatePromoCodeForEmail(context.Background(), "TEAM30", "member@gmail.com")
	require.ErrorIs(t, err, ErrPromoCodeEmailSuffixNotAllowed)
}

func TestPromoService_Create_NormalizesAllowedEmailSuffixes(t *testing.T) {
	repo := &promoRepoUnitStub{}
	service := NewPromoService(repo, nil, nil, nil, nil)

	code, err := service.Create(context.Background(), &CreatePromoCodeInput{
		Code:                 "team30",
		AllowedEmailSuffixes: []string{"Company.com", "@Example.org", "company.com"},
		BonusAmount:          100,
		MaxUses:              30,
	})
	require.NoError(t, err)
	require.NotNil(t, code)
	require.Equal(t, []string{"@company.com", "@example.org"}, code.AllowedEmailSuffixes)
	require.Equal(t, []string{"@company.com", "@example.org"}, repo.created.AllowedEmailSuffixes)
}

func TestPromoService_Update_ClearsAllowedEmailSuffixes(t *testing.T) {
	repo := &promoRepoUnitStub{
		code: &PromoCode{
			ID:                   1,
			Code:                 "TEAM30",
			AllowedEmailSuffixes: []string{"@company.com"},
			BonusAmount:          100,
			Status:               PromoCodeStatusActive,
		},
	}
	service := NewPromoService(repo, nil, nil, nil, nil)
	empty := []string{}

	updated, err := service.Update(context.Background(), 1, &UpdatePromoCodeInput{
		AllowedEmailSuffixes: &empty,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Empty(t, updated.AllowedEmailSuffixes)
	require.Empty(t, repo.updated.AllowedEmailSuffixes)
}
