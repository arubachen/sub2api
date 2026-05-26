//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	dbpromocode "github.com/Wei-Shaw/sub2api/ent/promocode"
	dbpromocodeusage "github.com/Wei-Shaw/sub2api/ent/promocodeusage"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type entUserRepoForAuthPromoTest struct {
	client *dbent.Client
}

func (r *entUserRepoForAuthPromoTest) Create(ctx context.Context, user *User) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	created, err := client.User.Create().
		SetEmail(user.Email).
		SetPasswordHash(user.PasswordHash).
		SetRole(user.Role).
		SetBalance(user.Balance).
		SetConcurrency(user.Concurrency).
		SetStatus(user.Status).
		Save(ctx)
	if err != nil {
		return err
	}
	user.ID = created.ID
	return nil
}

func (r *entUserRepoForAuthPromoTest) GetByID(ctx context.Context, id int64) (*User, error) {
	model, err := r.client.User.Query().Where(dbuser.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:          model.ID,
		Email:       model.Email,
		Balance:     model.Balance,
		Concurrency: model.Concurrency,
		Role:        model.Role,
		Status:      model.Status,
	}, nil
}

func (r *entUserRepoForAuthPromoTest) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}

func (r *entUserRepoForAuthPromoTest) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (r *entUserRepoForAuthPromoTest) Update(context.Context, *User) error {
	panic("unexpected Update call")
}

func (r *entUserRepoForAuthPromoTest) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (r *entUserRepoForAuthPromoTest) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}

func (r *entUserRepoForAuthPromoTest) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}

func (r *entUserRepoForAuthPromoTest) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}

func (r *entUserRepoForAuthPromoTest) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (r *entUserRepoForAuthPromoTest) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (r *entUserRepoForAuthPromoTest) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}

func (r *entUserRepoForAuthPromoTest) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}

func (r *entUserRepoForAuthPromoTest) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}

func (r *entUserRepoForAuthPromoTest) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	_, err := client.User.UpdateOneID(id).AddBalance(amount).Save(ctx)
	return err
}

func (r *entUserRepoForAuthPromoTest) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}

func (r *entUserRepoForAuthPromoTest) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}

func (r *entUserRepoForAuthPromoTest) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchSetConcurrency call")
}

func (r *entUserRepoForAuthPromoTest) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	panic("unexpected BatchAddConcurrency call")
}

func (r *entUserRepoForAuthPromoTest) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	return client.User.Query().Where(dbuser.EmailEQ(email)).Exist(ctx)
}

func (r *entUserRepoForAuthPromoTest) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (r *entUserRepoForAuthPromoTest) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (r *entUserRepoForAuthPromoTest) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}

func (r *entUserRepoForAuthPromoTest) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}

func (r *entUserRepoForAuthPromoTest) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}

func (r *entUserRepoForAuthPromoTest) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (r *entUserRepoForAuthPromoTest) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}

func (r *entUserRepoForAuthPromoTest) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

type entPromoRepoForAuthPromoTest struct {
	client *dbent.Client
}

func (r *entPromoRepoForAuthPromoTest) Create(ctx context.Context, code *PromoCode) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	created, err := client.PromoCode.Create().
		SetCode(code.Code).
		SetAllowedEmailSuffixes(code.AllowedEmailSuffixes).
		SetBonusAmount(code.BonusAmount).
		SetMaxUses(code.MaxUses).
		SetUsedCount(code.UsedCount).
		SetStatus(code.Status).
		SetNotes(code.Notes).
		Save(ctx)
	if err != nil {
		return err
	}
	code.ID = created.ID
	return nil
}

func (r *entPromoRepoForAuthPromoTest) GetByID(ctx context.Context, id int64) (*PromoCode, error) {
	model, err := r.client.PromoCode.Query().Where(dbpromocode.IDEQ(id)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return promoCodeEntityToServiceForAuthPromoTest(model), nil
}

func (r *entPromoRepoForAuthPromoTest) GetByCode(ctx context.Context, code string) (*PromoCode, error) {
	model, err := r.client.PromoCode.Query().Where(dbpromocode.CodeEqualFold(code)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPromoCodeNotFound
		}
		return nil, err
	}
	return promoCodeEntityToServiceForAuthPromoTest(model), nil
}

func (r *entPromoRepoForAuthPromoTest) GetByCodeForUpdate(ctx context.Context, code string) (*PromoCode, error) {
	client := r.client
	q := client.PromoCode.Query().Where(dbpromocode.CodeEqualFold(code))
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
		q = client.PromoCode.Query().Where(dbpromocode.CodeEqualFold(code))
	}
	model, err := q.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrPromoCodeNotFound
		}
		return nil, err
	}
	return promoCodeEntityToServiceForAuthPromoTest(model), nil
}

func (r *entPromoRepoForAuthPromoTest) Update(context.Context, *PromoCode) error {
	panic("unexpected Update call")
}

func (r *entPromoRepoForAuthPromoTest) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (r *entPromoRepoForAuthPromoTest) List(context.Context, pagination.PaginationParams) ([]PromoCode, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (r *entPromoRepoForAuthPromoTest) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string) ([]PromoCode, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (r *entPromoRepoForAuthPromoTest) CreateUsage(ctx context.Context, usage *PromoCodeUsage) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	created, err := client.PromoCodeUsage.Create().
		SetPromoCodeID(usage.PromoCodeID).
		SetUserID(usage.UserID).
		SetBonusAmount(usage.BonusAmount).
		SetUsedAt(usage.UsedAt).
		Save(ctx)
	if err != nil {
		return err
	}
	usage.ID = created.ID
	return nil
}

func (r *entPromoRepoForAuthPromoTest) GetUsageByPromoCodeAndUser(ctx context.Context, promoCodeID, userID int64) (*PromoCodeUsage, error) {
	model, err := r.client.PromoCodeUsage.Query().
		Where(dbpromocodeusage.PromoCodeIDEQ(promoCodeID), dbpromocodeusage.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &PromoCodeUsage{
		ID:          model.ID,
		PromoCodeID: model.PromoCodeID,
		UserID:      model.UserID,
		BonusAmount: model.BonusAmount,
		UsedAt:      model.UsedAt,
	}, nil
}

func (r *entPromoRepoForAuthPromoTest) ListUsagesByPromoCode(context.Context, int64, pagination.PaginationParams) ([]PromoCodeUsage, *pagination.PaginationResult, error) {
	panic("unexpected ListUsagesByPromoCode call")
}

func (r *entPromoRepoForAuthPromoTest) IncrementUsedCount(ctx context.Context, id int64) error {
	client := r.client
	if tx := dbent.TxFromContext(ctx); tx != nil {
		client = tx.Client()
	}
	_, err := client.PromoCode.UpdateOneID(id).AddUsedCount(1).Save(ctx)
	return err
}

func promoCodeEntityToServiceForAuthPromoTest(model *dbent.PromoCode) *PromoCode {
	if model == nil {
		return nil
	}
	return &PromoCode{
		ID:                   model.ID,
		Code:                 model.Code,
		AllowedEmailSuffixes: append([]string{}, model.AllowedEmailSuffixes...),
		BonusAmount:          model.BonusAmount,
		MaxUses:              model.MaxUses,
		UsedCount:            model.UsedCount,
		Status:               model.Status,
		ExpiresAt:            model.ExpiresAt,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
}

func TestAuthService_RegisterWithVerification_PromoBonusTransaction(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite", "file:auth_service_promo_bonus?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	userRepo := &entUserRepoForAuthPromoTest{client: client}
	promoRepo := &entPromoRepoForAuthPromoTest{client: client}
	promoService := NewPromoService(promoRepo, userRepo, nil, client, nil)

	_, err = promoService.Create(ctx, &CreatePromoCodeInput{
		Code:                 "TEAM30",
		AllowedEmailSuffixes: []string{"@company.com"},
		BonusAmount:          100,
		MaxUses:              30,
	})
	require.NoError(t, err)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireHour: 1,
		},
		Default: config.DefaultConfig{
			UserBalance:     0,
			UserConcurrency: 1,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyRegistrationEnabled: "true",
		SettingKeyPromoCodeEnabled:    "true",
		SettingKeyDefaultBalance:      "0",
	}}, cfg)

	authService := NewAuthService(
		client,
		userRepo,
		nil,
		nil,
		cfg,
		settingService,
		nil,
		nil,
		nil,
		promoService,
		nil,
		nil,
		nil,
	)

	token, user, err := authService.RegisterWithVerification(ctx, "member@company.com", "password", "", "TEAM30", "", "")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, user)
	require.Equal(t, 100.0, user.Balance)

	storedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 100.0, storedUser.Balance)

	promoAfter, err := promoRepo.GetByCode(ctx, "TEAM30")
	require.NoError(t, err)
	require.Equal(t, 1, promoAfter.UsedCount)

	_, _, err = authService.RegisterWithVerification(ctx, "other@gmail.com", "password", "", "TEAM30", "", "")
	require.ErrorIs(t, err, ErrPromoCodeEmailSuffixNotAllowed)

	exists, err := userRepo.ExistsByEmail(ctx, "other@gmail.com")
	require.NoError(t, err)
	require.False(t, exists)
}
