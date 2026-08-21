package domain

import "time"

// 実績の分類。何を伸ばすと解除できるかで分ける。
const (
	CategoryWins    = "wins"
	CategoryStreak  = "streak"
	CategoryMatches = "matches"
	CategoryRating  = "rating"
)

// AchievementDef は実績 1 つの定義。
//
// マスタを DB に置かないのは、実績を 1 つ足すたびにデータ投入の migration が
// 必要になるため。user_achievements には「いつ解除したか」だけを残す。
type AchievementDef struct {
	Code        string
	Name        string
	Description string
	Category    string
	// Target は解除に必要な値。Category が指す指標と対で読む。
	Target int
}

// Achievements は全実績の定義。順序は表示順を兼ねる。
//
// 追加するときはここに 1 行足すだけでよい。Code は解除済みの記録と
// 突き合わせる鍵なので、一度出したものは変えないこと。
var Achievements = []AchievementDef{
	{"first_win", "初勝利", "はじめて対戦に勝つ", CategoryWins, 1},
	{"win_10", "常勝", "通算 10 勝する", CategoryWins, 10},
	{"win_50", "百戦錬磨", "通算 50 勝する", CategoryWins, 50},
	{"win_100", "覇者", "通算 100 勝する", CategoryWins, 100},

	{"streak_3", "三連覇", "3 連勝する", CategoryStreak, 3},
	{"streak_5", "無敗", "5 連勝する", CategoryStreak, 5},
	{"streak_10", "不動", "10 連勝する", CategoryStreak, 10},

	{"played_10", "常連", "10 戦する", CategoryMatches, 10},
	{"played_50", "歴戦", "50 戦する", CategoryMatches, 50},
	{"played_100", "求道者", "100 戦する", CategoryMatches, 100},

	{"rating_1300", "上級者", "レーティング 1300 に到達する", CategoryRating, 1300},
	{"rating_1500", "熟練者", "レーティング 1500 に到達する", CategoryRating, 1500},
}

// Achievement は定義に、その人の進捗を重ねたもの。
type Achievement struct {
	AchievementDef
	Progress   int
	Unlocked   bool
	UnlockedAt *time.Time
}

// progressFor は指標ごとの現在値を返す。
func progressFor(def AchievementDef, s StatsSummary) int {
	switch def.Category {
	case CategoryWins:
		return s.Wins
	case CategoryStreak:
		return s.BestStreak
	case CategoryMatches:
		return s.TotalMatches()
	case CategoryRating:
		return s.Rating
	default:
		return 0
	}
}

// EvaluateAchievements は現在の統計から、全実績の進捗と解除状態を組み立てる。
//
// unlockedAt には解除済みの記録（code → 解除時刻）を渡す。
// 一度解除したものは、後で成績が下がっても外れない。レーティングのように
// 上下する指標だと、解除済みの実績が消えたり戻ったりしてしまうため。
func EvaluateAchievements(s StatsSummary, unlockedAt map[string]time.Time) []Achievement {
	result := make([]Achievement, 0, len(Achievements))

	for _, def := range Achievements {
		progress := progressFor(def, s)

		achievement := Achievement{
			AchievementDef: def,
			// 進捗バーが Target を超えないよう頭を打っておく。
			Progress: min(progress, def.Target),
		}

		if at, ok := unlockedAt[def.Code]; ok {
			achievement.Unlocked = true
			achievement.UnlockedAt = &at
		} else if progress >= def.Target {
			achievement.Unlocked = true
		}

		result = append(result, achievement)
	}

	return result
}

// NewlyUnlocked は、まだ記録されていないが条件を満たしている実績の code を返す。
// 対戦を記録した直後に呼び、返ってきたぶんを user_achievements へ書く。
func NewlyUnlocked(s StatsSummary, unlockedAt map[string]time.Time) []string {
	var codes []string

	for _, def := range Achievements {
		if _, already := unlockedAt[def.Code]; already {
			continue
		}
		if progressFor(def, s) >= def.Target {
			codes = append(codes, def.Code)
		}
	}

	return codes
}
