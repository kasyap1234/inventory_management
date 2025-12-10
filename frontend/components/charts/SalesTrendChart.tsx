'use client';

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { analyticsService } from '@/lib/services';
import { formatCurrency } from '@/lib/utils';
import { format, subDays } from 'date-fns';

interface SalesData {
  date: string;
  sales: number;
  orders: number;
}

export function SalesTrendChart() {
  const rangeEnd = useMemo(() => new Date(), []);
  const rangeStart = useMemo(() => subDays(rangeEnd, 30), [rangeEnd]);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['analytics', 'sales-trend', rangeStart.toISOString(), rangeEnd.toISOString()],
    queryFn: async () =>
      analyticsService.getSalesTrends({
        start_date: format(rangeStart, 'yyyy-MM-dd'),
        end_date: format(rangeEnd, 'yyyy-MM-dd'),
      }),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const chartData: SalesData[] = useMemo(() => {
    const trendMap = new Map<string, { sales: number; orders: number }>();

    data?.trends?.forEach((trend) => {
      if (!trend.date) return;
      const key = format(new Date(trend.date), 'yyyy-MM-dd');
      trendMap.set(key, {
        sales: trend.salesAmount ?? 0,
        orders: trend.orderCount ?? 0,
      });
    });

    const series: SalesData[] = [];
    for (let i = 30; i >= 0; i--) {
      const day = subDays(rangeEnd, i);
      const key = format(day, 'yyyy-MM-dd');
      const label = format(day, 'MMM dd');
      const point = trendMap.get(key);
      series.push({
        date: label,
        sales: point?.sales ?? 0,
        orders: point?.orders ?? 0,
      });
    }

    return series;
  }, [data?.trends, rangeEnd]);

  if (isLoading) {
    return (
      <div className="h-[350px] flex items-center justify-center text-muted-foreground">
        Loading chart...
      </div>
    );
  }

  if (isError || !chartData.length) {
    return (
      <div className="h-[350px] flex items-center justify-center text-muted-foreground">
        {isError ? 'Failed to load sales data' : 'No sales data available'}
      </div>
    );
  }

  const hasValues = chartData.some((point) => point.sales > 0 || point.orders > 0);

  return (
    <ResponsiveContainer width="100%" height={350}>
      <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
            <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
        <XAxis
          dataKey="date"
          className="text-xs"
          tick={{ fill: 'currentColor' }}
          tickLine={{ stroke: 'currentColor' }}
        />
        <YAxis
          className="text-xs"
          tick={{ fill: 'currentColor' }}
          tickLine={{ stroke: 'currentColor' }}
          tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
          allowDecimals={false}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '8px',
          }}
          formatter={(value: number, name: string) => [
            name === 'sales' ? formatCurrency(value) : value,
            name === 'sales' ? 'Sales' : 'Orders',
          ]}
        />
        <Legend />
        <Area
          type="monotone"
          dataKey="sales"
          stroke="#8884d8"
          fillOpacity={1}
          fill="url(#colorSales)"
          name="Sales"
          isAnimationActive={hasValues}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
