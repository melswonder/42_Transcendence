import {
  Group,
  Paper,
  Progress,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
} from "@mantine/core";
import {
  IconChartLine,
  IconFlame,
  IconSwords,
  IconTrophy,
} from "@tabler/icons-react";

import type { Achievement } from "@/lib/stats";

// 分類ごとにアイコンを変える。同じ絵が並ぶと一覧で見分けがつかないため。
const CATEGORY_ICONS: Record<string, typeof IconTrophy> = {
  wins: IconTrophy,
  streak: IconFlame,
  matches: IconSwords,
  rating: IconChartLine,
};

export function AchievementList({ items }: { items: Achievement[] }) {
  return (
    <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="md">
      {items.map((achievement) => {
        const Icon = CATEGORY_ICONS[achievement.category] ?? IconTrophy;
        const percent = (achievement.progress / achievement.target) * 100;

        return (
          <Paper
            key={achievement.code}
            p="md"
            // 未解除は沈ませて、解除済みとの差を一目で分かるようにする。
            className={achievement.unlocked ? undefined : "opacity-60"}
          >
            <Stack gap="xs">
              <Group gap="sm" wrap="nowrap">
                <ThemeIcon
                  size={38}
                  radius="md"
                  variant={achievement.unlocked ? "filled" : "light"}
                  color={achievement.unlocked ? "emerald" : "gray"}
                >
                  <Icon size={20} />
                </ThemeIcon>
                <Stack gap={0} className="min-w-0">
                  <Text fw={600} truncate>
                    {achievement.name}
                  </Text>
                  <Text size="xs" c="dimmed" truncate>
                    {achievement.description}
                  </Text>
                </Stack>
              </Group>

              <Progress
                value={percent}
                size="sm"
                radius="xl"
                color={achievement.unlocked ? "emerald" : "gray"}
                aria-label={`${achievement.name} の進捗`}
              />
              <Text size="xs" c="dimmed" ta="right" ff="monospace">
                {achievement.progress} / {achievement.target}
              </Text>
            </Stack>
          </Paper>
        );
      })}
    </SimpleGrid>
  );
}
