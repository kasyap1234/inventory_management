'use client';

import { Suspense, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import Link from 'next/link';
import { ArrowRight, Sparkles, Shield, Zap, Package } from 'lucide-react';
import { AxiosError } from 'axios';

function LoginContent() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const searchParams = useSearchParams();
  const router = useRouter();
  const { login, isAuthenticated, isLoading } = useAuth();

  const resetSuccessful = searchParams.get('reset') === 'success';

  // Redirect if already authenticated
  if (!isLoading && isAuthenticated) {
    router.push('/dashboard');
    return null;
  }

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
      <div className="relative flex flex-1 items-center justify-center px-6 py-12 lg:px-14 auth-gradient">
        <div className="auth-gradient-content w-full max-w-md space-y-8 animate-fade-in">
          {/* Logo */}
          <div className="mb-8 text-center">
            <div className="inline-flex items-center gap-3 mb-4">
              <div className="w-14 h-14 rounded-2xl gradient-primary flex items-center justify-center shadow-colored">
                <Package className="w-8 h-8 text-white" />
              </div>
              <h1 className="text-4xl font-bold text-gray-900">AgroMart</h1>
            </div>
            <p className="text-gray-500 text-base font-medium">Inventory Management Platform</p>
          </div>

          {/* Form Card */}
          <Card className="border border-[var(--border)] shadow-[0_24px_48px_-32px_rgba(15,23,42,0.35)] rounded-3xl bg-white">
            <CardContent className="space-y-6 p-8 sm:p-10">
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
                    <Link href="/forgot-password" className="text-sm text-blue-600 hover:text-blue-700 font-medium">
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
                  className="w-full btn-modern gradient-primary text-white font-semibold shadow-colored hover:shadow-colored hover:scale-[1.02] transition-all duration-300"
                  isLoading={login.isPending}
                  loadingText="Signing in..."
                >
                  <span className="flex items-center justify-center gap-2">
                    Sign in to your account
                    <ArrowRight className="w-5 h-5" />
                  </span>
                </Button>
              </form>
            </CardContent>
          </Card>

          {/* Sign up link */}
          <p className="text-center text-sm text-gray-500">
            Don&apos;t have an account?{' '}
            <Link href="/signup" className="text-blue-600 hover:text-blue-700 font-semibold">
              Create a free account
            </Link>
          </p>
        </div>
      </div>

      {/* Right side - Brand */}
      <div className="hidden lg:flex flex-1 bg-white border-l border-[var(--border)] items-center justify-center px-16">
        <div className="max-w-lg space-y-10 animate-slide-in">
          <div className="inline-flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface-secondary)] px-4 py-1.5">
            <Sparkles className="w-4 h-4 text-blue-600" />
            <span className="text-sm font-semibold text-gray-700">Modern Platform</span>
          </div>

          <h2 className="text-4xl font-semibold text-gray-900 leading-tight tracking-tight">
            Complete Inventory Management Solution
          </h2>

          <p className="text-base text-gray-600 leading-relaxed">
            Professional-grade system for managing your inventory. Track products, streamline operations, and optimize your supply chain.
          </p>

          <div className="space-y-6">
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-100 text-blue-600">
                <Shield className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <h3 className="text-base font-semibold text-gray-900">Secure & Reliable</h3>
                <p className="text-sm text-gray-600">Enterprise-grade security for your data</p>
              </div>
            </div>
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-purple-100 text-purple-600">
                <Zap className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <h3 className="text-base font-semibold text-gray-900">Real-time Tracking</h3>
                <p className="text-sm text-gray-600">Monitor inventory levels in real-time</p>
              </div>
            </div>
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-600">
                <Sparkles className="w-6 h-6" />
              </div>
              <div className="space-y-1">
                <h3 className="text-base font-semibold text-gray-900">Analytics & Insights</h3>
                <p className="text-sm text-gray-600">Data-driven decisions for your business</p>
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
