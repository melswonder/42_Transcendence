// 対局の WebSocket 入口。メッセージの解釈と JSON への詰め替えだけを担い、
// 盤面とルールの判断はすべて usecase / domain に任せる（サーバー権威型）。
package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// wsKeepAlive は接続維持の ping 間隔。中継機器のアイドル切断より短くする。
const wsKeepAlive = 25 * time.Second

type GameConfig struct {
	// AllowedOrigins はハンドシェイクを受け付ける Origin のホスト（host:port）。
	// WebSocket は CORS の保護を受けないので、ここで検証しないと
	// 悪意あるサイトから Cookie 付きで接続できてしまう（CSWSH）。
	AllowedOrigins []string
}

type GameHandler struct {
	uc          *usecase.GameUsecase
	currentUser func(*http.Request) (*domain.User, error)
	cfg         GameConfig
}

func NewGameHandler(uc *usecase.GameUsecase, currentUser func(*http.Request) (*domain.User, error), cfg GameConfig) *GameHandler {
	return &GameHandler{uc: uc, currentUser: currentUser, cfg: cfg}
}

// クライアント → サーバー。
type wsClientMessage struct {
	Type string `json:"type"` // join_queue | leave_queue | action

	// Type == action のとき。
	ActionID        string `json:"actionId"`
	ExpectedVersion int    `json:"expectedVersion"`
	Action          string `json:"action"` // move | wall | resign
	Row             int    `json:"row"`
	Col             int    `json:"col"`
	Orientation     string `json:"orientation"` // wall のとき h | v
}

// サーバー → クライアント。1 つの形に寄せて、type で見分ける。
type wsServerMessage struct {
	Type string `json:"type"` // queued | state | opponent_disconnected | opponent_reconnected | error

	State *wsGameState `json:"state,omitempty"`

	GraceSeconds int `json:"graceSeconds,omitempty"`

	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// wsGameState は盤面の全量。差分は送らない。
type wsGameState struct {
	MatchID      string      `json:"matchId"`
	Mode         string      `json:"mode"`
	Seat         int         `json:"seat"` // 受け取る側の座席
	Version      int         `json:"version"`
	Turn         int         `json:"turn"`
	Pawns        [2]wsCell   `json:"pawns"`
	WallsLeft    [2]int      `json:"wallsLeft"`
	Walls        []wsWall    `json:"walls"`
	MoveCount    int         `json:"moveCount"`
	LegalMoves   []wsCell    `json:"legalMoves"`
	Finished     bool        `json:"finished"`
	Winner       *int        `json:"winner,omitempty"`
	ResultType   string      `json:"resultType,omitempty"`
	TurnDeadline time.Time   `json:"turnDeadline"`
	Players      [2]wsPlayer `json:"players"`
	Connected    [2]bool     `json:"connected"`
}

type wsCell struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type wsWall struct {
	Orientation string `json:"orientation"`
	Row         int    `json:"row"`
	Col         int    `json:"col"`
}

type wsPlayer struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Handle      string `json:"handle"`
	Rating      int    `json:"rating"`
}

// WS - GET /game/ws
//
// Cookie セッションで認証してから WebSocket へ upgrade する。
// 進行中の対局があれば usecase 側が自動で復帰させ、最初の state が届く。
func (h *GameHandler) WS(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.cfg.AllowedOrigins,
	})
	if err != nil {
		// Accept が自身で 4xx を書いている。
		return
	}
	defer conn.CloseNow() //nolint:errcheck // 後始末の失敗は握りつぶしていい

	client, err := h.uc.Connect(r.Context(), user)
	if err != nil {
		log.Printf("game ws connect: %v", err)
		conn.Close(websocket.StatusInternalError, "connect failed")
		return
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// 書き込みは 1 goroutine に集める。websocket は並行書き込みできない。
	go h.writePump(ctx, conn, client)

	for {
		var msg wsClientMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return // 切断。client.Close() が猶予タイマーを起こす
		}
		h.dispatch(ctx, conn, client, msg)
	}
}

// writePump は usecase から届くイベントを JSON にして流し、あわせて keep-alive を打つ。
func (h *GameHandler) writePump(ctx context.Context, conn *websocket.Conn, client *usecase.GameClient) {
	ticker := time.NewTicker(wsKeepAlive)
	defer ticker.Stop()

	for {
		select {
		case ev := <-client.Events():
			if err := wsjson.Write(ctx, conn, h.toServerMessage(ev, client.Seat())); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.Ping(ctx); err != nil {
				return
			}
		case <-client.Done():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (h *GameHandler) dispatch(ctx context.Context, conn *websocket.Conn, client *usecase.GameClient, msg wsClientMessage) {
	switch msg.Type {
	case "join_queue":
		if err := client.JoinQueue(ctx); err != nil {
			h.writeError(ctx, conn, err)
		}
	case "leave_queue":
		client.LeaveQueue()
	case "action":
		input, err := toActionInput(msg)
		if err != nil {
			h.writeError(ctx, conn, err)
			return
		}
		if err := client.Act(ctx, input); err != nil {
			h.writeError(ctx, conn, err)
		}
	default:
		h.writeError(ctx, conn, domain.ErrUnknownGameAction)
	}
}

func toActionInput(msg wsClientMessage) (usecase.GameActionInput, error) {
	actionID, err := uuid.Parse(msg.ActionID)
	if err != nil {
		return usecase.GameActionInput{}, domain.ErrUnknownGameAction
	}
	input := usecase.GameActionInput{
		ActionID:        actionID,
		ExpectedVersion: msg.ExpectedVersion,
		Kind:            msg.Action,
	}
	switch msg.Action {
	case domain.GameActionMove:
		input.Cell = domain.Cell{Row: msg.Row, Col: msg.Col}
	case domain.GameActionWall:
		input.Wall = domain.Wall{Orientation: msg.Orientation, Row: msg.Row, Col: msg.Col}
	case domain.GameActionResign:
		// 追加の情報は要らない
	default:
		return usecase.GameActionInput{}, domain.ErrUnknownGameAction
	}
	return input, nil
}

func (h *GameHandler) toServerMessage(ev usecase.GameEvent, seat int) wsServerMessage {
	msg := wsServerMessage{Type: ev.Type, GraceSeconds: ev.GraceSeconds}
	if ev.State != nil {
		msg.State = toWSGameState(ev.State, seat)
	}
	return msg
}

func toWSGameState(v *usecase.GameStateView, seat int) *wsGameState {
	state := &wsGameState{
		MatchID:      v.MatchID.String(),
		Mode:         v.Mode,
		Seat:         seat,
		Version:      v.Version,
		Turn:         v.Turn,
		Pawns:        [2]wsCell{toWSCell(v.Pawns[0]), toWSCell(v.Pawns[1])},
		WallsLeft:    v.WallsLeft,
		Walls:        make([]wsWall, 0, len(v.Walls)),
		MoveCount:    v.MoveCount,
		LegalMoves:   make([]wsCell, 0, len(v.LegalMoves)),
		Finished:     v.Finished,
		ResultType:   v.ResultType,
		TurnDeadline: v.TurnDeadline,
		Connected:    v.Connected,
	}
	for _, w := range v.Walls {
		state.Walls = append(state.Walls, wsWall{Orientation: w.Orientation, Row: w.Row, Col: w.Col})
	}
	for _, c := range v.LegalMoves {
		state.LegalMoves = append(state.LegalMoves, toWSCell(c))
	}
	if v.Finished && v.Winner >= 0 {
		winner := v.Winner
		state.Winner = &winner
	}
	for i, p := range v.Players {
		state.Players[i] = wsPlayer{
			ID:          p.UserID.String(),
			DisplayName: p.DisplayName,
			Handle:      p.Handle,
			Rating:      p.Rating,
		}
	}
	return state
}

func toWSCell(c domain.Cell) wsCell {
	return wsCell{Row: c.Row, Col: c.Col}
}

// writeError は失敗をコード付きで返す。盤面は動かさないので state は付けない。
func (h *GameHandler) writeError(ctx context.Context, conn *websocket.Conn, err error) {
	code := "internal"
	switch {
	case errors.Is(err, domain.ErrNotYourTurn):
		code = "not_your_turn"
	case errors.Is(err, domain.ErrIllegalPawnMove), errors.Is(err, domain.ErrCellOutOfBoard):
		code = "illegal_move"
	case errors.Is(err, domain.ErrWallOverlaps):
		code = "wall_overlaps"
	case errors.Is(err, domain.ErrWallSealsGoal):
		code = "wall_seals_goal"
	case errors.Is(err, domain.ErrWallOutOfBoard), errors.Is(err, domain.ErrInvalidWall):
		code = "invalid_wall"
	case errors.Is(err, domain.ErrNoWallsLeft):
		code = "no_walls_left"
	case errors.Is(err, domain.ErrStaleGameVersion):
		code = "stale_version"
	case errors.Is(err, domain.ErrGameFinished):
		code = "game_finished"
	case errors.Is(err, domain.ErrMatchNotFound):
		code = "no_active_match"
	case errors.Is(err, domain.ErrAlreadyInMatch):
		code = "already_in_match"
	case errors.Is(err, domain.ErrNotMatchPlayer):
		code = "not_match_player"
	case errors.Is(err, domain.ErrUnknownGameAction):
		code = "invalid_message"
	default:
		log.Printf("game ws: %v", err)
	}
	_ = wsjson.Write(ctx, conn, wsServerMessage{Type: "error", Code: code, Message: err.Error()})
}
