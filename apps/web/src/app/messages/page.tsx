"use client";

import React, { useEffect, useState, useRef, useCallback } from "react";
import Link from "next/link";
import {
  Mail,
  Settings,
  Send,
  Image as ImageIcon,
  Loader2,
  Check,
  CheckCheck,
  Plus,
  ArrowLeft,
  X,
  AlertCircle,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { AuthModal } from "@/components/auth/AuthModal";
import { Conversation, Message, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useRealtime } from "@/lib/useRealtime";

export default function MessagesPage() {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConvId, setActiveConvId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [messageInput, setMessageInput] = useState("");
  const [loadingConvs, setLoadingConvs] = useState(true);
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [sendingMessage, setSendingMessage] = useState(false);
  const [newChatModalOpen, setNewChatModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [searching, setSearching] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isTypingRef = useRef(false);

  const {
    subscribeToConversation,
    unsubscribeFromConversation,
    sendTyping,
    typingUsers,
    lastMessage,
  } = useRealtime();

  // Load current user
  useEffect(() => {
    apiRequest<User>("/me")
      .then((u) => setCurrentUser(u))
      .catch(() => setCurrentUser(null));
  }, []);

  // Load conversations
  const loadConversations = useCallback(async () => {
    try {
      setLoadingConvs(true);
      const res = await apiRequest<{ conversations: Conversation[] }>("/conversations");
      setConversations(res.conversations || []);
      if (!activeConvId && res.conversations && res.conversations.length > 0) {
        setActiveConvId(res.conversations[0].id);
      }
    } catch (e) {
      console.error("Failed to load conversations", e);
    } finally {
      setLoadingConvs(false);
    }
  }, [activeConvId]);

  useEffect(() => {
    if (currentUser) {
      loadConversations();
    } else {
      setLoadingConvs(false);
    }
  }, [currentUser, loadConversations]);

  // Load messages for active conversation & subscribe to realtime
  useEffect(() => {
    if (!activeConvId) return;

    subscribeToConversation(activeConvId);
    setLoadingMessages(true);
    setErrorMessage(null);

    apiRequest<{ messages: Message[] }>(`/conversations/${activeConvId}/messages?limit=50`)
      .then((res) => {
        setMessages(res.messages || []);
        // Mark conversation as read
        apiRequest(`/conversations/${activeConvId}/read`, { method: "POST" }).catch(() => {});
      })
      .catch((err) => {
        console.error("Failed to load messages", err);
        setErrorMessage(err.message || "Failed to load messages");
      })
      .finally(() => {
        setLoadingMessages(false);
      });

    return () => {
      unsubscribeFromConversation(activeConvId);
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    };
  }, [activeConvId, subscribeToConversation, unsubscribeFromConversation]);

  // Handle incoming realtime message
  useEffect(() => {
    if (!lastMessage) return;

    // If message is for currently active conversation
    if (lastMessage.conversation_id === activeConvId) {
      setMessages((prev) => {
        // Prevent duplicate if already in state (e.g. from optimistic send)
        if (prev.some((m) => m.id === lastMessage.id || (m.client_message_id && m.client_message_id === lastMessage.client_message_id))) {
          return prev.map((m) =>
            m.client_message_id === lastMessage.client_message_id ? lastMessage : m
          );
        }
        return [...prev, lastMessage];
      });

      // Mark as read
      apiRequest(`/conversations/${activeConvId}/read`, { method: "POST" }).catch(() => {});
    }

    // Update conversation preview in sidebar
    setConversations((prev) =>
      prev.map((c) =>
        c.id === lastMessage.conversation_id
          ? { ...c, last_message: lastMessage, updated_at: lastMessage.created_at }
          : c
      )
    );
  }, [lastMessage, activeConvId]);

  // Auto-scroll to bottom of chat
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // Typing debounce handler
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMessageInput(e.target.value);
    if (!activeConvId) return;

    if (!isTypingRef.current) {
      isTypingRef.current = true;
      sendTyping(activeConvId, true);
    }

    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    typingTimeoutRef.current = setTimeout(() => {
      isTypingRef.current = false;
      sendTyping(activeConvId, false);
    }, 2500);
  };

  // Send message
  const handleSendMessage = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!activeConvId || !messageInput.trim() || sendingMessage) return;

    const content = messageInput.trim();
    const clientMsgId = "client_" + Date.now() + "_" + Math.random().toString(36).substring(2, 9);
    setMessageInput("");
    setErrorMessage(null);

    // Optimistic message
    const optimisticMsg: Message = {
      id: clientMsgId,
      conversation_id: activeConvId,
      sender_id: currentUser?.id || "",
      sender: currentUser || undefined,
      client_message_id: clientMsgId,
      content,
      message_type: "text",
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimisticMsg]);

    try {
      setSendingMessage(true);
      const savedMsg = await apiRequest<Message>(`/conversations/${activeConvId}/messages`, {
        method: "POST",
        body: JSON.stringify({
          content,
          client_message_id: clientMsgId,
        }),
      });

      // Reconcile optimistic message
      setMessages((prev) =>
        prev.map((m) => (m.client_message_id === clientMsgId ? savedMsg : m))
      );
    } catch (err: any) {
      setErrorMessage(err.message || "Failed to send message");
      // Remove or mark failed
    } finally {
      setSendingMessage(false);
    }
  };

  // Search user to start new chat
  const handleSearchUsers = async (query: string) => {
    setSearchQuery(query);
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    setSearching(true);
    try {
      const res = await apiRequest<{ users: User[] }>(`/search?type=users&q=${encodeURIComponent(query)}`);
      setSearchResults(res.users || []);
    } catch (e) {
      console.error("Search failed", e);
    } finally {
      setSearching(false);
    }
  };

  // Start new conversation
  const startConversationWith = async (targetUser: User) => {
    try {
      const conv = await apiRequest<Conversation>("/conversations", {
        method: "POST",
        body: JSON.stringify({ recipient_id: targetUser.id }),
      });
      conv.other_member = targetUser;
      setConversations((prev) => {
        const exists = prev.find((c) => c.id === conv.id);
        if (exists) return prev;
        return [conv, ...prev];
      });
      setActiveConvId(conv.id);
      setNewChatModalOpen(false);
      setSearchQuery("");
      setSearchResults([]);
    } catch (e: any) {
      setErrorMessage(e.message || "Could not start conversation");
    }
  };

  const activeConv = conversations.find((c) => c.id === activeConvId);

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        <LeftNavRail
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        {/* 2-Column Messages Container */}
        <div className="flex-1 flex border-r border-border min-h-screen">
          {/* Conversation List Column */}
          <div className="w-full sm:w-[380px] border-r border-border flex flex-col">
            <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center justify-between px-4 h-[53px]">
              <h1 className="font-extrabold text-[20px]">Messages</h1>
              <div className="flex items-center gap-1">
                {currentUser && (
                  <button
                    onClick={() => setNewChatModalOpen(true)}
                    title="New message"
                    className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
                  >
                    <Plus className="w-5 h-5" />
                  </button>
                )}
                <button className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors">
                  <Settings className="w-5 h-5" />
                </button>
              </div>
            </header>

            <div className="flex-1 flex flex-col overflow-y-auto divide-y divide-border">
              {loadingConvs ? (
                <div className="flex justify-center p-8 text-brand">
                  <Loader2 className="w-6 h-6 animate-spin" />
                </div>
              ) : !currentUser ? (
                <div className="text-center py-12 px-4">
                  <h3 className="font-bold text-[18px] mb-2">Welcome to your inbox!</h3>
                  <p className="text-muted text-[14px] mb-4">
                    Drop a line, share posts, and more with private conversations.
                  </p>
                  <button
                    onClick={() => setAuthModalOpen(true)}
                    className="bg-brand text-white font-bold py-2 px-5 rounded-full text-[14px]"
                  >
                    Log in
                  </button>
                </div>
              ) : conversations.length === 0 ? (
                <div className="text-center py-12 px-4">
                  <h3 className="font-bold text-[18px] mb-2">No messages yet</h3>
                  <p className="text-muted text-[14px] mb-4">
                    Send a message to start a conversation with anyone on X.
                  </p>
                  <button
                    onClick={() => setNewChatModalOpen(true)}
                    className="bg-brand text-white font-bold py-2 px-5 rounded-full text-[14px]"
                  >
                    Write a message
                  </button>
                </div>
              ) : (
                conversations.map((c) => {
                  const other = c.other_member;
                  const isSelected = c.id === activeConvId;
                  return (
                    <div
                      key={c.id}
                      onClick={() => setActiveConvId(c.id)}
                      className={cn(
                        "flex items-center gap-3 p-4 hover:bg-subtle/50 cursor-pointer transition-colors relative",
                        isSelected && "bg-subtle/70"
                      )}
                    >
                      <img
                        src={
                          other?.avatar_url ||
                          "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                        }
                        alt={other?.display_name || "User"}
                        className="w-12 h-12 rounded-full object-cover border border-border shrink-0"
                      />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-[15px] truncate">
                            {other?.display_name || other?.username || "Conversation"}
                          </span>
                        </div>
                        <span className="text-[14px] text-muted truncate block">
                          @{other?.username || "user"}
                        </span>
                        <p className="text-[14px] text-muted truncate mt-0.5">
                          {c.last_message?.content || "No messages yet"}
                        </p>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* Active Chat Column */}
          <div className="hidden sm:flex flex-1 flex-col justify-between h-screen">
            {activeConv ? (
              <>
                {/* Chat Header */}
                <div className="h-[53px] border-b border-border px-4 flex items-center justify-between bg-background/80 backdrop-blur-md">
                  <div className="flex items-center gap-3">
                    <img
                      src={
                        activeConv.other_member?.avatar_url ||
                        "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                      }
                      alt="Avatar"
                      className="w-8 h-8 rounded-full object-cover"
                    />
                    <div>
                      <Link
                        href={`/${activeConv.other_member?.username || ""}`}
                        className="font-bold text-[16px] hover:underline block leading-tight"
                      >
                        {activeConv.other_member?.display_name || activeConv.other_member?.username}
                      </Link>
                      <span className="text-[12px] text-muted">
                        @{activeConv.other_member?.username}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Error Banner */}
                {errorMessage && (
                  <div className="bg-red-500/10 border-b border-red-500/20 px-4 py-2 flex items-center gap-2 text-red-500 text-sm">
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span>{errorMessage}</span>
                  </div>
                )}

                {/* Chat Body */}
                <div className="flex-1 p-4 overflow-y-auto flex flex-col gap-3">
                  {loadingMessages ? (
                    <div className="flex justify-center p-8 text-brand">
                      <Loader2 className="w-6 h-6 animate-spin" />
                    </div>
                  ) : messages.length === 0 ? (
                    <div className="text-center py-12 text-muted">
                      <p>Say hello to @{activeConv.other_member?.username}!</p>
                    </div>
                  ) : (
                    messages.map((m) => {
                      const isMe = m.sender_id === currentUser?.id;
                      return (
                        <div
                          key={m.id}
                          className={cn(
                            "max-w-[70%] p-3 text-[15px] flex flex-col gap-1",
                            isMe
                              ? "self-end bg-brand text-white rounded-2xl rounded-tr-sm"
                              : "self-start bg-card text-foreground rounded-2xl rounded-tl-sm border border-border"
                          )}
                        >
                          <p className="whitespace-pre-wrap break-words">{m.content}</p>
                          <div
                            className={cn(
                              "text-[11px] self-end flex items-center gap-1",
                              isMe ? "text-white/80" : "text-muted"
                            )}
                          >
                            <span>
                              {new Date(m.created_at).toLocaleTimeString([], {
                                hour: "2-digit",
                                minute: "2-digit",
                              })}
                            </span>
                            {isMe && <Check className="w-3.5 h-3.5" />}
                          </div>
                        </div>
                      );
                    })
                  )}

                  {/* Realtime Ephemeral Typing Indicator */}
                  {typingUsers[activeConv.id] && (
                    <div className="self-start bg-card border border-border px-3 py-2 rounded-2xl rounded-tl-sm text-xs text-muted flex items-center gap-1.5 animate-pulse">
                      <span>{activeConv.other_member?.display_name || "User"} is typing</span>
                      <span className="flex gap-0.5">
                        <span className="w-1 h-1 bg-muted rounded-full animate-bounce" />
                        <span className="w-1 h-1 bg-muted rounded-full animate-bounce [animation-delay:0.2s]" />
                        <span className="w-1 h-1 bg-muted rounded-full animate-bounce [animation-delay:0.4s]" />
                      </span>
                    </div>
                  )}

                  <div ref={messagesEndRef} />
                </div>

                {/* Input Area */}
                <form
                  onSubmit={handleSendMessage}
                  className="p-3 border-t border-border flex items-center gap-2 bg-background"
                >
                  <button
                    type="button"
                    className="p-2 hover:bg-subtle text-brand rounded-full transition-colors"
                  >
                    <ImageIcon className="w-5 h-5" />
                  </button>
                  <input
                    type="text"
                    placeholder="Start a new message"
                    value={messageInput}
                    onChange={handleInputChange}
                    className="flex-1 bg-card rounded-full py-2 px-4 outline-none text-[15px] border border-transparent focus:border-brand transition-colors text-foreground"
                  />
                  <button
                    type="submit"
                    disabled={!messageInput.trim() || sendingMessage}
                    className="p-2 hover:bg-subtle text-brand disabled:opacity-40 rounded-full transition-colors"
                  >
                    <Send className="w-5 h-5" />
                  </button>
                </form>
              </>
            ) : (
              <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-muted">
                <Mail className="w-12 h-12 mb-3 text-muted/60" />
                <h3 className="font-bold text-[20px] text-foreground mb-1">
                  Select a message
                </h3>
                <p className="max-w-xs text-sm">
                  Choose from your existing conversations, start a new one, or just keep swimming.
                </p>
                {currentUser && (
                  <button
                    onClick={() => setNewChatModalOpen(true)}
                    className="mt-4 bg-brand text-white font-bold py-2.5 px-6 rounded-full text-sm hover:bg-brand-hover transition-colors"
                  >
                    New message
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

        {/* New Chat Modal */}
        {newChatModalOpen && (
          <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
            <div className="bg-background border border-border rounded-2xl w-full max-w-lg max-h-[80vh] flex flex-col shadow-2xl">
              <div className="flex items-center justify-between p-4 border-b border-border">
                <h2 className="font-bold text-lg">New message</h2>
                <button
                  onClick={() => setNewChatModalOpen(false)}
                  className="p-1.5 hover:bg-subtle rounded-full text-muted hover:text-foreground"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <div className="p-3 border-b border-border">
                <input
                  type="text"
                  placeholder="Search people"
                  value={searchQuery}
                  onChange={(e) => handleSearchUsers(e.target.value)}
                  className="w-full bg-card rounded-full py-2 px-4 text-sm outline-none border border-border focus:border-brand"
                  autoFocus
                />
              </div>

              <div className="flex-1 overflow-y-auto divide-y divide-border min-h-[200px]">
                {searching ? (
                  <div className="flex justify-center p-6 text-brand">
                    <Loader2 className="w-6 h-6 animate-spin" />
                  </div>
                ) : searchResults.length === 0 && searchQuery ? (
                  <div className="text-center p-8 text-muted text-sm">
                    No users found matching &ldquo;{searchQuery}&rdquo;
                  </div>
                ) : (
                  searchResults.map((u) => (
                    <div
                      key={u.id}
                      onClick={() => startConversationWith(u)}
                      className="flex items-center gap-3 p-3 hover:bg-subtle cursor-pointer transition-colors"
                    >
                      <img
                        src={
                          u.avatar_url ||
                          "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                        }
                        alt={u.display_name}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                      <div>
                        <div className="font-bold text-sm">{u.display_name}</div>
                        <div className="text-xs text-muted">@{u.username}</div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        )}

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
      </div>
    </div>
  );
}
