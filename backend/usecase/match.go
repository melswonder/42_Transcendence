package usecase

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// MatchFilter は履歴・集計の共通の絞り込み。ゼロ値のフィールドは絞らない。
type MatchFilter struct {
	From     *time.Time
	To       *time.Time
	Mode     string
	Outcome  string
	Opponent *uuid.UUID
	Limit    int
	Offset   int
}

// RecordMatch は 1 対戦とその参加者を保存し、参加者のレーティング・XP・レベルを更新する。
// 1 トランザクションで行うので、途中で失敗したら何も残さない。
// ListMatches は自分が参加した決着済みの対戦を新しい順に返す。
// UsersByID は存在しない ID を結果に入れない。
type MatchRepository interface {
	RecordMatch(ctx context.Context, match domain.Match) error
	ListMatches(ctx context.Context, userID uuid.UUID, f MatchFilter) ([]domain.MatchRecord, error)
	CountMatches(ctx context.Context, userID uuid.UUID, f MatchFilter) (int, error)
	UsersByID(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]domain.User, error)
}

// AchievementSyncer は対戦後に実績の解除を判定する。
// 実装は AchievementUsecase。同じ層の中なので interface 越しに繋ぐ。
type AchievementSyncer interface {
	SyncAfterMatch(ctx context.Context, userID uuid.UUID) error
}

// MatchNotifier は対戦が記録されたことを外へ知らせる。
// SSE の hub が実装するが、usecase は「誰かに伝わる」ことしか知らない。
type MatchNotifier interface {
	NotifyMatchRecorded(match domain.Match)
}

type MatchUsecase struct {
	repo         MatchRepository
	notifier     MatchNotifier
	achievements AchievementSyncer
	now          func() time.Time
}

func NewMatchUsecase(
	repo MatchRepository, notifier MatchNotifier, achievements AchievementSyncer,
) *MatchUsecase {
	return &MatchUsecase{repo: repo, notifier: notifier, achievements: achievements, now: time.Now}
}

// RecordMatch はレーティングと XP をここで計算する。クライアントの申告をそのまま保存すると
// いくらでも詐称できるため、受け取るのは「誰が・どの席で・勝ったか負けたか」だけ。
func (u *MatchUsecase) RecordMatch(
	ctx context.Context, match domain.Match,
) (*domain.Match, map[uuid.UUID]domain.User, error) {
	if err := domain.ValidateMatchInput(
		match.Mode, match.ResultType, match.StartedAt, match.FinishedAt, match.Participants,
	); err != nil {
		return nil, nil, err
	}

	userIDs := make([]uuid.UUID, 0, len(match.Participants))
	for _, p := range match.Participants {
		userIDs = append(userIDs, p.UserID)
	}

	users, err := u.repo.UsersByID(ctx, userIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, id := range userIDs {
		if _, ok := users[id]; !ok {
			return nil, nil, domain.ErrUserNotFound
		}
	}

	match.ID = uuid.New()
	match.Participants = applyRatings(match.Mode, match.Participants, users)

	if err := u.repo.RecordMatch(ctx, match); err != nil {
		return nil, nil, err
	}

	// 実績と通知は保存が終わってから。どちらも失敗しても記録は取り消さない。
	// 実績は次の対戦のときに改めて判定されるので、取りこぼしても自然に追いつく。
	u.syncAchievements(ctx, match)

	if u.notifier != nil {
		u.notifier.NotifyMatchRecorded(match)
	}

	return &match, users, nil
}

// applyRatings でレーティングが動くのはランク戦だけ。練習や身内戦で下がると、
// 対戦相手を選ぶ動機が歪むため。XP は全モードで入る。
func applyRatings(
	mode string, participants []domain.MatchParticipant, users map[uuid.UUID]domain.User,
) []domain.MatchParticipant {
	rated := make([]domain.MatchParticipant, len(participants))

	for i, p := range participants {
		opponent := participants[(i+1)%len(participants)]
		before := users[p.UserID].Rating

		after := before
		if mode == domain.ModeRanked {
			after = domain.NextRating(before, users[opponent.UserID].Rating, domain.OutcomeScore(p.Outcome))
		}

		p.RatingBefore = before
		p.RatingAfter = after
		p.XPGained = domain.XPForOutcome(p.Outcome)
		rated[i] = p
	}

	return rated
}

func (u *MatchUsecase) syncAchievements(ctx context.Context, match domain.Match) {
	if u.achievements == nil {
		return
	}

	for _, p := range match.Participants {
		if err := u.achievements.SyncAfterMatch(ctx, p.UserID); err != nil {
			log.Printf("sync achievements for %s: %v", p.UserID, err)
		}
	}
}

func (u *MatchUsecase) ListMatches(
	ctx context.Context, userID uuid.UUID, f MatchFilter,
) ([]domain.MatchRecord, int, error) {
	total, err := u.repo.CountMatches(ctx, userID, f)
	if err != nil {
		return nil, 0, err
	}

	records, err := u.repo.ListMatches(ctx, userID, f)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
