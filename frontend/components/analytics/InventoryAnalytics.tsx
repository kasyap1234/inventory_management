'use client'

import React, { useState } from 'react'
import { Package, AlertTriangle, TrendingDown, Search, Filter, Download } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { analyticsService } from '@/lib/services'

interface InventoryAnalyticsProps {
  className?: string
}

export default function InventoryAnalytics({ className = '' }: InventoryAnalyticsProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [filterStatus, setFilterStatus] = useState<'all' | 'low' | 'out' | 'optimal'>('all')

  // Fetch inventory data
  const { data: inventoryData, isLoading } = useQuery({
    queryKey: ['inventory-analytics', filterStatus],
    queryFn: async () => {
      const [valuation, lowStock] = await Promise.all([
        analyticsService.getInventoryValuation(),
        analyticsService.getLowStockReport(),
      ])

      const products = (lowStock || []).map((item) => ({
        id: item.productId,
        name: item.productName,
        category: '',
        quantity: item.currentStock,
        minimum_level: item.threshold,
        value: item.stockValue,
      }))

      const categories = Object.entries(valuation.byCategory || {}).map(([name, value]) => ({
        name,
        count: typeof value === 'number' ? value : 0,
        quantity: 0,
      }))

      const summary = {
        total_products: products.length,
        low_stock_count: products.filter((p) => p.quantity > 0 && p.quantity <= p.minimum_level).length,
        out_of_stock_count: products.filter((p) => p.quantity === 0).length,
        total_value: products.reduce((sum, p) => sum + (p.value || 0), 0),
        categories,
      }

      return { products, summary }
    },
  })

  const products = inventoryData?.products || []
  const summary = inventoryData?.summary || {
    total_products: 0,
    low_stock_count: 0,
    out_of_stock_count: 0,
    total_value: 0,
    categories: []
  }

  // Filter products by search term
  const filteredProducts = products.filter((product: any) =>
    product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    product.category?.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6']

  const getStockStatus = (quantity: number, minLevel: number) => {
    if (quantity === 0) return { label: 'Out of Stock', color: 'text-red-600', bg: 'bg-red-100' }
    if (quantity <= minLevel) return { label: 'Low Stock', color: 'text-orange-600', bg: 'bg-orange-100' }
    return { label: 'In Stock', color: 'text-green-600', bg: 'bg-green-100' }
  }

  const handleExport = () => {
    const headers = ['Product Name', 'Category', 'Quantity', 'Min Level', 'Status', 'Value']
    const rows = filteredProducts.map((product: any) => [
      product.name,
      product.category || 'N/A',
      product.quantity,
      product.minimum_level,
      getStockStatus(product.quantity, product.minimum_level).label,
      product.value || 0
    ])
    
    const csv = [
      headers.join(','),
      ...rows.map((row: any[]) => row.join(','))
    ].join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `inventory-analytics-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return (
    <div className={`bg-card text-card-foreground rounded-lg shadow ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold">Inventory Analytics</h2>
            <p className="text-muted-foreground mt-1">Monitor stock levels and identify low inventory</p>
          </div>

          <button
            onClick={handleExport}
            className="flex items-center space-x-2 px-4 py-2 text-foreground bg-muted rounded-lg hover:bg-muted/80 mt-4 md:mt-0"
          >
            <Download className="w-4 h-4" />
            <span>Export Report</span>
          </button>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="p-4 bg-blue-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-blue-600 font-medium">Total Products</p>
                <p className="text-2xl font-bold text-blue-900 mt-1">{summary.total_products}</p>
              </div>
              <Package className="w-8 h-8 text-blue-500" />
            </div>
          </div>

          <div className="p-4 bg-orange-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-orange-600 font-medium">Low Stock</p>
                <p className="text-2xl font-bold text-orange-900 mt-1">{summary.low_stock_count}</p>
              </div>
              <AlertTriangle className="w-8 h-8 text-orange-500" />
            </div>
          </div>

          <div className="p-4 bg-red-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-red-600 font-medium">Out of Stock</p>
                <p className="text-2xl font-bold text-red-900 mt-1">{summary.out_of_stock_count}</p>
              </div>
              <TrendingDown className="w-8 h-8 text-red-500" />
            </div>
          </div>

          <div className="p-4 bg-green-50 rounded-lg">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-green-600 font-medium">Total Value</p>
                <p className="text-2xl font-bold text-green-900 mt-1">
                  ₹{(summary.total_value / 1000).toFixed(0)}k
                </p>
              </div>
              <Package className="w-8 h-8 text-green-500" />
            </div>
          </div>
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          {/* Category Distribution */}
          <div className="p-4 border border-border rounded-lg">
            <h3 className="text-lg font-semibold mb-4">Inventory by Category</h3>
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={summary.categories}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }: any) => `${name}: ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {summary.categories.map((_entry: any, index: number) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Stock Levels */}
          <div className="p-4 border border-border rounded-lg">
            <h3 className="text-lg font-semibold mb-4">Stock Levels by Category</h3>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={summary.categories}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" style={{ fontSize: '12px' }} />
                <YAxis style={{ fontSize: '12px' }} />
                <Tooltip />
                <Legend />
                <Bar dataKey="quantity" fill="#3B82F6" name="Quantity" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Filters and Search */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-4 space-y-4 md:space-y-0">
          <div className="flex items-center space-x-2">
            <div className="relative flex-1 md:w-64">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search products..."
                className="w-full pl-10 pr-4 py-2 border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-ring focus:border-ring"
              />
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Filter className="w-4 h-4 text-muted-foreground" />
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value as any)}
              className="px-4 py-2 border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-ring focus:border-ring"
            >
              <option value="all">All Products</option>
              <option value="low">Low Stock</option>
              <option value="out">Out of Stock</option>
              <option value="optimal">Optimal Stock</option>
            </select>
          </div>
        </div>

        {/* Products Table */}
        <div className="overflow-x-auto">
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading inventory data...</div>
          ) : filteredProducts.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No products found matching your criteria
            </div>
          ) : (
            <table className="w-full">
              <thead className="bg-muted border-b border-border">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Product
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Category
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Quantity
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Min Level
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Value
                  </th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {filteredProducts.map((product: any) => {
                  const status = getStockStatus(product.quantity, product.minimum_level)
                  return (
                    <tr key={product.id} className="hover:bg-muted/50">
                      <td className="px-4 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <div className="flex-shrink-0 h-10 w-10 bg-muted rounded flex items-center justify-center">
                            <Package className="w-5 h-5 text-muted-foreground" />
                          </div>
                          <div className="ml-4">
                            <div className="text-sm font-medium">{product.name}</div>
                            <div className="text-sm text-muted-foreground">{product.sku || 'N/A'}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap">
                        <span className="px-2 py-1 text-xs bg-muted rounded">
                          {product.category || 'Uncategorized'}
                        </span>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap">
                        <div className="text-sm">{product.quantity}</div>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap">
                        <div className="text-sm">{product.minimum_level}</div>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap">
                        <span className={`px-2 py-1 text-xs rounded ${status.bg} ${status.color}`}>
                          {status.label}
                        </span>
                      </td>
                      <td className="px-4 py-4 whitespace-nowrap">
                        <div className="text-sm">
                          ₹{(product.value || 0).toLocaleString()}
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          )}
        </div>

        {/* Low Stock Alerts */}
        {summary.low_stock_count > 0 && (
          <div className="mt-6 p-4 bg-orange-50 border border-orange-200 rounded-lg">
            <div className="flex items-start">
              <AlertTriangle className="w-5 h-5 text-orange-600 mt-0.5 mr-3" />
              <div>
                <h4 className="text-sm font-semibold text-orange-900">Low Stock Alert</h4>
                <p className="text-sm text-orange-700 mt-1">
                  {summary.low_stock_count} product{summary.low_stock_count !== 1 ? 's are' : ' is'} running low on stock. 
                  Consider reordering to maintain optimal inventory levels.
                </p>
                <button className="mt-2 text-sm text-orange-600 hover:text-orange-800 font-medium">
                  View low stock items →
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}