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

import { UserAvatar } from "@/components/user-avatar";
import { apiUrl } from "@/lib/api";
import type { User } from "@/lib/auth";

const locales = [
  { value: "ja", label: "日本語" },
  { value: "en", label: "English" },
  { value: "fr", label: "Français" },
];

/** アバターは png / jpeg / webp、最大 5MB。バックエンドの検証と揃える。 */
const avatarMaxBytes = 5 * 1024 * 1024;

async function requestJSON(path: string, init?: RequestInit) {
  const res = await fetch(apiUrl(path), { credentials: "include", ...init });
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? `リクエストに失敗しました (${res.status})`);
  }
  return res.status === 204 ? null : res.json();
}

export function SettingsForm({ user }: { user: User }) {
  const router = useRouter();

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
      router.refresh(); // サイドバーの表示名なども追従させる
    } catch (e) {
      setError(e instanceof Error ? e.message : "保存に失敗しました");
    } finally {
      setSaving(false);
    }
  };

  /** アップロード → その場で avatar_asset_id を差して反映、の 2 段階。 */
  const uploadAvatar = async (file: File | null) => {
    if (!file) return;
    if (file.size > avatarMaxBytes) {
      setError("画像は 5MB 以下にしてください");
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
      setError(e instanceof Error ? e.message : "アップロードに失敗しました");
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
      setError(e instanceof Error ? e.message : "削除に失敗しました");
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
          保存しました
        </Alert>
      )}

      <Paper p="lg">
        <Stack gap="md">
          <Title order={3} size="h5">
            アバター
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
                      画像を選ぶ
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
                    削除する
                  </Button>
                )}
              </Group>
              <Text size="xs" c="dimmed">
                png / jpeg / webp、5MB まで。削除するとデフォルトに戻ります。
              </Text>
            </Stack>
          </Group>
        </Stack>
      </Paper>

      <Paper p="lg">
        <Stack gap="md">
          <Title order={3} size="h5">
            プロフィール
          </Title>
          <TextInput
            label="表示名"
            value={displayName}
            maxLength={50}
            onChange={(e) => setDisplayName(e.currentTarget.value)}
          />
          <TextInput
            label="ハンドル"
            description="英小文字・数字・アンダースコア、3〜30 文字"
            value={handle}
            maxLength={30}
            onChange={(e) => setHandle(e.currentTarget.value.toLowerCase())}
          />
          <Select
            label="言語"
            data={locales}
            value={locale}
            allowDeselect={false}
            onChange={(v) => v && setLocale(v)}
          />
          <Group justify="flex-end">
            <Button onClick={save} loading={saving}>
              保存する
            </Button>
          </Group>
        </Stack>
      </Paper>
    </Stack>
  );
}
