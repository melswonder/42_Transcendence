package domain

import "testing"

func TestNextRating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rating   int
		opponent int
		outcome  string
		want     int
	}{
		// 同格同士なら期待値 0.5。勝てば K/2 ぶん上がり、負ければ同じだけ下がる。
		{"同格に勝つ", 1200, 1200, OutcomeWin, 1216},
		{"同格に負ける", 1200, 1200, OutcomeLoss, 1184},
		{"同格と引き分け", 1200, 1200, OutcomeDraw, 1200},
		// 格上に勝つと大きく上がり、格下に勝ってもほとんど上がらない。
		{"格上に勝つ", 1200, 1600, OutcomeWin, 1229},
		{"格下に勝つ", 1600, 1200, OutcomeWin, 1603},
		{"格下に負ける", 1600, 1200, OutcomeLoss, 1571},
		// 下限。0 未満にはしない。
		{"下限で止まる", 5, 2000, OutcomeLoss, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NextRating(tt.rating, tt.opponent, OutcomeScore(tt.outcome))
			if got != tt.want {
				t.Errorf("NextRating(%d, %d, %s) = %d, want %d", tt.rating, tt.opponent, tt.outcome, got, tt.want)
			}
		})
	}
}

// 勝った側の増加と負けた側の減少は同じ幅になる（ゼロサム）。
func TestNextRatingIsZeroSum(t *testing.T) {
	t.Parallel()

	const a, b = 1450, 1310
	gained := NextRating(a, b, OutcomeScore(OutcomeWin)) - a
	lost := b - NextRating(b, a, OutcomeScore(OutcomeLoss))

	if gained != lost {
		t.Errorf("勝者 +%d に対して敗者 -%d。同じ幅になるはず", gained, lost)
	}
}

func TestLevelForXP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		xp   int
		want int
	}{
		{0, 1},
		{99, 1},
		{100, 2}, // Lv1→2 に 100
		{299, 2}, // Lv2→3 に 200（累計 300）
		{300, 3},
		{599, 3}, // Lv3→4 に 300（累計 600）
		{600, 4},
	}

	for _, tt := range tests {
		if got := LevelForXP(tt.xp); got != tt.want {
			t.Errorf("LevelForXP(%d) = %d, want %d", tt.xp, got, tt.want)
		}
	}
}

// LevelForXP と XPRangeForLevel が食い違わないこと。
// 片方だけ直したときに気付けるようにしておく。
func TestXPRangeMatchesLevel(t *testing.T) {
	t.Parallel()

	for xp := range 3000 {
		level := LevelForXP(xp)
		floor, ceiling := XPRangeForLevel(level)

		if xp < floor || xp >= ceiling {
			t.Fatalf("xp=%d は Lv%d と判定されたが、その範囲は [%d, %d)", xp, level, floor, ceiling)
		}
	}
}
