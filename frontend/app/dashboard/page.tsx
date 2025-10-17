'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package, Warehouse, ShoppingCart, FileText, DollarSign } from 'lucide-react';
import { analyticsService, invoiceService, type AnalyticsDashboard, type LowStockItem } from '@/lib/services';
import { formatCurrency } from '@/lib/utils';
import { format, formatDistance } from 'date-fns';

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
  const { data: analytics, isLoading: analyticsLoading } = useQuery<AnalyticsDashboard>({
    queryKey: ['dashboard-analytics'],
    queryFn: analyticsService.getDashboardAnalytics,
  });

  const { data: lowStock, isLoading: lowStockLoading } = useQuery<LowStockItem[]>({
    queryKey: ['dashboard-low-stock'],
    queryFn: () => analyticsService.getLowStockReport({ threshold: 10 }),
  });

  const { data: unpaidInvoices, isLoading: invoicesLoading } = useQuery<UnpaidInvoicesResponse>({
    queryKey: ['dashboard-unpaid-invoices'],
    queryFn: async () => {
      const response = await invoiceService.getUnpaid();
      return response.data;
    },
  });

  const unpaidInvoiceItems: InvoiceSummary[] = Array.isArray(unpaidInvoices?.invoices)
    ? unpaidInvoices.invoices
    : [];

  const lowStockItems: LowStockItem[] = lowStock ?? [];

  const isLoading = analyticsLoading || lowStockLoading || invoicesLoading;

  const statCards = [
    {
      title: 'Total Sales',
      value: analytics ? formatCurrency(analytics.totalSales ?? 0) : formatCurrency(0),
      icon: DollarSign,
      color: 'text-emerald-600',
      bgColor: 'bg-gradient-to-br from-emerald-50 to-emerald-100/50',
      iconBg: 'bg-gradient-to-br from-emerald-500 to-emerald-600',
      helper: analytics?.lastUpdated,
    },
    {
      title: 'Inventory Value',
      value: analytics ? formatCurrency(analytics.totalStockValue ?? 0) : formatCurrency(0),
      icon: Warehouse,
      color: 'text-indigo-600',
      bgColor: 'bg-gradient-to-br from-indigo-50 to-indigo-100/50',
      iconBg: 'bg-gradient-to-br from-indigo-500 to-indigo-600',
    },
    {
      title: 'Orders Processed',
      value: analytics?.orderCount ?? 0,
      icon: ShoppingCart,
      color: 'text-purple-600',
      bgColor: 'bg-gradient-to-br from-purple-50 to-purple-100/50',
      iconBg: 'bg-gradient-to-br from-purple-500 to-purple-600',
    },
    {
      title: 'Unpaid Invoices',
      value: unpaidInvoiceItems.length,
      icon: FileText,
      color: 'text-pink-600',
      bgColor: 'bg-gradient-to-br from-pink-50 to-pink-100/50',
      iconBg: 'bg-gradient-to-br from-pink-500 to-pink-600',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header with gradient */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold gradient-text mb-2">
            Dashboard
          </h1>
          <p className="text-gray-600 text-lg">Welcome back! Here&rsquo;s what&rsquo;s happening today.</p>
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
                {isLoading ? (
                  <div className="h-9 w-24 skeleton rounded"></div>
                ) : (
                  stat.value
                )}
              </div>
              {stat.helper && !isLoading && (
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
          <CardHeader className="border-b border-gray-100 bg-gradient-to-r from-indigo-50/50 to-purple-50/50">
            <CardTitle className="text-lg font-bold gradient-text">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {invoicesLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((item) => (
                  <div key={item} className="h-14 bg-gray-200 rounded-lg animate-pulse"></div>
                ))}
              </div>
            ) : unpaidInvoiceItems.length ? (
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
              <div className="space-y-4 text-center py-12 text-gray-500">
                <Warehouse className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                No unpaid invoices at the moment
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-0 shadow-elegant hover-lift">
          <CardHeader className="border-b border-gray-100 bg-gradient-to-r from-amber-50/50 to-orange-50/50">
            <CardTitle className="text-lg font-bold gradient-text">Low Stock Alerts</CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {lowStockLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((item) => (
                  <div key={item} className="h-14 bg-gray-200 rounded-lg animate-pulse"></div>
                ))}
              </div>
            ) : lowStockItems.length > 0 ? (
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
              <div className="space-y-4 text-center py-12 text-gray-500">
                <Package className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                All items are well stocked
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions with modern cards */}
      <Card className="border-0 shadow-elegant hover-lift bg-gradient-to-br from-white via-indigo-50/20 to-purple-50/20">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold gradient-text">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <a
              href="/dashboard/products"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-indigo-50 to-indigo-100/50 rounded-2xl hover:shadow-colored hover:-translate-y-2 transition-all duration-300 border border-indigo-100 hover:border-indigo-200"
            >
              <div className="p-3 bg-gradient-to-br from-indigo-500 to-indigo-600 rounded-xl mb-3 group-hover:scale-125 transition-transform duration-300 shadow-colored">
                <Package className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-indigo-900">Add Product</span>
            </a>
            <a
              href="/dashboard/orders"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-emerald-50 to-emerald-100/50 rounded-2xl hover:shadow-colored hover:-translate-y-2 transition-all duration-300 border border-emerald-100 hover:border-emerald-200"
            >
              <div className="p-3 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-xl mb-3 group-hover:scale-125 transition-transform duration-300 shadow-colored">
                <ShoppingCart className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-emerald-900">New Order</span>
            </a>
            <a
              href="/dashboard/inventory"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-purple-50 to-purple-100/50 rounded-2xl hover:shadow-colored hover:-translate-y-2 transition-all duration-300 border border-purple-100 hover:border-purple-200"
            >
              <div className="p-3 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl mb-3 group-hover:scale-125 transition-transform duration-300 shadow-colored">
                <Warehouse className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-purple-900">Check Inventory</span>
            </a>
            <a
              href="/dashboard/invoices"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-pink-50 to-pink-100/50 rounded-2xl hover:shadow-colored hover:-translate-y-2 transition-all duration-300 border border-pink-100 hover:border-pink-200"
            >
              <div className="p-3 bg-gradient-to-br from-pink-500 to-pink-600 rounded-xl mb-3 group-hover:scale-125 transition-transform duration-300 shadow-colored">
                <FileText className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-pink-900">View Invoices</span>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
