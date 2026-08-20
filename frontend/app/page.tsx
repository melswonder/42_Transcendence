import { redirect } from "next/navigation";
import { Stack, Text, Title } from "@mantine/core";

import { AppShell } from "@/components/app-shell";
import { getCurrentUser } from "@/lib/auth";

export default async function HomePage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return (
    <AppShell user={user}>
      <Stack gap="xs">
        <Title order={2}>おかえりなさい、{user.display_name} さん</Title>
        <Text c="dimmed">@{user.handle}</Text>
      </Stack>
    </AppShell>
  );
}
