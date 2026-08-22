package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// fakeGameRepo は DB の代わり。ユニーク制約（冪等キー・版数）の挙動も再現する。
type fakeGameRepo struct {
	mu       sync.Mutex
	games    map[uuid.UUID]*StoredGame
	users    map[uuid.UUID]GamePlayer
	finished map[uuid.UUID]string // matchID → resultType
	winners  map[uuid.UUID]int
	settled  map[uuid.UUID][]domain.MatchParticipant
}

func newFakeGameRepo(users ...GamePlayer) *fakeGameRepo {
	r := &fakeGameRepo{
		games:    make(map[uuid.UUID]*StoredGame),
		users:    make(map[uuid.UUID]GamePlayer),
		finished: make(map[uuid.UUID]string),
		winners:  make(map[uuid.UUID]int),
		settled:  make(map[uuid.UUID][]domain.MatchParticipant),
	}
	for _, u := range users {
		r.users[u.UserID] = u
	}
	return r
}

func (r *fakeGameRepo) CreateMatch(_ context.Context, mode string, userIDs [2]uuid.UUID) (*StoredGame, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g := &StoredGame{
		MatchID:   uuid.New(),
		Mode:      mode,
		Players:   [2]GamePlayer{r.users[userIDs[0]], r.users[userIDs[1]]},
		StartedAt: time.Now(),
	}
	r.games[g.MatchID] = g
	return g, nil
}

func (r *fakeGameRepo) AppendAction(_ context.Context, record MatchActionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.games[record.MatchID]
	if !ok {
		return domain.ErrMatchNotFound
	}
	for _, a := range g.Actions {
		if a.ActionID == record.ActionID {
			return domain.ErrDuplicateGameAction
		}
		if a.Seq == record.Seq {
			return domain.ErrStaleGameVersion
		}
	}
	g.Actions = append(g.Actions, StoredAction{
		Seq:      record.Seq,
		ActionID: record.ActionID,
		Seat:     record.Seat,
		Type:     record.Type,
		Payload:  record.Payload,
	})
	return nil
}

func (r *fakeGameRepo) FinishMatch(_ context.Context, matchID uuid.UUID, resultType string, _ int, participants []domain.MatchParticipant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.finished[matchID]; ok {
		return domain.ErrMatchNotFound
	}
	r.finished[matchID] = resultType
	r.winners[matchID] = -1
	for _, p := range participants {
		if p.Outcome == domain.OutcomeWin {
			r.winners[matchID] = p.Seat
		}
	}
	r.settled[matchID] = participants
	return nil
}

func (r *fakeGameRepo) FindActiveMatchByID(_ context.Context, matchID uuid.UUID) (*StoredGame, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.games[matchID]
	if !ok {
		return nil, domain.ErrMatchNotFound
	}
	if _, done := r.finished[matchID]; done {
		return nil, domain.ErrMatchNotFound
	}
	copied := *g
	return &copied, nil
}

func (r *fakeGameRepo) ListLiveMatches(_ context.Context, _, _ int) ([]LiveMatch, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []LiveMatch
	for _, g := range r.games {
		if _, done := r.finished[g.MatchID]; done {
			continue
		}
		out = append(out, LiveMatch{
			MatchID:   g.MatchID,
			Mode:      g.Mode,
			StartedAt: g.StartedAt,
			Players:   g.Players,
			MoveCount: len(g.Actions),
		})
	}
	return out, len(out), nil
}

func (r *fakeGameRepo) FindActiveMatch(_ context.Context, userID uuid.UUID) (*StoredGame, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, g := range r.games {
		if _, done := r.finished[g.MatchID]; done {
			continue
		}
		if g.Players[0].UserID == userID || g.Players[1].UserID == userID {
			copied := *g
			return &copied, nil
		}
	}
	return nil, domain.ErrMatchNotFound
}

func (r *fakeGameRepo) result(matchID uuid.UUID) (string, int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt, ok := r.finished[matchID]
	return rt, r.winners[matchID], ok
}

func (r *fakeGameRepo) settledOf(matchID uuid.UUID) []domain.MatchParticipant {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settled[matchID]
}

func testUser(name string) *domain.User {
	return &domain.User{ID: uuid.New(), DisplayName: name, Handle: name}
}

func asPlayer(u *domain.User) GamePlayer {
	return GamePlayer{UserID: u.ID, DisplayName: u.DisplayName, Handle: u.Handle, Rating: 1200}
}

// waitEvent は指定した種類のイベントが届くまで読み進める。
func waitEvent(t *testing.T, c *GameClient, eventType string) GameEvent {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-c.Events():
			if ev.Type == eventType {
				return ev
			}
		case <-deadline:
			t.Fatalf("イベント %q が届かない", eventType)
		}
	}
}

// startTestMatch は 2 人をキューに入れて対戦を始め、初期盤面まで受け取る。
func startTestMatch(t *testing.T) (*fakeGameRepo, *GameUsecase, *GameClient, *GameClient, *GameStateView) {
	t.Helper()
	ctx := context.Background()
	u1, u2 := testUser("alice"), testUser("bob")
	repo := newFakeGameRepo(asPlayer(u1), asPlayer(u2))
	uc := NewGameUsecase(repo, nil, nil)

	c1, err := uc.Connect(ctx, u1)
	if err != nil {
		t.Fatalf("connect c1: %v", err)
	}
	c2, err := uc.Connect(ctx, u2)
	if err != nil {
		t.Fatalf("connect c2: %v", err)
	}

	if err := c1.JoinQueue(ctx); err != nil {
		t.Fatalf("join c1: %v", err)
	}
	waitEvent(t, c1, GameEventQueued)
	if err := c2.JoinQueue(ctx); err != nil {
		t.Fatalf("join c2: %v", err)
	}

	st1 := waitEvent(t, c1, GameEventState)
	waitEvent(t, c2, GameEventState)
	if c1.Seat() != 0 || c2.Seat() != 1 {
		t.Fatalf("先に並んだ側が seat 0 のはず: %d, %d", c1.Seat(), c2.Seat())
	}
	return repo, uc, c1, c2, st1.State
}

func move(c *GameClient, version, row, col int) error {
	return c.Act(context.Background(), GameActionInput{
		ActionID:        uuid.New(),
		ExpectedVersion: version,
		Kind:            domain.GameActionMove,
		Cell:            domain.Cell{Row: row, Col: col},
	})
}

func TestQueuePairsTwoPlayers(t *testing.T) {
	t.Parallel()

	_, _, _, _, state := startTestMatch(t)
	if state.Version != 0 || state.Turn != 0 || state.Finished {
		t.Errorf("初期盤面が変: %+v", state)
	}
	if !state.Connected[0] || !state.Connected[1] {
		t.Errorf("両者接続中のはず: %+v", state.Connected)
	}
}

func TestActAdvancesStateForBothPlayers(t *testing.T) {
	t.Parallel()

	_, _, c1, c2, _ := startTestMatch(t)
	if err := move(c1, 0, 1, 4); err != nil {
		t.Fatalf("合法手が拒否された: %v", err)
	}
	st1 := waitEvent(t, c1, GameEventState)
	st2 := waitEvent(t, c2, GameEventState)
	if st1.State.Version != 1 || st2.State.Version != 1 {
		t.Errorf("両者の版数が進むはず: %d, %d", st1.State.Version, st2.State.Version)
	}
	if st2.State.Turn != 1 {
		t.Errorf("手番が渡るはず: %d", st2.State.Turn)
	}
}

func TestDuplicateActionIsIdempotent(t *testing.T) {
	t.Parallel()

	_, _, c1, c2, _ := startTestMatch(t)
	actionID := uuid.New()
	in := GameActionInput{
		ActionID:        actionID,
		ExpectedVersion: 0,
		Kind:            domain.GameActionMove,
		Cell:            domain.Cell{Row: 1, Col: 4},
	}
	if err := c1.Act(context.Background(), in); err != nil {
		t.Fatalf("1 回目: %v", err)
	}
	waitEvent(t, c1, GameEventState)

	// 同じ actionID の再送。エラーにも二重適用にもならず、いまの盤面が返る。
	if err := c1.Act(context.Background(), in); err != nil {
		t.Fatalf("再送がエラーになった: %v", err)
	}
	st := waitEvent(t, c1, GameEventState)
	if st.State.Version != 1 {
		t.Errorf("再送で盤面が進んではいけない: version=%d", st.State.Version)
	}
	_ = c2
}

func TestStaleVersionIsRejected(t *testing.T) {
	t.Parallel()

	_, _, c1, c2, _ := startTestMatch(t)
	if err := move(c1, 0, 1, 4); err != nil {
		t.Fatalf("1 手目: %v", err)
	}
	// 古い盤面（version 0）を前提にした操作は拒否される。
	err := move(c2, 0, 7, 4)
	if !errors.Is(err, domain.ErrStaleGameVersion) {
		t.Errorf("古い版数は拒否されるはず: %v", err)
	}
	// 正しい版数なら通る。
	if err := move(c2, 1, 7, 4); err != nil {
		t.Errorf("正しい版数が拒否された: %v", err)
	}
}

func TestOutOfTurnIsRejected(t *testing.T) {
	t.Parallel()

	_, _, _, c2, _ := startTestMatch(t)
	err := move(c2, 0, 7, 4)
	if !errors.Is(err, domain.ErrNotYourTurn) {
		t.Errorf("手番外は拒否されるはず: %v", err)
	}
}

func TestResignFinishesMatch(t *testing.T) {
	t.Parallel()

	repo, _, c1, c2, state := startTestMatch(t)
	err := c2.Act(context.Background(), GameActionInput{
		ActionID: uuid.New(),
		Kind:     domain.GameActionResign,
	})
	if err != nil {
		t.Fatalf("投了できない: %v", err)
	}
	st := waitEvent(t, c1, GameEventState)
	if !st.State.Finished || st.State.Winner != 0 || st.State.ResultType != "resign" {
		t.Errorf("seat 0 の投了勝ちのはず: %+v", st.State)
	}
	if rt, winner, ok := repo.result(state.MatchID); !ok || rt != "resign" || winner != 0 {
		t.Errorf("resign / winner=0 で記録されるはず: %s %d %v", rt, winner, ok)
	}
}

// ランク戦なので決着で Elo と XP が動く。同レート同士は ±16（K=32）。
func TestRankedFinishSettlesRatings(t *testing.T) {
	t.Parallel()

	repo, _, c1, c2, state := startTestMatch(t)
	err := c2.Act(context.Background(), GameActionInput{
		ActionID: uuid.New(),
		Kind:     domain.GameActionResign,
	})
	if err != nil {
		t.Fatalf("投了できない: %v", err)
	}
	st := waitEvent(t, c1, GameEventState)

	settled := repo.settledOf(state.MatchID)
	if len(settled) != 2 {
		t.Fatalf("参加者 2 人ぶんが清算されるはず: %+v", settled)
	}
	if settled[0].RatingAfter != 1216 || settled[1].RatingAfter != 1184 {
		t.Errorf("同レートの勝敗は ±16 のはず: %d, %d", settled[0].RatingAfter, settled[1].RatingAfter)
	}
	if settled[0].XPGained != domain.XPForOutcome(domain.OutcomeWin) ||
		settled[1].XPGained != domain.XPForOutcome(domain.OutcomeLoss) {
		t.Errorf("XP が勝敗に応じて入るはず: %+v", settled)
	}
	if st.State.RatingAfter == nil || st.State.RatingAfter[0] != 1216 {
		t.Errorf("決着の state に新レーティングが載るはず: %+v", st.State.RatingAfter)
	}
}

func TestReconnectRestoresLatestState(t *testing.T) {
	t.Parallel()

	_, uc, c1, c2, _ := startTestMatch(t)
	if err := move(c1, 0, 1, 4); err != nil {
		t.Fatalf("1 手目: %v", err)
	}
	waitEvent(t, c2, GameEventState)

	// c2 が切断 → 新しい接続で戻ってくる。
	user2 := c2.user
	c2.Close()
	waitEvent(t, c1, GameEventOpponentDisconnected)

	c2b, err := uc.Connect(context.Background(), user2)
	if err != nil {
		t.Fatalf("再接続: %v", err)
	}
	st := waitEvent(t, c2b, GameEventState)
	if st.State.Version != 1 || st.State.Pawns[0] != (domain.Cell{Row: 1, Col: 4}) {
		t.Errorf("最新の盤面が復元されるはず: %+v", st.State)
	}
	if c2b.Seat() != 1 {
		t.Errorf("座席が戻るはず: %d", c2b.Seat())
	}
	waitEvent(t, c1, GameEventOpponentReconnected)
}

func TestRestoreFromRepositoryAfterRestart(t *testing.T) {
	t.Parallel()

	repo, _, c1, c2, state := startTestMatch(t)
	if err := move(c1, 0, 1, 4); err != nil {
		t.Fatalf("1 手目: %v", err)
	}
	waitEvent(t, c2, GameEventState)

	// サーバー再起動を模して、同じ repo から新しい usecase を作る。
	uc2 := NewGameUsecase(repo, nil, nil)
	c1b, err := uc2.Connect(context.Background(), c1.user)
	if err != nil {
		t.Fatalf("復元: %v", err)
	}
	st := waitEvent(t, c1b, GameEventState)
	if st.State.MatchID != state.MatchID || st.State.Version != 1 {
		t.Errorf("手のログから局面が復元されるはず: %+v", st.State)
	}
}

func TestDisconnectGraceTimeout(t *testing.T) {
	// タイマーを縮めるので並列にしない。
	restoreGrace, restoreTurn := ReconnectGrace, TurnTimeLimit
	ReconnectGrace, TurnTimeLimit = 50*time.Millisecond, time.Hour
	defer func() { ReconnectGrace, TurnTimeLimit = restoreGrace, restoreTurn }()

	repo, _, c1, c2, state := startTestMatch(t)
	c2.Close()
	waitEvent(t, c1, GameEventOpponentDisconnected)

	// 猶予を超えたら、切断した側の時間切れ負けで決着する。
	st := waitEvent(t, c1, GameEventState)
	if !st.State.Finished {
		t.Fatalf("猶予超過で決着するはず: %+v", st.State)
	}
	if rt, winner, ok := repo.result(state.MatchID); !ok || rt != "timeout" || winner != 0 {
		t.Errorf("timeout / winner=0 で記録されるはず: %s %d %v", rt, winner, ok)
	}
}

func TestTurnTimeout(t *testing.T) {
	// タイマーを縮めるので並列にしない。
	restoreTurn := TurnTimeLimit
	TurnTimeLimit = 60 * time.Millisecond
	defer func() { TurnTimeLimit = restoreTurn }()

	repo, _, c1, _, state := startTestMatch(t)

	// seat 0 が持ち時間を使い切る。
	st := waitEvent(t, c1, GameEventState)
	if !st.State.Finished {
		t.Fatalf("持ち時間切れで決着するはず: %+v", st.State)
	}
	if rt, winner, ok := repo.result(state.MatchID); !ok || rt != "timeout" || winner != 1 {
		t.Errorf("timeout / winner=1 で記録されるはず: %s %d %v", rt, winner, ok)
	}
}

func TestSpectatorFlow(t *testing.T) {
	t.Parallel()

	_, uc, c1, c2, state := startTestMatch(t)

	// 第三者が途中から観戦に入る。
	carol := testUser("carol")
	c3, err := uc.Connect(context.Background(), carol)
	if err != nil {
		t.Fatalf("connect c3: %v", err)
	}
	if err := c3.Watch(context.Background(), state.MatchID); err != nil {
		t.Fatalf("観戦に入れるはず: %v", err)
	}

	// 観戦者には最新の盤面が届き、座席は無い。
	st := waitEvent(t, c3, GameEventState)
	if st.State.Version != 0 || c3.Seat() != -1 {
		t.Errorf("最新盤面と seat -1 のはず: v=%d seat=%d", st.State.Version, c3.Seat())
	}
	if st.State.Spectators != 1 {
		t.Errorf("観戦者数 1 のはず: %d", st.State.Spectators)
	}
	// 対局者にも観戦者数が伝わる。
	stPlayer := waitEvent(t, c1, GameEventState)
	if stPlayer.State.Spectators != 1 {
		t.Errorf("対局者にも観戦者数が届くはず: %d", stPlayer.State.Spectators)
	}

	// 対局が進めば観戦者にもリアルタイムに届く。
	if err := move(c1, 0, 1, 4); err != nil {
		t.Fatalf("1 手目: %v", err)
	}
	st = waitEvent(t, c3, GameEventState)
	if st.State.Version != 1 {
		t.Errorf("観戦者にも手が届くはず: v=%d", st.State.Version)
	}

	// 観戦者は操作できない。
	err = c3.Act(context.Background(), GameActionInput{
		ActionID:        uuid.New(),
		ExpectedVersion: 1,
		Kind:            domain.GameActionMove,
		Cell:            domain.Cell{Row: 7, Col: 4},
	})
	if !errors.Is(err, domain.ErrNotMatchPlayer) {
		t.Errorf("観戦者の操作は拒否されるはず: %v", err)
	}

	// 観戦をやめると観戦者数が減って全員に伝わる。
	// チャネルには古い盤面も溜まっているので、観戦者数 0 の通知が来るまで読み進める。
	c3.Unwatch()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-c2.Events():
			if ev.Type == GameEventState && ev.State.Spectators == 0 {
				return
			}
		case <-deadline:
			t.Fatal("観戦者数 0 の通知が届かない")
		}
	}
}

func TestMultipleSpectatorsAndListLive(t *testing.T) {
	t.Parallel()

	_, uc, _, _, state := startTestMatch(t)

	// 観戦者 2 人。
	for i := range 2 {
		c, err := uc.Connect(context.Background(), testUser(fmt.Sprintf("watcher%d", i)))
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Watch(context.Background(), state.MatchID); err != nil {
			t.Fatal(err)
		}
		st := waitEvent(t, c, GameEventState)
		if st.State.Spectators != i+1 {
			t.Errorf("観戦者数が積み上がるはず: got %d want %d", st.State.Spectators, i+1)
		}
	}

	live, total, err := uc.ListLive(context.Background(), 20, 0)
	if err != nil || total != 1 || len(live) != 1 {
		t.Fatalf("進行中の対局が 1 件のはず: %v %d", err, total)
	}
	if live[0].MatchID != state.MatchID || live[0].Spectators != 2 {
		t.Errorf("一覧に観戦者数が載るはず: %+v", live[0])
	}
}

func TestWatchRejectsUnknownOrFinishedMatch(t *testing.T) {
	t.Parallel()

	repo, uc, _, c2, state := startTestMatch(t)

	// 存在しない対局。
	c, err := uc.Connect(context.Background(), testUser("nosy"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Watch(context.Background(), uuid.New()); !errors.Is(err, domain.ErrMatchNotFound) {
		t.Errorf("存在しない対局の観戦は 404 のはず: %v", err)
	}

	// 決着済みの対局。
	if err := c2.Act(context.Background(), GameActionInput{ActionID: uuid.New(), Kind: domain.GameActionResign}); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := repo.result(state.MatchID); !ok {
		t.Fatal("決着しているはず")
	}
	if err := c.Watch(context.Background(), state.MatchID); !errors.Is(err, domain.ErrMatchNotFound) {
		t.Errorf("決着済みの観戦は 404 のはず: %v", err)
	}
}

func TestQueueRejectsWhileInMatch(t *testing.T) {
	t.Parallel()

	_, _, c1, _, _ := startTestMatch(t)
	err := c1.JoinQueue(context.Background())
	if !errors.Is(err, domain.ErrAlreadyInMatch) {
		t.Errorf("対局中はキューに入れないはず: %v", err)
	}
}
