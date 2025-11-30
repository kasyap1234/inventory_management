'use client'

import React from 'react'
import { TrendingUp, TrendingDown, DollarSign, Package, ShoppingCart, Users, AlertTriangle, CheckCircle } from 'lucide-react'
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

interface MetricCardProps {
  title: string
  value: string | number
  change?: number
  icon: React.ReactNode
  trend?: 'up' | 'down'
  color: string
}

interface DashboardChartsProps {
  salesData?: any[]
  inventoryData?: any[]
  revenueData?: any[]
  orderData?: any[]
  metrics?: {
    totalRevenue: number
    totalOrders: number
    totalProducts: number
    activeUsers: number
    revenueChange: number
    ordersChange: number
    productsChange: number
    usersChange: number
  }
}

const MetricCard = ({ title, value, change, icon, trend, color }: MetricCardProps) => (
  <div className="bg-card text-card-foreground rounded-lg shadow p-6">
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium text-muted-foreground">{title}</p>
        <p className="text-2xl font-bold mt-2">{value}</p>
        {change !== undefined && (
          <div className={`flex items-center mt-2 text-sm ${
            trend === 'up' ? 'text-green-600' : 'text-red-600'
          }`}>
            {trend === 'up' ? (
              <TrendingUp className="w-4 h-4 mr-1" />
            ) : (
              <TrendingDown className="w-4 h-4 mr-1" />
            )}
            <span>{Math.abs(change)}% from last month</span>
          </div>
        )}
      </div>
      <div className={`p-3 rounded-full ${color}`}>
        {icon}
      </div>
    </div>
  </div>
)

export default function DashboardCharts({
  salesData = [],
  inventoryData = [],
  revenueData = [],
  orderData = [],
  metrics = {
    totalRevenue: 0,
    totalOrders: 0,
    totalProducts: 0,
    activeUsers: 0,
    revenueChange: 0,
    ordersChange: 0,
    productsChange: 0,
    usersChange: 0,
  }
}: DashboardChartsProps) {
  const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899']

  // Sample data if none provided
  const defaultSalesData = [
    { month: 'Jan', sales: 4000, orders: 240 },
    { month: 'Feb', sales: 3000, orders: 198 },
    { month: 'Mar', sales: 5000, orders: 320 },
    { month: 'Apr', sales: 4500, orders: 280 },
    { month: 'May', sales: 6000, orders: 390 },
    { month: 'Jun', sales: 5500, orders: 350 },
  ]

  const defaultInventoryData = [
    { category: 'Electronics', value: 400 },
    { category: 'Clothing', value: 300 },
    { category: 'Food', value: 200 },
    { category: 'Books', value: 150 },
    { category: 'Other', value: 100 },
  ]

  const defaultRevenueData = [
    { date: '01', revenue: 12000, cost: 8000, profit: 4000 },
    { date: '05', revenue: 15000, cost: 9000, profit: 6000 },
    { date: '10', revenue: 18000, cost: 11000, profit: 7000 },
    { date: '15', revenue: 16000, cost: 10000, profit: 6000 },
    { date: '20', revenue: 20000, cost: 12000, profit: 8000 },
    { date: '25', revenue: 22000, cost: 13000, profit: 9000 },
    { date: '30', revenue: 25000, cost: 14000, profit: 11000 },
  ]

  const displaySalesData = salesData.length > 0 ? salesData : defaultSalesData
  const displayInventoryData = inventoryData.length > 0 ? inventoryData : defaultInventoryData
  const displayRevenueData = revenueData.length > 0 ? revenueData : defaultRevenueData

  return (
    <div className="space-y-6">
      {/* Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricCard
          title="Total Revenue"
          value={`₹${metrics.totalRevenue.toLocaleString()}`}
          change={metrics.revenueChange}
          trend={metrics.revenueChange >= 0 ? 'up' : 'down'}
          icon={<DollarSign className="w-6 h-6 text-white" />}
          color="bg-blue-500"
        />
        
        <MetricCard
          title="Total Orders"
          value={metrics.totalOrders.toLocaleString()}
          change={metrics.ordersChange}
          trend={metrics.ordersChange >= 0 ? 'up' : 'down'}
          icon={<ShoppingCart className="w-6 h-6 text-white" />}
          color="bg-green-500"
        />
        
        <MetricCard
          title="Total Products"
          value={metrics.totalProducts.toLocaleString()}
          change={metrics.productsChange}
          trend={metrics.productsChange >= 0 ? 'up' : 'down'}
          icon={<Package className="w-6 h-6 text-white" />}
          color="bg-yellow-500"
        />
        
        <MetricCard
          title="Active Users"
          value={metrics.activeUsers.toLocaleString()}
          change={metrics.usersChange}
          trend={metrics.usersChange >= 0 ? 'up' : 'down'}
          icon={<Users className="w-6 h-6 text-white" />}
          color="bg-purple-500"
        />
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Sales Trend Chart */}
        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Sales Trend</h3>
            <div className="flex items-center space-x-2">
              <span className="text-sm text-muted-foreground">Last 6 months</span>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={displaySalesData}>
              <defs>
                <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.8}/>
                  <stop offset="95%" stopColor="#3B82F6" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Area 
                type="monotone" 
                dataKey="sales" 
                stroke="#3B82F6" 
                fillOpacity={1} 
                fill="url(#colorSales)" 
                name="Sales (₹)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Orders Chart */}
        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Order Volume</h3>
            <div className="flex items-center space-x-2">
              <span className="text-sm text-muted-foreground">Monthly</span>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={displaySalesData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="orders" fill="#10B981" name="Orders" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Revenue Breakdown */}
        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Revenue Breakdown</h3>
            <div className="flex items-center space-x-2">
              <span className="text-sm text-muted-foreground">This month</span>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={displayRevenueData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="revenue" stroke="#3B82F6" strokeWidth={2} name="Revenue" />
              <Line type="monotone" dataKey="cost" stroke="#EF4444" strokeWidth={2} name="Cost" />
              <Line type="monotone" dataKey="profit" stroke="#10B981" strokeWidth={2} name="Profit" />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Inventory Distribution */}
        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Inventory Distribution</h3>
            <div className="flex items-center space-x-2">
              <span className="text-sm text-muted-foreground">By category</span>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={displayInventoryData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }: any) => `${name}: ${(percent * 100).toFixed(0)}%`}
                outerRadius={80}
                fill="#8884d8"
                dataKey="value"
              >
                {displayInventoryData.map((_entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Quick Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Low Stock Items</p>
              <p className="text-3xl font-bold text-orange-600 mt-2">12</p>
            </div>
            <AlertTriangle className="w-12 h-12 text-orange-500" />
          </div>
          <button className="mt-4 text-sm text-primary hover:underline">
            View details →
          </button>
        </div>

        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Pending Orders</p>
              <p className="text-3xl font-bold text-yellow-600 mt-2">28</p>
            </div>
            <ShoppingCart className="w-12 h-12 text-yellow-500" />
          </div>
          <button className="mt-4 text-sm text-primary hover:underline">
            Process orders →
          </button>
        </div>

        <div className="bg-card text-card-foreground rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Completed Today</p>
              <p className="text-3xl font-bold text-green-600 mt-2">45</p>
            </div>
            <CheckCircle className="w-12 h-12 text-green-500" />
          </div>
          <button className="mt-4 text-sm text-primary hover:underline">
            View completed →
          </button>
        </div>
      </div>
    </div>
  )
}