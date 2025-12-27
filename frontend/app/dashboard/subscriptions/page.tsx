'use client';

import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { CreditCard, Pause, Play, X, Trash2, Calendar, Wallet } from 'lucide-react';
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

  const getStatusVariant = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'default'; // dark/primary
      case 'paused':
        return 'secondary'; // yellow-ish usually, but secondary fits "paused" state
      case 'cancelled':
      case 'expired':
        return 'destructive'; // red
      default:
        return 'outline';
    }
  };

  return (
    <div className="space-y-8 max-w-7xl mx-auto py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-foreground">
            Subscriptions
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage your subscription plans and billing history
          </p>
        </div>
        <Button
          onClick={() => router.push('/dashboard/subscriptions/plans')}
          className="shadow-sm"
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
      <div className="space-y-4">
        <h2 className="text-xl font-semibold tracking-tight">Active Subscriptions</h2>
        <div className="grid grid-cols-1 gap-4">
          {isLoading ? (
            <>
              {[1].map((i) => (
                <div key={i} className="h-48 bg-muted/20 rounded-lg animate-pulse border border-border"></div>
              ))}
            </>
          ) : subscriptions.length > 0 ? (
            subscriptions.map((subscription) => {
              const intervalLabel = planIntervalMap.get(subscription.plan_name) ?? 'monthly';
              const amountDisplay = formatCurrency(subscription.amount || 0, subscription.currency || 'INR');

              return (
                <Card key={subscription.id} className="transition-all hover:border-primary/50">
                  <CardHeader className="pb-4">
                    <div className="flex items-start justify-between">
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <CardTitle className="text-xl">
                            {subscription.plan_name || 'Subscription'}
                          </CardTitle>
                          <Badge variant={getStatusVariant(subscription.status)} className="capitalize">
                            {subscription.status}
                          </Badge>
                        </div>
                        <CardDescription className="flex items-center gap-2">
                          <span className="font-semibold text-foreground text-lg">{amountDisplay}</span>
                          <span>/ {intervalLabel}</span>
                        </CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 pb-6">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Calendar className="h-4 w-4" />
                        <span>Start Date</span>
                      </div>
                      <p className="font-medium">
                        {format(new Date(subscription.start_date), 'MMM dd, yyyy')}
                      </p>
                    </div>

                    {subscription.end_date && (
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Calendar className="h-4 w-4" />
                          <span>End Date</span>
                        </div>
                        <p className="font-medium">
                          {format(new Date(subscription.end_date), 'MMM dd, yyyy')}
                        </p>
                      </div>
                    )}

                    {subscription.razorpay_subscription_id && (
                      <div className="space-y-1">
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Wallet className="h-4 w-4" />
                          <span>Reference ID</span>
                        </div>
                        <p className="font-medium font-mono text-xs truncate" title={subscription.razorpay_subscription_id}>
                          {subscription.razorpay_subscription_id}
                        </p>
                      </div>
                    )}
                  </CardContent>

                  <CardFooter className="bg-muted/30 border-t flex gap-2 justify-end py-3">
                    {subscription.status === 'active' && (
                      <>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => pauseMutation.mutate(subscription.id)}
                        >
                          <Pause className="h-4 w-4 mr-2" />
                          Pause
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => cancelMutation.mutate(subscription.id)}
                        >
                          <X className="h-4 w-4 mr-2" />
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
                        >
                          <Play className="h-4 w-4 mr-2" />
                          Resume
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => cancelMutation.mutate(subscription.id)}
                        >
                          <X className="h-4 w-4 mr-2" />
                          Cancel
                        </Button>
                      </>
                    )}
                    {(subscription.status === 'cancelled' || subscription.status === 'expired') && (
                      <Button
                        size="sm"
                        variant="outline"
                        className="text-destructive hover:text-destructive"
                        onClick={() => deleteMutation.mutate(subscription.id)}
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </Button>
                    )}
                  </CardFooter>
                </Card>
              );
            })
          ) : (
            <Card className="border-dashed shadow-none">
              <CardContent className="py-16">
                <div className="text-center space-y-4">
                  <div className="mx-auto w-12 h-12 rounded-full bg-muted flex items-center justify-center">
                    <CreditCard className="h-6 w-6 text-muted-foreground" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-foreground">
                      No active subscriptions
                    </h3>
                    <p className="text-muted-foreground max-w-sm mx-auto mt-1">
                      You don't have any active subscriptions. Choose a plan to unlock full features.
                    </p>
                  </div>
                  <Button
                    onClick={() => router.push('/dashboard/subscriptions/plans')}
                  >
                    View Plans
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
