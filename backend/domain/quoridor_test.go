package domain

import (
	"errors"
	"slices"
	"testing"
)

// sortCells はマスの一覧を比較しやすいように並べ替えたコピーを返す。
func sortCells(cells []Cell) []Cell {
	sorted := slices.Clone(cells)
	slices.SortFunc(sorted, func(a, b Cell) int {
		if a.Row != b.Row {
			return a.Row - b.Row
		}
		return a.Col - b.Col
	})
	return sorted
}

func TestNewQuoridor(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()

	if q.Pawns[0] != (Cell{Row: 0, Col: 4}) || q.Pawns[1] != (Cell{Row: 8, Col: 4}) {
		t.Errorf("初期配置が違う: %+v", q.Pawns)
	}
	if q.WallsLeft[0] != WallsPerPlayer || q.WallsLeft[1] != WallsPerPlayer {
		t.Errorf("持ち壁は各 %d 枚のはず: %+v", WallsPerPlayer, q.WallsLeft)
	}
	if q.Turn != 0 {
		t.Errorf("先手は seat 0 のはず: %d", q.Turn)
	}
	if q.Finished() {
		t.Error("開始直後に決着している")
	}
}

func TestMovePawn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *Quoridor
		seat    int
		to      Cell
		wantErr error
	}{
		{
			"前へ 1 マス進める",
			NewQuoridor,
			0, Cell{Row: 1, Col: 4}, nil,
		},
		{
			"手番でない側は指せない",
			NewQuoridor,
			1, Cell{Row: 7, Col: 4}, ErrNotYourTurn,
		},
		{
			"斜めには動けない",
			NewQuoridor,
			0, Cell{Row: 1, Col: 5}, ErrIllegalPawnMove,
		},
		{
			"2 マスは進めない",
			NewQuoridor,
			0, Cell{Row: 2, Col: 4}, ErrIllegalPawnMove,
		},
		{
			"盤の外には出られない",
			NewQuoridor,
			0, Cell{Row: -1, Col: 4}, ErrCellOutOfBoard,
		},
		{
			"壁の向こうへは動けない",
			func() *Quoridor {
				q := NewQuoridor()
				// (0,4) の下（row 0-1 の間、col 4）を横壁で塞ぐ。
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 0, Col: 4}}
				return q
			},
			0, Cell{Row: 1, Col: 4}, ErrIllegalPawnMove,
		},
		{
			"存在しない座席は拒否する",
			NewQuoridor,
			2, Cell{Row: 1, Col: 4}, ErrInvalidSeat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q := tt.setup()
			err := q.MovePawn(tt.seat, tt.to)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MovePawn(%d, %+v) = %v, want %v", tt.seat, tt.to, err, tt.wantErr)
			}
		})
	}
}

func TestMovePawnUpdatesState(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()
	if err := q.MovePawn(0, Cell{Row: 1, Col: 4}); err != nil {
		t.Fatalf("合法手が拒否された: %v", err)
	}
	if q.Pawns[0] != (Cell{Row: 1, Col: 4}) {
		t.Errorf("駒が動いていない: %+v", q.Pawns[0])
	}
	if q.Turn != 1 {
		t.Errorf("手番が渡っていない: %d", q.Turn)
	}
	if q.MoveCount != 1 {
		t.Errorf("手数が増えていない: %d", q.MoveCount)
	}
}

func TestLegalPawnMovesJump(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Quoridor
		seat  int
		want  []Cell
	}{
		{
			// 相手が正面にいて後ろが空いていれば飛び越える。斜めには行けない。
			"正面ジャンプ",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns = [2]Cell{{Row: 3, Col: 4}, {Row: 4, Col: 4}}
				return q
			},
			0,
			[]Cell{
				{Row: 2, Col: 4}, {Row: 3, Col: 3}, {Row: 3, Col: 5}, // 通常移動
				{Row: 5, Col: 4}, // ジャンプ
			},
		},
		{
			// ジャンプ先が壁で塞がれていたら、相手駒の左右へ斜めに動ける。
			"壁ごしのサイドステップ",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns = [2]Cell{{Row: 3, Col: 4}, {Row: 4, Col: 4}}
				// (4,4) の下（row 4-5 の間）を塞ぐ。
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			0,
			[]Cell{
				{Row: 2, Col: 4}, {Row: 3, Col: 3}, {Row: 3, Col: 5},
				{Row: 4, Col: 3}, {Row: 4, Col: 5}, // サイドステップ
			},
		},
		{
			// 相手が盤端にいてジャンプ先が盤外でもサイドステップになる。
			"盤端でのサイドステップ",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns = [2]Cell{{Row: 7, Col: 4}, {Row: 8, Col: 4}}
				return q
			},
			0,
			[]Cell{
				{Row: 6, Col: 4}, {Row: 7, Col: 3}, {Row: 7, Col: 5},
				{Row: 8, Col: 3}, {Row: 8, Col: 5},
			},
		},
		{
			// サイドステップ先の片方が壁で塞がれていれば残りだけ。
			// v(7,4) は row 7-8 の col 4-5 間を塞ぐので (7,5) と (8,5) の両方が消える。
			"片側だけのサイドステップ",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns = [2]Cell{{Row: 7, Col: 4}, {Row: 8, Col: 4}}
				q.Walls = []Wall{{Orientation: WallVertical, Row: 7, Col: 4}}
				return q
			},
			0,
			[]Cell{
				{Row: 6, Col: 4}, {Row: 7, Col: 3},
				{Row: 8, Col: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q := tt.setup()
			got := sortCells(q.LegalPawnMoves(tt.seat))
			want := sortCells(tt.want)
			if !slices.Equal(got, want) {
				t.Errorf("LegalPawnMoves(%d) = %+v, want %+v", tt.seat, got, want)
			}
		})
	}
}

func TestPlaceWall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *Quoridor
		wall    Wall
		wantErr error
	}{
		{
			"空いている場所には置ける",
			NewQuoridor,
			Wall{Orientation: WallHorizontal, Row: 4, Col: 4}, nil,
		},
		{
			"同じ位置・同じ向きは重なる",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallHorizontal, Row: 4, Col: 4}, ErrWallOverlaps,
		},
		{
			"横壁は 1 マスずれても半分重なる",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallHorizontal, Row: 4, Col: 5}, ErrWallOverlaps,
		},
		{
			"縦壁は 1 マスずれても半分重なる",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallVertical, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallVertical, Row: 5, Col: 4}, ErrWallOverlaps,
		},
		{
			"同じアンカーの縦横は十字に交差する",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallVertical, Row: 4, Col: 4}, ErrWallOverlaps,
		},
		{
			"2 マス離れた同じ向きは置ける",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallHorizontal, Row: 4, Col: 6}, nil,
		},
		{
			"隣のアンカーの縦横は交差しない",
			func() *Quoridor {
				q := NewQuoridor()
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 4, Col: 4}}
				return q
			},
			Wall{Orientation: WallVertical, Row: 4, Col: 5}, nil,
		},
		{
			"アンカーは 0..7 まで",
			NewQuoridor,
			Wall{Orientation: WallHorizontal, Row: 8, Col: 0}, ErrWallOutOfBoard,
		},
		{
			"向きは h か v だけ",
			NewQuoridor,
			Wall{Orientation: "x", Row: 4, Col: 4}, ErrInvalidWall,
		},
		{
			"持ち壁がなければ置けない",
			func() *Quoridor {
				q := NewQuoridor()
				q.WallsLeft[0] = 0
				return q
			},
			Wall{Orientation: WallHorizontal, Row: 4, Col: 4}, ErrNoWallsLeft,
		},
		{
			// (0,0) の駒を h(0,0) と v(0,1) で {(0,0),(0,1)} に閉じ込める。
			"ゴールへの経路を完全に塞ぐ壁は置けない",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns[0] = Cell{Row: 0, Col: 0}
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 0, Col: 0}}
				return q
			},
			Wall{Orientation: WallVertical, Row: 0, Col: 1}, ErrWallSealsGoal,
		},
		{
			// 経路が長くなるだけなら置ける。
			"遠回りになるだけの壁は置ける",
			func() *Quoridor {
				q := NewQuoridor()
				q.Pawns[0] = Cell{Row: 0, Col: 0}
				q.Walls = []Wall{{Orientation: WallHorizontal, Row: 0, Col: 0}}
				return q
			},
			Wall{Orientation: WallVertical, Row: 1, Col: 1}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			q := tt.setup()
			err := q.PlaceWall(0, tt.wall)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("PlaceWall(0, %+v) = %v, want %v", tt.wall, err, tt.wantErr)
			}
		})
	}
}

func TestPlaceWallUpdatesState(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()
	w := Wall{Orientation: WallVertical, Row: 3, Col: 3}
	if err := q.PlaceWall(0, w); err != nil {
		t.Fatalf("置けるはずの壁が拒否された: %v", err)
	}
	if q.WallsLeft[0] != WallsPerPlayer-1 {
		t.Errorf("持ち壁が減っていない: %d", q.WallsLeft[0])
	}
	if !slices.Contains(q.Walls, w) {
		t.Errorf("壁が盤に置かれていない: %+v", q.Walls)
	}
	if q.Turn != 1 {
		t.Errorf("手番が渡っていない: %d", q.Turn)
	}
	// 拒否された壁では状態が変わらないことも確認する。
	if err := q.PlaceWall(1, w); !errors.Is(err, ErrWallOverlaps) {
		t.Fatalf("重なる壁が通ってしまった: %v", err)
	}
	if q.WallsLeft[1] != WallsPerPlayer || len(q.Walls) != 1 {
		t.Error("拒否された壁で状態が変わった")
	}
}

func TestWinByGoal(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()
	q.Pawns[0] = Cell{Row: 7, Col: 4}
	q.Pawns[1] = Cell{Row: 8, Col: 0} // ゴールマスを空けておく

	if err := q.MovePawn(0, Cell{Row: 8, Col: 4}); err != nil {
		t.Fatalf("ゴールへの移動が拒否された: %v", err)
	}
	if !q.Finished() || q.Winner != 0 {
		t.Errorf("seat 0 の勝ちのはず: winner=%d", q.Winner)
	}
	// 決着後はどんな操作も受け付けない。
	if err := q.MovePawn(1, Cell{Row: 7, Col: 0}); !errors.Is(err, ErrGameFinished) {
		t.Errorf("決着後の移動が通った: %v", err)
	}
	if err := q.PlaceWall(1, Wall{Orientation: WallHorizontal, Row: 4, Col: 4}); !errors.Is(err, ErrGameFinished) {
		t.Errorf("決着後の壁配置が通った: %v", err)
	}
}

func TestResign(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()
	// 投了は手番に関係なくできる。
	if err := q.Resign(1); err != nil {
		t.Fatalf("投了が拒否された: %v", err)
	}
	if q.Winner != 0 {
		t.Errorf("投了した側の相手が勝つはず: winner=%d", q.Winner)
	}
	if err := q.Resign(0); !errors.Is(err, ErrGameFinished) {
		t.Errorf("決着後の投了が通った: %v", err)
	}
}

// 短い 1 局を最初から最後まで通す。
func TestFullGame(t *testing.T) {
	t.Parallel()

	q := NewQuoridor()
	type step struct {
		seat int
		to   Cell
	}
	// seat 0 が一直線に進み、seat 1 は横で足踏みする。
	steps := []step{
		{0, Cell{Row: 1, Col: 4}}, {1, Cell{Row: 8, Col: 3}},
		{0, Cell{Row: 2, Col: 4}}, {1, Cell{Row: 8, Col: 4}},
		{0, Cell{Row: 3, Col: 4}}, {1, Cell{Row: 8, Col: 3}},
		{0, Cell{Row: 4, Col: 4}}, {1, Cell{Row: 8, Col: 4}},
		{0, Cell{Row: 5, Col: 4}}, {1, Cell{Row: 8, Col: 3}},
		{0, Cell{Row: 6, Col: 4}}, {1, Cell{Row: 8, Col: 4}},
		{0, Cell{Row: 7, Col: 4}}, {1, Cell{Row: 8, Col: 3}},
		{0, Cell{Row: 8, Col: 4}},
	}
	for i, s := range steps {
		if err := q.MovePawn(s.seat, s.to); err != nil {
			t.Fatalf("%d 手目 %+v が拒否された: %v", i+1, s, err)
		}
	}
	if q.Winner != 0 {
		t.Errorf("seat 0 が勝つはず: winner=%d", q.Winner)
	}
	if q.MoveCount != len(steps) {
		t.Errorf("手数が合わない: %d != %d", q.MoveCount, len(steps))
	}
}
