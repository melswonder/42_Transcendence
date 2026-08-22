"use client";

import { useEffect, useState } from "react";
import { Badge, Group, Paper, Text } from "@mantine/core";

import type { GamePlayer } from "@/components/use-game-socket";

/** 残り時間の表示。60 秒制なので 0:47 の形にする。 */
export function formatSeconds(total: number): string {
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

/** 残り持ち時間。サーバーの締切時刻から手元で数えるだけ。 */
export function useCountdown(deadline: string | undefined, running: boolean) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!running) return;
    const timer = setInterval(() => setNow(Date.now()), 500);
    return () => clearInterval(timer);
  }, [running]);
  if (!deadline || !running) return null;
  return Math.max(0, Math.ceil((new Date(deadline).getTime() - now) / 1000));
}

/** 1 人ぶんのカード。手番のプレイヤーには枠の強調と持ち時間が付く。
 * 「あなたの手番です」のような文は置かず、時計の位置で手番を伝える。
 */
export function PlayerPanel({
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
