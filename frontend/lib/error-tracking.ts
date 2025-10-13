/**
 * Error Tracking Service
 * Integrates with error tracking services like Sentry, LogRocket, etc.
 */

interface ErrorContext {
  [key: string]: any;
}

interface ErrorTrackingConfig {
  dsn?: string;
  environment?: string;
  release?: string;
  enabled?: boolean;
}

// Define Sentry interface at runtime
interface SentryService {
  init(config: any): void;
  captureException(error: Error): void;
  captureMessage(message: string, level?: string): void;
  setUser(user: any): void;
  setContext(name: string, data: any): void;
  addBreadcrumb(breadcrumb: any): void;
  browserTracingIntegration?(): any;
  replayIntegration?(config?: any): any;
}

class ErrorTrackingService {
  private config: ErrorTrackingConfig = {
    enabled: false,
  };
  private initialized = false;
  private sentry: SentryService | null = null;

  /**
   * Initialize error tracking service
   * @param config Configuration for error tracking
   */
  init(config: ErrorTrackingConfig) {
    this.config = {
      ...this.config,
      ...config,
      enabled: config.enabled !== false && Boolean(config.dsn),
    };

    if (!this.config.enabled) {
      console.log('[Error Tracking] Disabled or DSN not configured');
      return;
    }

    // Initialize Sentry if DSN is provided
    if (typeof window !== 'undefined' && this.config.dsn) {
      this.initializeSentry();
    }

    this.initialized = true;
  }

  /**
   * Initialize Sentry (lazy loaded)
   */
  private async initializeSentry() {
    try {
      // Dynamic import to keep bundle size small
      // Note: @sentry/react must be installed: npm install @sentry/react
      const SentryModule = await (globalThis as any).import?.('@sentry/react').catch((e: any) => {
        console.log('[Error Tracking] @sentry/react not installed. Install with: npm install @sentry/react');
        console.debug('Import error:', e);
        return null;
      });
      
      if (!SentryModule) {
        this.config.enabled = false;
        return;
      }
      
      // Type assertion to handle dynamic import
      this.sentry = SentryModule.default || SentryModule;
      
      if (!this.sentry) {
        this.config.enabled = false;
        return;
      }
      
      this.sentry.init({
        dsn: this.config.dsn,
        environment: this.config.environment || process.env.NODE_ENV || 'development',
        release: this.config.release,
        integrations: [
          this.sentry.browserTracingIntegration?.() || {},
          this.sentry.replayIntegration?.({
            maskAllText: true,
            blockAllMedia: true,
          }) || {},
        ].filter(Boolean),
        // Performance Monitoring
        tracesSampleRate: this.config.environment === 'production' ? 0.1 : 1.0,
        // Session Replay
        replaysSessionSampleRate: 0.1,
        replaysOnErrorSampleRate: 1.0,
        // Ignore common non-critical errors
        ignoreErrors: [
          // Browser extensions
          'top.GLOBALS',
          // Network errors
          'NetworkError',
          'Failed to fetch',
          // Random plugins/extensions
          'Non-Error promise rejection captured',
        ],
        beforeSend(event: any, hint: any) {
          // Filter out errors from browser extensions
          if (event.exception) {
            const firstException = event.exception.values?.[0];
            if (firstException?.stacktrace?.frames) {
              const frames = firstException.stacktrace.frames;
              if (frames.some((frame: any) => frame.filename?.includes('chrome-extension://'))) {
                return null;
              }
            }
          }
          return event;
        },
      });

      console.log('[Error Tracking] Sentry initialized');
    } catch (error) {
      console.error('[Error Tracking] Failed to initialize Sentry:', error);
      console.log('[Error Tracking] Falling back to console logging only');
    }
  }

  /**
   * Capture an exception
   * @param error Error object
   * @param context Additional context
   */
  captureException(error: Error, context?: ErrorContext) {
    if (!this.config.enabled) {
      console.error('[Error Tracking] Exception (not sent):', error, context);
      return;
    }

    if (this.sentry) {
      if (context) {
        this.sentry.setContext('additional', context);
      }
      this.sentry.captureException(error);
    } else {
      // Fallback to local logging
      console.error('[Error Tracking] Exception (Sentry not available):', error, context);
    }
  }

  /**
   * Capture a message
   * @param message Message string
   * @param level Severity level
   * @param context Additional context
   */
  captureMessage(
    message: string,
    level: 'info' | 'warning' | 'error' = 'info',
    context?: ErrorContext
  ) {
    if (!this.config.enabled) {
      console.log(`[Error Tracking] ${level.toUpperCase()}: ${message}`, context);
      return;
    }

    if (this.sentry) {
      if (context) {
        this.sentry.setContext('additional', context);
      }
      this.sentry.captureMessage(message, level);
    } else {
      // Fallback to local logging
      console.log(`[Error Tracking] ${level.toUpperCase()} (Sentry not available): ${message}`, context);
    }
  }

  /**
   * Set user context
   * @param user User information
   */
  setUser(user: { id?: string; email?: string; username?: string } | null) {
    if (!this.config.enabled) {
      return;
    }

    if (this.sentry) {
      this.sentry.setUser(user);
    }
  }

  /**
   * Set custom context
   * @param name Context name
   * @param data Context data
   */
  setContext(name: string, data: ErrorContext) {
    if (!this.config.enabled) {
      return;
    }

    if (this.sentry) {
      this.sentry.setContext(name, data);
    }
  }

  /**
   * Add breadcrumb for debugging
   * @param message Breadcrumb message
   * @param category Category
   * @param data Additional data
   */
  addBreadcrumb(message: string, category?: string, data?: ErrorContext) {
    if (!this.config.enabled) {
      return;
    }

    if (this.sentry) {
      this.sentry.addBreadcrumb({
        message,
        category,
        data,
        level: 'info',
      });
    }
  }
}

// Export singleton instance
export const errorTracking = new ErrorTrackingService();

// Initialize with environment variables
if (typeof window !== 'undefined') {
  errorTracking.init({
    dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
    environment: process.env.NEXT_PUBLIC_ENV || process.env.NODE_ENV,
    release: process.env.NEXT_PUBLIC_APP_VERSION,
    enabled: process.env.NEXT_PUBLIC_ENABLE_ERROR_TRACKING !== 'false',
  });
}

/**
 * React Error Boundary integration
 */
export function logErrorToService(error: Error, errorInfo: React.ErrorInfo) {
  errorTracking.captureException(error, {
    react: {
      componentStack: errorInfo.componentStack,
    },
  });
}

/**
 * Manually trigger Sentry loading (optional)
 * Call this from your app code to load Sentry on demand
 */
export async function loadSentryManually() {
  try {
    const SentryModule = await (globalThis as any).import?.('@sentry/react');
    return SentryModule?.default || SentryModule;
  } catch (error) {
    console.log('[Error Tracking] Sentry not available');
    return null;
  }
}
