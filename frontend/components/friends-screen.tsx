"use client";

import { useCallback, useEffect, useState } from "react";
import {
  ActionIcon,
  Alert,
  Badge,
  Button,
  Group,
  Indicator,
  Paper,
  Stack,
  Tabs,
  Text,
  TextInput,
} from "@mantine/core";
import {
  IconAlertTriangle,
  IconCheck,
  IconSearch,
  IconUserMinus,
  IconUserPlus,
  IconX,
} from "@tabler/icons-react";

import { UserAvatar } from "@/components/user-avatar";
import { apiUrl } from "@/lib/api";

/** backend の handler/user.go userPublicResponse と対。 */
interface PublicUser {
  id: string;
  display_name: string;
  handle: string;
  avatar_url: string | null;
  level: number;
}

/** backend の handler/friend.go friendshipResponse と対。 */
interface Friendship {
  user: PublicUser;
  status: "pending" | "accepted" | "rejected";
  requested_by_me: boolean;
  online: boolean;
}

async function requestJSON(path: string, init?: RequestInit) {
  const res = await fetch(apiUrl(path), { credentials: "include", ...init });
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? `リクエストに失敗しました (${res.status})`);
  }
  return res.status === 204 ? null : res.json();
}

/** 名前・handle・オンライン表示つきの 1 行。 */
function UserRow({
  user,
  online,
  right,
}: {
  user: PublicUser;
  online?: boolean;
  right: React.ReactNode;
}) {
  return (
    <Paper p="sm">
      <Group justify="space-between" wrap="nowrap">
        <Group gap="sm" wrap="nowrap" className="min-w-0">
          <Indicator
            color="emerald"
            size={12}
            offset={4}
            position="bottom-end"
            withBorder
            disabled={online === undefined || !online}
          >
            <UserAvatar
              displayName={user.display_name}
              avatarUrl={user.avatar_url}
            />
          </Indicator>
          <Stack gap={0} className="min-w-0">
            <Group gap="xs">
              <Text size="sm" fw={600} truncate>
                {user.display_name}
              </Text>
              {online !== undefined && (
                <Badge
                  size="xs"
                  variant="light"
                  color={online ? "emerald" : "gray"}
                >
                  {online ? "オンライン" : "オフライン"}
                </Badge>
              )}
            </Group>
            <Text size="xs" c="dimmed" truncate>
              @{user.handle} ・ Lv.{user.level}
            </Text>
          </Stack>
        </Group>
        <Group gap="xs" wrap="nowrap">
          {right}
        </Group>
      </Group>
    </Paper>
  );
}

export function FriendsScreen() {
  const [friends, setFriends] = useState<Friendship[]>([]);
  const [incoming, setIncoming] = useState<Friendship[]>([]);
  const [outgoing, setOutgoing] = useState<Friendship[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [results, setResults] = useState<PublicUser[] | null>(null);
  const [searching, setSearching] = useState(false);

  const reload = useCallback(async () => {
    try {
      const [f, inc, out] = (await Promise.all([
        requestJSON("/friends"),
        requestJSON("/friends/requests?direction=incoming"),
        requestJSON("/friends/requests?direction=outgoing"),
      ])) as [
        { items: Friendship[] },
        { items: Friendship[] },
        { items: Friendship[] },
      ];
      setFriends(f.items);
      setIncoming(inc.items);
      setOutgoing(out.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : "読み込みに失敗しました");
    }
  }, []);

  useEffect(() => {
    const load = () => void reload();
    // 初回もタスクに乗せて、effect 本体では setState しない（cascading render 対策）。
    const kickoff = setTimeout(load, 0);
    // オンライン表示を保つため定期的に引き直す。
    const timer = setInterval(load, 30_000);
    return () => {
      clearTimeout(kickoff);
      clearInterval(timer);
    };
  }, [reload]);

  const act = async (fn: () => Promise<unknown>, done?: string) => {
    setError(null);
    setNotice(null);
    try {
      await fn();
      if (done) setNotice(done);
      await reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : "操作に失敗しました");
    }
  };

  const search = () =>
    act(async () => {
      setSearching(true);
      try {
        const res = (await requestJSON(
          `/users?q=${encodeURIComponent(query)}`,
        )) as { items: PublicUser[] };
        setResults(res.items);
      } finally {
        setSearching(false);
      }
    });

  const sendRequest = (userId: string) =>
    act(
      () =>
        requestJSON("/friends/requests", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ user_id: userId }),
        }),
      "申請を送りました",
    );

  const decide = (userId: string, action: "accept" | "reject") =>
    act(() =>
      requestJSON(`/friends/requests/${userId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action }),
      }),
    );

  const remove = (userId: string) =>
    act(() => requestJSON(`/friends/${userId}`, { method: "DELETE" }));

  const empty = (label: string) => (
    <Text size="sm" c="dimmed" ta="center" py="lg">
      {label}
    </Text>
  );

  return (
    <Stack gap="lg">
      {error && (
        <Alert
          color="red"
          icon={<IconAlertTriangle size={16} />}
          variant="light"
        >
          {error}
        </Alert>
      )}
      {notice && (
        <Alert color="emerald" icon={<IconCheck size={16} />} variant="light">
          {notice}
        </Alert>
      )}

      <Tabs defaultValue="friends" keepMounted={false}>
        <Tabs.List>
          <Tabs.Tab value="friends">フレンド ({friends.length})</Tabs.Tab>
          <Tabs.Tab value="requests">
            届いた申請 {incoming.length > 0 && `(${incoming.length})`}
          </Tabs.Tab>
          <Tabs.Tab value="outgoing">送った申請</Tabs.Tab>
          <Tabs.Tab value="search">探す</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="friends" pt="md">
          <Stack gap="sm">
            {friends.length === 0 &&
              empty("まだフレンドがいません。「探す」から申請を送りましょう。")}
            {friends.map((f) => (
              <UserRow
                key={f.user.id}
                user={f.user}
                online={f.online}
                right={
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    aria-label="フレンドを解除"
                    onClick={() => remove(f.user.id)}
                  >
                    <IconUserMinus size={18} />
                  </ActionIcon>
                }
              />
            ))}
          </Stack>
        </Tabs.Panel>

        <Tabs.Panel value="requests" pt="md">
          <Stack gap="sm">
            {incoming.length === 0 && empty("届いている申請はありません。")}
            {incoming.map((f) => (
              <UserRow
                key={f.user.id}
                user={f.user}
                right={
                  <>
                    <Button
                      size="xs"
                      onClick={() => decide(f.user.id, "accept")}
                    >
                      承認
                    </Button>
                    <ActionIcon
                      variant="subtle"
                      color="red"
                      aria-label="申請を拒否"
                      onClick={() => decide(f.user.id, "reject")}
                    >
                      <IconX size={18} />
                    </ActionIcon>
                  </>
                }
              />
            ))}
          </Stack>
        </Tabs.Panel>

        <Tabs.Panel value="outgoing" pt="md">
          <Stack gap="sm">
            {outgoing.length === 0 && empty("送信中の申請はありません。")}
            {outgoing.map((f) => (
              <UserRow
                key={f.user.id}
                user={f.user}
                right={
                  <Button
                    size="xs"
                    variant="subtle"
                    onClick={() => remove(f.user.id)}
                  >
                    取り下げる
                  </Button>
                }
              />
            ))}
          </Stack>
        </Tabs.Panel>

        <Tabs.Panel value="search" pt="md">
          <Stack gap="sm">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void search();
              }}
            >
              <Group>
                <TextInput
                  className="flex-1"
                  placeholder="表示名か @handle で検索"
                  value={query}
                  onChange={(e) => setQuery(e.currentTarget.value)}
                />
                <Button
                  type="submit"
                  loading={searching}
                  leftSection={<IconSearch size={16} />}
                  disabled={query.trim() === ""}
                >
                  検索
                </Button>
              </Group>
            </form>
            {results !== null &&
              results.length === 0 &&
              empty("見つかりませんでした。")}
            {results?.map((u) => (
              <UserRow
                key={u.id}
                user={u}
                right={
                  <Button
                    size="xs"
                    variant="light"
                    leftSection={<IconUserPlus size={16} />}
                    onClick={() => sendRequest(u.id)}
                  >
                    申請
                  </Button>
                }
              />
            ))}
          </Stack>
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
}
