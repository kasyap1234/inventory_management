'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent } from '@/components/ui/card';
import { Bell, Check, Trash2, Mail, MessageSquare } from 'lucide-react';
import { notificationService, type NotificationDto, type NotificationListResult } from '@/lib/services';
import { Button } from '@/components/ui/button';
import { formatDistance } from 'date-fns';

export default function NotificationsPage() {
  const queryClient = useQueryClient();

  const { data: notificationsData, isLoading } = useQuery<NotificationListResult>({
    queryKey: ['notifications'],
    queryFn: () => notificationService.list(),
  });

  const notifications: NotificationDto[] = notificationsData?.items ?? [];

  const markAsReadMutation = useMutation({
    mutationFn: (id: string) => notificationService.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => notificationService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'email':
        return <Mail className="h-5 w-5 text-blue-600" />;
      case 'sms':
        return <MessageSquare className="h-5 w-5 text-green-600" />;
      default:
        return <Bell className="h-5 w-5 text-purple-600" />;
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
            Notifications
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Stay updated with all your notifications
          </p>
        </div>
      </div>

      {/* Notifications List */}
      <div className="space-y-4">
        {isLoading ? (
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-24 bg-gray-200 rounded-lg animate-pulse"></div>
            ))}
          </div>
        ) : notifications.length > 0 ? (
          notifications.map((notification) => {
            const isRead = notification.status === 'read';

            return (
              <Card
                key={notification.id}
                className={`border-0 shadow-md hover:shadow-lg transition-all ${
                  isRead ? 'bg-white' : 'bg-blue-50 border-l-4 border-blue-600'
                }`}
              >
                <CardContent className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-4 flex-1">
                      <div className={`p-3 rounded-full ${
                        isRead ? 'bg-gray-100' : 'bg-blue-100'
                      }`}>
                        {getNotificationIcon(notification.type)}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-semibold text-foreground">
                            {notification.subject || 'Notification'}
                          </h3>
                          <span className="px-2 py-0.5 text-xs font-medium bg-gray-200 text-foreground rounded-full">
                            {notification.status || 'pending'}
                          </span>
                        </div>
                        <p className="text-foreground mb-2">{notification.body}</p>
                        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
                          <span className="capitalize">Channel: {notification.type}</span>
                          {notification.event_type && (
                            <span className="capitalize">Event: {notification.event_type}</span>
                          )}
                          {notification.recipient && (
                            <span>Recipient: {notification.recipient}</span>
                          )}
                          <span>
                            {formatDistance(new Date(notification.created_at), new Date(), {
                              addSuffix: true,
                            })}
                          </span>
                        </div>
                        {notification.error && (
                          <p className="mt-2 text-sm text-red-600">Error: {notification.error}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-2 ml-4">
                      {!isRead && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => markAsReadMutation.mutate(notification.id)}
                          className="text-blue-600 hover:text-blue-700 hover:bg-blue-50"
                        >
                          <Check className="h-4 w-4 mr-1" />
                          Mark Read
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => deleteMutation.mutate(notification.id)}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })
        ) : (
          <Card className="border-0 shadow-md">
            <CardContent className="py-16">
              <div className="text-center">
                <Bell className="h-16 w-16 text-muted-foreground/50 mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-foreground mb-2">
                  No notifications yet
                </h3>
                <p className="text-muted-foreground">
                  When you receive notifications, they&rsquo;ll appear here
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
