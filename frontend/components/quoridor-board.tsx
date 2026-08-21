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
 * 盤の向きは固定（seat 0 が上から下へ、seat 1 が下から上へ）。
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
  const [hoverWall, setHoverWall] = useState<GameWall | null>(null);

  const legalMoves = interactive ? state.legalMoves : [];
  const isLegal = (row: number, col: number) =>
    legalMoves.some((c) => c.row === row && c.col === col);

  const canPlaceWall = interactive && state.wallsLeft[state.seat] > 0;

  const cells = [];
  for (let row = 0; row < 9; row++) {
    for (let col = 0; col < 9; col++) {
      const pawnSeat = state.pawns.findIndex(
        (p) => p.row === row && p.col === col,
      );
      const legal = isLegal(row, col);
      cells.push(
        <button
          key={`cell-${row}-${col}`}
          type="button"
          disabled={!legal}
          onClick={() => onMove({ row, col })}
          aria-label={`マス ${row}-${col}`}
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
  const walls = state.walls.map((w) => (
    <div
      key={`wall-${w.orientation}-${w.row}-${w.col}`}
      className="rounded-full bg-[var(--mantine-color-yellow-6)]"
      style={
        w.orientation === "h"
          ? {
              gridRow: gutterTrack(w.row),
              gridColumn: `${cellTrack(w.col)} / span 3`,
            }
          : {
              gridColumn: gutterTrack(w.col),
              gridRow: `${cellTrack(w.row)} / span 3`,
            }
      }
    />
  ));

  // 壁を置ける溝のクリック領域。ホバーで置き先のプレビューを出す。
  const gutters = [];
  if (canPlaceWall) {
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 9; col++) {
        const wall: GameWall = { orientation: "h", row, col: clampAnchor(col) };
        gutters.push(
          <button
            key={`gh-${row}-${col}`}
            type="button"
            aria-label={`横壁 ${wall.row}-${wall.col}`}
            onClick={() => onWall(wall)}
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
        gutters.push(
          <button
            key={`gv-${row}-${col}`}
            type="button"
            aria-label={`縦壁 ${wall.row}-${wall.col}`}
            onClick={() => onWall(wall)}
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
