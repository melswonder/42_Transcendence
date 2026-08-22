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

import { useTranslations } from "next-intl";

import type { Achievement } from "@/lib/stats";

// 分類ごとにアイコンを変える。同じ絵が並ぶと一覧で見分けがつかないため。
const CATEGORY_ICONS: Record<string, typeof IconTrophy> = {
  wins: IconTrophy,
  streak: IconFlame,
  matches: IconSwords,
  rating: IconChartLine,
};

export function AchievementList({ items }: { items: Achievement[] }) {
  const t = useTranslations("achievements");
  // 実績の定義は backend（日本語）にあるので、code を鍵に各言語へ引き直す。
  // 未知の code が来ても壊れないよう、サーバーの文言へフォールバックする。
  const name = (a: Achievement) =>
    t.has(`items.${a.code}.name`) ? t(`items.${a.code}.name`) : a.name;
  const description = (a: Achievement) =>
    t.has(`items.${a.code}.description`)
      ? t(`items.${a.code}.description`)
      : a.description;
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
                    {name(achievement)}
                  </Text>
                  <Text size="xs" c="dimmed" truncate>
                    {description(achievement)}
                  </Text>
                </Stack>
              </Group>

              <Progress
                value={percent}
                size="sm"
                radius="xl"
                color={achievement.unlocked ? "emerald" : "gray"}
                aria-label={t("progressAria", { name: name(achievement) })}
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
