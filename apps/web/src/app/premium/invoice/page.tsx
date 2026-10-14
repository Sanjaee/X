'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ShieldCheck,
  Check,
  ArrowLeft,
  Sparkles,
  ExternalLink,
  Loader2,
  Lock,
  Wallet,
  Zap,
} from 'lucide-react';
import { LeftNavRail } from '@/components/layout/LeftNavRail';
import { RightSidebar } from '@/components/layout/RightSidebar';
import { VerifiedBadge } from '@/components/common/VerifiedBadge';
import { User } from '@/types';
import { apiRequest } from '@/lib/api';
import { cn } from '@/lib/utils';

function InvoiceContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const orderId = searchParams.get('order_id');
  const isSimulated = searchParams.get('simulated') === 'true';

  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [orderDetails, setOrderDetails] = useState<any>(null);
  const [isCompleted, setIsCompleted] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    apiRequest<User>('/me')
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));

    if (orderId) {
      // Fetch billing history to find this specific order
      apiRequest<{ orders: any[] }>('/billing/history')
        .then((res) => {
          const found = res.orders?.find((o) => o.id === orderId);
          if (found) {
            setOrderDetails(found);
            if (found.status === 'completed') {
              setIsCompleted(true);
            }
          }
        })
        .catch((e) => console.error(e));
    }
  }, [orderId]);

  const handleConfirmPayment = async () => {
    if (!orderId) return;
    setLoading(true);
    setErrorMessage(null);

    try {
      const res = await apiRequest<{ status: string; message: string }>('/subscriptions/simulate-payment', {
        method: 'POST',
        body: JSON.stringify({ order_id: orderId }),
      });

      if (res.status === 'completed') {
        setIsCompleted(true);
      }
    } catch (err: any) {
      setErrorMessage(err.message || 'Payment simulation failed');
    } finally {
      setLoading(false);
    }
  };

  const planName = orderDetails?.plan_id ? orderDetails.plan_id.replace('_', ' ') : 'Premium';
  const priceDisplay = orderDetails?.amount_cents
    ? `$${(orderDetails.amount_cents / 100).toFixed(2)}`
    : '$8.00';

  return (
    <main className="w-full max-w-[600px] border-r border-border min-h-screen pb-16">
      {/* Header */}
      <div className="sticky top-0 z-20 backdrop-blur-md bg-background/80 border-b border-border px-4 py-3 flex items-center gap-6">
        <Link
          href="/premium"
          className="p-2 hover:bg-subtle rounded-full transition-colors text-foreground"
          aria-label="Back"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-[20px] font-bold leading-6">Checkout</h1>
          <span className="text-[13px] text-muted">X Premium Subscription</span>
        </div>
      </div>

      <div className="p-4 sm:p-6 space-y-6">
        {/* Success Modal / State */}
        {isCompleted ? (
          <div className="border border-green-500/40 bg-gradient-to-b from-green-500/10 via-background to-background rounded-3xl p-6 sm:p-8 text-center space-y-4 shadow-2xl animate-in zoom-in-95 duration-200">
            <div className="w-16 h-16 bg-[#1d9bf0]/20 rounded-full flex items-center justify-center mx-auto border border-[#1d9bf0]/50 shadow-[0_0_20px_rgba(29,155,240,0.4)]">
              <VerifiedBadge type="blue" size="lg" showPopover={false} />
            </div>

            <div className="space-y-1">
              <h2 className="text-[26px] font-extrabold tracking-tight">You are now Verified!</h2>
              <p className="text-muted text-[14px]">
                Your subscription to X {planName} is active. Your verified blue badge will now display beside your username everywhere.
              </p>
            </div>

            <div className="bg-subtle/50 rounded-2xl p-4 text-left text-[14px] space-y-2 border border-border">
              <div className="flex justify-between">
                <span className="text-muted">Plan</span>
                <span className="font-bold capitalize">{planName}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted">Verification Status</span>
                <span className="font-bold text-[#1d9bf0] flex items-center gap-1">
                  Active <VerifiedBadge type="blue" size="sm" showPopover={false} />
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted">Order ID</span>
                <span className="font-mono text-[12px] text-muted line-clamp-1">{orderId}</span>
              </div>
            </div>

            <div className="pt-2 flex flex-col sm:flex-row gap-3">
              <Link
                href={`/${currentUser?.username || ''}`}
                className="flex-1 bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white font-bold py-3 px-6 rounded-full transition-colors text-center shadow-md"
              >
                View Your Verified Profile
              </Link>
              <Link
                href="/"
                className="flex-1 bg-subtle hover:bg-neutral-800 text-foreground font-semibold py-3 px-6 rounded-full transition-colors text-center border border-border"
              >
                Go to Timeline
              </Link>
            </div>
          </div>
        ) : (
          <>
            {/* Order Summary Card */}
            <div className="border border-border rounded-2xl p-5 bg-subtle/30 space-y-4">
              <div className="flex justify-between items-start border-b border-border pb-4">
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-[20px] font-extrabold capitalize">X {planName}</h2>
                    <VerifiedBadge type="blue" size="md" showPopover={false} />
                  </div>
                  <p className="text-[13px] text-muted">
                    {orderDetails?.billing_cycle === 'annual' ? 'Billed Annually' : 'Billed Monthly'}
                  </p>
                </div>
                <div className="text-right">
                  <span className="text-[26px] font-black">{priceDisplay}</span>
                  <span className="text-muted text-[13px] block">USD</span>
                </div>
              </div>

              {/* Benefits list */}
              <ul className="space-y-2 text-[14px] text-neutral-300">
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-[#1d9bf0]" />
                  <span>Blue checkmark badge on your profile and posts</span>
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-[#1d9bf0]" />
                  <span>50% fewer ads in For You and Following</span>
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-[#1d9bf0]" />
                  <span>Post up to 4,000 characters per tweet</span>
                </li>
                <li className="flex items-center gap-2">
                  <Check className="w-4 h-4 text-[#1d9bf0]" />
                  <span>60-minute 1080p video uploads stored on Cloudflare R2</span>
                </li>
              </ul>
            </div>

            {/* Payment Method / Crypto Gateway */}
            <div className="border border-border rounded-2xl p-5 bg-subtle/20 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 font-bold text-[16px]">
                  <Wallet className="w-5 h-5 text-[#1d9bf0]" />
                  <span>Cryptocurrency Payment</span>
                </div>
                <span className="text-[12px] bg-[#1d9bf0]/10 text-[#1d9bf0] font-bold px-2 py-0.5 rounded-full">
                  Plisio Gateway
                </span>
              </div>

              <p className="text-[13px] text-muted leading-relaxed">
                Pay securely using BTC, USDT (TRC-20 / ERC-20), ETH, LTC, or SOL. Once the blockchain transaction confirms, your subscription and blue checkmark will activate authoritatively.
              </p>

              {/* Crypto Pills */}
              <div className="flex flex-wrap gap-2 text-[12px] font-semibold text-neutral-400">
                <span className="bg-subtle px-2.5 py-1 rounded-lg border border-border">Bitcoin (BTC)</span>
                <span className="bg-subtle px-2.5 py-1 rounded-lg border border-border">Tether (USDT)</span>
                <span className="bg-subtle px-2.5 py-1 rounded-lg border border-border">Ethereum (ETH)</span>
                <span className="bg-subtle px-2.5 py-1 rounded-lg border border-border">Solana (SOL)</span>
              </div>
            </div>

            {/* Dev / Simulation Section */}
            {(isSimulated || true) && (
              <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-2xl p-5 space-y-3">
                <div className="flex items-center gap-2 text-yellow-400 font-bold text-[15px]">
                  <Zap className="w-5 h-5" />
                  <span>Simulated Payment Test Mode</span>
                </div>
                <p className="text-[13px] text-neutral-300 leading-relaxed">
                  You are evaluating the payment and entitlement activation workflow. Click below to simulate authoritative Plisio webhook settlement without spending real currency.
                </p>

                {errorMessage && (
                  <p className="text-red-400 text-[13px] font-semibold">{errorMessage}</p>
                )}

                <button
                  type="button"
                  onClick={handleConfirmPayment}
                  disabled={loading}
                  className="w-full bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white font-extrabold text-[15px] py-3.5 px-6 rounded-full shadow-lg transition-all flex items-center justify-center gap-2 active:scale-98 disabled:opacity-50"
                >
                  {loading ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    <>
                      <span>Confirm & Activate Subscription</span>
                      <ShieldCheck className="w-5 h-5" />
                    </>
                  )}
                </button>
              </div>
            )}

            <div className="flex items-center justify-center gap-1.5 text-muted text-[12px] text-center">
              <Lock className="w-3.5 h-3.5" />
              <span>Encrypted payment processing. Non-refundable after activation.</span>
            </div>
          </>
        )}
      </div>
    </main>
  );
}

export default function InvoicePage() {
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  useEffect(() => {
    apiRequest<User>('/me')
      .then((u) => setCurrentUser(u))
      .catch(() => setCurrentUser(null));
  }, []);

  return (
    <div className="flex min-h-screen bg-background text-foreground justify-center">
      <LeftNavRail currentUser={currentUser} />
      <Suspense
        fallback={
          <main className="w-full max-w-[600px] border-r border-border min-h-screen flex items-center justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-[#1d9bf0]" />
          </main>
        }
      >
        <InvoiceContent />
      </Suspense>
      <RightSidebar />
    </div>
  );
}
