'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import {
  Check,
  Zap,
  ShieldCheck,
  Sparkles,
  ArrowRight,
  ExternalLink,
  Loader2,
  Clock,
  FileText,
  AlertCircle,
  HelpCircle,
} from 'lucide-react';
import { LeftNavRail } from '@/components/layout/LeftNavRail';
import { RightSidebar } from '@/components/layout/RightSidebar';
import { AuthModal } from '@/components/auth/AuthModal';
import { ConfirmModal } from '@/components/common/ConfirmModal';
import { VerifiedBadge } from '@/components/common/VerifiedBadge';
import { User, Subscription, Entitlements } from '@/types';
import { apiRequest } from '@/lib/api';
import { cn } from '@/lib/utils';
import { useCurrentUser } from '@/lib/useCurrentUser';

export default function PremiumPage() {
  const router = useRouter();
  const { currentUser, updateUser: setCurrentUser } = useCurrentUser();

  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'annual'>('annual');
  const [selectedTier, setSelectedTier] = useState<'basic' | 'premium' | 'premium_plus'>('premium');
  const [loading, setLoading] = useState(false);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [entitlements, setEntitlements] = useState<Entitlements | null>(null);
  const [checkoutUrl, setCheckoutUrl] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'subscribe' | 'history'>('subscribe');
  const [billingOrders, setBillingOrders] = useState<any[]>([]);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);

  useEffect(() => {
    loadUserData();
  }, [currentUser]);

  const loadUserData = () => {
    apiRequest<{ subscription: Subscription | null; entitlements: Entitlements }>('/subscriptions/me')
      .then((res) => {
        setSubscription(res.subscription);
        setEntitlements(res.entitlements);
      })
      .catch(() => {
        setCurrentUser(null);
      });
  };

  const handleCheckout = async (simulate = false) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    setLoading(true);
    setCheckoutUrl(null);
    setSuccessMessage(null);

    try {
      const res = await apiRequest<{
        order_id: string;
        invoice_url: string;
        amount_cents: number;
        currency: string;
        white_label?: any;
      }>('/subscriptions/checkout', {
        method: 'POST',
        body: JSON.stringify({
          plan_id: selectedTier,
          billing_cycle: billingCycle,
          currency: 'USDT_TRX',
        }),
      });

      // White-Label checkout: navigate directly to internal payment page without external redirect
      router.push(`/premium/invoice?order_id=${res.order_id}${simulate ? '&simulated=true' : ''}`);
    } catch (err: any) {
      alert(err.message || 'Failed to initiate checkout');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    if (!subscription) return;
    setCancelModalOpen(true);
  };

  const handleConfirmCancel = async () => {
    if (!subscription) return;
    setIsCanceling(true);
    try {
      await apiRequest(`/subscriptions/${subscription.id}/cancel`, { method: 'POST' });
      setCancelModalOpen(false);
      loadUserData();
    } catch (err: any) {
      alert(err.message || 'Failed to cancel subscription');
    } finally {
      setIsCanceling(false);
    }
  };

  const loadHistory = async () => {
    try {
      const res = await apiRequest<{ orders: any[]; invoices: any[] }>('/billing/history');
      setBillingOrders(res.orders || []);
    } catch (e) {
      console.error(e);
    }
  };

  const pricing = {
    basic: { monthly: 3, annual: 30, discount: 'Save 17%' },
    premium: { monthly: 8, annual: 84, discount: 'Save 12%' },
    premium_plus: { monthly: 16, annual: 168, discount: 'Save 12%' },
  };

  return (
    <>
      <main className="w-full max-w-[600px] border-r border-border min-h-screen pb-16">
        {/* Top Header */}
        <div className="sticky top-0 z-20 backdrop-blur-md bg-background/80 border-b border-border px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-[20px] font-bold">Premium</h1>
          </div>
          <div className="flex bg-subtle p-0.5 rounded-full text-[13px] font-semibold">
            <button
              onClick={() => setActiveTab('subscribe')}
              className={cn(
                'px-3 py-1 rounded-full transition-colors',
                activeTab === 'subscribe' ? 'bg-background text-foreground shadow-sm' : 'text-muted hover:text-foreground'
              )}
            >
              Plans
            </button>
            <button
              onClick={() => {
                setActiveTab('history');
                loadHistory();
              }}
              className={cn(
                'px-3 py-1 rounded-full transition-colors',
                activeTab === 'history' ? 'bg-background text-foreground shadow-sm' : 'text-muted hover:text-foreground'
              )}
            >
              Billing History
            </button>
          </div>
        </div>

        {activeTab === 'subscribe' ? (
          <div className="p-4 sm:p-6 space-y-6">
            {/* Active Subscription Banner */}
            {subscription && subscription.status === 'active' && (
              <div className="bg-gradient-to-r from-blue-900/40 via-blue-800/20 to-neutral-900 border border-blue-500/40 rounded-2xl p-5 shadow-lg">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-[20px] font-bold capitalize">{subscription.plan_id.replace('_', ' ')} Member</span>
                    <VerifiedBadge type={entitlements?.badge_type || 'blue'} size="md" />
                  </div>
                  <span className="text-[12px] bg-blue-500/20 text-[#1d9bf0] font-bold px-2.5 py-0.5 rounded-full uppercase tracking-wider">
                    Active
                  </span>
                </div>
                <p className="text-[14px] text-muted mb-3">
                  Your subscription renews on {new Date(subscription.current_period_end).toLocaleDateString()}.
                  {subscription.cancel_at_period_end && ' (Cancels at end of period)'}
                </p>
                <div className="flex gap-2">
                  {!subscription.cancel_at_period_end && (
                    <button
                      onClick={handleCancel}
                      className="text-[13px] text-red-400 hover:text-red-300 border border-red-500/30 hover:bg-red-500/10 px-3 py-1 rounded-full font-semibold transition-colors"
                    >
                      Cancel renewal
                    </button>
                  )}
                </div>
              </div>
            )}

            {/* Hero Heading */}
            <div className="text-center space-y-2 py-2">
              <h2 className="text-[28px] sm:text-[32px] font-extrabold tracking-tight">
                Upgrade to Premium
              </h2>
              <p className="text-muted text-[15px] max-w-[460px] mx-auto leading-relaxed">
                Enjoy an enhanced experience, verified profile checkmark, creator revenue eligibility, and higher posting limits.
              </p>
            </div>

            {/* Monthly / Annual Toggle */}
            <div className="flex justify-center">
              <div className="bg-subtle p-1 rounded-full flex items-center border border-border">
                <button
                  type="button"
                  onClick={() => setBillingCycle('annual')}
                  className={cn(
                    'px-4 py-1.5 rounded-full text-[14px] font-bold transition-all flex items-center gap-1.5',
                    billingCycle === 'annual'
                      ? 'bg-foreground text-background shadow-md'
                      : 'text-muted hover:text-foreground'
                  )}
                >
                  Annual
                  <span className="text-[11px] bg-green-500/20 text-green-400 px-1.5 py-0.5 rounded-full font-extrabold">
                    SAVE 12%
                  </span>
                </button>
                <button
                  type="button"
                  onClick={() => setBillingCycle('monthly')}
                  className={cn(
                    'px-4 py-1.5 rounded-full text-[14px] font-bold transition-all',
                    billingCycle === 'monthly'
                      ? 'bg-foreground text-background shadow-md'
                      : 'text-muted hover:text-foreground'
                  )}
                >
                  Monthly
                </button>
              </div>
            </div>

            {/* Plan Cards Grid */}
            <div className="grid grid-cols-1 gap-4">
              {/* Basic Plan */}
              <div
                onClick={() => setSelectedTier('basic')}
                className={cn(
                  'border rounded-2xl p-5 cursor-pointer transition-all relative',
                  selectedTier === 'basic'
                    ? 'border-[#1d9bf0] bg-[#1d9bf0]/5 shadow-[0_0_20px_rgba(29,155,240,0.15)]'
                    : 'border-border hover:border-neutral-700 bg-subtle/30'
                )}
              >
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <h3 className="text-[19px] font-bold">Basic</h3>
                    <p className="text-[13px] text-muted">Essential enhancement tools</p>
                  </div>
                  <div className="text-right">
                    <span className="text-[22px] font-extrabold">
                      ${billingCycle === 'annual' ? pricing.basic.annual : pricing.basic.monthly}
                    </span>
                    <span className="text-muted text-[13px]">
                      /{billingCycle === 'annual' ? 'year' : 'month'}
                    </span>
                  </div>
                </div>
                <ul className="space-y-2 mt-4 text-[14px] text-neutral-300">
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Edit posts within 1 hour
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Post up to 1,000 characters
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    10-minute 1080p video uploads
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Bookmark folders & Undo send
                  </li>
                </ul>
              </div>

              {/* Premium Plan (Most Popular) */}
              <div
                onClick={() => setSelectedTier('premium')}
                className={cn(
                  'border-2 rounded-2xl p-5 cursor-pointer transition-all relative overflow-hidden',
                  selectedTier === 'premium'
                    ? 'border-[#1d9bf0] bg-[#1d9bf0]/10 shadow-[0_0_25px_rgba(29,155,240,0.25)]'
                    : 'border-border hover:border-neutral-700 bg-subtle/40'
                )}
              >
                <div className="absolute top-0 right-0 bg-[#1d9bf0] text-white text-[11px] font-extrabold px-3 py-0.5 rounded-bl-xl uppercase tracking-wider">
                  Most Popular
                </div>
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <div className="flex items-center gap-1.5">
                      <h3 className="text-[20px] font-bold">Premium</h3>
                      <VerifiedBadge type="blue" size="md" showPopover={false} />
                    </div>
                    <p className="text-[13px] text-muted">The signature X verified experience</p>
                  </div>
                  <div className="text-right">
                    <span className="text-[24px] font-extrabold">
                      ${billingCycle === 'annual' ? pricing.premium.annual : pricing.premium.monthly}
                    </span>
                    <span className="text-muted text-[13px]">
                      /{billingCycle === 'annual' ? 'year' : 'month'}
                    </span>
                  </div>
                </div>
                <ul className="space-y-2.5 mt-4 text-[14px]">
                  <li className="flex items-center gap-2 font-semibold text-white">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>Blue Verified Badge on profile and replies</span>
                  </li>
                  <li className="flex items-center gap-2 text-neutral-200">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>50% fewer ads in For You and Following</span>
                  </li>
                  <li className="flex items-center gap-2 text-neutral-200">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>Post longer content up to 4,000 characters</span>
                  </li>
                  <li className="flex items-center gap-2 text-neutral-200">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>60-minute 1080p video uploads to Cloudflare R2</span>
                  </li>
                  <li className="flex items-center gap-2 text-neutral-200">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>Prioritized rankings in conversations</span>
                  </li>
                  <li className="flex items-center gap-2 text-neutral-200">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    <span>Creator analytics & revenue sharing eligibility</span>
                  </li>
                </ul>
              </div>

              {/* Premium+ Plan */}
              <div
                onClick={() => setSelectedTier('premium_plus')}
                className={cn(
                  'border rounded-2xl p-5 cursor-pointer transition-all relative',
                  selectedTier === 'premium_plus'
                    ? 'border-[#1d9bf0] bg-[#1d9bf0]/5 shadow-[0_0_20px_rgba(29,155,240,0.15)]'
                    : 'border-border hover:border-neutral-700 bg-subtle/30'
                )}
              >
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <h3 className="text-[19px] font-bold">Premium+</h3>
                    <p className="text-[13px] text-muted">The ultimate ad-free experience</p>
                  </div>
                  <div className="text-right">
                    <span className="text-[22px] font-extrabold">
                      ${billingCycle === 'annual' ? pricing.premium_plus.annual : pricing.premium_plus.monthly}
                    </span>
                    <span className="text-muted text-[13px]">
                      /{billingCycle === 'annual' ? 'year' : 'month'}
                    </span>
                  </div>
                </div>
                <ul className="space-y-2 mt-4 text-[14px] text-neutral-300">
                  <li className="flex items-center gap-2 font-semibold text-white">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Zero ads in For You and Following
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Largest reply ranking boost
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    Write articles & post up to 10,000 characters
                  </li>
                  <li className="flex items-center gap-2">
                    <Check className="w-4 h-4 text-[#1d9bf0]" />
                    3-hour video uploads
                  </li>
                </ul>
              </div>
            </div>

            {/* Success notification */}
            {successMessage && (
              <div className="bg-green-500/10 border border-green-500/40 rounded-xl p-4 text-green-400 text-[14px] flex items-center gap-3">
                <Check className="w-5 h-5 shrink-0" />
                <span>{successMessage}</span>
              </div>
            )}

            {/* Plisio Invoice Redirection Banner */}
            {checkoutUrl && (
              <div className="bg-[#1d9bf0]/10 border border-[#1d9bf0]/40 rounded-xl p-4 text-[14px] flex flex-col gap-2">
                <span className="font-bold text-white flex items-center gap-2">
                  <ExternalLink className="w-4 h-4 text-[#1d9bf0]" />
                  Plisio Cryptocurrency Invoice Created
                </span>
                <p className="text-muted text-[13px]">
                  Your invoice has been generated with Plisio gateway. Complete the payment in cryptocurrency (BTC, USDT, ETH, etc.) to activate your subscription.
                </p>
                <a
                  href={checkoutUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white font-bold py-2 px-4 rounded-full text-center transition-colors inline-block mt-1"
                >
                  Open Plisio Invoice
                </a>
              </div>
            )}

            {/* Checkout Actions */}
            <div className="space-y-3 pt-2">
              <button
                type="button"
                onClick={() => handleCheckout(false)}
                disabled={loading}
                className="w-full bg-[#1d9bf0] hover:bg-[#1a8cd8] text-white font-extrabold text-[16px] py-3.5 px-6 rounded-full shadow-lg transition-all flex items-center justify-center gap-2 active:scale-98"
              >
                {loading ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : (
                  <>
                    <span>Subscribe with Crypto (Plisio)</span>
                    <ArrowRight className="w-5 h-5" />
                  </>
                )}
              </button>

              {/* Dev Simulation Button */}
              <button
                type="button"
                onClick={() => handleCheckout(true)}
                disabled={loading}
                className="w-full bg-subtle hover:bg-neutral-800 border border-border text-neutral-300 font-semibold text-[14px] py-2.5 px-4 rounded-full transition-colors flex items-center justify-center gap-2"
              >
                <Zap className="w-4 h-4 text-yellow-400" />
                <span>Simulate Instant Payment (Dev Mode)</span>
              </button>
            </div>

            {/* Feature Comparison Matrix */}
            <div className="mt-8 border-t border-border pt-6">
              <h3 className="font-bold text-[18px] mb-4">Compare all features</h3>
              <div className="border border-border rounded-xl overflow-hidden text-[13px]">
                <table className="w-full text-left">
                  <thead className="bg-subtle/50 border-b border-border text-muted">
                    <tr>
                      <th className="p-3">Feature</th>
                      <th className="p-3 text-center">Free</th>
                      <th className="p-3 text-center">Basic</th>
                      <th className="p-3 text-center text-[#1d9bf0]">Premium</th>
                      <th className="p-3 text-center">Premium+</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    <tr>
                      <td className="p-3 font-medium">Post Length</td>
                      <td className="p-3 text-center text-muted">280</td>
                      <td className="p-3 text-center font-medium">1,000</td>
                      <td className="p-3 text-center font-bold text-[#1d9bf0]">4,000</td>
                      <td className="p-3 text-center font-bold">10,000</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium">Verified Blue Checkmark</td>
                      <td className="p-3 text-center text-muted">-</td>
                      <td className="p-3 text-center text-muted">-</td>
                      <td className="p-3 text-center text-[#1d9bf0]">✔</td>
                      <td className="p-3 text-center text-[#1d9bf0]">✔</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium">Ad Reduction</td>
                      <td className="p-3 text-center text-muted">None</td>
                      <td className="p-3 text-center text-muted">None</td>
                      <td className="p-3 text-center font-medium">50% Less</td>
                      <td className="p-3 text-center font-bold text-green-400">100% Ad-Free</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium">Video Upload Length</td>
                      <td className="p-3 text-center text-muted">2 min</td>
                      <td className="p-3 text-center font-medium">10 min</td>
                      <td className="p-3 text-center font-bold">60 min</td>
                      <td className="p-3 text-center font-bold">3 hours</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium">Edit Post Window</td>
                      <td className="p-3 text-center text-muted">-</td>
                      <td className="p-3 text-center">1 Hour</td>
                      <td className="p-3 text-center">1 Hour</td>
                      <td className="p-3 text-center">1 Hour</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        ) : (
          /* Billing History Tab */
          <div className="p-4 sm:p-6 space-y-4">
            <h3 className="font-bold text-[18px]">Payment & Order History</h3>
            {billingOrders.length === 0 ? (
              <div className="p-8 text-center text-muted border border-dashed border-border rounded-xl">
                <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p>No billing transactions found for your account.</p>
              </div>
            ) : (
              <div className="space-y-3">
                {billingOrders.map((order) => (
                  <div
                    key={order.id}
                    className="border border-border rounded-xl p-4 flex items-center justify-between bg-subtle/20"
                  >
                    <div>
                      <div className="font-bold capitalize text-[15px]">
                        {order.plan_id.replace('_', ' ')} Plan ({order.billing_cycle})
                      </div>
                      <div className="text-[13px] text-muted">
                        {new Date(order.created_at).toLocaleDateString()} · Plisio Crypto
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-extrabold text-[16px]">
                        ${(order.amount_cents / 100).toFixed(2)}
                      </div>
                      <span
                        className={cn(
                          'text-[12px] font-bold uppercase tracking-wider',
                          order.status === 'completed'
                            ? 'text-green-400'
                            : order.status === 'pending'
                            ? 'text-yellow-400'
                            : 'text-red-400'
                        )}
                      >
                        {order.status}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </main>

      <RightSidebar />

      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        onSuccess={(user) => {
          setCurrentUser(user);
          loadUserData();
        }}
      />

      <ConfirmModal
        isOpen={cancelModalOpen}
        onClose={() => setCancelModalOpen(false)}
        onConfirm={handleConfirmCancel}
        title="Cancel subscription?"
        description="Your subscription renewal will be canceled and verification benefits will end at the end of the current billing period."
        confirmText="Cancel subscription"
        cancelText="Keep subscription"
        variant="destructive"
        isLoading={isCanceling}
      />
    </>
  );
}
