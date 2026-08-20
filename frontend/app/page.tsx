import { redirect } from "next/navigation";
import {
  Avatar,
  Container,
  Group,
  Paper,
  Stack,
  Text,
  Title,
} from "@mantine/core";

import { LogoutButton } from "@/components/logout-button";
import { getCurrentUser } from "@/lib/auth";

export default async function HomePage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return (
    <main className="flex min-h-dvh items-center justify-center bg-body p-6">
      <Container size="sm" w="100%">
        <Paper p="xl" shadow="md">
          <Stack gap="lg">
            <Group justify="space-between">
              <Title order={1} className="text-2xl tracking-[0.15em]">
                TRANSCENDENCE
              </Title>
              <LogoutButton />
            </Group>

            <Group gap="md">
              <Avatar size={56} radius="xl" color="emerald" variant="filled">
                {user.display_name.charAt(0).toUpperCase()}
              </Avatar>
              <Stack gap={2}>
                <Text fw={600}>{user.display_name}</Text>
                <Text size="sm" c="dimmed">
                  @{user.handle}
                </Text>
                <Text size="sm" c="emerald">
                  Lv.{user.level}
                </Text>
              </Stack>
            </Group>
          </Stack>
        </Paper>
      </Container>
    </main>
  );
}
