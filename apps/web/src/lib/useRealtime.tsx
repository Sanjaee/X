"use client";

import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from "react";
import Link from "next/link";
import { Message, Notification, Conversation, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { Bell, Mail, UserPlus, UserCheck, Heart, Repeat2, MessageCircle, X } from "lucide-react";

interface TypingEvent {
  conversation_id: string;
  user_id: string;
  is_typing: boolean;
}

export interface ToastNotification {
  id: string;
  type: "message" | "follow" | "follow_request" | "like" | "repost" | "reply" | "system";
  title: string;
  message: string;
  avatar?: string;
  link?: string;
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
  addToast: (t: Omit<ToastNotification, "id">) => void;
}

const RealtimeContext = createContext<RealtimeContextType | null>(null);

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
  const [typingUsers, setTypingUsers] = useState<Record<string, boolean>>({});
  const [lastMessage, setLastMessage] = useState<Message | null>(null);
  const [unreadNotifCount, setUnreadNotifCount] = useState<number>(0);
  const [unreadMessageCount, setUnreadMessageCount] = useState<number>(0);
  const [toasts, setToasts] = useState<ToastNotification[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const backoffRef = useRef<number>(1000);
  const currentUserIdRef = useRef<string | null>(null);
  const seenMessageIdsRef = useRef<Set<string>>(new Set());

  const addToast = useCallback((t: Omit<ToastNotification, "id">) => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev.slice(-3), { ...t, id }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((item) => item.id !== id));
    }, 5000);
  }, []);

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
      if (!currentUserIdRef.current) {
        apiRequest<User>("/me")
          .then((u) => {
            if (u?.id) currentUserIdRef.current = u.id;
          })
          .catch(() => {});
      }

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

  const retryCountRef = useRef<number>(0);
  const MAX_RETRIES = 3;

  const connect = useCallback(() => {
    if (typeof window === "undefined") return;

    const token = localStorage.getItem("x_session_token");
    if (!token) {
      setIsConnected(false);
      return;
    }

    if (retryCountRef.current >= MAX_RETRIES) {
      return;
    }

    let wsUrl = "";
    if (typeof window !== "undefined") {
      const host = window.location.hostname;
      const port = window.location.port ? `:${window.location.port}` : "";
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      if (host === "localhost" && !window.location.port) {
        wsUrl = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(token)}`;
      }
    }

    if (!wsUrl && process.env.NEXT_PUBLIC_API_URL) {
      try {
        const parsed = new URL(process.env.NEXT_PUBLIC_API_URL, window.location.origin);
        const protocol = parsed.protocol === "https:" ? "wss:" : "ws:";
        const base = `${protocol}//${parsed.host}${parsed.pathname.replace(/\/+$/, "")}`;
        wsUrl = `${base}/ws?token=${encodeURIComponent(token)}`;
      } catch {
        // Fallback
      }
    }

    if (!wsUrl) {
      const host = window.location.hostname;
      const port = window.location.port ? `:${window.location.port}` : "";
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsUrl = `${protocol}//${host}${port}/api/v1/ws?token=${encodeURIComponent(token)}`;
    }

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        retryCountRef.current = 0;
        backoffRef.current = 1000;
        refreshCounts();

        // Heartbeat
        const pingInterval = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ event_type: "ping", type: "ping" }));
          } else {
            clearInterval(pingInterval);
          }
        }, 25000);
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          const type = data.event_type || data.type;
          const payload = data.payload || data;

          if (type === "message.created") {
            const msg = payload as Message;
            setLastMessage(msg);

            // Deduplicate count increment
            if (msg.id && !seenMessageIdsRef.current.has(msg.id)) {
              seenMessageIdsRef.current.add(msg.id);
              // Only increment if message is from another user
              if (!currentUserIdRef.current || msg.sender_id !== currentUserIdRef.current) {
                setUnreadMessageCount((prev) => prev + 1);
                addToast({
                  type: "message",
                  title: msg.sender?.display_name || (msg.sender?.username ? `@${msg.sender.username}` : "New Message"),
                  message: msg.content || "Sent an attachment",
                  avatar: msg.sender?.avatar_url,
                  link: msg.conversation_id ? `/messages/${msg.conversation_id}` : "/messages",
                });
                if (typeof window !== "undefined") {
                  window.dispatchEvent(new CustomEvent("x_new_message", { detail: msg }));
                }
              }
            }
          } else if (type === "typing.updated") {
            const typingData = payload as TypingEvent;
            setTypingUsers((prev) => ({
              ...prev,
              [typingData.conversation_id]: typingData.is_typing,
            }));
          } else if (type === "notification.created") {
            setUnreadNotifCount((prev) => prev + 1);
            refreshCounts();

            const notifType = payload.type;
            let toastTitle = "New Notification";
            let toastMsg = "You have a new alert";
            let toastLink = "/notifications";
            let toastType: ToastNotification["type"] = "system";

            if (notifType === "follow") {
              toastType = "follow";
              toastTitle = "New Follower";
              toastMsg = payload.follower_username ? `@${payload.follower_username} followed you` : "Someone followed you";
              toastLink = payload.follower_username ? `/${payload.follower_username}` : "/notifications";
            } else if (notifType === "follow_request") {
              toastType = "follow_request";
              toastTitle = "Follow Request";
              toastMsg = payload.requester_username ? `@${payload.requester_username} requested to follow you` : "Someone requested to follow you";
              toastLink = "/notifications";
            } else if (notifType === "follow_accept") {
              toastType = "follow";
              toastTitle = "Follow Request Accepted";
              toastMsg = "Your follow request was accepted";
              toastLink = "/notifications";
            } else if (notifType === "dm") {
              toastType = "message";
              toastTitle = payload.sender_name || (payload.sender_username ? `@${payload.sender_username}` : "Direct Message");
              toastMsg = payload.content || "Sent you a direct message";
              toastLink = payload.conversation_id ? `/messages/${payload.conversation_id}` : "/messages";
            } else if (notifType === "like") {
              toastType = "like";
              toastTitle = "New Like";
              toastMsg = "Someone liked your post";
              toastLink = "/notifications";
            } else if (notifType === "repost") {
              toastType = "repost";
              toastTitle = "New Repost";
              toastMsg = "Someone reposted your post";
              toastLink = "/notifications";
            } else if (notifType === "reply") {
              toastType = "reply";
              toastTitle = "New Reply";
              toastMsg = payload.content ? `"${payload.content}"` : "Someone replied to your post";
              toastLink = "/notifications";
            }

            addToast({
              type: toastType,
              title: toastTitle,
              message: toastMsg,
              avatar: payload.follower_avatar || payload.requester_avatar || payload.sender_avatar,
              link: toastLink,
            });

            if (typeof window !== "undefined") {
              window.dispatchEvent(new CustomEvent("x_new_notification", { detail: payload }));
            }
          } else if (type === "notification.count.updated") {
            if (typeof payload?.unread_count === "number") {
              setUnreadNotifCount(payload.unread_count);
            }
          }
        } catch (e) {
          console.error("Failed to parse websocket message", e);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        wsRef.current = null;
        retryCountRef.current += 1;

        if (retryCountRef.current >= MAX_RETRIES) {
          console.warn("WebSocket disconnected: maximum retries reached. Awaiting valid authentication.");
          return;
        }

        // Exponential backoff reconnect
        const nextDelay = Math.min(backoffRef.current * 1.5, 30000);
        backoffRef.current = nextDelay;
        reconnectTimeoutRef.current = setTimeout(connect, nextDelay);
      };

      ws.onerror = () => {
        // ws.onerror is followed by onclose in browser WebSocket lifecycle
      };
    } catch (e) {
      console.warn("Failed to establish WebSocket connection", e);
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
        retryCountRef.current = 0;
        backoffRef.current = 1000;
        connect();
        refreshCounts();
      }
    };

    const handleAuthChange = () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      retryCountRef.current = 0;
      backoffRef.current = 1000;
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
      wsRef.current.send(
        JSON.stringify({
          event_type: type,
          type: type,
          payload,
        })
      );
    }
  }, []);

  const subscribeToConversation = useCallback((convId: string) => {
    sendWS("conversation.subscribe", { conversation_id: convId });
  }, [sendWS]);

  const unsubscribeFromConversation = useCallback((convId: string) => {
    sendWS("conversation.unsubscribe", { conversation_id: convId });
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
        addToast,
      }}
    >
      {children}

      {/* Real-time In-App Floating Toasts */}
      {toasts.length > 0 && (
        <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-2.5 max-w-sm w-full pointer-events-none px-4 sm:px-0">
          {toasts.map((t) => (
            <div
              key={t.id}
              className="pointer-events-auto bg-[#16181c]/95 border border-border/80 hover:border-brand/50 shadow-2xl rounded-2xl p-3.5 backdrop-blur-md flex items-start gap-3 transition-all animate-in slide-in-from-bottom-5 duration-300"
            >
              <div className="shrink-0 pt-0.5">
                {t.type === "message" ? (
                  <div className="w-8 h-8 rounded-full bg-blue-500/15 flex items-center justify-center text-blue-400">
                    <Mail className="w-4 h-4" />
                  </div>
                ) : t.type === "follow" ? (
                  <div className="w-8 h-8 rounded-full bg-brand/15 flex items-center justify-center text-brand">
                    <UserPlus className="w-4 h-4" />
                  </div>
                ) : t.type === "follow_request" ? (
                  <div className="w-8 h-8 rounded-full bg-amber-500/15 flex items-center justify-center text-amber-500">
                    <UserCheck className="w-4 h-4" />
                  </div>
                ) : t.type === "like" ? (
                  <div className="w-8 h-8 rounded-full bg-rose-500/15 flex items-center justify-center text-rose-500">
                    <Heart className="w-4 h-4 fill-rose-500" />
                  </div>
                ) : t.type === "repost" ? (
                  <div className="w-8 h-8 rounded-full bg-emerald-500/15 flex items-center justify-center text-emerald-500">
                    <Repeat2 className="w-4 h-4" />
                  </div>
                ) : (
                  <div className="w-8 h-8 rounded-full bg-brand/15 flex items-center justify-center text-brand">
                    <Bell className="w-4 h-4" />
                  </div>
                )}
              </div>

              <div className="flex-1 min-w-0">
                {t.link ? (
                  <Link href={t.link} className="block group">
                    <p className="text-[14px] font-bold text-foreground group-hover:text-brand transition-colors truncate">
                      {t.title}
                    </p>
                    <p className="text-[13px] text-muted line-clamp-2 mt-0.5">
                      {t.message}
                    </p>
                  </Link>
                ) : (
                  <div>
                    <p className="text-[14px] font-bold text-foreground truncate">{t.title}</p>
                    <p className="text-[13px] text-muted line-clamp-2 mt-0.5">{t.message}</p>
                  </div>
                )}
              </div>

              <button
                onClick={() => setToasts((prev) => prev.filter((item) => item.id !== t.id))}
                className="shrink-0 p-1 text-muted hover:text-foreground rounded-full hover:bg-subtle transition-colors"
                title="Dismiss"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}
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
