import { CircleAlertIcon } from "lucide-react";

import { GoogleLoginButton, resolveLoginError } from "@/features/auth-google";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { Card, CardContent, CardHeader } from "@/shared/ui/card";
import { PongMark } from "@/shared/ui/pong-mark";

export function LoginPage({ error }: { error?: string }) {
  const message = resolveLoginError(error);

  return (
    <main className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="items-center text-center">
          <div className="mb-4 inline-flex size-16 items-center justify-center rounded-xl bg-primary/20 text-primary">
            <PongMark />
          </div>
          <h1 className="text-3xl font-bold tracking-wider text-foreground">
            TRANSCENDENCE
          </h1>
          <p className="mt-2 text-muted-foreground">Sign in to play</p>
        </CardHeader>

        <CardContent className="flex flex-col gap-6">
          {message && (
            // Alert は内部で role="alert" を持つので、ここで付け直す必要はない。
            <Alert variant="destructive">
              <CircleAlertIcon />
              <AlertDescription>{message}</AlertDescription>
            </Alert>
          )}

          <GoogleLoginButton />

          <p className="text-center text-xs text-muted-foreground">
            続行すると、Google
            アカウントの表示名・メールアドレス・プロフィール画像を取得します。
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
