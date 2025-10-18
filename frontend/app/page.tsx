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
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-white to-gray-50 relative overflow-hidden">
        <div className="absolute inset-0 animated-gradient opacity-5"></div>
        <div className="absolute inset-0 dot-pattern"></div>
        <div className="text-center relative z-10">
          <div className="relative mb-8">
            <div className="animate-spin rounded-full h-20 w-20 border-4 border-gray-200 border-t-blue-600 mx-auto"></div>
          </div>
          <div className="flex items-center justify-center gap-3 mb-4">
            <h1 className="text-4xl font-bold gradient-text">AgroMart</h1>
          </div>
          <p className="text-gray-600 text-base">Loading your workspace...</p>
          <div className="mt-6 flex items-center justify-center gap-2">
            <div className="w-2 h-2 rounded-full bg-blue-600 animate-pulse"></div>
            <div className="w-2 h-2 rounded-full bg-purple-600 animate-pulse" style={{ animationDelay: '0.2s' }}></div>
            <div className="w-2 h-2 rounded-full bg-pink-600 animate-pulse" style={{ animationDelay: '0.4s' }}></div>
          </div>
        </div>
      </div>
    );
  }

  return null;
}
