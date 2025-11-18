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
      <div className="flex flex-1 items-center justify-center p-8 bg-background">
        <div className="w-full max-w-md">
          <div className="mb-8 flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <KeyRound className="h-5 w-5 text-primary" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-foreground">Create a new password</h1>
              <p className="text-muted-foreground">Choose a secure password to protect your account.</p>
            </div>
          </div>

          <Card className="shadow-lg border-border bg-card">
            <CardContent className="p-8">
              {!tokenParam ? (
                <div className="text-center space-y-4">
                  <p className="text-sm text-muted-foreground">
                    This reset link is invalid or expired. Please request a new one.
                  </p>
                  <Link href="/forgot-password">
                    <Button className="w-full">Request new link</Button>
                  </Link>
                </div>
              ) : (
                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="space-y-2">
                    <label htmlFor="password" className="text-sm font-medium text-foreground">
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
                    <div className="text-xs text-muted-foreground space-y-1">
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
                    <label htmlFor="confirmPassword" className="text-sm font-medium text-foreground">
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
                    <div className="bg-destructive/10 border border-destructive/20 text-destructive px-4 py-3 rounded-lg text-sm">
                      {error ?? 'Unable to reset password. Please try again.'}
                    </div>
                  )}

                  <Button
                    type="submit"
                    className="w-full font-semibold shadow-sm"
                    disabled={disableForm || !passwordStrength.isAcceptable}
                  >
                    {resetPassword.isPending ? 'Updating password...' : 'Update password'}
                  </Button>
                </form>
              )}
            </CardContent>
          </Card>

          <p className="text-center text-sm text-muted-foreground mt-6">
            Remembered it?{' '}
            <Link href="/login" className="text-primary hover:text-primary/90 font-semibold">
              Go back to sign in
            </Link>
          </p>
        </div>
      </div>

      <div className="hidden lg:flex flex-1 bg-muted/30 border-l border-border items-center justify-center p-12">
        <div className="max-w-md space-y-6 text-foreground">
          <h2 className="text-4xl font-bold leading-tight">Security is our priority</h2>
          <p className="text-lg text-muted-foreground">
            Strong passwords keep your organization safe. Make sure to use a unique password for Agromart.
          </p>
          <div className="flex items-center gap-3 bg-background border border-border rounded-xl px-4 py-3 shadow-sm">
            <ShieldCheck className="h-6 w-6 text-primary" />
            <span className="text-sm text-foreground">Tip: Use a mix of letters, numbers, and symbols for better protection.</span>
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
