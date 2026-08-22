"use client";

import Link from "next/link";
import { useState } from "react";
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
  IconEye,
  IconPlugOff,
  IconSearch,
} from "@tabler/icons-react";

import { useTranslations } from "next-intl";

import { AppModal } from "@/components/app-modal";
import { ConfirmModal } from "@/components/confirm-modal";
import { PlayerPanel, useCountdown } from "@/components/player-panel";
import { QuoridorBoard } from "@/components/quoridor-board";
import { type GameState, useGameSocket } from "@/components/use-game-socket";

/** 決着の見出し。自分視点の勝敗と理由を 1 行で。 */
function resultText(
  t: ReturnType<typeof useTranslations<"game">>,
  state: GameState,
): string {
  if (state.winner === undefined) return t("result.aborted");
  const result = t(state.winner === state.seat ? "result.win" : "result.loss");
  const reasonKey = state.resultType;
  if (
    reasonKey === "goal" ||
    reasonKey === "resign" ||
    reasonKey === "timeout"
  ) {
    return t("resultWithReason", { result, reason: t(`result.${reasonKey}`) });
  }
  return result;
}

export function GameScreen() {
  const t = useTranslations("game");
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

  // 勝敗モーダル。決着したら開き、閉じたら盤面を見返せる。
  // 「どの対局で閉じたか」を持つので、次の対局では自然にまた開く。
  const [dismissedFor, setDismissedFor] = useState<string | null>(null);
  const resultDismissed = dismissedFor === state?.matchId;

  // 投了は取り消せないので、確認を一枚挟む。
  const [resignOpen, setResignOpen] = useState(false);

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
          {t("title")}
        </Title>
        {state !== null && state.spectators > 0 && (
          <Badge variant="light" leftSection={<IconEye size={12} />}>
            {t("spectatorBadge", { count: state.spectators })}
          </Badge>
        )}
        {status !== "online" && (
          <Badge
            color="yellow"
            variant="light"
            leftSection={<IconPlugOff size={12} />}
          >
            {status === "connecting" ? t("connecting") : t("reconnecting")}
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
          {t("opponentDisconnected", { seconds: opponentGrace })}
        </Alert>
      )}

      {state === null && !queued && (
        <Paper p="xl">
          <Stack align="center" gap="md">
            <Text c="dimmed">{t("queueHint")}</Text>
            <Button
              size="md"
              leftSection={<IconSearch size={18} />}
              onClick={joinQueue}
              disabled={status !== "online"}
            >
              {t("findOpponent")}
            </Button>
          </Stack>
        </Paper>
      )}

      {state === null && queued && (
        <Paper p="xl">
          <Stack align="center" gap="md">
            <Loader />
            <Text c="dimmed">{t("searching")}</Text>
            <Button variant="subtle" onClick={leaveQueue}>
              {t("cancel")}
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
              {t("boardHint")}
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
              <Stack gap="sm">
                <Button onClick={() => setDismissedFor(null)}>
                  {t("viewResult")}
                </Button>
                <Button component={Link} href="/" variant="subtle">
                  {t("backHome")}
                </Button>
              </Stack>
            ) : (
              <Button
                color="red"
                variant="light"
                onClick={() => setResignOpen(true)}
              >
                {t("resign")}
              </Button>
            )}
          </Stack>
        </div>
      )}

      {/* 勝敗モーダル。共通モーダルの中身を差し替えて使う。 */}
      {state !== null && (
        <AppModal
          opened={state.finished && !resultDismissed}
          onClose={() => setDismissedFor(state.matchId)}
          title={resultText(t, state)}
        >
          <Stack gap="md">
            <Text size="sm" c="dimmed">
              {t("totalMoves", { count: state.moveCount })}
            </Text>
            {state.ratingAfter && (
              <Text>
                {t("ratingChange", { before: state.players[mySeat].rating })}{" "}
                <Text
                  span
                  fw={700}
                  c={
                    state.ratingAfter[mySeat] >= state.players[mySeat].rating
                      ? "emerald"
                      : "red"
                  }
                >
                  {state.ratingAfter[mySeat]}
                </Text>
              </Text>
            )}
            <Group justify="flex-end" gap="sm">
              <Button
                variant="default"
                onClick={() => setDismissedFor(state.matchId)}
              >
                {t("viewBoard")}
              </Button>
              <Button component={Link} href="/">
                {t("backHome")}
              </Button>
            </Group>
          </Stack>
        </AppModal>
      )}

      {/* 投了の確認。 */}
      <ConfirmModal
        opened={resignOpen}
        onClose={() => setResignOpen(false)}
        onConfirm={() => {
          resign();
          setResignOpen(false);
        }}
        title={t("resignConfirmTitle")}
        message={t("resignConfirmMessage")}
        confirmLabel={t("resign")}
        color="red"
      />
    </Stack>
  );
}
