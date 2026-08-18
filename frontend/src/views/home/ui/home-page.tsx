import Link from "next/link";

import { Button } from "@/shared/ui/button";
import { PongMark } from "@/shared/ui/pong-mark";

export function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-10 bg-background p-8 text-center">
      <div className="flex flex-col items-center gap-6">
        <div className="inline-flex size-20 items-center justify-center rounded-2xl bg-primary/20 text-primary">
          <PongMark className="size-11" />
        </div>
        <h1 className="text-5xl font-bold tracking-wider text-foreground">
          TRANSCENDENCE
        </h1>
        <p className="max-w-md text-lg text-muted-foreground">
          ブラウザで対戦できる Pong。ログインしてゲームを始めましょう。
        </p>
      </div>

      <Button asChild size="lg" className="h-12 px-8 text-base">
        <Link href="/login">ログインして始める</Link>
      </Button>
    </main>
  );
}
