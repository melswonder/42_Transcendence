// 統計の集計 API の仕様。
//
// 元データは match_participants で、ここは読み取り専用の集計しか提供しない。
// 集計は毎回クエリで求める（キャッシュ表は持たない）。対戦数が数万件までなら
// インデックスで足りるし、二重管理でズレるほうが面倒なため。
package apispec

import "time"

// StatsSummary は自分の現在地。統計画面の上段の数値タイルに対応する。
type StatsSummary struct {
	Wins           int     `json:"wins"            example:"42"`
	Losses         int     `json:"losses"          example:"18"`
	Draws          int     `json:"draws"           example:"3"`
	TotalMatches   int     `json:"total_matches"   example:"63"`
	WinRate        float64 `json:"win_rate"        example:"0.667"` // 0.0-1.0。引き分けは母数に含む
	CurrentStreak  int     `json:"current_streak"  example:"3"`     // 連勝は正、連敗は負
	BestStreak     int     `json:"best_streak"     example:"7"`
	Rating         int     `json:"rating"          example:"1444"`
	Ranking        int     `json:"ranking"         example:"12"`  // rating の降順順位。同率は同順位
	TotalPlayers   int     `json:"total_players"   example:"250"` // 「250 人中 12 位」を作るため
	Level          int     `json:"level"           example:"7"`
	XP             int     `json:"xp"              example:"1420"`
	XPForNextLevel int     `json:"xp_for_next_level" example:"1600"` // このレベルの上限。進捗バーの分母
}

// TimeseriesPoint は時系列 1 コマ。折れ線と棒の両方に使う。
type TimeseriesPoint struct {
	Date    string `json:"date"    example:"2026-08-20"` // interval=week のときは週の開始日
	Wins    int    `json:"wins"    example:"3"`
	Losses  int    `json:"losses"  example:"1"`
	Draws   int    `json:"draws"   example:"0"`
	Matches int    `json:"matches" example:"4"`
	Rating  int    `json:"rating"  example:"1444"` // そのコマの最後の対戦後のレーティング
}

// TimeseriesResponse は時系列の集計結果。
type TimeseriesResponse struct {
	Interval string            `json:"interval" enums:"day,week" example:"day"`
	Points   []TimeseriesPoint `json:"points"`
}

// BreakdownSlice は内訳の 1 区分。円グラフの 1 切れ。
type BreakdownSlice struct {
	Key   string `json:"key"   example:"goal"`
	Count int    `json:"count" example:"28"`
}

// BreakdownResponse は結果種別とモードの内訳。
type BreakdownResponse struct {
	ByResultType []BreakdownSlice `json:"by_result_type"`
	ByMode       []BreakdownSlice `json:"by_mode"`
	ByOutcome    []BreakdownSlice `json:"by_outcome"`
}

// StatsUpdatedEvent は SSE で流れるイベントの本体。
//
// 統計そのものは載せない。載せると「イベントで運ばれた値」と
// 「API で取り直した値」の 2 経路ができて食い違うため、更新の合図だけを送り、
// 受け取った側が改めて取得する。
type StatsUpdatedEvent struct {
	MatchID   string    `json:"match_id" format:"uuid"`
	Outcome   string    `json:"outcome"  enums:"win,loss,draw"`
	Rating    int       `json:"rating"   example:"1444"`
	OccuredAt time.Time `json:"occurred_at"`
}

// GetMyStats godoc
//
//	@Summary		自分の統計サマリ
//	@Description	勝敗・勝率・連勝・レーティング・順位・レベル・XP をまとめて返す。
//	@Tags			stats
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query		string	false	"集計対象の開始日時"	format(date-time)
//	@Param			to		query		string	false	"集計対象の終了日時"	format(date-time)
//	@Param			mode	query		string	false	"対戦モード"			Enums(ranked, casual, ai, friend)
//	@Success		200		{object}	StatsSummary
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/stats/me [get]
func GetMyStats() {}

// GetMyTimeseries godoc
//
//	@Summary		自分の統計の推移
//	@Description	レーティングの推移（折れ線）と日別の対戦数（棒）に使う。
//	@Tags			stats
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from		query		string	false	"集計対象の開始日時"	format(date-time)
//	@Param			to			query		string	false	"集計対象の終了日時"	format(date-time)
//	@Param			mode		query		string	false	"対戦モード"			Enums(ranked, casual, ai, friend)
//	@Param			interval	query		string	false	"刻み幅"				Enums(day, week)	default(day)
//	@Success		200			{object}	TimeseriesResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Router			/stats/me/timeseries [get]
func GetMyTimeseries() {}

// GetMyBreakdown godoc
//
//	@Summary		自分の統計の内訳
//	@Description	結果種別・モード・勝敗ごとの件数。円グラフに使う。
//	@Tags			stats
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query		string	false	"集計対象の開始日時"	format(date-time)
//	@Param			to		query		string	false	"集計対象の終了日時"	format(date-time)
//	@Success		200		{object}	BreakdownResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/stats/me/breakdown [get]
func GetMyBreakdown() {}

// StreamStats godoc
//
//	@Summary		統計の更新通知 (SSE)
//	@Description	自分が参加した対戦が記録されるたびに 1 イベント流す。
//	@Description	接続は開きっぱなしになり、切断はクライアント側が閉じるか
//	@Description	サーバーが落ちるまで続く。ブラウザの EventSource は自動で再接続する。
//	@Tags			stats
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Success		200	{object}	StatsUpdatedEvent	"data: に JSON が 1 行ずつ流れる"
//	@Failure		401	{object}	ErrorResponse
//	@Router			/stats/stream [get]
func StreamStats() {}
