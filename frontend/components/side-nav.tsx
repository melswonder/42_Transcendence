"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { NavLink } from "@mantine/core";
import { useTranslations } from "next-intl";
import {
  IconActivity,
  IconAward,
  IconChartBar,
  IconEye,
  IconHistory,
  IconSettings,
  IconTrophy,
  IconUsers,
} from "@tabler/icons-react";

const items = [
  { href: "/", key: "play", icon: IconActivity },
  { href: "/watch", key: "watch", icon: IconEye },
  { href: "/stats", key: "stats", icon: IconChartBar },
  { href: "/matches", key: "matches", icon: IconHistory },
  { href: "/achievements", key: "achievements", icon: IconAward },
  { href: "/leaderboard", key: "leaderboard", icon: IconTrophy },
  { href: "/friends", key: "friends", icon: IconUsers },
  { href: "/settings", key: "settings", icon: IconSettings },
] as const;

/** 現在地の判定に usePathname を使うのでクライアント側に置く。 */
export function SideNav() {
  const pathname = usePathname();
  const t = useTranslations("nav");

  return (
    <>
      {items.map(({ href, key, icon: Icon }) => (
        <NavLink
          key={href}
          component={Link}
          href={href}
          label={t(key)}
          leftSection={<Icon size={20} />}
          active={pathname === href}
          variant="light"
        />
      ))}
    </>
  );
}
