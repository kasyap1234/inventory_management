'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Check, CreditCard, Loader2, Sparkles } from 'lucide-react';
import { subscriptionService, type SubscriptionPlanConfig } from '@/lib/services';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';
import { useRazorpayCheckout } from '@/hooks/useRazorpayCheckout';

declare global {
  interface Window {
    Razorpay: any;
  }
}

export default function SubscriptionPlansPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { createSubscription: createRazorpaySubscription, isLoading: razorpayLoading } = useRazorpayCheckout();
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  const { data: plans, isLoading } = useQuery<SubscriptionPlanConfig[]>({
    queryKey: ['subscription-plans'],
    queryFn: async () => {
      const response = await subscriptionService.getAvailablePlans();
      return response;
    },
  });

  const { data: currentUser, isError: userError } = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const response = await api.get('/me');
      return response.data;
    },
    retry: false, // Don't retry on auth failures
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });

  const handleSubscribe = async (planId: string) => {
    if (!currentUser?.email) {
      alert('Please log in to subscribe');
      return;
    }

    setSelectedPlan(planId);
    setIsProcessing(true);

    try {
      // Use Razorpay Checkout integration
      await createRazorpaySubscription(
        planId,
        currentUser.email,
        (subscriptionId) => {
          // Success callback
          queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
          alert('Subscription activated successfully!');
          router.push('/dashboard/subscriptions?success=true');
        },
        (error) => {
          // Error callback
          console.error('Subscription error:', error);
          alert(error?.message || 'Failed to process subscription payment');
          setIsProcessing(false);
          setSelectedPlan(null);
        }
      );
    } catch (error: any) {
      console.error('Subscription error:', error);
      alert(error?.message || 'Failed to create subscription');
      setIsProcessing(false);
      setSelectedPlan(null);
    }
  };

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
    }).format(amount);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-8 max-w-7xl mx-auto py-8">
      {/* Header */}
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold tracking-tight text-foreground">
          Choose Your Plan
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Select the perfect plan for your business needs. All plans include a 14-day free trial.
        </p>
      </div>

      {/* Plans Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-12">
        {plans?.map((plan) => {
          const isPopular = plan.id === 'premium';
          // Use isPopular directly for styling. basic and enterprise use standard styles.
          const isProcessingThis = isProcessing && selectedPlan === plan.id;

          return (
            <Card
              key={plan.id}
              className={`relative flex flex-col h-full border-2 transition-all hover:shadow-lg ${isPopular ? 'border-primary shadow-sm z-10' : 'border-border'
                }`}
            >
              {isPopular && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <Badge variant="default" className="px-4 py-1 flex items-center gap-1 text-sm font-semibold">
                    <Sparkles className="h-3.5 w-3.5" />
                    Most Popular
                  </Badge>
                </div>
              )}

              <CardHeader className="text-center space-y-4 pb-8 pt-8">
                <div className={`w-16 h-16 mx-auto rounded-full flex items-center justify-center ${isPopular
                  ? 'bg-primary/10 text-primary'
                  : 'bg-muted text-muted-foreground'
                  }`}>
                  <CreditCard className="h-8 w-8" />
                </div>
                <CardTitle className="text-2xl font-bold">{plan.name}</CardTitle>
                <CardDescription className="text-base line-clamp-2 md:h-12">
                  {plan.description}
                </CardDescription>
                <div className="pt-4">
                  <div className="text-4xl font-bold text-foreground">
                    {formatCurrency(plan.amount, plan.currency)}
                  </div>
                  <div className="text-muted-foreground mt-1 text-sm font-medium">per {plan.interval}</div>
                </div>
              </CardHeader>

              <CardContent className="space-y-6 flex-grow">
                {/* Features */}
                <div className="space-y-3">
                  {plan.features?.map((feature, index) => (
                    <div key={index} className="flex items-start gap-3">
                      <div className="flex-shrink-0 w-5 h-5 rounded-full bg-primary/10 flex items-center justify-center mt-0.5">
                        <Check className="h-3 w-3 text-primary" />
                      </div>
                      <span className="text-sm text-foreground/80">{feature}</span>
                    </div>
                  ))}
                </div>
              </CardContent>

              <CardFooter className="pt-0 pb-8 px-6">
                <Button
                  onClick={() => handleSubscribe(plan.id)}
                  disabled={isProcessingThis}
                  variant={isPopular ? "default" : "outline"}
                  className="w-full h-11 text-base font-semibold"
                >
                  {isProcessingThis ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    <>
                      <CreditCard className="h-4 w-4 mr-2" />
                      Subscribe Now
                    </>
                  )}
                </Button>
              </CardFooter>

              <div className="pb-4 text-center">
                <p className="text-xs text-muted-foreground">
                  14-day free trial • Cancel anytime
                </p>
              </div>
            </Card>
          );
        })}
      </div>

      {/* Additional Info */}
      <div className="mt-16 text-center space-y-4">
        <h3 className="text-2xl font-bold text-foreground">All plans include:</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mt-8">
          <div className="text-center p-4 rounded-lg bg-card border border-border">
            <div className="text-3xl font-bold text-primary mb-2">24/7</div>
            <div className="text-muted-foreground">Customer Support</div>
          </div>
          <div className="text-center p-4 rounded-lg bg-card border border-border">
            <div className="text-3xl font-bold text-primary mb-2">99.9%</div>
            <div className="text-muted-foreground">Uptime SLA</div>
          </div>
          <div className="text-center p-4 rounded-lg bg-card border border-border">
            <div className="text-3xl font-bold text-primary mb-2">SSL</div>
            <div className="text-muted-foreground">Secure Encryption</div>
          </div>
          <div className="text-center p-4 rounded-lg bg-card border border-border">
            <div className="text-3xl font-bold text-primary mb-2">Free</div>
            <div className="text-muted-foreground">Updates & Backups</div>
          </div>
        </div>
      </div>
    </div>
  );
}
