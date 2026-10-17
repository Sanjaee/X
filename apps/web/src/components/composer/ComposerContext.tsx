"use client";

import React, { createContext, useContext, useState, useEffect } from "react";
import { User, Post } from "@/types";
import { apiRequest } from "@/lib/api";
import { PostModal } from "./PostModal";
import { useQueryClient } from "@tanstack/react-query";
import { AuthModal } from "@/components/auth/AuthModal";

interface ComposerContextValue {
  isOpen: boolean;
  openComposer: () => void;
  closeComposer: () => void;
}

const ComposerContext = createContext<ComposerContextValue | undefined>(undefined);

export function ComposerProvider({ children }: { children: React.ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const queryClient = useQueryClient();

  useEffect(() => {
    const syncUser = () => {
      const token = typeof window !== "undefined" ? localStorage.getItem("x_session_token") : null;
      if (token) {
        apiRequest<User>("/me")
          .then((u) => setCurrentUser(u))
          .catch(() => setCurrentUser(null));
      } else {
        setCurrentUser(null);
      }
    };

    syncUser();
    window.addEventListener("x_auth_state_change", syncUser);
    return () => window.removeEventListener("x_auth_state_change", syncUser);
  }, []);

  const openComposer = () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    setIsOpen(true);
  };

  const closeComposer = () => setIsOpen(false);

  const handlePostSuccess = (newPost: Post) => {
    queryClient.invalidateQueries({ queryKey: ["feed"] });
    queryClient.invalidateQueries({ queryKey: ["posts"] });
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("x_post_created", { detail: newPost }));
    }
  };

  return (
    <ComposerContext.Provider value={{ isOpen, openComposer, closeComposer }}>
      {children}
      <PostModal
        isOpen={isOpen}
        onClose={closeComposer}
        currentUser={currentUser}
        onPostSuccess={handlePostSuccess}
        onOpenAuth={() => setAuthModalOpen(true)}
      />
      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        onSuccess={() => {
          setAuthModalOpen(false);
          apiRequest<User>("/me")
            .then((u) => setCurrentUser(u))
            .catch(() => {});
        }}
      />
    </ComposerContext.Provider>
  );
}

export function useComposer(): ComposerContextValue {
  const ctx = useContext(ComposerContext);
  if (!ctx) {
    return {
      isOpen: false,
      openComposer: () => {},
      closeComposer: () => {},
    };
  }
  return ctx;
}
