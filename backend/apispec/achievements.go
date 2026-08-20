// 実績（Achievements）と進捗の仕様。
//
// 実績の定義そのものは DB ではなく domain 側の定数で持つ。DB に置くと
// 実績を 1 つ足すたびにデータ投入の migration が要るため。
// user_achievements テーブルには「いつ解除したか」だけを残す。
package apispec

import "time"

// Achievement は 1 つの実績と、その人の進捗。
type Achievement struct {
	Code        string `json:"code"        example:"first_win"`
	Name        string `json:"name"        example:"初勝利"`
	Description string `json:"description" example:"はじめて対戦に勝つ"`
	Category    string `json:"category"    enums:"wins,streak,matches,rating" example:"wins"`
	// Progress / Target は未解除の実績でも「10 戦中 7 戦」を出せるように常に入れる。
	Progress   int        `json:"progress"    example:"7"`
	Target     int        `json:"target"      example:"10"`
	Unlocked   bool       `json:"unlocked"    example:"false"`
	UnlockedAt *time.Time `json:"unlocked_at"` // 未解除なら null
}

// AchievementListResponse は実績の一覧。未解除も含めて全件返す。
type AchievementListResponse struct {
	Items         []Achievement `json:"items"`
	UnlockedCount int           `json:"unlocked_count" example:"4"`
	TotalCount    int           `json:"total_count"    example:"12"`
}

// ListMyAchievements godoc
//
//	@Summary		自分の実績
//	@Description	解除済み・未解除をまとめて返す。未解除のものには現在の進捗が入る。
//	@Tags			achievements
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AchievementListResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/achievements/me [get]
func ListMyAchievements() {}
