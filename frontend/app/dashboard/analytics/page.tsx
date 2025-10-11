'use client';

import { useMemo, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { 
  TrendingUp, 
  DollarSign, 
  Package, 
  AlertTriangle,
  BarChart3,
  PieChart as PieChartIcon,
  RefreshCw,
  LineChart as LineChartIcon
} from 'lucide-react';
import {
  analyticsService,
  categoryService,
  type AnalyticsDashboard,
  type AnalyticsSalesTrends,
  type ProductSales,
  type LowStockItem,
  type OrderStatusEntry,
  type RevenueByCategory,
  type InventoryValuation,
} from '@/lib/services';
import { formatCurrency } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
} from 'recharts';
import { format } from 'date-fns';
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart';

type CategoryOption = {
  id: string;
  name: string;
};

const SALES_TREND_CHART_CONFIG: ChartConfig = {
  revenue: {
    label: 'Revenue',
    color: 'hsl(217 91% 60%)',
  },
  orders: {
    label: 'Orders',
    color: 'hsl(24 95% 53%)',
  },
};

const TOP_PRODUCTS_CHART_CONFIG: ChartConfig = {
  unitsSold: {
    label: 'Units Sold',
    color: 'hsl(217 83% 53%)',
  },
};

const REVENUE_BY_CATEGORY_CHART_CONFIG: ChartConfig = {
  totalRevenue: {
    label: 'Revenue',
    color: 'hsl(142 71% 45%)',
  },
};

export default function AnalyticsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const queryClient = useQueryClient();

  const { data: dashboardData, isLoading: dashboardLoading } = useQuery<AnalyticsDashboard>({
    queryKey: ['analytics-dashboard'],
    queryFn: analyticsService.getDashboardAnalytics,
  });

  const { data: salesTrendsData, isLoading: salesTrendLoading } = useQuery<AnalyticsSalesTrends>({
    queryKey: ['analytics-sales-trends'],
    queryFn: () => analyticsService.getSalesTrends(),
  });

  const { data: topProductsData, isLoading: topProductsLoading } = useQuery<ProductSales[]>({
    queryKey: ['analytics-top-products'],
    queryFn: () => analyticsService.getTopProducts({ limit: 10 }),
  });

  const { data: lowStockData, isLoading: lowStockLoading } = useQuery<LowStockItem[]>({
    queryKey: ['analytics-low-stock'],
    queryFn: () => analyticsService.getLowStockReport({ threshold: 10 }),
  });

  const { data: orderStatusData, isLoading: orderStatusLoading } = useQuery<OrderStatusEntry[]>({
    queryKey: ['analytics-order-status'],
    queryFn: () => analyticsService.getOrderStatusDistribution(),
  });

  const { data: revenueByCategoryData, isLoading: revenueByCategoryLoading } = useQuery<RevenueByCategory[]>({
    queryKey: ['analytics-revenue-category'],
    queryFn: () => analyticsService.getRevenueByCategory(),
  });

  const { data: inventoryValuationData, isLoading: inventoryValuationLoading } = useQuery<InventoryValuation>({
    queryKey: ['analytics-inventory-valuation'],
    queryFn: analyticsService.getInventoryValuation,
  });

  const { data: categoriesData } = useQuery<CategoryOption[]>({
    queryKey: ['categories-all'],
    queryFn: async () => {
      const response = await categoryService.list();
      const raw = Array.isArray(response.data?.categories) ? response.data.categories : [];
      return raw
        .filter((category: any): category is { id: string; name: string } =>
          typeof category?.id === 'string' && typeof category?.name === 'string'
        )
        .map((category: any) => ({ id: category.id, name: category.name }));
    },
    staleTime: 5 * 60 * 1000,
  });

  const salesTrendChartData = useMemo((): Array<{
    dateISO: string | null;
    label: string;
    revenue: number;
    orders: number;
  }> => {
    if (!salesTrendsData?.trends?.length) {
      return [];
    }

    return salesTrendsData.trends.map((trend) => ({
      dateISO: trend.date ?? null,
      label: trend.date ? format(new Date(trend.date), 'MMM dd') : 'Unknown',
      revenue: trend.salesAmount ?? 0,
      orders: trend.orderCount ?? 0,
    }));
  }, [salesTrendsData]);

  const revenueByCategoryChartData = useMemo((): Array<{
    categoryId: string;
    label: string;
    totalRevenue: number;
  }> => {
    if (!revenueByCategoryData?.length) {
      return [];
    }

    const categoryMap = new Map<string, string>(
      (categoriesData ?? []).map((category) => [category.id, category.name])
    );

    return revenueByCategoryData.map((item) => ({
      categoryId: item.categoryId,
      label:
        categoryMap.get(item.categoryId) ??
        (item.categoryId === 'uncategorized'
          ? 'Uncategorized'
          : `Category ${item.categoryId.slice(0, 8)}`),
      totalRevenue: item.totalRevenue,
    }));
  }, [revenueByCategoryData, categoriesData]);

  const revenueChange = useMemo(() => computePercentageChange(salesTrendChartData.map((trend) => trend.revenue)), [
    salesTrendChartData,
  ]);

  const orderCountChange = useMemo(
    () => computePercentageChange(salesTrendChartData.map((trend) => trend.orders)),
    [salesTrendChartData]
  );

  const statCards = useMemo(
    () => [
      {
        title: 'Total Revenue',
        value: dashboardData ? formatCurrency(dashboardData.totalSales) : '—',
        icon: DollarSign,
        color: 'text-green-600',
        bgColor: 'bg-green-50',
        change: revenueChange,
        isLoading: dashboardLoading || salesTrendLoading,
        changeLabel: 'vs. first data point',
      },
      {
        title: 'Total Orders',
        value: dashboardData?.orderCount ?? 0,
        icon: TrendingUp,
        color: 'text-blue-600',
        bgColor: 'bg-blue-50',
        change: orderCountChange,
        isLoading: dashboardLoading || salesTrendLoading,
        changeLabel: 'vs. first data point',
      },
      {
        title: 'Inventory Value',
        value: inventoryValuationData ? formatCurrency(inventoryValuationData.totalValue) : '—',
        icon: Package,
        color: 'text-purple-600',
        bgColor: 'bg-purple-50',
        change: 'N/A',
        isLoading: inventoryValuationLoading,
        changeLabel: '',
      },
      {
        title: 'Low Stock Items',
        value: lowStockData?.length ?? 0,
        icon: AlertTriangle,
        color: 'text-orange-600',
        bgColor: 'bg-orange-50',
        change: 'N/A',
        isLoading: lowStockLoading,
        changeLabel: '',
      },
    ],
    [
      dashboardData,
      revenueChange,
      orderCountChange,
      inventoryValuationData,
      dashboardLoading,
      salesTrendLoading,
      inventoryValuationLoading,
      lowStockData,
      lowStockLoading,
    ]
  );

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await analyticsService.refreshAnalytics();
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['analytics-dashboard'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-sales-trends'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-top-products'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-low-stock'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-order-status'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-revenue-category'] }),
        queryClient.invalidateQueries({ queryKey: ['analytics-inventory-valuation'] }),
      ]);
    } catch (error) {
      // Error is already handled by react-query
      // Log to error tracking service in production
    } finally {
      setRefreshing(false);
    }
  };

  const topProducts: ProductSales[] = useMemo(() => topProductsData ?? [], [topProductsData]);
  const lowStockItems: LowStockItem[] = useMemo(() => lowStockData ?? [], [lowStockData]);
  const orderStatus: OrderStatusEntry[] = useMemo(() => orderStatusData ?? [], [orderStatusData]);
  const orderStatusChartConfig = useMemo<ChartConfig>(() => {
    if (!orderStatus.length) {
      return {};
    }

    return orderStatus.reduce<ChartConfig>((acc, item, index) => {
      const statusKey = item.status ?? `status-${index}`;
      const readable = statusKey.replace(/_/g, ' ');
      acc[statusKey] = {
        label: readable.charAt(0).toUpperCase() + readable.slice(1),
        color: STATUS_COLORS[item.status ?? 'default'] ?? STATUS_COLORS.default,
      };
      return acc;
    }, {});
  }, [orderStatus]);

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
            Analytics & Reports
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Comprehensive insights into your business performance
          </p>
        </div>
        <Button
          onClick={handleRefresh}
          disabled={refreshing}
          className="bg-gradient-to-r from-blue-600 to-purple-600 text-white hover:shadow-lg transition-all"
        >
          <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
          Refresh Data
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <Card
            key={stat.title}
            className="card-hover border-0 shadow-md bg-white overflow-hidden"
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-3 pt-6">
              <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
                {stat.title}
              </CardTitle>
              <div className={`p-3 rounded-xl ${stat.bgColor} shadow-sm`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-foreground mb-1">
                {stat.isLoading ? (
                  <div className="h-9 w-24 bg-gray-200 rounded animate-pulse" />
                ) : typeof stat.value === 'number' ? (
                  stat.value.toLocaleString()
                ) : (
                  stat.value
                )}
              </div>
              {stat.change !== 'N/A' && (
                <p className="text-sm text-muted-foreground flex items-center">
                  <span
                    className={`font-medium mr-1 ${
                      stat.change.startsWith('-')
                        ? 'text-red-600'
                        : stat.change.startsWith('+')
                        ? 'text-green-600'
                        : 'text-muted-foreground'
                    }`}
                  >
                    {stat.change}
                  </span>
                  {stat.changeLabel}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="border-0 shadow-md">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold text-foreground flex items-center">
            <LineChartIcon className="h-5 w-5 mr-2 text-blue-600" />
            Sales Trend (Last 30 Days)
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
        {salesTrendLoading ? (
          <div className="h-72 bg-gray-200 rounded animate-pulse" />
        ) : salesTrendChartData.length ? (
          <ChartContainer config={SALES_TREND_CHART_CONFIG} className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={salesTrendChartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="4 4" className="stroke-muted/40" />
                <XAxis dataKey="label" tickLine={false} axisLine={false} stroke="#888" />
                <YAxis tickLine={false} axisLine={false} stroke="#888" />
                <ChartTooltip
                  cursor={{ strokeDasharray: '4 4' }}
                  content={
                    <ChartTooltipContent
                      indicator="line"
                      formatter={({ value, name }) =>
                        name === 'revenue'
                          ? formatCurrency(Number(value ?? 0))
                          : `${Number(value ?? 0)} orders`
                      }
                    />
                  }
                />
                <ChartLegend verticalAlign="top" content={<ChartLegendContent className="pt-2" />} />
                <Line
                  type="monotone"
                  dataKey="revenue"
                  stroke="var(--color-revenue)"
                  strokeWidth={3}
                  dot={{ fill: 'var(--color-revenue)', r: 4 }}
                  activeDot={{ r: 6 }}
                  name="Revenue"
                />
                <Line
                  type="monotone"
                  dataKey="orders"
                  stroke="var(--color-orders)"
                  strokeWidth={2}
                  dot={false}
                  name="Orders"
                />
              </LineChart>
            </ResponsiveContainer>
          </ChartContainer>
        ) : (
          <p className="text-center text-muted-foreground py-10">
            No sales data available for the selected period.
          </p>
        )}
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground flex items-center">
              <BarChart3 className="h-5 w-5 mr-2 text-blue-600" />
              Top Selling Products
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {topProductsLoading ? (
              <div className="h-80 bg-gray-200 rounded animate-pulse" />
            ) : topProducts.length ? (
              <ChartContainer config={TOP_PRODUCTS_CHART_CONFIG} className="h-[320px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={topProducts.slice(0, 5)} margin={{ top: 5, right: 30, left: 20, bottom: 50 }}>
                    <CartesianGrid strokeDasharray="4 4" className="stroke-muted/40" />
                    <XAxis
                      dataKey="productName"
                      stroke="#888"
                      angle={-15}
                      textAnchor="end"
                      height={80}
                    />
                    <YAxis stroke="#888" />
                    <ChartTooltip
                      cursor={{ fill: 'rgba(148, 163, 184, 0.15)' }}
                      content={
                        <ChartTooltipContent
                          formatter={({ value }) => `${Number(value ?? 0)} units`}
                        />
                      }
                    />
                    <Bar
                      dataKey="unitsSold"
                      fill="var(--color-unitsSold)"
                      name="Units Sold"
                      radius={[8, 8, 0, 0]}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </ChartContainer>
            ) : (
              <p className="text-center text-muted-foreground py-10">No product sales data available.</p>
            )}
          </CardContent>
        </Card>

        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground flex items-center">
              <AlertTriangle className="h-5 w-5 mr-2 text-orange-600" />
              Low Stock Alert
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {lowStockLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="h-12 bg-gray-200 rounded animate-pulse" />
                ))}
              </div>
            ) : lowStockItems.length ? (
              <div className="space-y-4">
                {lowStockItems.slice(0, 5).map((item) => (
                  <div
                    key={`${item.productId}-${item.warehouseId}`}
                    className="flex items-center justify-between p-3 bg-orange-50 rounded-lg border border-orange-200"
                  >
                    <div>
                      <p className="font-semibold text-foreground">{item.productName || 'Unnamed Product'}</p>
                      <p className="text-sm text-muted-foreground">Warehouse: {item.warehouseId?.slice(0, 8) ?? 'N/A'}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-bold text-orange-600">{item.currentStock} units</p>
                      <p className="text-xs text-muted-foreground">Threshold: {item.threshold}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-center text-muted-foreground py-12">
                <Package className="h-12 w-12 text-muted-foreground/50 mx-auto mb-3" />
                All items are well stocked
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-0 shadow-md">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground flex items-center">
              <PieChartIcon className="h-5 w-5 mr-2 text-purple-600" />
              Order Status Distribution
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {orderStatusLoading ? (
              <div className="h-80 bg-gray-200 rounded animate-pulse" />
            ) : orderStatus.length ? (
              <ChartContainer config={orderStatusChartConfig} className="h-[320px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={orderStatus}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ status, count, percent }: any) =>
                        `${status}: ${count} (${Math.round((Number(percent) || 0) * 100)}%)`
                      }
                      outerRadius={110}
                      dataKey="count"
                      nameKey="status"
                    >
                      {orderStatus.map((entry, index) => (
                        <Cell
                          key={`cell-${index}`}
                          fill={`var(--color-${entry.status ?? `status-${index}`})`}
                        />
                      ))}
                    </Pie>
                    <ChartTooltip
                      content={
                        <ChartTooltipContent
                          formatter={({ value, item }) => {
                            const count = typeof value === 'number' ? value : Number(value ?? 0);
                            const percent = (item?.payload?.percent ?? 0) * 100;
                            return `${count} orders (${Math.round(percent)}%)`;
                          }}
                          labelFormatter={() => null}
                        />
                      }
                    />
                    <ChartLegend
                      layout="horizontal"
                      verticalAlign="bottom"
                      align="center"
                      content={<ChartLegendContent className="pt-4" />}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </ChartContainer>
            ) : (
              <p className="text-center text-muted-foreground py-10">No orders available for the selected range.</p>
            )}
          </CardContent>
        </Card>

        <Card className="border-0 shadow-md lg:col-span-2">
          <CardHeader className="border-b border-gray-100">
            <CardTitle className="text-lg font-bold text-foreground flex items-center">
              <DollarSign className="h-5 w-5 mr-2 text-green-600" />
              Revenue by Category
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            {revenueByCategoryLoading ? (
              <div className="h-80 bg-gray-200 rounded animate-pulse" />
            ) : revenueByCategoryChartData.length ? (
              <ChartContainer config={REVENUE_BY_CATEGORY_CHART_CONFIG} className="h-[320px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={revenueByCategoryChartData} margin={{ top: 5, right: 30, left: 20, bottom: 50 }}>
                    <CartesianGrid strokeDasharray="4 4" className="stroke-muted/40" />
                    <XAxis
                      dataKey="label"
                      stroke="#888"
                      angle={-15}
                      textAnchor="end"
                      height={80}
                    />
                    <YAxis stroke="#888" />
                    <ChartTooltip
                      cursor={{ fill: 'rgba(16, 185, 129, 0.08)' }}
                      content={
                        <ChartTooltipContent
                          formatter={({ value }) => formatCurrency(Number(value ?? 0))}
                        />
                      }
                    />
                    <Bar
                      dataKey="totalRevenue"
                      fill="var(--color-totalRevenue)"
                      name="Revenue"
                      radius={[8, 8, 0, 0]}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </ChartContainer>
            ) : (
              <p className="text-center text-muted-foreground py-10">No category revenue data available.</p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

// Helper constants and functions
const STATUS_COLORS: Record<string, string> = {
  pending: '#f59e0b',
  approved: '#3b82f6',
  processing: '#8b5cf6',
  shipped: '#06b6d4',
  delivered: '#10b981',
  cancelled: '#ef4444',
  default: '#6b7280',
};

function computePercentageChange(values: number[]): string {
  if (!values || values.length < 2) {
    return 'N/A';
  }

  const filtered = values.filter((value) => typeof value === 'number');
  if (filtered.length < 2) {
    return 'N/A';
  }

  const first = filtered[0];
  const last = filtered[filtered.length - 1];

  if (first === 0) {
    return last === 0 ? '0%' : 'N/A';
  }

  const change = ((last - first) / Math.abs(first)) * 100;
  const sign = change > 0 ? '+' : '';
  return `${sign}${change.toFixed(1)}%`;
}
