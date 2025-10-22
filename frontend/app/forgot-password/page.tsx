'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { CheckCircle2, Mail, Undo2 } from 'lucide-react';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const { requestPasswordReset } = useAuth();

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    requestPasswordReset.mutate({ email: email.trim() });
  };

  const isCompleted = requestPasswordReset.isSuccess;

  return (
    <div className="flex min-h-screen">
      <div className="relative flex flex-1 items-center justify-center px-6 py-12 lg:px-14 auth-gradient">
        <div className="auth-gradient-content w-full max-w-md space-y-8 animate-fade-in">
          <div className="flex items-start gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-100 text-blue-600">
              <Mail className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-3xl font-semibold text-gray-900">Reset your password</h1>
              <p className="text-gray-500">Enter the email associated with your account.</p>
            </div>
          </div>

          <Card className="border border-[var(--border)] shadow-[0_24px_48px_-32px_rgba(15,23,42,0.35)] rounded-3xl bg-white">
            <CardContent className="space-y-6 p-8 sm:p-10">
              {isCompleted ? (
                <div className="space-y-5 text-center">
                  <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-600">
                    <CheckCircle2 className="h-7 w-7" />
                  </div>
                  <div className="space-y-2">
                    <h2 className="text-xl font-semibold text-gray-900">Check your inbox</h2>
                    <p className="text-sm text-gray-600">
                      If an account exists for <span className="font-semibold">{email}</span>, you&apos;ll receive a link to reset your password shortly.
                    </p>
                  </div>
                  <Link href="/login">
                    <Button variant="outline" className="w-full">
                      <Undo2 className="mr-2 h-4 w-4" /> Return to login
                    </Button>
                  </Link>
                </div>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="space-y-2">
                    <label htmlFor="email" className="text-sm font-medium text-gray-700">
                      Email address
                    </label>
                    <Input
                      id="email"
                      type="email"
                      placeholder="you@company.com"
                      value={email}
                      onChange={(event) => setEmail(event.target.value)}
                      required
                    />
                  </div>

                  {requestPasswordReset.isError && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm animate-fade-in">
                      Something went wrong. Please try again in a moment.
                    </div>
                  )}

                  <Button
                    type="submit"
                    className="w-full btn-modern gradient-primary text-white font-semibold shadow-colored hover:shadow-colored"
                    disabled={requestPasswordReset.isPending}
                  >
                    {requestPasswordReset.isPending ? 'Sending reset link...' : 'Send reset link'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>

          <p className="text-center text-sm text-gray-500">
            Remember your password?{' '}
            <Link href="/login" className="text-blue-600 hover:text-blue-700 font-semibold">
              Back to sign in
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 bg-white border-l border-[var(--border)] items-center justify-center px-16">
        <div className="max-w-lg space-y-8 animate-slide-in">
          <div className="inline-flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface-secondary)] px-4 py-1.5 text-gray-700">
            Account security matters
          </div>
          <div className="space-y-4 text-gray-900">
            <h2 className="text-4xl font-semibold leading-tight">We keep your account secure</h2>
            <p className="text-base text-gray-600 leading-relaxed">
              Reset your password safely knowing your account is protected with enterprise-grade security and continuous monitoring.
            </p>
          </div>
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-100 text-blue-600">
                <Mail className="h-5 w-5" />
              </div>
              <div className="space-y-1">
                <h3 className="text-sm font-semibold text-gray-900">Instant notifications</h3>
                <p className="text-sm text-gray-600">Stay informed about important account activity.</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-purple-100 text-purple-600">
                <CheckCircle2 className="h-5 w-5" />
              </div>
              <div className="space-y-1">
                <h3 className="text-sm font-semibold text-gray-900">Strong verification</h3>
                <p className="text-sm text-gray-600">Multi-layer checks ensure only you can access your workspace.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
