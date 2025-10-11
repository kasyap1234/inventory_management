'use client';

import { useEffect, useMemo, useState, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import api from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { ShieldCheck, MailCheck, Loader2, MailWarning } from 'lucide-react';

type VerificationState = 'idle' | 'verifying' | 'success' | 'error' | 'resending';

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token') ?? '';
  const email = searchParams.get('email') ?? '';

  const [state, setState] = useState<VerificationState>(token ? 'verifying' : 'idle');
  const [message, setMessage] = useState<string>('');
  const [resendEmail, setResendEmail] = useState<string>(email);
  const [resendMessage, setResendMessage] = useState<string>('');
  const [isResending, setIsResending] = useState<boolean>(false);

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
      } catch (error) {
        if (!cancelled) {
          // Log error for debugging
          if (process.env.NODE_ENV === 'development') {
            console.error('Email verification failed:', error);
          }
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

  const handleResendEmail = async () => {
    if (!resendEmail || !resendEmail.includes('@')) {
      setResendMessage('Please enter a valid email address.');
      return;
    }

    setIsResending(true);
    setResendMessage('');

    try {
      const response = await api.post<{ message: string }>('/auth/verify/resend', { email: resendEmail });
      setResendMessage(response.data?.message ?? 'Verification email has been sent.');
    } catch (error) {
      // Extract error message safely
      let errorMessage = 'Failed to resend verification email. Please try again.';
      if (error && typeof error === 'object' && 'response' in error) {
        const axiosError = error as { response?: { data?: { message?: string } } };
        errorMessage = axiosError.response?.data?.message || errorMessage;
      }
      setResendMessage(errorMessage);
    } finally {
      setIsResending(false);
    }
  };

  const header = useMemo(() => {
    switch (state) {
      case 'verifying':
        return { title: 'Verifying your email...', description: 'Please wait a moment while we confirm your account.' };
      case 'success':
        return { title: 'Your email is verified!', description: 'You can now sign in and start using Agromart.' };
      case 'error':
        return { title: 'We could not verify your email', description: 'The verification link may have expired. Request a new one below.' };
      case 'resending':
        return { title: 'Sending verification email...', description: 'Please wait while we send your verification link.' };
      default:
        return { title: 'Check your inbox', description: 'We sent you a verification link. Click it to activate your account.' };
    }
  }, [state]);

  const icon = useMemo(() => {
    switch (state) {
      case 'verifying':
        return <Loader2 className="h-12 w-12 text-green-600 animate-spin" />;
      case 'success':
        return <ShieldCheck className="h-12 w-12 text-green-600" />;
      case 'error':
        return <MailWarning className="h-12 w-12 text-red-500" />;
      case 'resending':
        return <Loader2 className="h-12 w-12 text-green-600 animate-spin" />;
      default:
        return <MailCheck className="h-12 w-12 text-green-600" />;
    }
  }, [state]);

  return (
    <div className="flex min-h-screen">
      <div className="flex flex-1 items-center justify-center p-8 bg-white">
        <div className="w-full max-w-xl animate-fade-in text-center space-y-6">
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

          <Card className="shadow-xl">
            <CardContent className="p-8 space-y-4">
              {state === 'success' && (
                <div className="space-y-4">
                  <p className="text-sm text-foreground">{message}</p>
                  <Link href="/login" className="block">
                    <Button className="w-full h-11 btn-modern gradient-green text-white">
                      Proceed to sign in
                    </Button>
                  </Link>
                </div>
              )}

              {state === 'error' && (
                <div className="space-y-4">
                  <p className="text-sm text-foreground">{message}</p>
                  <div className="space-y-2">
                    <input
                      type="email"
                      placeholder="Enter your email"
                      value={resendEmail}
                      onChange={(e) => setResendEmail(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                    <Button 
                      onClick={handleResendEmail}
                      disabled={isResending}
                      className="w-full h-11 btn-modern gradient-green text-white"
                    >
                      {isResending ? (
                        <>
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                          Sending...
                        </>
                      ) : (
                        'Resend Verification Email'
                      )}
                    </Button>
                    {resendMessage && (
                      <p className={`text-sm ${resendMessage.includes('sent') ? 'text-green-600' : 'text-red-600'}`}>
                        {resendMessage}
                      </p>
                    )}
                  </div>
                </div>
              )}

              {state === 'idle' && (
                <div className="space-y-4">
                  <div className="space-y-3 text-sm text-muted-foreground">
                    <p>Didn&apos;t receive the email?</p>
                    <ul className="text-left list-disc list-inside space-y-1">
                      <li>Check your spam or promotions folder.</li>
                      <li>Make sure you entered the correct email address.</li>
                      <li>Request another link below.</li>
                    </ul>
                  </div>
                  <div className="space-y-2 pt-4 border-t">
                    <p className="text-sm font-medium text-foreground">Resend verification email:</p>
                    <input
                      type="email"
                      placeholder="Enter your email"
                      value={resendEmail}
                      onChange={(e) => setResendEmail(e.target.value)}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                    <Button 
                      onClick={handleResendEmail}
                      disabled={isResending}
                      className="w-full h-11 btn-modern gradient-green text-white"
                    >
                      {isResending ? (
                        <>
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                          Sending...
                        </>
                      ) : (
                        'Resend Verification Email'
                      )}
                    </Button>
                    {resendMessage && (
                      <p className={`text-sm ${resendMessage.includes('sent') ? 'text-green-600' : 'text-red-600'}`}>
                        {resendMessage}
                      </p>
                    )}
                  </div>
                </div>
              )}

              {state === 'verifying' && (
                <div className="text-sm text-muted-foreground">Please wait while we verify the token...</div>
              )}
            </CardContent>
          </Card>

          <p className="text-sm text-muted-foreground">
            Need assistance?{' '}
            <Link href="/login" className="text-green-700 hover:text-green-800 font-semibold">
              Contact support
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 gradient-agro items-center justify-center p-12">
        <div className="max-w-md space-y-6 animate-slide-in text-white">
          <h2 className="text-4xl font-bold leading-tight">Secure access for your inventory</h2>
          <p className="text-lg text-white/90">
            Email verification keeps your chemical inventory management safe. Finish this step to explore the full Agromart platform.
          </p>
        </div>
      </div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    }>
      <VerifyEmailContent />
    </Suspense>
  );
}
