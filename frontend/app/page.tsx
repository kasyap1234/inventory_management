'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

import { tokenStorage } from '@/lib/security';
import { Loader2 } from 'lucide-react';

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
      <div className="flex items-center justify-center min-h-screen bg-background relative overflow-hidden">
        <div className="text-center relative z-10 p-8">
          <div className="relative mb-8 inline-block">
            <Loader2 className="h-12 w-12 text-primary animate-spin" />
          </div>

          <div className="flex items-center justify-center gap-3 mb-2">
            <h1 className="text-3xl font-bold text-foreground tracking-tight">AgroMart</h1>
          </div>

          <p className="text-muted-foreground text-sm font-medium">Initializing workspace...</p>
        </div>
      </div>
    );
  }

  return null;
}
