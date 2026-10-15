"use client";

import React, { useEffect } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface PhotoNavigationProps {
  currentIndex: number;
  total: number;
  onPrev: () => void;
  onNext: () => void;
}

export function PhotoNavigation({
  currentIndex,
  total,
  onPrev,
  onNext,
}: PhotoNavigationProps) {
  // Arrow key navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Guard against typing in active input or textarea
      if (
        document.activeElement instanceof HTMLInputElement ||
        document.activeElement instanceof HTMLTextAreaElement ||
        (document.activeElement as HTMLElement)?.isContentEditable
      ) {
        return;
      }

      if (e.key === "ArrowLeft") {
        e.preventDefault();
        onPrev();
      } else if (e.key === "ArrowRight") {
        e.preventDefault();
        onNext();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onPrev, onNext]);

  if (total <= 1) return null;

  const hasPrev = currentIndex > 0;
  const hasNext = currentIndex < total - 1;

  return (
    <>
      {/* Top Center Gallery Indicator Pill */}
      <div className="absolute top-4 left-1/2 -translate-x-1/2 bg-black/60 backdrop-blur-md border border-white/10 rounded-full px-3 py-1 text-white text-xs font-semibold select-none z-30">
        {currentIndex + 1} / {total}
      </div>

      {/* Left Chevron (Previous) */}
      {hasPrev && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onPrev();
          }}
          title="Previous photo (Arrow Left)"
          aria-label="Previous photo"
          className="absolute left-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
        >
          <ChevronLeft className="w-6 h-6 group-hover:-translate-x-0.5 transition-transform" />
        </button>
      )}

      {/* Right Chevron (Next) */}
      {hasNext && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onNext();
          }}
          title="Next photo (Arrow Right)"
          aria-label="Next photo"
          className="absolute right-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
        >
          <ChevronRight className="w-6 h-6 group-hover:translate-x-0.5 transition-transform" />
        </button>
      )}
    </>
  );
}
