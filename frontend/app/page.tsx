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
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50">
        <div className="text-center">
          <div className="relative mb-8">
            <div className="animate-spin rounded-full h-20 w-20 border-4 border-gray-200 border-t-blue-600 mx-auto"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-600 to-purple-600 opacity-20 animate-pulse"></div>
            </div>
          </div>
          <h1 className="text-2xl font-bold gradient-text mb-2">Agromart</h1>
          <p className="text-gray-600">Loading your workspace...</p>
        </div>
      </div>
    );
  }

  return null;
}
