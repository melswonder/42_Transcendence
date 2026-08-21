package domain

import (
	"time"

	"github.com/google/uuid"
)

// 対戦モード。matches.mode の CHECK 制約と同じ値にする。
const (
	ModeRanked = "ranked"
	ModeCasual = "casual"
	ModeAI     = "ai"
	ModeFriend = "friend"
)

// 決着のつき方。matches.result_type の CHECK 制約と同じ値にする。
const (
	ResultGoal    = "goal"    // ゴール到達
	ResultResign  = "resign"  // 投了
	ResultTimeout = "timeout" // 時間切れ
	ResultDraw    = "draw"
	ResultAbort   = "abort" // 中断
)

// 勝敗。match_participants.outcome の CHECK 制約と同じ値にする。
const (
	OutcomeWin  = "win"
	OutcomeLoss = "loss"
	OutcomeDraw = "draw"
)

// ParticipantsPerMatch は 1 対戦の参加者数。コリドールは常に 2 人。
const ParticipantsPerMatch = 2

// Match は決着した 1 対戦。
type Match struct {
	ID           uuid.UUID
	Mode         string
	ResultType   string
	TotalMoves   int
	StartedAt    time.Time
	FinishedAt   time.Time
	Participants []MatchParticipant
}

// MatchParticipant は 1 対戦の 1 人ぶんの結果。
// RatingAfter / XPGained は記録時にサーバーが計算して埋める。
type MatchParticipant struct {
	UserID       uuid.UUID
	Seat         int
	Outcome      string
	RatingBefore int
	RatingAfter  int
	XPGained     int
}

// MatchRecord は履歴として返す 1 件。自分から見た形に組み替えてある。
type MatchRecord struct {
	Match
	Opponent     User
	Outcome      string
	ResultType   string
	RatingBefore int
	RatingAfter  int
	XPGained     int
}

// ValidateMatchInput は記録しようとしている対戦が成立しているかを見る。
//
// DB の CHECK 制約は 1 行の中しか見られないので、
// 「2 人ぶんの勝敗が噛み合っているか」はここで弾く。
func ValidateMatchInput(mode, resultType string, startedAt, finishedAt time.Time, participants []MatchParticipant) error {
	if !isValidMode(mode) {
		return ErrInvalidMatchMode
	}
	if !isValidResultType(resultType) {
		return ErrInvalidResultType
	}
	if finishedAt.Before(startedAt) {
		return ErrInvalidMatchPeriod
	}
	if len(participants) != ParticipantsPerMatch {
		return ErrInvalidParticipants
	}

	seats := map[int]bool{}
	users := map[uuid.UUID]bool{}
	wins, draws := 0, 0

	for _, p := range participants {
		if !isValidOutcome(p.Outcome) {
			return ErrInvalidOutcome
		}
		if p.Seat != 0 && p.Seat != 1 {
			return ErrInvalidParticipants
		}
		// 同じ席・同じ人が 2 回出てくる（自分と自分の対戦）を弾く。
		if seats[p.Seat] || users[p.UserID] {
			return ErrInvalidParticipants
		}
		seats[p.Seat] = true
		users[p.UserID] = true

		switch p.Outcome {
		case OutcomeWin:
			wins++
		case OutcomeDraw:
			draws++
		}
	}

	// 成立するのは「1 勝 1 敗」か「2 人とも引き分け」だけ。
	// 2 人とも勝ちや、勝者がいるのに片方だけ引き分け、は矛盾している。
	switch {
	case wins == 1 && draws == 0:
	case wins == 0 && draws == ParticipantsPerMatch:
	default:
		return ErrInconsistentOutcome
	}

	// 引き分けなのに result_type が draw でない（またはその逆）も弾く。
	if (draws == ParticipantsPerMatch) != (resultType == ResultDraw) {
		return ErrInconsistentOutcome
	}

	return nil
}

func isValidMode(v string) bool {
	switch v {
	case ModeRanked, ModeCasual, ModeAI, ModeFriend:
		return true
	}

	return false
}

func isValidResultType(v string) bool {
	switch v {
	case ResultGoal, ResultResign, ResultTimeout, ResultDraw, ResultAbort:
		return true
	}

	return false
}

func isValidOutcome(v string) bool {
	switch v {
	case OutcomeWin, OutcomeLoss, OutcomeDraw:
		return true
	}

	return false
}
