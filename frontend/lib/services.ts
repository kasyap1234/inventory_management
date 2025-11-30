import { api } from './api';

export type AnalyticsDashboard = {
  tenantId: string | null;
  totalSales: number;
  totalStockValue: number;
  gstCollected: number;
  orderCount: number;
  lowStockItems: number;
  lastUpdated: string | null;
};

type RawTrend = {
  date?: string;
  Date?: string;
  sales_amount?: number;
  SalesAmount?: number;
  order_count?: number;
  OrderCount?: number;
};

export type AnalyticsTrendPoint = {
  date: string | null;
  salesAmount: number;
  orderCount: number;
};

export type AnalyticsSalesTrends = {
  startDate: string | null;
  endDate: string | null;
  trends: AnalyticsTrendPoint[];
};

type RawProductSales = {
  product_id?: string;
  ProductID?: string;
  product_name?: string;
  ProductName?: string;
  total_sales?: number;
  TotalSales?: number;
  units_sold?: number;
  UnitsSold?: number;
  order_count?: number;
  OrderCount?: number;
};

export type ProductSales = {
  productId: string;
  productName: string;
  totalSales: number;
  unitsSold: number;
  orderCount: number;
};

type RawLowStockItem = {
  product_id?: string;
  ProductID?: string;
  product_name?: string;
  ProductName?: string;
  warehouse_id?: string;
  WarehouseID?: string;
  current_stock?: number;
  CurrentStock?: number;
  threshold?: number;
  Threshold?: number;
  unit_price?: number;
  UnitPrice?: number;
  stock_value?: number;
  StockValue?: number;
};

export type LowStockItem = {
  productId: string;
  productName: string;
  warehouseId: string;
  currentStock: number;
  threshold: number;
  unitPrice: number;
  stockValue: number;
};

type RawInventoryValuation = {
  tenant_id?: string;
  tenantId?: string;
  total_value?: number;
  totalValue?: number;
  total_items?: number;
  totalItems?: number;
  total_quantity?: number;
  totalQuantity?: number;
  by_warehouse?: Record<string, number>;
  byWarehouse?: Record<string, number>;
  by_category?: Record<string, number>;
  byCategory?: Record<string, number>;
  last_calculated?: string;
  lastCalculated?: string;
};

export type InventoryValuation = {
  tenantId: string | null;
  totalValue: number;
  totalItems: number;
  totalQuantity: number;
  byWarehouse: Record<string, number>;
  byCategory: Record<string, number>;
  lastCalculated: string | null;
};

export type RevenueByCategory = {
  categoryId: string;
  totalRevenue: number;
};

export type OrderStatusEntry = {
  status: string;
  count: number;
  percent: number;
};

type RawSubscription = {
  id: string;
  plan_name: string;
  amount: number | string;
  currency: string;
  status: string;
  start_date: string;
  end_date?: string | null;
  razorpay_subscription_id?: string | null;
};

export type SubscriptionDto = {
  id: string;
  plan_name: string;
  amount: number;
  currency: string;
  status: string;
  start_date: string;
  end_date?: string | null;
  razorpay_subscription_id?: string | null;
};

export type SubscriptionListResult = {
  items: SubscriptionDto[];
  limit: number;
  offset: number;
};

export type SubscriptionPlanConfig = {
  id: string;
  name: string;
  amount: number;
  currency: string;
  interval: string;
  description?: string;
  features?: string[];
};

type RawNotification = {
  id: string;
  type: string;
  status: string;
  recipient: string;
  subject?: string | null;
  body: string;
  event_type?: string | null;
  error?: string | null;
  created_at: string;
};

export type NotificationDto = RawNotification;

export type NotificationListResult = {
  items: NotificationDto[];
  count: number;
};

export type AuditLogEntry = {
  id: string;
  table_name: string;
  record_id?: string | null;
  action: string;
  new_values?: Record<string, unknown> | null;
  old_values?: Record<string, unknown> | null;
  changed_by?: string | null;
  created_at: string;
};

type RawAuditLogEntry = {
  id: string;
  table_name: string;
  record_id?: string | null;
  action: string;
  new_values?: Record<string, unknown> | null;
  old_values?: Record<string, unknown> | null;
  changed_by?: string | null;
  created_at: string;
};

export type AuditLogListResult = {
  items: AuditLogEntry[];
  total: number;
  limit: number;
  offset: number;
  page: number;
};

export type AuditSummary = {
  total_logs: number;
  table_breakdown?: Record<string, number>;
  action_breakdown?: Record<string, number>;
  user_activity?: Record<string, number>;
  period_start?: string;
  period_end?: string;
};

// Analytics Services
export const analyticsService = {
  getDashboardAnalytics: async (): Promise<AnalyticsDashboard> => {
    const { data } = await api.get('/analytics/dashboard');
    return {
      tenantId: data?.tenant_id ?? null,
      totalSales: data?.total_sales ?? 0,
      totalStockValue: data?.total_stock_value ?? 0,
      gstCollected: data?.gst_collected ?? 0,
      orderCount: data?.order_count ?? 0,
      lowStockItems: data?.low_stock_items ?? 0,
      lastUpdated: data?.last_updated ?? null,
    };
  },
  getSalesTrends: async (params?: { start_date?: string; end_date?: string }): Promise<AnalyticsSalesTrends> => {
    const { data } = await api.get('/analytics/sales-trends', { params });

    const trendsSource: RawTrend[] = Array.isArray(data?.trends) ? data.trends : [];

    const trends: AnalyticsTrendPoint[] = trendsSource
      .map((trend) => {
        const date = trend.date ?? null;
        const salesAmount = trend.sales_amount ?? 0;
        const orderCount = trend.order_count ?? 0;
        return { date, salesAmount, orderCount };
      })
      .sort((a, b) => {
        const aTime = a.date ? new Date(a.date).getTime() : 0;
        const bTime = b.date ? new Date(b.date).getTime() : 0;
        return aTime - bTime;
      });

    return {
      startDate: data?.start_date ?? null,
      endDate: data?.end_date ?? null,
      trends,
    };
  },
  getGSTTotals: async (params?: { start_date?: string; end_date?: string }): Promise<{
    cgst: number;
    sgst: number;
    igst: number;
    total: number;
  }> => {
    const { data } = await api.get('/analytics/gst-totals', { params });
    return {
      cgst: data?.cgst ?? 0,
      sgst: data?.sgst ?? 0,
      igst: data?.igst ?? 0,
      total: data?.total ?? 0,
    };
  },
  getTopProducts: async (params?: {
    limit?: number;
    start_date?: string;
    end_date?: string;
  }): Promise<ProductSales[]> => {
    const { data } = await api.get('/analytics/top-products', { params });

    const productsSource: RawProductSales[] = Array.isArray(data?.products) ? data.products : [];

    return productsSource.map((product) => ({
      productId: product.product_id ?? '',
      productName: product.product_name ?? '',
      totalSales: product.total_sales ?? 0,
      unitsSold: product.units_sold ?? 0,
      orderCount: product.order_count ?? 0,
    }));
  },
  getLowStockReport: async (params?: { threshold?: number }): Promise<LowStockItem[]> => {
    const { data } = await api.get('/analytics/low-stock', { params });

    const itemsSource: RawLowStockItem[] = Array.isArray(data?.low_stock_items) ? data.low_stock_items : [];

    return itemsSource.map((item) => ({
      productId: item.product_id ?? '',
      productName: item.product_name ?? '',
      warehouseId: item.warehouse_id ?? '',
      currentStock: item.current_stock ?? 0,
      threshold: item.threshold ?? params?.threshold ?? 0,
      unitPrice: item.unit_price ?? 0,
      stockValue: item.stock_value ?? 0,
    }));
  },
  getInventoryValuation: async (): Promise<InventoryValuation> => {
    const { data } = await api.get('/analytics/inventory-valuation');

    const payload = data as RawInventoryValuation;

    return {
      tenantId: payload.tenant_id ?? null,
      totalValue: payload.total_value ?? 0,
      totalItems: payload.total_items ?? 0,
      totalQuantity: payload.total_quantity ?? 0,
      byWarehouse: payload.by_warehouse ?? {},
      byCategory: payload.by_category ?? {},
      lastCalculated: payload.last_calculated ?? null,
    };
  },
  getRevenueByCategory: async (params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<RevenueByCategory[]> => {
    const { data } = await api.get('/analytics/revenue-by-category', { params });

    const revenueMap = data?.revenue_by_category ?? {};

    return Object.entries(revenueMap).map(([categoryId, total]) => ({
      categoryId,
      totalRevenue: typeof total === 'number' ? total : Number(total) || 0,
    }));
  },
  getOrderStatusDistribution: async (params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<OrderStatusEntry[]> => {
    const { data } = await api.get('/analytics/order-status', { params });

    const statusMap = (data as Record<string, unknown>) ?? {};
    const statusEntries = Object.entries(statusMap).map(([status, count]) => ({
      status,
      count: typeof count === 'number' ? count : Number(count) || 0,
    }));

    const total = statusEntries.reduce((acc, entry) => acc + entry.count, 0);

    return statusEntries.map((entry) => ({
      ...entry,
      percent: total > 0 ? entry.count / total : 0,
    }));
  },
  refreshAnalytics: async (): Promise<void> => {
    await api.post('/analytics/refresh');
  },

  // Combined analytics endpoint - fetches all analytics data in a single API call
  // This significantly reduces page load time by eliminating multiple round trips
  getCombinedAnalytics: async (params?: {
    start_date?: string;
    end_date?: string;
    top_products_limit?: number;
    low_stock_threshold?: number;
  }): Promise<CombinedAnalyticsData> => {
    const { data } = await api.get('/analytics/combined', { params });

    // Parse dashboard data
    const dashboardRaw = data?.dashboard ?? {};
    const dashboard: AnalyticsDashboard = {
      tenantId: dashboardRaw?.tenant_id ?? null,
      totalSales: dashboardRaw?.total_sales ?? 0,
      totalStockValue: dashboardRaw?.total_stock_value ?? 0,
      gstCollected: dashboardRaw?.gst_collected ?? 0,
      orderCount: dashboardRaw?.order_count ?? 0,
      lowStockItems: dashboardRaw?.low_stock_items ?? 0,
      lastUpdated: dashboardRaw?.last_updated ?? null,
    };

    // Parse sales trends
    const salesTrendsRaw = data?.sales_trends ?? {};
    const trendsSource: RawTrend[] = Array.isArray(salesTrendsRaw?.trends) ? salesTrendsRaw.trends : [];
    const salesTrends: AnalyticsSalesTrends = {
      startDate: salesTrendsRaw?.start_date ?? null,
      endDate: salesTrendsRaw?.end_date ?? null,
      trends: trendsSource.map((trend) => ({
        date: trend.date ?? null,
        salesAmount: trend.sales_amount ?? 0,
        orderCount: trend.order_count ?? 0,
      })).sort((a, b) => {
        const aTime = a.date ? new Date(a.date).getTime() : 0;
        const bTime = b.date ? new Date(b.date).getTime() : 0;
        return aTime - bTime;
      }),
    };

    // Parse top products
    const topProductsRaw = data?.top_products ?? {};
    const productsSource: RawProductSales[] = Array.isArray(topProductsRaw?.products) ? topProductsRaw.products : [];
    const topProducts: ProductSales[] = productsSource.map((product) => ({
      productId: product.product_id ?? '',
      productName: product.product_name ?? '',
      totalSales: product.total_sales ?? 0,
      unitsSold: product.units_sold ?? 0,
      orderCount: product.order_count ?? 0,
    }));

    // Parse low stock
    const lowStockRaw = data?.low_stock ?? {};
    const lowStockSource: RawLowStockItem[] = Array.isArray(lowStockRaw?.low_stock_items) ? lowStockRaw.low_stock_items : [];
    const lowStock: LowStockItem[] = lowStockSource.map((item) => ({
      productId: item.product_id ?? '',
      productName: item.product_name ?? '',
      warehouseId: item.warehouse_id ?? '',
      currentStock: item.current_stock ?? 0,
      threshold: item.threshold ?? params?.low_stock_threshold ?? 0,
      unitPrice: item.unit_price ?? 0,
      stockValue: item.stock_value ?? 0,
    }));

    // Parse inventory valuation
    const inventoryRaw = data?.inventory_valuation ?? {};
    const inventoryValuation: InventoryValuation = {
      tenantId: inventoryRaw.tenant_id ?? null,
      totalValue: inventoryRaw.total_value ?? 0,
      totalItems: inventoryRaw.total_items ?? 0,
      totalQuantity: inventoryRaw.total_quantity ?? 0,
      byWarehouse: inventoryRaw.by_warehouse ?? {},
      byCategory: inventoryRaw.by_category ?? {},
      lastCalculated: inventoryRaw.last_calculated ?? null,
    };

    // Parse revenue by category
    const revenueRaw = data?.revenue_by_category ?? {};
    const revenueMap = revenueRaw?.revenue_by_category ?? {};
    const revenueByCategory: RevenueByCategory[] = Object.entries(revenueMap).map(([categoryId, total]) => ({
      categoryId,
      totalRevenue: typeof total === 'number' ? total : Number(total) || 0,
    }));

    // Parse order status distribution
    const orderStatusRaw = data?.order_status ?? {};
    const statusEntries = Object.entries(orderStatusRaw).map(([status, count]) => ({
      status,
      count: typeof count === 'number' ? count : Number(count) || 0,
    }));
    const total = statusEntries.reduce((acc, entry) => acc + entry.count, 0);
    const orderStatus: OrderStatusEntry[] = statusEntries.map((entry) => ({
      ...entry,
      percent: total > 0 ? entry.count / total : 0,
    }));

    // Parse GST totals
    const gstRaw = data?.gst_totals ?? {};
    const gstTotals = {
      cgst: gstRaw?.cgst ?? 0,
      sgst: gstRaw?.sgst ?? 0,
      igst: gstRaw?.igst ?? 0,
      total: gstRaw?.total ?? 0,
    };

    return {
      dashboard,
      salesTrends,
      topProducts,
      lowStock,
      inventoryValuation,
      revenueByCategory,
      orderStatus,
      gstTotals,
      fetchedAt: data?.fetched_at ?? new Date().toISOString(),
    };
  },
};

// Combined analytics data type
export type CombinedAnalyticsData = {
  dashboard: AnalyticsDashboard;
  salesTrends: AnalyticsSalesTrends;
  topProducts: ProductSales[];
  lowStock: LowStockItem[];
  inventoryValuation: InventoryValuation;
  revenueByCategory: RevenueByCategory[];
  orderStatus: OrderStatusEntry[];
  gstTotals: { cgst: number; sgst: number; igst: number; total: number };
  fetchedAt: string;
};

// Subscription Services
export const subscriptionService = {
  list: async (params?: { limit?: number; offset?: number }): Promise<SubscriptionListResult> => {
    const { data } = await api.get('/subscriptions', { params });
    const rawItems: RawSubscription[] = Array.isArray(data?.subscriptions) ? data.subscriptions : [];

    const items: SubscriptionDto[] = rawItems.map((subscription) => ({
      id: subscription.id,
      plan_name: subscription.plan_name,
      amount: Number(subscription.amount ?? 0),
      currency: subscription.currency,
      status: subscription.status,
      start_date: subscription.start_date,
      end_date: subscription.end_date ?? null,
      razorpay_subscription_id: subscription.razorpay_subscription_id ?? null,
    }));

    const limit = typeof data?.limit === 'number' ? data.limit : params?.limit ?? items.length;
    const offset = typeof data?.offset === 'number' ? data.offset : params?.offset ?? 0;

    return {
      items,
      limit,
      offset,
    };
  },
  getById: async (id: string): Promise<SubscriptionDto> => {
    const { data } = await api.get(`/subscriptions/${id}`);
    const payload = data as RawSubscription;
    return {
      id: payload.id,
      plan_name: payload.plan_name,
      amount: Number(payload.amount ?? 0),
      currency: payload.currency,
      status: payload.status,
      start_date: payload.start_date,
      end_date: payload.end_date ?? null,
      razorpay_subscription_id: payload.razorpay_subscription_id ?? null,
    };
  },
  create: async (payload: Record<string, unknown>): Promise<SubscriptionDto> => {
    const { data } = await api.post('/subscriptions', payload);
    return subscriptionService.getById(data?.id ?? '');
  },
  updatePlan: async (id: string, payload: Record<string, unknown>): Promise<SubscriptionDto> => {
    await api.put(`/subscriptions/${id}`, payload);
    return subscriptionService.getById(id);
  },
  cancel: async (id: string): Promise<void> => {
    await api.post(`/subscriptions/${id}/cancel`);
  },
  pause: async (id: string): Promise<void> => {
    await api.post(`/subscriptions/${id}/pause`);
  },
  resume: async (id: string): Promise<void> => {
    await api.post(`/subscriptions/${id}/resume`);
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/subscriptions/${id}`);
  },
  getAvailablePlans: async (): Promise<SubscriptionPlanConfig[]> => {
    const { data } = await api.get('/subscriptions/plans');
    const plansSource = data?.plans ?? {};

    // Backend returns object with plan IDs as keys
    return Object.entries(plansSource as Record<string, any>).map(([id, plan]) => ({
      id: id,
      name: plan.name,
      amount: plan.amount,
      currency: plan.currency,
      interval: plan.interval,
      description: plan.description,
      features: plan.features ?? [],
    }));
  },
};

// Notification Services
export const notificationService = {
  list: async (params?: { type?: string; event_type?: string; status?: string }): Promise<NotificationListResult> => {
    const { data } = await api.get('/notifications', { params });
    const rawItems: RawNotification[] = Array.isArray(data?.notifications) ? data.notifications : [];

    const items: NotificationDto[] = rawItems.map((notification) => ({
      id: notification.id,
      type: notification.type,
      status: notification.status,
      recipient: notification.recipient,
      subject: notification.subject ?? null,
      body: notification.body,
      event_type: notification.event_type ?? null,
      error: notification.error ?? null,
      created_at: notification.created_at,
    }));

    return {
      items,
      count: data?.count ?? items.length,
    };
  },
  getById: async (id: string): Promise<NotificationDto> => {
    const { data } = await api.get(`/notifications/${id}`);
    const payload = data as RawNotification;
    return {
      id: payload.id,
      type: payload.type,
      status: payload.status,
      recipient: payload.recipient,
      subject: payload.subject ?? null,
      body: payload.body,
      event_type: payload.event_type ?? null,
      error: payload.error ?? null,
      created_at: payload.created_at,
    };
  },
  send: async (payload: {
    type: string;
    eventType: string;
    recipient: string;
    subject?: string;
    body: string;
    eventId?: string;
  }): Promise<NotificationDto> => {
    const { data } = await api.post('/notifications/send', {
      type: payload.type,
      event_type: payload.eventType,
      event_id: payload.eventId,
      recipient: payload.recipient,
      subject: payload.subject,
      body: payload.body,
    });
    return notificationService.getById(data?.id ?? '');
  },
  markAsRead: async (id: string): Promise<void> => {
    await api.put(`/notifications/${id}/read`);
  },
  markAllAsRead: async (): Promise<void> => {
    await api.put(`/notifications/mark-all-read`);
  },
  archive: async (id: string): Promise<void> => {
    await api.put(`/notifications/${id}/archive`);
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/notifications/${id}`);
  },
};

// Audit Logs Services
export const auditLogsService = {
  list: async (filters?: {
    page?: number;
    limit?: number;
    table_name?: string;
    action?: string;
    start_date?: string;
    end_date?: string;
    user_id?: string;
    record_id?: string;
    include_deleted?: boolean;
  }): Promise<AuditLogListResult> => {
    const limit = filters?.limit && filters.limit > 0 ? filters.limit : 50;
    const page = filters?.page && filters.page > 0 ? filters.page : 1;
    const offset = (page - 1) * limit;

    const query: Record<string, string | number | boolean> = {
      limit,
      offset,
    };

    if (filters?.table_name) query.table = filters.table_name;
    if (filters?.action) query.action = filters.action;
    if (filters?.start_date) query.start_date = filters.start_date;
    if (filters?.end_date) query.end_date = filters.end_date;
    if (filters?.user_id) query.user_id = filters.user_id;
    if (filters?.record_id) query.record_id = filters.record_id;
    if (filters?.include_deleted) query.include_deleted = 'true';

    const { data } = await api.get('/audit-logs', { params: query });

    const rawItems: RawAuditLogEntry[] = Array.isArray((data as { data?: unknown }).data)
      ? ((data as { data: RawAuditLogEntry[] }).data ?? [])
      : [];

    const items: AuditLogEntry[] = rawItems.map((entry) => ({
      id: entry.id,
      table_name: entry.table_name,
      record_id: entry.record_id ?? null,
      action: entry.action,
      new_values: entry.new_values ?? null,
      old_values: entry.old_values ?? null,
      changed_by: entry.changed_by ?? null,
      created_at: entry.created_at,
    }));

    const responseLimit = data?.limit ?? limit;
    const responseOffset = data?.offset ?? offset;
    const total = data?.total ?? items.length;
    const pageNumber = responseLimit > 0 ? Math.floor(responseOffset / responseLimit) + 1 : 1;

    return {
      items,
      total,
      limit: responseLimit,
      offset: responseOffset,
      page: pageNumber,
    };
  },
  getById: async (id: string): Promise<AuditLogEntry> => {
    const { data } = await api.get(`/audit-logs/${id}`);
    const entry = data as RawAuditLogEntry;
    return {
      id: entry.id,
      table_name: entry.table_name,
      record_id: entry.record_id ?? null,
      action: entry.action,
      new_values: entry.new_values ?? null,
      old_values: entry.old_values ?? null,
      changed_by: entry.changed_by ?? null,
      created_at: entry.created_at,
    };
  },
  getEntityHistory: async (
    table: string,
    recordId: string,
    params?: { limit?: number; offset?: number }
  ): Promise<AuditLogListResult> => {
    const { data } = await api.get(`/audit-logs/entity/${table}/${recordId}`, { params });
    const rawItems: RawAuditLogEntry[] = Array.isArray((data as { data?: unknown }).data)
      ? ((data as { data: RawAuditLogEntry[] }).data ?? [])
      : [];

    const items: AuditLogEntry[] = rawItems.map((entry) => ({
      id: entry.id,
      table_name: entry.table_name,
      record_id: entry.record_id ?? null,
      action: entry.action,
      new_values: entry.new_values ?? null,
      old_values: entry.old_values ?? null,
      changed_by: entry.changed_by ?? null,
      created_at: entry.created_at,
    }));

    const limit = typeof data?.limit === 'number' ? data.limit : params?.limit ?? items.length;
    const offset = typeof data?.offset === 'number' ? data.offset : params?.offset ?? 0;
    const total = typeof data?.total === 'number' ? data.total : items.length;

    return {
      items,
      total,
      limit,
      offset,
      page: limit > 0 ? Math.floor(offset / limit) + 1 : 1,
    };
  },
  getUserActivity: async (
    userId: string,
    params?: {
      start_date?: string;
      end_date?: string;
      limit?: number;
      offset?: number;
    }
  ): Promise<AuditLogListResult> => {
    const { data } = await api.get(`/audit-logs/user/${userId}`, { params });
    const rawItems: RawAuditLogEntry[] = Array.isArray((data as { data?: unknown }).data)
      ? ((data as { data: RawAuditLogEntry[] }).data ?? [])
      : [];

    const items: AuditLogEntry[] = rawItems.map((entry) => ({
      id: entry.id,
      table_name: entry.table_name,
      record_id: entry.record_id ?? null,
      action: entry.action,
      new_values: entry.new_values ?? null,
      old_values: entry.old_values ?? null,
      changed_by: entry.changed_by ?? null,
      created_at: entry.created_at,
    }));

    const limit = typeof data?.limit === 'number' ? data.limit : params?.limit ?? items.length;
    const offset = typeof data?.offset === 'number' ? data.offset : params?.offset ?? 0;
    const total = typeof data?.total === 'number' ? data.total : items.length;

    return {
      items,
      total,
      limit,
      offset,
      page: limit > 0 ? Math.floor(offset / limit) + 1 : 1,
    };
  },
  getSummary: async (params?: { start_date?: string; end_date?: string }): Promise<AuditSummary> => {
    const { data } = await api.get('/audit-logs/summary', { params });
    return (data ?? {}) as AuditSummary;
  },
  getTableNames: async (): Promise<string[]> => {
    const { data } = await api.get('/audit-logs/tables');
    return Array.isArray(data?.table_names) ? data.table_names : [];
  },
  getActions: async (): Promise<string[]> => {
    const { data } = await api.get('/audit-logs/actions');
    return Array.isArray(data?.actions) ? data.actions : [];
  },
};

// Tally Services
export const tallyService = {
  exportData: (data: {
    data_type: 'invoices' | 'orders' | 'products';
    start_date?: string;
    end_date?: string
  }) => api.post('/api/tally/export', data),
  importData: (data: {
    data_type: string;
    file_path?: string
  }) => api.post('/api/tally/import', data),
};

// User Services
export const userService = {
  list: () => api.get('/users'),
  getById: (id: string) => api.get(`/users/${id}`),
  create: (data: Record<string, unknown>) => api.post('/users', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/users/${id}`, data),
  delete: (id: string) => api.delete(`/users/${id}`),
  me: () => api.get('/me'),
};

// Tenant Services
export const tenantService = {
  list: () => api.get('/tenants'),
  create: (data: { name: string; subdomain: string; license: string }) => api.post('/tenants', data),
  getById: (id: string) => api.get(`/tenants/${id}`),
  update: (id: string, data: Record<string, unknown>) => api.put(`/tenants/${id}`, data),
  delete: (id: string) => api.delete(`/tenants/${id}`),
};

// Product Services
export const productService = {
  list: (params?: { page?: number; limit?: number; category_id?: string }) =>
    api.get('/products', { params }),
  getById: (id: string) => api.get(`/products/${id}`),
  create: (data: Record<string, unknown>) => api.post('/products', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/products/${id}`, data),
  delete: (id: string) => api.delete(`/products/${id}`),
  search: (query: string) => api.get('/products/search', { params: { q: query } }),
  bulkCreate: (data: Record<string, unknown>[]) => api.post('/products/bulk/create', { products: data }),
  bulkUpdate: (data: Record<string, unknown>[]) => api.post('/products/bulk/update', { products: data }),
  bulkPriceUpdate: (data: {
    product_ids: string[];
    adjustment: {
      type: 'percentage' | 'fixed';
      value: number;
    };
  }) => api.post('/products/bulk-price-update', data),
  uploadImage: (id: string, file: File) => {
    const formData = new FormData();
    formData.append('image', file);
    return api.post(`/products/${id}/images`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
  },
  getImages: (id: string) => api.get(`/products/${id}/images`),
  getImageURL: (productId: string, imageId: string) =>
    api.get(`/products/${productId}/images/${imageId}/url`),
  deleteImage: (productId: string, imageId: string) =>
    api.delete(`/products/${productId}/images/${imageId}`),
};

// Category Services
export const categoryService = {
  list: () => api.get('/categories'),
  getById: (id: string) => api.get(`/categories/${id}`),
  create: (data: Record<string, unknown>) => api.post('/categories', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/categories/${id}`, data),
  delete: (id: string) => api.delete(`/categories/${id}`),
};

// Warehouse Services
export const warehouseService = {
  list: () => api.get('/warehouses'),
  getById: (id: string) => api.get(`/warehouses/${id}`),
  create: (data: Record<string, unknown>) => api.post('/warehouses', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/warehouses/${id}`, data),
  delete: (id: string) => api.delete(`/warehouses/${id}`),
};

// Supplier Services
export const supplierService = {
  list: () => api.get('/suppliers'),
  getById: (id: string) => api.get(`/suppliers/${id}`),
  create: (data: Record<string, unknown>) => api.post('/suppliers', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/suppliers/${id}`, data),
  delete: (id: string) => api.delete(`/suppliers/${id}`),
};

// Distributor Services
export const distributorService = {
  list: () => api.get('/distributors'),
  getById: (id: string) => api.get(`/distributors/${id}`),
  create: (data: Record<string, unknown>) => api.post('/distributors', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/distributors/${id}`, data),
  delete: (id: string) => api.delete(`/distributors/${id}`),
};

// Inventory Services
export const inventoryService = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get('/inventory', { params }),
  getById: (id: string) => api.get(`/inventory/${id}`),
  create: (data: Record<string, unknown>) => api.post('/inventory', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/inventory/${id}`, data),
  delete: (id: string) => api.delete(`/inventory/${id}`),
  search: (params: { product_id?: string; warehouse_id?: string }) =>
    api.get('/inventory/search', { params }),
  adjustStock: (data: {
    warehouse_id: string;
    product_id: string;
    quantity_change: number;
    reason?: string;
  }) => api.post('/inventory/adjust', data),
  getHistory: (id: string, params?: { limit?: number; offset?: number }) =>
    api.get(`/inventory/${id}/history`, { params }),
};

// Order Services
export const orderService = {
  list: (params?: { status?: string; page?: number; limit?: number }) =>
    api.get('/orders', { params }),
  getById: (id: string) => api.get(`/orders/${id}`),
  create: (data: Record<string, unknown>) => api.post('/orders', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/orders/${id}`, data),
  delete: (id: string) => api.delete(`/orders/${id}`),
  // Status transition methods
  approve: (id: string) => api.post(`/orders/${id}/approve`),
  process: (id: string) => api.post(`/orders/${id}/process`),
  receive: (id: string) => api.post(`/orders/${id}/receive`),
  ship: (id: string, expectedDelivery?: string) =>
    api.post(`/orders/${id}/ship`, expectedDelivery ? { expected_delivery: expectedDelivery } : {}),
  deliver: (id: string) => api.post(`/orders/${id}/deliver`),
  cancel: (id: string, notes?: string) =>
    api.post(`/orders/${id}/cancel`, notes ? { notes } : {}),
  // Advanced features
  search: (params: {
    query?: string;
    status?: string;
    order_type?: string;
    supplier_id?: string;
    distributor_id?: string;
    product_id?: string;
    warehouse_id?: string;
    min_quantity?: number;
    max_quantity?: number;
    order_date_from?: string;
    order_date_to?: string;
    limit?: number;
    offset?: number;
  }) => api.get('/orders/search', { params }),
  getHistory: (id: string, params?: { limit?: number; offset?: number }) =>
    api.get(`/orders/${id}/history`, { params }),
  getAnalytics: (params?: { start_date?: string; end_date?: string }) =>
    api.get('/orders/analytics', { params }),
};

// Invoice Services
export const invoiceService = {
  list: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get('/invoices', { params }),
  getById: (id: string) => api.get(`/invoices/${id}`),
  create: (data: Record<string, unknown>) => api.post('/invoices', data),
  bulkCreate: (data: Record<string, unknown>[]) =>
    api.post('/invoices/bulk-create', { invoices: data }),
  update: (id: string, data: Record<string, unknown>) => api.put(`/invoices/${id}`, data),
  updateStatus: (id: string, status: string) =>
    api.put(`/invoices/${id}/status`, { status }),
  getUnpaid: () => api.get('/invoices/unpaid'),
  generatePDF: (id: string) => api.post(`/invoices/${id}/generate-pdf`),
  delete: (id: string) => api.delete(`/invoices/${id}`),
};

// Webhook Services
export const webhookService = {
  list: () => api.get('/webhooks/subscriptions'),
  create: (data: Record<string, unknown>) => api.post('/webhooks/subscriptions', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/webhooks/subscriptions/${id}`, data),
  delete: (id: string) => api.delete(`/webhooks/subscriptions/${id}`),
};
