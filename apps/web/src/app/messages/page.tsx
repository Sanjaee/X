"use client";

import React, { useState } from "react";
import { Mail, Settings, Send, Image as ImageIcon } from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";

export default function MessagesPage() {
  const [selectedConversation, setSelectedConversation] = useState<string | null>(null);
  const [messageInput, setMessageInput] = useState("");

  const conversations = [
    {
      id: "1",
      user: {
        name: "NASA",
        username: "NASA",
        avatar: "https://images.unsplash.com/photo-1614728894747-a83421e2b9c9?w=150&auto=format&fit=crop&q=80",
      },
      lastMessage: "Welcome to Artemis mission updates!",
      time: "Oct 12",
    },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        <LeftNavRail currentUser={null} />

        {/* Messages 2-Column Container */}
        <div className="flex-1 flex border-r border-border min-h-screen">
          {/* Conversation List Column (380px) */}
          <div className="w-full sm:w-[380px] border-r border-border flex flex-col">
            <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center justify-between px-4 h-[53px]">
              <h1 className="font-extrabold text-[20px]">Messages</h1>
              <div className="flex items-center gap-1">
                <button className="p-2 hover:bg-subtle rounded-full transition-colors">
                  <Settings className="w-5 h-5" />
                </button>
                <button className="p-2 hover:bg-subtle rounded-full transition-colors">
                  <Mail className="w-5 h-5" />
                </button>
              </div>
            </header>

            <div className="flex-1 flex flex-col divide-y divide-border">
              {conversations.map((c) => (
                <div
                  key={c.id}
                  onClick={() => setSelectedConversation(c.id)}
                  className="flex items-center gap-3 p-4 hover:bg-subtle/50 cursor-pointer transition-colors"
                >
                  <img
                    src={c.user.avatar}
                    alt={c.user.name}
                    className="w-12 h-12 rounded-full object-cover border border-border"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-[15px] truncate">{c.user.name}</span>
                      <span className="text-[13px] text-muted">{c.time}</span>
                    </div>
                    <span className="text-[14px] text-muted truncate block">@{c.user.username}</span>
                    <p className="text-[14px] text-muted truncate mt-0.5">{c.lastMessage}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Active Chat Column */}
          <div className="hidden sm:flex flex-1 flex-col justify-between h-screen">
            {selectedConversation ? (
              <>
                {/* Chat Header */}
                <div className="h-[53px] border-b border-border px-4 flex items-center justify-between">
                  <span className="font-bold text-[16px]">NASA</span>
                </div>

                {/* Chat Body */}
                <div className="flex-1 p-4 overflow-y-auto flex flex-col gap-3">
                  <div className="self-start max-w-[70%] bg-card p-3 rounded-2xl rounded-tl-sm text-[15px]">
                    Welcome to Artemis mission updates! Let us know if you have questions.
                  </div>
                </div>

                {/* Input Area */}
                <div className="p-3 border-t border-border flex items-center gap-2">
                  <button className="p-2 hover:bg-subtle text-brand rounded-full">
                    <ImageIcon className="w-5 h-5" />
                  </button>
                  <input
                    type="text"
                    placeholder="Start a new message"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    className="flex-1 bg-card rounded-full py-2 px-4 outline-none text-[15px] border border-transparent focus:border-brand"
                  />
                  <button className="p-2 hover:bg-subtle text-brand rounded-full">
                    <Send className="w-5 h-5" />
                  </button>
                </div>
              </>
            ) : (
              <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
                <h2 className="text-[28px] font-extrabold mb-2">Select a message</h2>
                <p className="text-muted text-[15px] max-w-sm">
                  Choose from your existing conversations, start a new one, or just keep swimming.
                </p>
                <button className="mt-6 bg-brand hover:bg-brand-hover text-white font-bold py-3 px-6 rounded-full text-[15px]">
                  New message
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
