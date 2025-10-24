'use client';

import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface ProductPerformanceData {
  product_name: string;
  units_sold: number;
  revenue: number;
  growth_rate: number;
  margin_percent: number;
}

interface ProductPerformanceChartProps {
  data: ProductPerformanceData[];
  isLoading?: boolean;
}

export function ProductPerformanceChart({ data, isLoading }: ProductPerformanceChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Product Performance</CardTitle>
          <CardDescription>Loading product metrics...</CardDescription>
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
        <CardTitle>Product Performance</CardTitle>
        <CardDescription>Units sold and revenue by product</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="product_name" angle={-45} textAnchor="end" height={80} />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="units_sold" stroke="#8884d8" name="Units Sold" />
            <Line type="monotone" dataKey="revenue" stroke="#82ca9d" name="Revenue" />
          </LineChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
