'use client'

import React, { useState } from 'react'
import { DollarSign, TrendingUp, TrendingDown, PieChart as PieChartIcon, BarChart3 } from 'lucide-react'
import { BarChart, Bar, LineChart, Line, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { analyticsService } from '@/lib/services'

type ViewType = 'category' | 'product' | 'customer' | 'region'

interface RevenueAnalyticsProps {
  className?: string
}

export default function RevenueAnalytics({ className = '' }: RevenueAnalyticsProps) {
  const [viewType, setViewType] = useState<ViewType>('category')
  const [timePeriod, setTimePeriod] = useState<'7d' | '30d' | '90d' | '1y'>('30d')

  // Fetch revenue data
  const { data: revenueData, isLoading } = useQuery({
    queryKey: ['revenue-analytics', viewType, timePeriod],
    queryFn: async () => {
      // derive date range
      const now = new Date()
      const end_date = now.toISOString().slice(0, 10)
      const start = new Date(now)
      if (timePeriod === '7d') start.setDate(now.getDate() - 7)
      if (timePeriod === '30d') start.setDate(now.getDate() - 30)
      if (timePeriod === '90d') start.setDate(now.getDate() - 90)
      if (timePeriod === '1y') start.setFullYear(now.getFullYear() - 1)
      const start_date = start.toISOString().slice(0, 10)

      const [byCategory, salesTrends] = await Promise.all([
        analyticsService.getRevenueByCategory({ start_date, end_date }),
        analyticsService.getSalesTrends({ start_date, end_date }),
      ])

      const breakdown = byCategory.map((entry) => ({
        name: entry.categoryId,
        revenue: entry.totalRevenue,
        orders: undefined as number | undefined,
      }))

      const trend = salesTrends.trends.map((t) => ({
        date: t.date,
        revenue: t.salesAmount,
        profit: 0,
      }))

      const totalRevenue = breakdown.reduce((sum, b) => sum + (b.revenue || 0), 0)

      // compute revenue change vs previous half-period
      let revenue_change = 0
      if (trend.length > 1) {
        const mid = Math.floor(trend.length / 2)
        const first = trend.slice(0, mid).reduce((s, p) => s + (p.revenue || 0), 0)
        const second = trend.slice(mid).reduce((s, p) => s + (p.revenue || 0), 0)
        if (first > 0) revenue_change = ((second - first) / first) * 100
      }

      const top = breakdown.reduce((acc, cur) => (cur.revenue > (acc?.revenue || 0) ? cur : acc), undefined as any)

      const summary = {
        total_revenue: totalRevenue,
        revenue_change: revenue_change,
        gross_profit: 0,
        profit_margin: 0,
        avg_order_value: 0,
        top_category: top?.name || 'N/A',
      }

      return { summary, breakdown, trend }
    },
  })

  const summary = revenueData?.summary || {
    total_revenue: 0,
    revenue_change: 0,
    gross_profit: 0,
    profit_margin: 0,
    avg_order_value: 0,
    top_category: 'N/A'
  }

  const breakdown = revenueData?.breakdown || []
  const trend = revenueData?.trend || []

  const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#14B8A6', '#F97316']

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value)
  }

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-popover text-popover-foreground p-4 rounded-lg shadow-lg border border-border">
          <p className="font-semibold mb-2">{payload[0].payload.name}</p>
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center justify-between space-x-4">
              <span className="text-sm text-muted-foreground">{entry.name}:</span>
              <span className="text-sm font-semibold" style={{ color: entry.color }}>
                {formatCurrency(entry.value)}
              </span>
            </div>
          ))}
        </div>
      )
    }
    return null
  }

  return (
    <div className={`bg-card text-card-foreground rounded-lg shadow ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold">Revenue Analytics</h2>
            <p className="text-muted-foreground mt-1">Analyze revenue breakdown and trends</p>
          </div>

          <div className="flex items-center space-x-2 mt-4 md:mt-0">
            {/* Time Period Selector */}
            <div className="flex items-center space-x-1 bg-muted rounded-lg p-1">
              {[
                { value: '7d', label: '7D' },
                { value: '30d', label: '30D' },
                { value: '90d', label: '90D' },
                { value: '1y', label: '1Y' },
              ].map((period) => (
                <button
                  key={period.value}
                  onClick={() => setTimePeriod(period.value as any)}
                  className={`px-3 py-1 text-sm rounded-md transition-colors ${
                    timePeriod === period.value
                      ? 'bg-background text-primary shadow'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {period.label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div className="p-4 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium opacity-90">Total Revenue</p>
                <p className="text-2xl font-bold mt-1">{formatCurrency(summary.total_revenue)}</p>
                <div className="flex items-center mt-2 text-sm">
                  {summary.revenue_change >= 0 ? (
                    <TrendingUp className="w-4 h-4 mr-1" />
                  ) : (
                    <TrendingDown className="w-4 h-4 mr-1" />
                  )}
                  <span>{Math.abs(summary.revenue_change).toFixed(1)}% vs last period</span>
                </div>
              </div>
              <DollarSign className="w-12 h-12 opacity-80" />
            </div>
          </div>

          <div className="p-4 bg-gradient-to-br from-green-500 to-green-600 rounded-lg text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium opacity-90">Gross Profit</p>
                <p className="text-2xl font-bold mt-1">{formatCurrency(summary.gross_profit)}</p>
                <p className="text-sm mt-2 opacity-90">
                  Margin: {summary.profit_margin.toFixed(1)}%
                </p>
              </div>
              <TrendingUp className="w-12 h-12 opacity-80" />
            </div>
          </div>

          <div className="p-4 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium opacity-90">Avg Order Value</p>
                <p className="text-2xl font-bold mt-1">{formatCurrency(summary.avg_order_value)}</p>
                <p className="text-sm mt-2 opacity-90">Per transaction</p>
              </div>
              <BarChart3 className="w-12 h-12 opacity-80" />
            </div>
          </div>

          <div className="p-4 bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium opacity-90">Top Category</p>
                <p className="text-xl font-bold mt-1">{summary.top_category}</p>
                <p className="text-sm mt-2 opacity-90">Highest revenue</p>
              </div>
              <PieChartIcon className="w-12 h-12 opacity-80" />
            </div>
          </div>
        </div>

        {/* View Type Selector */}
        <div className="flex items-center space-x-2 mb-6">
          <span className="text-sm font-medium text-muted-foreground">Breakdown by:</span>
          <div className="flex items-center space-x-1 bg-muted rounded-lg p-1">
            {[
              { value: 'category', label: 'Category' },
              { value: 'product', label: 'Product' },
              { value: 'customer', label: 'Customer' },
              { value: 'region', label: 'Region' },
            ].map((view) => (
              <button
                key={view.value}
                onClick={() => setViewType(view.value as ViewType)}
                className={`px-3 py-1 text-sm rounded-md transition-colors ${
                  viewType === view.value
                    ? 'bg-background text-primary shadow'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {view.label}
              </button>
            ))}
          </div>
        </div>

        {/* Charts */}
        {isLoading ? (
          <div className="flex items-center justify-center h-96">
            <div className="text-muted-foreground">Loading revenue data...</div>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Revenue Breakdown - Pie Chart */}
            <div className="p-4 border border-border rounded-lg">
              <h3 className="text-lg font-semibold mb-4">
                Revenue Distribution by {viewType.charAt(0).toUpperCase() + viewType.slice(1)}
              </h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={breakdown}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }: any) => `${name}: ${((percent as number) * 100).toFixed(0)}%`}
                    outerRadius={100}
                    fill="#8884d8"
                    dataKey="revenue"
                  >
                    {breakdown.map((entry: any, index: number) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip content={<CustomTooltip />} />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* Revenue Breakdown - Bar Chart */}
            <div className="p-4 border border-border rounded-lg">
              <h3 className="text-lg font-semibold mb-4">
                Top 10 by Revenue
              </h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={breakdown.slice(0, 10)} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis type="number" tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`} />
                  <YAxis dataKey="name" type="category" width={100} style={{ fontSize: '12px' }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Bar dataKey="revenue" fill="#3B82F6" name="Revenue" />
                </BarChart>
              </ResponsiveContainer>
            </div>

            {/* Revenue Trend */}
            <div className="p-4 border border-border rounded-lg lg:col-span-2">
              <h3 className="text-lg font-semibold mb-4">Revenue Trend</h3>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={trend}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" style={{ fontSize: '12px' }} />
                  <YAxis tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`} style={{ fontSize: '12px' }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <Line 
                    type="monotone" 
                    dataKey="revenue" 
                    stroke="#3B82F6" 
                    strokeWidth={3}
                    dot={{ fill: '#3B82F6', r: 4 }}
                    name="Revenue"
                  />
                  <Line 
                    type="monotone" 
                    dataKey="profit" 
                    stroke="#10B981" 
                    strokeWidth={3}
                    dot={{ fill: '#10B981', r: 4 }}
                    name="Profit"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        )}

        {/* Detailed Breakdown Table */}
        <div className="mt-6">
          <h3 className="text-lg font-semibold mb-4">Detailed Breakdown</h3>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    {viewType.charAt(0).toUpperCase() + viewType.slice(1)}
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Revenue
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Orders
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Avg Order Value
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    % of Total
                  </th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {breakdown.map((item: any, index: number) => {
                  const percentage = (item.revenue / summary.total_revenue) * 100
                  return (
                    <tr key={index} className="hover:bg-muted/50">
                      <td className="px-4 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div 
                            className="w-3 h-3 rounded-full mr-3" 
                            style={{ backgroundColor: COLORS[index % COLORS.length] }}
                          />
                          <span className="text-sm font-medium">{item.name}</span>
                        </div>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap text-right">
                        <span className="text-sm">{formatCurrency(item.revenue)}</span>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap text-right">
                        <span className="text-sm">{item.orders?.toLocaleString() || 'N/A'}</span>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap text-right">
                        <span className="text-sm">
                          {item.orders ? formatCurrency(item.revenue / item.orders) : 'N/A'}
                        </span>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap text-right">
                        <div className="flex items-center justify-end">
                          <div className="w-16 bg-muted rounded-full h-2 mr-2">
                            <div 
                              className="bg-blue-600 h-2 rounded-full" 
                              style={{ width: `${percentage}%` }}
                            />
                          </div>
                          <span className="text-sm">{percentage.toFixed(1)}%</span>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>

        {/* Insights */}
        <div className="mt-6 p-4 bg-blue-50 rounded-lg">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">Key Insights</h3>
          <ul className="space-y-2 text-sm text-blue-800">
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Revenue is {summary.revenue_change >= 0 ? 'up' : 'down'} {Math.abs(summary.revenue_change).toFixed(1)}% compared to the previous period
              </span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                {summary.top_category} is your top-performing category
              </span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Profit margin is at {summary.profit_margin.toFixed(1)}%
              </span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Average order value is {formatCurrency(summary.avg_order_value)}
              </span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  )
}