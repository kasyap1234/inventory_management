'use client';

import { useQuery } from '@tanstack/react-query';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import api from '@/lib/api';
import { formatCurrency } from '@/lib/utils';
import { format, subDays } from 'date-fns';

interface SalesData {
    date: string;
    sales: number;
    orders: number;
}

export function SalesTrendChart() {
    const { data, isLoading } = useQuery({
        queryKey: ['sales-trend'],
        queryFn: async () => {
            // Get sales data for the last 30 days
            const endDate = new Date();
            const startDate = subDays(endDate, 30);

            const response = await api.get(`/orders?start_date=${startDate.toISOString()}&end_date=${endDate.toISOString()}`);
            const orders = response.data.orders || [];

            // Group orders by date
            const salesByDate: Record<string, { sales: number; orders: number }> = {};

            orders.forEach((order: { created_at: string; total_amount: number }) => {
                const date = format(new Date(order.created_at), 'MMM dd');
                if (!salesByDate[date]) {
                    salesByDate[date] = { sales: 0, orders: 0 };
                }
                salesByDate[date].sales += order.total_amount || 0;
                salesByDate[date].orders += 1;
            });

            // Convert to array and sort by date
            const salesData: SalesData[] = Object.entries(salesByDate).map(([date, data]) => ({
                date,
                sales: data.sales,
                orders: data.orders,
            }));

            // Fill in missing dates with zero values
            const fullData: SalesData[] = [];
            for (let i = 30; i >= 0; i--) {
                const date = format(subDays(endDate, i), 'MMM dd');
                const existing = salesData.find(d => d.date === date);
                fullData.push({
                    date,
                    sales: existing?.sales || 0,
                    orders: existing?.orders || 0,
                });
            }

            return fullData;
        },
        staleTime: 5 * 60 * 1000, // 5 minutes
    });

    if (isLoading) {
        return (
            <div className="h-[350px] flex items-center justify-center text-muted-foreground">
                Loading chart...
            </div>
        );
    }

    if (!data || data.length === 0) {
        return (
            <div className="h-[350px] flex items-center justify-center text-muted-foreground">
                No sales data available
            </div>
        );
    }

    return (
        <ResponsiveContainer width="100%" height={350}>
            <AreaChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
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
                />
                <Tooltip
                    contentStyle={{
                        backgroundColor: 'hsl(var(--card))',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '8px',
                    }}
                    formatter={(value: number, name: string) => [
                        name === 'sales' ? formatCurrency(value) : value,
                        name === 'sales' ? 'Sales' : 'Orders'
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
                />
            </AreaChart>
        </ResponsiveContainer>
    );
}

export function OrdersLineChart() {
    const { data, isLoading } = useQuery({
        queryKey: ['orders-trend'],
        queryFn: async () => {
            const endDate = new Date();
            const startDate = subDays(endDate, 14);

            const response = await api.get(`/orders?start_date=${startDate.toISOString()}&end_date=${endDate.toISOString()}`);
            const orders = response.data.orders || [];

            const ordersByDate: Record<string, number> = {};

            orders.forEach((order: { created_at: string }) => {
                const date = format(new Date(order.created_at), 'MMM dd');
                ordersByDate[date] = (ordersByDate[date] || 0) + 1;
            });

            const chartData = [];
            for (let i = 14; i >= 0; i--) {
                const date = format(subDays(endDate, i), 'MMM dd');
                chartData.push({
                    date,
                    orders: ordersByDate[date] || 0,
                });
            }

            return chartData;
        },
        staleTime: 5 * 60 * 1000,
    });

    if (isLoading) {
        return (
            <div className="h-[200px] flex items-center justify-center text-muted-foreground">
                Loading...
            </div>
        );
    }

    return (
        <ResponsiveContainer width="100%" height={200}>
            <LineChart data={data}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                    dataKey="date"
                    className="text-xs"
                    tick={{ fill: 'currentColor' }}
                />
                <YAxis
                    className="text-xs"
                    tick={{ fill: 'currentColor' }}
                />
                <Tooltip
                    contentStyle={{
                        backgroundColor: 'hsl(var(--card))',
                        border: '1px solid hsl(var(--border))',
                        borderRadius: '8px',
                    }}
                />
                <Line
                    type="monotone"
                    dataKey="orders"
                    stroke="hsl(var(--primary))"
                    strokeWidth={2}
                    dot={{ fill: 'hsl(var(--primary))', r: 4 }}
                    activeDot={{ r: 6 }}
                />
            </LineChart>
        </ResponsiveContainer>
    );
}
