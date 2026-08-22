package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// FriendRepo はフレンド関係の永続化。
// friendships は (user_low_id < user_high_id) に正規化されているので、
// 読み書きの前に必ずここで並べ替える。外の層には low/high を見せない。
type FriendRepo struct {
	db *gorm.DB
}

func NewFriendRepo(db *gorm.DB) *FriendRepo {
	return &FriendRepo{db: db}
}

// normalizePair は 2 人を low/high の順に並べる。
func normalizePair(a, b uuid.UUID) (low, high uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

func (r *FriendRepo) GetPair(ctx context.Context, me, other uuid.UUID) (*usecase.FriendPair, error) {
	low, high := normalizePair(me, other)
	var row Friendship
	err := r.db.WithContext(ctx).
		First(&row, "user_low_id = ? AND user_high_id = ?", low, high).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrFriendshipNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find friendship: %w", err)
	}
	return &usecase.FriendPair{
		Status:        row.Status,
		RequestedByMe: row.RequestedByUserID == me,
	}, nil
}

func (r *FriendRepo) InsertPending(ctx context.Context, from, to uuid.UUID) error {
	low, high := normalizePair(from, to)
	row := Friendship{
		UserLowID:         low,
		UserHighID:        high,
		RequestedByUserID: from,
		Status:            domain.FriendStatusPending,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		// チェックと挿入の間に相手が先に申請した場合は複合 PK に当たる。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.ErrFriendAlreadyRequested
		}
		return fmt.Errorf("insert friendship: %w", err)
	}
	return nil
}

func (r *FriendRepo) UpdateStatus(ctx context.Context, me, other uuid.UUID, status string, requestedBy uuid.UUID) error {
	low, high := normalizePair(me, other)
	res := r.db.WithContext(ctx).Model(&Friendship{}).
		Where("user_low_id = ? AND user_high_id = ?", low, high).
		Updates(map[string]any{
			"status":               status,
			"requested_by_user_id": requestedBy,
		})
	if res.Error != nil {
		return fmt.Errorf("update friendship: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrFriendshipNotFound
	}
	return nil
}

func (r *FriendRepo) DeletePair(ctx context.Context, me, other uuid.UUID) error {
	low, high := normalizePair(me, other)
	res := r.db.WithContext(ctx).
		Where("user_low_id = ? AND user_high_id = ?", low, high).
		Delete(&Friendship{})
	if res.Error != nil {
		return fmt.Errorf("delete friendship: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrFriendshipNotFound
	}
	return nil
}

func (r *FriendRepo) GetFriendship(ctx context.Context, me, other uuid.UUID) (*domain.Friendship, error) {
	low, high := normalizePair(me, other)
	var row Friendship
	err := r.db.WithContext(ctx).
		Preload("UserLow").Preload("UserHigh").
		First(&row, "user_low_id = ? AND user_high_id = ?", low, high).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrFriendshipNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find friendship: %w", err)
	}
	return r.toDomainFriendship(&row, me)
}

func (r *FriendRepo) toDomainFriendship(row *Friendship, me uuid.UUID) (*domain.Friendship, error) {
	otherSide := row.UserHigh
	if row.UserHighID == me {
		otherSide = row.UserLow
	}
	if otherSide == nil {
		return nil, fmt.Errorf("friendship %s/%s is missing the other user", row.UserLowID, row.UserHighID)
	}
	return &domain.Friendship{
		Other:         *toDomainUser(otherSide),
		Status:        row.Status,
		RequestedByMe: row.RequestedByUserID == me,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (r *FriendRepo) ListFriendships(ctx context.Context, me uuid.UUID, f usecase.FriendListFilter) ([]domain.Friendship, int, error) {
	query := r.db.WithContext(ctx).Model(&Friendship{}).
		Where("user_low_id = ? OR user_high_id = ?", me, me)
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	switch f.Direction {
	case "incoming":
		query = query.Where("requested_by_user_id <> ?", me)
	case "outgoing":
		query = query.Where("requested_by_user_id = ?", me)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count friendships: %w", err)
	}

	var rows []Friendship
	err := query.Preload("UserLow").Preload("UserHigh").
		Order("updated_at DESC").
		Limit(f.Limit).Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list friendships: %w", err)
	}

	friendships := make([]domain.Friendship, 0, len(rows))
	for i := range rows {
		friendship, err := r.toDomainFriendship(&rows[i], me)
		if err != nil {
			return nil, 0, err
		}
		friendships = append(friendships, *friendship)
	}
	return friendships, int(total), nil
}

func (r *FriendRepo) HasBlockRelation(ctx context.Context, a, b uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Block{}).
		Where("(blocker_user_id = ? AND blocked_user_id = ?) OR (blocker_user_id = ? AND blocked_user_id = ?)",
			a, b, b, a).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check blocks: %w", err)
	}
	return count > 0, nil
}

func (r *FriendRepo) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND anonymized_at IS NULL AND status = 'active'", userID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check user: %w", err)
	}
	return count > 0, nil
}
