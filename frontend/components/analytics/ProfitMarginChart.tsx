'use client';

import { PieChart, Pie, Cell, Legend, Tooltip, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface ProfitMarginData {
  total_revenue: number;
  total_cost: number;
  profit_margin: number;
  margin_trend: string;
  invoice_count: number;
}

interface ProfitMarginChartProps {
  data: ProfitMarginData;
  isLoading?: boolean;
}

const COLORS = ['#8884d8', '#82ca9d'];

export function ProfitMarginChart({ data, isLoading }: ProfitMarginChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Profit Margin Analysis</CardTitle>
          <CardDescription>Loading margin data...</CardDescription>
        </CardHeader>
        <CardContent className="h-80 flex items-center justify-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  const chartData = [
    { name: 'Cost', value: data.total_cost },
    { name: 'Profit', value: data.total_revenue - data.total_cost },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profit Margin Analysis</CardTitle>
        <CardDescription>Revenue breakdown and margin percentage</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={chartData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, value }) => `${name}: $${(value as number).toFixed(2)}`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {chartData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip formatter={(value) => `$${(value as number).toFixed(2)}`} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-gray-600">Total Revenue</p>
              <p className="text-lg font-semibold">${data.total_revenue.toFixed(2)}</p>
            </div>
            <div>
              <p className="text-gray-600">Profit Margin</p>
              <p className="text-lg font-semibold">{data.profit_margin.toFixed(2)}%</p>
            </div>
            <div>
              <p className="text-gray-600">Total Cost</p>
              <p className="text-lg font-semibold">${data.total_cost.toFixed(2)}</p>
            </div>
            <div>
              <p className="text-gray-600">Trend</p>
              <p className="text-lg font-semibold capitalize">{data.margin_trend}</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
