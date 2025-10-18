'use client';

import { useEffect } from 'react';
import { wsClient } from '@/lib/websocket';
import { useAuth } from '@/hooks/useAuth';
import toast from 'react-hot-toast';

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  useEffect(() => {
    if (!isAuthenticated) {
      wsClient.disconnect();
      return;
    }

    // Connect to WebSocket
    wsClient.connect();

    // Listen for notifications
    const unsubscribeNotification = wsClient.on('notification', (message) => {
      toast(message.data.message || 'New notification', {
        icon: '🔔',
        duration: 5000,
      });
    });

    // Listen for low stock alerts
    const unsubscribeLowStock = wsClient.on('low_stock_alert', (message) => {
      toast.error(`Low stock alert: ${message.data.product_name}`, {
        duration: 7000,
      });
    });

    // Cleanup on unmount
    return () => {
      if (unsubscribeNotification) unsubscribeNotification();
      if (unsubscribeLowStock) unsubscribeLowStock();
    };
  }, [isAuthenticated]);

  return <>{children}</>;
}
