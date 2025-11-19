'use client'

import React, { useState, useMemo, useCallback } from 'react'
import { Check, Star, Zap, Building, Users, HardDrive, Activity, Shield, CreditCard } from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import axios from 'axios'
import { toast } from 'react-hot-toast'
import { useStripeCheckout } from '@/hooks/useStripeCheckout'
import { useRazorpayCheckout } from '@/hooks/useRazorpayCheckout'

interface Plan {
  id: string
  name: string
  type: 'basic' | 'professional' | 'enterprise'
  description: string
  price_monthly: number
  price_yearly: number
  currency: string
  features: string[]
  limits: {
    users: number
    storage: number // in bytes
    api_calls: number
    warehouses: number
    products: number
  }
  popular?: boolean
  current?: boolean
}

interface PlanSelectorProps {
  currentPlan?: string
  onPlanSelected?: (planId: string, billingCycle: 'monthly' | 'yearly') => void
}

const SAMPLE_PLANS: Plan[] = [
  {
    id: 'basic',
    name: 'Basic',
    type: 'basic',
    description: 'Perfect for small businesses getting started',
    price_monthly: 2900, // $29.00
    price_yearly: 29000, // $290.00 (save $58)
    currency: 'USD',
    features: [
      'Up to 5 users',
      '10GB storage',
      '10,000 API calls/month',
      '2 warehouses',
      '1,000 products',
      'Basic analytics',
      'Email support',
      'Mobile app access'
    ],
    limits: {
      users: 5,
      storage: 10 * 1024 * 1024 * 1024, // 10GB
      api_calls: 10000,
      warehouses: 2,
      products: 1000
    }
  },
  {
    id: 'professional',
    name: 'Professional',
    type: 'professional',
    description: 'Ideal for growing businesses with advanced needs',
    price_monthly: 7900, // $79.00
    price_yearly: 79000, // $790.00 (save $158)
    currency: 'USD',
    features: [
      'Up to 25 users',
      '100GB storage',
      '100,000 API calls/month',
      '10 warehouses',
      '10,000 products',
      'Advanced analytics',
      'Priority support',
      'Custom integrations',
      'Bulk operations',
      'Advanced reporting'
    ],
    limits: {
      users: 25,
      storage: 100 * 1024 * 1024 * 1024, // 100GB
      api_calls: 100000,
      warehouses: 10,
      products: 10000
    },
    popular: true
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    type: 'enterprise',
    description: 'For large organizations with complex requirements',
    price_monthly: 19900, // $199.00
    price_yearly: 199000, // $1,990.00 (save $398)
    currency: 'USD',
    features: [
      'Unlimited users',
      '1TB storage',
      'Unlimited API calls',
      'Unlimited warehouses',
      'Unlimited products',
      'Enterprise analytics',
      '24/7 phone support',
      'Custom integrations',
      'Advanced security',
      'SLA guarantee',
      'Dedicated account manager',
      'Custom training'
    ],
    limits: {
      users: -1, // unlimited
      storage: 1024 * 1024 * 1024 * 1024, // 1TB
      api_calls: -1, // unlimited
      warehouses: -1, // unlimited
      products: -1 // unlimited
    }
  }
]

const PlanSelector = React.memo(function PlanSelector({ currentPlan, onPlanSelected }: PlanSelectorProps) {
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly')
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null)
  const [paymentMethod, setPaymentMethod] = useState<'stripe' | 'razorpay'>('stripe')

  // Payment hooks
  const stripeCheckout = useStripeCheckout()
  const razorpayCheckout = useRazorpayCheckout()

  // Subscribe to plan mutation
  const subscribeToPlanMutation = useMutation({
    mutationFn: async ({ planId, cycle }: { planId: string, cycle: 'monthly' | 'yearly' }) => {
      const response = await axios.post('/api/subscription/subscribe', {
        plan_id: planId,
        billing_cycle: cycle
      })
      return response.data
    },
    onSuccess: (data) => {
      toast.success('Subscription updated successfully!')
      onPlanSelected?.(data.plan_id, billingCycle)
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to update subscription')
    }
  })

  const formatPrice = useCallback((price: number, currency: string) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(price / 100)
  }, [])

  const formatLimit = useCallback((value: number, unit: string) => {
    if (value === -1) return 'Unlimited'

    if (unit === 'storage') {
      const gb = value / (1024 * 1024 * 1024)
      if (gb >= 1024) {
        return `${(gb / 1024).toFixed(0)}TB`
      }
      return `${gb.toFixed(0)}GB`
    }

    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    }
    if (value >= 1000) {
      return `${(value / 1000).toFixed(0)}K`
    }

    return value.toLocaleString()
  }, [])

  const calculateYearlySavings = useCallback((monthlyPrice: number, yearlyPrice: number) => {
    const monthlyCost = monthlyPrice * 12
    const savings = monthlyCost - yearlyPrice
    const percentage = Math.round((savings / monthlyCost) * 100)
    return { amount: savings, percentage }
  }, [])

  const getPlanIcon = useCallback((type: string) => {
    switch (type) {
      case 'basic':
        return <Users className="w-8 h-8 text-blue-500" />
      case 'professional':
        return <Zap className="w-8 h-8 text-purple-500" />
      case 'enterprise':
        return <Building className="w-8 h-8 text-gray-700" />
      default:
        return <Star className="w-8 h-8 text-gray-500" />
    }
  }, [])

  const handleSelectPlan = useCallback(async (planId: string) => {
    setSelectedPlan(planId)

    try {
      if (paymentMethod === 'stripe') {
        // Use Stripe Checkout
        const userEmail = localStorage.getItem('user_email') || ''
        await stripeCheckout.createSubscription(
          planId,
          userEmail,
          (sessionId) => {
            toast.success('Redirecting to payment...')
          },
          (error) => {
            toast.error(error.message || 'Payment failed')
            setSelectedPlan(null)
          }
        )
      } else {
        // Use Razorpay Checkout
        const userEmail = localStorage.getItem('user_email') || ''
        await razorpayCheckout.createSubscription(
          planId,
          userEmail,
          (subscriptionId) => {
            toast.success('Subscription created successfully!')
            onPlanSelected?.(planId, billingCycle)
          },
          (error) => {
            toast.error('Payment failed')
            setSelectedPlan(null)
          }
        )
      }
    } catch (error) {
      console.error('Payment error:', error)
      setSelectedPlan(null)
    }
  }, [paymentMethod, stripeCheckout, razorpayCheckout, billingCycle, onPlanSelected])

  return (
    <div className="max-w-7xl mx-auto p-6">
      <div className="text-center mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">Choose Your Plan</h1>
        <p className="text-lg text-gray-600 mb-8">
          Select the perfect plan for your business needs
        </p>

        {/* Billing Toggle */}
        <div className="flex items-center justify-center space-x-4 mb-6">
          <span className={`text-sm ${billingCycle === 'monthly' ? 'text-gray-900 font-medium' : 'text-gray-500'}`}>
            Monthly
          </span>
          <button
            onClick={() => setBillingCycle(billingCycle === 'monthly' ? 'yearly' : 'monthly')}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${billingCycle === 'yearly' ? 'bg-blue-600' : 'bg-gray-200'
              }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${billingCycle === 'yearly' ? 'translate-x-6' : 'translate-x-1'
                }`}
            />
          </button>
          <span className={`text-sm ${billingCycle === 'yearly' ? 'text-gray-900 font-medium' : 'text-gray-500'}`}>
            Yearly
          </span>
          {billingCycle === 'yearly' && (
            <span className="text-sm text-green-600 font-medium">Save up to 20%</span>
          )}
        </div>

        {/* Payment Method Selector */}
        <div className="flex items-center justify-center gap-3 mb-8">
          <span className="text-sm text-gray-600 font-medium flex items-center gap-2">
            <CreditCard className="w-4 h-4" />
            Payment Method:
          </span>
          <div className="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-1">
            <button
              onClick={() => setPaymentMethod('stripe')}
              className={`px-4 py-2 text-sm font-medium rounded-md transition-all ${paymentMethod === 'stripe'
                  ? 'bg-white text-blue-600 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
                }`}
            >
              Stripe
            </button>
            <button
              onClick={() => setPaymentMethod('razorpay')}
              className={`px-4 py-2 text-sm font-medium rounded-md transition-all ${paymentMethod === 'razorpay'
                  ? 'bg-white text-blue-600 shadow-sm'
                  : 'text-gray-600 hover:text-gray-900'
                }`}
            >
              Razorpay
            </button>
          </div>
        </div>
      </div>

      {/* Plans Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {SAMPLE_PLANS.map((plan) => {
          const price = billingCycle === 'monthly' ? plan.price_monthly : plan.price_yearly
          const isCurrentPlan = currentPlan === plan.id
          const savings = calculateYearlySavings(plan.price_monthly, plan.price_yearly)

          return (
            <div
              key={plan.id}
              className={`relative bg-white rounded-2xl shadow-lg border-2 transition-all duration-200 ${plan.popular
                  ? 'border-blue-500 ring-2 ring-blue-200'
                  : isCurrentPlan
                    ? 'border-green-500 ring-2 ring-green-200'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
            >
              {/* Popular Badge */}
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <span className="bg-blue-500 text-white px-4 py-1 rounded-full text-sm font-medium">
                    Most Popular
                  </span>
                </div>
              )}

              {/* Current Plan Badge */}
              {isCurrentPlan && (
                <div className="absolute -top-4 right-4">
                  <span className="bg-green-500 text-white px-3 py-1 rounded-full text-sm font-medium">
                    Current Plan
                  </span>
                </div>
              )}

              <div className="p-8">
                {/* Plan Header */}
                <div className="text-center mb-6">
                  <div className="flex justify-center mb-4">
                    {getPlanIcon(plan.type)}
                  </div>
                  <h3 className="text-2xl font-bold text-gray-900 mb-2">{plan.name}</h3>
                  <p className="text-gray-600 mb-4">{plan.description}</p>

                  {/* Pricing */}
                  <div className="mb-4">
                    <div className="flex items-baseline justify-center">
                      <span className="text-4xl font-bold text-gray-900">
                        {formatPrice(price, plan.currency)}
                      </span>
                      <span className="text-gray-600 ml-2">
                        /{billingCycle === 'monthly' ? 'month' : 'year'}
                      </span>
                    </div>

                    {billingCycle === 'yearly' && (
                      <div className="text-sm text-green-600 mt-2">
                        Save {formatPrice(savings.amount, plan.currency)} ({savings.percentage}%) annually
                      </div>
                    )}
                  </div>
                </div>

                {/* Features */}
                <div className="space-y-4 mb-8">
                  {/* Key Limits */}
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div className="flex items-center space-x-2">
                      <Users className="w-4 h-4 text-gray-400" />
                      <span>{formatLimit(plan.limits.users, 'users')} users</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <HardDrive className="w-4 h-4 text-gray-400" />
                      <span>{formatLimit(plan.limits.storage, 'storage')}</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Activity className="w-4 h-4 text-gray-400" />
                      <span>{formatLimit(plan.limits.api_calls, 'api')} API calls</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Building className="w-4 h-4 text-gray-400" />
                      <span>{formatLimit(plan.limits.warehouses, 'warehouses')} warehouses</span>
                    </div>
                  </div>

                  {/* Feature List */}
                  <div className="border-t pt-4">
                    <ul className="space-y-3">
                      {plan.features.map((feature, index) => (
                        <li key={index} className="flex items-start space-x-3">
                          <Check className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" />
                          <span className="text-gray-700 text-sm">{feature}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>

                {/* CTA Button */}
                <button
                  onClick={() => handleSelectPlan(plan.id)}
                  disabled={isCurrentPlan || stripeCheckout.isLoading || razorpayCheckout.isLoading}
                  className={`w-full py-3 px-4 rounded-lg font-medium transition-colors ${isCurrentPlan
                      ? 'bg-green-100 text-green-700 cursor-not-allowed'
                      : plan.popular
                        ? 'bg-blue-600 text-white hover:bg-blue-700'
                        : 'bg-gray-900 text-white hover:bg-gray-800'
                    } ${(stripeCheckout.isLoading || razorpayCheckout.isLoading) && selectedPlan === plan.id ? 'opacity-50' : ''}`}
                >
                  {(stripeCheckout.isLoading || razorpayCheckout.isLoading) && selectedPlan === plan.id
                    ? 'Processing...'
                    : isCurrentPlan
                      ? 'Current Plan'
                      : currentPlan
                        ? 'Switch to This Plan'
                        : 'Get Started'
                  }
                </button>

                {/* Additional Info */}
                {plan.type === 'enterprise' && (
                  <div className="mt-4 text-center">
                    <p className="text-sm text-muted-foreground mt-1">
                      Don&apos;t worry, you can change your plan at any time.
                    </p>    <button className="text-blue-600 hover:text-blue-800 font-medium">
                      Contact Sales
                    </button>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {/* FAQ Section */}
      <div className="mt-16 text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-8">Frequently Asked Questions</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 text-left">
          <div>
            <h3 className="font-semibold text-gray-900 mb-2">Can I change my plan anytime?</h3>
            <p className="text-gray-600">
              Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately.
            </p>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900 mb-2">What happens if I exceed my limits?</h3>
            <p className="text-gray-600">
              We'll notify you when you approach your limits. You can upgrade your plan or purchase additional resources.
            </p>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900 mb-2">Is there a free trial?</h3>
            <p className="text-gray-600">
              Yes, all plans come with a 14-day free trial. No credit card required to start.
            </p>
          </div>
          <div>
            <h3 className="font-semibold text-gray-900 mb-2">Do you offer refunds?</h3>
            <p className="text-gray-600">
              We offer a 30-day money-back guarantee for all annual plans.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
})

export default PlanSelector