"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  Group,
  Loader,
  Paper,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import {
  IconAlertTriangle,
  IconPlugOff,
  IconSearch,
} from "@tabler/icons-react";

import { QuoridorBoard } from "@/components/quoridor-board";
import {
  type GamePlayer,
  type GameState,
  useGameSocket,
} from "@/components/use-game-socket";

/** 決着の見出し。自分視点の勝敗と理由を 1 行で。 */
function resultText(state: GameState): string {
  if (state.winner === undefined) return "対局は中断されました";
  const won = state.winner === state.seat;
  const reason = {
    goal: "ゴール到達",
    resign: "投了",
    timeout: "時間切れ",
  }[state.resultType ?? ""];
  return `${won ? "勝ち" : "負け"}${reason ? `（${reason}）` : ""}`;
}

/** 残り持ち時間の表示。サーバーの締切時刻から手元で数えるだけ。 */
function useCountdown(deadline: string | undefined, running: boolean) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!running) return;
    const timer = setInterval(() => setNow(Date.now()), 500);
    return () => clearInterval(timer);
  }, [running]);
  if (!deadline || !running) return null;
  return Math.max(0, Math.ceil((new Date(deadline).getTime() - now) / 1000));
}

/** 残り時間の表示。60 秒制なので 0:47 の形にする。 */
function formatSeconds(total: number): string {
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

/** 1 人ぶんのカード。手番のプレイヤーには枠の強調と持ち時間が付く。
 * 「あなたの手番です」のような文は置かず、時計の位置で手番を伝える。
 */
function PlayerPanel({
  player,
  wallsLeft,
  isMe,
  connected,
  secondsLeft,
}: {
  player: GamePlayer;
  wallsLeft: number;
  isMe: boolean;
  connected: boolean;
  /** 手番でなければ null。 */
  secondsLeft: number | null;
}) {
  const isTurn = secondsLeft !== null;
  return (
    <Paper p="sm" className={isTurn ? "border-emerald-500" : undefined}>
      <Group justify="space-between" align="center">
        <div>
          <Group gap="xs">
            <Text fw={600}>{player.displayName}</Text>
            {isMe && (
              <Badge size="xs" variant="light">
                自分
              </Badge>
            )}
            {!connected && (
              <Badge size="xs" color="red" variant="light">
                切断中
              </Badge>
            )}
          </Group>
          <Text size="xs" c="dimmed">
            @{player.handle}
          </Text>
        </div>
        <Group gap="sm" align="center">
          <Text size="sm" c="dimmed">
            壁 {wallsLeft}
          </Text>
          {isTurn && (
            <Text
              ff="monospace"
              fw={700}
              size="lg"
              c={secondsLeft <= 10 ? "red" : undefined}
              className="rounded-sm bg-dark-500 px-2 py-0.5"
            >
              {formatSeconds(secondsLeft)}
            </Text>
          )}
        </Group>
      </Group>
    </Paper>
  );
}

export function GameScreen() {
  const {
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
  } = useGameSocket();

  const mySeat = state?.seat ?? 0;
  const myTurn = state !== null && !state.finished && state.turn === mySeat;
  const secondsLeft = useCountdown(
    state?.turnDeadline,
    state !== null && !state.finished,
  );

  return (
    <Stack gap="lg" maw={900} mx="auto">
      <Group justify="space-between">
        <Title order={2} size="h3">
          クイックマッチ
        </Title>
        {status !== "online" && (
          <Badge
            color="yellow"
            variant="light"
            leftSection={<IconPlugOff size={12} />}
          >
            {status === "connecting" ? "接続中…" : "再接続中…"}
          </Badge>
        )}
      </Group>

      {lastError && (
        <Alert
          color="yellow"
          icon={<IconAlertTriangle size={16} />}
          variant="light"
        >
          {lastError}
        </Alert>
      )}

      {opponentGrace !== null && state !== null && !state.finished && (
        <Alert color="red" icon={<IconPlugOff size={16} />} variant="light">
          相手の接続が切れました。{opponentGrace}
          秒以内に戻らなければあなたの勝ちになります。
        </Alert>
      )}

      {state === null && !queued && (
        <Paper p="xl">
          <Stack align="center" gap="md">
            <Text c="dimmed">
              待機列に入ると、相手が見つかり次第すぐに対局が始まります。
            </Text>
            <Button
              size="md"
              leftSection={<IconSearch size={18} />}
              onClick={joinQueue}
              disabled={status !== "online"}
            >
              対戦相手を探す
            </Button>
          </Stack>
        </Paper>
      )}

      {state === null && queued && (
        <Paper p="xl">
          <Stack align="center" gap="md">
            <Loader />
            <Text c="dimmed">対戦相手を探しています…</Text>
            <Button variant="subtle" onClick={leaveQueue}>
              やめる
            </Button>
          </Stack>
        </Paper>
      )}

      {state !== null && (
        <div className="flex flex-col gap-6 lg:flex-row">
          <Stack gap="xs" className="w-full max-w-[560px]">
            <QuoridorBoard
              state={state}
              interactive={myTurn && status === "online"}
              onMove={movePawn}
              onWall={placeWall}
            />
            <Text
              size="xs"
              c="dimmed"
              ta="center"
              className={myTurn && !state.finished ? undefined : "invisible"}
            >
              移動: 光るマス ／ 壁: マスの間をクリック
            </Text>
          </Stack>

          <Stack gap="sm" className="lg:w-72">
            <PlayerPanel
              player={state.players[1 - mySeat]}
              wallsLeft={state.wallsLeft[1 - mySeat]}
              isMe={false}
              connected={state.connected[1 - mySeat]}
              secondsLeft={
                !state.finished && state.turn !== mySeat ? secondsLeft : null
              }
            />
            <PlayerPanel
              player={state.players[mySeat]}
              wallsLeft={state.wallsLeft[mySeat]}
              isMe
              connected
              secondsLeft={myTurn ? secondsLeft : null}
            />

            {state.finished ? (
              <Paper p="md">
                <Stack gap="sm">
                  <Title order={3} size="h4">
                    {resultText(state)}
                  </Title>
                  <Text size="sm" c="dimmed">
                    総手数 {state.moveCount}
                  </Text>
                  {state.ratingAfter && (
                    <Text size="sm">
                      レート {state.players[mySeat].rating} →{" "}
                      <Text
                        span
                        fw={700}
                        c={
                          state.ratingAfter[mySeat] >=
                          state.players[mySeat].rating
                            ? "emerald"
                            : "red"
                        }
                      >
                        {state.ratingAfter[mySeat]}
                      </Text>
                    </Text>
                  )}
                  <Button component={Link} href="/">
                    ホームへ戻る
                  </Button>
                </Stack>
              </Paper>
            ) : (
              <Button color="red" variant="light" onClick={resign}>
                投了する
              </Button>
            )}
          </Stack>
        </div>
      )}
    </Stack>
  );
}
