'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { CheckCircle, Trash2, Settings } from 'lucide-react';
import { format } from 'date-fns';

interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
  is_read: boolean;
  created_at: string;
  action_url?: string;
}

interface NotificationManagerProps {
  notifications: Notification[];
  isLoading?: boolean;
  onMarkAsRead?: (notificationId: string) => Promise<void>;
  onDelete?: (notificationId: string) => Promise<void>;
  onOpenSettings?: () => void;
}

const getTypeColor = (type: string) => {
  switch (type) {
    case 'success':
      return 'bg-green-100 text-green-800';
    case 'warning':
      return 'bg-yellow-100 text-yellow-800';
    case 'error':
      return 'bg-red-100 text-red-800';
    case 'info':
    default:
      return 'bg-blue-100 text-blue-800';
  }
};

export function NotificationManager({
  notifications,
  isLoading,
  onMarkAsRead,
  onDelete,
  onOpenSettings,
}: NotificationManagerProps) {
  const [loadingIds, setLoadingIds] = useState<Set<string>>(new Set());

  const handleMarkAsRead = async (notificationId: string) => {
    setLoadingIds((prev) => new Set(prev).add(notificationId));
    try {
      if (onMarkAsRead) {
        await onMarkAsRead(notificationId);
      }
    } finally {
      setLoadingIds((prev) => {
        const next = new Set(prev);
        next.delete(notificationId);
        return next;
      });
    }
  };

  const handleDelete = async (notificationId: string) => {
    setLoadingIds((prev) => new Set(prev).add(notificationId));
    try {
      if (onDelete) {
        await onDelete(notificationId);
      }
    } finally {
      setLoadingIds((prev) => {
        const next = new Set(prev);
        next.delete(notificationId);
        return next;
      });
    }
  };

  const unreadCount = notifications.filter((n) => !n.is_read).length;

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>Loading notifications...</CardDescription>
        </CardHeader>
        <CardContent className="h-96 flex items-center justify-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>
            {unreadCount > 0 ? `${unreadCount} unread notification${unreadCount !== 1 ? 's' : ''}` : 'All caught up!'}
          </CardDescription>
        </div>
        {onOpenSettings && (
          <Button variant="outline" size="sm" onClick={onOpenSettings}>
            <Settings className="w-4 h-4 mr-2" />
            Settings
          </Button>
        )}
      </CardHeader>
      <CardContent>
        {notifications.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <p>No notifications yet</p>
          </div>
        ) : (
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {notifications.map((notification) => (
              <div
                key={notification.id}
                className={`p-4 rounded-lg border ${
                  notification.is_read
                    ? 'bg-white border-gray-200'
                    : 'bg-blue-50 border-blue-200'
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h4 className="font-semibold text-sm">{notification.title}</h4>
                      <Badge className={getTypeColor(notification.type)}>
                        {notification.type}
                      </Badge>
                      {!notification.is_read && (
                        <div className="w-2 h-2 rounded-full bg-blue-600" />
                      )}
                    </div>
                    <p className="text-sm text-gray-700 mb-2">{notification.message}</p>
                    <p className="text-xs text-gray-500">
                      {format(new Date(notification.created_at), 'MMM d, yyyy HH:mm')}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {!notification.is_read && onMarkAsRead && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleMarkAsRead(notification.id)}
                        disabled={loadingIds.has(notification.id)}
                        className="h-8 w-8 p-0"
                      >
                        <CheckCircle className="w-4 h-4 text-green-600" />
                      </Button>
                    )}
                    {onDelete && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDelete(notification.id)}
                        disabled={loadingIds.has(notification.id)}
                        className="h-8 w-8 p-0 text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
