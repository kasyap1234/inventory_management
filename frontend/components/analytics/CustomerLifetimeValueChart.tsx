'use client';

import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface CustomerLTVData {
  customer_name: string;
  total_spent: number;
  order_count: number;
  average_order_value: number;
}

interface CustomerLifetimeValueChartProps {
  data: CustomerLTVData[];
  isLoading?: boolean;
}

export function CustomerLifetimeValueChart({ data, isLoading }: CustomerLifetimeValueChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Customer Lifetime Value</CardTitle>
          <CardDescription>Loading customer LTV data...</CardDescription>
        </CardHeader>
        <CardContent className="h-80 flex items-center justify-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Customer Lifetime Value</CardTitle>
        <CardDescription>Total spending and order frequency by customer</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="customer_name" angle={-45} textAnchor="end" height={80} />
            <YAxis />
            <Tooltip />
            <Legend />
            <Area type="monotone" dataKey="total_spent" fill="#8884d8" stroke="#8884d8" name="Total Spent" />
            <Area type="monotone" dataKey="average_order_value" fill="#82ca9d" stroke="#82ca9d" name="Avg Order Value" />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
