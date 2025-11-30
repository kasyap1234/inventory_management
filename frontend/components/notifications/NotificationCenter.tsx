'use client'

import React, { useState } from 'react'
import { Bell, X, AlertCircle, Info, CheckCircle, XCircle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationService, type NotificationListResult } from '@/lib/services'
import { toast } from 'react-hot-toast'

interface Notification {
  id: string
  title: string
  message: string
  notification_type: 'info' | 'warning' | 'error' | 'success'
  priority: 'low' | 'normal' | 'high' | 'critical'
  status: 'unread' | 'read' | 'archived'
  created_at: string
  expires_at?: string
  event_type?: string
}

interface NotificationCenterProps {
  className?: string
}

export default function NotificationCenter({ className = '' }: NotificationCenterProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [filter, setFilter] = useState<'all' | 'unread' | 'read'>('all')
  const queryClient = useQueryClient()

  // Fetch notifications
  const { data: listResult, isLoading } = useQuery<NotificationListResult>({
    queryKey: ['notifications', filter],
    queryFn: async () => {
      const status = filter === 'all' ? undefined : filter
      return notificationService.list({ status })
    },
    refetchInterval: 30000,
  })

  const notifications: Notification[] = (listResult?.items as unknown as Notification[]) ?? []

  // Mark notification as read
  const markAsReadMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      await notificationService.markAsRead(notificationId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
    onError: () => {
      toast.error('Failed to mark notification as read')
    }
  })

  // Mark all as read
  const markAllAsReadMutation = useMutation({
    mutationFn: async () => {
      await notificationService.markAllAsRead()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      toast.success('All notifications marked as read')
    },
    onError: () => {
      toast.error('Failed to mark all notifications as read')
    }
  })

  // Archive notification
  const deleteNotificationMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      await notificationService.archive(notificationId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
    },
    onError: () => {
      toast.error('Failed to archive notification')
    }
  })

  const unreadCount = notifications.filter((n: Notification) => n.status === 'unread').length

  const getNotificationIcon = (type: string, priority: string) => {
    const iconClass = `w-5 h-5 ${
      priority === 'critical' ? 'text-red-600' :
      priority === 'high' ? 'text-orange-500' :
      priority === 'normal' ? 'text-blue-500' :
      'text-gray-500'
    }`

    switch (type) {
      case 'error':
        return <XCircle className={iconClass} />
      case 'warning':
        return <AlertCircle className={iconClass} />
      case 'success':
        return <CheckCircle className={iconClass} />
      default:
        return <Info className={iconClass} />
    }
  }

  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60))

    if (diffInMinutes < 1) return 'Just now'
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}h ago`
    return `${Math.floor(diffInMinutes / 1440)}d ago`
  }

  return (
    <div className={`relative ${className}`}>
      {/* Notification Bell */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-muted-foreground hover:text-foreground focus:outline-none focus:ring-2 focus:ring-ring rounded-lg"
      >
        <Bell className="w-6 h-6" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {/* Notification Panel */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-96 bg-popover text-popover-foreground rounded-lg shadow-lg border border-border z-50">
          {/* Header */}
          <div className="p-4 border-b border-border">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">Notifications</h3>
              <button
                onClick={() => setIsOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Filter Tabs */}
            <div className="flex space-x-1 mt-3">
              {(['all', 'unread', 'read'] as const).map((filterOption) => (
                <button
                  key={filterOption}
                  onClick={() => setFilter(filterOption)}
                  className={`px-3 py-1 text-sm rounded-md capitalize ${
                    filter === filterOption
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  {filterOption}
                </button>
              ))}
            </div>

            {/* Actions */}
            {unreadCount > 0 && (
              <button
                onClick={() => markAllAsReadMutation.mutate()}
                disabled={markAllAsReadMutation.isPending}
                className="mt-2 text-sm text-primary hover:underline disabled:opacity-50"
              >
                Mark all as read
              </button>
            )}
          </div>

          {/* Notifications List */}
          <div className="max-h-96 overflow-y-auto">
            {isLoading ? (
              <div className="p-4 text-center text-muted-foreground">Loading notifications...</div>
            ) : notifications.length === 0 ? (
              <div className="p-4 text-center text-muted-foreground">
                {filter === 'unread' ? 'No unread notifications' : 'No notifications'}
              </div>
            ) : (
              <div className="divide-y divide-border">
                {notifications.map((notification: Notification) => (
                  <div
                    key={notification.id}
                    className={`p-4 hover:bg-muted/50 ${
                      notification.status === 'unread' ? 'bg-primary/5' : ''
                    }`}
                  >
                    <div className="flex items-start space-x-3">
                      {getNotificationIcon(notification.notification_type, notification.priority)}
                      
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <p className={`text-sm font-medium ${
                            notification.status === 'unread' ? '' : 'text-muted-foreground'
                          }`}>
                            {notification.title}
                          </p>
                          <span className="text-xs text-muted-foreground">
                            {formatTimeAgo(notification.created_at)}
                          </span>
                        </div>
                        
                        <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                          {notification.message}
                        </p>

                        {notification.event_type && (
                          <span className="inline-block mt-2 px-2 py-1 text-xs bg-muted text-muted-foreground rounded">
                            {notification.event_type}
                          </span>
                        )}

            {/* Actions */}
                        <div className="flex items-center space-x-2 mt-2">
                          {notification.status === 'unread' && (
                            <button
                              onClick={() => markAsReadMutation.mutate(notification.id)}
                              disabled={markAsReadMutation.isPending}
                              className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                            >
                              Mark as read
                            </button>
                          )}

              <button
                onClick={() => deleteNotificationMutation.mutate(notification.id)}
                disabled={deleteNotificationMutation.isPending}
                className="text-xs text-gray-600 hover:text-gray-800 disabled:opacity-50"
              >
                Archive
              </button>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="p-3 border-t border-gray-200">
            <button
              onClick={() => {
                setIsOpen(false)
                // Navigate to full notifications page
                window.location.href = '/notifications'
              }}
              className="w-full text-center text-sm text-blue-600 hover:text-blue-800"
            >
              View all notifications
            </button>
          </div>
        </div>
      )}

      {/* Backdrop */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setIsOpen(false)}
        />
      )}
    </div>
  )
}