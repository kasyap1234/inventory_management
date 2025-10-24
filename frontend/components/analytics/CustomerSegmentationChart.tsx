'use client';

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface CustomerSegmentationData {
  segment: string;
  customer_count: number;
  average_order_value: number;
  total_revenue: number;
}

interface CustomerSegmentationChartProps {
  data: CustomerSegmentationData[];
  isLoading?: boolean;
}

export function CustomerSegmentationChart({ data, isLoading }: CustomerSegmentationChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Customer Segmentation</CardTitle>
          <CardDescription>Loading segmentation data...</CardDescription>
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
        <CardTitle>Customer Segmentation</CardTitle>
        <CardDescription>Revenue and order distribution by customer segment</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="segment" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="customer_count" fill="#8884d8" name="Customer Count" />
            <Bar dataKey="total_revenue" fill="#82ca9d" name="Total Revenue" />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
