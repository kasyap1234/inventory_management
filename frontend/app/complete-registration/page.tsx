'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Package, ArrowRight, CheckCircle2 } from 'lucide-react';
import { AxiosError } from 'axios';

function CompleteRegistrationContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { completeGoogleSignup } = useAuth();
    const [token, setToken] = useState<string | null>(null);

    const [formData, setFormData] = useState({
        tenant_name: '',
        subdomain: '',
    });

    useEffect(() => {
        const tokenParam = searchParams.get('token');
        if (!tokenParam) {
            router.push('/login');
            return;
        }
        setToken(tokenParam);
    }, [searchParams, router]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!token) return;

        completeGoogleSignup.mutate({
            token,
            tenant_name: formData.tenant_name,
            subdomain: formData.subdomain,
        });
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData(prev => ({
            ...prev,
            [e.target.name]: e.target.value
        }));
    };

    if (!token) return null;

    return (
        <div className="flex min-h-screen">
            {/* Left side - Form */}
            <div className="relative flex flex-1 items-center justify-center px-6 py-12 lg:px-14 bg-background">
                <div className="w-full max-w-md space-y-8">
                    {/* Logo */}
                    <div className="mb-8 text-center">
                        <div className="inline-flex items-center gap-3 mb-4">
                            <div className="w-14 h-14 rounded-2xl bg-primary flex items-center justify-center shadow-sm">
                                <Package className="w-8 h-8 text-primary-foreground" />
                            </div>
                            <h1 className="text-4xl font-bold text-foreground">AgroMart</h1>
                        </div>
                        <h2 className="text-2xl font-semibold text-foreground">Complete Registration</h2>
                        <p className="text-muted-foreground text-base mt-2">
                            Please provide a few more details to set up your workspace.
                        </p>
                    </div>

                    {/* Form Card */}
                    <Card className="border-border shadow-lg rounded-3xl bg-card">
                        <CardContent className="space-y-6 p-8 sm:p-10">
                            <form onSubmit={handleSubmit} className="space-y-6">
                                {completeGoogleSignup.isError && (
                                    <div className="bg-destructive/10 border border-destructive/20 text-destructive px-4 py-3 rounded-lg text-sm flex items-start gap-2">
                                        <svg className="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                                        </svg>
                                        <span>
                                            {(
                                                (completeGoogleSignup.error as AxiosError<{ error?: { message?: string } }> | undefined)?.response?.data?.error?.message ??
                                                (completeGoogleSignup.error as AxiosError<{ message?: string }> | undefined)?.response?.data?.message ??
                                                'Failed to complete registration. Please try again.'
                                            )}
                                        </span>
                                    </div>
                                )}

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
                                            className="rounded-r-none border-r-0"
                                            required
                                        />
                                        <span className="inline-flex items-center px-4 border border-l-0 border-input bg-muted text-muted-foreground text-sm font-medium rounded-r-md">
                                            .agromart.com
                                        </span>
                                    </div>
                                </div>

                                <Button
                                    type="submit"
                                    className="w-full font-semibold shadow-sm"
                                    disabled={completeGoogleSignup.isPending}
                                >
                                    {completeGoogleSignup.isPending ? (
                                        <span className="flex items-center justify-center gap-2">
                                            <svg className="animate-spin h-5 w-5 text-primary-foreground" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                            </svg>
                                            Setting up workspace...
                                        </span>
                                    ) : (
                                        <span className="flex items-center justify-center gap-2">
                                            Complete Registration
                                            <ArrowRight className="w-5 h-5" />
                                        </span>
                                    )}
                                </Button>
                            </form>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Right side - Info */}
            <div className="hidden lg:flex flex-1 bg-muted/30 border-l border-border items-center justify-center px-16">
                <div className="max-w-lg space-y-10">
                    <div className="space-y-6">
                        <div className="flex items-start gap-4">
                            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-100 text-blue-600">
                                <CheckCircle2 className="w-6 h-6" />
                            </div>
                            <div className="space-y-1">
                                <h3 className="text-base font-semibold text-foreground">Almost there!</h3>
                                <p className="text-sm text-muted-foreground">
                                    Just a few more details to set up your dedicated workspace.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function CompleteRegistrationPage() {
    return (
        <Suspense fallback={
            <div className="flex min-h-screen items-center justify-center bg-background px-4">
                <div className="w-16 h-16 rounded-2xl bg-primary flex items-center justify-center shadow-sm animate-pulse">
                    <Package className="w-8 h-8 text-primary-foreground" />
                </div>
            </div>
        }>
            <CompleteRegistrationContent />
        </Suspense>
    );
}
