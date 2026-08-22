"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";

import { wsUrl } from "@/lib/api";

/** backend の handler/game.go の wsGameState と対。 */
export interface GameCell {
  row: number;
  col: number;
}

export interface GameWall {
  orientation: "h" | "v";
  row: number;
  col: number;
}

export interface GamePlayer {
  id: string;
  displayName: string;
  handle: string;
  rating: number;
}

export interface GameState {
  matchId: string;
  mode: string;
  seat: number;
  version: number;
  turn: number;
  pawns: [GameCell, GameCell];
  wallsLeft: [number, number];
  walls: GameWall[];
  moveCount: number;
  legalMoves: GameCell[];
  finished: boolean;
  winner?: number;
  resultType?: string;
  /** 決着後の新レーティング。中断では付かない。 */
  ratingAfter?: [number, number];
  turnDeadline: string;
  players: [GamePlayer, GamePlayer];
  connected: [boolean, boolean];
  /** いま観戦している人数。 */
  spectators: number;
}

type ServerMessage =
  | { type: "queued" }
  | { type: "state"; state: GameState }
  | { type: "opponent_disconnected"; graceSeconds: number }
  | { type: "opponent_reconnected" }
  | { type: "error"; code: string; message: string };

export type ConnectionStatus = "connecting" | "online" | "reconnecting";

const maxReconnectDelayMs = 10_000;

/** 対局 WebSocket の接続・再接続・メッセージ処理をまとめたフック。
 *
 * 盤面はサーバーが毎回全量を送ってくる（サーバー権威型）ので、
 * このフックは「最後に届いた state を持つだけ」でよく、差分の適用はしない。
 * 切断したら指数バックオフで繋ぎ直す。進行中の対局があれば
 * サーバー側が接続時に自動で復帰させ、最新の盤面を送ってくる。
 */
export function useGameSocket() {
  // サーバーはエラーをコードで返し、文言はクライアント側の言語で組み立てる。
  const t = useTranslations("game.wsErrors");
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const [queued, setQueued] = useState(false);
  const [state, setState] = useState<GameState | null>(null);
  const [opponentGrace, setOpponentGrace] = useState<number | null>(null);
  const [lastError, setLastError] = useState<string | null>(null);

  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let stopped = false;
    let attempts = 0;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    const connect = () => {
      const socket = new WebSocket(wsUrl("/game/ws"));
      socketRef.current = socket;

      socket.onopen = () => {
        attempts = 0;
        setStatus("online");
      };

      socket.onmessage = (event) => {
        const msg: ServerMessage = JSON.parse(event.data as string);
        switch (msg.type) {
          case "queued":
            // 新しい対戦に並び直したときは前の対局の表示を片付ける。
            setQueued(true);
            setState(null);
            setOpponentGrace(null);
            break;
          case "state": {
            setQueued(false);
            setState(msg.state);
            // 相手が繋がっている、または決着したら切断表示は要らない。
            // 観戦（seat -1）では相手という概念が無いので常に消す。
            const oppSeat = msg.state.seat >= 0 ? 1 - msg.state.seat : -1;
            if (
              msg.state.finished ||
              oppSeat < 0 ||
              msg.state.connected[oppSeat]
            ) {
              setOpponentGrace(null);
            }
            break;
          }
          case "opponent_disconnected":
            setOpponentGrace(msg.graceSeconds);
            break;
          case "opponent_reconnected":
            setOpponentGrace(null);
            break;
          case "error":
            setLastError(t.has(msg.code) ? t(msg.code) : msg.message);
            break;
        }
      };

      socket.onclose = () => {
        if (stopped) return;
        setStatus("reconnecting");
        const delay = Math.min(1000 * 2 ** attempts, maxReconnectDelayMs);
        attempts += 1;
        retryTimer = setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      stopped = true;
      clearTimeout(retryTimer);
      socketRef.current?.close();
    };
    // t は接続の張り直しを起こしたくないので依存に入れない（文言は次のエラーから反映される）。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const send = useCallback(
    (payload: Record<string, unknown>) => {
      const socket = socketRef.current;
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        setLastError(t("not_connected"));
        return;
      }
      socket.send(JSON.stringify(payload));
    },
    [t],
  );

  const joinQueue = useCallback(() => {
    setLastError(null);
    send({ type: "join_queue" });
  }, [send]);

  const leaveQueue = useCallback(() => {
    setQueued(false);
    send({ type: "leave_queue" });
  }, [send]);

  /** 操作には毎回 actionId（冪等キー）と、いま見ている盤面の版数を添える。
   * 再送されても二重適用されず、古い盤面前提の操作はサーバーが拒否する。
   */
  const act = useCallback(
    (action: Record<string, unknown>) => {
      setLastError(null);
      send({
        type: "action",
        actionId: crypto.randomUUID(),
        expectedVersion: state?.version ?? 0,
        ...action,
      });
    },
    [send, state?.version],
  );

  const movePawn = useCallback(
    (cell: GameCell) => act({ action: "move", row: cell.row, col: cell.col }),
    [act],
  );

  const placeWall = useCallback(
    (wall: GameWall) =>
      act({
        action: "wall",
        orientation: wall.orientation,
        row: wall.row,
        col: wall.col,
      }),
    [act],
  );

  const resign = useCallback(() => act({ action: "resign" }), [act]);

  /** 進行中の対局を観戦する。サーバーが最新盤面を送り返してくる。 */
  const watch = useCallback(
    (matchId: string) => {
      setLastError(null);
      send({ type: "watch", matchId });
    },
    [send],
  );

  /** 観戦をやめる。盤面表示も手元で片付ける。 */
  const unwatch = useCallback(() => {
    send({ type: "unwatch" });
    setState(null);
    setOpponentGrace(null);
  }, [send]);

  return {
    status,
    queued,
    state,
    opponentGrace,
    lastError,
    joinQueue,
    leaveQueue,
    movePawn,
    placeWall,
    resign,
    watch,
    unwatch,
  };
}
