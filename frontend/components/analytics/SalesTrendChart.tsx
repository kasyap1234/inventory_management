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
        <div className="bg-white p-4 rounded-lg shadow-lg border border-gray-200">
          <p className="font-semibold text-gray-900 mb-2">{label}</p>
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center justify-between space-x-4">
              <span className="text-sm text-gray-600">{entry.name}:</span>
              <span className="text-sm font-semibold" style={{ color: entry.color }}>
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
    <div className={`bg-white rounded-lg shadow ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Sales Trend Analysis</h2>
            <p className="text-gray-600 mt-1">Track your sales performance over time</p>
          </div>

          <div className="flex items-center space-x-2 mt-4 md:mt-0">
            {/* Time Period Selector */}
            <div className="flex items-center space-x-1 bg-gray-100 rounded-lg p-1">
              {[
                { value: '7d', label: '7D' },
                { value: '30d', label: '30D' },
                { value: '90d', label: '90D' },
                { value: '1y', label: '1Y' },
                { value: 'custom', label: 'Custom' },
              ].map((period) => (
                <button
                  key={period.value}
                  onClick={() => setTimePeriod(period.value as TimePeriod)}
                  className={`px-3 py-1 text-sm rounded-md transition-colors ${
                    timePeriod === period.value
                      ? 'bg-white text-blue-600 shadow'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  {period.label}
                </button>
              ))}
            </div>

            {/* Chart Type Toggle */}
            <div className="flex items-center space-x-1 bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setChartType('area')}
                className={`px-3 py-1 text-sm rounded-md ${
                  chartType === 'area' ? 'bg-white shadow' : ''
                }`}
              >
                Area
              </button>
              <button
                onClick={() => setChartType('line')}
                className={`px-3 py-1 text-sm rounded-md ${
                  chartType === 'line' ? 'bg-white shadow' : ''
                }`}
              >
                Line
              </button>
            </div>

            {/* Export Button */}
            <button
              onClick={handleExport}
              className="flex items-center space-x-2 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
            >
              <Download className="w-4 h-4" />
              <span className="hidden md:inline">Export</span>
            </button>
          </div>
        </div>

        {/* Custom Date Range */}
        {timePeriod === 'custom' && (
          <div className="flex items-center space-x-4 mb-6 p-4 bg-gray-50 rounded-lg">
            <Calendar className="w-5 h-5 text-gray-600" />
            <div className="flex items-center space-x-2">
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <span className="text-gray-600">to</span>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
        )}

        {/* Statistics Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div className="p-4 bg-blue-50 rounded-lg">
            <p className="text-sm text-blue-600 font-medium">Total Sales</p>
            <p className="text-2xl font-bold text-blue-900 mt-1">{formatCurrency(totalSales)}</p>
          </div>
          
          <div className="p-4 bg-green-50 rounded-lg">
            <p className="text-sm text-green-600 font-medium">Average Sales</p>
            <p className="text-2xl font-bold text-green-900 mt-1">{formatCurrency(avgSales)}</p>
          </div>
          
          <div className="p-4 bg-purple-50 rounded-lg">
            <p className="text-sm text-purple-600 font-medium">Peak Sales</p>
            <p className="text-2xl font-bold text-purple-900 mt-1">{formatCurrency(maxSales)}</p>
          </div>
          
          <div className="p-4 bg-orange-50 rounded-lg">
            <p className="text-sm text-orange-600 font-medium">Trend</p>
            <div className="flex items-center mt-1">
              {isPositiveTrend ? (
                <TrendingUp className="w-5 h-5 text-green-600 mr-1" />
              ) : (
                <TrendingDown className="w-5 h-5 text-red-600 mr-1" />
              )}
              <p className={`text-2xl font-bold ${isPositiveTrend ? 'text-green-900' : 'text-red-900'}`}>
                {Math.abs(trendPercentage).toFixed(1)}%
              </p>
            </div>
          </div>
        </div>

        {/* Chart */}
        <div className="mt-6">
          {isLoading ? (
            <div className="flex items-center justify-center h-96">
              <div className="text-gray-500">Loading sales data...</div>
            </div>
          ) : salesData.length === 0 ? (
            <div className="flex items-center justify-center h-96">
              <div className="text-center">
                <p className="text-gray-500">No sales data available for this period</p>
                <p className="text-sm text-gray-400 mt-2">Try selecting a different time range</p>
              </div>
            </div>
          ) : (
            <ResponsiveContainer width="100%" height={400}>
              {chartType === 'area' ? (
                <AreaChart data={salesData}>
                  <defs>
                    <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#3B82F6" stopOpacity={0.1}/>
                    </linearGradient>
                    <linearGradient id="colorOrders" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10B981" stopOpacity={0.8}/>
                      <stop offset="95%" stopColor="#10B981" stopOpacity={0.1}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                  <XAxis 
                    dataKey="date" 
                    stroke="#6B7280"
                    style={{ fontSize: '12px' }}
                  />
                  <YAxis 
                    stroke="#6B7280"
                    style={{ fontSize: '12px' }}
                    tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <Area 
                    type="monotone" 
                    dataKey="sales" 
                    stroke="#3B82F6" 
                    strokeWidth={2}
                    fillOpacity={1} 
                    fill="url(#colorSales)" 
                    name="Sales Amount"
                  />
                  <Area 
                    type="monotone" 
                    dataKey="orders" 
                    stroke="#10B981" 
                    strokeWidth={2}
                    fillOpacity={1} 
                    fill="url(#colorOrders)" 
                    name="Order Count"
                  />
                </AreaChart>
              ) : (
                <LineChart data={salesData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
                  <XAxis 
                    dataKey="date" 
                    stroke="#6B7280"
                    style={{ fontSize: '12px' }}
                  />
                  <YAxis 
                    stroke="#6B7280"
                    style={{ fontSize: '12px' }}
                    tickFormatter={(value) => `₹${(value / 1000).toFixed(0)}k`}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <Line 
                    type="monotone" 
                    dataKey="sales" 
                    stroke="#3B82F6" 
                    strokeWidth={3}
                    dot={{ fill: '#3B82F6', r: 4 }}
                    activeDot={{ r: 6 }}
                    name="Sales Amount"
                  />
                  <Line 
                    type="monotone" 
                    dataKey="orders" 
                    stroke="#10B981" 
                    strokeWidth={3}
                    dot={{ fill: '#10B981', r: 4 }}
                    activeDot={{ r: 6 }}
                    name="Order Count"
                  />
                </LineChart>
              )}
            </ResponsiveContainer>
          )}
        </div>

        {/* Insights */}
        <div className="mt-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Key Insights</h3>
          <ul className="space-y-2 text-sm text-gray-600">
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Sales are trending {isPositiveTrend ? 'upward' : 'downward'} with a {Math.abs(trendPercentage).toFixed(1)}% change
              </span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Peak sales day reached {formatCurrency(maxSales)}
              </span>
            </li>
            <li className="flex items-start">
              <span className="mr-2">•</span>
              <span>
                Average daily sales: {formatCurrency(avgSales)}
              </span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  )
}