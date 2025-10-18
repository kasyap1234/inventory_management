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
      <div className="flex flex-1 items-center justify-center p-8 bg-white">
        <div className="w-full max-w-md animate-fade-in">
          <div className="mb-8 flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
              <Mail className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Reset your password</h1>
              <p className="text-gray-600">Enter the email associated with your account.</p>
            </div>
          </div>

          <Card className="shadow-xl">
            <CardContent className="p-8">
              {isCompleted ? (
                <div className="flex flex-col items-center gap-4 text-center">
                  <CheckCircle2 className="h-12 w-12 text-green-500" />
                  <div>
                    <h2 className="text-xl font-semibold text-gray-900">Check your inbox</h2>
                    <p className="text-sm text-gray-600 mt-2">
                      If an account exists for <span className="font-semibold">{email}</span>, you&apos;ll receive a link to reset your password shortly.
                    </p>
                  </div>
                  <Link href="/login">
                    <Button variant="outline" className="mt-4 w-full">
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
                    className="w-full h-12 btn-modern gradient-blue text-white text-sm font-semibold shadow-lg hover:shadow-xl"
                    disabled={requestPasswordReset.isPending}
                  >
                    {requestPasswordReset.isPending ? 'Sending reset link...' : 'Send reset link'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>

          <p className="text-center text-sm text-gray-600 mt-6">
            Remember your password?{' '}
            <Link href="/login" className="text-blue-600 hover:text-blue-700 font-semibold">
              Back to sign in
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 bg-gradient-to-br from-gray-50 to-blue-50/30 items-center justify-center p-12">
        <div className="max-w-md space-y-6 animate-slide-in">
          <h2 className="text-4xl font-bold leading-tight text-gray-900">We keep your account secure</h2>
          <p className="text-lg text-gray-600">
            Reset your password safely knowing your account is protected with enterprise-grade security.
          </p>
        </div>
      </div>
    </div>
  );
}
