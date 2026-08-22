"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Group, Pagination, Text } from "@mantine/core";
import { useFormatter, useTranslations } from "next-intl";

/** ページ番号も URL のクエリに持つ。リロードや共有で位置が保たれる。 */
export function PaginationControls({
  total,
  limit,
  offset,
}: {
  total: number;
  limit: number;
  offset: number;
}) {
  const t = useTranslations("pagination");
  const format = useFormatter();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const pages = Math.ceil(total / limit);
  if (pages <= 1) return null;

  const goTo = (page: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("offset", `${(page - 1) * limit}`);
    router.push(`${pathname}?${params.toString()}`);
  };

  return (
    <Group justify="space-between" className="print:hidden">
      <Text size="sm" c="dimmed">
        {t("range", {
          total,
          from: format.number(offset + 1),
          to: format.number(Math.min(offset + limit, total)),
        })}
      </Text>
      <Pagination
        total={pages}
        value={Math.floor(offset / limit) + 1}
        onChange={goTo}
        size="sm"
      />
    </Group>
  );
}
