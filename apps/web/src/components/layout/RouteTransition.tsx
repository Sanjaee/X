"use client";

import React from "react";
import { cn } from "@/lib/utils";

export function RouteTransition({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex-1 flex flex-row min-w-0 transition-opacity duration-150 ease-out animate-in fade-in-95 motion-reduce:transition-none motion-reduce:animate-none",
        className
      )}
    >
      {children}
    </div>
  );
}

