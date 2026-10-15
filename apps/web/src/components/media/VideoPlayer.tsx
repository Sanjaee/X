"use client";

import React, { useState, useRef, useEffect, useCallback } from "react";
import { Play, Loader2, AlertCircle } from "lucide-react";
import { VideoControls } from "./VideoControls";

interface VideoPlayerProps {
  src: string;
  poster?: string;
  aspectRatio?: string;
  autoPlay?: boolean;
  mode?: "in_feed" | "viewer";
  onOpenViewer?: () => void;
}

export function VideoPlayer({
  src,
  poster,
  aspectRatio = "16/9",
  autoPlay = false,
  mode = "in_feed",
  onOpenViewer,
}: VideoPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const hideControlsTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [isMuted, setIsMuted] = useState(mode === "in_feed"); // Feed videos must be muted by default
  const [volume, setVolume] = useState(1);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [buffered, setBuffered] = useState(0);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const [isBuffering, setIsBuffering] = useState(false);
  const [error, setError] = useState(false);
  const [showControls, setShowControls] = useState(true);

  // Synchronize playback state
  const handleTogglePlay = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;

    if (video.paused) {
      // Broadcast event so other playing videos pause
      window.dispatchEvent(new CustomEvent("x_video_play", { detail: { src } }));
      video.play().catch(() => {});
    } else {
      video.pause();
    }
  }, [src]);

  // Handle global single-video playback: pause if another video starts
  useEffect(() => {
    const onOtherVideoPlay = (e: Event) => {
      const custom = e as CustomEvent<{ src: string }>;
      if (custom.detail?.src !== src && videoRef.current && !videoRef.current.paused) {
        videoRef.current.pause();
      }
    };

    window.addEventListener("x_video_play", onOtherVideoPlay);
    return () => window.removeEventListener("x_video_play", onOtherVideoPlay);
  }, [src]);

  // In-Feed IntersectionObserver: Auto-pause when scrolled out of view
  useEffect(() => {
    if (mode !== "in_feed") return;
    const el = containerRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry.isIntersecting && videoRef.current && !videoRef.current.paused) {
          videoRef.current.pause();
        }
      },
      { threshold: 0.5 }
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [mode]);

  // Autohide controls after 2.5s of inactivity
  const resetHideControlsTimer = useCallback(() => {
    setShowControls(true);
    if (hideControlsTimeoutRef.current) {
      clearTimeout(hideControlsTimeoutRef.current);
    }
    if (isPlaying) {
      hideControlsTimeoutRef.current = setTimeout(() => {
        setShowControls(false);
      }, 2500);
    }
  }, [isPlaying]);

  const handleMouseMove = useCallback(() => {
    resetHideControlsTimer();
  }, [resetHideControlsTimer]);

  // Keyboard Shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Guard against typing in active inputs
      if (
        document.activeElement instanceof HTMLInputElement ||
        document.activeElement instanceof HTMLTextAreaElement ||
        (document.activeElement as HTMLElement)?.isContentEditable
      ) {
        return;
      }

      const video = videoRef.current;
      if (!video) return;

      if (e.code === "Space") {
        e.preventDefault();
        handleTogglePlay();
      } else if (e.code === "ArrowLeft") {
        e.preventDefault();
        video.currentTime = Math.max(0, video.currentTime - 5);
      } else if (e.code === "ArrowRight") {
        e.preventDefault();
        video.currentTime = Math.min(video.duration, video.currentTime + 5);
      } else if (e.key.toLowerCase() === "m") {
        e.preventDefault();
        setIsMuted((prev) => {
          const next = !prev;
          video.muted = next;
          return next;
        });
      } else if (e.key.toLowerCase() === "f") {
        e.preventDefault();
        handleToggleFullscreen();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleTogglePlay]);

  const handleToggleMute = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;
    const nextMuted = !isMuted;
    video.muted = nextMuted;
    setIsMuted(nextMuted);
    if (!nextMuted && volume === 0) {
      setVolume(0.5);
      video.volume = 0.5;
    }
  }, [isMuted, volume]);

  const handleVolumeChange = useCallback((newVol: number) => {
    const video = videoRef.current;
    if (!video) return;
    video.volume = newVol;
    setVolume(newVol);
    const shouldMute = newVol === 0;
    video.muted = shouldMute;
    setIsMuted(shouldMute);
  }, []);

  const handleSeek = useCallback((time: number) => {
    const video = videoRef.current;
    if (!video) return;
    video.currentTime = time;
    setCurrentTime(time);
  }, []);

  const handleSpeedChange = useCallback((speed: number) => {
    const video = videoRef.current;
    if (!video) return;
    video.playbackRate = speed;
    setPlaybackSpeed(speed);
  }, []);

  const handleToggleFullscreen = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    if (!document.fullscreenElement) {
      container.requestFullscreen?.().catch(() => {});
      setIsFullscreen(true);
    } else {
      document.exitFullscreen?.().catch(() => {});
      setIsFullscreen(false);
    }
  }, []);

  // Listen to fullscreen changes
  useEffect(() => {
    const onFsChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", onFsChange);
    return () => document.removeEventListener("fullscreenchange", onFsChange);
  }, []);

  // Time and progress updates
  const handleTimeUpdate = () => {
    const video = videoRef.current;
    if (!video) return;
    setCurrentTime(video.currentTime);

    // Calculate buffered range
    if (video.buffered.length > 0 && video.duration > 0) {
      const bufferedEnd = video.buffered.end(video.buffered.length - 1);
      setBuffered(bufferedEnd / video.duration);
    }
  };

  return (
    <div
      ref={containerRef}
      onMouseMove={handleMouseMove}
      onMouseLeave={() => isPlaying && setShowControls(false)}
      onClick={(e) => {
        // If clicking video surface in feed mode, open viewer or toggle play
        if (mode === "in_feed" && onOpenViewer) {
          e.stopPropagation();
          onOpenViewer();
        } else {
          handleTogglePlay();
        }
      }}
      className={`relative w-full overflow-hidden bg-black flex items-center justify-center select-none group ${
        mode === "viewer" ? "h-full" : "rounded-2xl"
      }`}
      style={mode === "in_feed" ? { aspectRatio } : undefined}
    >
      {/* Video Element */}
      <video
        ref={videoRef}
        src={src}
        poster={poster}
        playsInline
        preload="metadata"
        muted={isMuted}
        onPlay={() => {
          setIsPlaying(true);
          resetHideControlsTimer();
        }}
        onPause={() => {
          setIsPlaying(false);
          setShowControls(true);
        }}
        onWaiting={() => setIsBuffering(true)}
        onPlaying={() => setIsBuffering(false)}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={() => {
          if (videoRef.current) {
            setDuration(videoRef.current.duration);
          }
        }}
        onError={() => {
          setError(true);
          setIsBuffering(false);
        }}
        className={`w-full h-full ${
          mode === "viewer" ? "object-contain max-h-[90vh]" : "object-cover"
        }`}
      />

      {/* Buffering Indicator */}
      {isBuffering && !error && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/30 pointer-events-none z-10">
          <Loader2 className="w-10 h-10 animate-spin text-brand" />
        </div>
      )}

      {/* Big Play Button Overlay when Paused */}
      {!isPlaying && !isBuffering && !error && (
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none z-10">
          <div className="w-16 h-16 rounded-full bg-brand/90 hover:bg-brand text-white flex items-center justify-center shadow-2xl transition-transform transform group-hover:scale-110">
            <Play className="w-7 h-7 fill-white text-white ml-1" />
          </div>
        </div>
      )}

      {/* Error Fallback */}
      {error && (
        <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/80 text-white p-4 z-10">
          <AlertCircle className="w-10 h-10 text-red-500 mb-2" />
          <p className="text-sm font-semibold">Video could not be loaded</p>
          <button
            onClick={(e) => {
              e.stopPropagation();
              setError(false);
              videoRef.current?.load();
            }}
            className="mt-2 text-xs font-bold px-3 py-1 bg-white/20 hover:bg-white/30 rounded-full transition-colors"
          >
            Retry
          </button>
        </div>
      )}

      {/* Custom Controls Bar */}
      <div
        className={`transition-opacity duration-200 ${
          showControls || !isPlaying ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
      >
        <VideoControls
          isPlaying={isPlaying}
          isMuted={isMuted}
          volume={volume}
          currentTime={currentTime}
          duration={duration}
          buffered={buffered}
          isFullscreen={isFullscreen}
          playbackSpeed={playbackSpeed}
          onTogglePlay={handleTogglePlay}
          onToggleMute={handleToggleMute}
          onVolumeChange={handleVolumeChange}
          onSeek={handleSeek}
          onToggleFullscreen={handleToggleFullscreen}
          onSpeedChange={handleSpeedChange}
        />
      </div>
    </div>
  );
}
