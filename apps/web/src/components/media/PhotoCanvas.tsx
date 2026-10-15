"use client";

import React, { useState, useRef, useEffect, useCallback } from "react";
import { MediaObject } from "@/types";
import { Loader2, AlertCircle, ZoomIn, ZoomOut, RotateCcw } from "lucide-react";

interface PhotoCanvasProps {
  media: MediaObject;
  onDoubleTap?: () => void;
}

export function PhotoCanvas({ media }: PhotoCanvasProps) {
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);
  const imgRef = useRef<HTMLImageElement>(null);

  // Reset zoom and pan when media changes
  useEffect(() => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
    setLoading(true);
    setError(false);
  }, [media.url]);

  const handleZoomIn = useCallback(() => {
    setScale((prev) => Math.min(prev + 0.5, 4));
  }, []);

  const handleZoomOut = useCallback(() => {
    setScale((prev) => {
      const next = Math.max(prev - 0.5, 1);
      if (next === 1) setPosition({ x: 0, y: 0 });
      return next;
    });
  }, []);

  const handleResetZoom = useCallback(() => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
  }, []);

  const handleDoubleClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setScale((prev) => {
      if (prev > 1) {
        setPosition({ x: 0, y: 0 });
        return 1;
      }
      return 2.5;
    });
  }, []);

  // Mouse pan handlers
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (scale <= 1) return;
    e.preventDefault();
    setIsDragging(true);
    setDragStart({ x: e.clientX - position.x, y: e.clientY - position.y });
  }, [scale, position]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging || scale <= 1) return;
    const maxOffset = (scale - 1) * 300;
    const newX = Math.max(-maxOffset, Math.min(maxOffset, e.clientX - dragStart.x));
    const newY = Math.max(-maxOffset, Math.min(maxOffset, e.clientY - dragStart.y));
    setPosition({ x: newX, y: newY });
  }, [isDragging, scale, dragStart]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  // Touch pan handlers for mobile
  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (scale <= 1 || e.touches.length !== 1) return;
    setIsDragging(true);
    const touch = e.touches[0];
    setDragStart({ x: touch.clientX - position.x, y: touch.clientY - position.y });
  }, [scale, position]);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isDragging || scale <= 1 || e.touches.length !== 1) return;
    const touch = e.touches[0];
    const maxOffset = (scale - 1) * 300;
    const newX = Math.max(-maxOffset, Math.min(maxOffset, touch.clientX - dragStart.x));
    const newY = Math.max(-maxOffset, Math.min(maxOffset, touch.clientY - dragStart.y));
    setPosition({ x: newX, y: newY });
  }, [isDragging, scale, dragStart]);

  return (
    <div
      ref={containerRef}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleMouseUp}
      onDoubleClick={handleDoubleClick}
      className={`relative w-full h-full flex items-center justify-center select-none overflow-hidden ${
        scale > 1 ? (isDragging ? "cursor-grabbing" : "cursor-grab") : "cursor-zoom-in"
      }`}
    >
      {/* Loading Spinner */}
      {loading && !error && (
        <div className="absolute inset-0 flex items-center justify-center z-10">
          <Loader2 className="w-10 h-10 animate-spin text-brand" />
        </div>
      )}

      {/* Error Fallback */}
      {error && (
        <div className="flex flex-col items-center justify-center text-center p-6 text-muted z-10">
          <AlertCircle className="w-12 h-12 text-red-500 mb-2" />
          <p className="text-foreground font-semibold text-base mb-1">Failed to load image</p>
          <button
            onClick={() => {
              setError(false);
              setLoading(true);
            }}
            className="mt-3 px-4 py-2 bg-subtle hover:bg-muted/40 rounded-full text-sm font-bold text-foreground transition-colors"
          >
            Try Again
          </button>
        </div>
      )}

      {/* Photo Image */}
      <img
        ref={imgRef}
        src={media.url}
        alt="Photo preview"
        loading="eager"
        decoding="async"
        onLoad={() => setLoading(false)}
        onError={() => {
          setLoading(false);
          setError(true);
        }}
        style={{
          transform: `translate(${position.x}px, ${position.y}px) scale(${scale})`,
          transition: isDragging ? "none" : "transform 0.15s ease-out",
          maxHeight: "92vh",
          maxWidth: "92vw",
        }}
        className={`object-contain transition-opacity duration-200 pointer-events-auto ${
          loading ? "opacity-0" : "opacity-100"
        }`}
      />

      {/* Floating Zoom Controls Bar (Desktop & Tablet) */}
      <div
        onClick={(e) => e.stopPropagation()}
        className="absolute bottom-6 left-1/2 -translate-x-1/2 bg-black/60 backdrop-blur-md border border-white/10 rounded-full px-3 py-1.5 flex items-center gap-2 text-white shadow-2xl z-20 transition-opacity hover:opacity-100 opacity-80"
      >
        <button
          onClick={handleZoomOut}
          disabled={scale <= 1}
          title="Zoom out"
          aria-label="Zoom out"
          className="p-1.5 hover:bg-white/10 rounded-full disabled:opacity-30 disabled:hover:bg-transparent transition-colors"
        >
          <ZoomOut className="w-4 h-4" />
        </button>
        <span className="text-xs font-semibold px-1 min-w-[36px] text-center">
          {Math.round(scale * 100)}%
        </span>
        <button
          onClick={handleZoomIn}
          disabled={scale >= 4}
          title="Zoom in"
          aria-label="Zoom in"
          className="p-1.5 hover:bg-white/10 rounded-full disabled:opacity-30 disabled:hover:bg-transparent transition-colors"
        >
          <ZoomIn className="w-4 h-4" />
        </button>
        {scale > 1 && (
          <button
            onClick={handleResetZoom}
            title="Reset zoom"
            aria-label="Reset zoom"
            className="p-1.5 hover:bg-white/10 rounded-full transition-colors ml-1 border-l border-white/20 pl-2"
          >
            <RotateCcw className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
}
