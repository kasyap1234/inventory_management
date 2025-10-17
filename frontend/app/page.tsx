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
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-indigo-100 via-purple-100 to-pink-100 relative overflow-hidden">
        <div className="absolute inset-0 animated-gradient opacity-40"></div>
        <div className="text-center relative z-10">
          <div className="relative mb-8">
            <div className="animate-spin rounded-full h-24 w-24 border-4 border-gray-200 border-t-indigo-600 mx-auto"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-14 w-14 rounded-full bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-600 opacity-30 animate-pulse pulse-glow"></div>
            </div>
          </div>
          <h1 className="text-4xl font-bold gradient-text mb-3">Agromart</h1>
          <p className="text-gray-700 font-medium text-lg">Loading your workspace...</p>
          <div className="mt-4 flex items-center justify-center gap-2">
            <div className="w-2 h-2 rounded-full bg-indigo-600 animate-pulse"></div>
            <div className="w-2 h-2 rounded-full bg-purple-600 animate-pulse" style={{ animationDelay: '0.2s' }}></div>
            <div className="w-2 h-2 rounded-full bg-pink-600 animate-pulse" style={{ animationDelay: '0.4s' }}></div>
          </div>
        </div>
      </div>
    );
  }

  return null;
}
