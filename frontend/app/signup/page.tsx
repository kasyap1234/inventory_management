'use client';

import { useMemo, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import Link from 'next/link';
import { ArrowRight, TrendingUp, Users, BarChart3, CheckCircle2 } from 'lucide-react';
import { AxiosError } from 'axios';
import { PasswordStrengthMeter } from '@/components/password-strength-meter';
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
  const { signup } = useAuth();

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
      <div className="flex flex-1 items-center justify-center p-8 bg-white">
        <div className="w-full max-w-lg animate-fade-in">
          {/* Logo */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">Create your account</h1>
            <p className="text-muted-foreground">Get started with Agromart in less than 2 minutes.</p>
          </div>

          {/* Form Card */}
          <Card className="shadow-xl">
            <CardContent className="p-8">
              <form onSubmit={handleSubmit} className="space-y-5">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label htmlFor="first_name" className="text-sm font-medium text-foreground">
                      First Name
                    </label>
                    <Input
                      id="first_name"
                      name="first_name"
                      placeholder="John"
                      value={formData.first_name}
                      onChange={handleChange}
                      
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <label htmlFor="last_name" className="text-sm font-medium text-foreground">
                      Last Name
                    </label>
                    <Input
                      id="last_name"
                      name="last_name"
                      placeholder="Doe"
                      value={formData.last_name}
                      onChange={handleChange}
                      
                      required
                    />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="email" className="text-sm font-medium text-foreground">
                    Work Email
                  </label>
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    placeholder="you@company.com"
                    value={formData.email}
                    onChange={handleChange}
                    
                    required
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="password" className="text-sm font-medium text-foreground">
                    Password
                  </label>
                  <Input
                    id="password"
                    name="password"
                    type="password"
                    placeholder="Create a strong password"
                    value={formData.password}
                    onChange={handleChange}
                    
                    required
                  />
                  <PasswordStrengthMeter password={formData.password} />
                  <p className="text-xs text-muted-foreground">Use at least 12 characters with numbers, symbols, and mixed case letters.</p>
                  {passwordError && (
                    <p className="text-xs text-red-600">{passwordError}</p>
                  )}
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="tenant_name" className="text-sm font-medium text-foreground">
                    Company Name
                  </label>
                  <Input
                    id="tenant_name"
                    name="tenant_name"
                    placeholder="Acme Inc."
                    value={formData.tenant_name}
                    onChange={handleChange}
                    
                    required
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="subdomain" className="text-sm font-medium text-foreground">
                    Workspace URL
                  </label>
                  <div className="flex items-stretch">
                    <Input
                      id="subdomain"
                      name="subdomain"
                      placeholder="acme"
                      value={formData.subdomain}
                      onChange={handleChange}
                      className="h-12 bg-gray-50 border-gray-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 rounded-r-none border-r-0"
                      required
                    />
                    <span className="inline-flex items-center h-12 px-4 border border-input bg-muted text-muted-foreground text-sm font-medium rounded-r-lg">
                      .agromart.com
                    </span>
                  </div>
                </div>
                
                {signup.isError && (
                  <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm animate-fade-in flex items-start gap-2">
                    <svg className="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                    <span>
                      {(
                        (signup.error as AxiosError<{ error?: { message?: string } }> | undefined)?.response?.data?.error?.message ??
                        (signup.error as AxiosError<{ message?: string }> | undefined)?.response?.data?.message ??
                        'Failed to create account. Please try again.'
                      )}
                    </span>
                  </div>
                )}
                
                <Button
                  type="submit"
                  className="w-full h-12 btn-modern gradient-green text-white text-sm font-semibold shadow-lg hover:shadow-xl"
                  disabled={signup.isPending || !passwordStrength.isAcceptable}
                >
                  {signup.isPending ? (
                    <span className="flex items-center justify-center gap-2">
                      <svg className="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Creating your account...
                    </span>
                  ) : (
                    <span className="flex items-center justify-center gap-2">
                      Create free account
                      <ArrowRight className="w-5 h-5" />
                    </span>
                  )}
                </Button>

                <p className="text-xs text-muted-foreground text-center">
                  By signing up, you agree to our Terms of Service and Privacy Policy
                </p>
              </form>
            </CardContent>
          </Card>

          {/* Sign in link */}
          <p className="text-center text-sm text-muted-foreground mt-6">
            Already have an account?{' '}
            <Link href="/login" className="text-green-700 hover:text-green-800 font-semibold">
              Sign in
            </Link>
          </p>
        </div>
      </div>

      {/* Right side - Features */}
      <div className="hidden lg:flex flex-1 gradient-agro items-center justify-center p-12">
        <div className="max-w-md space-y-8 animate-slide-in">
          <div className="inline-block">
            <div className="flex items-center gap-2 bg-white/80 backdrop-blur-sm px-4 py-2 rounded-full shadow-md">
              <TrendingUp className="w-5 h-5 text-green-600" />
              <span className="text-sm font-semibold text-foreground">Trusted by 500+ businesses</span>
            </div>
          </div>
          
          <h2 className="text-4xl font-bold text-foreground leading-tight">
            Everything you need to manage inventory
          </h2>
          
          <p className="text-lg text-foreground">
            Join thousands of businesses streamlining their operations with Agromart.
          </p>

          <div className="space-y-4 pt-4">
            {features.map((feature, index) => (
              <div key={index} className="flex items-center gap-3">
                <div className="flex-shrink-0 w-6 h-6 bg-green-600 rounded-full flex items-center justify-center">
                  <CheckCircle2 className="w-4 h-4 text-white" />
                </div>
                <span className="text-foreground font-medium">{feature}</span>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-3 gap-6 pt-8">
            <div className="text-center">
              <div className="flex items-center justify-center w-12 h-12 bg-white/90 rounded-xl mx-auto mb-2">
                <Users className="w-6 h-6 text-green-600" />
              </div>
              <div className="text-2xl font-bold text-foreground">500+</div>
              <div className="text-sm text-muted-foreground">Active Users</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center w-12 h-12 bg-white/90 rounded-xl mx-auto mb-2">
                <BarChart3 className="w-6 h-6 text-green-600" />
              </div>
              <div className="text-2xl font-bold text-foreground">98%</div>
              <div className="text-sm text-muted-foreground">Satisfaction</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center w-12 h-12 bg-white/90 rounded-xl mx-auto mb-2">
                <TrendingUp className="w-6 h-6 text-green-600" />
              </div>
              <div className="text-2xl font-bold text-foreground">24/7</div>
              <div className="text-sm text-muted-foreground">Support</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
