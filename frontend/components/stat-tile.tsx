import type { ReactNode } from "react";
import { Group, Paper, Stack, Text, ThemeIcon } from "@mantine/core";

interface StatTileProps {
  icon: ReactNode;
  label: string;
  value: string;
  /** 値の下に出す補足（「250 人中」「直近 40 戦」など）。 */
  hint?: string;
  /** 良い/悪いで色を変えたいときだけ渡す。 */
  tone?: "positive" | "negative";
}

export function StatTile({ icon, label, value, hint, tone }: StatTileProps) {
  const color =
    tone === "positive" ? "emerald" : tone === "negative" ? "red" : undefined;

  return (
    <Paper p="md">
      <Group gap="sm" wrap="nowrap" align="flex-start">
        <ThemeIcon size={38} radius="md" variant="light" color={color}>
          {icon}
        </ThemeIcon>
        <Stack gap={0} className="min-w-0">
          <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
            {label}
          </Text>
          <Text size="xl" fw={700} c={color} lh={1.2}>
            {value}
          </Text>
          {hint && (
            <Text size="xs" c="dimmed">
              {hint}
            </Text>
          )}
        </Stack>
      </Group>
    </Paper>
  );
}
