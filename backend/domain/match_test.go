package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func participants(outcomeA, outcomeB string) []MatchParticipant {
	return []MatchParticipant{
		{UserID: uuid.New(), Seat: 0, Outcome: outcomeA},
		{UserID: uuid.New(), Seat: 1, Outcome: outcomeB},
	}
}

func TestValidateMatchInput(t *testing.T) {
	t.Parallel()

	start := time.Now().Add(-10 * time.Minute)
	end := time.Now()
	same := uuid.New()

	tests := []struct {
		name         string
		mode         string
		resultType   string
		startedAt    time.Time
		finishedAt   time.Time
		participants []MatchParticipant
		want         error
	}{
		{"勝敗が揃っている", ModeRanked, ResultGoal, start, end, participants(OutcomeWin, OutcomeLoss), nil},
		{"引き分け", ModeRanked, ResultDraw, start, end, participants(OutcomeDraw, OutcomeDraw), nil},

		{"知らないモード", "battle", ResultGoal, start, end, participants(OutcomeWin, OutcomeLoss), ErrInvalidMatchMode},
		{"知らない結果種別", ModeRanked, "surrender", start, end, participants(OutcomeWin, OutcomeLoss), ErrInvalidResultType},
		{"終了が開始より前", ModeRanked, ResultGoal, end, start, participants(OutcomeWin, OutcomeLoss), ErrInvalidMatchPeriod},

		{"2 人とも勝ち", ModeRanked, ResultGoal, start, end, participants(OutcomeWin, OutcomeWin), ErrInconsistentOutcome},
		{"2 人とも負け", ModeRanked, ResultGoal, start, end, participants(OutcomeLoss, OutcomeLoss), ErrInconsistentOutcome},
		{"片方だけ引き分け", ModeRanked, ResultGoal, start, end, participants(OutcomeWin, OutcomeDraw), ErrInconsistentOutcome},
		{"引き分けなのに result_type が goal", ModeRanked, ResultGoal, start, end, participants(OutcomeDraw, OutcomeDraw), ErrInconsistentOutcome},
		{"決着なのに result_type が draw", ModeRanked, ResultDraw, start, end, participants(OutcomeWin, OutcomeLoss), ErrInconsistentOutcome},

		{"参加者が 1 人", ModeRanked, ResultGoal, start, end,
			[]MatchParticipant{{UserID: uuid.New(), Seat: 0, Outcome: OutcomeWin}}, ErrInvalidParticipants},
		{"自分と自分の対戦", ModeRanked, ResultGoal, start, end,
			[]MatchParticipant{
				{UserID: same, Seat: 0, Outcome: OutcomeWin},
				{UserID: same, Seat: 1, Outcome: OutcomeLoss},
			}, ErrInvalidParticipants},
		{"席が重複", ModeRanked, ResultGoal, start, end,
			[]MatchParticipant{
				{UserID: uuid.New(), Seat: 0, Outcome: OutcomeWin},
				{UserID: uuid.New(), Seat: 0, Outcome: OutcomeLoss},
			}, ErrInvalidParticipants},
		{"知らない勝敗", ModeRanked, ResultGoal, start, end,
			[]MatchParticipant{
				{UserID: uuid.New(), Seat: 0, Outcome: "forfeit"},
				{UserID: uuid.New(), Seat: 1, Outcome: OutcomeLoss},
			}, ErrInvalidOutcome},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateMatchInput(tt.mode, tt.resultType, tt.startedAt, tt.finishedAt, tt.participants)
			if !errors.Is(err, tt.want) {
				t.Errorf("got %v, want %v", err, tt.want)
			}
		})
	}
}
