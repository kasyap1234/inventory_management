'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
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

  const { data: currentUser } = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const response = await api.get('/me');
      return response.data;
    },
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

  const getPlanColor = (planId: string) => {
    switch (planId) {
      case 'basic':
        return 'from-blue-500 to-blue-600';
      case 'premium':
        return 'from-purple-500 to-purple-600';
      case 'enterprise':
        return 'from-orange-500 to-orange-600';
      default:
        return 'from-gray-500 to-gray-600';
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
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
          Choose Your Plan
        </h1>
        <p className="text-xl text-gray-600 max-w-2xl mx-auto">
          Select the perfect plan for your business needs. All plans include a 14-day free trial.
        </p>
      </div>

      {/* Plans Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-12">
        {plans?.map((plan) => {
          const isPopular = plan.id === 'premium';
          const isProcessingThis = isProcessing && selectedPlan === plan.id;

          return (
            <Card
              key={plan.id}
              className={`relative border-2 transition-all hover:shadow-xl ${
                isPopular ? 'border-purple-500 shadow-lg scale-105' : 'border-gray-200'
              }`}
            >
              {isPopular && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <span className="bg-gradient-to-r from-purple-500 to-purple-600 text-white px-4 py-1 rounded-full text-sm font-semibold flex items-center gap-1">
                    <Sparkles className="h-4 w-4" />
                    Most Popular
                  </span>
                </div>
              )}

              <CardHeader className="text-center space-y-4 pb-8">
                <div className={`w-16 h-16 mx-auto rounded-full bg-gradient-to-r ${getPlanColor(plan.id)} flex items-center justify-center`}>
                  <CreditCard className="h-8 w-8 text-white" />
                </div>
                <CardTitle className="text-2xl font-bold">{plan.name}</CardTitle>
                <CardDescription className="text-base">{plan.description}</CardDescription>
                <div className="pt-4">
                  <div className="text-4xl font-bold text-gray-900">
                    {formatCurrency(plan.amount, plan.currency)}
                  </div>
                  <div className="text-gray-600 mt-1">per {plan.interval}</div>
                </div>
              </CardHeader>

              <CardContent className="space-y-6">
                {/* Features */}
                <div className="space-y-3">
                  {plan.features?.map((feature, index) => (
                    <div key={index} className="flex items-start gap-3">
                      <div className="flex-shrink-0 w-5 h-5 rounded-full bg-green-100 flex items-center justify-center mt-0.5">
                        <Check className="h-3 w-3 text-green-600" />
                      </div>
                      <span className="text-sm text-gray-700">{feature}</span>
                    </div>
                  ))}
                </div>

                {/* Subscribe Button */}
                <Button
                  onClick={() => handleSubscribe(plan.id)}
                  disabled={isProcessingThis}
                  className={`w-full h-12 text-base font-semibold ${
                    isPopular
                      ? 'bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700'
                      : 'bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700'
                  }`}
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

                <p className="text-xs text-center text-gray-500">
                  14-day free trial • Cancel anytime
                </p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Additional Info */}
      <div className="mt-16 text-center space-y-4">
        <h3 className="text-2xl font-bold text-gray-900">All plans include:</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mt-8">
          <div className="text-center">
            <div className="text-3xl font-bold text-blue-600 mb-2">24/7</div>
            <div className="text-gray-600">Customer Support</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-blue-600 mb-2">99.9%</div>
            <div className="text-gray-600">Uptime SLA</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-blue-600 mb-2">SSL</div>
            <div className="text-gray-600">Secure Encryption</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-blue-600 mb-2">Free</div>
            <div className="text-gray-600">Updates & Backups</div>
          </div>
        </div>
      </div>
    </div>
  );
}
