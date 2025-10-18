'use client';

import { Suspense, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import Link from 'next/link';
import { ArrowRight, Sparkles, Shield, Zap } from 'lucide-react';
import { AxiosError } from 'axios';

function LoginContent() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const searchParams = useSearchParams();
  const router = useRouter();
  const { login } = useAuth();

  const resetSuccessful = searchParams.get('reset') === 'success';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    login.mutate({ email, password }, {
      onSuccess: (data) => {
        if (data?.requires_2fa && data.temp_token) {
          // Store temp token and redirect to MFA page
          localStorage.setItem('temp_token', data.temp_token);
          router.push('/mfa');
        }
      }
    });
  };

  return (
    <div className="flex min-h-screen">
      {/* Left side - Form */}
      <div className="flex flex-1 items-center justify-center p-8 bg-gradient-to-br from-white via-green-50/30 to-amber-50/20 agro-pattern">
        <div className="w-full max-w-md animate-fade-in">
          {/* Logo */}
          <div className="mb-8 text-center">
            <div className="inline-flex items-center gap-3 mb-4">
              <div className="w-14 h-14 rounded-2xl gradient-agro flex items-center justify-center shadow-growth">
                <span className="text-3xl">🌾</span>
              </div>
              <h1 className="text-4xl font-bold gradient-text">AgroMart</h1>
            </div>
            <p className="text-gray-700 text-lg font-medium">Agrotech Solutions Platform</p>
            <p className="text-gray-600 text-sm mt-1">Pesticides • Chemicals • Fertilizers</p>
          </div>

          {/* Form Card */}
          <Card className="shadow-elegant-lg border-0">
            <CardContent className="p-8">
              <form onSubmit={handleSubmit} className="space-y-6">
                {resetSuccessful && (
                  <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm animate-fade-in">
                    Password updated successfully. Please sign in with your new password.
                  </div>
                )}
                {login.isError && (
                  <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm animate-fade-in flex items-start gap-2">
                    <svg className="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                    <span>
                      {(
                        (login.error as AxiosError<{ error?: { message?: string } }> | undefined)?.response?.data?.error?.message ??
                        (login.error as AxiosError<{ message?: string }> | undefined)?.response?.data?.message ??
                        'Invalid credentials. Please try again.'
                      )}
                    </span>
                  </div>
                )}
                <div className="space-y-2">
                  <label htmlFor="email" className="text-sm font-medium text-gray-700">
                    Email address
                  </label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="you@company.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <label htmlFor="password" className="text-sm font-medium text-gray-700">
                      Password
                    </label>
                    <Link href="/forgot-password" className="text-sm text-green-700 hover:text-green-800 font-medium">
                      Forgot?
                    </Link>
                  </div>
                  <Input
                    id="password"
                    type="password"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                  />
                </div>
                <Button
                  type="submit"
                  className="w-full h-12 btn-modern gradient-agro text-white text-sm font-semibold shadow-growth hover:shadow-growth hover:scale-[1.02] transition-all duration-300"
                  disabled={login.isPending}
                >
                  {login.isPending ? (
                    <span className="flex items-center justify-center gap-2">
                      <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Signing in...
                    </span>
                  ) : (
                    <span className="flex items-center justify-center gap-2">
                      Sign in to your account
                      <ArrowRight className="w-5 h-5" />
                    </span>
                  )}
                </Button>
              </form>
            </CardContent>
          </Card>

          {/* Sign up link */}
          <p className="text-center text-sm text-gray-600 mt-6">
            Don&apos;t have an account?{' '}
            <Link href="/signup" className="text-green-700 hover:text-green-800 font-semibold">
              Create a free account
            </Link>
          </p>
        </div>
      </div>

      {/* Right side - Brand */}
      <div className="hidden lg:flex flex-1 gradient-natural items-center justify-center p-12 relative overflow-hidden">
        <div className="absolute inset-0 animated-gradient opacity-20"></div>
        <div className="absolute inset-0 agro-pattern"></div>
        <div className="max-w-md space-y-8 animate-slide-in relative z-10">
          <div className="inline-block">
            <div className="flex items-center gap-2 glass px-4 py-2 rounded-full shadow-elegant">
              <span className="text-xl">🌱</span>
              <span className="text-sm font-semibold gradient-text-green">Agricultural Excellence</span>
            </div>
          </div>
          
          <h2 className="text-4xl font-bold text-gray-900 leading-tight">
            Complete Agrotech Inventory Management
          </h2>
          
          <p className="text-lg text-gray-700">
            Professional-grade system for managing pesticides, chemicals, and fertilizers. Track batches, ensure compliance, and optimize your agricultural supply chain.
          </p>

          <div className="space-y-4 pt-4">
            <div className="flex items-start gap-4 group">
              <div className="flex-shrink-0 w-12 h-12 gradient-growth rounded-xl flex items-center justify-center shadow-growth group-hover:scale-110 transition-transform duration-300">
                <Shield className="w-6 h-6 text-white" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 mb-1 text-lg">Regulatory Compliance</h3>
                <p className="text-sm text-gray-600">Meet safety standards for chemical & pesticide handling</p>
              </div>
            </div>
            <div className="flex items-start gap-4 group">
              <div className="flex-shrink-0 w-12 h-12 gradient-harvest rounded-xl flex items-center justify-center shadow-harvest group-hover:scale-110 transition-transform duration-300">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 mb-1 text-lg">Batch Tracking</h3>
                <p className="text-sm text-gray-600">Complete traceability for chemicals and fertilizers</p>
              </div>
            </div>
            <div className="flex items-start gap-4 group">
              <div className="flex-shrink-0 w-12 h-12 gradient-emerald rounded-xl flex items-center justify-center shadow-growth group-hover:scale-110 transition-transform duration-300">
                <Sparkles className="w-6 h-6 text-white" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 mb-1 text-lg">Safety Management</h3>
                <p className="text-sm text-gray-600">Hazard tracking and safety protocol enforcement</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="text-center">Loading...</div>}>
      <LoginContent />
    </Suspense>
  );
}
