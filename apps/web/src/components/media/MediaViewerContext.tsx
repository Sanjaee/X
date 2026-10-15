"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from "react";
import { Post, MediaObject } from "@/types";

interface MediaViewerContextType {
  isOpen: boolean;
  post: Post | null;
  activeIndex: number;
  openMedia: (post: Post, initialIndex?: number) => void;
  closeMedia: () => void;
  nextMedia: () => void;
  prevMedia: () => void;
  setMediaIndex: (index: number) => void;
  activeMedia: MediaObject | null;
  totalMedia: number;
}

const MediaViewerContext = createContext<MediaViewerContextType | undefined>(undefined);

export function MediaViewerProvider({ children }: { children: React.ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const [post, setPost] = useState<Post | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const originPathRef = useRef<string>("/");
  const isNavigatingBackRef = useRef(false);

  const getMediaUrl = useCallback((p: Post, index: number) => {
    const username = p.author?.username || "user";
    const mediaItem = p.media?.[index];
    const type = mediaItem?.media_type === "video" ? "video" : "photo";
    return `/${username}/status/${p.id}/${type}/${index + 1}`;
  }, []);

  const openMedia = useCallback((targetPost: Post, initialIndex = 0) => {
    if (!targetPost.media || targetPost.media.length === 0) return;

    const clampedIndex = Math.max(0, Math.min(initialIndex, targetPost.media.length - 1));
    const currentPath = window.location.pathname + window.location.search;
    originPathRef.current = currentPath;

    setPost(targetPost);
    setActiveIndex(clampedIndex);
    setIsOpen(true);

    // Save anchor for scroll preservation
    try {
      sessionStorage.setItem("x_last_media_anchor", targetPost.id);
      sessionStorage.setItem("x_last_media_origin", currentPath);
    } catch {}

    // Lock body scroll with padding compensation to avoid layout shift
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${scrollbarWidth}px`;
    }

    // Push new history state
    const newUrl = getMediaUrl(targetPost, clampedIndex);
    window.history.pushState(
      { x_media_modal: true, postId: targetPost.id, index: clampedIndex, origin: currentPath },
      "",
      newUrl
    );
  }, [getMediaUrl]);

  const closeMedia = useCallback(() => {
    if (!isOpen) return;

    // Restore body scroll
    document.body.style.overflow = "";
    document.body.style.paddingRight = "";

    // If modal was opened via pushState, use history.back() to revert URL cleanly
    if (window.history.state?.x_media_modal) {
      isNavigatingBackRef.current = true;
      window.history.back();
    } else {
      window.history.replaceState(null, "", originPathRef.current || "/");
    }

    setIsOpen(false);
    setPost(null);
    setActiveIndex(0);

    // Smoothly restore scroll anchor if needed
    try {
      const anchorId = sessionStorage.getItem("x_last_media_anchor");
      if (anchorId) {
        requestAnimationFrame(() => {
          const el = document.querySelector(`[data-post-id="${anchorId}"]`);
          if (el) {
            el.scrollIntoView({ block: "nearest", behavior: "auto" });
          }
        });
      }
    } catch {}
  }, [isOpen]);

  const setMediaIndex = useCallback((index: number) => {
    if (!post || !post.media) return;
    const clamped = Math.max(0, Math.min(index, post.media.length - 1));
    setActiveIndex(clamped);

    // Use replaceState during gallery flips so Back button exits viewer in one press
    const updatedUrl = getMediaUrl(post, clamped);
    window.history.replaceState(
      { x_media_modal: true, postId: post.id, index: clamped, origin: originPathRef.current },
      "",
      updatedUrl
    );
  }, [post, getMediaUrl]);

  const nextMedia = useCallback(() => {
    if (!post || !post.media) return;
    if (activeIndex < post.media.length - 1) {
      setMediaIndex(activeIndex + 1);
    }
  }, [post, activeIndex, setMediaIndex]);

  const prevMedia = useCallback(() => {
    if (activeIndex > 0) {
      setMediaIndex(activeIndex - 1);
    }
  }, [activeIndex, setMediaIndex]);

  // Synchronize with browser Back and Forward buttons (popstate)
  useEffect(() => {
    const handlePopState = (e: PopStateEvent) => {
      if (e.state?.x_media_modal) {
        // Forward back into modal
        setIsOpen(true);
        setActiveIndex(e.state.index || 0);
        document.body.style.overflow = "hidden";
      } else {
        // Browser Back pressed: close viewer without reload
        setIsOpen(false);
        setPost(null);
        document.body.style.overflow = "";
        document.body.style.paddingRight = "";

        // Restore scroll position to anchor element
        try {
          const anchorId = sessionStorage.getItem("x_last_media_anchor");
          if (anchorId) {
            requestAnimationFrame(() => {
              const el = document.querySelector(`[data-post-id="${anchorId}"]`);
              if (el) {
                el.scrollIntoView({ block: "nearest", behavior: "auto" });
              }
            });
          }
        } catch {}
      }
    };

    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const activeMedia = post?.media?.[activeIndex] || null;
  const totalMedia = post?.media?.length || 0;

  return (
    <MediaViewerContext.Provider
      value={{
        isOpen,
        post,
        activeIndex,
        openMedia,
        closeMedia,
        nextMedia,
        prevMedia,
        setMediaIndex,
        activeMedia,
        totalMedia,
      }}
    >
      {children}
    </MediaViewerContext.Provider>
  );
}

export function useMediaViewer() {
  const context = useContext(MediaViewerContext);
  if (!context) {
    throw new Error("useMediaViewer must be used within a MediaViewerProvider");
  }
  return context;
}
