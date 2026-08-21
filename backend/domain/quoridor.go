package domain

import (
	"errors"
	"slices"
)

// コリドールの盤の大きさ。マスは 9x9、壁のアンカー（交点）は 8x8。
const (
	BoardSize      = 9
	WallGridSize   = BoardSize - 1
	WallsPerPlayer = 10

	// 1 局の人数。座席は 0（先手）と 1（後手）。
	quoridorSeats = 2
)

// 壁の向き。match_actions のペイロードにもこの値をそのまま入れる。
const (
	WallHorizontal = "h" // 横壁: 上下の移動を塞ぐ
	WallVertical   = "v" // 縦壁: 左右の移動を塞ぐ
)

// コリドールのルール違反。usecase は errors.Is で分岐して
// 「不正入力（クライアントに 4xx 相当で返す）」として扱う。
var (
	ErrNotYourTurn     = errors.New("not your turn")
	ErrGameFinished    = errors.New("game already finished")
	ErrCellOutOfBoard  = errors.New("cell out of board")
	ErrIllegalPawnMove = errors.New("illegal pawn move")
	ErrWallOutOfBoard  = errors.New("wall out of board")
	ErrInvalidWall     = errors.New("invalid wall orientation")
	ErrNoWallsLeft     = errors.New("no walls left")
	ErrWallOverlaps    = errors.New("wall overlaps or crosses an existing wall")
	ErrWallSealsGoal   = errors.New("wall would seal a player's path to goal")
	ErrInvalidSeat     = errors.New("invalid seat")
)

// Cell は盤上の 1 マス。Row / Col とも 0..8。
type Cell struct {
	Row int
	Col int
}

func (c Cell) inBoard() bool {
	return c.Row >= 0 && c.Row < BoardSize && c.Col >= 0 && c.Col < BoardSize
}

// Wall は 1 枚の壁。アンカー (Row, Col) は左上の交点で 0..7。
// 横壁はアンカーの右方向 2 マスぶんの上下移動を塞ぎ、
// 縦壁はアンカーの下方向 2 マスぶんの左右移動を塞ぐ。
type Wall struct {
	Orientation string
	Row         int
	Col         int
}

func (w Wall) inBoard() bool {
	return w.Row >= 0 && w.Row < WallGridSize && w.Col >= 0 && w.Col < WallGridSize
}

// Quoridor は進行中の 1 局。サーバーだけがこれを書き換える（サーバー権威型）。
// seat 0 は (0,4) から Row 8 を、seat 1 は (8,4) から Row 0 を目指す。
type Quoridor struct {
	Pawns     [quoridorSeats]Cell
	WallsLeft [quoridorSeats]int
	Walls     []Wall
	Turn      int // 次に指す座席
	MoveCount int // 駒移動と壁配置の合計手数
	Winner    int // 勝った座席。未決着なら -1
}

// NewQuoridor は初期配置の局面を返す。先手は seat 0。
func NewQuoridor() *Quoridor {
	return &Quoridor{
		Pawns:     [quoridorSeats]Cell{{Row: 0, Col: 4}, {Row: BoardSize - 1, Col: 4}},
		WallsLeft: [quoridorSeats]int{WallsPerPlayer, WallsPerPlayer},
		Turn:      0,
		Winner:    -1,
	}
}

// GoalRow は seat が到達すべき行。
func GoalRow(seat int) int {
	if seat == 0 {
		return BoardSize - 1
	}
	return 0
}

// Finished は決着済みかどうか。
func (q *Quoridor) Finished() bool {
	return q.Winner >= 0
}

// MovePawn は seat の駒を to へ動かす。合法手でなければ局面を変えずにエラーを返す。
func (q *Quoridor) MovePawn(seat int, to Cell) error {
	if err := q.checkTurn(seat); err != nil {
		return err
	}
	if !to.inBoard() {
		return ErrCellOutOfBoard
	}
	if !slices.Contains(q.LegalPawnMoves(seat), to) {
		return ErrIllegalPawnMove
	}

	q.Pawns[seat] = to
	q.MoveCount++
	if to.Row == GoalRow(seat) {
		q.Winner = seat
	} else {
		q.Turn = 1 - seat
	}
	return nil
}

// PlaceWall は seat が壁を置く。置けない壁なら局面を変えずにエラーを返す。
func (q *Quoridor) PlaceWall(seat int, w Wall) error {
	if err := q.checkTurn(seat); err != nil {
		return err
	}
	if q.WallsLeft[seat] <= 0 {
		return ErrNoWallsLeft
	}
	if err := q.validateWall(w); err != nil {
		return err
	}

	q.Walls = append(q.Walls, w)
	q.WallsLeft[seat]--
	q.MoveCount++
	q.Turn = 1 - seat
	return nil
}

// Resign は seat の投了。相手の勝ちで決着する。
func (q *Quoridor) Resign(seat int) error {
	if seat != 0 && seat != 1 {
		return ErrInvalidSeat
	}
	if q.Finished() {
		return ErrGameFinished
	}
	q.Winner = 1 - seat
	return nil
}

// LegalPawnMoves は seat の駒が動ける先の一覧。
// 隣接マスへの移動に加え、相手駒への正面ジャンプと、
// ジャンプ先が壁や盤端で塞がれている場合のサイドステップを含む。
func (q *Quoridor) LegalPawnMoves(seat int) []Cell {
	if seat != 0 && seat != 1 {
		return nil
	}
	me := q.Pawns[seat]
	opp := q.Pawns[1-seat]

	var moves []Cell
	for _, d := range directions() {
		next := Cell{Row: me.Row + d.Row, Col: me.Col + d.Col}
		if !next.inBoard() || q.blocked(me, next) {
			continue
		}
		if next != opp {
			moves = append(moves, next)
			continue
		}
		// 相手駒の上には乗れない。まず正面ジャンプを試す。
		jump := Cell{Row: next.Row + d.Row, Col: next.Col + d.Col}
		if jump.inBoard() && !q.blocked(next, jump) {
			moves = append(moves, jump)
			continue
		}
		// ジャンプできないときだけ、相手駒の左右へサイドステップできる。
		for _, p := range perpendicular(d) {
			side := Cell{Row: next.Row + p.Row, Col: next.Col + p.Col}
			if side.inBoard() && !q.blocked(next, side) {
				moves = append(moves, side)
			}
		}
	}
	return moves
}

func (q *Quoridor) checkTurn(seat int) error {
	if seat != 0 && seat != 1 {
		return ErrInvalidSeat
	}
	if q.Finished() {
		return ErrGameFinished
	}
	if q.Turn != seat {
		return ErrNotYourTurn
	}
	return nil
}

// validateWall は壁 1 枚の配置可否を見る。手番や残り枚数は呼び出し側で確認済み。
func (q *Quoridor) validateWall(w Wall) error {
	if w.Orientation != WallHorizontal && w.Orientation != WallVertical {
		return ErrInvalidWall
	}
	if !w.inBoard() {
		return ErrWallOutOfBoard
	}
	for _, e := range q.Walls {
		if wallsConflict(w, e) {
			return ErrWallOverlaps
		}
	}
	// どちらのプレイヤーのゴール経路も完全に塞いではいけない（公式ルール）。
	trial := append(slices.Clone(q.Walls), w)
	for seat := range quoridorSeats {
		if !hasPathToGoal(q.Pawns[seat], GoalRow(seat), trial) {
			return ErrWallSealsGoal
		}
	}
	return nil
}

// wallsConflict は 2 枚の壁が重なる・交差するかどうか。
func wallsConflict(a, b Wall) bool {
	if a.Orientation == b.Orientation {
		if a.Row == b.Row && a.Col == b.Col {
			return true // 完全に同じ位置
		}
		// 同じ向きで 1 マスずれた壁は半分重なる。
		if a.Orientation == WallHorizontal {
			return a.Row == b.Row && absInt(a.Col-b.Col) == 1
		}
		return a.Col == b.Col && absInt(a.Row-b.Row) == 1
	}
	// 向きが違う壁は、同じアンカーを共有すると十字に交差する。
	return a.Row == b.Row && a.Col == b.Col
}

// blocked は隣接する 2 マスの間が壁で塞がれているか。
func (q *Quoridor) blocked(from, to Cell) bool {
	return blockedByWalls(from, to, q.Walls)
}

func blockedByWalls(from, to Cell, walls []Wall) bool {
	for _, w := range walls {
		if wallBlocks(w, from, to) {
			return true
		}
	}
	return false
}

func wallBlocks(w Wall, from, to Cell) bool {
	// 上下の移動は横壁が塞ぐ。
	if from.Col == to.Col && absInt(from.Row-to.Row) == 1 {
		if w.Orientation != WallHorizontal {
			return false
		}
		return w.Row == min(from.Row, to.Row) && (w.Col == from.Col || w.Col == from.Col-1)
	}
	// 左右の移動は縦壁が塞ぐ。
	if from.Row == to.Row && absInt(from.Col-to.Col) == 1 {
		if w.Orientation != WallVertical {
			return false
		}
		return w.Col == min(from.Col, to.Col) && (w.Row == from.Row || w.Row == from.Row-1)
	}
	return false
}

// hasPathToGoal は from から goalRow のどこかへ壁を避けて到達できるかを幅優先探索で調べる。
func hasPathToGoal(from Cell, goalRow int, walls []Wall) bool {
	var visited [BoardSize][BoardSize]bool
	queue := []Cell{from}
	visited[from.Row][from.Col] = true

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.Row == goalRow {
			return true
		}
		for _, d := range directions() {
			next := Cell{Row: cur.Row + d.Row, Col: cur.Col + d.Col}
			if !next.inBoard() || visited[next.Row][next.Col] {
				continue
			}
			if blockedByWalls(cur, next, walls) {
				continue
			}
			visited[next.Row][next.Col] = true
			queue = append(queue, next)
		}
	}
	return false
}

func directions() []Cell {
	return []Cell{{Row: -1}, {Row: 1}, {Col: -1}, {Col: 1}}
}

func perpendicular(d Cell) []Cell {
	if d.Row != 0 {
		return []Cell{{Col: -1}, {Col: 1}}
	}
	return []Cell{{Row: -1}, {Row: 1}}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
