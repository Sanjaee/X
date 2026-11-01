'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import {
  Check,
  ArrowLeft,
  Clock,
  Loader2,
  Lock,
  Wallet,
  Copy,
  CheckCircle2,
  RefreshCw,
  QrCode,
} from 'lucide-react';
import { RightSidebar } from '@/components/layout/RightSidebar';
import { UserNameWithRole } from '@/components/ui/UserNameWithRole';
import { User } from '@/types';
import { apiRequest } from '@/lib/api';
import { cn } from '@/lib/utils';

interface WhiteLabelInvoice {
  txn_id: string;
  invoice_url: string;
  amount: string;
  pending_amount: string;
  wallet_hash: string;
  psys_cid: string;
  currency: string;
  status: string;
  source_currency: string;
  source_rate: string;
  expire_utc: number | string;
  expected_confirmations: string;
  qr_code: string;
}

function InvoiceContent() {
  const searchParams = useSearchParams();
  const orderId = searchParams.get('order_id');

  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [orderDetails, setOrderDetails] = useState<any>(null);
  const [planNameFromApi, setPlanNameFromApi] = useState<string>('');
  const [whiteLabel, setWhiteLabel] = useState<WhiteLabelInvoice | null>(null);
  const [isCompleted, setIsCompleted] = useState(false);
  const [copiedAddress, setCopiedAddress] = useState(false);
  const [copiedAmount, setCopiedAmount] = useState(false);
  const [timeLeft, setTimeLeft] = useState<number>(24 * 3600);
  const [isManualRefreshing, setIsManualRefreshing] = useState(false);

  // Load user details
  useEffect(() => {
    apiRequest<User>('/me')
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));
  }, []);

  // Fetch Order and White-Label data
  const fetchOrderStatus = async (isSilent = false) => {
    if (!orderId) return;
    if (!isSilent) setIsManualRefreshing(true);

    try {
      const res = await apiRequest<{
        order: any;
        plan_name: string;
        white_label?: WhiteLabelInvoice;
      }>(`/subscriptions/orders/${orderId}`);

      if (res?.order) {
        setOrderDetails(res.order);
        if (res.plan_name) {
          setPlanNameFromApi(res.plan_name);
        }

        const wl = res.white_label || res.order?.metadata?.white_label;
        if (wl) {
          setWhiteLabel(wl);
          if (wl.expire_utc) {
            const expSec =
              typeof wl.expire_utc === 'number'
                ? wl.expire_utc - Math.floor(Date.now() / 1000)
                : parseInt(String(wl.expire_utc), 10) - Math.floor(Date.now() / 1000);
            if (!isNaN(expSec) && expSec > 0) {
              setTimeLeft(expSec);
            }
          }
        }

        if (res.order.status === 'completed') {
          setIsCompleted(true);
        }
      }
    } catch (e: any) {
      if (!isSilent) {
        console.error('Failed to load order status:', e);
      }
    } finally {
      if (!isSilent) setIsManualRefreshing(false);
    }
  };

  useEffect(() => {
    fetchOrderStatus(false);
  }, [orderId]);

  // Polling loop every 3.5 seconds to detect live blockchain settlement
  useEffect(() => {
    if (!orderId || isCompleted) return;

    const interval = setInterval(() => {
      fetchOrderStatus(true);
    }, 3500);

    return () => clearInterval(interval);
  }, [orderId, isCompleted]);

  // Expiration Countdown Timer
  useEffect(() => {
    if (isCompleted || timeLeft <= 0) return;

    const timer = setInterval(() => {
      setTimeLeft((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);

    return () => clearInterval(timer);
  }, [isCompleted, timeLeft]);

  const copyToClipboard = (text: string, type: 'address' | 'amount') => {
    navigator.clipboard.writeText(text);
    if (type === 'address') {
      setCopiedAddress(true);
      setTimeout(() => setCopiedAddress(false), 2000);
    } else {
      setCopiedAmount(true);
      setTimeout(() => setCopiedAmount(false), 2000);
    }
  };

  const formatTimer = (sec: number) => {
    const h = Math.floor(sec / 3600);
    const m = Math.floor((sec % 3600) / 60);
    const s = sec % 60;
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const roleKey = orderDetails?.plan_id || 'vip';
  const planName = planNameFromApi || (orderDetails?.plan_id ? orderDetails.plan_id.toUpperCase() : 'VIP Member');
  const priceDisplay = orderDetails?.amount_cents
    ? `$${(orderDetails.amount_cents / 100).toFixed(2)}`
    : '$1.00';

  const cryptoCurrency = whiteLabel?.currency || whiteLabel?.psys_cid || 'USDT_TRX';
  const cryptoAmount =
    whiteLabel?.amount ||
    (orderDetails?.amount_cents ? (orderDetails.amount_cents / 100).toFixed(2) : '1.00');
  const depositAddress = whiteLabel?.wallet_hash || 'Awaiting wallet address...';
  const qrCodeImage =
    whiteLabel?.qr_code ||
    `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(depositAddress)}`;

  const previewUserName = currentUser?.display_name || currentUser?.username || 'User';

  return (
    <main className="w-full max-w-[620px] border-r border-border min-h-screen pb-16">
      {/* Top Header */}
      <div className="sticky top-0 z-20 backdrop-blur-md bg-background/80 border-b border-border px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-5">
          <Link
            href="/premium"
            className="p-2 hover:bg-subtle rounded-full transition-colors text-foreground"
            aria-label="Back to Premium"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-[19px] font-extrabold leading-6">Crypto Checkout</h1>
            <span className="text-[12px] text-muted flex items-center gap-1.5">
              <span>Plisio White-Label Gateway</span>
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              <span>Live Direct Settlement</span>
            </span>
          </div>
        </div>

        <button
          onClick={() => fetchOrderStatus(false)}
          disabled={isManualRefreshing}
          className="p-2 hover:bg-subtle rounded-full transition-colors text-muted hover:text-foreground"
          title="Refresh status"
        >
          <RefreshCw className={cn('w-4 h-4', isManualRefreshing && 'animate-spin text-[#1d9bf0]')} />
        </button>
      </div>

      <div className="p-4 sm:p-6 space-y-6">
        {/* State: Completed / Verified */}
        {isCompleted ? (
          <div className="border border-green-500/40 bg-gradient-to-b from-green-500/15 via-background to-background rounded-3xl p-6 sm:p-8 text-center space-y-4 shadow-2xl animate-in zoom-in-95 duration-200">
            <div className="w-20 h-20 bg-emerald-500/20 rounded-full flex items-center justify-center mx-auto border border-emerald-500/50 shadow-[0_0_25px_rgba(16,185,129,0.4)]">
              <CheckCircle2 className="w-10 h-10 text-emerald-400" strokeWidth={2.5} />
            </div>

            <div className="space-y-1.5">
              <h2 className="text-[26px] font-black tracking-tight text-foreground">Payment Received!</h2>
              <p className="text-muted text-[14px] max-w-sm mx-auto leading-relaxed">
                Your cryptocurrency payment has confirmed. Your <strong className="text-foreground">{planName}</strong> role and quotas are now active.
              </p>
            </div>

            <div className="bg-subtle/60 rounded-2xl p-4 text-left text-[14px] space-y-2.5 border border-border">
              <div className="flex justify-between items-center">
                <span className="text-muted">Activated Role</span>
                <UserNameWithRole displayName={previewUserName} role={roleKey} className="text-sm" />
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted">Paid Amount</span>
                <span className="font-mono text-[13px] font-bold text-foreground">
                  {cryptoAmount} {cryptoCurrency} ({priceDisplay} USD)
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted">Order ID</span>
                <span className="font-mono text-[11px] text-muted line-clamp-1">{orderId}</span>
              </div>
            </div>

            <div className="pt-2 flex flex-col sm:flex-row gap-3">
              <Link
                href={`/${currentUser?.username || ''}`}
                className="flex-1 bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white font-extrabold py-3 px-6 rounded-full transition-all text-center shadow-md flex items-center justify-center gap-2"
              >
                <span>View Profile</span>
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
            {/* Plan Overview & Price Banner */}
            <div className="border border-border rounded-2xl p-4 sm:p-5 bg-subtle/30 flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-[18px] font-extrabold">{planName}</h2>
                </div>
                <div className="mt-1 flex items-center gap-2 text-xs text-muted">
                  <span>Role Preview:</span>
                  <UserNameWithRole displayName={previewUserName} role={roleKey} className="text-sm" />
                </div>
              </div>
              <div className="text-right">
                <span className="text-[24px] font-black">{priceDisplay}</span>
                <span className="text-muted text-[12px] block uppercase tracking-wider font-semibold">
                  USD / mo
                </span>
              </div>
            </div>

            {/* WHITE LABEL PAYMENT CARD */}
            <div className="border border-border/80 rounded-3xl p-5 sm:p-6 bg-[#16181c] text-white shadow-2xl space-y-6 relative overflow-hidden">
              {/* Top Accent Gradient Bar */}
              <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-[#1d9bf0] via-cyan-400 to-[#1d9bf0]" />

              {/* Header with Expiry Countdown */}
              <div className="flex items-center justify-between pb-3 border-b border-white/10">
                <div className="flex items-center gap-2">
                  <div className="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center overflow-hidden">
                    <img
                      src={`/crypto-icons/${cryptoCurrency}.svg`}
                      alt={cryptoCurrency}
                      className="w-5 h-5 object-contain"
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).style.display = 'none';
                      }}
                    />
                    <Wallet className="w-4 h-4 text-[#1d9bf0] hidden" />
                  </div>
                  <div>
                    <h3 className="text-[14px] font-bold text-white uppercase tracking-wider">
                      {cryptoCurrency.replace('_', ' ')}
                    </h3>
                    <span className="text-[11px] text-neutral-400">Direct Merchant Deposit</span>
                  </div>
                </div>

                <div className="flex items-center gap-2 bg-white/5 border border-white/10 px-3 py-1.5 rounded-xl">
                  <Clock className="w-3.5 h-3.5 text-neutral-400 animate-pulse" />
                  <span className="font-mono text-[13px] font-bold text-neutral-200">
                    {formatTimer(timeLeft)}
                  </span>
                </div>
              </div>

              {/* Amount to Send (with 1-Click Copy) */}
              <div className="bg-black/50 border border-white/10 rounded-2xl p-4 text-center space-y-1 relative group">
                <span className="text-[11px] font-bold text-neutral-400 uppercase tracking-widest block">
                  Amount to Send
                </span>
                <div
                  onClick={() => copyToClipboard(cryptoAmount, 'amount')}
                  className="inline-flex items-center justify-center gap-2.5 cursor-pointer hover:opacity-85 transition-opacity px-2 py-1"
                  title="Click to copy amount"
                >
                  <span className="text-[26px] sm:text-[30px] font-black tracking-tight text-white font-mono">
                    {cryptoAmount}
                  </span>
                  <span className="text-[15px] font-extrabold text-[#1d9bf0] uppercase">
                    {cryptoCurrency}
                  </span>
                  <div className="ml-1 p-1.5 rounded-lg bg-white/10 hover:bg-white/20 transition-colors">
                    {copiedAmount ? (
                      <Check className="w-4 h-4 text-emerald-400" />
                    ) : (
                      <Copy className="w-4 h-4 text-neutral-300" />
                    )}
                  </div>
                </div>
                {copiedAmount && (
                  <p className="text-[11px] text-emerald-400 font-semibold animate-in fade-in duration-200">
                    Amount copied to clipboard!
                  </p>
                )}
              </div>

              {/* QR Code Section */}
              <div className="flex flex-col items-center justify-center gap-3 py-2">
                <div className="bg-white p-3.5 rounded-2xl shadow-xl inline-block border-4 border-white/20 relative group">
                  <img
                    src={qrCodeImage}
                    alt="Cryptocurrency Deposit QR Code"
                    className="w-48 h-48 sm:w-52 sm:h-52 object-contain rounded-lg"
                  />
                  <div className="absolute -bottom-2.5 left-1/2 -translate-x-1/2 bg-black text-white border border-white/20 text-[10px] font-bold px-2.5 py-0.5 rounded-full flex items-center gap-1 shadow-md whitespace-nowrap">
                    <QrCode className="w-3 h-3 text-[#1d9bf0]" />
                    <span>Scan in wallet app</span>
                  </div>
                </div>
                <p className="text-[12px] text-neutral-400 text-center max-w-xs mt-1">
                  Scan the QR code with your crypto wallet (e.g. Trust Wallet, Binance, MetaMask, Phantom).
                </p>
              </div>

              {/* Destination Wallet Address (with 1-Click Copy) */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between text-[12px]">
                  <span className="font-bold text-neutral-300 uppercase tracking-wider">
                    Deposit Wallet Address
                  </span>
                  <span className="text-[11px] text-neutral-400">Network: {cryptoCurrency}</span>
                </div>

                <div
                  onClick={() => copyToClipboard(depositAddress, 'address')}
                  className="bg-black/60 hover:bg-black/80 border border-white/15 hover:border-[#1d9bf0]/50 rounded-2xl p-3.5 flex items-center justify-between gap-3 cursor-pointer transition-all group shadow-inner"
                >
                  <p className="font-mono text-[13px] sm:text-[14px] text-neutral-100 font-semibold break-all select-all">
                    {depositAddress}
                  </p>
                  <div className="p-2 rounded-xl bg-white/10 group-hover:bg-[#1d9bf0] transition-colors flex-shrink-0 text-white">
                    {copiedAddress ? (
                      <Check className="w-4 h-4 text-emerald-400" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </div>
                </div>

                <div className="flex items-center justify-between text-[11px] text-neutral-400 pt-1">
                  <span>Confirmations required: {whiteLabel?.expected_confirmations || '1'} block</span>
                  {copiedAddress && (
                    <span className="text-emerald-400 font-semibold">Address copied!</span>
                  )}
                </div>
              </div>

              {/* Live Awaiting Confirmation Indicator */}
              <div className="bg-white/5 border border-white/10 rounded-2xl p-4 flex items-center gap-3">
                <div className="w-5 h-5 flex-shrink-0 flex items-center justify-center">
                  <div className="w-3 h-3 rounded-full bg-[#1d9bf0] animate-ping" />
                </div>
                <div className="text-[13px] leading-snug">
                  <p className="font-bold text-neutral-200">Listening for transaction on blockchain...</p>
                  <p className="text-neutral-400 text-[12px]">
                    Once you send funds, this page will automatically confirm and activate your {planName} role.
                  </p>
                </div>
              </div>
            </div>

            {/* Security Note */}
            <div className="flex items-center justify-center gap-2 text-muted text-[12px] text-center pt-2">
              <Lock className="w-3.5 h-3.5" />
              <span>Direct cryptographic settlement via Plisio API. No intermediary redirects.</span>
            </div>
          </>
        )}
      </div>
    </main>
  );
}

export default function InvoicePage() {
  return (
    <>
      <Suspense
        fallback={
          <main className="w-full max-w-[620px] border-r border-border min-h-screen flex items-center justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-[#1d9bf0]" />
          </main>
        }
      >
        <InvoiceContent />
      </Suspense>
      <RightSidebar />
    </>
  );
}
