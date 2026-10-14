'use client';

import React, { useState } from 'react';

interface VerifiedBadgeProps {
  type?: 'blue' | 'business' | 'government' | 'gold' | 'grey' | 'staff' | string;
  size?: 'sm' | 'md' | 'lg';
  showPopover?: boolean;
}

export function VerifiedBadge({ type = 'blue', size = 'sm', showPopover = true }: VerifiedBadgeProps) {
  const [open, setOpen] = useState(false);

  // Normalize type
  const normalizedType = type === 'business' ? 'gold' : type === 'government' ? 'grey' : type;

  const sizeClass = size === 'lg' ? 'w-5 h-5' : size === 'md' ? 'w-4 h-4' : 'w-[18px] h-[18px]';

  const badgeConfig = {
    blue: {
      color: '#1d9bf0',
      title: 'Verified account',
      desc: "This account is verified because it's subscribed to Premium and its phone/identity was verified.",
      badgeColor: 'text-[#1d9bf0]',
    },
    gold: {
      color: '#e7a414',
      title: 'Official business or organization',
      desc: "This account is a verified business or organization on X.",
      badgeColor: 'text-[#e7a414]',
    },
    grey: {
      color: '#829aab',
      title: 'Government or multilateral organization',
      desc: "This account is a verified government agency, official, or multilateral organization.",
      badgeColor: 'text-[#829aab]',
    },
    staff: {
      color: '#00ba7c',
      title: 'Platform Staff',
      desc: "This account belongs to an official employee of the platform.",
      badgeColor: 'text-[#00ba7c]',
    },
  };

  const config = badgeConfig[normalizedType as keyof typeof badgeConfig] || badgeConfig.blue;

  return (
    <span className="relative inline-flex items-center ml-1 align-middle">
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          if (showPopover) setOpen(!open);
        }}
        className="focus:outline-none hover:opacity-85 transition-opacity"
        aria-label="Verified account"
      >
        {normalizedType === 'gold' ? (
          // Gold square icon
          <svg viewBox="0 0 24 24" aria-hidden="true" className={`${sizeClass} ${config.badgeColor} fill-current`}>
            <path d="M4 3h16a1 1 0 0 1 1 1v16a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1zm6.2 12.8 8-8-1.4-1.4-6.6 6.6-2.6-2.6-1.4 1.4 4 4z" />
          </svg>
        ) : (
          // Signature Blue & Grey checkmark badge
          <svg viewBox="0 0 22 22" aria-hidden="true" className={`${sizeClass} ${config.badgeColor} fill-current`}>
            <g>
              <path d="M20.396 11c-.018-.646-.215-1.275-.57-1.816-.354-.54-.852-.972-1.438-1.246.223-.607.27-1.264.14-1.897-.131-.634-.437-1.218-.882-1.687-.47-.445-1.053-.75-1.687-.882-.633-.13-1.29-.083-1.897.14-.273-.587-.704-1.086-1.245-1.44S11.647 1.62 11 1.604c-.646.017-1.273.213-1.813.568s-.969.854-1.24 1.44c-.608-.223-1.267-.272-1.902-.14-.635.13-1.22.436-1.69.882-.445.47-.749 1.055-.878 1.688-.13.633-.08 1.29.144 1.896-.587.274-1.087.705-1.443 1.245-.356.54-.555 1.17-.574 1.817.02.647.218 1.276.574 1.817.356.54.856.972 1.443 1.245-.224.606-.274 1.263-.144 1.896.13.634.433 1.218.877 1.688.47.443 1.054.747 1.687.878.633.132 1.29.084 1.897-.136.274.586.705 1.084 1.246 1.439.54.354 1.17.551 1.816.569.647-.016 1.276-.213 1.817-.567s.972-.854 1.245-1.44c.604.239 1.266.296 1.903.164.636-.132 1.22-.447 1.68-.907.46-.46.776-1.044.908-1.681s.075-1.299-.165-1.903c.586-.274 1.084-.705 1.439-1.246.354-.54.551-1.17.569-1.816zM9.662 14.85l-3.429-3.428 1.293-1.302 2.136 2.136 5.25-5.25 1.302 1.293-6.552 6.551z" />
            </g>
          </svg>
        )}
      </button>

      {/* Verification details popover dialog (X style) */}
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute top-full left-0 mt-2 z-50 w-72 p-4 bg-[#000000] border border-[#2f3336] rounded-2xl shadow-[0_0_15px_rgba(255,255,255,0.15)] text-white text-left animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-start justify-between mb-2">
              <h4 className="font-bold text-[15px]">{config.title}</h4>
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="text-[#71767b] hover:text-white p-1 rounded-full -mr-1 -mt-1"
              >
                ✕
              </button>
            </div>
            <p className="text-[13px] text-[#71767b] leading-relaxed mb-3">
              {config.desc}
            </p>
            <a
              href="/premium"
              className="text-[13px] text-[#1d9bf0] font-semibold hover:underline block"
            >
              Learn more about verified subscriptions
            </a>
          </div>
        </>
      )}
    </span>
  );
}
