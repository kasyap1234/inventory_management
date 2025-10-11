'use client';

import { FormEvent, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ShieldAlert } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { mfaChallengeStore } from '@/lib/security';

export default function MFAPage() {
  const router = useRouter();
  const [challengeToken, setChallengeToken] = useState<string | null>(null);
  const [code, setCode] = useState('');
  const [feedback, setFeedback] = useState<string | null>(null);

  useEffect(() => {
    const challenge = mfaChallengeStore.get();
    if (!challenge) {
      router.replace('/login');
      return;
    }
    setChallengeToken(challenge);
  }, [router]);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setFeedback(
      'Multi-factor authentication will be available soon. Please contact your administrator if you believe this is an error.'
    );
  };

  const handleCancel = () => {
    mfaChallengeStore.clear();
    router.replace('/login');
  };

  if (!challengeToken) {
    return null;
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-6">
      <Card className="w-full max-w-md shadow-xl">
        <CardContent className="p-8 space-y-6">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
              <ShieldAlert className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-foreground">Verify your identity</h1>
              <p className="text-sm text-muted-foreground">Enter the 6-digit code from your authenticator app.</p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="space-y-2">
              <label htmlFor="code" className="text-sm font-medium text-foreground">
                Authentication code
              </label>
              <Input
                id="code"
                name="code"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="123456"
                value={code}
                onChange={(event) => setCode(event.target.value)}
                maxLength={6}
                required
              />
              <p className="text-xs text-muted-foreground">We&apos;ll prompt you for MFA whenever your account requires additional verification.</p>
            </div>

            {feedback && <p className="text-xs text-red-600">{feedback}</p>}

            <div className="flex items-center gap-3">
              <Button type="submit" className="flex-1">
                Verify code
              </Button>
              <Button type="button" variant="secondary" onClick={handleCancel}>
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
