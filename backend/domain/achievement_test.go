package domain

import (
	"testing"
	"time"
)

func TestEvaluateAchievements(t *testing.T) {
	t.Parallel()

	stats := StatsSummary{Wins: 12, Losses: 5, Draws: 1, BestStreak: 3, Rating: 1250}
	got := EvaluateAchievements(stats, nil)

	if len(got) != len(Achievements) {
		t.Fatalf("未解除も含めて全件返すはずが %d 件", len(got))
	}

	byCode := map[string]Achievement{}
	for _, a := range got {
		byCode[a.Code] = a
	}

	// 条件を満たしているものは記録が無くても解除扱いになる。
	for _, code := range []string{"first_win", "win_10", "streak_3", "played_10"} {
		if !byCode[code].Unlocked {
			t.Errorf("%s は解除されているはず", code)
		}
	}

	// 届いていないものは未解除で、進捗だけ入る。
	if a := byCode["win_50"]; a.Unlocked || a.Progress != 12 {
		t.Errorf("win_50: unlocked=%v progress=%d, want false / 12", a.Unlocked, a.Progress)
	}

	// 進捗は Target で頭打ちにする（バーが振り切れないように）。
	if a := byCode["first_win"]; a.Progress != 1 {
		t.Errorf("first_win の progress = %d, want 1 (Target で頭打ち)", a.Progress)
	}
}

// 一度解除したものは、後で成績が下がっても外れない。
func TestUnlockedStaysUnlocked(t *testing.T) {
	t.Parallel()

	at := time.Now()
	unlocked := map[string]time.Time{"rating_1300": at}

	// レーティングが 1300 を割り込んだ状態。
	got := EvaluateAchievements(StatsSummary{Rating: 1100}, unlocked)

	for _, a := range got {
		if a.Code != "rating_1300" {
			continue
		}
		if !a.Unlocked || a.UnlockedAt == nil {
			t.Errorf("記録済みの実績が外れている: %+v", a)
		}
	}
}

func TestNewlyUnlocked(t *testing.T) {
	t.Parallel()

	// 10 勝 = 10 戦なので played_10 も同時に解除される。
	stats := StatsSummary{Wins: 10, BestStreak: 3, Rating: 1200}
	already := map[string]time.Time{"first_win": time.Now()}

	got := NewlyUnlocked(stats, already)

	want := map[string]bool{"win_10": true, "streak_3": true, "played_10": true}
	for _, code := range got {
		if !want[code] {
			t.Errorf("%s は新規解除に含まれないはず", code)
		}
		delete(want, code)
	}
	for code := range want {
		t.Errorf("%s が新規解除に含まれていない", code)
	}
}
