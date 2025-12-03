'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, Warehouse, ShoppingCart, FileText, DollarSign, TrendingUp, AlertTriangle } from 'lucide-react';
import { analyticsService, invoiceService, type AnalyticsDashboard, type LowStockItem } from '@/lib/services';
import { formatCurrency } from '@/lib/utils';
import { format, formatDistance } from 'date-fns';
import { DashboardSkeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { EmptyState } from '@/components/ui/empty-state';
import { useMemo, useState, useEffect } from 'react';
import Link from 'next/link';
import { SalesTrendChart } from '@/components/charts/SalesTrendChart';

type InvoiceSummary = {
  id: string;
  invoice_number: string;
  total_amount: number;
  status: string;
  due_date: string;
};

type UnpaidInvoicesResponse = {
  invoices: InvoiceSummary[];
  limit: number;
  offset: number;
};

export default function DashboardPage() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const { data: analytics, isLoading: analyticsLoading, error: analyticsError } = useQuery<AnalyticsDashboard>({
    queryKey: ['dashboard-analytics'],
    queryFn: analyticsService.getDashboardAnalytics,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
  });

  const { data: lowStock, isLoading: lowStockLoading, error: lowStockError } = useQuery<LowStockItem[]>({
    queryKey: ['dashboard-low-stock'],
    queryFn: () => analyticsService.getLowStockReport({ threshold: 10 }),
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
  });

  const { data: unpaidInvoices, isLoading: invoicesLoading, error: invoicesError } = useQuery<UnpaidInvoicesResponse>({
    queryKey: ['dashboard-unpaid-invoices'],
    queryFn: async () => {
      const response = await invoiceService.getUnpaid();
      return response.data;
    },
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
  });

  const unpaidInvoiceItems: InvoiceSummary[] = useMemo(
    () => (Array.isArray(unpaidInvoices?.invoices) ? unpaidInvoices.invoices : []),
    [unpaidInvoices]
  );

  const lowStockItems: LowStockItem[] = useMemo(() => lowStock ?? [], [lowStock]);

  const isLoading = analyticsLoading || lowStockLoading || invoicesLoading;

  const statCards = useMemo(() => [
    {
      title: 'REVENUE',
      value: analytics ? formatCurrency(analytics.totalSales ?? 0) : formatCurrency(0),
      icon: DollarSign,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
      helper: analytics?.lastUpdated,
    },
    {
      title: 'STOCK VALUE',
      value: analytics ? formatCurrency(analytics.totalStockValue ?? 0) : formatCurrency(0),
      icon: Warehouse,
      color: 'text-secondary',
      bgColor: 'bg-secondary/10',
    },
    {
      title: 'ORDERS',
      value: analytics?.orderCount ?? 0,
      icon: ShoppingCart,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
    },
    {
      title: 'Pending Invoices',
      value: unpaidInvoiceItems.length,
      icon: FileText,
      color: 'text-pink-600 dark:text-pink-400',
      bgColor: 'bg-pink-50 dark:bg-pink-900/20',
    },
  ], [analytics, unpaidInvoiceItems.length]);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  const hasErrors = analyticsError || lowStockError || invoicesError;

  if (hasErrors) {
    const errorMessage = analyticsError?.message || lowStockError?.message || invoicesError?.message || 'Unknown error';
    console.error('Dashboard errors:', {
      analyticsError,
      lowStockError,
      invoicesError,
    });

    return (
      <div className="space-y-6 p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error Loading Dashboard</AlertTitle>
          <AlertDescription className="space-y-2">
            <p>
              {errorMessage.includes('Network')
                ? 'Network error. Please check your connection and try again.'
                : errorMessage.includes('Missing token') || errorMessage.includes('Unauthorized')
                  ? 'Authentication error. Please try logging out and logging back in.'
                  : 'We encountered an error while loading your dashboard data. Please try refreshing the page.'}
            </p>
            {process.env.NODE_ENV === 'development' && (
              <details className="mt-2">
                <summary className="cursor-pointer text-sm">Error details (dev only)</summary>
                <pre className="mt-2 text-xs bg-muted p-2 rounded overflow-auto">
                  {JSON.stringify({ analyticsError, lowStockError, invoicesError }, null, 2)}
                </pre>
              </details>
            )}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between border-b border-border pb-4">
        <h1 className="text-4xl font-bold tracking-tighter text-foreground">DASHBOARD</h1>
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs text-muted-foreground">
            SYNCED: {format(new Date(), 'HH:mm:ss')}
          </span>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat) => (
          <Card key={stat.title} className="border-l-2 border-l-primary/50 hover:border-l-primary transition-all">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-xs font-medium font-mono tracking-widest text-muted-foreground uppercase">
                {stat.title}
              </CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold font-mono tracking-tighter">{stat.value}</div>
              {stat.helper && (
                <p className="text-xs text-muted-foreground mt-1 font-mono">
                  {stat.helper}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4 border-none bg-transparent">
          <CardHeader>
            <CardTitle className="text-lg font-mono uppercase tracking-widest">Sales Overview - Last 30 Days</CardTitle>
          </CardHeader>
          <CardContent className="pl-2">
            <SalesTrendChart />
          </CardContent>
        </Card>

        {/* Low Stock Alerts */}
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle className="text-lg font-mono uppercase tracking-widest">Low Stock Alert</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {lowStockItems.length === 0 ? (
                <EmptyState
                  icon={Package}
                  title="Stock Healthy"
                  description="All products are well stocked"
                />
              ) : (
                lowStockItems.map((item) => (
                  <div key={item.productId} className="flex items-center justify-between border-b border-border pb-2 last:border-0">
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none font-mono">{item.productName}</p>
                      <p className="text-xs text-muted-foreground">
                        ID: {item.productId.substring(0, 8)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="text-right">
                        <p className="text-sm font-bold text-destructive font-mono">{item.currentStock} left</p>
                        <p className="text-xs text-muted-foreground">
                          Min: {item.threshold}
                        </p>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Unpaid Invoices */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-7">
          <CardHeader>
            <CardTitle className="text-lg font-mono uppercase tracking-widest">Unpaid Invoices</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {unpaidInvoiceItems.length === 0 ? (
                <EmptyState
                  icon={FileText}
                  title="No Unpaid Invoices"
                  description="All invoices are settled"
                />
              ) : (
                unpaidInvoiceItems.map((invoice) => (
                  <div key={invoice.id} className="flex items-center justify-between border-b border-border pb-2 last:border-0">
                    <div className="space-y-1">
                      <p className="text-sm font-medium leading-none font-mono">{invoice.invoice_number}</p>
                      <p className="text-xs text-muted-foreground">
                        Due: {format(new Date(invoice.due_date), 'MMM d, yyyy')}
                      </p>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right">
                        <p className="text-sm font-bold font-mono">{formatCurrency(invoice.total_amount)}</p>
                        <p className="text-xs text-destructive font-medium uppercase tracking-wider">
                          {invoice.status}
                        </p>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
