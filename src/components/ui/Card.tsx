import type { HTMLAttributes, ReactNode } from "react";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function Card({ children, className = "", ...props }: CardProps) {
  return (
    <section className={`card ${className}`.trim()} {...props}>
      {children}
    </section>
  );
}