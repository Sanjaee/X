"use client";

import React, { useState, useRef, useCallback } from "react";

interface VideoProgressBarProps {
  currentTime: number;
  duration: number;
  buffered: number; // 0 to 1
  onSeek: (time: number) => void;
}

function formatTime(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

export function VideoProgressBar({
  currentTime,
  duration,
  buffered,
  onSeek,
}: VideoProgressBarProps) {
  const [isHovering, setIsHovering] = useState(false);
  const [hoverPosition, setHoverPosition] = useState(0); // 0 to 1
  const [hoverTime, setHoverTime] = useState(0);
  const barRef = useRef<HTMLDivElement>(null);

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;

  const calculateTimeFromEvent = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!barRef.current || duration <= 0) return 0;
      const rect = barRef.current.getBoundingClientRect();
      const pos = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
      return pos * duration;
    },
    [duration]
  );

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!barRef.current || duration <= 0) return;
      const rect = barRef.current.getBoundingClientRect();
      const pos = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
      setHoverPosition(pos);
      setHoverTime(pos * duration);
    },
    [duration]
  );

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation();
      const targetTime = calculateTimeFromEvent(e);
      onSeek(targetTime);
    },
    [calculateTimeFromEvent, onSeek]
  );

  return (
    <div
      ref={barRef}
      onClick={handleClick}
      onMouseEnter={() => setIsHovering(true)}
      onMouseMove={handleMouseMove}
      onMouseLeave={() => setIsHovering(false)}
      className="relative w-full h-3 flex items-center cursor-pointer group py-1"
    >
      {/* Background Track */}
      <div className="w-full h-1 group-hover:h-1.5 bg-white/20 rounded-full overflow-hidden transition-all relative">
        {/* Buffered Progress */}
        <div
          className="absolute left-0 top-0 bottom-0 bg-white/40 transition-all duration-200"
          style={{ width: `${Math.min(100, buffered * 100)}%` }}
        />
        {/* Played Progress */}
        <div
          className="absolute left-0 top-0 bottom-0 bg-brand rounded-full transition-all"
          style={{ width: `${Math.min(100, progress)}%` }}
        />
      </div>

      {/* Scrubber Knob */}
      <div
        className="absolute w-3.5 h-3.5 bg-brand rounded-full shadow-md -translate-x-1/2 scale-0 group-hover:scale-100 transition-transform pointer-events-none"
        style={{ left: `${Math.min(100, progress)}%` }}
      />

      {/* Hover Timestamp Tooltip */}
      {isHovering && duration > 0 && (
        <div
          className="absolute -top-7 -translate-x-1/2 bg-black/80 backdrop-blur-md border border-white/10 text-white text-[11px] font-semibold px-2 py-0.5 rounded shadow pointer-events-none whitespace-nowrap"
          style={{ left: `${hoverPosition * 100}%` }}
        >
          {formatTime(hoverTime)}
        </div>
      )}
    </div>
  );
}
