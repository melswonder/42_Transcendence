package domain

import "time"

// StatsSummary は 1 人ぶんの現在地。統計画面の数値タイルに対応する。
type StatsSummary struct {
	Wins          int
	Losses        int
	Draws         int
	CurrentStreak int // 連勝は正、連敗は負、引き分けを挟むと 0
	BestStreak    int
	Rating        int
	Ranking       int
	TotalPlayers  int
	Level         int
	XP            int
}

// TotalMatches は集計対象の対戦数。
func (s StatsSummary) TotalMatches() int {
	return s.Wins + s.Losses + s.Draws
}

// WinRate は勝率 (0.0-1.0)。引き分けも母数に含める。
// 「勝ちでも負けでもない試合」を無かったことにすると、
// 引き分けの多い人の勝率が実態より高く出るため。
func (s StatsSummary) WinRate() float64 {
	total := s.TotalMatches()
	if total == 0 {
		return 0
	}

	return float64(s.Wins) / float64(total)
}

// TimeseriesPoint は時系列 1 コマ。折れ線と棒の両方に使う。
type TimeseriesPoint struct {
	Date    time.Time
	Wins    int
	Losses  int
	Draws   int
	Matches int
	Rating  int // そのコマの最後の対戦を終えた時点のレーティング
}

// BreakdownSlice は内訳の 1 区分。
type BreakdownSlice struct {
	Key   string
	Count int
}

// Breakdown は円グラフ用の内訳。
type Breakdown struct {
	ByResultType []BreakdownSlice
	ByMode       []BreakdownSlice
	ByOutcome    []BreakdownSlice
}

// LeaderboardEntry はランキングの 1 行。
type LeaderboardEntry struct {
	Rank   int
	User   User
	Rating int
	Wins   int
	Losses int
}

// WinRate は勝率 (0.0-1.0)。引き分けは記録から引けないのでここでは母数に含めない。
func (e LeaderboardEntry) WinRate() float64 {
	total := e.Wins + e.Losses
	if total == 0 {
		return 0
	}

	return float64(e.Wins) / float64(total)
}

// 時系列の刻み幅。
const (
	IntervalDay  = "day"
	IntervalWeek = "week"
)

// NormalizeInterval は未知の値を既定 (day) に丸める。
func NormalizeInterval(v string) string {
	if v == IntervalWeek {
		return IntervalWeek
	}

	return IntervalDay
}

// Streaks は古い順に並んだ勝敗から、現在の連勝・連敗と最長連勝を求める。
//
// SQL の窓関数でも書けるが、連続の判定はクエリが読みにくくなるうえ
// テストしづらいので、勝敗の列だけ取ってきてここで数える。
func Streaks(outcomes []string) (current, best int) {
	run := 0
	for _, o := range outcomes {
		switch o {
		case OutcomeWin:
			// 連敗中なら 1 からやり直す。
			if run < 0 {
				run = 0
			}
			run++
			best = max(best, run)
		case OutcomeLoss:
			if run > 0 {
				run = 0
			}
			run--
		default:
			// 引き分けは連勝も連敗も切る。
			run = 0
		}
	}

	return run, best
}
