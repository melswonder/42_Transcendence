import { notFound, redirect } from "next/navigation";
import { cookies } from "next/headers";
import { Group, Paper, Stack, Text, Title } from "@mantine/core";
import { getFormatter, getTranslations } from "next-intl/server";

import { AppShell } from "@/components/app-shell";
import { UserAvatar } from "@/components/user-avatar";
import { serverApiUrl } from "@/lib/api";
import { getCurrentUser } from "@/lib/auth";

/** backend の handler/user.go userPublicResponse と対。 */
interface PublicUser {
  id: string;
  display_name: string;
  handle: string;
  avatar_url: string | null;
  level: number;
  experience_points: number;
  created_at: string;
}

async function getPublicUser(userId: string): Promise<PublicUser | null> {
  const cookie = (await cookies()).toString();
  const res = await fetch(serverApiUrl(`/users/${userId}`), {
    headers: { cookie },
    cache: "no-store",
  });
  if (!res.ok) return null;
  return res.json();
}

/** 公開プロフィールページ。自分のページなら設定への導線を出す。 */
export default async function UserProfilePage({
  params,
}: {
  params: Promise<{ userId: string }>;
}) {
  const me = await getCurrentUser();
  if (!me) redirect("/login");

  const { userId } = await params;
  const user = await getPublicUser(userId);
  if (!user) notFound();

  const t = await getTranslations("profile");
  const format = await getFormatter();

  return (
    <AppShell user={me}>
      <Stack gap="lg" maw={640} mx="auto">
        <Title order={2} size="h3">
          {t("title")}
        </Title>

        <Paper p="xl">
          <Stack gap="lg">
            <Group gap="lg">
              <UserAvatar
                displayName={user.display_name}
                avatarUrl={user.avatar_url}
                size={96}
              />
              <Stack gap={4}>
                <Title order={3}>{user.display_name}</Title>
                <Text c="dimmed">@{user.handle}</Text>
                <Text size="sm" c="dimmed">
                  {t("memberSince", {
                    date: format.dateTime(new Date(user.created_at), {
                      dateStyle: "long",
                    }),
                  })}
                </Text>
              </Stack>
            </Group>

            <Group gap="xl">
              <Stack gap={0}>
                <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
                  {t("level")}
                </Text>
                <Text size="xl" fw={700}>
                  {user.level}
                </Text>
              </Stack>
              <Stack gap={0}>
                <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
                  {t("xp")}
                </Text>
                <Text size="xl" fw={700}>
                  {user.experience_points}
                </Text>
              </Stack>
            </Group>
          </Stack>
        </Paper>
      </Stack>
    </AppShell>
  );
}
