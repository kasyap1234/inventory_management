import { tokenStorage } from './security';

export type WebSocketMessage = {
  type: 'inventory_update' | 'order_update' | 'notification' | 'low_stock_alert' | 'ping' | 'pong';
  data: any;
  timestamp: string;
};

export type WebSocketEventHandler = (message: WebSocketMessage) => void;

class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private handlers: Map<string, Set<WebSocketEventHandler>> = new Map();
  private pingInterval: NodeJS.Timeout | null = null;
  private url: string;
  private isIntentionallyClosed = false;

  constructor() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || `${protocol}//${window.location.hostname}:8080/v1/ws`;
    this.url = wsUrl;
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.isIntentionallyClosed = false;
    const token = tokenStorage.getAccessToken();
    
    if (!token) {
      console.warn('No access token available for WebSocket connection');
      return;
    }

    try {
      this.ws = new WebSocket(`${this.url}?token=${token}`);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.startPing();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          if (message.type === 'ping') {
            this.send({ type: 'pong', data: {}, timestamp: new Date().toISOString() });
            return;
          }

          const handlers = this.handlers.get(message.type);
          if (handlers) {
            handlers.forEach((handler) => handler(message));
          }

          // Also notify wildcard handlers
          const wildcardHandlers = this.handlers.get('*');
          if (wildcardHandlers) {
            wildcardHandlers.forEach((handler) => handler(message));
          }
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        this.stopPing();
        
        if (!this.isIntentionallyClosed) {
          this.reconnect();
        }
      };
    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      this.reconnect();
    }
  }

  private reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, delay);
  }

  private startPing() {
    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping', data: {}, timestamp: new Date().toISOString() });
      }
    }, 30000); // Ping every 30 seconds
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  send(message: WebSocketMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  on(type: string, handler: WebSocketEventHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
  }

  off(type: string, handler: WebSocketEventHandler) {
    const handlers = this.handlers.get(type);
    if (handlers) {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.handlers.delete(type);
      }
    }
  }

  disconnect() {
    this.isIntentionallyClosed = true;
    this.stopPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.handlers.clear();
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsClient = new WebSocketClient();

// React hook for WebSocket
export function useWebSocket(type: string, handler: WebSocketEventHandler) {
  if (typeof window === 'undefined') return;

  const handleMessage = (message: WebSocketMessage) => {
    handler(message);
  };

  // Subscribe
  wsClient.on(type, handleMessage);

  // Cleanup
  return () => {
    wsClient.off(type, handleMessage);
  };
}
