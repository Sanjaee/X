"use client";

import React from "react";
import { RealtimeProvider } from "@/lib/useRealtime";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <RealtimeProvider>
      {children}
    </RealtimeProvider>
  );
}
