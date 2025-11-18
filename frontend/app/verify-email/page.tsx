'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ShieldCheck, MailCheck, Loader2, MailWarning } from 'lucide-react';

type VerificationState = 'idle' | 'verifying' | 'success' | 'error';

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';
  const email = searchParams.get('email') ?? '';

  const [state, setState] = useState<VerificationState>(token ? 'verifying' : 'idle');
  const [message, setMessage] = useState<string>('');

  useEffect(() => {
    if (!token) {
      return;
    }

    let cancelled = false;
    const verify = async () => {
      try {
        setState('verifying');
        const response = await api.post<{ message: string }>('/auth/verify', { token });
        if (!cancelled) {
          setMessage(response.data?.message ?? 'Email verified successfully.');
          setState('success');
        }
      } catch {
        if (!cancelled) {
          setMessage('The verification link is invalid or has expired.');
          setState('error');
        }
      }
    };

    verify();
    return () => {
      cancelled = true;
    };
  }, [token]);

  const header = useMemo(() => {
    switch (state) {
      case 'verifying':
        return { title: 'Verifying your email...', description: 'Please wait a moment while we confirm your account.' };
      case 'success':
        return { title: 'Your email is verified!', description: 'You can now sign in and start using Agromart.' };
      case 'error':
        return { title: 'We could not verify your email', description: 'The verification link may have expired. Request a new one below.' };
      default:
        return { title: 'Check your inbox', description: 'We sent you a verification link. Click it to activate your account.' };
    }
  }, [state]);

  const icon = useMemo(() => {
    switch (state) {
      case 'verifying':
        return <Loader2 className="h-12 w-12 text-primary animate-spin" />;
      case 'success':
        return <ShieldCheck className="h-12 w-12 text-green-600" />;
      case 'error':
        return <MailWarning className="h-12 w-12 text-destructive" />;
      default:
        return <MailCheck className="h-12 w-12 text-primary" />;
    }
  }, [state]);

  return (
    <div className="flex min-h-screen">
      <div className="flex flex-1 items-center justify-center p-8 bg-background">
        <div className="w-full max-w-xl text-center space-y-6">
          <div className="flex justify-center">{icon}</div>
          <div>
            <h1 className="text-3xl font-bold text-foreground mb-2">{header.title}</h1>
            <p className="text-muted-foreground">{header.description}</p>
            {state === 'idle' && email && (
              <p className="text-sm text-muted-foreground mt-2">
                Verification email sent to <span className="font-semibold">{email}</span>.
              </p>
            )}
          </div>

          <Card className="shadow-lg border-border bg-card">
            <CardContent className="p-8 space-y-4">
              {state === 'success' && (
                <div className="space-y-4">
                  <p className="text-sm text-foreground">{message}</p>
                  <Link href="/login">
                    <Button className="w-full text-primary-foreground">
                      Proceed to sign in
                    </Button>
                  </Link>
                </div>
              )}

              {state === 'error' && (
                <div className="space-y-4">
                  <p className="text-sm text-foreground">{message}</p>
                  <Link href="/forgot-password">
                    <Button variant="outline" className="w-full">
                      Request new verification link
                    </Button>
                  </Link>
                </div>
              )}

              {state === 'idle' && (
                <div className="space-y-3 text-sm text-muted-foreground">
                  <p>Didn&apos;t receive the email?</p>
                  <ul className="text-left list-disc list-inside space-y-1">
                    <li>Check your spam or promotions folder.</li>
                    <li>Make sure you entered the correct email address.</li>
                    <li>Request another link after a few minutes.</li>
                  </ul>
                </div>
              )}

              {state === 'verifying' && (
                <div className="text-sm text-muted-foreground">Please wait while we verify the token...</div>
              )}
            </CardContent>
          </Card>

          <p className="text-sm text-muted-foreground">
            Need assistance?{' '}
            <Link href="/login" className="text-primary hover:text-primary/90 font-semibold">
              Contact support
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 bg-muted/30 border-l border-border items-center justify-center p-12">
        <div className="max-w-md space-y-6 text-foreground">
          <h2 className="text-4xl font-bold leading-tight">Secure access for your team</h2>
          <p className="text-lg text-muted-foreground">
            Email verification keeps your organization safe. Finish this step to explore the full Agromart platform.
          </p>
        </div>
      </div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={<div className="text-center">Loading...</div>}>
      <VerifyEmailContent />
    </Suspense>
  );
}
