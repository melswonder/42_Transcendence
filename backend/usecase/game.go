package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// 対局の時間まわり。テストで短縮できるよう var にしてある。
var (
	// TurnTimeLimit は 1 手の持ち時間。超えると手番側の時間切れ負け。
	TurnTimeLimit = 60 * time.Second
	// ReconnectGrace は切断から再接続までの猶予。超えると切断側の負け。
	ReconnectGrace = 45 * time.Second
)

const (
	// gameEventBuffer は 1 接続あたりのイベントバッファ。
	// 溢れたら最も古いイベントを捨てる。state は毎回全量を送るので、
	// 最新さえ届けば取りこぼしても盤面は壊れない。
	gameEventBuffer = 32

	// repoTimeout はタイマー起点など、リクエスト外で DB に触るときの上限。
	repoTimeout = 5 * time.Second

	// gameModeQuickMatch はクイックマッチで作る対戦のモード。
	// レーティング計算は統計モジュール側の仕事なので、当面 casual で記録する。
	gameModeQuickMatch = "casual"
)

// 決着のつき方。matches.result_type の CHECK 制約と同じ値にする。
const (
	gameResultGoal    = "goal"
	gameResultResign  = "resign"
	gameResultTimeout = "timeout"
	gameResultAbort   = "abort"
)

// GamePlayer は対局に参加している 1 人。盤面と一緒にクライアントへ返す表示用の情報。
type GamePlayer struct {
	UserID      uuid.UUID
	DisplayName string
	Handle      string
	Rating      int
}

// StoredAction は永続化された 1 手。
type StoredAction struct {
	Seq      int
	ActionID uuid.UUID
	Seat     int
	Type     string
	Payload  []byte
}

// StoredGame は DB から読み出した進行中の対局。リプレイで局面を復元する材料。
type StoredGame struct {
	MatchID   uuid.UUID
	Mode      string
	Players   [2]GamePlayer
	Actions   []StoredAction
	StartedAt time.Time
}

// MatchActionRecord は永続化する 1 手。
type MatchActionRecord struct {
	MatchID  uuid.UUID
	Seq      int
	ActionID uuid.UUID
	Seat     int
	Type     string
	Payload  []byte
}

// GameRepository は対局の永続化の受け口。
type GameRepository interface {
	// CreateMatch は matches 1 行と participants 2 行（未決着）を作る。
	CreateMatch(ctx context.Context, mode string, userIDs [2]uuid.UUID) (*StoredGame, error)
	// AppendAction は 1 手を追記する。同じ (match_id, action_id) なら
	// domain.ErrDuplicateGameAction、同じ (match_id, seq) なら domain.ErrStaleGameVersion。
	AppendAction(ctx context.Context, record MatchActionRecord) error
	// FinishMatch は対局を決着させる。winnerSeat が負なら中断（aborted）。
	FinishMatch(ctx context.Context, matchID uuid.UUID, resultType string, totalMoves, winnerSeat int) error
	// FindActiveMatch は user が参加している進行中の対局を返す。無ければ domain.ErrMatchNotFound。
	FindActiveMatch(ctx context.Context, userID uuid.UUID) (*StoredGame, error)
}

// GameActionInput はクライアントから届いた 1 操作。
type GameActionInput struct {
	// ActionID はクライアントが生成する冪等キー。再送されても 1 回しか適用しない。
	ActionID uuid.UUID
	// ExpectedVersion は操作の前提にした盤面の版数。ズレていたら古い操作として拒否する。
	ExpectedVersion int
	Kind            string // domain.GameActionMove / GameActionWall / GameActionResign
	Cell            domain.Cell
	Wall            domain.Wall
}

// クライアントへ流すイベントの種類。
const (
	GameEventQueued               = "queued"
	GameEventState                = "state"
	GameEventOpponentDisconnected = "opponent_disconnected"
	GameEventOpponentReconnected  = "opponent_reconnected"
)

// GameEvent は 1 接続へ届ける 1 通。
type GameEvent struct {
	Type         string
	State        *GameStateView // Type == state のとき
	GraceSeconds int            // Type == opponent_disconnected のとき
}

// GameStateView は盤面のスナップショット。差分ではなく毎回この全量を送る。
// 遅延や取りこぼしがあっても、最新の 1 通で盤面が揃う。
type GameStateView struct {
	MatchID      uuid.UUID
	Mode         string
	Version      int
	Turn         int
	Pawns        [2]domain.Cell
	WallsLeft    [2]int
	Walls        []domain.Wall
	MoveCount    int
	LegalMoves   []domain.Cell // 手番側の駒が動ける先
	Finished     bool
	Winner       int    // 未決着なら -1
	ResultType   string // 決着後のみ
	TurnDeadline time.Time
	Players      [2]GamePlayer
	Connected    [2]bool
}

// 手の中身。match_actions.payload に入れる形。
type movePayload struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type wallPayload struct {
	Orientation string `json:"orientation"`
	Row         int    `json:"row"`
	Col         int    `json:"col"`
}

// GameUsecase は進行中の対局と待機キューを束ねる。サーバー権威型の本体。
type GameUsecase struct {
	repo GameRepository

	mu       sync.Mutex
	sessions map[uuid.UUID]*gameSession // matchID → セッション
	byUser   map[uuid.UUID]*gameSession // 参加中ユーザー → セッション
	waiting  []*GameClient              // クイックマッチの待機列
}

func NewGameUsecase(repo GameRepository) *GameUsecase {
	return &GameUsecase{
		repo:     repo,
		sessions: make(map[uuid.UUID]*gameSession),
		byUser:   make(map[uuid.UUID]*gameSession),
	}
}

// GameClient は WebSocket 1 接続ぶんの窓口。handler が接続ごとに 1 つ作る。
type GameClient struct {
	u      *GameUsecase
	user   *domain.User
	events chan GameEvent
	done   chan struct{} // Close で閉じる。events 自体は閉じない（push との競合を避ける）

	mu      sync.Mutex
	session *gameSession
	seat    int
	closed  bool
}

// Connect は接続を登録する。進行中の対局があれば復帰させ、最新の盤面を events へ流す。
func (u *GameUsecase) Connect(ctx context.Context, user *domain.User) (*GameClient, error) {
	c := &GameClient{
		u:      u,
		user:   user,
		events: make(chan GameEvent, gameEventBuffer),
		done:   make(chan struct{}),
		seat:   -1,
	}

	s, err := u.findOrRestore(ctx, user.ID)
	if err != nil && !errors.Is(err, domain.ErrMatchNotFound) {
		return nil, err
	}
	if s != nil {
		s.attach(c)
		s.sendState(c)
	}
	return c, nil
}

// findOrRestore はメモリ上のセッションを探し、無ければ DB の進行中対局から復元する。
// サーバー再起動後の再接続はここを通る。
func (u *GameUsecase) findOrRestore(ctx context.Context, userID uuid.UUID) (*gameSession, error) {
	u.mu.Lock()
	if s, ok := u.byUser[userID]; ok {
		u.mu.Unlock()
		return s, nil
	}
	u.mu.Unlock()

	stored, err := u.repo.FindActiveMatch(ctx, userID)
	if err != nil {
		return nil, err
	}
	s, err := u.buildSession(stored)
	if err != nil {
		return nil, err
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	// 復元している間に別の接続が先に復元していたらそちらを使う。
	if existing, ok := u.sessions[stored.MatchID]; ok {
		return existing, nil
	}
	u.register(s)
	return s, nil
}

// buildSession は保存された手を初期局面から順に適用して局面を復元する。
func (u *GameUsecase) buildSession(stored *StoredGame) (*gameSession, error) {
	game := domain.NewQuoridor()
	applied := make(map[uuid.UUID]int, len(stored.Actions))
	for _, a := range stored.Actions {
		if err := replayAction(game, a); err != nil {
			return nil, fmt.Errorf("replay match %s seq %d: %w", stored.MatchID, a.Seq, err)
		}
		applied[a.ActionID] = a.Seq
	}
	return &gameSession{
		u:            u,
		matchID:      stored.MatchID,
		mode:         stored.Mode,
		players:      stored.Players,
		game:         game,
		version:      len(stored.Actions),
		applied:      applied,
		subs:         make(map[*GameClient]struct{}),
		turnDeadline: time.Now().Add(TurnTimeLimit),
	}, nil
}

func replayAction(game *domain.Quoridor, a StoredAction) error {
	switch a.Type {
	case domain.GameActionMove:
		var p movePayload
		if err := json.Unmarshal(a.Payload, &p); err != nil {
			return err
		}
		return game.MovePawn(a.Seat, domain.Cell{Row: p.Row, Col: p.Col})
	case domain.GameActionWall:
		var p wallPayload
		if err := json.Unmarshal(a.Payload, &p); err != nil {
			return err
		}
		return game.PlaceWall(a.Seat, domain.Wall{Orientation: p.Orientation, Row: p.Row, Col: p.Col})
	case domain.GameActionResign, domain.GameActionTimeout, domain.GameActionAbort:
		// 決着済みの対局は復元対象にならないはずだが、来ても壊れないように。
		return game.Resign(a.Seat)
	default:
		return domain.ErrUnknownGameAction
	}
}

// register は u.mu を握った状態で呼ぶ。
func (u *GameUsecase) register(s *gameSession) {
	u.sessions[s.matchID] = s
	u.byUser[s.players[0].UserID] = s
	u.byUser[s.players[1].UserID] = s
	s.armTurnTimer()
}

// evict は決着したセッションを外す。session の Lock を持たずに呼ぶこと。
func (u *GameUsecase) evict(s *gameSession) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.sessions, s.matchID)
	if u.byUser[s.players[0].UserID] == s {
		delete(u.byUser, s.players[0].UserID)
	}
	if u.byUser[s.players[1].UserID] == s {
		delete(u.byUser, s.players[1].UserID)
	}
}

// Events はこの接続へ流れるイベント。
func (c *GameClient) Events() <-chan GameEvent { return c.events }

// Done は Close されたら閉じる。イベント読み側の終了合図。
func (c *GameClient) Done() <-chan struct{} { return c.done }

// Seat はこの接続の座席。対局に入っていなければ -1。
func (c *GameClient) Seat() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.seat
}

// JoinQueue はクイックマッチの待機列に入る。相手がいれば即対戦を作る。
func (c *GameClient) JoinQueue(ctx context.Context) error {
	c.mu.Lock()
	if c.session != nil {
		c.mu.Unlock()
		return domain.ErrAlreadyInMatch
	}
	c.mu.Unlock()

	u := c.u
	u.mu.Lock()
	if _, ok := u.byUser[c.user.ID]; ok {
		u.mu.Unlock()
		return domain.ErrAlreadyInMatch
	}
	// 同じユーザーが並び直したら古いエントリを置き換える（別タブの開き直しなど）。
	for i, w := range u.waiting {
		if w.user.ID == c.user.ID {
			u.waiting[i] = c
			u.mu.Unlock()
			c.push(GameEvent{Type: GameEventQueued})
			return nil
		}
	}
	var opponent *GameClient
	if len(u.waiting) > 0 {
		opponent = u.waiting[0]
		u.waiting = u.waiting[1:]
	}
	if opponent == nil {
		u.waiting = append(u.waiting, c)
		u.mu.Unlock()
		c.push(GameEvent{Type: GameEventQueued})
		return nil
	}
	u.mu.Unlock()

	if err := u.startMatch(ctx, opponent, c); err != nil {
		// 対戦を作れなかったら相手を列へ戻す。
		u.mu.Lock()
		u.waiting = append([]*GameClient{opponent}, u.waiting...)
		u.mu.Unlock()
		return err
	}
	return nil
}

// startMatch は 2 人の対戦を作って両者に初期盤面を配る。seat 0 = 先に並んだ側。
func (u *GameUsecase) startMatch(ctx context.Context, first, second *GameClient) error {
	stored, err := u.repo.CreateMatch(ctx, gameModeQuickMatch, [2]uuid.UUID{first.user.ID, second.user.ID})
	if err != nil {
		return err
	}
	s, err := u.buildSession(stored)
	if err != nil {
		return err
	}

	u.mu.Lock()
	u.register(s)
	u.mu.Unlock()

	// 両者を合流させてから初期盤面を配る。順に配ると
	// 先に届いた側の Connected が [true, false] になってしまう。
	s.attach(first)
	s.attach(second)
	s.broadcastState()
	return nil
}

// LeaveQueue は待機列から抜ける。並んでいなければ何もしない。
func (c *GameClient) LeaveQueue() {
	u := c.u
	u.mu.Lock()
	defer u.mu.Unlock()
	for i, w := range u.waiting {
		if w == c {
			u.waiting = append(u.waiting[:i], u.waiting[i+1:]...)
			return
		}
	}
}

// Act は 1 操作を適用する。冪等・版数チェック・ルール検証を通ったものだけが盤面を進める。
func (c *GameClient) Act(ctx context.Context, in GameActionInput) error {
	c.mu.Lock()
	s, seat := c.session, c.seat
	c.mu.Unlock()
	if s == nil {
		return domain.ErrMatchNotFound
	}

	finished, err := s.act(ctx, c, seat, in)
	if err != nil {
		return err
	}
	if finished {
		c.u.evict(s)
	}
	return nil
}

// Close は接続の終了。待機列から抜き、対局中なら切断として扱う。
func (c *GameClient) Close() {
	c.LeaveQueue()

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.done)
	s, seat := c.session, c.seat
	c.mu.Unlock()

	if s != nil {
		s.detach(c, seat)
	}
}

// push はイベントを 1 通渡す。バッファが一杯なら一番古いものを捨てて必ず入れる。
func (c *GameClient) push(ev GameEvent) {
	for {
		select {
		case c.events <- ev:
			return
		default:
			select {
			case <-c.events:
			default:
			}
		}
	}
}

// gameSession は進行中の 1 対局。盤面・版数・接続状態・タイマーを 1 つの鍵で守る。
type gameSession struct {
	u *GameUsecase

	mu           sync.Mutex
	matchID      uuid.UUID
	mode         string
	players      [2]GamePlayer
	game         *domain.Quoridor
	version      int
	applied      map[uuid.UUID]int // actionID → 適用済み seq（冪等処理）
	resultType   string
	subs         map[*GameClient]struct{}
	conns        [2]int // 座席ごとの接続数
	turnTimer    *time.Timer
	graceTimers  [2]*time.Timer
	turnDeadline time.Time
}

func (s *gameSession) seatOf(userID uuid.UUID) int {
	if s.players[0].UserID == userID {
		return 0
	}
	if s.players[1].UserID == userID {
		return 1
	}
	return -1
}

// attach は接続を対局に合流させる。盤面は送らないので、
// 呼び出し側が sendState / broadcastState で配ること。
func (s *gameSession) attach(c *GameClient) {
	seat := s.seatOf(c.user.ID)
	if seat < 0 {
		return
	}

	s.mu.Lock()
	s.subs[c] = struct{}{}
	s.conns[seat]++
	reconnected := s.conns[seat] == 1
	if reconnected && s.graceTimers[seat] != nil {
		s.graceTimers[seat].Stop()
		s.graceTimers[seat] = nil
	}
	s.mu.Unlock()

	c.mu.Lock()
	c.session = s
	c.seat = seat
	c.mu.Unlock()

	if reconnected {
		s.broadcastExcept(c, GameEvent{Type: GameEventOpponentReconnected})
	}
}

// sendState はいまの盤面を 1 接続へ送る。
func (s *gameSession) sendState(c *GameClient) {
	s.mu.Lock()
	view := s.viewLocked()
	s.mu.Unlock()
	c.push(GameEvent{Type: GameEventState, State: view})
}

// broadcastState はいまの盤面を全接続へ送る。
func (s *gameSession) broadcastState() {
	s.mu.Lock()
	view := s.viewLocked()
	s.mu.Unlock()
	s.broadcastExcept(nil, GameEvent{Type: GameEventState, State: view})
}

// detach は接続の離脱。座席の接続が 0 になったら相手に知らせて猶予タイマーを起こす。
func (s *gameSession) detach(c *GameClient, seat int) {
	s.mu.Lock()
	delete(s.subs, c)
	if seat >= 0 {
		s.conns[seat]--
	}
	lastConn := seat >= 0 && s.conns[seat] == 0 && !s.game.Finished()
	if lastConn {
		s.graceTimers[seat] = time.AfterFunc(ReconnectGrace, func() { s.onGraceExpired(seat) })
	}
	s.mu.Unlock()

	if lastConn {
		s.broadcastExcept(nil, GameEvent{
			Type:         GameEventOpponentDisconnected,
			GraceSeconds: int(ReconnectGrace / time.Second),
		})
	}
}

// act は 1 操作の本体。適用できたら全員に新しい盤面を配る。
func (s *gameSession) act(ctx context.Context, from *GameClient, seat int, in GameActionInput) (finished bool, err error) {
	s.mu.Lock()

	if seat < 0 {
		s.mu.Unlock()
		return false, domain.ErrNotMatchPlayer
	}
	if s.game.Finished() {
		s.mu.Unlock()
		return false, domain.ErrGameFinished
	}

	// 冪等処理: 適用済みの actionID なら何もせず、いまの盤面を送り直す。
	if _, ok := s.applied[in.ActionID]; ok {
		view := s.viewLocked()
		s.mu.Unlock()
		from.push(GameEvent{Type: GameEventState, State: view})
		return false, nil
	}

	// 古い盤面を前提にした操作は拒否する（投了はいつでも受ける）。
	if in.Kind != domain.GameActionResign && in.ExpectedVersion != s.version {
		s.mu.Unlock()
		return false, domain.ErrStaleGameVersion
	}

	// 先に複製へ適用してルールを検証し、永続化に成功してから盤面を差し替える。
	trial := s.game.Clone()
	var payload []byte
	switch in.Kind {
	case domain.GameActionMove:
		err = trial.MovePawn(seat, in.Cell)
		payload, _ = json.Marshal(movePayload{Row: in.Cell.Row, Col: in.Cell.Col})
	case domain.GameActionWall:
		err = trial.PlaceWall(seat, in.Wall)
		payload, _ = json.Marshal(wallPayload{Orientation: in.Wall.Orientation, Row: in.Wall.Row, Col: in.Wall.Col})
	case domain.GameActionResign:
		err = trial.Resign(seat)
		payload = []byte("{}")
	default:
		err = domain.ErrUnknownGameAction
	}
	if err != nil {
		s.mu.Unlock()
		return false, err
	}

	record := MatchActionRecord{
		MatchID:  s.matchID,
		Seq:      s.version + 1,
		ActionID: in.ActionID,
		Seat:     seat,
		Type:     in.Kind,
		Payload:  payload,
	}
	if err := s.u.repo.AppendAction(ctx, record); err != nil {
		if errors.Is(err, domain.ErrDuplicateGameAction) {
			// 別経路で適用済み（再送の競り合い）。いまの盤面を返すだけでいい。
			view := s.viewLocked()
			s.mu.Unlock()
			from.push(GameEvent{Type: GameEventState, State: view})
			return false, nil
		}
		s.mu.Unlock()
		return false, err
	}

	s.game = trial
	s.version++
	s.applied[in.ActionID] = s.version

	if s.game.Finished() {
		result := gameResultGoal
		if in.Kind == domain.GameActionResign {
			result = gameResultResign
		}
		if err := s.finishLocked(ctx, result, s.game.Winner); err != nil {
			s.mu.Unlock()
			return false, err
		}
		view := s.viewLocked()
		s.mu.Unlock()
		s.broadcastExcept(nil, GameEvent{Type: GameEventState, State: view})
		return true, nil
	}

	s.armTurnTimerLocked()
	view := s.viewLocked()
	s.mu.Unlock()
	s.broadcastExcept(nil, GameEvent{Type: GameEventState, State: view})
	return false, nil
}

// onTurnTimeout は持ち時間切れ。手番側の負けで決着させる。
func (s *gameSession) onTurnTimeout() {
	s.forceFinish(func(s *gameSession) (loserSeat int, actionType, resultType string, skip bool) {
		return s.game.Turn, domain.GameActionTimeout, gameResultTimeout, false
	})
}

// onGraceExpired は再接続の猶予切れ。戻ってこなかった側の負け。
// 両者とも居なければ中断として決着させる。
func (s *gameSession) onGraceExpired(seat int) {
	s.forceFinish(func(s *gameSession) (loserSeat int, actionType, resultType string, skip bool) {
		if s.conns[seat] > 0 {
			return 0, "", "", true // 猶予内に戻ってきた
		}
		if s.conns[1-seat] == 0 {
			return seat, domain.GameActionAbort, gameResultAbort, false
		}
		return seat, domain.GameActionTimeout, gameResultTimeout, false
	})
}

// forceFinish は盤外の理由（時間切れ・中断）で対局を終わらせる。
// decide はロック中に呼ばれ、負ける座席と記録の種類を決める。
func (s *gameSession) forceFinish(decide func(*gameSession) (loserSeat int, actionType, resultType string, skip bool)) {
	ctx, cancel := context.WithTimeout(context.Background(), repoTimeout)
	defer cancel()

	s.mu.Lock()
	if s.game.Finished() {
		s.mu.Unlock()
		return
	}
	loserSeat, actionType, resultType, skip := decide(s)
	if skip {
		s.mu.Unlock()
		return
	}

	actionID := uuid.New() // サーバー起点の手なのでサーバーが冪等キーを振る
	record := MatchActionRecord{
		MatchID:  s.matchID,
		Seq:      s.version + 1,
		ActionID: actionID,
		Seat:     loserSeat,
		Type:     actionType,
		Payload:  []byte("{}"),
	}
	if err := s.u.repo.AppendAction(ctx, record); err != nil {
		s.mu.Unlock()
		return
	}
	_ = s.game.Resign(loserSeat) // 未決着は上で確認済み
	s.version++
	s.applied[actionID] = s.version

	winner := s.game.Winner
	if resultType == gameResultAbort {
		winner = -1
	}
	if err := s.finishLocked(ctx, resultType, winner); err != nil {
		s.mu.Unlock()
		return
	}
	view := s.viewLocked()
	s.mu.Unlock()

	s.broadcastExcept(nil, GameEvent{Type: GameEventState, State: view})
	s.u.evict(s)
}

// finishLocked は決着の永続化とタイマー停止。s.mu を握って呼ぶ。
// ロックは解放しないので、呼び出し側が viewLocked → Unlock → broadcast の順で締める。
func (s *gameSession) finishLocked(ctx context.Context, resultType string, winnerSeat int) error {
	if err := s.u.repo.FinishMatch(ctx, s.matchID, resultType, s.game.MoveCount, winnerSeat); err != nil {
		return err
	}
	s.resultType = resultType
	s.stopTimersLocked()
	return nil
}

func (s *gameSession) armTurnTimer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.armTurnTimerLocked()
}

func (s *gameSession) armTurnTimerLocked() {
	if s.turnTimer != nil {
		s.turnTimer.Stop()
	}
	s.turnDeadline = time.Now().Add(TurnTimeLimit)
	s.turnTimer = time.AfterFunc(TurnTimeLimit, s.onTurnTimeout)
}

func (s *gameSession) stopTimersLocked() {
	if s.turnTimer != nil {
		s.turnTimer.Stop()
	}
	for i, t := range s.graceTimers {
		if t != nil {
			t.Stop()
			s.graceTimers[i] = nil
		}
	}
}

// broadcastExcept は except 以外の全接続へ配る。except が nil なら全員へ。
func (s *gameSession) broadcastExcept(except *GameClient, ev GameEvent) {
	s.mu.Lock()
	targets := make([]*GameClient, 0, len(s.subs))
	for sub := range s.subs {
		if sub != except {
			targets = append(targets, sub)
		}
	}
	s.mu.Unlock()
	for _, sub := range targets {
		sub.push(ev)
	}
}

// viewLocked は現在の盤面スナップショット。s.mu を握って呼ぶ。
func (s *gameSession) viewLocked() *GameStateView {
	view := &GameStateView{
		MatchID:      s.matchID,
		Mode:         s.mode,
		Version:      s.version,
		Turn:         s.game.Turn,
		Pawns:        s.game.Pawns,
		WallsLeft:    s.game.WallsLeft,
		Walls:        append([]domain.Wall(nil), s.game.Walls...),
		MoveCount:    s.game.MoveCount,
		Finished:     s.game.Finished(),
		Winner:       s.game.Winner,
		ResultType:   s.resultType,
		TurnDeadline: s.turnDeadline,
		Players:      s.players,
		Connected:    [2]bool{s.conns[0] > 0, s.conns[1] > 0},
	}
	if !view.Finished {
		view.LegalMoves = s.game.LegalPawnMoves(s.game.Turn)
	}
	return view
}
