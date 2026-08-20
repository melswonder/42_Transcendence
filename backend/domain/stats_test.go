package domain

import "testing"

func TestStreaks(t *testing.T) {
	t.Parallel()

	w, l, d := OutcomeWin, OutcomeLoss, OutcomeDraw

	tests := []struct {
		name        string
		outcomes    []string
		wantCurrent int
		wantBest    int
	}{
		{"記録なし", nil, 0, 0},
		{"3 連勝中", []string{w, w, w}, 3, 3},
		{"2 連敗中", []string{l, l}, -2, 0},
		// 過去に 3 連勝したが、今は負けが続いている。
		{"最長は過去", []string{w, w, w, l, l}, -2, 3},
		// 引き分けは連勝も連敗も切る。
		{"引き分けで切れる", []string{w, w, d, w}, 1, 2},
		{"引き分けの直後", []string{w, w, d}, 0, 2},
		// 負けを挟んで勝ち直すと現在の連勝は 1 に戻る。
		{"負けを挟む", []string{w, w, l, w}, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			current, best := Streaks(tt.outcomes)
			if current != tt.wantCurrent || best != tt.wantBest {
				t.Errorf("Streaks(%v) = (%d, %d), want (%d, %d)",
					tt.outcomes, current, best, tt.wantCurrent, tt.wantBest)
			}
		})
	}
}

func TestWinRateIncludesDraws(t *testing.T) {
	t.Parallel()

	// 勝ち 1・負け 1・引き分け 2 なら 1/4。引き分けを母数から外すと 1/2 になってしまう。
	s := StatsSummary{Wins: 1, Losses: 1, Draws: 2}
	if got := s.WinRate(); got != 0.25 {
		t.Errorf("WinRate() = %v, want 0.25", got)
	}

	if got := (StatsSummary{}).WinRate(); got != 0 {
		t.Errorf("記録なしの WinRate() = %v, want 0", got)
	}
}
