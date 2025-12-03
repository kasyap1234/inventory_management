'use client';

import { useMemo, useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import Link from 'next/link';
import { ArrowRight, TrendingUp, Users, BarChart3, CheckCircle2 } from 'lucide-react';
import { AxiosError } from 'axios';
import { PasswordStrengthMeter } from '@/components/password-strength-meter';
import { GoogleSignInButton } from '@/components/auth/GoogleSignInButton';
import { evaluatePasswordStrength } from '@/lib/password';

export default function SignupPage() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    tenant_name: '',
    subdomain: '',
  });
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const { signup, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  // Redirect if already authenticated - must be in useEffect to avoid setState during render
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.push('/dashboard');
    }
  }, [isLoading, isAuthenticated, router]);

  // Show nothing while redirecting
  if (!isLoading && isAuthenticated) {
    return null;
  }

  const passwordStrength = useMemo(() => evaluatePasswordStrength(formData.password), [formData.password]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!passwordStrength.isAcceptable) {
      setPasswordError('Password is too weak. Follow the suggestions to create a stronger password.');
      return;
    }
    setPasswordError(null);
    signup.mutate(formData);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value
    }));
    if (e.target.name === 'password') {
      setPasswordError(null);
    }
  };

  const features = [
    "Real-time inventory tracking",
    "Multi-warehouse management",
    "Advanced analytics & reports",
    "Automated stock alerts"
  ];

  return (
    <div className="flex min-h-screen">
      {/* Left side - Form */}
      <div className="relative flex flex-1 items-center justify-center px-6 py-12 lg:px-14 bg-background">
        <div className="w-full max-w-lg space-y-8">
          {/* Logo */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">AgroMart</h1>
            <p className="text-muted-foreground">Inventory Management for Modern Businesses</p>
          </div>

          {/* Invite Only Message */}
          <Card className="border-border shadow-lg rounded-3xl bg-card">
            <CardContent className="space-y-6 p-8 sm:p-10 text-center">
              <div className="flex justify-center mb-4">
                <div className="p-3 bg-blue-100 rounded-full text-blue-600">
                  <Users className="w-8 h-8" />
                </div>
              </div>
              <h2 className="text-xl font-semibold">Invite Only Access</h2>
              <p className="text-muted-foreground">
                AgroMart is currently available by invitation only. If you have received an invitation, please follow the link in your email to set up your account.
              </p>
              <p className="text-sm text-muted-foreground">
                Already have an account?{' '}
                <Link href="/login" className="text-primary hover:text-primary/90 font-semibold">
                  Sign in
                </Link>
              </p>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Right side - Features */}
      <div className="hidden lg:flex flex-1 bg-muted/30 border-l border-border items-center justify-center px-16 relative overflow-hidden">
        <div className="max-w-md space-y-8 relative z-10">
          <div className="inline-flex items-center gap-2 rounded-full border border-border bg-background px-4 py-1.5">
            <TrendingUp className="w-5 h-5 text-primary" />
            <span className="text-sm font-semibold text-foreground">Trusted by 500+ businesses</span>
          </div>

          <h2 className="text-4xl font-semibold text-foreground leading-tight">
            Everything you need to manage inventory
          </h2>

          <p className="text-base text-muted-foreground leading-relaxed">
            Join thousands of businesses streamlining operations with AgroMart.
          </p>

          <div className="space-y-4 pt-2">
            {features.map((feature, index) => (
              <div key={index} className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-100 text-blue-600">
                  <CheckCircle2 className="w-5 h-5" />
                </div>
                <span className="text-foreground font-medium">{feature}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
