import type { ReactNode } from "react";

interface BadgeProps {
  children: ReactNode;
  tone: string;
}

export function Badge({ children, tone }: BadgeProps) {
  return <span className={`badge badge--${tone}`}>{children}</span>;
}