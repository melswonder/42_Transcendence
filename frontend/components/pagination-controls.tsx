"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Group, Pagination, Text } from "@mantine/core";

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
        {total} 件中 {offset + 1}–{Math.min(offset + limit, total)} 件
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
