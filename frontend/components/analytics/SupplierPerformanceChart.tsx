'use client';

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface SupplierPerformanceData {
  supplier_name: string;
  order_count: number;
  on_time_delivery: number;
  quality_score: number;
  total_spent: number;
}

interface SupplierPerformanceChartProps {
  data: SupplierPerformanceData[];
  isLoading?: boolean;
}

export function SupplierPerformanceChart({ data, isLoading }: SupplierPerformanceChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Supplier Performance</CardTitle>
          <CardDescription>Loading supplier metrics...</CardDescription>
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
        <CardTitle>Supplier Performance</CardTitle>
        <CardDescription>On-time delivery and quality scores by supplier</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="supplier_name" angle={-45} textAnchor="end" height={80} />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="on_time_delivery" fill="#8884d8" name="On-Time Delivery %" />
            <Bar dataKey="quality_score" fill="#82ca9d" name="Quality Score" />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
