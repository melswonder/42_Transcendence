"use client";

import { useState } from "react";

import type {
  GameCell,
  GameState,
  GameWall,
} from "@/components/use-game-socket";

interface QuoridorBoardProps {
  state: GameState;
  /** 自分の手番で操作できるときだけ true。 */
  interactive: boolean;
  onMove: (cell: GameCell) => void;
  onWall: (wall: GameWall) => void;
}

/** 9x9 のマスと 8x8 の壁アンカーを 17 トラックの CSS グリッドで表す。
 * 奇数トラック（1,3,...,17）がマス、偶数トラックが壁の溝。
 *
 * 盤の座標（サーバーと同じ）と表示は分けて考える。seat 0 はそのままだと
 * 自分の駒が上から下へ進んで分かりにくいので、盤を 180° 回して
 * 「どちらのプレイヤーも自分が手前（下）から上へ進む」見え方に揃える。
 */
const cellTrack = (i: number) => i * 2 + 1;
const gutterTrack = (i: number) => i * 2 + 2;

/** 壁の溝をクリックしたときのアンカー。右端・下端の溝は 1 つ内側に寄せる。 */
const clampAnchor = (i: number) => Math.min(i, 7);

export function QuoridorBoard({
  state,
  interactive,
  onMove,
  onWall,
}: QuoridorBoardProps) {
  // hoverWall は表示座標で持つ（描画にしか使わないため）。
  const [hoverWall, setHoverWall] = useState<GameWall | null>(null);

  // seat 0 だけ盤を 180° 回す。変換は自身が逆変換にもなっている。
  const flipped = state.seat === 0;
  const toModelCell = (c: GameCell): GameCell =>
    flipped ? { row: 8 - c.row, col: 8 - c.col } : c;
  const toDisplayWall = (w: GameWall): GameWall =>
    flipped ? { ...w, row: 7 - w.row, col: 7 - w.col } : w;
  const toModelWall = toDisplayWall;

  const legalMoves = interactive ? state.legalMoves : [];
  const isLegal = (cell: GameCell) =>
    legalMoves.some((c) => c.row === cell.row && c.col === cell.col);

  const canPlaceWall = interactive && state.wallsLeft[state.seat] > 0;

  // 以降のループ変数はすべて表示座標。サーバーへ返すときだけ盤座標に戻す。
  const cells = [];
  for (let row = 0; row < 9; row++) {
    for (let col = 0; col < 9; col++) {
      const model = toModelCell({ row, col });
      const pawnSeat = state.pawns.findIndex(
        (p) => p.row === model.row && p.col === model.col,
      );
      const legal = isLegal(model);
      cells.push(
        <button
          key={`cell-${row}-${col}`}
          type="button"
          disabled={!legal}
          onClick={() => onMove(model)}
          aria-label={`マス ${model.row}-${model.col}`}
          style={{ gridRow: cellTrack(row), gridColumn: cellTrack(col) }}
          className={`flex aspect-square items-center justify-center rounded-sm ${
            legal
              ? "cursor-pointer bg-emerald-500/25 hover:bg-emerald-500/50"
              : "bg-dark-600"
          }`}
        >
          {pawnSeat >= 0 && (
            <span
              className={`block h-3/5 w-3/5 rounded-full ${
                pawnSeat === state.seat
                  ? "bg-emerald-500"
                  : "bg-[var(--mantine-color-red-6)]"
              } ${pawnSeat === state.turn && !state.finished ? "ring-2 ring-text/70" : ""}`}
            />
          )}
        </button>,
      );
    }
  }

  // 置かれている壁。横壁はマス 2 つと間の溝、縦壁も同様に 3 トラックへ跨がせる。
  const walls = state.walls.map((w) => {
    const d = toDisplayWall(w);
    return (
      <div
        key={`wall-${w.orientation}-${w.row}-${w.col}`}
        className="rounded-full bg-[var(--mantine-color-yellow-6)]"
        style={
          d.orientation === "h"
            ? {
                gridRow: gutterTrack(d.row),
                gridColumn: `${cellTrack(d.col)} / span 3`,
              }
            : {
                gridColumn: gutterTrack(d.col),
                gridRow: `${cellTrack(d.row)} / span 3`,
              }
        }
      />
    );
  });

  // 壁を置ける溝のクリック領域。ホバーで置き先のプレビューを出す。
  const gutters = [];
  if (canPlaceWall) {
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 9; col++) {
        const wall: GameWall = { orientation: "h", row, col: clampAnchor(col) };
        const model = toModelWall(wall);
        gutters.push(
          <button
            key={`gh-${row}-${col}`}
            type="button"
            aria-label={`横壁 ${model.row}-${model.col}`}
            onClick={() => onWall(model)}
            onMouseEnter={() => setHoverWall(wall)}
            onMouseLeave={() => setHoverWall(null)}
            style={{ gridRow: gutterTrack(row), gridColumn: cellTrack(col) }}
            className="cursor-pointer"
          />,
        );
      }
    }
    for (let row = 0; row < 9; row++) {
      for (let col = 0; col < 8; col++) {
        const wall: GameWall = { orientation: "v", row: clampAnchor(row), col };
        const model = toModelWall(wall);
        gutters.push(
          <button
            key={`gv-${row}-${col}`}
            type="button"
            aria-label={`縦壁 ${model.row}-${model.col}`}
            onClick={() => onWall(model)}
            onMouseEnter={() => setHoverWall(wall)}
            onMouseLeave={() => setHoverWall(null)}
            style={{ gridColumn: gutterTrack(col), gridRow: cellTrack(row) }}
            className="cursor-pointer"
          />,
        );
      }
    }
  }

  return (
    <div
      className="grid w-full max-w-[560px] rounded-md border border-default-border bg-dark-700 p-3"
      style={{
        gridTemplateColumns: "repeat(8, 1fr 8px) 1fr",
        gridTemplateRows: "repeat(8, 1fr 8px) 1fr",
        gap: 0,
      }}
    >
      {cells}
      {gutters}
      {hoverWall && (
        <div
          className="pointer-events-none rounded-full bg-[var(--mantine-color-yellow-6)]/40"
          style={
            hoverWall.orientation === "h"
              ? {
                  gridRow: gutterTrack(hoverWall.row),
                  gridColumn: `${cellTrack(hoverWall.col)} / span 3`,
                }
              : {
                  gridColumn: gutterTrack(hoverWall.col),
                  gridRow: `${cellTrack(hoverWall.row)} / span 3`,
                }
          }
        />
      )}
      {walls}
    </div>
  );
}
