'use client'

import React, { useMemo, useState } from 'react'
import { Plus, Edit, Trash2, Globe, CheckCircle, XCircle, AlertCircle, TestTube } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { webhookService } from '@/lib/services'
import { toast } from 'react-hot-toast'
import { api } from '@/lib/api'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { tokenStorage } from '@/lib/security'

interface WebhookSubscription {
  id: string
  name: string
  url: string
  event_types: string[]
  headers: Record<string, string>
  timeout_seconds: number
  retry_count: number
  is_active: boolean
  last_success_at?: string
  last_failure_at?: string
  failure_count: number
  created_at: string
}

interface WebhookFormData {
  name: string
  url: string
  event_types: string[]
  headers: Record<string, string>
  timeout_seconds: number
  retry_count: number
  is_active: boolean
}

const EVENT_TYPES = [
  'order_created',
  'order_approved',
  'order_shipped',
  'order_delivered',
  'order_cancelled',
  'payment_received',
  'payment_failed',
  'low_stock',
  'user_registered',
  'inventory_updated',
]

type WebhookTestResult = {
  success: boolean
  target_status?: number
  response_headers?: Record<string, string>
  response_body_snippet?: string
  duration_ms: number
  signature: { algorithm: string; header_name: string }
  error?: string
}

function decodeJwtPerms(): Set<string> {
  const token = tokenStorage.getAccessToken()
  if (!token) return new Set()
  try {
    const parts = token.split('.')
    if (parts.length < 2) return new Set()
    const json = JSON.parse(typeof atob !== 'undefined' ? atob(parts[1]) : Buffer.from(parts[1], 'base64').toString())
    const perms: string[] = Array.isArray(json?.permissions) ? json.permissions : []
    return new Set(perms)
  } catch {
    return new Set()
  }
}

export default function WebhookSubscriptionManager() {
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [editingWebhook, setEditingWebhook] = useState<WebhookSubscription | null>(null)
  const [formData, setFormData] = useState<WebhookFormData>({
    name: '',
    url: '',
    event_types: [],
    headers: {},
    timeout_seconds: 30,
    retry_count: 3,
    is_active: true,
  })
  const [newHeaderKey, setNewHeaderKey] = useState('')
  const [newHeaderValue, setNewHeaderValue] = useState('')

  const queryClient = useQueryClient()

  const [isTestModalOpen, setIsTestModalOpen] = useState(false)
  const [testResult, setTestResult] = useState<WebhookTestResult | null>(null)
  const [isTesting, setIsTesting] = useState(false)

  const hasWebhookTestPermission = useMemo(() => {
    const perms = decodeJwtPerms()
    return perms.has('webhooks:test') || perms.has('notifications:manage')
  }, [])

  // Fetch webhooks
  const { data: webhooks = [], isLoading } = useQuery({
    queryKey: ['webhook-subscriptions'],
    queryFn: async () => {
      const { data } = await webhookService.list()
      // backend returns { webhook_subscriptions: [...] }
      return (data as any)?.webhook_subscriptions ?? []
    }
  })

  // Save webhook mutation
  const saveWebhookMutation = useMutation({
    mutationFn: async (webhookData: WebhookFormData) => {
      if (editingWebhook) {
        await webhookService.update(editingWebhook.id, webhookData as any)
      } else {
        await webhookService.create(webhookData as any)
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhook-subscriptions'] })
      toast.success(`Webhook ${editingWebhook ? 'updated' : 'created'} successfully`)
      resetForm()
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to save webhook')
    }
  })

  // Delete webhook mutation
  const deleteWebhookMutation = useMutation({
    mutationFn: async (webhookId: string) => {
      await webhookService.delete(webhookId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhook-subscriptions'] })
      toast.success('Webhook deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error?.message || 'Failed to delete webhook')
    }
  })

  // Test webhook mutation
  const handleTestWebhook = async (webhook: WebhookSubscription) => {
    if (!hasWebhookTestPermission) {
      toast.error('You do not have permission to test webhooks')
      return
    }
    setIsTesting(true)
    const t = toast.loading('Testing webhook...')
    try {
      const res = await (api as any).webhooks.test({
        target_url: webhook.url,
        method: 'POST',
        headers: webhook.headers,
        payload: { test: true, message: 'webhook delivery test' },
        secret_id: webhook.id,
        event_type: 'webhook_test',
      })
      setTestResult(res as WebhookTestResult)
      setIsTestModalOpen(true)
      toast.success('Test webhook executed')
    } catch (e: any) {
      const msg = e?.response?.data?.error?.message || e?.message || 'Failed to test webhook'
      setTestResult({
        success: false,
        duration_ms: 0,
        error: msg,
        signature: { algorithm: 'HMAC-SHA256', header_name: 'X-Webhook-Signature' },
      })
      setIsTestModalOpen(true)
      toast.error(msg)
    } finally {
      setIsTesting(false)
      toast.dismiss(t)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      url: '',
      event_types: [],
      headers: {},
      timeout_seconds: 30,
      retry_count: 3,
      is_active: true,
    })
    setEditingWebhook(null)
    setIsFormOpen(false)
    setNewHeaderKey('')
    setNewHeaderValue('')
  }

  const handleEdit = (webhook: WebhookSubscription) => {
    setEditingWebhook(webhook)
    setFormData({
      name: webhook.name,
      url: webhook.url,
      event_types: webhook.event_types,
      headers: webhook.headers,
      timeout_seconds: webhook.timeout_seconds,
      retry_count: webhook.retry_count,
      is_active: webhook.is_active,
    })
    setIsFormOpen(true)
  }

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error('Webhook name is required')
      return
    }
    
    if (!formData.url.trim()) {
      toast.error('Webhook URL is required')
      return
    }

    if (formData.event_types.length === 0) {
      toast.error('At least one event type must be selected')
      return
    }

    saveWebhookMutation.mutate(formData)
  }

  const handleDelete = (webhook: WebhookSubscription) => {
    if (window.confirm(`Are you sure you want to delete "${webhook.name}"?`)) {
      deleteWebhookMutation.mutate(webhook.id)
    }
  }

  const addHeader = () => {
    if (newHeaderKey.trim() && newHeaderValue.trim()) {
      setFormData({
        ...formData,
        headers: {
          ...formData.headers,
          [newHeaderKey]: newHeaderValue
        }
      })
      setNewHeaderKey('')
      setNewHeaderValue('')
    }
  }

  const removeHeader = (key: string) => {
    const newHeaders = { ...formData.headers }
    delete newHeaders[key]
    setFormData({ ...formData, headers: newHeaders })
  }

  const getStatusIcon = (webhook: WebhookSubscription) => {
    if (!webhook.is_active) {
      return <XCircle className="w-5 h-5 text-gray-400" />
    }
    
    if (webhook.failure_count > 0) {
      return <AlertCircle className="w-5 h-5 text-yellow-500" />
    }
    
    if (webhook.last_success_at) {
      return <CheckCircle className="w-5 h-5 text-green-500" />
    }
    
    return <Globe className="w-5 h-5 text-blue-500" />
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never'
    return new Date(dateString).toLocaleString()
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Webhook Subscriptions</h1>
          <p className="text-gray-600">Manage webhook endpoints for receiving event notifications</p>
        </div>
        <button
          onClick={() => setIsFormOpen(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          <span>Add Webhook</span>
        </button>
      </div>

      {/* Webhooks List */}
      <div className="bg-white rounded-lg shadow">
        {isLoading ? (
          <div className="p-8 text-center text-gray-500">Loading webhooks...</div>
        ) : webhooks.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <Globe className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p>No webhook subscriptions configured</p>
            <button
              onClick={() => setIsFormOpen(true)}
              className="mt-4 text-blue-600 hover:text-blue-800"
            >
              Create your first webhook
            </button>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {webhooks.map((webhook: WebhookSubscription) => (
              <div key={webhook.id} className="p-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    {getStatusIcon(webhook)}
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900">{webhook.name}</h3>
                      <p className="text-sm text-gray-600">{webhook.url}</p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => handleTestWebhook(webhook)}
                      disabled={!hasWebhookTestPermission || isTesting}
                      title={!hasWebhookTestPermission ? 'Insufficient permission' : 'Test webhook delivery'}
                      className={`flex items-center space-x-1 px-3 py-1 rounded ${!hasWebhookTestPermission || isTesting ? 'text-gray-400 bg-gray-100 cursor-not-allowed' : 'text-blue-600 bg-blue-100 hover:bg-blue-200'}`}
                    >
                      <TestTube className="w-4 h-4" />
                      <span>Test</span>
                    </button>

                    <button
                      onClick={() => handleEdit(webhook)}
                      className="flex items-center space-x-1 px-3 py-1 text-gray-600 bg-gray-100 rounded hover:bg-gray-200"
                    >
                      <Edit className="w-4 h-4" />
                      <span>Edit</span>
                    </button>
                    
                    <button
                      onClick={() => handleDelete(webhook)}
                      disabled={deleteWebhookMutation.isPending}
                      className="flex items-center space-x-1 px-3 py-1 text-red-600 bg-red-100 rounded hover:bg-red-200 disabled:opacity-50"
                    >
                      <Trash2 className="w-4 h-4" />
                      <span>Delete</span>
                    </button>
                  </div>
                </div>

                <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                  <div>
                    <span className="font-medium text-gray-700">Event Types:</span>
                    <div className="mt-1 flex flex-wrap gap-1">
                      {webhook.event_types.map(eventType => (
                        <span key={eventType} className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">
                          {eventType}
                        </span>
                      ))}
                    </div>
                  </div>
                  
                  <div>
                    <span className="font-medium text-gray-700">Last Success:</span>
                    <p className="text-gray-600">{formatDate(webhook.last_success_at)}</p>
                  </div>
                  
                  <div>
                    <span className="font-medium text-gray-700">Failure Count:</span>
                    <p className={`${webhook.failure_count > 0 ? 'text-red-600' : 'text-gray-600'}`}>
                      {webhook.failure_count}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Test Result Modal */}
      <Dialog open={isTestModalOpen} onOpenChange={setIsTestModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Webhook Test Result</DialogTitle>
            <DialogDescription>
              Review the response details returned by the target endpoint.
            </DialogDescription>
          </DialogHeader>
          {testResult ? (
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <div><span className="font-medium">Success:</span> {String(testResult.success)}</div>
                <div><span className="font-medium">Status:</span> {testResult.target_status ?? '—'}</div>
              </div>
              <div>
                <span className="font-medium">Duration:</span> {testResult.duration_ms} ms
              </div>
              <div>
                <span className="font-medium">Content-Type:</span>{' '}
                {testResult.response_headers?.['content-type'] ?? '—'}
              </div>
              <div>
                <span className="font-medium">Body Snippet:</span>
                <pre className="mt-1 p-2 bg-gray-100 rounded max-h-64 overflow-auto text-xs whitespace-pre-wrap">
                  {testResult.response_body_snippet ?? '—'}
                </pre>
              </div>
              {testResult.error && (
                <div className="text-red-600"><span className="font-medium">Error:</span> {testResult.error}</div>
              )}
              <div className="text-xs text-gray-500">
                Signature: {testResult.signature?.algorithm} via header {testResult.signature?.header_name}
              </div>
            </div>
          ) : (
            <div>Loading...</div>
          )}
          <div className="mt-4 flex justify-end">
            <button
              onClick={() => setIsTestModalOpen(false)}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
            >
              Close
            </button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Webhook Form Modal */}
      {isFormOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-screen overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-bold text-gray-900">
                  {editingWebhook ? 'Edit' : 'Create'} Webhook Subscription
                </h2>
                <button
                  onClick={resetForm}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <XCircle className="w-6 h-6" />
                </button>
              </div>

              <div className="space-y-6">
                {/* Basic Information */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Name
                    </label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      placeholder="Enter webhook name"
                    />
                  </div>

                  <div className="flex items-center pt-8">
                    <input
                      type="checkbox"
                      id="is_active"
                      checked={formData.is_active}
                      onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                      className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                    />
                    <label htmlFor="is_active" className="ml-2 text-sm text-gray-700">
                      Active
                    </label>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Webhook URL
                  </label>
                  <input
                    type="url"
                    value={formData.url}
                    onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="https://your-domain.com/webhook"
                  />
                </div>

                {/* Event Types */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Event Types
                  </label>
                  <div className="grid grid-cols-2 gap-2">
                    {EVENT_TYPES.map(eventType => (
                      <label key={eventType} className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          checked={formData.event_types.includes(eventType)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setFormData({
                                ...formData,
                                event_types: [...formData.event_types, eventType]
                              })
                            } else {
                              setFormData({
                                ...formData,
                                event_types: formData.event_types.filter(t => t !== eventType)
                              })
                            }
                          }}
                          className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-700">
                          {eventType.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                        </span>
                      </label>
                    ))}
                  </div>
                </div>

                {/* Configuration */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Timeout (seconds)
                    </label>
                    <input
                      type="number"
                      min="1"
                      max="300"
                      value={formData.timeout_seconds}
                      onChange={(e) => setFormData({ ...formData, timeout_seconds: parseInt(e.target.value) || 30 })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Retry Count
                    </label>
                    <input
                      type="number"
                      min="0"
                      max="10"
                      value={formData.retry_count}
                      onChange={(e) => setFormData({ ...formData, retry_count: parseInt(e.target.value) || 3 })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                </div>

                {/* Custom Headers */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Custom Headers
                  </label>
                  
                  {/* Existing Headers */}
                  {Object.entries(formData.headers).map(([key, value]) => (
                    <div key={key} className="flex items-center space-x-2 mb-2">
                      <input
                        type="text"
                        value={key}
                        readOnly
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg bg-gray-50"
                      />
                      <input
                        type="text"
                        value={value}
                        onChange={(e) => setFormData({
                          ...formData,
                          headers: { ...formData.headers, [key]: e.target.value }
                        })}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                      />
                      <button
                        onClick={() => removeHeader(key)}
                        className="p-2 text-red-600 hover:text-red-800"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}

                  {/* Add New Header */}
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={newHeaderKey}
                      onChange={(e) => setNewHeaderKey(e.target.value)}
                      placeholder="Header name"
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <input
                      type="text"
                      value={newHeaderValue}
                      onChange={(e) => setNewHeaderValue(e.target.value)}
                      placeholder="Header value"
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <button
                      onClick={addHeader}
                      disabled={!newHeaderKey.trim() || !newHeaderValue.trim()}
                      className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      Add
                    </button>
                  </div>
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
                    disabled={saveWebhookMutation.isPending}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {editingWebhook ? 'Update' : 'Create'} Webhook
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