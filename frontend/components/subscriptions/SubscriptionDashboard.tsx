'use client'

import React, { useState } from 'react'
import { CreditCard, Calendar, TrendingUp, AlertCircle, CheckCircle, Clock, DollarSign } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'

interface Subscription {
  id: string
  plan_name: string
  plan_type: 'basic' | 'professional' | 'enterprise'
  status: 'active' | 'cancelled' | 'past_due' | 'trialing'
  current_period_start: string
  current_period_end: string
  amount: number
  currency: string
  billing_cycle: 'monthly' | 'yearly'
  features: string[]
  usage: {
    users: { current: number; limit: number }
    storage: { current: number; limit: number }
    api_calls: { current: number; limit: number }
  }
}

interface BillingHistory {
  id: string
  amount: number
  currency: string
  status: 'paid' | 'pending' | 'failed'
  invoice_date: string
  due_date: string
  description: string
  invoice_url?: string
}

export default function SubscriptionDashboard() {
  const [activeTab, setActiveTab] = useState<'overview' | 'billing' | 'usage'>('overview')

  // Fetch subscription data
  const { data: subscription, isLoading: subscriptionLoading } = useQuery({
    queryKey: ['subscription'],
    queryFn: async () => {
      const response = await axios.get('/api/subscription')
      return response.data.subscription
    }
  })

  // Fetch billing history
  const { data: billingHistory = [], isLoading: billingLoading } = useQuery({
    queryKey: ['billing-history'],
    queryFn: async () => {
      const response = await axios.get('/api/subscription/billing-history')
      return response.data.history
    }
  })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'trialing':
        return <Clock className="w-5 h-5 text-blue-500" />
      case 'past_due':
        return <AlertCircle className="w-5 h-5 text-yellow-500" />
      case 'cancelled':
        return <AlertCircle className="w-5 h-5 text-red-500" />
      default:
        return <AlertCircle className="w-5 h-5 text-gray-500" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'text-green-700 bg-green-100'
      case 'trialing':
        return 'text-blue-700 bg-blue-100'
      case 'past_due':
        return 'text-yellow-700 bg-yellow-100'
      case 'cancelled':
        return 'text-red-700 bg-red-100'
      default:
        return 'text-gray-700 bg-gray-100'
    }
  }

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(amount / 100)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  const calculateUsagePercentage = (current: number, limit: number) => {
    return Math.min((current / limit) * 100, 100)
  }

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-500'
    if (percentage >= 75) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  if (subscriptionLoading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-32 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Subscription Management</h1>
        <p className="text-gray-600">Manage your subscription, billing, and usage</p>
      </div>

      {/* Subscription Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {/* Current Plan */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              {getStatusIcon(subscription?.status)}
              <h3 className="text-lg font-semibold text-gray-900">Current Plan</h3>
            </div>
            <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(subscription?.status)}`}>
              {subscription?.status?.replace('_', ' ').toUpperCase()}
            </span>
          </div>
          <div>
            <p className="text-2xl font-bold text-gray-900 capitalize">{subscription?.plan_name}</p>
            <p className="text-gray-600">
              {formatCurrency(subscription?.amount, subscription?.currency)} / {subscription?.billing_cycle}
            </p>
          </div>
        </div>

        {/* Billing Period */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center space-x-2 mb-4">
            <Calendar className="w-5 h-5 text-blue-500" />
            <h3 className="text-lg font-semibold text-gray-900">Billing Period</h3>
          </div>
          <div>
            <p className="text-sm text-gray-600">Current period ends</p>
            <p className="text-lg font-semibold text-gray-900">
              {formatDate(subscription?.current_period_end)}
            </p>
            <p className="text-sm text-gray-500 mt-1">
              Started {formatDate(subscription?.current_period_start)}
            </p>
          </div>
        </div>

        {/* Next Payment */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center space-x-2 mb-4">
            <DollarSign className="w-5 h-5 text-green-500" />
            <h3 className="text-lg font-semibold text-gray-900">Next Payment</h3>
          </div>
          <div>
            <p className="text-2xl font-bold text-gray-900">
              {formatCurrency(subscription?.amount, subscription?.currency)}
            </p>
            <p className="text-gray-600">
              Due {formatDate(subscription?.current_period_end)}
            </p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="bg-white rounded-lg shadow">
        <div className="border-b border-gray-200">
          <nav className="flex space-x-8 px-6">
            {[
              { id: 'overview', label: 'Overview', icon: TrendingUp },
              { id: 'billing', label: 'Billing History', icon: CreditCard },
              { id: 'usage', label: 'Usage', icon: TrendingUp },
            ].map(tab => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`flex items-center space-x-2 py-4 border-b-2 font-medium text-sm ${activeTab === tab.id
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                >
                  <Icon className="w-4 h-4" />
                  <span>{tab.label}</span>
                </button>
              )
            })}
          </nav>
        </div>

        <div className="p-6">
          {/* Overview Tab */}
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Plan Features */}
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Plan Features</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {subscription?.features?.map((feature: string, index: number) => (
                    <div key={index} className="flex items-center space-x-2">
                      <CheckCircle className="w-4 h-4 text-green-500" />
                      <span className="text-gray-700">{feature}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Quick Actions */}
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
                <div className="flex flex-wrap gap-4">
                  <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                    Upgrade Plan
                  </button>
                  <button className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
                    Update Payment Method
                  </button>
                  <button className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200">
                    Download Invoice
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Billing History Tab */}
          {activeTab === 'billing' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Billing History</h3>
              {billingLoading ? (
                <div className="space-y-4">
                  {[1, 2, 3].map(i => (
                    <div key={i} className="animate-pulse h-16 bg-gray-200 rounded"></div>
                  ))}
                </div>
              ) : billingHistory.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <CreditCard className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                  <p>No billing history available</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {billingHistory.map((invoice: BillingHistory) => (
                    <div key={invoice.id} className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                      <div className="flex items-center space-x-4">
                        <div className={`w-3 h-3 rounded-full ${invoice.status === 'paid' ? 'bg-green-500' :
                            invoice.status === 'pending' ? 'bg-yellow-500' : 'bg-red-500'
                          }`}></div>
                        <div>
                          <p className="font-medium text-gray-900">{invoice.description}</p>
                          <p className="text-sm text-gray-600">
                            {formatDate(invoice.invoice_date)} • Due {formatDate(invoice.due_date)}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-4">
                        <div className="text-right">
                          <p className="font-semibold text-gray-900">
                            {formatCurrency(invoice.amount, invoice.currency)}
                          </p>
                          <p className={`text-sm capitalize ${invoice.status === 'paid' ? 'text-green-600' :
                              invoice.status === 'pending' ? 'text-yellow-600' : 'text-red-600'
                            }`}>
                            {invoice.status}
                          </p>
                        </div>
                        {invoice.invoice_url && (
                          <button
                            onClick={() => window.open(invoice.invoice_url, '_blank')}
                            className="px-3 py-1 text-blue-600 bg-blue-100 rounded hover:bg-blue-200"
                          >
                            Download
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Usage Tab */}
          {activeTab === 'usage' && (
            <div className="space-y-6">
              <h3 className="text-lg font-semibold text-gray-900">Current Usage</h3>

              {subscription?.usage && (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Users */}
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-medium text-gray-900">Users</h4>
                      <span className="text-sm text-gray-600">
                        {subscription.usage.users.current} / {subscription.usage.users.limit}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className={`h-2 rounded-full ${getUsageColor(
                          calculateUsagePercentage(subscription.usage.users.current, subscription.usage.users.limit)
                        )}`}
                        style={{
                          width: `${calculateUsagePercentage(subscription.usage.users.current, subscription.usage.users.limit)}%`
                        }}
                      ></div>
                    </div>
                  </div>

                  {/* Storage */}
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-medium text-gray-900">Storage</h4>
                      <span className="text-sm text-gray-600">
                        {(subscription.usage.storage.current / 1024 / 1024 / 1024).toFixed(1)} GB / {(subscription.usage.storage.limit / 1024 / 1024 / 1024).toFixed(0)} GB
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className={`h-2 rounded-full ${getUsageColor(
                          calculateUsagePercentage(subscription.usage.storage.current, subscription.usage.storage.limit)
                        )}`}
                        style={{
                          width: `${calculateUsagePercentage(subscription.usage.storage.current, subscription.usage.storage.limit)}%`
                        }}
                      ></div>
                    </div>
                  </div>

                  {/* API Calls */}
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="font-medium text-gray-900">API Calls</h4>
                      <span className="text-sm text-gray-600">
                        {subscription.usage.api_calls.current.toLocaleString()} / {subscription.usage.api_calls.limit.toLocaleString()}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div
                        className={`h-2 rounded-full ${getUsageColor(
                          calculateUsagePercentage(subscription.usage.api_calls.current, subscription.usage.api_calls.limit)
                        )}`}
                        style={{
                          width: `${calculateUsagePercentage(subscription.usage.api_calls.current, subscription.usage.api_calls.limit)}%`
                        }}
                      ></div>
                    </div>
                  </div>
                </div>
              )}

              {/* Usage Alerts */}
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div className="flex items-center space-x-2">
                  <AlertCircle className="w-5 h-5 text-yellow-600" />
                  <h4 className="font-medium text-yellow-800">Usage Alerts</h4>
                </div>
                <p className="text-yellow-700 mt-2">
                  You're approaching your usage limits. Consider upgrading your plan to avoid service interruptions.
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}