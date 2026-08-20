import { Avatar, Group, Paper, Stack, Text } from "@mantine/core";

import { LogoutButton } from "@/components/logout-button";
import type { User } from "@/lib/auth";

/** サイドバー下部の自分のカード。 */
export function UserCard({ user }: { user: User }) {
  return (
    <Paper p="sm" bg="dark.7">
      <Group justify="space-between" wrap="nowrap">
        <Group gap="sm" wrap="nowrap" className="min-w-0">
          <Avatar radius="xl" color="emerald" variant="filled">
            {user.display_name.charAt(0).toUpperCase()}
          </Avatar>
          <Stack gap={0} className="min-w-0">
            <Text size="sm" fw={600} truncate>
              {user.display_name}
            </Text>
            <Text size="xs" c="emerald">
              Lv.{user.level}
            </Text>
          </Stack>
        </Group>
        <LogoutButton iconOnly />
      </Group>
    </Paper>
  );
}
