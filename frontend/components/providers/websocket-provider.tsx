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

    // Define handlers
    const handleNotification = (message: any) => {
      toast(message.data.message || 'New notification', {
        icon: '🔔',
        duration: 5000,
      });
    };

    const handleLowStock = (message: any) => {
      toast.error(`Low stock alert: ${message.data.product_name}`, {
        duration: 7000,
      });
    };

    // Listen for events
    wsClient.on('notification', handleNotification);
    wsClient.on('low_stock_alert', handleLowStock);

    // Cleanup on unmount
    return () => {
      wsClient.off('notification', handleNotification);
      wsClient.off('low_stock_alert', handleLowStock);
    };
  }, [isAuthenticated]);

  return <>{children}</>;
}
