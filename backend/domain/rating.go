package domain

import "math"

// レーティング（Elo）の定数。
const (
	// InitialRating は users.rating のデフォルトと同じ値にする。
	InitialRating = 1200

	// kFactor は 1 戦で動く幅の上限。大きいほど直近の結果に振られる。
	// チェス連盟は実力帯で変えるが、ここでは一律で扱う。
	kFactor = 32
)

// 経験値とレベルの定数。
const (
	xpWin  = 50
	xpDraw = 30
	xpLoss = 20

	// xpPerLevel はレベルを 1 上げるのに必要な XP の基準値。
	// 必要量は level に比例して増える（レベル n → n+1 に xpPerLevel * n）。
	xpPerLevel = 100
)

// ExpectedScore は a が b に勝つ確率を返す（0.0-1.0）。
//
// レーティング差 400 でおよそ 10 倍の勝率差、というのが Elo の定義。
func ExpectedScore(ratingA, ratingB int) float64 {
	return 1 / (1 + math.Pow(10, float64(ratingB-ratingA)/400))
}

// NextRating は 1 戦ぶんのレーティングを返す。
//
// score は勝ち 1.0 / 引き分け 0.5 / 負け 0.0。
// 期待より良い結果なら上がり、悪ければ下がる。
// 0 未満にはしない（DB 側の CHECK と揃える）。
func NextRating(rating, opponentRating int, score float64) int {
	delta := kFactor * (score - ExpectedScore(rating, opponentRating))
	next := rating + int(math.Round(delta))

	return max(next, 0)
}

// OutcomeScore は勝敗を Elo の score に直す。
func OutcomeScore(outcome string) float64 {
	switch outcome {
	case OutcomeWin:
		return 1.0
	case OutcomeDraw:
		return 0.5
	default:
		return 0.0
	}
}

// XPForOutcome は 1 戦で得る経験値を返す。負けても入るのは、
// 遊んだこと自体を進捗として見せたいため。
func XPForOutcome(outcome string) int {
	switch outcome {
	case OutcomeWin:
		return xpWin
	case OutcomeDraw:
		return xpDraw
	default:
		return xpLoss
	}
}

// LevelForXP は累計 XP からレベルを求める。
//
// レベルは users.level に持っているが、正本はあくまで experience_points 側。
// 両方を独立に更新するとズレるので、level は常にこの関数の結果で上書きする。
func LevelForXP(xp int) int {
	level, remaining := 1, xp
	for {
		need := xpPerLevel * level
		if remaining < need {
			return level
		}
		remaining -= need
		level++
	}
}

// XPRangeForLevel は そのレベルの下限と上限の累計 XP を返す。
// 進捗バーの分母と、現在位置の計算に使う。
func XPRangeForLevel(level int) (floor, ceiling int) {
	for l := 1; l < level; l++ {
		floor += xpPerLevel * l
	}

	return floor, floor + xpPerLevel*level
}
