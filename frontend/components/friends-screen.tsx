"use client";

import Link from "next/link";
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

import { useTranslations } from "next-intl";

import { UserAvatar } from "@/components/user-avatar";
import { ApiError, requestJSON } from "@/lib/request";

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
  const t = useTranslations("friends");
  return (
    <Paper p="sm">
      <Group justify="space-between" wrap="nowrap">
        <Link
          href={`/users/${user.id}`}
          className="min-w-0 no-underline text-inherit"
        >
          <Group gap="sm" wrap="nowrap">
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
                    {online ? t("online") : t("offline")}
                  </Badge>
                )}
              </Group>
              <Text size="xs" c="dimmed" truncate>
                {t("userLine", { handle: user.handle, level: user.level })}
              </Text>
            </Stack>
          </Group>
        </Link>
        <Group gap="xs" wrap="nowrap">
          {right}
        </Group>
      </Group>
    </Paper>
  );
}

export function FriendsScreen() {
  const t = useTranslations("friends");
  const tErr = useTranslations("apiErrors");
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
      setError(e instanceof Error ? e.message : t("loadFailed"));
    }
  }, [t]);

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
      if (e instanceof ApiError && e.code && tErr.has(e.code)) {
        setError(tErr(e.code));
      } else {
        setError(e instanceof Error ? e.message : t("actionFailed"));
      }
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
      t("requestSent"),
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
          <Tabs.Tab value="friends">
            {t("tabFriends", { count: friends.length })}
          </Tabs.Tab>
          <Tabs.Tab value="requests">
            {t("tabRequests")} {incoming.length > 0 && `(${incoming.length})`}
          </Tabs.Tab>
          <Tabs.Tab value="outgoing">{t("tabOutgoing")}</Tabs.Tab>
          <Tabs.Tab value="search">{t("tabSearch")}</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="friends" pt="md">
          <Stack gap="sm">
            {friends.length === 0 && empty(t("emptyFriends"))}
            {friends.map((f) => (
              <UserRow
                key={f.user.id}
                user={f.user}
                online={f.online}
                right={
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    aria-label={t("removeAria")}
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
            {incoming.length === 0 && empty(t("emptyIncoming"))}
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
                      {t("accept")}
                    </Button>
                    <ActionIcon
                      variant="subtle"
                      color="red"
                      aria-label={t("rejectAria")}
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
            {outgoing.length === 0 && empty(t("emptyOutgoing"))}
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
                    {t("withdraw")}
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
                  placeholder={t("searchPlaceholder")}
                  value={query}
                  onChange={(e) => setQuery(e.currentTarget.value)}
                />
                <Button
                  type="submit"
                  loading={searching}
                  leftSection={<IconSearch size={16} />}
                  disabled={query.trim() === ""}
                >
                  {t("search")}
                </Button>
              </Group>
            </form>
            {results !== null &&
              results.length === 0 &&
              empty(t("emptySearch"))}
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
                    {t("request")}
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
