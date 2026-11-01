'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  Check,
  Zap,
  Shield,
  Sparkles,
  Star,
  Loader2,
  FileText,
  Search,
  X,
} from 'lucide-react';
import { RightSidebar } from '@/components/layout/RightSidebar';
import { AuthModal } from '@/components/auth/AuthModal';
import { ConfirmModal } from '@/components/common/ConfirmModal';
import { UserNameWithRole } from '@/components/ui/UserNameWithRole';
import { Subscription, Entitlements, MembershipPlan } from '@/types';
import { apiRequest } from '@/lib/api';
import { cn } from '@/lib/utils';
import { useCurrentUser } from '@/lib/useCurrentUser';

interface CryptoCurrencyOption {
  cid: string;
  currency: string;
  name: string;
  network?: string;
  icon: string;
  price_usd?: string;
  priceUsd?: string;
}

const ROLE_TIER_DEFAULTS: Record<
  string,
  {
    role: string;
    name: string;
    defaultPriceCents: number;
    icon: React.ReactNode;
    highlights: string[];
  }
> = {
  vip: {
    role: 'vip',
    name: 'VIP Member',
    defaultPriceCents: 100,
    icon: <Star className="w-5 h-5 text-blue-400" />,
    highlights: ['VIP Role Badge & Styled Username', 'Basic analytics', 'Custom profile colors'],
  },
  mvp: {
    role: 'mvp',
    name: 'MVP',
    defaultPriceCents: 100,
    icon: <Zap className="w-5 h-5 text-yellow-400" />,
    highlights: ['Everything in VIP', 'Animated MVP Role Badge', 'Increased upload limits'],
  },
  mod: {
    role: 'mod',
    name: 'Moderator',
    defaultPriceCents: 100,
    icon: <Shield className="w-5 h-5 text-green-400" />,
    highlights: ['Everything in MVP', 'Moderator Role Badge & Tools', 'Content management'],
  },
  god: {
    role: 'god',
    name: 'God',
    defaultPriceCents: 100,
    icon: <Sparkles className="w-5 h-5 text-purple-400" />,
    highlights: ['Everything in Mod', 'God Mode Role & Privileges', 'Bypass standard limits'],
  },
};

const PURCHASABLE_KEYS = ['vip', 'mvp', 'mod', 'god'] as const;

export default function PremiumPage() {
  const router = useRouter();
  const { currentUser, updateUser: setCurrentUser } = useCurrentUser();

  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [selectedTier, setSelectedTier] = useState<string>('vip');
  const [loading, setLoading] = useState(false);
  const [plansLoading, setPlansLoading] = useState(true);
  const [membershipPlans, setMembershipPlans] = useState<MembershipPlan[]>([]);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [entitlements, setEntitlements] = useState<Entitlements | null>(null);
  const [activeTab, setActiveTab] = useState<'subscribe' | 'history'>('subscribe');
  const [billingOrders, setBillingOrders] = useState<any[]>([]);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [isCanceling, setIsCanceling] = useState(false);

  // Crypto Modal State
  const [showCryptoModal, setShowCryptoModal] = useState(false);
  const [currencies, setCurrencies] = useState<CryptoCurrencyOption[]>([]);
  const [loadingCryptos, setLoadingCryptos] = useState(false);
  const [creatingInvoice, setCreatingInvoice] = useState(false);
  const [cryptoSearch, setCryptoSearch] = useState('');

  useEffect(() => {
    loadPlans();
  }, []);

  useEffect(() => {
    if (currentUser) {
      loadUserData();
    }
  }, [currentUser]);

  const loadPlans = async () => {
    setPlansLoading(true);
    try {
      const res = await apiRequest<{ plans: MembershipPlan[] }>('/membership/plans');
      setMembershipPlans(res.plans || []);
    } catch (err) {
      console.error('Failed to load membership plans', err);
    } finally {
      setPlansLoading(false);
    }
  };

  const loadUserData = () => {
    apiRequest<{ subscription: Subscription | null; entitlements: Entitlements }>('/subscriptions/me')
      .then((res) => {
        setSubscription(res.subscription);
        setEntitlements(res.entitlements);
      })
      .catch(() => {
        // Ignore if not logged in
      });
  };

  const getTierDisplayList = () => {
    return PURCHASABLE_KEYS.map((key) => {
      const dbPlan = membershipPlans.find((p) => p.key === key || p.id === key);
      const defaults = ROLE_TIER_DEFAULTS[key];
      const priceCents = dbPlan ? dbPlan.price_amount_cents : defaults.defaultPriceCents;
      const formattedPrice =
        priceCents % 100 === 0
          ? `$${(priceCents / 100).toFixed(0)}`
          : `$${(priceCents / 100).toFixed(2)}`;

      const quotaFeatures: string[] = [];
      if (dbPlan?.entitlements && dbPlan.entitlements.length > 0) {
        const entMap: Record<string, number> = {};
        dbPlan.entitlements.forEach((e) => {
          entMap[e.entitlement_key] = e.limit_value;
        });
        if (entMap['posts_per_day']) {
          quotaFeatures.push(`${entMap['posts_per_day'].toLocaleString()} posts per day`);
        }
        if (entMap['max_post_length']) {
          quotaFeatures.push(`Up to ${entMap['max_post_length'].toLocaleString()} characters per post`);
        }
        if (entMap['max_video_minutes'] && entMap['max_file_size_mb']) {
          quotaFeatures.push(
            `${entMap['max_video_minutes']}-min video uploads (up to ${entMap['max_file_size_mb']} MB)`
          );
        }
      }

      return {
        role: key,
        name: dbPlan?.name || defaults.name,
        description: dbPlan?.description || '',
        priceCents,
        formattedPrice,
        icon: defaults.icon,
        features: [...defaults.highlights, ...quotaFeatures],
      };
    });
  };

  const tiers = getTierDisplayList();
  const activeTierObj = tiers.find((t) => t.role === selectedTier) || tiers[0];

  const handleOpenCryptoModal = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    setShowCryptoModal(true);
    setCryptoSearch('');
    if (currencies.length === 0) {
      setLoadingCryptos(true);
      try {
        const res = await apiRequest<{ cryptocurrencies: CryptoCurrencyOption[] }>('/subscriptions/cryptos');
        setCurrencies(res.cryptocurrencies || []);
      } catch (err) {
        console.error('Failed to load cryptocurrencies', err);
      } finally {
        setLoadingCryptos(false);
      }
    }
  };

  const handleSelectCryptoAndCheckout = async (currencyCode: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    setCreatingInvoice(true);
    setLoading(true);
    try {
      const res = await apiRequest<{
        order_id: string;
        invoice_url: string;
        amount_cents: number;
        currency: string;
      }>('/subscriptions/checkout', {
        method: 'POST',
        body: JSON.stringify({
          plan_id: selectedTier,
          billing_cycle: 'monthly',
          currency: currencyCode,
        }),
      });

      setShowCryptoModal(false);
      router.push(`/premium/invoice?order_id=${res.order_id}`);
    } catch (err: any) {
      alert(err.message || 'Failed to initiate cryptocurrency checkout');
    } finally {
      setCreatingInvoice(false);
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

  const previewUserName =
    currentUser?.display_name || currentUser?.username || 'Username';

  return (
    <>
      <main className="w-full max-w-[600px] border-r border-border min-h-screen pb-28">
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
                activeTab === 'subscribe'
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted hover:text-foreground'
              )}
            >
              Role Plans
            </button>
            <button
              onClick={() => {
                setActiveTab('history');
                loadHistory();
              }}
              className={cn(
                'px-3 py-1 rounded-full transition-colors',
                activeTab === 'history'
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted hover:text-foreground'
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
                    <UserNameWithRole
                      displayName={previewUserName}
                      role={currentUser?.role || subscription.plan_id}
                      className="text-[18px]"
                    />
                  </div>
                  <span className="text-[12px] bg-blue-500/20 text-[#1d9bf0] font-bold px-2.5 py-0.5 rounded-full uppercase tracking-wider">
                    Active
                  </span>
                </div>
                <p className="text-[14px] text-muted mb-3">
                  Your membership renews on {new Date(subscription.current_period_end).toLocaleDateString()}.
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
            <div className="text-center space-y-2 py-1">
              <h2 className="text-[26px] sm:text-[30px] font-extrabold tracking-tight">
                Upgrade Your Role
              </h2>
              <p className="text-muted text-[14px] max-w-[460px] mx-auto leading-relaxed">
                Unlock exclusive role badges, styled display names, higher daily post quotas, and enhanced media uploads.
              </p>
            </div>

            {/* Role Tier Cards Grid (2x2 on sm+, 1 col on mobile) */}
            {plansLoading ? (
              <div className="flex justify-center py-12">
                <Loader2 className="w-8 h-8 animate-spin text-[#1d9bf0]" />
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {tiers.map((tier) => (
                  <div
                    key={tier.role}
                    onClick={() => setSelectedTier(tier.role)}
                    className={cn(
                      'bg-[#16181c] rounded-2xl p-5 flex flex-col relative border-2 cursor-pointer transition-all',
                      selectedTier === tier.role
                        ? 'border-[#1d9bf0] shadow-[0_0_20px_rgba(29,155,240,0.3)] scale-[1.01]'
                        : 'border-border/70 hover:border-[#333c45]'
                    )}
                  >
                    <div className="mb-4 border-b border-[#2f3336] pb-4">
                      <div className="flex items-center gap-2 mb-2.5">
                        {tier.icon}
                        <h3 className="text-[19px] font-bold text-white">{tier.name}</h3>
                      </div>

                      <div className="text-xs text-gray-400 mb-3.5 bg-black/50 px-3 py-2 rounded-lg border border-white/5 flex items-center gap-1.5 min-w-0">
                        <span className="shrink-0">Preview:</span>
                        <UserNameWithRole
                          displayName={previewUserName}
                          role={tier.role}
                          className="text-sm"
                        />
                      </div>

                      <div className="flex items-baseline gap-1">
                        <span className="text-3xl font-extrabold text-white">
                          {tier.formattedPrice}
                        </span>
                        <span className="text-[#71767b] text-sm font-medium">/ mo</span>
                      </div>
                    </div>

                    <ul className="flex flex-col gap-2.5 text-[13px] font-medium mt-auto">
                      {tier.features.map((feat, i) => (
                        <li key={i} className="flex items-start gap-2.5">
                          <Check className="text-[#1d9bf0] w-4 h-4 shrink-0 mt-0.5" />
                          <span className="text-gray-300 leading-snug">{feat}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </div>
            )}

            {/* Subscribe Action Bar */}
            {activeTierObj && (
              <div className="bg-[#16181c] border border-[#2f3336] rounded-2xl p-4 sm:p-5 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-xl">
                <div className="flex flex-col text-center sm:text-left">
                  <div className="flex items-center justify-center sm:justify-start gap-2">
                    <span className="font-bold text-base text-white">{activeTierObj.name} Tier</span>
                  </div>
                  <span className="text-[#71767b] text-xs mt-0.5">
                    {activeTierObj.formattedPrice} / month · Instant role activation via Plisio Crypto
                  </span>
                </div>

                <button
                  type="button"
                  onClick={handleOpenCryptoModal}
                  disabled={loading}
                  className="w-full sm:w-auto min-w-[220px] flex items-center justify-center gap-2 bg-white hover:bg-gray-200 text-black font-bold py-3 px-6 rounded-full text-[15px] transition-colors disabled:opacity-70"
                >
                  {loading ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    `Subscribe to ${activeTierObj.name}`
                  )}
                </button>
              </div>
            )}
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
                    onClick={() => router.push(`/premium/invoice?order_id=${order.id}`)}
                    className="border border-border rounded-xl p-4 flex items-center justify-between bg-subtle/20 hover:bg-subtle/40 cursor-pointer transition-colors"
                  >
                    <div>
                      <div className="font-bold uppercase text-[14px] flex items-center gap-2">
                        <span>{order.plan_id} Role Plan</span>
                        <UserNameWithRole
                          displayName={previewUserName}
                          role={order.plan_id}
                          className="text-xs"
                        />
                      </div>
                      <div className="text-[13px] text-muted mt-0.5">
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

      {/* Crypto Selection Modal */}
      {showCryptoModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 animate-in fade-in">
          <div className="bg-[#16181c] border border-[#333] rounded-2xl w-full max-w-md p-6 relative shadow-2xl">
            <button
              onClick={() => !creatingInvoice && setShowCryptoModal(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-white p-1"
              disabled={creatingInvoice}
            >
              <X className="w-5 h-5" />
            </button>

            <h2 className="text-xl font-bold text-white mb-1">Select Crypto</h2>
            <p className="text-gray-400 text-xs mb-4">
              Choose your preferred cryptocurrency to complete the{' '}
              <span className="text-white font-semibold">{activeTierObj?.formattedPrice}</span> payment for the{' '}
              <span className="text-white font-semibold">{activeTierObj?.name}</span> role.
            </p>

            {currencies.length > 0 && (
              <div className="mb-4">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
                  <input
                    type="text"
                    placeholder="Search cryptocurrency..."
                    value={cryptoSearch}
                    onChange={(e) => setCryptoSearch(e.target.value)}
                    className="w-full pl-9 pr-3 py-2.5 bg-[#202327] border border-[#333] rounded-xl text-sm outline-none focus:border-[#1d9bf0] transition-colors text-white"
                  />
                </div>
              </div>
            )}

            {loadingCryptos && !currencies.length ? (
              <div className="flex justify-center py-10">
                <Loader2 className="w-8 h-8 animate-spin text-[#1d9bf0]" />
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-3 max-h-[320px] overflow-y-auto pr-1">
                {currencies
                  .filter(
                    (c) =>
                      c.name.toLowerCase().includes(cryptoSearch.toLowerCase()) ||
                      c.currency.toLowerCase().includes(cryptoSearch.toLowerCase())
                  )
                  .map((c) => {
                    const usdRate = c.price_usd || c.priceUsd;
                    return (
                      <button
                        key={c.cid || c.currency}
                        onClick={() => handleSelectCryptoAndCheckout(c.currency)}
                        disabled={creatingInvoice}
                        className="flex flex-col items-center justify-center gap-1.5 p-3.5 rounded-xl border border-[#333] hover:border-[#1d9bf0] hover:bg-[#1d9bf0]/10 transition-all disabled:opacity-50"
                      >
                        <img
                          src={c.icon || `/crypto-icons/${c.currency}.svg`}
                          alt={c.name}
                          className="w-8 h-8 object-contain"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).src = '/crypto-icons/USDT.svg';
                          }}
                        />
                        <span className="text-xs font-semibold text-white text-center line-clamp-1">
                          {c.name}
                        </span>
                        <span className="text-[11px] text-gray-400 font-mono">{c.currency}</span>
                        {usdRate && (
                          <span className="text-[11px] font-medium text-emerald-400">
                            $
                            {parseFloat(usdRate).toLocaleString('en-US', {
                              minimumFractionDigits: 2,
                              maximumFractionDigits: 6,
                            })}
                          </span>
                        )}
                      </button>
                    );
                  })}
              </div>
            )}

            {creatingInvoice && (
              <div className="absolute inset-0 bg-[#16181c]/85 flex flex-col items-center justify-center rounded-2xl">
                <Loader2 className="w-9 h-9 animate-spin text-[#1d9bf0] mb-3" />
                <span className="font-semibold text-sm text-white">Preparing Plisio invoice...</span>
              </div>
            )}
          </div>
        </div>
      )}

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
        description="Your subscription renewal will be canceled and role benefits will end at the end of the current billing period."
        confirmText="Cancel subscription"
        cancelText="Keep subscription"
        variant="destructive"
        isLoading={isCanceling}
      />
    </>
  );
}
