'use client'

import React, { useState } from 'react'
import { Download, Eye, CreditCard, Calendar, Filter, Search, CheckCircle, Clock, XCircle } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'

interface Invoice {
  id: string
  invoice_number: string
  amount: number
  currency: string
  status: 'paid' | 'pending' | 'failed' | 'refunded'
  invoice_date: string
  due_date: string
  paid_date?: string
  description: string
  plan_name: string
  billing_period: {
    start: string
    end: string
  }
  payment_method?: {
    type: 'card' | 'bank_transfer' | 'paypal'
    last_four?: string
    brand?: string
  }
  invoice_url?: string
  receipt_url?: string
  tax_amount?: number
  discount_amount?: number
  subtotal: number
}

interface BillingFilters {
  status?: string
  year?: string
  search?: string
}

export default function BillingHistory() {
  const [filters, setFilters] = useState<BillingFilters>({})
  const [showFilters, setShowFilters] = useState(false)

  // Fetch billing history
  const { data: invoices = [], isLoading } = useQuery({
    queryKey: ['billing-history', filters],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filters.status) params.append('status', filters.status)
      if (filters.year) params.append('year', filters.year)
      if (filters.search) params.append('search', filters.search)
      
      const response = await axios.get(`/api/subscription/billing-history?${params}`)
      return response.data.invoices
    }
  })

  // Fetch billing summary
  const { data: summary } = useQuery({
    queryKey: ['billing-summary'],
    queryFn: async () => {
      const response = await axios.get('/api/subscription/billing-summary')
      return response.data
    }
  })

  const formatCurrency = (amount: number, currency: string) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(amount / 100)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'paid':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'pending':
        return <Clock className="w-5 h-5 text-yellow-500" />
      case 'failed':
        return <XCircle className="w-5 h-5 text-red-500" />
      case 'refunded':
        return <XCircle className="w-5 h-5 text-gray-500" />
      default:
        return <Clock className="w-5 h-5 text-gray-500" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'paid':
        return 'text-green-700 bg-green-100'
      case 'pending':
        return 'text-yellow-700 bg-yellow-100'
      case 'failed':
        return 'text-red-700 bg-red-100'
      case 'refunded':
        return 'text-gray-700 bg-gray-100'
      default:
        return 'text-gray-700 bg-gray-100'
    }
  }

  const getPaymentMethodDisplay = (paymentMethod?: Invoice['payment_method']) => {
    if (!paymentMethod) return 'N/A'
    
    switch (paymentMethod.type) {
      case 'card':
        return `${paymentMethod.brand?.toUpperCase()} •••• ${paymentMethod.last_four}`
      case 'bank_transfer':
        return 'Bank Transfer'
      case 'paypal':
        return 'PayPal'
      default:
        return paymentMethod.type
    }
  }

  const handleDownloadInvoice = (invoice: Invoice) => {
    if (invoice.invoice_url) {
      window.open(invoice.invoice_url, '_blank')
    }
  }

  const handleViewInvoice = (invoice: Invoice) => {
    // This would open an invoice preview modal
    console.log('View invoice:', invoice.id)
  }

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: 5 }, (_, i) => currentYear - i)

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Billing History</h1>
        <p className="text-gray-600">View and download your invoices and payment history</p>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center space-x-2 mb-2">
              <CreditCard className="w-5 h-5 text-blue-500" />
              <h3 className="font-semibold text-gray-900">Total Spent</h3>
            </div>
            <p className="text-2xl font-bold text-gray-900">
              {formatCurrency(summary.total_spent, summary.currency)}
            </p>
            <p className="text-sm text-gray-600">All time</p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center space-x-2 mb-2">
              <Calendar className="w-5 h-5 text-green-500" />
              <h3 className="font-semibold text-gray-900">This Year</h3>
            </div>
            <p className="text-2xl font-bold text-gray-900">
              {formatCurrency(summary.year_to_date, summary.currency)}
            </p>
            <p className="text-sm text-gray-600">{currentYear}</p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center space-x-2 mb-2">
              <CheckCircle className="w-5 h-5 text-green-500" />
              <h3 className="font-semibold text-gray-900">Paid Invoices</h3>
            </div>
            <p className="text-2xl font-bold text-gray-900">{summary.paid_invoices}</p>
            <p className="text-sm text-gray-600">Total count</p>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center space-x-2 mb-2">
              <Clock className="w-5 h-5 text-yellow-500" />
              <h3 className="font-semibold text-gray-900">Pending</h3>
            </div>
            <p className="text-2xl font-bold text-gray-900">{summary.pending_invoices}</p>
            <p className="text-sm text-gray-600">Awaiting payment</p>
          </div>
        </div>
      )}

      {/* Filters and Search */}
      <div className="bg-white rounded-lg shadow mb-6">
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="relative">
                <Search className="w-5 h-5 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
                <input
                  type="text"
                  placeholder="Search invoices..."
                  value={filters.search || ''}
                  onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              
              <button
                onClick={() => setShowFilters(!showFilters)}
                className="flex items-center space-x-2 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                <Filter className="w-4 h-4" />
                <span>Filters</span>
              </button>
            </div>

            <button
              onClick={() => window.open('/api/subscription/export-billing-history', '_blank')}
              className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <Download className="w-4 h-4" />
              <span>Export All</span>
            </button>
          </div>

          {/* Filter Options */}
          {showFilters && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Status</label>
                  <select
                    value={filters.status || ''}
                    onChange={(e) => setFilters({ ...filters, status: e.target.value || undefined })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="">All Statuses</option>
                    <option value="paid">Paid</option>
                    <option value="pending">Pending</option>
                    <option value="failed">Failed</option>
                    <option value="refunded">Refunded</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Year</label>
                  <select
                    value={filters.year || ''}
                    onChange={(e) => setFilters({ ...filters, year: e.target.value || undefined })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="">All Years</option>
                    {years.map(year => (
                      <option key={year} value={year}>{year}</option>
                    ))}
                  </select>
                </div>

                <div className="flex items-end">
                  <button
                    onClick={() => setFilters({})}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                  >
                    Clear Filters
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Invoices Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading billing history...</p>
          </div>
        ) : invoices.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <CreditCard className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p>No invoices found</p>
            {Object.keys(filters).length > 0 && (
              <button
                onClick={() => setFilters({})}
                className="mt-2 text-blue-600 hover:text-blue-800"
              >
                Clear filters to see all invoices
              </button>
            )}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Invoice
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Plan
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Amount
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Date
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Payment Method
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {invoices.map((invoice: Invoice) => (
                  <tr key={invoice.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-gray-900">
                          {invoice.invoice_number}
                        </div>
                        <div className="text-sm text-gray-500">
                          {formatDate(invoice.billing_period.start)} - {formatDate(invoice.billing_period.end)}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{invoice.plan_name}</div>
                      <div className="text-sm text-gray-500">{invoice.description}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {formatCurrency(invoice.amount, invoice.currency)}
                      </div>
                      {invoice.tax_amount && (
                        <div className="text-sm text-gray-500">
                          Tax: {formatCurrency(invoice.tax_amount, invoice.currency)}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center space-x-2">
                        {getStatusIcon(invoice.status)}
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(invoice.status)}`}>
                          {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{formatDate(invoice.invoice_date)}</div>
                      {invoice.paid_date && (
                        <div className="text-sm text-gray-500">
                          Paid: {formatDate(invoice.paid_date)}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">
                        {getPaymentMethodDisplay(invoice.payment_method)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => handleViewInvoice(invoice)}
                          className="text-blue-600 hover:text-blue-900"
                          title="View Invoice"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        {invoice.invoice_url && (
                          <button
                            onClick={() => handleDownloadInvoice(invoice)}
                            className="text-gray-600 hover:text-gray-900"
                            title="Download Invoice"
                          >
                            <Download className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}