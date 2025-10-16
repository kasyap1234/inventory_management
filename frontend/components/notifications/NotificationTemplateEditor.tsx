'use client'

import React, { useState, useEffect } from 'react'
import { Save, Eye, TestTube, Code, Mail, MessageSquare, Webhook, Smartphone } from 'lucide-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationTemplateAPI } from '@/lib/api'
import { toast } from 'react-hot-toast'

interface NotificationTemplate {
    id?: string
    name: string
    type: 'email' | 'sms' | 'webhook' | 'in_app'
    event_type: string
    subject?: string
    body_template: string
    variables: Record<string, any>
    is_active: boolean
}

interface NotificationTemplateEditorProps {
    template?: NotificationTemplate
    onSave?: (template: NotificationTemplate) => void
    onCancel?: () => void
}

const TEMPLATE_TYPES = [
    { value: 'email', label: 'Email', icon: Mail },
    { value: 'sms', label: 'SMS', icon: MessageSquare },
    { value: 'webhook', label: 'Webhook', icon: Webhook },
    { value: 'in_app', label: 'In-App Notification', icon: Smartphone },
]

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
    'password_reset',
]

const TEMPLATE_VARIABLES = {
    order_created: {
        order_id: 'Order ID',
        customer_name: 'Customer Name',
        total_amount: 'Total Amount',
        order_date: 'Order Date',
        items: 'Order Items (array)',
    },
    low_stock: {
        product_name: 'Product Name',
        current_stock: 'Current Stock Level',
        minimum_level: 'Minimum Stock Level',
        warehouse: 'Warehouse Name',
    },
    user_registered: {
        user_name: 'User Name',
        email: 'Email Address',
        registration_date: 'Registration Date',
    },
}

export default function NotificationTemplateEditor({
    template,
    onSave,
    onCancel
}: NotificationTemplateEditorProps) {
    const [formData, setFormData] = useState<NotificationTemplate>({
        name: '',
        type: 'email',
        event_type: 'order_created',
        subject: '',
        body_template: '',
        variables: {},
        is_active: true,
        ...template
    })

    const [previewMode, setPreviewMode] = useState(false)
    const [testData, setTestData] = useState<Record<string, any>>({})
    const [renderedPreview, setRenderedPreview] = useState('')

    const queryClient = useQueryClient()

    // Save template mutation
    const saveTemplateMutation = useMutation({
        mutationFn: async (templateData: NotificationTemplate) => {
            if (templateData.id) {
                await notificationTemplateAPI.update(templateData.id, {
                    name: templateData.name,
                    type: templateData.type,
                    event_type: templateData.event_type,
                    subject: templateData.subject,
                    body_template: templateData.body_template,
                    variables: templateData.variables,
                    is_active: templateData.is_active,
                })
            } else {
                await notificationTemplateAPI.create({
                    name: templateData.name,
                    type: templateData.type,
                    event_type: templateData.event_type,
                    subject: templateData.subject,
                    body_template: templateData.body_template,
                    variables: templateData.variables,
                    is_active: templateData.is_active,
                })
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notification-templates'] })
            toast.success('Template saved successfully')
            onSave?.(formData)
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.error?.message || 'Failed to save template')
        }
    })

    // Test template mutation
    const testTemplateMutation = useMutation({
        mutationFn: async ({ templateId, testData }: { templateId: string, testData: Record<string, any> }) => {
            const response = await notificationTemplateAPI.test(templateId, testData)
            return response.data
        },
        onSuccess: (data) => {
            setRenderedPreview(data.rendered_body)
            toast.success('Template test successful')
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.error?.message || 'Template test failed')
        }
    })

    // Update test data when event type changes
    useEffect(() => {
        const variables = TEMPLATE_VARIABLES[formData.event_type as keyof typeof TEMPLATE_VARIABLES] || {}
        const newTestData: Record<string, any> = {}

        Object.keys(variables).forEach(key => {
            switch (key) {
                case 'order_id':
                    newTestData[key] = 'ORD-12345'
                    break
                case 'customer_name':
                    newTestData[key] = 'John Doe'
                    break
                case 'total_amount':
                    newTestData[key] = 299.99
                    break
                case 'order_date':
                    newTestData[key] = new Date().toISOString()
                    break
                case 'current_stock':
                    newTestData[key] = 5
                    break
                case 'minimum_level':
                    newTestData[key] = 10
                    break
                case 'product_name':
                    newTestData[key] = 'Sample Product'
                    break
                case 'warehouse':
                    newTestData[key] = 'Main Warehouse'
                    break
                case 'user_name':
                    newTestData[key] = 'Jane Smith'
                    break
                case 'email':
                    newTestData[key] = 'jane@example.com'
                    break
                case 'registration_date':
                    newTestData[key] = new Date().toISOString()
                    break
                default:
                    newTestData[key] = `Sample ${key}`
            }
        })

        setTestData(newTestData)
    }, [formData.event_type])

    const handleSave = () => {
        if (!formData.name.trim()) {
            toast.error('Template name is required')
            return
        }

        if (!formData.body_template.trim()) {
            toast.error('Template body is required')
            return
        }

        if (formData.type === 'email' && !formData.subject?.trim()) {
            toast.error('Subject is required for email templates')
            return
        }

        saveTemplateMutation.mutate(formData)
    }

    const handleTest = () => {
        if (!formData.id) {
            toast.error('Please save the template first before testing')
            return
        }

        testTemplateMutation.mutate({ templateId: formData.id, testData })
    }

    const getTypeIcon = (type: string) => {
        const typeConfig = TEMPLATE_TYPES.find(t => t.value === type)
        const Icon = typeConfig?.icon || Mail
        return <Icon className="w-4 h-4" />
    }

    return (
        <div className="max-w-4xl mx-auto p-6 bg-white rounded-lg shadow-lg">
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-2xl font-bold text-gray-900">
                    {template?.id ? 'Edit' : 'Create'} Notification Template
                </h2>
                <div className="flex space-x-2">
                    <button
                        onClick={() => setPreviewMode(!previewMode)}
                        className="flex items-center space-x-2 px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                    >
                        <Eye className="w-4 h-4" />
                        <span>{previewMode ? 'Edit' : 'Preview'}</span>
                    </button>

                    {formData.id && (
                        <button
                            onClick={handleTest}
                            disabled={testTemplateMutation.isPending}
                            className="flex items-center space-x-2 px-4 py-2 text-blue-700 bg-blue-100 rounded-lg hover:bg-blue-200 disabled:opacity-50"
                        >
                            <TestTube className="w-4 h-4" />
                            <span>Test</span>
                        </button>
                    )}

                    <button
                        onClick={handleSave}
                        disabled={saveTemplateMutation.isPending}
                        className="flex items-center space-x-2 px-4 py-2 text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                        <Save className="w-4 h-4" />
                        <span>Save</span>
                    </button>

                    {onCancel && (
                        <button
                            onClick={onCancel}
                            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
                        >
                            Cancel
                        </button>
                    )}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Form Section */}
                <div className="space-y-6">
                    {/* Basic Information */}
                    <div className="space-y-4">
                        <h3 className="text-lg font-semibold text-gray-900">Basic Information</h3>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Template Name
                            </label>
                            <input
                                type="text"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                placeholder="Enter template name"
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    Type
                                </label>
                                <select
                                    value={formData.type}
                                    onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    {TEMPLATE_TYPES.map(type => (
                                        <option key={type.value} value={type.value}>
                                            {type.label}
                                        </option>
                                    ))}
                                </select>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    Event Type
                                </label>
                                <select
                                    value={formData.event_type}
                                    onChange={(e) => setFormData({ ...formData, event_type: e.target.value })}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                >
                                    {EVENT_TYPES.map(eventType => (
                                        <option key={eventType} value={eventType}>
                                            {eventType.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>

                        {formData.type === 'email' && (
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    Subject
                                </label>
                                <input
                                    type="text"
                                    value={formData.subject || ''}
                                    onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    placeholder="Enter email subject"
                                />
                            </div>
                        )}

                        <div className="flex items-center">
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

                    {/* Template Content */}
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <h3 className="text-lg font-semibold text-gray-900">Template Content</h3>
                            <div className="flex items-center space-x-2 text-sm text-gray-600">
                                <Code className="w-4 h-4" />
                                <span>Go Template Syntax</span>
                            </div>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Body Template
                            </label>
                            <textarea
                                value={formData.body_template}
                                onChange={(e) => setFormData({ ...formData, body_template: e.target.value })}
                                rows={12}
                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                placeholder="Enter template content using Go template syntax..."
                            />
                        </div>

                        {/* Available Variables */}
                        <div className="bg-gray-50 p-4 rounded-lg">
                            <h4 className="text-sm font-medium text-gray-900 mb-2">Available Variables</h4>
                            <div className="grid grid-cols-2 gap-2 text-sm">
                                {Object.entries(TEMPLATE_VARIABLES[formData.event_type as keyof typeof TEMPLATE_VARIABLES] || {}).map(([key, description]) => (
                                    <div key={key} className="flex justify-between">
                                        <code className="text-blue-600">{`{{${key}}}`}</code>
                                        <span className="text-gray-600">{description}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Preview Section */}
                <div className="space-y-6">
                    <div className="flex items-center justify-between">
                        <h3 className="text-lg font-semibold text-gray-900">Preview</h3>
                        <div className="flex items-center space-x-2">
                            {getTypeIcon(formData.type)}
                            <span className="text-sm text-gray-600 capitalize">{formData.type}</span>
                        </div>
                    </div>

                    {/* Test Data */}
                    <div className="space-y-4">
                        <h4 className="text-sm font-medium text-gray-900">Test Data</h4>
                        <div className="space-y-2">
                            {Object.entries(testData).map(([key, value]) => (
                                <div key={key} className="flex items-center space-x-2">
                                    <label className="text-xs text-gray-600 w-24">{key}:</label>
                                    <input
                                        type="text"
                                        value={String(value)}
                                        onChange={(e) => setTestData({ ...testData, [key]: e.target.value })}
                                        className="flex-1 px-2 py-1 text-xs border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                                    />
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Rendered Preview */}
                    <div className="border border-gray-300 rounded-lg p-4 bg-gray-50 min-h-64">
                        <h4 className="text-sm font-medium text-gray-900 mb-2">Rendered Output</h4>
                        {renderedPreview ? (
                            <div className="whitespace-pre-wrap text-sm text-gray-800 bg-white p-3 rounded border">
                                {formData.type === 'email' && formData.subject && (
                                    <div className="border-b pb-2 mb-2">
                                        <strong>Subject:</strong> {formData.subject}
                                    </div>
                                )}
                                {renderedPreview}
                            </div>
                        ) : (
                            <div className="text-sm text-gray-500 italic">
                                {formData.id ? 'Click "Test" to see rendered output' : 'Save template first to test rendering'}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}