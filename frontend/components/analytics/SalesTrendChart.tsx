'use client'

import React, { useState } from 'react'
import { Calendar, Download, TrendingUp, TrendingDown } from 'lucide-react'
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { analyticsService } from '@/lib/services'

type TimePeriod = '7d' | '30d' | '90d' | '1y' | 'custom'
type ChartType = 'line' | 'area'

interface SalesTrendChartProps {
  className?: string
}

export default function SalesTrendChart({ className = '' }: SalesTrendChartProps) {
  const [timePeriod, setTimePeriod] = useState<TimePeriod>('30d')
  const [chartType, setChartType] = useState<ChartType>('area')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  // Fetch sales trends via analyticsService
  const { data: salesTrends, isLoading } = useQuery({
    queryKey: ['sales-trend', timePeriod, startDate, endDate],
    queryFn: async () => {
      // Map timePeriod to date range (default to last 30 days)
      const now = new Date()
      let params: { start_date?: string; end_date?: string } = {}
      const toISO = (d: Date) => d.toISOString().slice(0, 10)
      switch (timePeriod) {
        case '7d':
          params = { start_date: toISO(new Date(now.getTime() - 7 * 24 * 3600 * 1000)), end_date: toISO(now) }
          break
        case '30d':
          params = { start_date: toISO(new Date(now.getTime() - 30 * 24 * 3600 * 1000)), end_date: toISO(now) }
          break
        case '90d':
          params = { start_date: toISO(new Date(now.getTime() - 90 * 24 * 3600 * 1000)), end_date: toISO(now) }
          break
        case '1y':
          params = { start_date: toISO(new Date(now.getTime() - 365 * 24 * 3600 * 1000)), end_date: toISO(now) }
          break
        case 'custom':
          if (startDate && endDate) params = { start_date: startDate, end_date: endDate }
          break
      }
      return analyticsService.getSalesTrends(params)
    }
  })

  const salesData = (salesTrends?.trends ?? []).map(pt => ({
    date: pt.date,
    sales: pt.salesAmount,
    orders: pt.orderCount,
  }))

  // Calculate statistics
  const totalSales = salesData.reduce((sum: number, item: any) => sum + (item.sales || 0), 0)
  const avgSales = salesData.length > 0 ? totalSales / salesData.length : 0
  const maxSales = Math.max(...salesData.map((item: any) => item.sales || 0))
  const minSales = Math.min(...salesData.map((item: any) => item.sales || 0))

  // Calculate trend
  const firstHalf = salesData.slice(0, Math.floor(salesData.length / 2))
  const secondHalf = salesData.slice(Math.floor(salesData.length / 2))
  const firstHalfAvg = firstHalf.reduce((sum: number, item: any) => sum + (item.sales || 0), 0) / (firstHalf.length || 1)
  const secondHalfAvg = secondHalf.reduce((sum: number, item: any) => sum + (item.sales || 0), 0) / (secondHalf.length || 1)
  const trendPercentage = firstHalfAvg > 0 ? ((secondHalfAvg - firstHalfAvg) / firstHalfAvg) * 100 : 0
  const isPositiveTrend = trendPercentage >= 0

  const handleExport = () => {
    // Convert data to CSV
    const headers = ['Date', 'Sales', 'Orders', 'Average Order Value']
    const rows = salesData.map((item: any) => [
      item.date || item.month,
      item.sales || 0,
      item.orders || 0,
      item.avg_order_value || 0
    ])
    
    const csv = [
      headers.join(','),
      ...rows.map((row: any[]) => row.join(','))
    ].join('\n')

    // Download CSV
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `sales-trend-${timePeriod}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value)
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-popover text-popover-foreground p-3 rounded-none border border-border shadow-none">
          <p className="font-mono text-xs mb-2 text-muted-foreground uppercase tracking-wider">{label}</p>
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center justify-between space-x-4">
              <span className="text-sm font-mono text-muted-foreground">{entry.name}:</span>
              <span className="text-sm font-mono font-bold" style={{ color: entry.color }}>
                {entry.name.includes('Sales') || entry.name.includes('Value') 
                  ? formatCurrency(entry.value)
                  : entry.value.toLocaleString()}
              </span>
            </div>
          ))}
        </div>
      )
    }
    return null
  }

  return (
    <div className={`bg-card text-card-foreground rounded-none border border-border ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
          <div>
            <h2 className="text-xl font-bold font-mono uppercase tracking-widest">Sales Trend</h2>
            <p className="text-xs text-muted-foreground mt-1 font-mono">FLOW ANALYSIS</p>
          </div>

          <div className="flex items-center space-x-2 mt-4 md:mt-0">
            {/* Time Period Selector */}
            <div className="flex items-center space-x-1 bg-muted/20 p-1">
              {[
                { value: '7d', label: '7D' },
                { value: '30d', label: '30D' },
                { value: '90d', label: '90D' },
                { value: '1y', label: '1Y' },
                { value: 'custom', label: 'CUSTOM' },
              ].map((period) => (
                <button
                  key={period.value}
                  onClick={() => setTimePeriod(period.value as TimePeriod)}
                  className={`px-3 py-1 text-xs font-mono transition-colors ${
                    timePeriod === period.value
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {period.label}
                </button>
              ))}
            </div>

            {/* Export Button */}
            <button
              onClick={handleExport}
              className="flex items-center space-x-2 px-4 py-2 text-foreground bg-muted/20 hover:bg-muted/40 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span className="hidden md:inline font-mono text-xs uppercase">Export</span>
            </button>
          </div>
        </div>

        {/* Custom Date Range */}
        {timePeriod === 'custom' && (
          <div className="flex items-center space-x-4 mb-6 p-4 bg-muted/10 border border-border">
            <Calendar className="w-5 h-5 text-muted-foreground" />
            <div className="flex items-center space-x-2">
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="px-3 py-2 border border-input bg-background text-foreground focus:ring-1 focus:ring-primary focus:border-primary font-mono text-sm"
              />
              <span className="text-muted-foreground font-mono text-xs">TO</span>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="px-3 py-2 border border-input bg-background text-foreground focus:ring-1 focus:ring-primary focus:border-primary font-mono text-sm"
              />
            </div>
          </div>
        )}

        {/* Statistics Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6 border-b border-border pb-6">
          <div className="p-4 border-l-2 border-primary/20">
            <p className="text-xs text-muted-foreground font-mono uppercase">Total Sales</p>
            <p className="text-xl font-bold font-mono mt-1 text-primary">{formatCurrency(totalSales)}</p>
          </div>
          
          <div className="p-4 border-l-2 border-primary/20">
            <p className="text-xs text-muted-foreground font-mono uppercase">Average Sales</p>
            <p className="text-xl font-bold font-mono mt-1">{formatCurrency(avgSales)}</p>
          </div>
          
          <div className="p-4 border-l-2 border-primary/20">
            <p className="text-xs text-muted-foreground font-mono uppercase">Peak Sales</p>
            <p className="text-xl font-bold font-mono mt-1">{formatCurrency(maxSales)}</p>
          </div>
          
          <div className="p-4 border-l-2 border-primary/20">
            <p className="text-xs text-muted-foreground font-mono uppercase">Trend</p>
            <div className="flex items-center mt-1">
              {isPositiveTrend ? (
                <TrendingUp className="w-4 h-4 text-primary mr-2" />
              ) : (
                <TrendingDown className="w-4 h-4 text-destructive mr-2" />
              )}
              <p className={`text-xl font-bold font-mono ${isPositiveTrend ? 'text-primary' : 'text-destructive'}`}>
                {Math.abs(trendPercentage).toFixed(1)}%
              </p>
            </div>
          </div>
        </div>

        {/* Chart */}
        <div className="mt-6">
          {isLoading ? (
            <div className="flex items-center justify-center h-96">
              <div className="text-muted-foreground font-mono animate-pulse">LOADING DATA STREAM...</div>
            </div>
          ) : salesData.length === 0 ? (
            <div className="flex items-center justify-center h-96">
              <div className="text-center">
                <p className="text-muted-foreground font-mono">NO DATA SIGNAL</p>
                <p className="text-xs text-muted-foreground/50 mt-2 font-mono">ADJUST TIME RANGE</p>
              </div>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={400}>
              <AreaChart data={salesData}>
                <defs>
                  <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="colorOrders" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--secondary))" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="hsl(var(--secondary))" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" opacity={0.2} vertical={false} />
                <XAxis 
                  dataKey="date" 
                  stroke="hsl(var(--muted-foreground))"
                  style={{ fontSize: '10px', fontFamily: 'var(--font-jetbrains-mono)' }}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis 
                  stroke="hsl(var(--muted-foreground))"
                  style={{ fontSize: '10px', fontFamily: 'var(--font-jetbrains-mono)' }}
                  tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend wrapperStyle={{ fontFamily: 'var(--font-jetbrains-mono)', fontSize: '12px' }} />
                <Area 
                  type="monotone" 
                  dataKey="sales" 
                  stroke="hsl(var(--primary))" 
                  strokeWidth={2}
                  fillOpacity={1} 
                  fill="url(#colorSales)" 
                  name="Sales Amount"
                  animationDuration={2000}
                />
                <Area 
                  type="monotone" 
                  dataKey="orders" 
                  stroke="hsl(var(--secondary))" 
                  strokeWidth={2}
                  fillOpacity={1} 
                  fill="url(#colorOrders)" 
                  name="Order Count"
                  animationDuration={2000}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>
    </div>
  )
}