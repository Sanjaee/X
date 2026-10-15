"use client";

import React, { useState } from "react";
import {
  Play,
  Pause,
  Volume2,
  VolumeX,
  Maximize,
  Minimize,
  Gauge,
} from "lucide-react";
import { VideoProgressBar } from "./VideoProgressBar";

interface VideoControlsProps {
  isPlaying: boolean;
  isMuted: boolean;
  volume: number;
  currentTime: number;
  duration: number;
  buffered: number;
  isFullscreen: boolean;
  playbackSpeed: number;
  onTogglePlay: () => void;
  onToggleMute: () => void;
  onVolumeChange: (vol: number) => void;
  onSeek: (time: number) => void;
  onToggleFullscreen: () => void;
  onSpeedChange: (speed: number) => void;
}

function formatTime(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

const SPEED_OPTIONS = [0.5, 1.0, 1.25, 1.5, 2.0];

export function VideoControls({
  isPlaying,
  isMuted,
  volume,
  currentTime,
  duration,
  buffered,
  isFullscreen,
  playbackSpeed,
  onTogglePlay,
  onToggleMute,
  onVolumeChange,
  onSeek,
  onToggleFullscreen,
  onSpeedChange,
}: VideoControlsProps) {
  const [showSpeedMenu, setShowSpeedMenu] = useState(false);

  return (
    <div
      onClick={(e) => e.stopPropagation()}
      className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/90 via-black/50 to-transparent pt-8 pb-3 px-4 flex flex-col gap-1 z-20 select-none text-white transition-opacity duration-300"
    >
      {/* Progress Scrubber */}
      <VideoProgressBar
        currentTime={currentTime}
        duration={duration}
        buffered={buffered}
        onSeek={onSeek}
      />

      {/* Buttons & Indicators Row */}
      <div className="flex items-center justify-between mt-1">
        {/* Left Side: Play/Pause, Volume, Time */}
        <div className="flex items-center gap-3">
          {/* Play/Pause */}
          <button
            onClick={onTogglePlay}
            title={isPlaying ? "Pause (Space)" : "Play (Space)"}
            aria-label={isPlaying ? "Pause" : "Play"}
            className="p-1.5 hover:bg-white/10 rounded-full transition-colors"
          >
            {isPlaying ? (
              <Pause className="w-5 h-5 fill-white text-white" />
            ) : (
              <Play className="w-5 h-5 fill-white text-white" />
            )}
          </button>

          {/* Volume & Slider Container */}
          <div className="flex items-center group/vol">
            <button
              onClick={onToggleMute}
              title={isMuted ? "Unmute (M)" : "Mute (M)"}
              aria-label={isMuted ? "Unmute" : "Mute"}
              className="p-1.5 hover:bg-white/10 rounded-full transition-colors"
            >
              {isMuted || volume === 0 ? (
                <VolumeX className="w-5 h-5 text-white" />
              ) : (
                <Volume2 className="w-5 h-5 text-white" />
              )}
            </button>

            {/* Expandable Volume Slider */}
            <div className="w-0 overflow-hidden group-hover/vol:w-16 transition-all duration-200 flex items-center pl-1">
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={isMuted ? 0 : volume}
                onChange={(e) => onVolumeChange(parseFloat(e.target.value))}
                className="w-14 h-1 bg-white/30 rounded-lg appearance-none cursor-pointer accent-brand"
              />
            </div>
          </div>

          {/* Timestamp */}
          <div className="text-xs font-medium text-white/90 font-mono ml-1">
            <span>{formatTime(currentTime)}</span>
            <span className="mx-1 text-white/40">/</span>
            <span>{formatTime(duration)}</span>
          </div>
        </div>

        {/* Right Side: Speed Selector & Fullscreen */}
        <div className="flex items-center gap-2 relative">
          {/* Playback Speed Menu */}
          <div className="relative">
            <button
              onClick={() => setShowSpeedMenu(!showSpeedMenu)}
              title="Playback speed"
              aria-label="Playback speed"
              className="px-2 py-1 hover:bg-white/10 rounded-md text-xs font-bold transition-colors flex items-center gap-1"
            >
              <Gauge className="w-3.5 h-3.5" />
              <span>{playbackSpeed}x</span>
            </button>

            {showSpeedMenu && (
              <>
                <div
                  className="fixed inset-0 z-30"
                  onClick={() => setShowSpeedMenu(false)}
                />
                <div className="absolute right-0 bottom-8 bg-black/90 backdrop-blur-md border border-white/10 rounded-xl py-1 shadow-2xl z-40 min-w-[80px]">
                  {SPEED_OPTIONS.map((spd) => (
                    <button
                      key={spd}
                      onClick={() => {
                        onSpeedChange(spd);
                        setShowSpeedMenu(false);
                      }}
                      className={`w-full text-left px-3 py-1.5 text-xs font-semibold hover:bg-white/10 transition-colors flex items-center justify-between ${
                        playbackSpeed === spd ? "text-brand" : "text-white"
                      }`}
                    >
                      <span>{spd}x</span>
                      {playbackSpeed === spd && <span className="text-[10px]">✓</span>}
                    </button>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Fullscreen Toggle */}
          <button
            onClick={onToggleFullscreen}
            title={isFullscreen ? "Exit fullscreen (F / Esc)" : "Fullscreen (F)"}
            aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
            className="p-1.5 hover:bg-white/10 rounded-full transition-colors"
          >
            {isFullscreen ? (
              <Minimize className="w-5 h-5 text-white" />
            ) : (
              <Maximize className="w-5 h-5 text-white" />
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
