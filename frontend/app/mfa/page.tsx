'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Shield, ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import api from '@/lib/api';
import { tokenStorage } from '@/lib/security';

export default function MFAPage() {
  const [code, setCode] = useState('');
  const router = useRouter();

  useEffect(() => {
    // Check if we have a temp token
    const tempToken = localStorage.getItem('temp_token');
    if (!tempToken) {
      router.push('/login');
    }
  }, [router]);

  const verifyMFA = useMutation({
    mutationFn: async (mfaCode: string) => {
      const tempToken = localStorage.getItem('temp_token');
      if (!tempToken) {
        throw new Error('No temporary token found');
      }

      const response = await api.post('/auth/2fa/verify', {
        token: tempToken,
        code: mfaCode,
      });
      return response.data;
    },
    onSuccess: (data) => {
      // Store the JWT tokens
      tokenStorage.setTokens();
      // Clear temp token
      localStorage.removeItem('temp_token');
      // Redirect to dashboard
      router.push('/dashboard');
    },
    onError: () => {
      // Clear temp token on error
      localStorage.removeItem('temp_token');
      router.push('/login');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (code.length === 6) {
      verifyMFA.mutate(code);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-8 bg-gray-50">
      <div className="w-full max-w-md">
        <Card className="shadow-xl">
          <CardHeader className="text-center">
            <div className="mx-auto w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mb-4">
              <Shield className="w-6 h-6 text-blue-600" />
            </div>
            <CardTitle className="text-2xl">Two-Factor Authentication</CardTitle>
            <p className="text-gray-600 mt-2">
              Enter the 6-digit code from your authenticator app
            </p>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="space-y-2">
                <Input
                  type="text"
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="000000"
                  maxLength={6}
                  className="text-center text-2xl tracking-widest font-mono"
                  required
                  autoFocus
                />
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={verifyMFA.isPending || code.length !== 6}
              >
                {verifyMFA.isPending ? 'Verifying...' : 'Verify Code'}
              </Button>
            </form>

            {verifyMFA.isError && (
              <div className="mt-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm text-center">
                Invalid code. Please try again.
              </div>
            )}

            <div className="mt-6 text-center">
              <Link
                href="/login"
                className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900"
                onClick={() => localStorage.removeItem('temp_token')}
              >
                <ArrowLeft className="w-4 h-4 mr-1" />
                Back to login
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
