// matches / match_participants テーブルに対応する公開 API の仕様。
//
// DB では 1 対戦につき match_participants が 2 行できるが、API では
// 「自分の結果」と「相手」に組み替えて 1 件として返す。行が 2 つに割れているのは
// 集計を素直に書くための都合なので、外には出さない。
package apispec

import "time"

// MatchResponse は自分から見た 1 対戦。
type MatchResponse struct {
	ID           string     `json:"id"            format:"uuid"`
	Mode         string     `json:"mode"          enums:"ranked,casual,ai,friend" example:"ranked"`
	Opponent     UserPublic `json:"opponent"`                                                            // AI 戦では handle が "bot" のダミーを返す
	Outcome      string     `json:"outcome"       enums:"win,loss,draw"           example:"win"`         // 自分から見た勝敗
	ResultType   string     `json:"result_type"   enums:"goal,resign,timeout,draw,abort" example:"goal"` // 決着のつき方
	RatingBefore int        `json:"rating_before" example:"1432"`
	RatingAfter  int        `json:"rating_after"  example:"1444"`
	RatingDiff   int        `json:"rating_diff"   example:"12"` // rating_after - rating_before。クライアントで引き算させない
	XPGained     int        `json:"xp_gained"     example:"50"`
	TotalMoves   int        `json:"total_moves"   example:"42"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   time.Time  `json:"finished_at"`
}

// MatchListResponse は対戦履歴の一覧。
type MatchListResponse struct {
	Items []MatchResponse `json:"items"`
	Pagination
}

// MatchParticipantInput は対戦結果に含まれる 1 人ぶんの申告。
type MatchParticipantInput struct {
	UserID  string `json:"user_id" binding:"required" format:"uuid" example:"3f0c1a2e-9a5b-4c7d-8e1f-2b3c4d5e6f70"`
	Seat    int    `json:"seat"    binding:"required" enums:"0,1" example:"0"` // 0=先手 1=後手
	Outcome string `json:"outcome" binding:"required" enums:"win,loss,draw" example:"win"`
}

// MatchCreateRequest は決着した対戦の記録。
//
// レーティングと XP はサーバー側で計算するのでクライアントからは受け取らない。
// 申告された値をそのまま保存すると、いくらでも詐称できてしまうため。
type MatchCreateRequest struct {
	Mode         string                  `json:"mode"         binding:"required" enums:"ranked,casual,ai,friend" example:"ranked"`
	ResultType   string                  `json:"result_type"  binding:"required" enums:"goal,resign,timeout,draw,abort" example:"goal"`
	TotalMoves   int                     `json:"total_moves"  example:"42" minimum:"0"`
	StartedAt    time.Time               `json:"started_at"   binding:"required"`
	FinishedAt   time.Time               `json:"finished_at"  binding:"required"`
	Participants []MatchParticipantInput `json:"participants" binding:"required"` // ちょうど 2 件
}

// ListMatches godoc
//
//	@Summary		対戦履歴
//	@Description	自分が参加した決着済みの対戦を新しい順に返す。日付範囲とモード・勝敗で絞り込める。
//	@Tags			matches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query		string	false	"この日時以降に終了した対戦"	format(date-time)
//	@Param			to		query		string	false	"この日時以前に終了した対戦"	format(date-time)
//	@Param			mode	query		string	false	"対戦モード"					Enums(ranked, casual, ai, friend)
//	@Param			outcome	query		string	false	"自分から見た勝敗"			Enums(win, loss, draw)
//	@Param			limit	query		int		false	"取得件数 (1-100)"			default(20)	minimum(1)	maximum(100)
//	@Param			offset	query		int		false	"取得開始位置"				default(0)	minimum(0)
//	@Success		200		{object}	MatchListResponse
//	@Failure		400		{object}	ErrorResponse	"from > to、または範囲外の値"
//	@Failure		401		{object}	ErrorResponse
//	@Router			/matches [get]
func ListMatches() {}

// ExportMatchesCSV godoc
//
//	@Summary		対戦履歴を CSV で書き出す
//	@Description	ListMatches と同じ絞り込みで、ページングだけ無視して全件を返す。
//	@Description	集計クエリの結果をそのまま流すのでサーバー側で生成する。
//	@Tags			matches
//	@Produce		text/csv
//	@Security		BearerAuth
//	@Param			from	query		string	false	"この日時以降に終了した対戦"	format(date-time)
//	@Param			to		query		string	false	"この日時以前に終了した対戦"	format(date-time)
//	@Param			mode	query		string	false	"対戦モード"					Enums(ranked, casual, ai, friend)
//	@Param			outcome	query		string	false	"自分から見た勝敗"			Enums(win, loss, draw)
//	@Success		200		{string}	string			"CSV 本文"
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/matches/export.csv [get]
func ExportMatchesCSV() {}

// CreateMatch godoc
//
//	@Summary		対戦結果を記録する
//	@Description	ゲーム側が決着時に呼ぶ入口。レーティング・XP・レベルの更新と実績の判定を行い、
//	@Description	参加者に対して SSE で更新を通知する。
//	@Tags			matches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		MatchCreateRequest	true	"決着した対戦"
//	@Success		201		{object}	MatchResponse		"呼び出したユーザーから見た結果"
//	@Failure		400		{object}	ErrorResponse	"参加者が 2 人でない、勝敗の組み合わせが矛盾している、終了が開始より前"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"参加者が存在しない"
//	@Router			/matches [post]
func CreateMatch() {}
