'use client'

import React, { useState } from 'react'
import { Plus, Edit, Trash2, CreditCard, Building, Shield, CheckCircle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import { toast } from 'react-hot-toast'

interface PaymentMethod {
  id: string
  type: 'card' | 'bank_account' | 'paypal'
  is_default: boolean
  card?: {
    brand: string
    last_four: string
    exp_month: number
    exp_year: number
    funding: 'credit' | 'debit' | 'prepaid'
  }
  bank_account?: {
    bank_name: string
    account_type: 'checking' | 'savings'
    last_four: string
  }
  billing_address: {
    line1: string
    line2?: string
    city: string
    state: string
    postal_code: string
    country: string
  }
  created_at: string
}

interface PaymentMethodFormData {
  type: 'card' | 'bank_account' | 'paypal'
  card_number?: string
  exp_month?: number
  exp_year?: number
  cvc?: string
  cardholder_name?: string
  account_number?: string
  routing_number?: string
  account_type?: 'checking' | 'savings'
  account_holder_name?: string
  paypal_email?: string
  billing_address: {
    line1: string
    line2?: string
    city: string
    state: string
    postal_code: string
    country: string
  }
}

export default function PaymentMethodManager() {
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingMethod, setEditingMethod] = useState<PaymentMethod | null>(null)
  const [formData, setFormData] = useState<PaymentMethodFormData>({
    type: 'card',
    billing_address: {
      line1: '',
      city: '',
      state: '',
      postal_code: '',
      country: 'US'
    }
  })

  const queryClient = useQueryClient()

  // Fetch payment methods
  const { data: paymentMethods = [], isLoading } = useQuery({
    queryKey: ['payment-methods'],
    queryFn: async () => {
      const response = await axios.get('/api/subscription/payment-methods')
      return response.data.payment_methods
    }
  })

  // Add payment method mutation
  const addPaymentMethodMutation = useMutation({
    mutationFn: async (methodData: PaymentMethodFormData) => {
      const response = await axios.post('/api/subscription/payment-methods', methodData)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Payment method added successfully')
      resetForm()
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to add payment method')
    }
  })

  // Update payment method mutation
  const updatePaymentMethodMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string, data: Partial<PaymentMethodFormData> }) => {
      const response = await axios.put(`/api/subscription/payment-methods/${id}`, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Payment method updated successfully')
      resetForm()
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to update payment method')
    }
  })

  // Delete payment method mutation
  const deletePaymentMethodMutation = useMutation({
    mutationFn: async (methodId: string) => {
      await axios.delete(`/api/subscription/payment-methods/${methodId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Payment method deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to delete payment method')
    }
  })

  // Set default payment method mutation
  const setDefaultMutation = useMutation({
    mutationFn: async (methodId: string) => {
      await axios.patch(`/api/subscription/payment-methods/${methodId}/set-default`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Default payment method updated')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to update default payment method')
    }
  })

  const resetForm = () => {
    setFormData({
      type: 'card',
      billing_address: {
        line1: '',
        city: '',
        state: '',
        postal_code: '',
        country: 'US'
      }
    })
    setEditingMethod(null)
    setIsFormOpen(false)
  }

  const handleEdit = (method: PaymentMethod) => {
    setEditingMethod(method)
    setFormData({
      type: method.type,
      billing_address: method.billing_address
    })
    setIsFormOpen(true)
  }

  const handleSave = () => {
    // Basic validation
    if (formData.type === 'card') {
      if (!formData.card_number || !formData.exp_month || !formData.exp_year || !formData.cvc) {
        toast.error('Please fill in all card details')
        return
      }
    } else if (formData.type === 'bank_account') {
      if (!formData.account_number || !formData.routing_number) {
        toast.error('Please fill in all bank account details')
        return
      }
    }

    if (!formData.billing_address.line1 || !formData.billing_address.city ||
      !formData.billing_address.state || !formData.billing_address.postal_code) {
      toast.error('Please fill in all billing address fields')
      return
    }

    if (editingMethod) {
      updatePaymentMethodMutation.mutate({ id: editingMethod.id, data: formData })
    } else {
      addPaymentMethodMutation.mutate(formData)
    }
  }

  const handleDelete = (method: PaymentMethod) => {
    if (method.is_default) {
      toast.error('Cannot delete the default payment method')
      return
    }

    if (window.confirm('Are you sure you want to delete this payment method?')) {
      deletePaymentMethodMutation.mutate(method.id)
    }
  }

  const getPaymentMethodIcon = (type: string) => {
    switch (type) {
      case 'card':
        return <CreditCard className="w-6 h-6 text-blue-500" />
      case 'bank_account':
        return <Building className="w-6 h-6 text-green-500" />
      case 'paypal':
        return <Shield className="w-6 h-6 text-purple-500" />
      default:
        return <CreditCard className="w-6 h-6 text-gray-500" />
    }
  }

  const getCardBrandIcon = (brand: string) => {
    // In a real app, you'd use actual card brand icons
    const brandColors = {
      visa: 'text-blue-600',
      mastercard: 'text-red-600',
      amex: 'text-green-600',
      discover: 'text-orange-600'
    }
    return brandColors[brand.toLowerCase() as keyof typeof brandColors] || 'text-gray-600'
  }

  const formatExpiryDate = (month: number, year: number) => {
    return `${month.toString().padStart(2, '0')}/${year.toString().slice(-2)}`
  }

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: 20 }, (_, i) => currentYear + i)
  const months = Array.from({ length: 12 }, (_, i) => i + 1)

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Payment Methods</h1>
          <p className="text-gray-600">Manage your payment methods and billing information</p>
        </div>
        <button
          onClick={() => setIsFormOpen(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          <span>Add Payment Method</span>
        </button>
      </div>

      {/* Payment Methods List */}
      <div className="space-y-4">
        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="animate-pulse bg-gray-200 h-24 rounded-lg"></div>
            ))}
          </div>
        ) : paymentMethods.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center">
            <CreditCard className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No payment methods</h3>
            <p className="text-gray-600 mb-4">Add a payment method to manage your subscription</p>
            <button
              onClick={() => setIsFormOpen(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Add Payment Method
            </button>
          </div>
        ) : (
          paymentMethods.map((method: PaymentMethod) => (
            <div key={method.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  {getPaymentMethodIcon(method.type)}

                  <div>
                    <div className="flex items-center space-x-2">
                      {method.type === 'card' && method.card && (
                        <>
                          <span className={`font-semibold ${getCardBrandIcon(method.card.brand)}`}>
                            {method.card.brand.toUpperCase()}
                          </span>
                          <span className="text-gray-600">•••• {method.card.last_four}</span>
                          <span className="text-gray-500">
                            Expires {formatExpiryDate(method.card.exp_month, method.card.exp_year)}
                          </span>
                        </>
                      )}

                      {method.type === 'bank_account' && method.bank_account && (
                        <>
                          <span className="font-semibold text-gray-900">
                            {method.bank_account.bank_name}
                          </span>
                          <span className="text-gray-600">
                            {method.bank_account.account_type} •••• {method.bank_account.last_four}
                          </span>
                        </>
                      )}

                      {method.is_default && (
                        <span className="flex items-center space-x-1 px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs font-medium">
                          <CheckCircle className="w-3 h-3" />
                          <span>Default</span>
                        </span>
                      )}
                    </div>

                    <div className="text-sm text-gray-600 mt-1">
                      {method.billing_address.line1}, {method.billing_address.city}, {method.billing_address.state} {method.billing_address.postal_code}
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-2">
                  {!method.is_default && (
                    <button
                      onClick={() => setDefaultMutation.mutate(method.id)}
                      disabled={setDefaultMutation.isPending}
                      className="px-3 py-1 text-blue-600 bg-blue-100 rounded hover:bg-blue-200 disabled:opacity-50"
                    >
                      Set Default
                    </button>
                  )}

                  <button
                    onClick={() => handleEdit(method)}
                    className="p-2 text-gray-600 hover:text-gray-900"
                  >
                    <Edit className="w-4 h-4" />
                  </button>

                  <button
                    onClick={() => handleDelete(method)}
                    disabled={method.is_default || deletePaymentMethodMutation.isPending}
                    className="p-2 text-red-600 hover:text-red-900 disabled:opacity-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Payment Method Form Modal */}
      {isFormOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-screen overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-gray-900">
                  {editingMethod ? 'Edit' : 'Add'} Payment Method
                </h2>
                <button
                  onClick={resetForm}
                  className="text-gray-400 hover:text-gray-600"
                >
                  ×
                </button>
              </div>

              <div className="space-y-6">
                {/* Payment Method Type */}
                {!editingMethod && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-3">
                      Payment Method Type
                    </label>
                    <div className="grid grid-cols-2 gap-4">
                      <button
                        onClick={() => setFormData({ ...formData, type: 'card' })}
                        className={`p-4 border-2 rounded-lg flex items-center space-x-3 ${formData.type === 'card'
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300'
                          }`}
                      >
                        <CreditCard className="w-6 h-6 text-blue-500" />
                        <span className="font-medium">Credit/Debit Card</span>
                      </button>

                      <button
                        onClick={() => setFormData({ ...formData, type: 'bank_account' })}
                        className={`p-4 border-2 rounded-lg flex items-center space-x-3 ${formData.type === 'bank_account'
                            ? 'border-blue-500 bg-blue-50'
                            : 'border-gray-200 hover:border-gray-300'
                          }`}
                      >
                        <Building className="w-6 h-6 text-green-500" />
                        <span className="font-medium">Bank Account</span>
                      </button>
                    </div>
                  </div>
                )}

                {/* Card Details */}
                {formData.type === 'card' && !editingMethod && (
                  <div className="space-y-4">
                    <h3 className="text-lg font-semibold text-gray-900">Card Information</h3>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Cardholder Name
                      </label>
                      <input
                        type="text"
                        value={formData.cardholder_name || ''}
                        onChange={(e) => setFormData({ ...formData, cardholder_name: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="John Doe"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Card Number
                      </label>
                      <input
                        type="text"
                        value={formData.card_number || ''}
                        onChange={(e) => setFormData({ ...formData, card_number: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="1234 5678 9012 3456"
                        maxLength={19}
                      />
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Month
                        </label>
                        <select
                          value={formData.exp_month || ''}
                          onChange={(e) => setFormData({ ...formData, exp_month: parseInt(e.target.value) })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        >
                          <option value="">Month</option>
                          {months.map(month => (
                            <option key={month} value={month}>
                              {month.toString().padStart(2, '0')}
                            </option>
                          ))}
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Year
                        </label>
                        <select
                          value={formData.exp_year || ''}
                          onChange={(e) => setFormData({ ...formData, exp_year: parseInt(e.target.value) })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        >
                          <option value="">Year</option>
                          {years.map(year => (
                            <option key={year} value={year}>{year}</option>
                          ))}
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          CVC
                        </label>
                        <input
                          type="text"
                          value={formData.cvc || ''}
                          onChange={(e) => setFormData({ ...formData, cvc: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          placeholder="123"
                          maxLength={4}
                        />
                      </div>
                    </div>
                  </div>
                )}

                {/* Bank Account Details */}
                {formData.type === 'bank_account' && !editingMethod && (
                  <div className="space-y-4">
                    <h3 className="text-lg font-semibold text-gray-900">Bank Account Information</h3>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Account Holder Name
                      </label>
                      <input
                        type="text"
                        value={formData.account_holder_name || ''}
                        onChange={(e) => setFormData({ ...formData, account_holder_name: e.target.value })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="John Doe"
                      />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Routing Number
                        </label>
                        <input
                          type="text"
                          value={formData.routing_number || ''}
                          onChange={(e) => setFormData({ ...formData, routing_number: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          placeholder="123456789"
                          maxLength={9}
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Account Number
                        </label>
                        <input
                          type="text"
                          value={formData.account_number || ''}
                          onChange={(e) => setFormData({ ...formData, account_number: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                          placeholder="1234567890"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Account Type
                      </label>
                      <select
                        value={formData.account_type || 'checking'}
                        onChange={(e) => setFormData({ ...formData, account_type: e.target.value as 'checking' | 'savings' })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="checking">Checking</option>
                        <option value="savings">Savings</option>
                      </select>
                    </div>
                  </div>
                )}

                {/* Billing Address */}
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-gray-900">Billing Address</h3>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Address Line 1
                    </label>
                    <input
                      type="text"
                      value={formData.billing_address.line1}
                      onChange={(e) => setFormData({
                        ...formData,
                        billing_address: { ...formData.billing_address, line1: e.target.value }
                      })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="123 Main St"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Address Line 2 (Optional)
                    </label>
                    <input
                      type="text"
                      value={formData.billing_address.line2 || ''}
                      onChange={(e) => setFormData({
                        ...formData,
                        billing_address: { ...formData.billing_address, line2: e.target.value }
                      })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Apt 4B"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        City
                      </label>
                      <input
                        type="text"
                        value={formData.billing_address.city}
                        onChange={(e) => setFormData({
                          ...formData,
                          billing_address: { ...formData.billing_address, city: e.target.value }
                        })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="New York"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        State
                      </label>
                      <input
                        type="text"
                        value={formData.billing_address.state}
                        onChange={(e) => setFormData({
                          ...formData,
                          billing_address: { ...formData.billing_address, state: e.target.value }
                        })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="NY"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Postal Code
                      </label>
                      <input
                        type="text"
                        value={formData.billing_address.postal_code}
                        onChange={(e) => setFormData({
                          ...formData,
                          billing_address: { ...formData.billing_address, postal_code: e.target.value }
                        })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="10001"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Country
                      </label>
                      <select
                        value={formData.billing_address.country}
                        onChange={(e) => setFormData({
                          ...formData,
                          billing_address: { ...formData.billing_address, country: e.target.value }
                        })}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="US">United States</option>
                        <option value="CA">Canada</option>
                        <option value="GB">United Kingdom</option>
                        <option value="AU">Australia</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Security Notice */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                  <div className="flex items-center space-x-2">
                    <Shield className="w-5 h-5 text-blue-600" />
                    <h4 className="font-medium text-blue-800">Secure Payment Processing</h4>
                  </div>
                  <p className="text-blue-700 text-sm mt-2">
                    Your payment information is encrypted and processed securely. We never store your full card number or bank account details.
                  </p>
                </div>

                {/* Actions */}
                <div className="flex justify-end space-x-4 pt-6 border-t">
                  <button
                    onClick={resetForm}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleSave}
                    disabled={addPaymentMethodMutation.isPending || updatePaymentMethodMutation.isPending}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {editingMethod ? 'Update' : 'Add'} Payment Method
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}