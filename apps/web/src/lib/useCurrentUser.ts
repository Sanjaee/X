"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useCallback } from "react";
import { User } from "@/types";
import {
  apiRequest,
  getStoredToken,
  getStoredUser,
  setStoredToken,
  setStoredUser,
} from "@/lib/api";

export function useCurrentUser() {
  const queryClient = useQueryClient();

  const { data: currentUser = null, isLoading, refetch } = useQuery<User | null>({
    queryKey: ["currentUser"],
    queryFn: async () => {
      const token = getStoredToken();
      if (!token) {
        setStoredUser(null);
        return null;
      }
      try {
        const user = await apiRequest<User>("/me");
        setStoredUser(user);
        return user;
      } catch (err) {
        setStoredUser(null);
        return null;
      }
    },
    initialData: () => {
      return getStoredUser() as User | null;
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  useEffect(() => {
    const handleAuthChange = () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
    };
    window.addEventListener("x_auth_state_change", handleAuthChange);
    return () => window.removeEventListener("x_auth_state_change", handleAuthChange);
  }, [queryClient]);

  const logout = useCallback(async () => {
    try {
      await apiRequest("/auth/logout", { method: "POST" });
    } catch {}
    try {
      const { signOut } = await import("next-auth/react");
      await signOut({ redirect: false });
    } catch {}
    setStoredToken(null);
    setStoredUser(null);
    queryClient.setQueryData(["currentUser"], null);
    queryClient.clear();
    window.dispatchEvent(new Event("x_auth_state_change"));
  }, [queryClient]);

  const updateUser = useCallback(
    (updated: User | null) => {
      setStoredUser(updated);
      queryClient.setQueryData(["currentUser"], updated);
    },
    [queryClient]
  );

  return {
    currentUser,
    isLoading: isLoading && !currentUser,
    logout,
    updateUser,
    refetchUser: refetch,
  };
}
