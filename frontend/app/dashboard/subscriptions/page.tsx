'use client';

import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CreditCard, Pause, Play, X, Trash2 } from 'lucide-react';
import {
  subscriptionService,
  type SubscriptionDto,
  type SubscriptionListResult,
  type SubscriptionPlanConfig,
} from '@/lib/services';
import { Button } from '@/components/ui/button';
import { formatCurrency } from '@/lib/utils';
import { format } from 'date-fns';
import SubscriptionUsageCard from '@/components/subscriptions/SubscriptionUsageCard';
import api from '@/lib/api';
import { useRouter } from 'next/navigation';

export default function SubscriptionsPage() {
  const queryClient = useQueryClient();
  const router = useRouter();

  const { data: subscriptionsData, isLoading } = useQuery<SubscriptionListResult>({
    queryKey: ['subscriptions'],
    queryFn: () => subscriptionService.list(),
  });

  const { data: availablePlans } = useQuery<SubscriptionPlanConfig[]>({
    queryKey: ['subscription-plans'],
    queryFn: () => subscriptionService.getAvailablePlans(),
  });

  // Fetch usage and limits
  const { data: usageData } = useQuery({
    queryKey: ['subscription-usage'],
    queryFn: async () => {
      const response = await api.get('/subscriptions/usage');
      return response.data;
    },
  });

  const subscriptions: SubscriptionDto[] = subscriptionsData?.items ?? [];

  const cancelMutation = useMutation({
    mutationFn: (id: string) => subscriptionService.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
    },
  });

  const pauseMutation = useMutation({
    mutationFn: (id: string) => subscriptionService.pause(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
    },
  });

  const resumeMutation = useMutation({
    mutationFn: (id: string) => subscriptionService.resume(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => subscriptionService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
    },
  });

  const planIntervalMap = useMemo(() => {
    if (!availablePlans) {
      return new Map<string, string>();
    }
    const map = new Map<string, string>();
    availablePlans.forEach((plan) => {
      map.set(plan.name, plan.interval);
    });
    return map;
  }, [availablePlans]);

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'paused':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'cancelled':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'expired':
        return 'bg-gray-100 text-gray-800 border-gray-200';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
            Subscriptions
          </h1>
          <p className="text-gray-600 mt-2 text-lg">
            Manage your subscription plans and billing
          </p>
        </div>
        <Button 
          onClick={() => window.location.href = '/dashboard/subscriptions/plans'}
          className="bg-gradient-to-r from-blue-600 to-purple-600 text-white hover:shadow-lg transition-all"
        >
          <CreditCard className="h-4 w-4 mr-2" />
          View Plans
        </Button>
      </div>

      {/* Usage Card */}
      {usageData?.usage && usageData?.limits && (
        <SubscriptionUsageCard
          usage={usageData.usage}
          limits={usageData.limits}
          onUpgrade={() => router.push('/dashboard/subscriptions/plans')}
        />
      )}

      {/* Subscriptions List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {isLoading ? (
          <>
            {[1, 2].map((i) => (
              <div key={i} className="h-64 bg-gray-200 rounded-lg animate-pulse"></div>
            ))}
          </>
        ) : subscriptions.length > 0 ? (
          subscriptions.map((subscription) => {
            const intervalLabel = planIntervalMap.get(subscription.plan_name) ?? 'monthly';
            const amountDisplay = formatCurrency(subscription.amount || 0, subscription.currency || 'INR');

            return (
            <Card key={subscription.id} className="border-0 shadow-md hover:shadow-lg transition-all">
              <CardHeader className="border-b border-gray-100">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-xl font-bold text-gray-900">
                    {subscription.plan_name || 'Subscription'}
                  </CardTitle>
                  <span
                    className={`px-3 py-1 text-xs font-semibold rounded-full border ${getStatusColor(
                      subscription.status
                    )}`}
                  >
                    {subscription.status}
                  </span>
                </div>
              </CardHeader>
              <CardContent className="pt-6">
                <div className="space-y-4">
                  {/* Price */}
                  <div>
                    <div className="text-3xl font-bold text-gray-900">
                      {amountDisplay}
                      <span className="text-lg font-normal text-gray-500">
                        /{intervalLabel}
                      </span>
                    </div>
                  </div>

                  {/* Details */}
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Start Date:</span>
                      <span className="font-medium text-gray-900">
                        {format(new Date(subscription.start_date), 'MMM dd, yyyy')}
                      </span>
                    </div>
                    {subscription.end_date && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">End Date:</span>
                        <span className="font-medium text-gray-900">
                          {format(new Date(subscription.end_date), 'MMM dd, yyyy')}
                        </span>
                      </div>
                    )}
                    {subscription.currency && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">Currency:</span>
                        <span className="font-medium text-gray-900">{subscription.currency}</span>
                      </div>
                    )}
                    {subscription.razorpay_subscription_id && (
                      <div className="flex justify-between">
                        <span className="text-gray-600">Gateway ID:</span>
                        <span className="font-medium text-gray-900">
                          {subscription.razorpay_subscription_id}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Actions */}
                  <div className="pt-4 border-t border-gray-200 flex gap-2">
                    {subscription.status === 'active' && (
                      <>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => pauseMutation.mutate(subscription.id)}
                          className="flex-1"
                        >
                          <Pause className="h-4 w-4 mr-1" />
                          Pause
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => cancelMutation.mutate(subscription.id)}
                          className="flex-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <X className="h-4 w-4 mr-1" />
                          Cancel
                        </Button>
                      </>
                    )}
                    {subscription.status === 'paused' && (
                      <>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => resumeMutation.mutate(subscription.id)}
                          className="flex-1"
                        >
                          <Play className="h-4 w-4 mr-1" />
                          Resume
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => cancelMutation.mutate(subscription.id)}
                          className="flex-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <X className="h-4 w-4 mr-1" />
                          Cancel
                        </Button>
                      </>
                    )}
                    {(subscription.status === 'cancelled' || subscription.status === 'expired') && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => deleteMutation.mutate(subscription.id)}
                        className="flex-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        <Trash2 className="h-4 w-4 mr-1" />
                        Delete
                      </Button>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
            );
          })
        ) : (
          <Card className="border-0 shadow-md col-span-2">
            <CardContent className="py-16">
              <div className="text-center">
                <CreditCard className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  No active subscriptions
                </h3>
                <p className="text-gray-500 mb-6">
                  Get started by subscribing to a plan
                </p>
                <Button 
                  onClick={() => window.location.href = '/dashboard/subscriptions/plans'}
                  className="bg-gradient-to-r from-blue-600 to-purple-600 text-white"
                >
                  View Plans
                </Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
