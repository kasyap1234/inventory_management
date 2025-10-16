'use client';

import { Suspense, useMemo, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { KeyRound, ShieldCheck } from 'lucide-react';
import { PasswordStrengthMeter } from '@/components/password-strength-meter';
import { evaluatePasswordStrength } from '@/lib/password';

function ResetPasswordContent() {
  const searchParams = useSearchParams();
  const tokenParam = searchParams.get('token') ?? '';

  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  const { resetPassword } = useAuth();

  const passwordStrength = useMemo(() => evaluatePasswordStrength(password), [password]);

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);

    if (!tokenParam) {
      setError('Reset token is missing. Please request a new password reset link.');
      return;
    }

    if (!passwordStrength.isAcceptable) {
      setError('Password is too weak. Follow the suggested improvements.');
      return;
    }

    if (password.trim() !== confirmPassword.trim()) {
      setError('Passwords do not match.');
      return;
    }

    resetPassword.mutate({
      token: tokenParam,
      password: password.trim(),
      confirm_password: confirmPassword.trim(),
    });
  };

  const disableForm = resetPassword.isPending || !tokenParam;

  return (
    <div className="flex min-h-screen">
      <div className="flex flex-1 items-center justify-center p-8 bg-white">
        <div className="w-full max-w-md animate-fade-in">
          <div className="mb-8 flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center">
              <KeyRound className="h-5 w-5 text-purple-600" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Create a new password</h1>
              <p className="text-gray-600">Choose a secure password to protect your account.</p>
            </div>
          </div>

          <Card className="shadow-xl">
            <CardContent className="p-8">
              {!tokenParam ? (
                <div className="text-center space-y-4">
                  <p className="text-sm text-gray-600">
                    This reset link is invalid or expired. Please request a new one.
                  </p>
                  <Link href="/forgot-password">
                    <Button className="w-full">Request new link</Button>
                  </Link>
                </div>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="space-y-2">
                    <label htmlFor="password" className="text-sm font-medium text-gray-700">
                      New password
                    </label>
                    <Input
                      id="password"
                      type="password"
                      placeholder="Enter a strong password"
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                      required
                    />
                    <PasswordStrengthMeter password={password} />
                    <div className="text-xs text-gray-500 space-y-1">
                      <p>Password must meet the following requirements:</p>
                      <ul className="list-disc list-inside ml-2 space-y-0.5">
                        <li>At least 12 characters long</li>
                        <li>At least one uppercase letter</li>
                        <li>At least one lowercase letter</li>
                        <li>At least one number</li>
                        <li>At least one special character (!@#$%^&*)</li>
                      </ul>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label htmlFor="confirmPassword" className="text-sm font-medium text-gray-700">
                      Confirm password
                    </label>
                    <Input
                      id="confirmPassword"
                      type="password"
                      placeholder="Re-enter your password"
                      value={confirmPassword}
                      onChange={(event) => setConfirmPassword(event.target.value)}
                      required
                    />
                  </div>

                  {(error || resetPassword.isError) && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm animate-fade-in">
                      {error ?? 'Unable to reset password. Please try again.'}
                    </div>
                  )}

                  <Button
                    type="submit"
                    className="w-full h-12 btn-modern gradient-blue text-white text-sm font-semibold shadow-lg hover:shadow-xl"
                    disabled={disableForm || !passwordStrength.isAcceptable}
                  >
                    {resetPassword.isPending ? 'Updating password...' : 'Update password'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>

          <p className="text-center text-sm text-gray-600 mt-6">
            Remembered it?{' '}
            <Link href="/login" className="text-blue-600 hover:text-blue-700 font-semibold">
              Go back to sign in
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 gradient-bg items-center justify-center p-12">
        <div className="max-w-md space-y-6 animate-slide-in text-white">
          <h2 className="text-4xl font-bold leading-tight">Security is our priority</h2>
          <p className="text-lg text-white/90">
            Strong passwords keep your organization safe. Make sure to use a unique password for Agromart.
          </p>
          <div className="flex items-center gap-3 bg-white/10 rounded-xl px-4 py-3">
            <ShieldCheck className="h-6 w-6 text-white" />
            <span className="text-sm text-white/90">Tip: Use a mix of letters, numbers, and symbols for better protection.</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<div className="text-center">Loading...</div>}>
      <ResetPasswordContent />
    </Suspense>
  );
}
