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
  LineChart as LineChartIcon,
  Download
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import toast from 'react-hot-toast';
import {
  analyticsService,
  type ProductSales,
  type LowStockItem,
  type OrderStatusEntry,
  type CombinedAnalyticsData,
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
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart';



export default function AnalyticsPage() {
  const [refreshing, setRefreshing] = useState(false);
  const queryClient = useQueryClient();

  // Single combined API call instead of 8 separate calls
  // This significantly reduces page load time by eliminating multiple round trips
  const { data: combinedData, isLoading: isLoadingCombined } = useQuery<CombinedAnalyticsData>({
    queryKey: ['analytics-combined'],
    queryFn: () => analyticsService.getCombinedAnalytics({
      top_products_limit: 10,
      low_stock_threshold: 10,
    }),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  });

  // Extract individual data from combined response
  const dashboardData = combinedData?.dashboard;
  const salesTrendsData = combinedData?.salesTrends;
  const topProductsData = combinedData?.topProducts;
  const lowStockData = combinedData?.lowStock;
  const orderStatusData = combinedData?.orderStatus;
  const inventoryValuationData = combinedData?.inventoryValuation;

  // Single loading state for all analytics data
  const dashboardLoading = isLoadingCombined;
  const salesTrendLoading = isLoadingCombined;
  const lowStockLoading = isLoadingCombined;
  const inventoryValuationLoading = isLoadingCombined;

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
      // Only need to invalidate the combined query now
      await queryClient.invalidateQueries({ queryKey: ['analytics-combined'] });
    } catch (error) {
      console.error('Failed to refresh analytics:', error);
    } finally {
      setRefreshing(false);
    }
  };

  const topProducts: ProductSales[] = useMemo(() => topProductsData ?? [], [topProductsData]);
  const lowStockItems: LowStockItem[] = useMemo(() => lowStockData ?? [], [lowStockData]);
  const orderStatus: OrderStatusEntry[] = useMemo(() => orderStatusData ?? [], [orderStatusData]);

  const handleExport = async (type: 'csv' | 'pdf', report: 'sales' | 'inventory' | 'low-stock') => {
    try {
      const blob = await analyticsService.exportAnalytics({ type, report });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `${report}_report_${new Date().toISOString().split('T')[0]}.${type}`);
      document.body.appendChild(link);
      link.click();
      link.parentNode?.removeChild(link);
      toast.success(`${report} report exported successfully`);
    } catch (error) {
      console.error('Export failed:', error);
      toast.error('Failed to export report');
    }
  };

  return (
    <div className="space-y-8 p-6">
      <div className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tighter text-foreground uppercase">Analytics & Reports</h1>
          <p className="text-xs font-mono text-muted-foreground mt-1 uppercase tracking-widest">BUSINESS INTELLIGENCE</p>
        </div>
        <div className="flex items-center gap-2">
          <DropdownMenu
            trigger={
              <Button variant="outline" className="rounded-none font-mono uppercase tracking-wider">
                <Download className="h-4 w-4 mr-2" />
                EXPORT
              </Button>
            }
          >
            <DropdownMenuItem onSelect={() => handleExport('csv', 'sales')}>
              Sales Report (CSV)
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => handleExport('pdf', 'sales')}>
              Sales Report (PDF)
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => handleExport('csv', 'inventory')}>
              Inventory Valuation (CSV)
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => handleExport('pdf', 'inventory')}>
              Inventory Valuation (PDF)
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => handleExport('csv', 'low-stock')}>
              Low Stock Report (CSV)
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => handleExport('pdf', 'low-stock')}>
              Low Stock Report (PDF)
            </DropdownMenuItem>
          </DropdownMenu>
          <Button
            onClick={handleRefresh}
            disabled={refreshing}
            className="rounded-none font-mono uppercase tracking-wider"
            variant="outline"
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            REFRESH DATA
          </Button>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((card, index) => (
          <Card key={index} className={`rounded-none border-l-4 ${index === 0 ? 'border-l-primary' :
            index === 1 ? 'border-l-purple-500' :
              index === 2 ? 'border-l-emerald-500' :
                'border-l-amber-500'
            }`}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">
                {card.title}
              </CardTitle>
              <card.icon className={`h-4 w-4 ${index === 0 ? 'text-primary' :
                index === 1 ? 'text-purple-500' :
                  index === 2 ? 'text-emerald-500' :
                    'text-amber-500'
                }`} />
            </CardHeader>
            <CardContent>
              {card.isLoading ? (
                <div className="h-8 w-24 animate-pulse bg-muted" />
              ) : (
                <>
                  <div className="text-3xl font-bold font-mono">{card.value}</div>
                  {card.change !== 'N/A' && (
                    <p className="text-xs text-muted-foreground font-mono mt-1">
                      {card.change} {card.changeLabel}
                    </p>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4 rounded-none border border-border">
          <CardHeader className="border-b border-border bg-muted/10">
            <CardTitle className="flex items-center gap-2 text-sm font-mono uppercase tracking-wider">
              <LineChartIcon className="h-4 w-4 text-primary" />
              Sales Trends
            </CardTitle>
          </CardHeader>
          <CardContent className="pl-2 pt-6">
            {salesTrendLoading ? (
              <div className="flex h-[350px] items-center justify-center">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              </div>
            ) : (
              <div className="h-[350px] w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={salesTrendChartData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" vertical={false} />
                    <XAxis
                      dataKey="label"
                      stroke="hsl(var(--muted-foreground))"
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                      fontFamily="var(--font-mono)"
                    />
                    <YAxis
                      stroke="hsl(var(--muted-foreground))"
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                      tickFormatter={(value) => `$${value}`}
                      fontFamily="var(--font-mono)"
                    />
                    <ChartTooltip
                      content={<ChartTooltipContent className="rounded-none border-border font-mono uppercase" />}
                    />
                    <Line
                      type="monotone"
                      dataKey="revenue"
                      stroke="hsl(var(--primary))"
                      strokeWidth={2}
                      dot={{ r: 4, fill: "hsl(var(--primary))" }}
                      activeDot={{ r: 6, strokeWidth: 0 }}
                    />
                    <Line
                      type="monotone"
                      dataKey="orders"
                      stroke="#a855f7"
                      strokeWidth={2}
                      dot={{ r: 4, fill: "#a855f7" }}
                      activeDot={{ r: 6, strokeWidth: 0 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="col-span-3 rounded-none border border-border">
          <CardHeader className="border-b border-border bg-muted/10">
            <CardTitle className="flex items-center gap-2 text-sm font-mono uppercase tracking-wider">
              <PieChartIcon className="h-4 w-4 text-purple-500" />
              Order Status Distribution
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="h-[350px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={orderStatus}
                    dataKey="count"
                    nameKey="status"
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={2}
                  >
                    {orderStatus.map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={
                          index === 0 ? 'hsl(var(--primary))' :
                            index === 1 ? '#a855f7' :
                              index === 2 ? '#10b981' :
                                index === 3 ? '#f59e0b' :
                                  '#ef4444'
                        }
                        stroke="hsl(var(--background))"
                        strokeWidth={2}
                      />
                    ))}
                  </Pie>
                  <ChartTooltip
                    content={<ChartTooltipContent className="rounded-none border-border font-mono uppercase" />}
                  />
                  <ChartLegend
                    content={<ChartLegendContent className="font-mono text-xs uppercase" />}
                    verticalAlign="bottom"
                    height={36}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="rounded-none border border-border">
          <CardHeader className="border-b border-border bg-muted/10">
            <CardTitle className="flex items-center gap-2 text-sm font-mono uppercase tracking-wider">
              <BarChart3 className="h-4 w-4 text-emerald-500" />
              Top Products
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="h-[350px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={topProducts} layout="vertical" margin={{ left: 40 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" horizontal={false} />
                  <XAxis type="number" hide />
                  <YAxis
                    dataKey="productName"
                    type="category"
                    width={100}
                    tick={{ fontSize: 10, fontFamily: 'var(--font-mono)', fill: 'hsl(var(--muted-foreground))' }}
                    interval={0}
                  />
                  <ChartTooltip
                    content={<ChartTooltipContent className="rounded-none border-border font-mono uppercase" />}
                  />
                  <Bar
                    dataKey="unitsSold"
                    fill="hsl(var(--primary))"
                    radius={[0, 4, 4, 0]}
                    barSize={20}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-none border border-border">
          <CardHeader className="border-b border-border bg-muted/10">
            <CardTitle className="flex items-center gap-2 text-sm font-mono uppercase tracking-wider">
              <AlertTriangle className="h-4 w-4 text-amber-500" />
              Low Stock Alerts
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-border">
              {lowStockItems.length === 0 ? (
                <div className="p-6 text-center text-muted-foreground font-mono text-sm">
                  NO LOW STOCK ITEMS
                </div>
              ) : (
                lowStockItems.slice(0, 5).map((item) => (
                  <div key={item.productId} className="flex items-center justify-between p-4 hover:bg-muted/20 transition-colors">
                    <div>
                      <p className="font-medium font-mono text-sm">{item.productName}</p>
                      <p className="text-xs text-muted-foreground font-mono uppercase mt-1">
                        WAREHOUSE: {item.warehouseId || 'N/A'}
                      </p>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-bold font-mono text-destructive">
                        {item.currentStock} LEFT
                      </div>
                      <p className="text-[10px] text-muted-foreground font-mono uppercase">
                        MIN: {item.threshold}
                      </p>
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

// Helper functions
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
