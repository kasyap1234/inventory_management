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
      bgColor: 'bg-emerald-50',
      borderColor: 'border-emerald-200',
      helper: analytics?.lastUpdated,
    },
    {
      title: 'Inventory Value',
      value: analytics ? formatCurrency(analytics.totalStockValue ?? 0) : formatCurrency(0),
      icon: Warehouse,
      color: 'text-sky-600',
      bgColor: 'bg-sky-50',
      borderColor: 'border-sky-200',
    },
    {
      title: 'Orders Processed',
      value: analytics?.orderCount ?? 0,
      icon: ShoppingCart,
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
      borderColor: 'border-purple-200',
    },
    {
      title: 'Unpaid Invoices',
      value: unpaidInvoiceItems.length,
      icon: FileText,
      color: 'text-amber-600',
      bgColor: 'bg-amber-50',
      borderColor: 'border-amber-200',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header - Clean and modern */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold text-foreground tracking-tight">
            Dashboard
          </h1>
          <p className="text-muted-foreground mt-2">Welcome back! Here&rsquo;s what&rsquo;s happening today.</p>
        </div>
      </div>

      {/* Stats Cards - Clean, modern design */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <Card 
            key={stat.title}
            className={`card-hover border ${stat.borderColor} bg-white shadow-sm overflow-hidden`}
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-2 pt-5">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
              <div className={`p-2.5 rounded-lg ${stat.bgColor}`}>
                <stat.icon className={`h-5 w-5 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent className="pb-5">
              <div className="text-2xl font-bold text-foreground">
                {isLoading ? (
                  <div className="h-8 w-24 bg-muted rounded animate-pulse"></div>
                ) : (
                  stat.value
                )}
              </div>
              {stat.helper && !isLoading && (
                <p className="text-xs text-muted-foreground mt-1">
                  Updated {formatDistance(new Date(stat.helper), new Date(), { addSuffix: true })}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground">Recent Activity</CardTitle>
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
                      <p className="font-semibold text-foreground">Invoice #{invoice.invoice_number}</p>
                      <p className="text-xs text-muted-foreground">
                        Due {format(new Date(invoice.due_date), 'MMM dd, yyyy')} · Status: {invoice.status}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-foreground">{formatCurrency(invoice.total_amount || 0)}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-4 text-center py-12 text-muted-foreground">
                <Warehouse className="h-12 w-12 text-muted-foreground/50 mx-auto mb-3" />
                No unpaid invoices at the moment
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground">Low Stock Alerts</CardTitle>
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
                      <p className="font-semibold text-foreground">{item.productName || 'Unnamed Product'}</p>
                      <p className="text-xs text-muted-foreground">Warehouse: {item.warehouseId?.slice(0, 8) ?? 'N/A'}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-orange-600">{item.currentStock} units</p>
                      <p className="text-xs text-muted-foreground">Threshold: {item.threshold}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-4 text-center py-12 text-muted-foreground">
                <Package className="h-12 w-12 text-muted-foreground/50 mx-auto mb-3" />
                All items are well stocked
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions with modern cards */}
      <Card className="border-0 shadow-md bg-gradient-to-br from-white to-gray-50">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold text-foreground">Quick Actions</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <a
              href="/dashboard/products"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-green-50 to-green-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-green-100"
            >
              <div className="p-3 bg-green-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <Package className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-green-900">Add Product</span>
            </a>
            <a
              href="/dashboard/orders"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-green-50 to-green-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-green-100"
            >
              <div className="p-3 bg-green-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <ShoppingCart className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-green-900">New Order</span>
            </a>
            <a
              href="/dashboard/inventory"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-amber-50 to-amber-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-amber-100"
            >
              <div className="p-3 bg-amber-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <Warehouse className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-amber-900">Check Inventory</span>
            </a>
            <a
              href="/dashboard/invoices"
              className="group flex flex-col items-center justify-center p-6 bg-gradient-to-br from-orange-50 to-orange-100/50 rounded-2xl hover:shadow-lg hover:-translate-y-1 transition-all duration-200 border border-orange-100"
            >
              <div className="p-3 bg-orange-600 rounded-xl mb-3 group-hover:scale-110 transition-transform duration-200 shadow-md">
                <FileText className="h-6 w-6 text-white" />
              </div>
              <span className="text-sm font-semibold text-orange-900">View Invoices</span>
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
