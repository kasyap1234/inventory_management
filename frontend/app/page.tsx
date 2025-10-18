'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

import { tokenStorage } from '@/lib/security';

export default function Home() {
  const router = useRouter();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      if (tokenStorage.hasAccessToken()) {
        router.push('/dashboard');
      } else {
        router.push('/login');
      }
      setIsChecking(false);
    }
  }, [router]);

  if (isChecking) {
    return (
      <div className="flex items-center justify-center min-h-screen gradient-natural relative overflow-hidden">
        <div className="absolute inset-0 animated-gradient opacity-30"></div>
        <div className="absolute inset-0 agro-pattern"></div>
        <div className="text-center relative z-10">
          <div className="relative mb-8">
            <div className="animate-spin rounded-full h-24 w-24 border-4 border-green-200 border-t-green-700 mx-auto"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-14 w-14 rounded-full gradient-agro opacity-30 animate-pulse pulse-glow"></div>
            </div>
          </div>
          <div className="flex items-center justify-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-xl gradient-agro flex items-center justify-center shadow-growth">
              <span className="text-2xl">🌾</span>
            </div>
            <h1 className="text-4xl font-bold gradient-text">AgroMart</h1>
          </div>
          <p className="text-gray-700 font-semibold text-lg">Agrotech Solutions Platform</p>
          <p className="text-gray-600 text-sm mt-1 mb-6">Pesticides • Chemicals • Fertilizers</p>
          <div className="mt-4 flex items-center justify-center gap-2">
            <div className="w-2 h-2 rounded-full bg-green-700 animate-pulse"></div>
            <div className="w-2 h-2 rounded-full bg-green-600 animate-pulse" style={{ animationDelay: '0.2s' }}></div>
            <div className="w-2 h-2 rounded-full bg-amber-600 animate-pulse" style={{ animationDelay: '0.4s' }}></div>
          </div>
        </div>
      </div>
    );
  }

  return null;
}
