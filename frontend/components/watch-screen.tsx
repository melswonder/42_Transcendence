"use client";

import { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  Group,
  Paper,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import {
  IconAlertTriangle,
  IconArrowLeft,
  IconEye,
  IconPlugOff,
  IconRefresh,
} from "@tabler/icons-react";

import { PlayerPanel, useCountdown } from "@/components/player-panel";
import { QuoridorBoard } from "@/components/quoridor-board";
import { type GamePlayer, useGameSocket } from "@/components/use-game-socket";
import { apiUrl } from "@/lib/api";

/** backend の handler/game.go liveMatchResponse と対。 */
interface LiveMatch {
  match_id: string;
  mode: string;
  started_at: string;
  players: [GamePlayer, GamePlayer];
  move_count: number;
  spectators: number;
}

export function WatchScreen() {
  const { status, state, lastError, watch, unwatch } = useGameSocket();

  const [matches, setMatches] = useState<LiveMatch[] | null>(null);
  const [listError, setListError] = useState<string | null>(null);

  const secondsLeft = useCountdown(
    state?.turnDeadline,
    state !== null && !state.finished,
  );

  const reload = useCallback(async () => {
    try {
      const res = await fetch(apiUrl("/game/live"), {
        credentials: "include",
      });
      if (!res.ok) throw new Error(`一覧の取得に失敗しました (${res.status})`);
      const body = (await res.json()) as { items: LiveMatch[] };
      setMatches(body.items);
      setListError(null);
    } catch (e) {
      setListError(e instanceof Error ? e.message : "読み込みに失敗しました");
    }
  }, []);

  // 観戦していない間は一覧を定期的に引き直す。
  const watching = state !== null;
  useEffect(() => {
    if (watching) return;
    const load = () => void reload();
    const kickoff = setTimeout(load, 0);
    const timer = setInterval(load, 10_000);
    return () => {
      clearTimeout(kickoff);
      clearInterval(timer);
    };
  }, [watching, reload]);

  const backToList = () => {
    unwatch();
    void reload();
  };

  // ---- 観戦中 ----
  if (state !== null) {
    return (
      <Stack gap="lg" maw={900} mx="auto">
        <Group justify="space-between">
          <Group gap="sm">
            <Button
              variant="subtle"
              size="compact-sm"
              leftSection={<IconArrowLeft size={16} />}
              onClick={backToList}
            >
              一覧へ戻る
            </Button>
            <Title order={2} size="h3">
              観戦中
            </Title>
          </Group>
          <Group gap="xs">
            <Badge variant="light" leftSection={<IconEye size={12} />}>
              観戦 {state.spectators} 人
            </Badge>
            {status !== "online" && (
              <Badge
                color="yellow"
                variant="light"
                leftSection={<IconPlugOff size={12} />}
              >
                再接続中…
              </Badge>
            )}
          </Group>
        </Group>

        <div className="flex flex-col gap-6 lg:flex-row">
          {/* 観戦者は操作できない。盤は表示専用。 */}
          <QuoridorBoard
            state={state}
            interactive={false}
            onMove={() => {}}
            onWall={() => {}}
          />

          <Stack gap="sm" className="lg:w-72">
            {([1, 0] as const).map((seat) => (
              <PlayerPanel
                key={seat}
                player={state.players[seat]}
                wallsLeft={state.wallsLeft[seat]}
                isMe={false}
                connected={state.connected[seat]}
                secondsLeft={
                  !state.finished && state.turn === seat ? secondsLeft : null
                }
              />
            ))}

            {state.finished && (
              <Paper p="md">
                <Stack gap="sm">
                  <Title order={3} size="h4">
                    {state.winner !== undefined
                      ? `${state.players[state.winner].displayName} の勝ち`
                      : "対局は中断されました"}
                  </Title>
                  <Text size="sm" c="dimmed">
                    総手数 {state.moveCount}
                  </Text>
                  <Button onClick={backToList}>一覧へ戻る</Button>
                </Stack>
              </Paper>
            )}
          </Stack>
        </div>
      </Stack>
    );
  }

  // ---- 一覧 ----
  return (
    <Stack gap="lg" maw={720} mx="auto">
      <Group justify="space-between">
        <Title order={2} size="h3">
          観戦
        </Title>
        <Button
          variant="subtle"
          size="compact-sm"
          leftSection={<IconRefresh size={16} />}
          onClick={() => void reload()}
        >
          更新
        </Button>
      </Group>

      {(lastError ?? listError) && (
        <Alert
          color="yellow"
          icon={<IconAlertTriangle size={16} />}
          variant="light"
        >
          {lastError ?? listError}
        </Alert>
      )}

      {matches !== null && matches.length === 0 && (
        <Paper p="xl">
          <Text c="dimmed" ta="center">
            いま進行中の対局はありません。
          </Text>
        </Paper>
      )}

      {matches?.map((m) => (
        <Paper
          key={m.match_id}
          component="button"
          type="button"
          p="md"
          onClick={() => watch(m.match_id)}
          disabled={status !== "online"}
          className="text-left transition-colors enabled:hover:border-emerald-500"
        >
          <Group justify="space-between" wrap="nowrap">
            <Group gap="xs" wrap="nowrap" className="min-w-0">
              <Text fw={600} truncate>
                {m.players[0].displayName}
              </Text>
              <Text c="dimmed" size="sm">
                vs
              </Text>
              <Text fw={600} truncate>
                {m.players[1].displayName}
              </Text>
            </Group>
            <Group gap="sm" wrap="nowrap">
              <Text size="sm" c="dimmed">
                {m.move_count} 手目
              </Text>
              <Badge
                variant="light"
                leftSection={<IconEye size={12} />}
                color={m.spectators > 0 ? "emerald" : "gray"}
              >
                {m.spectators}
              </Badge>
            </Group>
          </Group>
        </Paper>
      ))}
    </Stack>
  );
}
