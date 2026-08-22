import Link from "next/link";
import { Group, Paper, Stack, Text } from "@mantine/core";
import { useTranslations } from "next-intl";

import { LogoutButton } from "@/components/logout-button";
import { UserAvatar } from "@/components/user-avatar";
import type { User } from "@/lib/auth";

/** サイドバー下部の自分のカード。 */
export function UserCard({ user }: { user: User }) {
  const t = useTranslations("nav");
  return (
    <Paper p="sm" bg="dark.7">
      <Group justify="space-between" wrap="nowrap">
        <Link
          href={`/users/${user.id}`}
          className="min-w-0 no-underline text-inherit"
        >
          <Group gap="sm" wrap="nowrap">
            <UserAvatar
              displayName={user.display_name}
              avatarUrl={user.avatar_url}
            />
            <Stack gap={0} className="min-w-0">
              <Text size="sm" fw={600} truncate>
                {user.display_name}
              </Text>
              <Text size="xs" c="emerald">
                {t("level", { level: user.level })}
              </Text>
            </Stack>
          </Group>
        </Link>
        <LogoutButton iconOnly />
      </Group>
    </Paper>
  );
}
