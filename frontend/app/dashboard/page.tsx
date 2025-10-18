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
      color: 'text-blue-600',
      bgColor: 'bg-gradient-to-br from-blue-50 to-blue-100/50',
      iconBg: 'gradient-blue',
      helper: analytics?.lastUpdated,
    },
    {
      title: 'Stock Value',
      value: analytics ? formatCurrency(analytics.totalStockValue ?? 0) : formatCurrency(0),
      icon: Warehouse,
      color: 'text-purple-600',
      bgColor: 'bg-gradient-to-br from-purple-50 to-purple-100/50',
      iconBg: 'gradient-purple',
    },
    {
      title: 'Orders',
      value: analytics?.orderCount ?? 0,
      icon: ShoppingCart,
      color: 'text-emerald-600',
      bgColor: 'bg-gradient-to-br from-emerald-50 to-emerald-100/50',
      iconBg: 'gradient-emerald',
    },
    {
      title: 'Pending Invoices',
      value: unpaidInvoiceItems.length,
      icon: FileText,
      color: 'text-pink-600',
      bgColor: 'bg-gradient-to-br from-pink-50 to-pink-100/50',
      iconBg: 'gradient-pink',
    },
  ], [analytics, unpaidInvoiceItems.length]);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  const hasErrors = analyticsError || lowStockError || invoicesError;

  if (hasErrors) {
    const errorMessage = analyticsError?.message || lowStockError?.message || invoicesError?.message || 'Unknown error';
    return (
      <div className="space-y-6">
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
    <div className="space-y-8">
      {/* Header with gradient */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl gradient-primary flex items-center justify-center shadow-colored">
              <TrendingUp className="w-6 h-6 text-white" />
            </div>
            <h1 className="text-4xl font-bold gradient-text">
              Dashboard
            </h1>
          </div>
          <p className="text-gray-500 text-base">Operations Overview</p>
        </div>
      </div>

      {/* Stats Cards with modern design */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <Card 
            key={stat.title}
            className="hover-lift border-0 shadow-elegant bg-white overflow-hidden relative group"
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <div className={`absolute inset-0 ${stat.bgColor} opacity-50 group-hover:opacity-70 transition-opacity duration-300`}></div>
            <CardHeader className="relative flex flex-row items-center justify-between pb-3 pt-6">
              <CardTitle className="text-sm font-semibold text-gray-600 uppercase tracking-wide">
                {stat.title}
              </CardTitle>
              <div className={`p-3 rounded-xl ${stat.iconBg} shadow-colored`}>
                <stat.icon className="h-6 w-6 text-white" />
              </div>
            </CardHeader>
            <CardContent className="relative">
              <div className={`text-3xl font-bold mb-1 ${stat.color}`}>
                {stat.value}
              </div>
              {stat.helper && (
                <p className="text-sm text-gray-600">
                  Updated {formatDistance(new Date(stat.helper), new Date(), { addSuffix: true })}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-0 shadow-elegant hover-lift">
          <CardHeader className="border-b border-gray-100 bg-gradient-to-r from-blue-50/30 to-purple-50/30">
            <CardTitle className="text-lg font-bold gradient-text-blue">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {unpaidInvoiceItems.length ? (
              <div className="space-y-4">
                {unpaidInvoiceItems.slice(0, 5).map((invoice) => (
                  <div
                    key={invoice.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-gray-200 hover:border-blue-200 transition-colors"
                  >
                    <div>
                      <p className="font-semibold text-gray-900">Invoice #{invoice.invoice_number}</p>
                      <p className="text-xs text-gray-500">
                        Due {format(new Date(invoice.due_date), 'MMM dd, yyyy')} · Status: {invoice.status}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-gray-900">{formatCurrency(invoice.total_amount || 0)}</p>
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

        <Card className="border-0 shadow-elegant hover-lift">
          <CardHeader className="border-b border-gray-100 bg-gradient-to-r from-amber-50/30 to-orange-50/30">
            <CardTitle className="text-lg font-bold text-amber-700">Low Stock Alerts</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {lowStockItems.length > 0 ? (
              <div className="space-y-4">
                {lowStockItems.slice(0, 5).map((item) => (
                  <div
                    key={`${item.productId}-${item.warehouseId}`}
                    className="flex items-center justify-between p-4 bg-orange-50 border border-orange-200 rounded-lg"
                  >
                    <div>
                      <p className="font-semibold text-gray-900">{item.productName || 'Unnamed Product'}</p>
                      <p className="text-xs text-gray-600">Warehouse: {item.warehouseId?.slice(0, 8) ?? 'N/A'}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-orange-600">{item.currentStock} units</p>
                      <p className="text-xs text-gray-500">Threshold: {item.threshold}</p>
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

      {/* Quick Actions with modern cards */}
      <Card className="border-0 shadow-elegant hover-lift bg-gradient-to-br from-gray-50/50 to-white">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold gradient-text">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <a
              href="/dashboard/products"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-blue-50 to-blue-100/50 rounded-xl hover:shadow-colored hover:-translate-y-1 transition-all duration-300 border border-blue-100 hover:border-blue-200"
            >
              <div className="p-3 gradient-blue rounded-lg mb-3 group-hover:scale-110 transition-transform duration-300 shadow-colored">
                <Package className="h-5 w-5 text-white" />
              </div>
              <span className="text-sm font-semibold text-gray-700">Add Product</span>
            </a>
            <a
              href="/dashboard/orders"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-xl hover:shadow-purple hover:-translate-y-1 transition-all duration-300 border border-purple-100 hover:border-purple-200"
            >
              <div className="p-3 gradient-purple rounded-lg mb-3 group-hover:scale-110 transition-transform duration-300 shadow-purple">
                <ShoppingCart className="h-5 w-5 text-white" />
              </div>
              <span className="text-sm font-semibold text-gray-700">Process Order</span>
            </a>
            <a
              href="/dashboard/inventory"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-emerald-50 to-emerald-100/50 rounded-xl hover:shadow-emerald hover:-translate-y-1 transition-all duration-300 border border-emerald-100 hover:border-emerald-200"
            >
              <div className="p-3 gradient-emerald rounded-lg mb-3 group-hover:scale-110 transition-transform duration-300 shadow-emerald">
                <Warehouse className="h-5 w-5 text-white" />
              </div>
              <span className="text-sm font-semibold text-gray-700">Stock Levels</span>
            </a>
            <a
              href="/dashboard/invoices"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-pink-50 to-pink-100/50 rounded-xl hover:shadow-colored hover:-translate-y-1 transition-all duration-300 border border-pink-100 hover:border-pink-200"
            >
              <div className="p-3 gradient-pink rounded-lg mb-3 group-hover:scale-110 transition-transform duration-300">
                <FileText className="h-5 w-5 text-white" />
              </div>
              <span className="text-sm font-semibold text-gray-700">Billing</span>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
