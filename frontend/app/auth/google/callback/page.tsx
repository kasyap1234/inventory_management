'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useQueryClient } from '@tanstack/react-query';
import { tokenStorage, csrfTokenManager } from '@/lib/security';
import { Card, CardContent } from '@/components/ui/card';
import { Package } from 'lucide-react';

function CallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const queryClient = useQueryClient();
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const accessToken = searchParams.get('access_token');
        const refreshToken = searchParams.get('refresh_token');
        const errorParam = searchParams.get('error');

        if (errorParam) {
            setError(errorParam);
            return;
        }

        if (accessToken && refreshToken) {
            // Successful login
            tokenStorage.setTokens(accessToken, refreshToken);
            csrfTokenManager.clearToken();
            // Invalidate user query to fetch fresh data
            queryClient.invalidateQueries({ queryKey: ['user'] });
            router.push('/dashboard');
        } else {
            // Check for temp token for completion
            // Actually, the backend redirects to /complete-registration if needed
            // So if we are here, it should be success or error
            setError('Invalid callback parameters');
        }
    }, [searchParams, router, queryClient]);

    return (
        <div className="flex min-h-screen items-center justify-center bg-background px-4">
            <Card className="w-full max-w-md border-border shadow-lg rounded-3xl bg-card">
                <CardContent className="flex flex-col items-center justify-center p-10 space-y-6">
                    <div className="w-16 h-16 rounded-2xl bg-primary flex items-center justify-center shadow-sm animate-pulse">
                        <Package className="w-8 h-8 text-primary-foreground" />
                    </div>

                    <div className="text-center space-y-2">
                        <h1 className="text-2xl font-bold text-foreground">
                            {error ? 'Authentication Failed' : 'Authenticating...'}
                        </h1>
                        <p className="text-muted-foreground">
                            {error ? error : 'Please wait while we complete your sign in.'}
                        </p>
                    </div>

                    {error && (
                        <button
                            onClick={() => router.push('/login')}
                            className="text-primary hover:underline font-medium text-sm"
                        >
                            Back to Login
                        </button>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

export default function GoogleCallbackPage() {
    return (
        <Suspense fallback={
            <div className="flex min-h-screen items-center justify-center bg-background px-4">
                <div className="w-16 h-16 rounded-2xl bg-primary flex items-center justify-center shadow-sm animate-pulse">
                    <Package className="w-8 h-8 text-primary-foreground" />
                </div>
            </div>
        }>
            <CallbackContent />
        </Suspense>
    );
}
