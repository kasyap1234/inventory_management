'use client';

import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface InventoryTurnoverData {
  product_name: string;
  turnover_ratio: number;
  days_in_stock: number;
  stock_level: number;
}

interface InventoryTurnoverChartProps {
  data: InventoryTurnoverData[];
  isLoading?: boolean;
}

export function InventoryTurnoverChart({ data, isLoading }: InventoryTurnoverChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Inventory Turnover</CardTitle>
          <CardDescription>Loading turnover metrics...</CardDescription>
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
        <CardTitle>Inventory Turnover</CardTitle>
        <CardDescription>Turnover ratio vs days in stock</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <ScatterChart margin={{ top: 20, right: 20, bottom: 20, left: 20 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="days_in_stock" name="Days in Stock" />
            <YAxis dataKey="turnover_ratio" name="Turnover Ratio" />
            <Tooltip cursor={{ strokeDasharray: '3 3' }} />
            <Legend />
            <Scatter name="Products" data={data} fill="#8884d8" />
          </ScatterChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
