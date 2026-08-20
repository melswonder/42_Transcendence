import { rem } from "@mantine/core";

interface QuoridorMarkProps {
  size?: number | string;
}

/** ロゴ代わりの盤・壁・駒。色は currentColor 任せで、置いた側の c / color に従う。 */
export function QuoridorMark({ size = 32 }: QuoridorMarkProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      width={rem(size)}
      height={rem(size)}
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <rect x="3" y="3" width="18" height="18" rx="2" />
      {/* マス目 */}
      <path d="M9 3v18M15 3v18M3 9h18M3 15h18" opacity="0.4" />
      {/* 壁 */}
      <path d="M9 15h9" strokeWidth="2.5" />
      <circle cx="6" cy="6" r="1.75" fill="currentColor" stroke="none" />
    </svg>
  );
}
