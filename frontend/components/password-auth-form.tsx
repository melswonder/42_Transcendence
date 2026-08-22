"use client";

import { useState } from "react";
import { Alert, Button, PasswordInput, Stack, TextInput } from "@mantine/core";
import { IconAlertTriangle } from "@tabler/icons-react";
import { useTranslations } from "next-intl";

import { ApiError, requestJSON } from "@/lib/request";

/** メール+パスワードの認証フォーム。signup は登録、login はログイン。
 * 成功したらセッション Cookie が付くので、ホームへ遷移するだけでよい。
 */
export function PasswordAuthForm({ mode }: { mode: "login" | "signup" }) {
  const t = useTranslations("auth");
  const tErr = useTranslations("apiErrors");

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [handle, setHandle] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    setError(null);
    if (mode === "signup" && password !== confirm) {
      setError(t("passwordMismatch"));
      return;
    }
    setLoading(true);
    try {
      if (mode === "signup") {
        await requestJSON("/auth/register", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            email,
            password,
            display_name: displayName,
            handle,
          }),
        });
      } else {
        await requestJSON("/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });
      }
      // Server Component が握るユーザー情報ごと取り直すため、遷移で読み込み直す。
      window.location.href = "/";
    } catch (e) {
      if (e instanceof ApiError && e.code && tErr.has(e.code)) {
        setError(tErr(e.code));
      } else {
        setError(e instanceof Error ? e.message : t("failed"));
      }
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        void submit();
      }}
    >
      <Stack gap="sm">
        {error && (
          <Alert
            color="red"
            variant="light"
            icon={<IconAlertTriangle size={16} />}
          >
            {error}
          </Alert>
        )}

        {mode === "signup" && (
          <>
            <TextInput
              label={t("displayName")}
              value={displayName}
              onChange={(e) => setDisplayName(e.currentTarget.value)}
              maxLength={50}
              required
            />
            <TextInput
              label={t("handle")}
              description={t("handleHint")}
              value={handle}
              onChange={(e) => setHandle(e.currentTarget.value.toLowerCase())}
              maxLength={30}
              required
            />
          </>
        )}

        <TextInput
          label={t("email")}
          type="email"
          value={email}
          onChange={(e) => setEmail(e.currentTarget.value)}
          required
        />
        <PasswordInput
          label={t("password")}
          description={mode === "signup" ? t("passwordHint") : undefined}
          value={password}
          onChange={(e) => setPassword(e.currentTarget.value)}
          minLength={8}
          maxLength={72}
          required
        />
        {mode === "signup" && (
          <PasswordInput
            label={t("passwordConfirm")}
            value={confirm}
            onChange={(e) => setConfirm(e.currentTarget.value)}
            required
          />
        )}

        <Button type="submit" loading={loading} fullWidth>
          {mode === "signup" ? t("signupSubmit") : t("loginSubmit")}
        </Button>
      </Stack>
    </form>
  );
}
