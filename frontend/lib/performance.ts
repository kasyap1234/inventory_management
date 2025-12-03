/**
 * Performance monitoring utilities
 */

import React from 'react';

interface PerformanceMetric {
  name: string;
  duration: number;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

class PerformanceMonitor {
  private metrics: PerformanceMetric[] = [];
  private timers: Map<string, number> = new Map();

  /**
   * Start timing an operation
   */
  startTimer(name: string, metadata?: Record<string, unknown>) {
    this.timers.set(name, performance.now());
    if (metadata) {
      this.timers.set(`${name}_metadata`, metadata as unknown);
    }
  }

  /**
   * End timing an operation and record the metric
   */
  endTimer(name: string): number | null {
    const startTime = this.timers.get(name);
    if (!startTime) {
      console.warn(`Timer "${name}" was not started`);
      return null;
    }

    const duration = performance.now() - startTime;
    const metadata = this.timers.get(`${name}_metadata`) as Record<string, unknown> | undefined;

    const metric: PerformanceMetric = {
      name,
      duration,
      timestamp: Date.now(),
      metadata,
    };

    this.metrics.push(metric);
    this.timers.delete(name);
    this.timers.delete(`${name}_metadata`);

    // Log slow operations in development
    if (process.env.NODE_ENV === 'development' && duration > 1000) {
      console.warn(`Slow operation detected: ${name} took ${duration.toFixed(2)}ms`, metadata);
    }

    return duration;
  }

  /**
   * Get all recorded metrics
   */
  getMetrics(): PerformanceMetric[] {
    return [...this.metrics];
  }

  /**
   * Get metrics for a specific operation
   */
  getMetricsByName(name: string): PerformanceMetric[] {
    return this.metrics.filter((m) => m.name === name);
  }

  /**
   * Get average duration for an operation
   */
  getAverageDuration(name: string): number {
    const metrics = this.getMetricsByName(name);
    if (metrics.length === 0) return 0;

    const total = metrics.reduce((sum, m) => sum + m.duration, 0);
    return total / metrics.length;
  }

  /**
   * Clear all metrics
   */
  clearMetrics() {
    this.metrics = [];
  }

  /**
   * Get performance summary
   */
  getSummary(): Record<string, { count: number; avgDuration: number; maxDuration: number }> {
    const summary: Record<string, { count: number; avgDuration: number; maxDuration: number }> = {};

    this.metrics.forEach((metric) => {
      if (!summary[metric.name]) {
        summary[metric.name] = {
          count: 0,
          avgDuration: 0,
          maxDuration: 0,
        };
      }

      const current = summary[metric.name];
      current.count++;
      current.avgDuration = (current.avgDuration * (current.count - 1) + metric.duration) / current.count;
      current.maxDuration = Math.max(current.maxDuration, metric.duration);
    });

    return summary;
  }

  /**
   * Export metrics for analysis
   */
  exportMetrics(): string {
    return JSON.stringify(
      {
        metrics: this.metrics,
        summary: this.getSummary(),
        exportedAt: new Date().toISOString(),
      },
      null,
      2
    );
  }
}

// Singleton instance
export const performanceMonitor = new PerformanceMonitor();

/**
 * Decorator to measure function execution time
 */
export function measurePerformance(name?: string) {
  return function (target: object, propertyKey: string, descriptor: PropertyDescriptor) {
    const originalMethod = descriptor.value;
    const metricName = name || `${target.constructor.name}.${propertyKey}`;

    descriptor.value = async function (...args: unknown[]) {
      performanceMonitor.startTimer(metricName);
      try {
        const result = await originalMethod.apply(this, args);
        return result;
      } finally {
        performanceMonitor.endTimer(metricName);
      }
    };

    return descriptor;
  };
}

/**
 * HOC to measure React component render time
 */
export function withPerformanceTracking<P extends object>(
  Component: React.ComponentType<P>,
  componentName?: string
) {
  const name = componentName || Component.displayName || Component.name || 'Component';

  return function PerformanceTrackedComponent(props: P) {
    const renderName = `${name}_render`;

    React.useEffect(() => {
      performanceMonitor.startTimer(renderName);
      return () => {
        performanceMonitor.endTimer(renderName);
      };
    });

    return React.createElement(Component, props);
  };
}

/**
 * Hook to measure component lifecycle
 */
export function usePerformanceTracking(componentName: string) {
  React.useEffect(() => {
    const mountName = `${componentName}_mount`;
    performanceMonitor.startTimer(mountName);

    return () => {
      performanceMonitor.endTimer(mountName);
    };
  }, [componentName]);

  React.useEffect(() => {
    const renderName = `${componentName}_render`;
    performanceMonitor.startTimer(renderName);
    performanceMonitor.endTimer(renderName);
  });
}

/**
 * Measure API call performance
 */
export async function measureApiCall<T>(
  name: string,
  apiCall: () => Promise<T>,
  metadata?: Record<string, unknown>
): Promise<T> {
  performanceMonitor.startTimer(name, metadata);
  try {
    const result = await apiCall();
    return result;
  } finally {
    const duration = performanceMonitor.endTimer(name);

    // Log slow API calls
    if (duration && duration > 3000) {
      console.warn(`Slow API call: ${name} took ${duration.toFixed(2)}ms`, metadata);
    }
  }
}

/**
 * Get Web Vitals metrics
 */
export function getWebVitals() {
  if (typeof window === 'undefined' || !('performance' in window)) {
    return null;
  }

  const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
  const paint = performance.getEntriesByType('paint');

  return {
    // Time to First Byte
    ttfb: navigation?.responseStart - navigation?.requestStart,

    // First Contentful Paint
    fcp: paint.find((entry) => entry.name === 'first-contentful-paint')?.startTime,

    // Largest Contentful Paint (requires observer)
    // lcp: ...,

    // DOM Content Loaded
    domContentLoaded: navigation?.domContentLoadedEventEnd - navigation?.domContentLoadedEventStart,

    // Load Complete
    loadComplete: navigation?.loadEventEnd - navigation?.loadEventStart,

    // Total Page Load Time
    pageLoadTime: navigation?.loadEventEnd - navigation?.fetchStart,
  };
}

/**
 * Report performance metrics to analytics
 */
export function reportPerformanceMetrics() {
  if (process.env.NODE_ENV === 'production') {
    const summary = performanceMonitor.getSummary();
    const webVitals = getWebVitals();

    // Send to analytics service
    // analytics.track('performance_metrics', { summary, webVitals });

    console.log('Performance Summary:', summary);
    console.log('Web Vitals:', webVitals);
  }
}

// Auto-report metrics every 5 minutes in production
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'production') {
  setInterval(reportPerformanceMetrics, 5 * 60 * 1000);
}
