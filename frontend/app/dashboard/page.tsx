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
import { useMemo } from 'react';
import Link from 'next/link';

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
  const { data: analytics, isLoading: analyticsLoading, error: analyticsError } = useQuery<AnalyticsDashboard>({
    queryKey: ['dashboard-analytics'],
    queryFn: analyticsService.getDashboardAnalytics,
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 60000, // 1 minute
    gcTime: 5 * 60 * 1000, // 5 minutes
    refetchOnWindowFocus: false,
  });

  const { data: lowStock, isLoading: lowStockLoading, error: lowStockError } = useQuery<LowStockItem[]>({
    queryKey: ['dashboard-low-stock'],
    queryFn: () => analyticsService.getLowStockReport({ threshold: 10 }),
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
    staleTime: 2 * 60 * 1000, // 2 minutes
    gcTime: 5 * 60 * 1000,
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
    staleTime: 60000, // 1 minute
    gcTime: 5 * 60 * 1000,
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
      title: 'Revenue',
      value: analytics ? formatCurrency(analytics.totalSales ?? 0) : formatCurrency(0),
      icon: DollarSign,
      color: 'text-blue-600 dark:text-blue-400',
      bgColor: 'bg-blue-50 dark:bg-blue-900/20',
      helper: analytics?.lastUpdated,
    },
    {
      title: 'Stock Value',
      value: analytics ? formatCurrency(analytics.totalStockValue ?? 0) : formatCurrency(0),
      icon: Warehouse,
      color: 'text-purple-600 dark:text-purple-400',
      bgColor: 'bg-purple-50 dark:bg-purple-900/20',
    },
    {
      title: 'Orders',
      value: analytics?.orderCount ?? 0,
      icon: ShoppingCart,
      color: 'text-emerald-600 dark:text-emerald-400',
      bgColor: 'bg-emerald-50 dark:bg-emerald-900/20',
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
    return (
      <div className="space-y-6 p-6">
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Error Loading Dashboard</AlertTitle>
          <AlertDescription>
            {errorMessage.includes('Network')
              ? 'Network error. Please check your connection and try again.'
              : 'We encountered an error while loading your dashboard data. Please try refreshing the page.'}
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-8 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-3xl font-bold tracking-tight">
              Dashboard
            </h1>
          </div>
          <p className="text-muted-foreground">Operations Overview</p>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => (
          <Card key={stat.title}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.title}
              </CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              {stat.helper && (
                <p className="text-xs text-muted-foreground">
                  Updated {formatDistance(new Date(stat.helper), new Date(), { addSuffix: true })}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            {unpaidInvoiceItems.length ? (
              <div className="space-y-4">
                {unpaidInvoiceItems.slice(0, 5).map((invoice) => (
                  <div
                    key={invoice.id}
                    className="flex items-center justify-between p-4 rounded-lg border"
                  >
                    <div>
                      <p className="font-semibold">Invoice #{invoice.invoice_number}</p>
                      <p className="text-xs text-muted-foreground">
                        Due {format(new Date(invoice.due_date), 'MMM dd, yyyy')} · Status: {invoice.status}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold">{formatCurrency(invoice.total_amount || 0)}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState
                icon={FileText}
                title="No Unpaid Invoices"
                description="All invoices are paid or no invoices exist yet."
              />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Low Stock Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            {lowStockItems.length > 0 ? (
              <div className="space-y-4">
                {lowStockItems.slice(0, 5).map((item) => (
                  <div
                    key={`${item.productId}-${item.warehouseId}`}
                    className="flex items-center justify-between p-4 bg-muted/50 border rounded-lg"
                  >
                    <div>
                      <p className="font-semibold">{item.productName || 'Unnamed Product'}</p>
                      <p className="text-xs text-muted-foreground">Warehouse: {item.warehouseId?.slice(0, 8) ?? 'N/A'}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-destructive">{item.currentStock} units</p>
                      <p className="text-xs text-muted-foreground">Threshold: {item.threshold}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState
                icon={Package}
                title="All Items Well Stocked"
                description="No low stock alerts at this time. Great job!"
              />
            )}
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Link
              href="/dashboard/products"
              className="group flex flex-col items-center justify-center p-6 bg-muted/30 rounded-xl hover:bg-muted/50 transition-colors border"
            >
              <div className="p-3 bg-background rounded-lg mb-3 shadow-sm group-hover:scale-110 transition-transform">
                <Package className="h-5 w-5 text-primary" />
              </div>
              <span className="text-sm font-semibold">Add Product</span>
            </Link>
            <Link
              href="/dashboard/orders"
              className="group flex flex-col items-center justify-center p-6 bg-muted/30 rounded-xl hover:bg-muted/50 transition-colors border"
            >
              <div className="p-3 bg-background rounded-lg mb-3 shadow-sm group-hover:scale-110 transition-transform">
                <ShoppingCart className="h-5 w-5 text-primary" />
              </div>
              <span className="text-sm font-semibold">Process Order</span>
            </Link>
            <Link
              href="/dashboard/inventory"
              className="group flex flex-col items-center justify-center p-6 bg-muted/30 rounded-xl hover:bg-muted/50 transition-colors border"
            >
              <div className="p-3 bg-background rounded-lg mb-3 shadow-sm group-hover:scale-110 transition-transform">
                <Warehouse className="h-5 w-5 text-primary" />
              </div>
              <span className="text-sm font-semibold">Stock Levels</span>
            </Link>
            <Link
              href="/dashboard/invoices"
              className="group flex flex-col items-center justify-center p-6 bg-muted/30 rounded-xl hover:bg-muted/50 transition-colors border"
            >
              <div className="p-3 bg-background rounded-lg mb-3 shadow-sm group-hover:scale-110 transition-transform">
                <FileText className="h-5 w-5 text-primary" />
              </div>
              <span className="text-sm font-semibold">Billing</span>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
