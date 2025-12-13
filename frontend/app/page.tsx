'use client';

import { Loader2 } from 'lucide-react';

// Middleware handles the redirection logic (auth_token check).
// This page is only shown briefly while middleware resolves or if JavaScript takes over.
export default function Home() {
  return (
    <div className="flex items-center justify-center min-h-screen bg-background relative overflow-hidden">
      <div className="text-center relative z-10 p-8">
        <div className="relative mb-8 inline-block">
          <Loader2 className="h-12 w-12 text-primary animate-spin" />
        </div>

        <div className="flex items-center justify-center gap-3 mb-2">
          <h1 className="text-3xl font-bold text-foreground tracking-tight">AgroMart</h1>
        </div>

        <p className="text-muted-foreground text-sm font-medium">Loading...</p>
      </div>
    </div>
  );
}
