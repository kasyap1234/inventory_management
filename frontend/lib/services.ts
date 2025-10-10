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
      tenantId: typeof data?.tenant_id === 'string' ? data.tenant_id : data?.tenantId ?? null,
      totalSales: typeof data?.total_sales === 'number' ? data.total_sales : data?.totalSales ?? 0,
      totalStockValue:
        typeof data?.total_stock_value === 'number' ? data.total_stock_value : data?.totalStockValue ?? 0,
      gstCollected: typeof data?.gst_collected === 'number' ? data.gst_collected : data?.gstCollected ?? 0,
      orderCount: typeof data?.order_count === 'number' ? data.order_count : data?.orderCount ?? 0,
      lowStockItems:
        typeof data?.low_stock_items === 'number'
          ? data.low_stock_items
          : data?.low_stock_items_count ?? 0,
      lastUpdated: typeof data?.last_updated === 'string' ? data.last_updated : data?.lastUpdated ?? null,
    };
  },
  getSalesTrends: async (params?: { start_date?: string; end_date?: string }): Promise<AnalyticsSalesTrends> => {
    const { data } = await api.get('/analytics/sales-trends', { params });

    const trendsSource: RawTrend[] = Array.isArray((data as { trends?: unknown }).trends)
      ? ((data as { trends: RawTrend[] }).trends ?? [])
      : Array.isArray((data as { Trends?: unknown }).Trends)
        ? ((data as { Trends: RawTrend[] }).Trends ?? [])
        : [];

    const trends: AnalyticsTrendPoint[] = trendsSource
      .map((trend) => {
        const date = trend.date ?? trend.Date ?? null;
        const salesAmount =
          typeof trend.sales_amount === 'number'
            ? trend.sales_amount
            : typeof trend.SalesAmount === 'number'
              ? trend.SalesAmount
              : 0;
        const orderCount =
          typeof trend.order_count === 'number'
            ? trend.order_count
            : typeof trend.OrderCount === 'number'
              ? trend.OrderCount
              : 0;
        return { date, salesAmount, orderCount };
      })
      .sort((a, b) => {
        const aTime = a.date ? new Date(a.date).getTime() : 0;
        const bTime = b.date ? new Date(b.date).getTime() : 0;
        return aTime - bTime;
      });

    return {
      startDate: typeof data?.start_date === 'string' ? data.start_date : data?.startDate ?? null,
      endDate: typeof data?.end_date === 'string' ? data.end_date : data?.endDate ?? null,
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
      cgst: typeof data?.cgst === 'number' ? data.cgst : data?.CGST ?? 0,
      sgst: typeof data?.sgst === 'number' ? data.sgst : data?.SGST ?? 0,
      igst: typeof data?.igst === 'number' ? data.igst : data?.IGST ?? 0,
      total: typeof data?.total === 'number' ? data.total : data?.Total ?? 0,
    };
  },
  getTopProducts: async (params?: {
    limit?: number;
    start_date?: string;
    end_date?: string;
  }): Promise<ProductSales[]> => {
    const { data } = await api.get('/analytics/top-products', { params });

    const productsSource: RawProductSales[] = Array.isArray((data as { products?: unknown }).products)
      ? ((data as { products: RawProductSales[] }).products ?? [])
      : Array.isArray((data as { Products?: unknown }).Products)
        ? ((data as { Products: RawProductSales[] }).Products ?? [])
        : [];

    return productsSource.map((product) => ({
      productId: product.product_id ?? product.ProductID ?? '',
      productName: product.product_name ?? product.ProductName ?? '',
      totalSales:
        typeof product.total_sales === 'number'
          ? product.total_sales
          : typeof product.TotalSales === 'number'
            ? product.TotalSales
            : 0,
      unitsSold:
        typeof product.units_sold === 'number'
          ? product.units_sold
          : typeof product.UnitsSold === 'number'
            ? product.UnitsSold
            : 0,
      orderCount:
        typeof product.order_count === 'number'
          ? product.order_count
          : typeof product.OrderCount === 'number'
            ? product.OrderCount
            : 0,
    }));
  },
  getLowStockReport: async (params?: { threshold?: number }): Promise<LowStockItem[]> => {
    const { data } = await api.get('/analytics/low-stock', { params });

    const itemsSource: RawLowStockItem[] = Array.isArray((data as { low_stock_items?: unknown }).low_stock_items)
      ? ((data as { low_stock_items: RawLowStockItem[] }).low_stock_items ?? [])
      : Array.isArray(data)
        ? (data as RawLowStockItem[])
        : [];

    return itemsSource.map((item) => ({
      productId: item.product_id ?? item.ProductID ?? '',
      productName: item.product_name ?? item.ProductName ?? '',
      warehouseId: item.warehouse_id ?? item.WarehouseID ?? '',
      currentStock:
        typeof item.current_stock === 'number'
          ? item.current_stock
          : typeof item.CurrentStock === 'number'
            ? item.CurrentStock
            : 0,
      threshold:
        typeof item.threshold === 'number'
          ? item.threshold
          : typeof item.Threshold === 'number'
            ? item.Threshold
            : params?.threshold ?? 0,
      unitPrice:
        typeof item.unit_price === 'number'
          ? item.unit_price
          : typeof item.UnitPrice === 'number'
            ? item.UnitPrice
            : 0,
      stockValue:
        typeof item.stock_value === 'number'
          ? item.stock_value
          : typeof item.StockValue === 'number'
            ? item.StockValue
            : 0,
    }));
  },
  getInventoryValuation: async (): Promise<InventoryValuation> => {
    const { data } = await api.get('/analytics/inventory-valuation');

    const payload = data as RawInventoryValuation;

    return {
      tenantId: typeof payload.tenant_id === 'string' ? payload.tenant_id : payload.tenantId ?? null,
      totalValue:
        typeof payload.total_value === 'number'
          ? payload.total_value
          : payload.totalValue ?? 0,
      totalItems:
        typeof payload.total_items === 'number'
          ? payload.total_items
          : payload.totalItems ?? 0,
      totalQuantity:
        typeof payload.total_quantity === 'number'
          ? payload.total_quantity
          : payload.totalQuantity ?? 0,
      byWarehouse: payload.by_warehouse ?? payload.byWarehouse ?? {},
      byCategory: payload.by_category ?? payload.byCategory ?? {},
      lastCalculated:
        typeof payload.last_calculated === 'string'
          ? payload.last_calculated
          : payload.lastCalculated ?? null,
    };
  },
  getRevenueByCategory: async (params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<RevenueByCategory[]> => {
    const { data } = await api.get('/analytics/revenue-by-category', { params });

    const revenueMap =
      (data as { revenue_by_category?: Record<string, unknown>; revenueByCategory?: Record<string, unknown> })
        .revenue_by_category ??
      (data as { revenueByCategory?: Record<string, unknown> }).revenueByCategory ??
      (typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {});

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

    const statusMap = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
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
};

// Subscription Services
export const subscriptionService = {
  list: async (params?: { limit?: number; offset?: number }): Promise<SubscriptionListResult> => {
    const { data } = await api.get('/subscriptions', { params });
    const rawItems: RawSubscription[] = Array.isArray((data as { subscriptions?: unknown }).subscriptions)
      ? ((data as { subscriptions: RawSubscription[] }).subscriptions ?? [])
      : [];

    const items: SubscriptionDto[] = rawItems.map((subscription) => ({
      id: subscription.id,
      plan_name: subscription.plan_name,
      amount: typeof subscription.amount === 'number' ? subscription.amount : Number(subscription.amount) || 0,
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
      amount: typeof payload.amount === 'number' ? payload.amount : Number(payload.amount) || 0,
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
    const plansSource = data?.plans ?? [];

    if (Array.isArray(plansSource)) {
      return plansSource.map((plan: SubscriptionPlanConfig) => ({
        name: plan.name,
        amount: plan.amount,
        currency: plan.currency,
        interval: plan.interval,
        description: plan.description,
        features: plan.features ?? [],
      }));
    }

    return Object.values(plansSource as Record<string, SubscriptionPlanConfig>).map((plan) => ({
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
    const rawItems: RawNotification[] = Array.isArray((data as { notifications?: unknown }).notifications)
      ? ((data as { notifications: RawNotification[] }).notifications ?? [])
      : [];

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
      count: typeof data?.count === 'number' ? data.count : items.length,
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

    const responseLimit = typeof data?.limit === 'number' ? data.limit : limit;
    const responseOffset = typeof data?.offset === 'number' ? data.offset : offset;
    const total = typeof data?.total === 'number' ? data.total : items.length;
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
  uploadImage: (id: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
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
};

// Order Services
export const orderService = {
  list: (params?: { status?: string; page?: number; limit?: number }) => 
    api.get('/orders', { params }),
  getById: (id: string) => api.get(`/orders/${id}`),
  create: (data: Record<string, unknown>) => api.post('/orders', data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/orders/${id}`, data),
  delete: (id: string) => api.delete(`/orders/${id}`),
};

// Invoice Services
export const invoiceService = {
  list: (params?: { page?: number; limit?: number; status?: string }) => 
    api.get('/invoices', { params }),
  getById: (id: string) => api.get(`/invoices/${id}`),
  create: (data: Record<string, unknown>) => api.post('/invoices', data),
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
