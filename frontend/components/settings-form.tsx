"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import {
  Alert,
  Button,
  FileButton,
  Group,
  Paper,
  Select,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { IconAlertTriangle, IconCheck, IconUpload } from "@tabler/icons-react";
import { useTranslations } from "next-intl";

import { UserAvatar } from "@/components/user-avatar";
import type { User } from "@/lib/auth";
import { ApiError, requestJSON } from "@/lib/request";

// 言語名は各言語の自称のまま出す（他言語表示でも見つけやすい）。
const locales = [
  { value: "ja", label: "日本語" },
  { value: "en", label: "English" },
  { value: "fr", label: "Français" },
];

/** アバターは png / jpeg / webp、最大 5MB。バックエンドの検証と揃える。 */
const avatarMaxBytes = 5 * 1024 * 1024;

export function SettingsForm({ user }: { user: User }) {
  const t = useTranslations("settings");
  const tErr = useTranslations("apiErrors");
  const router = useRouter();

  /** code があれば翻訳、無ければサーバーの文言をそのまま。 */
  const errorMessage = (e: unknown, fallback: string) => {
    if (e instanceof ApiError && e.code && tErr.has(e.code)) {
      return tErr(e.code);
    }
    return e instanceof Error ? e.message : fallback;
  };

  const [displayName, setDisplayName] = useState(user.display_name);
  const [handle, setHandle] = useState(user.handle);
  const [locale, setLocale] = useState(user.preferred_locale);
  const [avatarUrl, setAvatarUrl] = useState(user.avatar_url);
  const [avatarAssetId, setAvatarAssetId] = useState(user.avatar_asset_id);

  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  const save = async () => {
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      await requestJSON("/users/me", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          display_name: displayName,
          handle,
          preferred_locale: locale,
        }),
      });
      setSaved(true);
      // 言語はその場で切り替える。Cookie が表示言語の正本（i18n/request.ts）。
      document.cookie = `NEXT_LOCALE=${locale}; path=/; max-age=31536000; samesite=lax`;
      router.refresh(); // サイドバーの表示名なども追従させる
    } catch (e) {
      setError(errorMessage(e, t("saveFailed")));
    } finally {
      setSaving(false);
    }
  };

  /** アップロード → その場で avatar_asset_id を差して反映、の 2 段階。 */
  const uploadAvatar = async (file: File | null) => {
    if (!file) return;
    if (file.size > avatarMaxBytes) {
      setError(t("avatarTooLarge"));
      return;
    }
    setUploading(true);
    setError(null);
    try {
      const form = new FormData();
      form.append("file", file);
      const asset = (await requestJSON("/media/avatars", {
        method: "POST",
        body: form,
      })) as { id: string; url: string };
      await requestJSON("/users/me", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ avatar_asset_id: asset.id }),
      });
      setAvatarUrl(asset.url);
      setAvatarAssetId(asset.id);
      router.refresh();
    } catch (e) {
      setError(errorMessage(e, t("uploadFailed")));
    } finally {
      setUploading(false);
    }
  };

  /** 削除するとデフォルト（頭文字）に戻る。 */
  const removeAvatar = async () => {
    if (!avatarAssetId) return;
    setUploading(true);
    setError(null);
    try {
      await requestJSON(`/media/${avatarAssetId}`, { method: "DELETE" });
      setAvatarUrl(null);
      setAvatarAssetId(null);
      router.refresh();
    } catch (e) {
      setError(errorMessage(e, t("deleteFailed")));
    } finally {
      setUploading(false);
    }
  };

  return (
    <Stack gap="lg" maw={560}>
      {error && (
        <Alert
          color="red"
          icon={<IconAlertTriangle size={16} />}
          variant="light"
        >
          {error}
        </Alert>
      )}
      {saved && (
        <Alert color="emerald" icon={<IconCheck size={16} />} variant="light">
          {t("saved")}
        </Alert>
      )}

      <Paper p="lg">
        <Stack gap="md">
          <Title order={3} size="h5">
            {t("avatarSection")}
          </Title>
          <Group>
            <UserAvatar
              displayName={displayName || user.display_name}
              avatarUrl={avatarUrl}
              size={72}
            />
            <Stack gap="xs">
              <Group gap="sm">
                <FileButton
                  onChange={uploadAvatar}
                  accept="image/png,image/jpeg,image/webp"
                >
                  {(props) => (
                    <Button
                      {...props}
                      variant="light"
                      loading={uploading}
                      leftSection={<IconUpload size={16} />}
                    >
                      {t("choose")}
                    </Button>
                  )}
                </FileButton>
                {avatarAssetId && (
                  <Button
                    variant="subtle"
                    color="red"
                    disabled={uploading}
                    onClick={removeAvatar}
                  >
                    {t("remove")}
                  </Button>
                )}
              </Group>
              <Text size="xs" c="dimmed">
                {t("avatarHint")}
              </Text>
            </Stack>
          </Group>
        </Stack>
      </Paper>

      <Paper p="lg">
        <Stack gap="md">
          <Title order={3} size="h5">
            {t("profileSection")}
          </Title>
          <TextInput
            label={t("displayName")}
            value={displayName}
            maxLength={50}
            onChange={(e) => setDisplayName(e.currentTarget.value)}
          />
          <TextInput
            label={t("handle")}
            description={t("handleHint")}
            value={handle}
            maxLength={30}
            onChange={(e) => setHandle(e.currentTarget.value.toLowerCase())}
          />
          <Select
            label={t("language")}
            data={locales}
            value={locale}
            allowDeselect={false}
            onChange={(v) => v && setLocale(v)}
          />
          <Group justify="flex-end">
            <Button onClick={save} loading={saving}>
              {t("save")}
            </Button>
          </Group>
        </Stack>
      </Paper>
    </Stack>
  );
}
