// Leaderboard の仕様。
//
// 順位は users.rating の降順で毎回求める。順位を列として持たないのは、
// 誰か 1 人のレーティングが動くたびに他の行も書き換えることになるため。
package apispec

// LeaderboardEntry はランキングの 1 行。
type LeaderboardEntry struct {
	Rank    int        `json:"rank"    example:"1"` // 同率は同順位（1,1,3 の付き方）
	User    UserPublic `json:"user"`
	Rating  int        `json:"rating"  example:"1830"`
	Wins    int        `json:"wins"    example:"120"`
	Losses  int        `json:"losses"  example:"40"`
	WinRate float64    `json:"win_rate" example:"0.75"`
}

// LeaderboardResponse はランキングと、その中での自分の位置。
type LeaderboardResponse struct {
	Items []LeaderboardEntry `json:"items"`
	// Me は自分が表示範囲の外にいても順位を出せるように別枠で返す。
	// 未対戦などで順位が付かない場合は null。
	Me *LeaderboardEntry `json:"me"`
	Pagination
}

// GetLeaderboard godoc
//
//	@Summary		ランキング
//	@Description	レーティングの高い順に返す。自分の順位は items の外でも me に入る。
//	@Tags			leaderboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"取得件数 (1-100)"	default(20)	minimum(1)	maximum(100)
//	@Param			offset	query		int	false	"取得開始位置"		default(0)	minimum(0)
//	@Success		200		{object}	LeaderboardResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/leaderboard [get]
func GetLeaderboard() {}
