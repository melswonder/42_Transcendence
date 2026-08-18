import { cn } from "@/shared/lib/utils";

/** ロゴ代わりのパドルとボール。プロダクト固有の意匠なので lucide には無く、
 * ここで持つ。色は currentColor 任せにして、置いた側の text-* に従わせる。
 */
export function PongMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      aria-hidden="true"
      className={cn("size-8", className)}
    >
      <path d="M4 7v10" />
      <path d="M20 7v10" />
      <circle cx="12" cy="12" r="2" fill="currentColor" stroke="none" />
    </svg>
  );
}
