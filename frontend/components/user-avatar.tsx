import { Avatar } from "@mantine/core";
import { useTranslations } from "next-intl";

import { apiUrl } from "@/lib/api";

/** アバター表示の共通部品。
 * avatar_url が無ければデフォルト（表示名の頭文字）に落とす。
 * URL はバックエンドの相対パス（/media/{id}/file）で届くので、ここで絶対 URL にする。
 */
export function UserAvatar({
  displayName,
  avatarUrl,
  size,
}: {
  displayName: string;
  avatarUrl: string | null | undefined;
  size?: number | string;
}) {
  const t = useTranslations("avatar");
  return (
    <Avatar
      radius="xl"
      size={size}
      color="emerald"
      variant="filled"
      src={avatarUrl ? apiUrl(avatarUrl) : null}
      alt={t("alt", { name: displayName })}
    >
      {displayName.charAt(0).toUpperCase()}
    </Avatar>
  );
}
