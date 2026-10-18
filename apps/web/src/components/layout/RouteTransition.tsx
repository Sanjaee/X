"use client";

import React from "react";

export function RouteTransition({ children }: { children: React.ReactNode }) {
  return (
    <div className="w-full flex-1 flex flex-row min-w-0 transition-opacity duration-150 ease-out animate-in fade-in-95 motion-reduce:transition-none motion-reduce:animate-none">
      {children}
    </div>
  );
}
