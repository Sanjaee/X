"use client";

import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from "react";
import { Message, Notification, Conversation } from "@/types";
import { apiRequest } from "@/lib/api";

interface TypingEvent {
  conversation_id: string;
  user_id: string;
  is_typing: boolean;
}

interface RealtimeContextType {
  isConnected: boolean;
  sendWS: (type: string, payload: any) => void;
  subscribeToConversation: (convId: string) => void;
  unsubscribeFromConversation: (convId: string) => void;
  sendTyping: (convId: string, isTyping: boolean) => void;
  typingUsers: Record<string, boolean>; // convId: true/false
  lastMessage: Message | null;
  unreadNotifCount: number;
  setUnreadNotifCount: React.Dispatch<React.SetStateAction<number>>;
  unreadMessageCount: number;
  setUnreadMessageCount: React.Dispatch<React.SetStateAction<number>>;
  refreshCounts: () => Promise<void>;
}

const RealtimeContext = createContext<RealtimeContextType | null>(null);

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
  const [typingUsers, setTypingUsers] = useState<Record<string, boolean>>({});
  const [lastMessage, setLastMessage] = useState<Message | null>(null);
  const [unreadNotifCount, setUnreadNotifCount] = useState<number>(0);
  const [unreadMessageCount, setUnreadMessageCount] = useState<number>(0);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const backoffRef = useRef<number>(1000);

  // Refresh counts from REST API
  const refreshCounts = useCallback(async () => {
    if (typeof window === "undefined") return;
    const token = localStorage.getItem("x_session_token");
    if (!token) {
      setUnreadNotifCount(0);
      setUnreadMessageCount(0);
      return;
    }

    try {
      const [notifRes, convsRes] = await Promise.allSettled([
        apiRequest<{ unread_count: number }>("/notifications/unread-count"),
        apiRequest<{ conversations: Conversation[] }>("/conversations"),
      ]);

      if (notifRes.status === "fulfilled" && typeof notifRes.value.unread_count === "number") {
        setUnreadNotifCount(notifRes.value.unread_count);
      }

      if (convsRes.status === "fulfilled" && Array.isArray(convsRes.value.conversations)) {
        const totalUnreadMsg = convsRes.value.conversations.reduce(
          (acc, c) => acc + (c.unread_count || 0),
          0
        );
        setUnreadMessageCount(totalUnreadMsg);
      }
    } catch (e) {
      console.warn("Failed to refresh unread counts", e);
    }
  }, []);

  const connect = useCallback(() => {
    if (typeof window === "undefined") return;

    const token = localStorage.getItem("x_session_token");
    if (!token) {
      setIsConnected(false);
      return;
    }

    const host = window.location.hostname;
    const port = window.location.port ? `:${window.location.port}` : "";
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${host}${port}/api/v1/ws?token=${encodeURIComponent(token)}`;

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        backoffRef.current = 1000;
        refreshCounts();

        // Heartbeat
        const pingInterval = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: "ping" }));
          } else {
            clearInterval(pingInterval);
          }
        }, 25000);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "message.created") {
            setLastMessage(data.payload as Message);
            setUnreadMessageCount((prev) => prev + 1);
          } else if (data.type === "typing.updated") {
            const typingData = data.payload as TypingEvent;
            setTypingUsers((prev) => ({
              ...prev,
              [typingData.conversation_id]: typingData.is_typing,
            }));
          } else if (data.type === "notification.created") {
            setUnreadNotifCount((prev) => prev + 1);
          } else if (data.type === "notification.count.updated") {
            if (typeof data.payload?.unread_count === "number") {
              setUnreadNotifCount(data.payload.unread_count);
            }
          }
        } catch (e) {
          console.error("Failed to parse websocket message", e);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        wsRef.current = null;
        // Exponential backoff reconnect
        const nextDelay = Math.min(backoffRef.current * 1.5, 30000);
        backoffRef.current = nextDelay;
        reconnectTimeoutRef.current = setTimeout(connect, nextDelay);
      };

      ws.onerror = (err) => {
        console.warn("WebSocket error:", err);
        ws.close();
      };
    } catch (e) {
      console.error("Failed to establish WebSocket connection", e);
    }
  }, [refreshCounts]);

  // Initial connect and listeners for auth state changes
  useEffect(() => {
    connect();

    const handleStorage = (e: StorageEvent) => {
      if (e.key === "x_session_token") {
        if (wsRef.current) {
          wsRef.current.close();
        }
        connect();
        refreshCounts();
      }
    };

    const handleAuthChange = () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      connect();
      refreshCounts();
    };

    window.addEventListener("storage", handleStorage);
    window.addEventListener("x_auth_state_change", handleAuthChange);

    return () => {
      window.removeEventListener("storage", handleStorage);
      window.removeEventListener("x_auth_state_change", handleAuthChange);
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      if (wsRef.current) wsRef.current.close();
    };
  }, [connect, refreshCounts]);

  // Standard X Tab Title synchronization: "(N) X" or "(N) Current Title"
  useEffect(() => {
    if (typeof document === "undefined") return;

    const totalUnread = unreadNotifCount + unreadMessageCount;
    const cleanTitle = document.title.replace(/^\(\d+\)\s*/, "");

    if (totalUnread > 0) {
      document.title = `(${totalUnread}) ${cleanTitle || "X"}`;
    } else {
      document.title = cleanTitle || "X";
    }
  }, [unreadNotifCount, unreadMessageCount]);

  const sendWS = useCallback((type: string, payload: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    }
  }, []);

  const subscribeToConversation = useCallback((convId: string) => {
    sendWS("subscribe", { conversation_id: convId });
  }, [sendWS]);

  const unsubscribeFromConversation = useCallback((convId: string) => {
    sendWS("unsubscribe", { conversation_id: convId });
  }, [sendWS]);

  const sendTyping = useCallback((convId: string, isTyping: boolean) => {
    sendWS("typing", { conversation_id: convId, is_typing: isTyping });
  }, [sendWS]);

  return (
    <RealtimeContext.Provider
      value={{
        isConnected,
        sendWS,
        subscribeToConversation,
        unsubscribeFromConversation,
        sendTyping,
        typingUsers,
        lastMessage,
        unreadNotifCount,
        setUnreadNotifCount,
        unreadMessageCount,
        setUnreadMessageCount,
        refreshCounts,
      }}
    >
      {children}
    </RealtimeContext.Provider>
  );
}

export function useRealtime() {
  const context = useContext(RealtimeContext);
  if (!context) {
    throw new Error("useRealtime must be used within a RealtimeProvider");
  }
  return context;
}
