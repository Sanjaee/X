'use client';

import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';

interface UpgradeModalContextType {
  isOpen: boolean;
  selectedPlanId: string;
  openUpgradeModal: (initialPlanId?: string) => void;
  closeUpgradeModal: () => void;
}

const UpgradeModalContext = createContext<UpgradeModalContextType | undefined>(undefined);

export function UpgradeModalProvider({ children }: { children: React.ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedPlanId, setSelectedPlanId] = useState('vip');

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      if (params.get('upgrade') === 'true') {
        const plan = params.get('plan') || 'vip';
        setSelectedPlanId(plan);
        setIsOpen(true);
      }
    }
  }, []);

  const openUpgradeModal = useCallback((initialPlanId?: string) => {
    if (initialPlanId) {
      setSelectedPlanId(initialPlanId);
    }
    setIsOpen(true);
  }, []);

  const closeUpgradeModal = useCallback(() => {
    setIsOpen(false);
    if (typeof window !== 'undefined' && window.location.search.includes('upgrade=true')) {
      const url = new URL(window.location.href);
      url.searchParams.delete('upgrade');
      url.searchParams.delete('plan');
      window.history.replaceState({}, '', url.pathname + (url.search ? url.search : ''));
    }
  }, []);

  return (
    <UpgradeModalContext.Provider
      value={{
        isOpen,
        selectedPlanId,
        openUpgradeModal,
        closeUpgradeModal,
      }}
    >
      {children}
    </UpgradeModalContext.Provider>
  );
}

export function useUpgradeModal() {
  const context = useContext(UpgradeModalContext);
  if (!context) {
    throw new Error('useUpgradeModal must be used within an UpgradeModalProvider');
  }
  return context;
}

