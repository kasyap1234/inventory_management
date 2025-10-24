'use client';

import { ComposedChart, Bar, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface MarketTrendData {
  trend_name: string;
  trend_value: number;
  change_percent: number;
  forecast: number;
  confidence: number;
}

interface MarketTrendsChartProps {
  data: MarketTrendData[];
  isLoading?: boolean;
}

export function MarketTrendsChart({ data, isLoading }: MarketTrendsChartProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Market Trends</CardTitle>
          <CardDescription>Loading trend analysis...</CardDescription>
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
        <CardTitle>Market Trends</CardTitle>
        <CardDescription>Current trends and forecasts with confidence levels</CardDescription>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <ComposedChart data={data}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="trend_name" angle={-45} textAnchor="end" height={80} />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="trend_value" fill="#8884d8" name="Current Value" />
            <Line type="monotone" dataKey="forecast" stroke="#82ca9d" name="Forecast" strokeWidth={2} />
          </ComposedChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
