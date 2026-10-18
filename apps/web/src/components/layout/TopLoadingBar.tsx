"use client";

import React, { useEffect, useState, Suspense } from "react";
import { usePathname, useSearchParams } from "next/navigation";

function TopLoadingBarInner() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [loading, setLoading] = useState(false);
  const [progress, setProgress] = useState(0);

  // Trigger progress on route or param change
  useEffect(() => {
    // Briefly show completion animation then reset
    setProgress(100);
    const completeTimer = setTimeout(() => {
      setLoading(false);
      setProgress(0);
    }, 250);

    return () => clearTimeout(completeTimer);
  }, [pathname, searchParams]);

  // Intercept internal link clicks to immediately indicate navigation start
  useEffect(() => {
    const handleAnchorClick = (e: MouseEvent) => {
      const target = (e.target as HTMLElement).closest("a");
      if (!target) return;

      const href = target.getAttribute("href");
      if (
        href &&
        href.startsWith("/") &&
        !href.startsWith("/#") &&
        !target.hasAttribute("download") &&
        target.target !== "_blank" &&
        !e.ctrlKey &&
        !e.metaKey &&
        !e.shiftKey
      ) {
        // Only start if navigating to a different destination
        const currentUrl = window.location.pathname + window.location.search;
        if (href !== currentUrl) {
          setLoading(true);
          setProgress(25);
          const t1 = setTimeout(() => setProgress(65), 100);
          const t2 = setTimeout(() => setProgress(85), 350);
          return () => {
            clearTimeout(t1);
            clearTimeout(t2);
          };
        }
      }
    };

    document.addEventListener("click", handleAnchorClick, { capture: true });
    return () => {
      document.removeEventListener("click", handleAnchorClick, { capture: true });
    };
  }, []);

  if (!loading && progress === 0) return null;

  return (
    <div
      role="progressbar"
      aria-label="Navigation progress"
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={progress}
      className="fixed top-0 left-0 right-0 z-[9999] pointer-events-none h-[2.5px] bg-transparent"
    >
      <div
        className="h-full bg-brand transition-all duration-200 ease-out shadow-[0_0_8px_rgba(29,155,240,0.8)]"
        style={{
          width: `${progress}%`,
          opacity: progress === 100 ? 0 : 1,
          transitionProperty: "width, opacity",
          transitionDuration: progress === 100 ? "250ms" : "200ms",
        }}
      />
    </div>
  );
}

export function TopLoadingBar() {
  return (
    <Suspense fallback={null}>
      <TopLoadingBarInner />
    </Suspense>
  );
}

interface RouteSpinnerProps {
  size?: "sm" | "md" | "lg";
  className?: string;
  label?: string;
}

export function RouteSpinner({
  size = "md",
  className = "",
  label = "Loading",
}: RouteSpinnerProps) {
  const sizeClasses = {
    sm: "w-4 h-4 border-2",
    md: "w-7 h-7 border-[2.5px]",
    lg: "w-10 h-10 border-3",
  };

  return (
    <div
      role="status"
      aria-label={label}
      className={`flex items-center justify-center p-6 ${className}`}
    >
      <div
        className={`${sizeClasses[size]} rounded-full border-brand/20 border-t-brand animate-spin motion-reduce:animate-none`}
      />
      <span className="sr-only">{label}</span>
    </div>
  );
}
