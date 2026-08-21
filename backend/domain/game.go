package domain

import "errors"

// 1 手の種類。match_actions.action_type の CHECK 制約と同じ値にする。
const (
	GameActionMove    = "move"    // 駒移動
	GameActionWall    = "wall"    // 壁配置
	GameActionResign  = "resign"  // 投了
	GameActionTimeout = "timeout" // 持ち時間切れ・再接続猶予切れ
	GameActionAbort   = "abort"   // 双方切断などによる中断
)

// 対局セッション（進行中の試合の出入りと同期）まわりの失敗。
// コリドールのルール違反は quoridor.go 側にある。
var (
	ErrMatchNotFound       = errors.New("match not found")
	ErrNotMatchPlayer      = errors.New("user is not a player of this match")
	ErrAlreadyInMatch      = errors.New("user already has an active match")
	ErrAlreadyQueued       = errors.New("user is already waiting in the queue")
	ErrStaleGameVersion    = errors.New("stale game version")
	ErrDuplicateGameAction = errors.New("duplicate game action")
	ErrUnknownGameAction   = errors.New("unknown game action")
)
