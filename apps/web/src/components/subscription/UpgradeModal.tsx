'use client';

import React, { useState, useEffect } from 'react';
import {
  X,
  Check,
  Loader2,
  ArrowLeft,
  Search,
  Copy,
  CheckCircle2,
  Clock,
  Wallet,
  XCircle,
  QrCode,
  RefreshCw,
} from 'lucide-react';
import { useUpgradeModal } from './UpgradeModalContext';
import { useCurrentUser } from '@/lib/useCurrentUser';
import { apiRequest } from '@/lib/api';
import { cn } from '@/lib/utils';
import { UserNameWithRole } from '@/components/ui/UserNameWithRole';
import { AuthModal } from '@/components/auth/AuthModal';
import { MembershipPlan } from '@/types';

interface CryptoCurrencyOption {
  cid: string;
  currency: string;
  name: string;
  icon: string;
  price_usd?: string;
  priceUsd?: string;
}

interface PaymentOrderStatus {
  order_id: string;
  status: string; // 'pending' | 'completed' | 'expired' | 'cancelled'
  plan_id: string;
  plan_name: string;
  billing_cycle: string;
  amount_cents: number;
  fiat_currency: string;
  coin?: string;
  crypto_amount?: string;
  pending_amount?: string;
  received_amount?: string;
  wallet_hash?: string;
  qr_code?: string;
  expire_at?: number;
  gateway_status?: string;
  is_simulated?: boolean;
}

interface ParsedRolePlan {
  id: string;
  key: string;
  name: string;
  badgeText: string;
  monthlyPriceCents: number;
  annualPriceCents: number;
  popular?: boolean;
  packageHeading: string;
  benefits: string[];
}

const ROLE_METADATA: Record<
  string,
  {
    badgeText: string;
    popular?: boolean;
    packageHeading: string;
    fallbackHighlights: string[];
  }
> = {
  vip: {
    badgeText: '25% off 6 months',
    packageHeading: 'Introductory package:',
    fallbackHighlights: [
      'VIP Role & Styled Username',
      'Basic analytics & dashboard',
      'Custom profile colors & theme',
      'Fewer ads in timeline',
    ],
  },
  mvp: {
    badgeText: 'Most popular',
    popular: true,
    packageHeading: 'Everything in VIP, and:',
    fallbackHighlights: [
      'Everything in VIP',
      'Animated MVP Role Badge',
      'Larger reply boost in timeline',
      'Increased upload limits',
    ],
  },
  mod: {
    badgeText: 'Staff & Mod',
    packageHeading: 'Everything in MVP, and:',
    fallbackHighlights: [
      'Everything in MVP',
      'Moderator Role Badge & Tools',
      'Report management & review tools',
      'Content moderation dashboard',
    ],
  },
  god: {
    badgeText: '50% off 1 month',
    packageHeading: 'Everything in Moderator, and:',
    fallbackHighlights: [
      'Everything in Mod',
      'God Mode Role & Privileges',
      'Bypass standard post & rate limits',
      'Highest daily quota allocation',
    ],
  },
};

const num = (s?: string) => parseFloat(s ?? '') || 0;

export function UpgradeModal() {
  const { isOpen, closeUpgradeModal, selectedPlanId } = useUpgradeModal();
  const { currentUser, updateUser } = useCurrentUser();

  // Wizard Steps: 'tiers' -> 'agreement' -> 'configure' -> 'payment'
  const [step, setStep] = useState<'tiers' | 'agreement' | 'configure' | 'payment'>('tiers');
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'annual'>('monthly');

  // Real Database Membership Plans
  const [plans, setPlans] = useState<ParsedRolePlan[]>([]);
  const [loadingPlans, setLoadingPlans] = useState(true);
  const [selectedPlan, setSelectedPlan] = useState<ParsedRolePlan | null>(null);

  // Auth Modal trigger if guest tries to subscribe
  const [authModalOpen, setAuthModalOpen] = useState(false);

  // Currencies state (fetched from real Go API)
  const [currencies, setCurrencies] = useState<CryptoCurrencyOption[]>([]);
  const [selectedCoin, setSelectedCoin] = useState<string>('');
  const [coinSearch, setCoinSearch] = useState('');
  const [loadingCoins, setLoadingCoins] = useState(false);

  // Checkout and Payment state
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [orderNotice, setOrderNotice] = useState<string | null>(null);
  const [activeOrderId, setActiveOrderId] = useState<string | null>(null);
  const [paymentData, setPaymentData] = useState<PaymentOrderStatus | null>(null);
  const [paymentError, setPaymentError] = useState<string | null>(null);
  const [copiedAddress, setCopiedAddress] = useState(false);
  const [copiedAmount, setCopiedAmount] = useState(false);

  // Load real dynamic plans from database
  useEffect(() => {
    if (isOpen) {
      setLoadingPlans(true);
      apiRequest<{ plans: MembershipPlan[] }>('/membership/plans')
        .then((res) => {
          if (res?.plans && res.plans.length > 0) {
            const purchasable = res.plans
              .filter((p) => p.is_active && p.key !== 'free' && p.price_amount_cents > 0)
              .map((p) => {
                const meta = ROLE_METADATA[p.key] || {
                  badgeText: 'Active',
                  packageHeading: 'Features included:',
                  fallbackHighlights: ['Verified Badge', 'Higher Limits'],
                };

                const quotaFeatures: string[] = [];
                if (p.entitlements && p.entitlements.length > 0) {
                  const entMap: Record<string, number> = {};
                  p.entitlements.forEach((e) => {
                    entMap[e.entitlement_key] = e.limit_value;
                  });
                  if (entMap['posts_per_day']) {
                    quotaFeatures.push(`${entMap['posts_per_day'].toLocaleString()} posts per day`);
                  }
                  if (entMap['max_post_length']) {
                    quotaFeatures.push(
                      `Up to ${entMap['max_post_length'].toLocaleString()} characters per post`
                    );
                  }
                  if (entMap['max_video_minutes'] && entMap['max_file_size_mb']) {
                    quotaFeatures.push(
                      `${entMap['max_video_minutes']}-min video uploads (up to ${entMap['max_file_size_mb']} MB)`
                    );
                  }
                  if (entMap['can_manage_colors']) {
                    quotaFeatures.push('Custom display name styling & colors');
                  }
                  if (entMap['can_view_analytics']) {
                    quotaFeatures.push('Advanced creator analytics');
                  }
                }

                const benefits =
                  quotaFeatures.length > 0
                    ? [meta.fallbackHighlights[0], ...quotaFeatures]
                    : meta.fallbackHighlights;

                const displayName = p.key === 'vip' ? 'VIP' : p.name;

                return {
                  id: p.id,
                  key: p.key,
                  name: displayName,
                  badgeText: meta.badgeText,
                  popular: meta.popular,
                  monthlyPriceCents: p.price_amount_cents,
                  annualPriceCents: Math.round(p.price_amount_cents * 0.8), // 20% discount on annual
                  packageHeading: meta.packageHeading,
                  benefits,
                };
              });

            setPlans(purchasable);
            const matched =
              purchasable.find((pl) => pl.id === selectedPlanId || pl.key === selectedPlanId) ||
              purchasable[0];
            setSelectedPlan(matched);
          }
        })
        .catch((err) => {
          console.error('Failed to load membership plans:', err);
        })
        .finally(() => {
          setLoadingPlans(false);
        });

      setStep('tiers');
      setPaymentData(null);
      setActiveOrderId(null);
      setPaymentError(null);
      setOrderNotice(null);

      // Check if user has an existing active transaction and redirect immediately
      if (currentUser) {
        apiRequest<{ order: PaymentOrderStatus | null }>('/subscriptions/orders/active')
          .then((res) => {
            if (res?.order && res.order.order_id) {
              setActiveOrderId(res.order.order_id);
              setPaymentData(res.order);
              setOrderNotice(
                'You have an active pending order. Please complete payment or cancel it to choose another role or coin.'
              );
              setStep('payment');
            }
          })
          .catch(() => {});
      }
    }
  }, [isOpen, selectedPlanId, currentUser]);

  // Fetch real supported cryptocurrencies from Go backend
  useEffect(() => {
    if (isOpen && currencies.length === 0) {
      setLoadingCoins(true);
      apiRequest<{ currencies: CryptoCurrencyOption[] }>('/payments/currencies')
        .then((res) => {
          if (res?.currencies && res.currencies.length > 0) {
            setCurrencies(res.currencies);
            if (!selectedCoin || !res.currencies.some((c) => c.currency === selectedCoin || c.cid === selectedCoin)) {
              setSelectedCoin(res.currencies[0].currency || res.currencies[0].cid);
            }
          }
        })
        .catch((err: any) => {
          console.warn('Could not load currencies:', err);
          setPaymentError(err?.message || 'Failed to load cryptocurrencies from payment provider');
        })
        .finally(() => {
          setLoadingCoins(false);
        });
    }
  }, [isOpen, currencies.length, selectedCoin]);

  // Real-time polling for payment status while in 'payment' step
  useEffect(() => {
    if (!activeOrderId || step !== 'payment') return;

    let stopped = false;
    let timer: ReturnType<typeof setInterval> | undefined;

    async function poll() {
      if (!activeOrderId || stopped) return;
      try {
        const res = await apiRequest<PaymentOrderStatus>(`/subscriptions/orders/${activeOrderId}`);
        if (stopped) return;
        setPaymentData(res);
        if (res.status === 'completed') {
          apiRequest('/me')
            .then((fresh) => {
              if (fresh) {
                updateUser(fresh);
                window.dispatchEvent(new Event('x_auth_state_change'));
              }
            })
            .catch(() => {});
        }
      } catch (err: any) {
        if (stopped) return;
        setPaymentError(err?.message || 'Transaction not found');
      }
    }

    poll();
    timer = setInterval(poll, 3000);

    return () => {
      stopped = true;
      if (timer) clearInterval(timer);
    };
  }, [activeOrderId, step, updateUser]);

  if (!isOpen) return null;

  // Handle proceed from Tier Card
  const handleSelectTierAndContinue = (tier: ParsedRolePlan) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    setSelectedPlan(tier);
    setStep('agreement');
  };

  // Handle cancel of an active pending order
  const handleCancelOrder = async () => {
    if (!activeOrderId) return;
    setIsCancelling(true);
    setPaymentError(null);
    try {
      await apiRequest(`/subscriptions/orders/${activeOrderId}/cancel`, {
        method: 'POST',
      });
      setActiveOrderId(null);
      setPaymentData(null);
      setOrderNotice(null);
      setStep('tiers');
    } catch (err: any) {
      setPaymentError(err?.message || 'Failed to cancel order. Please try again.');
    } finally {
      setIsCancelling(false);
    }
  };

  // Handle final checkout submit (hits real backend Plisio checkout)
  const handleInitiatePayment = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    if (!selectedPlan) return;

    setIsSubmitting(true);
    setPaymentError(null);
    setOrderNotice(null);

    try {
      const res = await apiRequest<PaymentOrderStatus & { is_existing?: boolean; message?: string }>(
        '/subscriptions/checkout',
        {
          method: 'POST',
          body: JSON.stringify({
            plan_id: selectedPlan.id,
            billing_cycle: billingCycle,
            currency: selectedCoin,
          }),
        }
      );

      if (res?.order_id) {
        setActiveOrderId(res.order_id);
        setPaymentData(res);
        if (res.is_existing) {
          setOrderNotice(
            res.message ||
              'An active pending order already exists. Please complete payment or cancel it to choose a different plan.'
          );
        }
        setStep('payment');
      } else {
        throw new Error('Failed to generate payment invoice');
      }
    } catch (err: any) {
      const msg = err?.message || 'Failed to initiate payment. Please try another coin.';
      if (msg.includes('unauthorized') || msg.includes('session') || msg.includes('401')) {
        setAuthModalOpen(true);
      }
      setPaymentError(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Copy helper
  const copyToClipboard = async (text: string, type: 'address' | 'amount') => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'address') {
        setCopiedAddress(true);
        setTimeout(() => setCopiedAddress(false), 2000);
      } else {
        setCopiedAmount(true);
        setTimeout(() => setCopiedAmount(false), 2000);
      }
    } catch {}
  };

  const previewUserName = currentUser?.display_name || currentUser?.username || 'Username';

  // Determine payment screen sub-state
  const required = num(paymentData?.crypto_amount);
  const received = num(paymentData?.received_amount);
  const pendingRemaining = num(paymentData?.pending_amount);
  const ps = paymentData?.gateway_status;
  const isPaid = paymentData?.status === 'completed';
  const isExpired = paymentData?.status === 'expired';
  const isCancelled = paymentData?.status === 'cancelled';

  let paymentScreen: 'loading' | 'error' | 'success' | 'expired' | 'qr' | 'processing' | 'partial' =
    'qr';
  if (paymentError) paymentScreen = 'error';
  else if (!paymentData) paymentScreen = 'loading';
  else if (isPaid) paymentScreen = 'success';
  else if (isExpired || isCancelled) paymentScreen = 'expired';
  else if (received > 0 && pendingRemaining > 0 && pendingRemaining < required)
    paymentScreen = 'partial';
  else if (received > 0 || ps === 'pending' || ps === 'pending internal')
    paymentScreen = 'processing';
  else paymentScreen = 'qr';

  return (
    <>
      <div className="fixed inset-0 z-[120] flex items-center justify-center bg-black/80 backdrop-blur-sm p-2 sm:p-4 animate-in fade-in overflow-y-auto">
        <div
          className={cn(
            'bg-[#000000] border border-[#2f3336] rounded-3xl w-full text-white shadow-2xl relative my-auto transition-all',
            step === 'tiers' ? 'max-w-6xl' : step === 'configure' ? 'max-w-4xl' : 'max-w-xl'
          )}
        >
          {/* ================================================================ */}
          {/* STEP 1: TIERS SELECTION MODAL — UPGRADE YOUR ROLE (IMAGE 1)       */}
          {/* ================================================================ */}
          {step === 'tiers' && (
            <div className="p-5 sm:p-7 space-y-6">
              {/* Header Bar */}
              <div className="flex items-center justify-between border-b border-[#2f3336] pb-4">
                <div className="flex items-center gap-3">
                  <img src="/LM.png" alt="LM Logo" className="w-7 h-7 object-contain" />
                </div>

                <div className="text-center font-extrabold text-base sm:text-xl hidden md:block">
                  One subscription. Access to everything.
                </div>

                <div className="flex items-center gap-3">
                  {/* Monthly / Annual Pill Toggle */}
                  <div className="bg-[#16181c] border border-[#2f3336] p-0.5 rounded-full flex items-center text-xs font-semibold">
                    <button
                      type="button"
                      onClick={() => setBillingCycle('monthly')}
                      className={cn(
                        'px-3 py-1 rounded-full transition-colors',
                        billingCycle === 'monthly'
                          ? 'bg-white text-black font-bold shadow'
                          : 'text-[#71767b] hover:text-white'
                      )}
                    >
                      Monthly
                    </button>
                    <button
                      type="button"
                      onClick={() => setBillingCycle('annual')}
                      className={cn(
                        'px-3 py-1 rounded-full transition-colors',
                        billingCycle === 'annual'
                          ? 'bg-white text-black font-bold shadow'
                          : 'text-[#71767b] hover:text-white'
                      )}
                    >
                      Annual
                    </button>
                  </div>

                  {/* Close X Button */}
                  <button
                    onClick={closeUpgradeModal}
                    className="p-1.5 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                  >
                    <X className="w-5 h-5" />
                  </button>
                </div>
              </div>

              {/* Subheading */}
              <div className="text-center space-y-1 py-1">
                <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
                  Upgrade Your Role
                </h2>
                <p className="text-gray-400 text-xs sm:text-sm max-w-lg mx-auto">
                  Unlock exclusive role badges, styled display names, higher daily post quotas, and
                  enhanced media uploads.
                </p>
              </div>

              {/* Dynamic Role Plans Grid (Real Database Plans) */}
              {loadingPlans ? (
                <div className="flex justify-center py-16">
                  <Loader2 className="w-8 h-8 animate-spin text-[#1d9bf0]" />
                </div>
              ) : (
                <div
                  className={cn(
                    'grid gap-4',
                    plans.length === 4
                      ? 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4'
                      : 'grid-cols-1 sm:grid-cols-3'
                  )}
                >
                  {plans.map((tier) => {
                    const priceCents =
                      billingCycle === 'annual' ? tier.annualPriceCents : tier.monthlyPriceCents;
                    const formattedPrice = `$${(priceCents / 100).toFixed(0)}`;
                    const isSelected = selectedPlan?.id === tier.id;

                    return (
                      <div
                        key={tier.id}
                        className={cn(
                          'rounded-2xl p-5 flex flex-col justify-between border transition-all relative',
                          isSelected || tier.popular
                            ? 'bg-[#121519] border-[#1d9bf0] shadow-[0_0_20px_rgba(29,155,240,0.2)]'
                            : 'bg-[#16181c] border-[#2f3336] hover:border-[#444a50]'
                        )}
                      >
                        <div>
                          {/* Role Title */}
                          <div className="flex items-center justify-between mb-2">
                            <span className="font-bold text-base text-white">{tier.name}</span>
                            <span
                              className={cn(
                                'text-[10px] font-bold px-2 py-0.5 rounded-full',
                                tier.popular
                                  ? 'bg-[#1d9bf0]/20 text-[#1d9bf0] border border-[#1d9bf0]/40'
                                  : 'bg-[#1d9bf0]/15 text-[#1d9bf0]'
                              )}
                            >
                              {tier.badgeText}
                            </span>
                          </div>

                          {/* Username Role Preview */}
                          <div className="text-[11px] text-gray-400 my-2.5 bg-black/60 px-3 py-2 rounded-xl border border-white/5 flex items-center gap-1.5 min-w-0">
                            <span className="shrink-0 text-gray-500">Preview:</span>
                            <UserNameWithRole
                              displayName={previewUserName}
                              role={tier.key}
                              className="text-xs font-semibold"
                            />
                          </div>

                          {/* Price */}
                          <div className="my-3">
                            <div className="flex items-baseline gap-1">
                              <span className="text-2xl sm:text-3xl font-extrabold text-white">
                                {formattedPrice}
                              </span>
                              <span className="text-xs text-[#71767b] font-medium">/month</span>
                            </div>
                            <p className="text-[11px] text-[#71767b] mt-0.5">
                              Billed {billingCycle} via crypto
                            </p>
                          </div>

                          {/* Claim Offer CTA Button */}
                          <button
                            type="button"
                            onClick={() => handleSelectTierAndContinue(tier)}
                            className={cn(
                              'w-full py-2.5 px-4 rounded-full font-bold text-xs sm:text-sm transition-all my-2 shadow',
                              tier.popular
                                ? 'bg-white hover:bg-gray-200 text-black'
                                : 'bg-white/10 hover:bg-white/20 text-white border border-[#2f3336]'
                            )}
                          >
                            Claim offer
                          </button>

                          <div className="h-[1px] bg-[#2f3336] my-3.5" />

                          {/* Package Heading */}
                          <div className="text-[11px] text-[#71767b] font-medium mb-2.5">
                            {tier.packageHeading}
                          </div>

                          {/* Benefit List */}
                          <ul className="space-y-2 text-[12px]">
                            {tier.benefits.map((b, i) => (
                              <li key={i} className="flex items-start gap-2">
                                <Check className="w-3.5 h-3.5 text-[#1d9bf0] shrink-0 mt-0.5" />
                                <span className="text-gray-300 leading-snug">{b}</span>
                              </li>
                            ))}
                          </ul>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}

              {/* Bottom Disclaimer */}
              <p className="text-[11px] text-[#71767b] text-center leading-relaxed max-w-4xl mx-auto pt-2">
                Prices shown do not include applicable local taxes. Subscriptions activate
                authoritatively upon blockchain confirmation. Cancel renewal anytime from your
                account.
              </p>
            </div>
          )}

          {/* ================================================================ */}
          {/* STEP 2: GROK & PERKS INTRO CONFIRMATION MODAL (IMAGE 2)          */}
          {/* ================================================================ */}
          {step === 'agreement' && (
            <div className="p-6 sm:p-8 space-y-6">
              {/* Header with Back Arrow */}
              <div className="flex items-center justify-between">
                <button
                  type="button"
                  onClick={() => setStep('tiers')}
                  className="p-2 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                >
                  <ArrowLeft className="w-5 h-5" />
                </button>
                <button
                  onClick={closeUpgradeModal}
                  className="p-1.5 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Active Existing Order Notice Banner */}
              {orderNotice && (
                <div className="bg-blue-500/10 border border-blue-500/30 rounded-2xl p-3.5 text-xs text-blue-200 text-center flex items-center justify-between gap-3">
                  <span className="leading-snug">{orderNotice}</span>
                  <button
                    type="button"
                    onClick={() => setOrderNotice(null)}
                    className="text-gray-400 hover:text-white shrink-0 p-1 rounded-full hover:bg-white/10"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              )}

              {/* Interlocking Logos */}
              <div className="flex items-center justify-center gap-4 py-2">
                <img src="/LM.png" alt="LM Logo" className="w-10 h-10 object-contain" />
                <svg viewBox="0 0 24 24" className="w-9 h-9 stroke-white fill-none stroke-[2]">
                  <circle cx="12" cy="12" r="9" />
                  <line x1="5.5" y1="18.5" x2="18.5" y2="5.5" />
                </svg>
              </div>

              {/* Main Headline */}
              <div className="text-center space-y-2">
                <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
                  One more thing before you can use your X Subscription
                </h2>
                <p className="text-gray-400 text-sm max-w-md mx-auto">
                  Your X Subscription now includes Grok, the most advanced AI, all in one package.
                </p>
              </div>

              {/* Bullet Features */}
              <div className="bg-[#16181c] border border-[#2f3336] rounded-2xl p-5 space-y-4 max-w-md mx-auto">
                <div className="flex items-start gap-3.5">
                  <RefreshCw className="w-5 h-5 text-gray-400 shrink-0 mt-0.5" />
                  <p className="text-xs text-gray-300 leading-relaxed">
                    Agree to data sharing between X and Grok to sync your role privileges and
                    workspace.
                  </p>
                </div>
                <div className="flex items-start gap-3.5">
                  <Clock className="w-5 h-5 text-gray-400 shrink-0 mt-0.5" />
                  <p className="text-xs text-gray-300 leading-relaxed">
                    If we find an existing Grok account linked to your email, you will confirm that
                    you want to use that account with your new X Subscription.
                  </p>
                </div>
              </div>

              {/* Agree and Continue Button */}
              <div className="max-w-md mx-auto pt-2">
                <button
                  type="button"
                  onClick={() => setStep('configure')}
                  className="w-full bg-white hover:bg-gray-200 text-black font-bold py-3.5 px-6 rounded-full text-sm transition-colors shadow"
                >
                  Agree and continue
                </button>
              </div>

              {/* Legal Notice */}
              <p className="text-[11px] text-[#71767b] text-center max-w-md mx-auto leading-relaxed">
                By continuing, you agree to SpaceXSI's Terms of Service and Privacy Policy, and data
                sharing between X and SpaceXSI to identify your account.
              </p>
            </div>
          )}

          {/* ================================================================ */}
          {/* STEP 3: CONFIGURE YOUR PLAN & SELECT COIN (IMAGE 3)              */}
          {/* ================================================================ */}
          {step === 'configure' && selectedPlan && (
            <div className="p-5 sm:p-7 space-y-6">
              {/* Header with Back Arrow */}
              <div className="flex items-center justify-between border-b border-[#2f3336] pb-3">
                <div className="flex items-center gap-3">
                  <button
                    type="button"
                    onClick={() => setStep('agreement')}
                    className="p-1.5 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                  >
                    <ArrowLeft className="w-5 h-5" />
                  </button>
                  <h2 className="text-lg sm:text-xl font-bold">Configure your plan</h2>
                </div>
                <button
                  onClick={closeUpgradeModal}
                  className="p-1.5 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Active Existing Order Notice Banner */}
              {orderNotice && (
                <div className="bg-blue-500/10 border border-blue-500/30 rounded-2xl p-3.5 text-xs text-blue-200 text-center flex items-center justify-between gap-3">
                  <span className="leading-snug">{orderNotice}</span>
                  <button
                    type="button"
                    onClick={() => setOrderNotice(null)}
                    className="text-gray-400 hover:text-white shrink-0 p-1 rounded-full hover:bg-white/10"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              )}

              {/* Two Column Layout (Image 3) */}
              <div className="grid grid-cols-1 md:grid-cols-12 gap-6">
                {/* Left Column: Crypto Selection & Account Info */}
                <div className="md:col-span-7 space-y-4">
                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-gray-400 mb-2">
                      Pay with
                    </label>

                    {/* Search Crypto */}
                    <div className="relative mb-3">
                      <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                      <input
                        type="text"
                        placeholder="Search cryptocurrency (e.g. USDT, BTC, ETH)..."
                        value={coinSearch}
                        onChange={(e) => setCoinSearch(e.target.value)}
                        className="w-full pl-9 pr-3 py-2 bg-[#16181c] border border-[#2f3336] rounded-xl text-xs text-white placeholder-gray-500 outline-none focus:border-[#1d9bf0]"
                      />
                    </div>

                    {/* Coins Grid / List */}
                    {loadingCoins && currencies.length === 0 ? (
                      <div className="flex justify-center py-8">
                        <Loader2 className="w-6 h-6 animate-spin text-[#1d9bf0]" />
                      </div>
                    ) : currencies.length === 0 ? (
                      <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-xl text-center text-xs text-red-400">
                        {paymentError || 'No cryptocurrencies available from payment provider.'}
                      </div>
                    ) : (
                      <div className="grid grid-cols-2 sm:grid-cols-3 gap-2.5 max-h-[260px] overflow-y-auto pr-1">
                        {currencies
                          .filter(
                            (c) =>
                              c.name.toLowerCase().includes(coinSearch.toLowerCase()) ||
                              c.currency.toLowerCase().includes(coinSearch.toLowerCase()) ||
                              c.cid.toLowerCase().includes(coinSearch.toLowerCase())
                          )
                          .map((coin) => {
                            const isCoinSelected =
                              selectedCoin === coin.currency || selectedCoin === coin.cid;
                            const rate = coin.price_usd || coin.priceUsd;

                            return (
                              <button
                                key={coin.cid || coin.currency}
                                type="button"
                                onClick={() => setSelectedCoin(coin.currency || coin.cid)}
                                className={cn(
                                  'flex flex-col items-center justify-center p-3 rounded-xl border text-center transition-all',
                                  isCoinSelected
                                    ? 'bg-[#1d9bf0]/15 border-[#1d9bf0] shadow-[0_0_12px_rgba(29,155,240,0.3)]'
                                    : 'bg-[#16181c] border-[#2f3336] hover:border-[#444a50]'
                                )}
                              >
                                <img
                                  src={coin.icon || `/crypto-icons/${coin.currency}.svg`}
                                  alt={coin.name}
                                  className="w-7 h-7 object-contain mb-1.5"
                                  onError={(e) => {
                                    (e.currentTarget as HTMLImageElement).src =
                                      '/crypto-icons/USDT_TRX.svg';
                                  }}
                                />
                                <span className="text-xs font-bold text-white line-clamp-1">
                                  {coin.currency}
                                </span>
                                <span className="text-[10px] text-gray-400 line-clamp-1">
                                  {coin.name}
                                </span>
                                {rate && (
                                  <span className="text-[10px] text-emerald-400 font-mono mt-0.5">
                                    ${parseFloat(rate).toFixed(2)}
                                  </span>
                                )}
                              </button>
                            );
                          })}
                      </div>
                    )}
                  </div>

                  {/* Account Summary */}
                  <div className="bg-[#16181c] border border-[#2f3336] rounded-xl p-3.5 space-y-1.5">
                    <span className="text-[11px] font-bold uppercase tracking-wider text-[#71767b]">
                      Receiving Account
                    </span>
                    <div className="flex items-center justify-between">
                      <UserNameWithRole
                        displayName={currentUser?.display_name || currentUser?.username || 'You'}
                        role={currentUser?.role || 'member'}
                        className="text-sm font-semibold"
                      />
                      <span className="text-xs text-gray-400 font-mono">
                        {currentUser?.email || `@${currentUser?.username}`}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Right Column: Pricing Breakdown & Checkout Action */}
                <div className="md:col-span-5 bg-[#16181c] border border-[#2f3336] rounded-2xl p-5 flex flex-col justify-between space-y-5">
                  <div className="space-y-4">
                    <div className="border-b border-[#2f3336] pb-3">
                      <span className="text-xs text-[#71767b] font-medium uppercase tracking-wider">
                        Selected Role Plan
                      </span>
                      <h3 className="text-xl font-extrabold text-white mt-0.5">
                        {selectedPlan.name}
                      </h3>
                      <div className="flex items-baseline gap-1 mt-1">
                        <span className="text-2xl font-bold text-white">
                          $
                          {(
                            (billingCycle === 'annual'
                              ? selectedPlan.annualPriceCents
                              : selectedPlan.monthlyPriceCents) / 100
                          ).toFixed(2)}
                        </span>
                        <span className="text-xs text-[#71767b]">/{billingCycle}</span>
                      </div>
                    </div>

                    {/* Breakdown */}
                    <div className="space-y-2 text-xs">
                      <div className="flex justify-between text-gray-400">
                        <span>Plan price</span>
                        <span>
                          $
                          {(
                            (billingCycle === 'annual'
                              ? selectedPlan.annualPriceCents * 12
                              : selectedPlan.monthlyPriceCents) / 100
                          ).toFixed(2)}
                        </span>
                      </div>
                      <div className="flex justify-between font-bold text-white text-sm pt-2 border-t border-[#2f3336]">
                        <span>Due today</span>
                        <span className="text-cyan-400">
                          $
                          {(
                            (billingCycle === 'annual'
                              ? selectedPlan.annualPriceCents * 12
                              : selectedPlan.monthlyPriceCents) / 100
                          ).toFixed(2)}{' '}
                          USD
                        </span>
                      </div>
                      <div className="text-[10px] text-[#71767b]">Includes applicable taxes</div>
                    </div>

                    {/* Error display */}
                    {paymentError && (
                      <div className="p-3 bg-red-950/40 border border-red-800 rounded-xl text-xs text-red-300">
                        {paymentError}
                      </div>
                    )}
                  </div>

                  {/* Submit Button */}
                  <div className="space-y-3 pt-2">
                    <button
                      type="button"
                      onClick={handleInitiatePayment}
                      disabled={isSubmitting}
                      className="w-full bg-white hover:bg-gray-200 text-black font-bold py-3.5 px-6 rounded-full text-sm transition-all disabled:opacity-60 flex items-center justify-center gap-2 shadow"
                    >
                      {isSubmitting ? (
                        <>
                          <Loader2 className="w-4 h-4 animate-spin" />
                          <span>Generating Payment...</span>
                        </>
                      ) : (
                        <span>Agree, Subscribe & Pay</span>
                      )}
                    </button>

                    <p className="text-[10px] text-[#71767b] text-center leading-snug">
                      By clicking Agree, Subscribe & Pay, you agree to our Terms. A direct
                      cryptographic payment session will be generated.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ================================================================ */}
          {/* STEP 4: LIVE PAYMENT PHASE (MATCHING SAWER PATTERN)              */}
          {/* ================================================================ */}
          {step === 'payment' && (
            <div className="p-6 sm:p-8 space-y-6">
              {/* Top Navigation */}
              <div className="flex items-center justify-between border-b border-[#2f3336] pb-3">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs text-gray-500">
                    Order #{paymentData?.order_id?.slice(0, 8)}
                  </span>
                </div>
                <button
                  onClick={closeUpgradeModal}
                  className="p-1.5 rounded-full hover:bg-white/10 text-gray-400 hover:text-white transition-colors"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Active Existing Order Notice Banner */}
              {orderNotice && (
                <div className="bg-blue-500/10 border border-blue-500/30 rounded-2xl p-3.5 text-xs text-blue-200 text-center flex items-center justify-between gap-3">
                  <span className="leading-snug">{orderNotice}</span>
                  <button
                    type="button"
                    onClick={() => setOrderNotice(null)}
                    className="text-gray-400 hover:text-white shrink-0 p-1 rounded-full hover:bg-white/10"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              )}

              {/* Sub-screens based on status */}
              {/* 1. Loading */}
              {paymentScreen === 'loading' && (
                <div className="flex flex-col items-center justify-center py-14 space-y-3 text-center">
                  <Loader2 className="w-10 h-10 animate-spin text-[#1d9bf0]" />
                  <p className="text-gray-400 text-sm">Loading payment details...</p>
                </div>
              )}

              {/* 2. Error */}
              {paymentScreen === 'error' && (
                <div className="text-center py-8 space-y-4">
                  <XCircle className="w-12 h-12 text-red-500 mx-auto" />
                  <h3 className="text-lg font-bold">Transaction Issue</h3>
                  <p className="text-xs text-gray-400 max-w-sm mx-auto">{paymentError}</p>
                  <button
                    type="button"
                    onClick={() => setStep('configure')}
                    className="px-5 py-2 rounded-full border border-[#2f3336] hover:bg-white/10 text-xs font-bold"
                  >
                    Go back and try again
                  </button>
                </div>
              )}

              {/* 3. Success (Role Activated!) */}
              {paymentScreen === 'success' && (
                <div className="text-center py-6 space-y-5 animate-in zoom-in-95">
                  <div className="w-20 h-20 rounded-full bg-emerald-500/20 border border-emerald-500/40 flex items-center justify-center mx-auto shadow-[0_0_30px_rgba(16,185,129,0.3)]">
                    <CheckCircle2 className="w-10 h-10 text-emerald-400" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-extrabold uppercase tracking-wide text-white">
                      Payment Successful!
                    </h2>
                    <p className="text-gray-400 text-sm mt-1 max-w-sm mx-auto">
                      Your subscription is active. Your verified checkmark and role privileges have
                      been activated.
                    </p>
                  </div>

                  <div className="bg-[#16181c] border border-[#2f3336] rounded-xl p-3 inline-block font-mono text-sm font-semibold text-emerald-400">
                    {paymentData?.crypto_amount} {paymentData?.coin || 'USDT'}
                  </div>

                  <div>
                    <button
                      type="button"
                      onClick={() => {
                        closeUpgradeModal();
                        window.location.reload();
                      }}
                      className="bg-white hover:bg-gray-200 text-black font-bold py-3 px-8 rounded-full text-sm transition-colors shadow"
                    >
                      Continue to X
                    </button>
                  </div>
                </div>
              )}

              {/* 4. Expired or Cancelled */}
              {paymentScreen === 'expired' && (
                <div className="text-center py-8 space-y-4">
                  <XCircle className="w-12 h-12 text-gray-500 mx-auto" />
                  <h3 className="text-lg font-bold">Payment Expired or Cancelled</h3>
                  <p className="text-xs text-gray-400">
                    No payment was received in the allotted time. You can initiate a new checkout.
                  </p>
                  <button
                    type="button"
                    onClick={() => setStep('configure')}
                    className="px-5 py-2.5 rounded-full bg-white text-black font-bold text-xs"
                  >
                    Try Again
                  </button>
                </div>
              )}

              {/* 5. QR Code & Deposit Address (Pending payment) */}
              {paymentScreen === 'qr' && paymentData && (
                <div className="text-center space-y-5 max-w-md mx-auto">
                  <div>
                    <h3 className="text-lg font-bold flex items-center justify-center gap-2">
                      <QrCode className="w-5 h-5 text-[#1d9bf0]" /> Scan to pay with{' '}
                      {paymentData.coin}
                    </h3>
                    <p className="text-xs text-gray-400 mt-1">
                      Send the exact amount below to the specified address.
                    </p>
                  </div>

                  {/* Click to Copy Amount */}
                  <div className="bg-[#16181c] border border-[#2f3336] rounded-2xl p-4">
                    <span className="text-[11px] text-[#71767b] uppercase tracking-wider font-semibold block mb-1">
                      Amount to Send
                    </span>
                    <button
                      type="button"
                      onClick={() =>
                        copyToClipboard(
                          paymentData.crypto_amount || paymentData.pending_amount || '',
                          'amount'
                        )
                      }
                      className="flex items-center justify-center gap-2 text-2xl sm:text-3xl font-extrabold text-[#1d9bf0] hover:text-[#1d9bf0]/80 transition-colors mx-auto"
                      title="Click to copy amount"
                    >
                      <span>{paymentData.crypto_amount || paymentData.pending_amount}</span>
                      <span className="text-lg text-white">{paymentData.coin}</span>
                      {copiedAmount ? (
                        <Check className="w-5 h-5 text-emerald-400" />
                      ) : (
                        <Copy className="w-4 h-4 opacity-60 hover:opacity-100" />
                      )}
                    </button>
                    <span className="text-[10px] text-gray-500 mt-1 block">
                      {copiedAmount ? 'Copied to clipboard!' : 'Click amount to copy'}
                    </span>
                  </div>

                  {/* QR Code Container */}
                  <div className="mx-auto w-56 h-56 bg-white p-3 rounded-2xl shadow flex items-center justify-center">
                    {paymentData.qr_code ? (
                      <img
                        src={paymentData.qr_code}
                        alt={`${paymentData.coin} payment QR`}
                        className="w-full h-full object-contain"
                      />
                    ) : (
                      <div className="flex flex-col items-center justify-center text-gray-800">
                        <QrCode className="w-12 h-12 mb-2 text-gray-400" />
                        <span className="text-xs font-semibold">Generating QR...</span>
                      </div>
                    )}
                  </div>

                  {/* Wallet Hash Address */}
                  {paymentData.wallet_hash && (
                    <div className="bg-[#16181c] border border-[#2f3336] rounded-xl p-3.5 text-left">
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="flex items-center gap-1.5 text-xs text-gray-400">
                          <Wallet className="w-4 h-4 text-[#1d9bf0]" /> Deposit Address
                        </span>
                        <button
                          type="button"
                          onClick={() => copyToClipboard(paymentData.wallet_hash || '', 'address')}
                          className="text-xs bg-white/10 hover:bg-white/20 text-white px-2.5 py-1 rounded-md flex items-center gap-1 font-semibold transition-colors"
                        >
                          {copiedAddress ? (
                            <>
                              <Check className="w-3.5 h-3.5 text-emerald-400" /> Copied
                            </>
                          ) : (
                            <>
                              <Copy className="w-3.5 h-3.5" /> Copy
                            </>
                          )}
                        </button>
                      </div>
                      <code className="block text-xs font-mono break-all text-gray-200 select-all">
                        {paymentData.wallet_hash}
                      </code>
                    </div>
                  )}

                  {/* Status Indicator & Cancel */}
                  <div className="flex items-center justify-center gap-2 text-xs text-[#71767b]">
                    <Clock className="w-4 h-4 text-[#1d9bf0]" />
                    <span>Awaiting your transfer. Invoice expires in 60 minutes.</span>
                  </div>

                  <div className="pt-3 flex flex-col items-center gap-2 border-t border-[#2f3336]/40 mt-3">
                    <button
                      type="button"
                      disabled={isCancelling}
                      onClick={handleCancelOrder}
                      className="text-xs font-semibold text-red-400 hover:text-red-300 hover:underline transition-colors flex items-center gap-1.5 disabled:opacity-50"
                    >
                      {isCancelling ? (
                        <>
                          <Loader2 className="w-3.5 h-3.5 animate-spin" />
                          <span>Cancelling order...</span>
                        </>
                      ) : (
                        <span>Cancel this order to choose another role or coin</span>
                      )}
                    </button>
                  </div>
                </div>
              )}

              {/* 6. Processing / Confirming */}
              {paymentScreen === 'processing' && paymentData && (
                <div className="relative text-center py-8 space-y-5 max-w-md mx-auto overflow-hidden">
                  <div className="w-20 h-20 rounded-full bg-cyan-500/20 border border-cyan-500/40 flex items-center justify-center mx-auto shadow-[0_0_30px_rgba(6,182,212,0.3)]">
                    <Clock className="w-10 h-10 text-cyan-400 animate-pulse" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-extrabold uppercase tracking-wide text-white">
                      Awaiting Confirmation
                    </h2>
                    <p className="text-gray-400 text-sm mt-2 leading-relaxed">
                      Payment detected on the blockchain. Please wait while nodes confirm the
                      transaction. Your role will activate automatically.
                    </p>
                  </div>

                  <div className="flex items-center justify-center gap-2 text-xs text-cyan-400">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span>Checking blockchain confirmations...</span>
                  </div>
                </div>
              )}

              {/* 7. Partial Payment Received (Amount kurang!) */}
              {paymentScreen === 'partial' && paymentData && (
                <div className="text-center space-y-5 max-w-md mx-auto">
                  <div className="w-16 h-16 rounded-full bg-amber-500/20 border border-amber-500/40 flex items-center justify-center mx-auto">
                    <Clock className="w-8 h-8 text-amber-400" />
                  </div>

                  <div>
                    <h3 className="text-lg font-bold text-amber-400">Partial Payment Received</h3>
                    <p className="text-xs text-gray-400 mt-1">
                      We received {received.toFixed(8)} {paymentData.coin}. To complete your
                      activation, please send the remaining amount:
                    </p>
                  </div>

                  {/* Red highlight for remaining amount */}
                  <div className="bg-[#16181c] border border-red-500/40 rounded-2xl p-4">
                    <span className="text-[11px] text-red-400 font-semibold uppercase tracking-wider block mb-1">
                      Remaining Amount Due
                    </span>
                    <button
                      type="button"
                      onClick={() =>
                        copyToClipboard(
                          (pendingRemaining || required - received).toFixed(8),
                          'amount'
                        )
                      }
                      className="flex items-center justify-center gap-2 text-3xl font-extrabold text-red-400 hover:text-red-300 transition-colors mx-auto"
                      title="Click to copy remaining amount"
                    >
                      <span>{(pendingRemaining || required - received).toFixed(8)}</span>
                      <span className="text-lg text-white">{paymentData.coin}</span>
                      {copiedAmount ? (
                        <Check className="w-5 h-5 text-emerald-400" />
                      ) : (
                        <Copy className="w-4 h-4 opacity-60 hover:opacity-100" />
                      )}
                    </button>
                    <span className="text-[10px] text-gray-500 mt-1 block">
                      {copiedAmount ? 'Copied to clipboard!' : 'Click to copy remaining amount'}
                    </span>
                  </div>

                  {/* Wallet Hash Address */}
                  {paymentData.wallet_hash && (
                    <div className="bg-[#16181c] border border-[#2f3336] rounded-xl p-3.5 text-left">
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="flex items-center gap-1.5 text-xs text-gray-400">
                          <Wallet className="w-4 h-4 text-[#1d9bf0]" /> Send remainder to this
                          address
                        </span>
                        <button
                          type="button"
                          onClick={() => copyToClipboard(paymentData.wallet_hash || '', 'address')}
                          className="text-xs bg-white/10 hover:bg-white/20 text-white px-2.5 py-1 rounded-md flex items-center gap-1 font-semibold transition-colors"
                        >
                          {copiedAddress ? (
                            <>
                              <Check className="w-3.5 h-3.5 text-emerald-400" /> Copied
                            </>
                          ) : (
                            <>
                              <Copy className="w-3.5 h-3.5" /> Copy
                            </>
                          )}
                        </button>
                      </div>
                      <code className="block text-xs font-mono break-all text-gray-200 select-all">
                        {paymentData.wallet_hash}
                      </code>
                    </div>
                  )}

                  <div className="flex items-center justify-center gap-2 text-xs text-amber-400">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span>Waiting for remaining payment...</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        onSuccess={(user) => {
          updateUser(user);
        }}
      />
    </>
  );
}
