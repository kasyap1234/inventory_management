'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useDebounce } from '@/hooks/useDebounce';
import { Plus, Search, CheckCircle, Package, Truck, Home, XCircle, MoreHorizontal } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select } from '@/components/ui/select';
import { DropdownMenu, DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu';
import { OrderStatusBadge } from "@/components/orders/OrderStatusBadge";
import { format } from "date-fns";
import api from '@/lib/api';
import { Order, Product, Warehouse, Supplier, Distributor } from '@/types';
import { formatCurrency, formatDate } from '@/lib/utils';

import { useRazorpayCheckout } from '@/hooks/useRazorpayCheckout';

export default function OrdersPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState<'all' | 'purchase' | 'sales'>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'approved' | 'processing' | 'shipped' | 'delivered' | 'cancelled'>('all');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [advancedFilters, setAdvancedFilters] = useState<Record<string, any>>({});
  const [selectedOrders, setSelectedOrders] = useState<string[]>([]);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [isDetailsDialogOpen, setIsDetailsDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  const { createOneTimePayment, isLoading: isPaymentLoading } = useRazorpayCheckout();

  // Debounce search query to reduce API calls
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  const { data: orders, isLoading } = useQuery<{ orders: Order[] }>({
    queryKey: ['orders'],
    queryFn: async () => {
      const response = await api.get('/orders?limit=100');
      return response.data;
    },
  });

  const { data: products } = useQuery<{ products: Product[] }>({
    queryKey: ['products'],
    queryFn: async () => {
      const response = await api.get('/products?limit=100');
      return response.data;
    },
  });

  const { data: warehouses } = useQuery<{ warehouses: Warehouse[] }>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const response = await api.get('/warehouses?limit=100');
      return response.data;
    },
  });

  const { data: suppliers } = useQuery<{ suppliers: Supplier[] }>({
    queryKey: ['suppliers'],
    queryFn: async () => {
      const response = await api.get('/suppliers?limit=100');
      return response.data;
    },
  });

  const { data: distributors } = useQuery<{ distributors: Distributor[] }>({
    queryKey: ['distributors'],
    queryFn: async () => {
      const response = await api.get('/distributors?limit=100');
      return response.data;
    },
  });

  const deleteOrder = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/orders/${id}`);
    },
    onMutate: async (id) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['orders'] });

      // Snapshot previous value
      const previousOrders = queryClient.getQueryData<{ orders: Order[] }>(['orders']);

      // Optimistically remove from list
      queryClient.setQueryData<{ orders: Order[] }>(
        ['orders'],
        (old) => ({
          ...old,
          orders: old?.orders?.filter((o) => o.id !== id) || [],
        })
      );

      return { previousOrders };
    },
    onError: (error, id, context) => {
      // Rollback on error
      if (context?.previousOrders) {
        queryClient.setQueryData(['orders'], context.previousOrders);
      }
      alert('Failed to delete order. Please try again.');
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['orders'], refetchType: 'none' });
      setSelectedOrders([]);
    },
  });

  // Order status action mutations
  const orderAction = useMutation({
    mutationFn: async ({ orderId, action }: { orderId: string; action: string }) => {
      await api.post(`/orders/${orderId}/${action}`);
    },
    onMutate: async ({ orderId, action }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['orders'] });

      // Snapshot previous value
      const previousOrders = queryClient.getQueryData<{ orders: Order[] }>(['orders']);

      // Determine new status based on action
      const statusMap: Record<string, string> = {
        'approve': 'approved',
        'process': 'processing',
        'ship': 'shipped',
        'deliver': 'delivered',
        'cancel': 'cancelled',
      };

      const newStatus = statusMap[action];

      // Optimistically update status
      if (newStatus) {
        queryClient.setQueryData<{ orders: Order[] }>(
          ['orders'],
          (old) => ({
            ...old,
            orders: old?.orders?.map((order) =>
              order.id === orderId ? { ...order, status: newStatus as Order['status'] } : order
            ) || [],
          })
        );
      }

      return { previousOrders };
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousOrders) {
        queryClient.setQueryData(['orders'], context.previousOrders);
      }
      const err = error as { response?: { data?: { error?: { message?: string } } } };
      alert(err.response?.data?.error?.message || `Failed to ${variables.action} order`);
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['orders'], refetchType: 'none' });
      setSelectedOrders([]);
    },
  });

  const bulkOrderAction = useMutation({
    mutationFn: async ({ action, eligibleOrderIds }: { action: string; eligibleOrderIds: string[] }) => {
      for (const id of eligibleOrderIds) {
        await api.post(`/orders/${id}/${action}`);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      setSelectedOrders([]);
    },
    onError: (error: any) => {
      alert(error.response?.data?.error?.message || 'Bulk action failed');
    },
  });

  const bulkDeleteOrders = useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map((id) => api.delete(`/orders/${id}`)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      setSelectedOrders([]);
    },
    onError: () => {
      alert('Failed to delete selected orders');
    },
  });

  const handleCollectPayment = (order: Order) => {
    const amount = Number(order.quantity || 0) * Number(order.unit_price || 0);
    if (!amount || amount <= 0) {
      alert('Order amount is invalid for payment');
      return;
    }

    createOneTimePayment({
      amount,
      currency: 'INR',
      receipt: `order-${order.id}`,
      orderId: order.id,
      notes: {
        order_type: order.order_type || 'sales',
      },
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['orders'] });
        alert('Payment captured successfully');
      },
      onError: (err) => {
        console.error(err);
        alert((err as any)?.message || 'Payment failed');
      },
    });
  };

  const filteredOrders = orders?.orders?.filter(order => {
    const product = products?.products?.find(p => p.id === order.product_id);
    const matchesSearch = product?.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = typeFilter === 'all' || order.order_type === typeFilter;
    const matchesStatus = statusFilter === 'all' || order.status === statusFilter;

    // Apply advanced filters
    if (advancedFilters.statuses && advancedFilters.statuses.length > 0) {
      if (!advancedFilters.statuses.includes(order.status)) return false;
    }
    if (advancedFilters.start_date && order.order_date < advancedFilters.start_date) return false;
    if (advancedFilters.end_date && order.order_date > advancedFilters.end_date) return false;
    if (advancedFilters.min_price && (order.quantity * order.unit_price) < parseFloat(advancedFilters.min_price)) return false;
    if (advancedFilters.max_price && (order.quantity * order.unit_price) > parseFloat(advancedFilters.max_price)) return false;
    if (advancedFilters.min_quantity && order.quantity < parseInt(advancedFilters.min_quantity)) return false;
    if (advancedFilters.max_quantity && order.quantity > parseInt(advancedFilters.max_quantity)) return false;

    return matchesSearch && matchesType && matchesStatus;
  }) || [];

  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'delivered': return 'success';
      case 'pending': return 'warning';
      case 'cancelled': return 'destructive';
      default: return 'secondary';
    }
  };

  const selectedOrderDetails = filteredOrders.filter((order) => selectedOrders.includes(order.id));

  const getEligibleOrderIds = (action: string) => {
    return selectedOrderDetails
      .filter((order) => {
        switch (action) {
          case 'approve':
            return order.status === 'pending';
          case 'process':
            return order.status === 'approved';
          case 'receive':
            return order.status === 'approved' && order.order_type === 'purchase';
          case 'ship':
            return order.status === 'received' && order.order_type === 'sales';
          case 'deliver':
            return order.status === 'shipped';
          case 'cancel':
            return !['delivered', 'cancelled'].includes(order.status);
          default:
            return false;
        }
      })
      .map((order) => order.id);
  };

  const handleBulkAction = (action: string) => {
    const eligibleOrderIds = getEligibleOrderIds(action);
    if (eligibleOrderIds.length === 0) {
      alert('No selected orders are eligible for this action');
      return;
    }
    if (confirm(`Apply ${action} to ${eligibleOrderIds.length} order(s)?`)) {
      bulkOrderAction.mutate({ action, eligibleOrderIds });
    }
  };

  const handleBulkDelete = () => {
    const deletableIds = selectedOrderDetails
      .filter((order) => ['delivered', 'cancelled'].includes(order.status))
      .map((order) => order.id);

    if (deletableIds.length === 0) {
      alert('Only delivered or cancelled orders can be deleted');
      return;
    }

    if (confirm(`Delete ${deletableIds.length} order(s)?`)) {
      bulkDeleteOrders.mutate(deletableIds);
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedOrders(filteredOrders.map((order) => order.id));
    } else {
      setSelectedOrders([]);
    }
  };

  const handleSelectOrder = (orderId: string, checked: boolean) => {
    if (checked) {
      setSelectedOrders((prev) => [...prev, orderId]);
    } else {
      setSelectedOrders((prev) => prev.filter((id) => id !== orderId));
    }
  };

  return (
    <div className="space-y-8 p-6">
      <div className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tighter text-foreground uppercase">Orders</h1>
          <p className="text-xs font-mono text-muted-foreground mt-1 uppercase tracking-widest">LOGISTICS CONTROL</p>
          {selectedOrders.length > 0 && (
            <div className="inline-flex items-center gap-2 mt-2 px-3 py-1 bg-primary/10 text-primary text-xs font-mono uppercase tracking-wider border border-primary/20">
              <CheckCircle className="w-3 h-3" />
              {selectedOrders.length} SELECTED
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          {selectedOrders.length > 0 && (
            <div className="flex flex-wrap gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkAction('approve')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Approve
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkAction('process')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Process
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkAction('receive')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Receive
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkAction('ship')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Ship
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleBulkAction('deliver')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Deliver
              </Button>
              <Button
                size="sm"
                variant="destructive"
                onClick={() => handleBulkAction('cancel')}
                disabled={bulkOrderAction.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Cancel
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={handleBulkDelete}
                disabled={bulkDeleteOrders.isPending}
                className="rounded-none font-mono uppercase text-xs"
              >
                Delete
              </Button>
            </div>
          )}
          <Button onClick={() => setIsAddDialogOpen(true)} className="rounded-none font-mono uppercase tracking-wider">
            <Plus className="h-4 w-4 mr-2" />
            New Order
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card className="rounded-none border-l-4 border-l-primary">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Total Orders</CardTitle>
            <Package className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">{orders?.orders?.length || 0}</div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              ALL TIME
            </p>
          </CardContent>
        </Card>

        <Card className="rounded-none border-l-4 border-l-purple-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Purchase Orders</CardTitle>
            <Home className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">
              {orders?.orders?.filter(o => o.order_type === 'purchase').length || 0}
            </div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              INCOMING
            </p>
          </CardContent>
        </Card>

        <Card className="rounded-none border-l-4 border-l-emerald-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Sales Orders</CardTitle>
            <Truck className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">
              {orders?.orders?.filter(o => o.order_type === 'sales').length || 0}
            </div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              OUTGOING
            </p>
          </CardContent>
        </Card>

        <Card className="rounded-none border-l-4 border-l-amber-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Pending</CardTitle>
            <CheckCircle className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">
              {orders?.orders?.filter(o => o.status === 'pending').length || 0}
            </div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              ACTION REQUIRED
            </p>
          </CardContent>
        </Card>
      </div>

      <Card className="rounded-none border border-border">
        <CardHeader className="border-b border-border bg-muted/10 pb-4">
          <div className="flex flex-col md:flex-row gap-4 justify-between">
            <div className="flex flex-1 gap-4">
              <div className="relative flex-1 max-w-sm">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="SEARCH ORDERS..."
                  className="pl-9 rounded-none border-border bg-background font-mono text-sm uppercase placeholder:text-muted-foreground/50"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />
              </div>
              <Select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value as 'all' | 'purchase' | 'sales')}
                className="w-[180px] rounded-none border-border font-mono uppercase text-xs"
              >
                <option value="all">All Types</option>
                <option value="purchase">Purchase</option>
                <option value="sales">Sales</option>
              </Select>
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as 'all' | 'pending' | 'approved' | 'shipped' | 'delivered' | 'cancelled')}
                className="w-[180px] rounded-none border-border font-mono uppercase text-xs"
              >
                <option value="all">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="approved">Approved</option>
                <option value="shipped">Shipped</option>
                <option value="delivered">Delivered</option>
                <option value="cancelled">Cancelled</option>
              </Select>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow className="border-border hover:bg-transparent">
                <TableHead className="w-[50px] border-r border-border">
                  <input
                    type="checkbox"
                    checked={
                      filteredOrders?.length > 0 &&
                      selectedOrders.length === filteredOrders?.length
                    }
                    onChange={(e) => handleSelectAll(e.target.checked)}
                    aria-label="Select all"
                    className="rounded-none border-muted-foreground accent-primary h-4 w-4"
                  />
                </TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Order ID</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Type</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Distributor/Supplier</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Date</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Total</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Status</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 border-r border-border text-muted-foreground">Timeline</TableHead>
                <TableHead className="font-mono text-xs uppercase tracking-wider h-12 text-right text-muted-foreground">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-24 text-center font-mono text-muted-foreground animate-pulse">
                    LOADING DATA STREAM...
                  </TableCell>
                </TableRow>
              ) : filteredOrders?.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-24 text-center font-mono text-muted-foreground">
                    NO ORDERS FOUND
                  </TableCell>
                </TableRow>
              ) : (
                filteredOrders?.map((order) => (
                  <TableRow key={order.id} className="border-border hover:bg-muted/20 transition-colors group">
                    <TableCell className="border-r border-border">
                      <input
                        type="checkbox"
                        checked={selectedOrders.includes(order.id)}
                        onChange={(e) =>
                          handleSelectOrder(order.id, e.target.checked)
                        }
                        aria-label={`Select order ${order.id}`}
                        className="rounded-none border-muted-foreground accent-primary h-4 w-4"
                      />
                    </TableCell>
                    <TableCell className="font-mono text-sm border-r border-border font-medium text-foreground">
                      #{order.id.substring(0, 8)}
                    </TableCell>
                    <TableCell className="font-mono text-sm border-r border-border">
                      <span className={`px-2 py-0.5 text-[10px] uppercase tracking-wider border ${order.order_type === 'sales'
                          ? 'border-emerald-500/50 text-emerald-500 bg-emerald-500/10'
                          : 'border-purple-500/50 text-purple-500 bg-purple-500/10'
                        }`}>
                        {order.order_type}
                      </span>
                    </TableCell>
                    <TableCell className="font-mono text-sm border-r border-border text-muted-foreground">
                      {order.distributor_id || order.supplier_id || '-'}
                    </TableCell>
                    <TableCell className="font-mono text-sm border-r border-border text-muted-foreground">
                      {format(new Date(order.created_at), 'yyyy-MM-dd')}
                    </TableCell>
                    <TableCell className="font-mono text-sm border-r border-border font-bold text-foreground">
                      ${(Number(order.unit_price) * Number(order.quantity)).toFixed(2)}
                    </TableCell>
                    <TableCell className="border-r border-border">
                      <OrderStatusBadge status={order.status} />
                    </TableCell>
                    <TableCell className="border-r border-border">
                      <div className="flex items-center space-x-1">
                        {['pending', 'approved', 'shipped', 'delivered'].map((step, idx, arr) => {
                          const currentStatusIdx = arr.indexOf(order.status.toLowerCase());
                          const isCompleted = idx <= currentStatusIdx;
                          const isCurrent = idx === currentStatusIdx;

                          return (
                            <div key={step} className="flex items-center">
                              <div className={`h-1.5 w-1.5 rounded-full transition-all duration-500 ${isCurrent
                                  ? 'bg-primary shadow-[0_0_8px_rgba(0,255,136,0.8)] scale-125'
                                  : isCompleted
                                    ? 'bg-primary/70'
                                    : 'bg-muted'
                                }`} title={step} />
                              {idx < arr.length - 1 && (
                                <div className={`h-[1px] w-3 transition-colors duration-500 ${idx < currentStatusIdx ? 'bg-primary/50' : 'bg-muted/30'
                                  }`} />
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu trigger={
                        <Button variant="ghost" className="h-8 w-8 p-0 rounded-none hover:bg-primary/10 hover:text-primary">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      }>
                        <div className="font-mono uppercase text-xs text-muted-foreground px-2 py-1.5">Actions</div>
                        <DropdownMenuItem
                          onSelect={() => {
                            setSelectedOrder(order);
                            setIsDetailsDialogOpen(true);
                          }}
                          className="font-mono text-xs uppercase cursor-pointer focus:bg-primary/10 focus:text-primary"
                        >
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onSelect={() => handleCollectPayment(order)}
                          className="font-mono text-xs uppercase cursor-pointer focus:bg-primary/10 focus:text-primary"
                          disabled={isPaymentLoading}
                        >
                          Collect Payment
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        {order.status === 'pending' && (
                          <DropdownMenuItem
                            onSelect={() => orderAction.mutate({ orderId: order.id, action: 'approve' })}
                            className="font-mono text-xs uppercase cursor-pointer focus:bg-primary/10 focus:text-primary"
                          >
                            Approve Order
                          </DropdownMenuItem>
                        )}
                        {order.status === 'approved' && (
                          <DropdownMenuItem
                            onSelect={() => orderAction.mutate({ orderId: order.id, action: 'ship' })}
                            className="font-mono text-xs uppercase cursor-pointer focus:bg-primary/10 focus:text-primary"
                          >
                            Ship Order
                          </DropdownMenuItem>
                        )}
                        {order.status === 'shipped' && (
                          <DropdownMenuItem
                            onSelect={() => orderAction.mutate({ orderId: order.id, action: 'deliver' })}
                            className="font-mono text-xs uppercase cursor-pointer focus:bg-primary/10 focus:text-primary"
                          >
                            Mark Delivered
                          </DropdownMenuItem>
                        )}
                        {['pending', 'approved'].includes(order.status) && (
                          <DropdownMenuItem
                            onSelect={() => orderAction.mutate({ orderId: order.id, action: 'cancel' })}
                            className="font-mono text-xs uppercase cursor-pointer text-destructive focus:bg-destructive/10 focus:text-destructive"
                          >
                            Cancel Order
                          </DropdownMenuItem>
                        )}
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <OrderFormDialog
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        products={products?.products || []}
        warehouses={warehouses?.warehouses || []}
        suppliers={suppliers?.suppliers || []}
        distributors={distributors?.distributors || []}
      />
      <OrderDetailsDialog
        open={isDetailsDialogOpen}
        onOpenChange={setIsDetailsDialogOpen}
        order={selectedOrder}
        products={products?.products || []}
        warehouses={warehouses?.warehouses || []}
        suppliers={suppliers?.suppliers || []}
        distributors={distributors?.distributors || []}
      />
    </div>
  );
}

function OrderFormDialog({
  open,
  onOpenChange,
  products,
  warehouses,
  suppliers,
  distributors,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  products: Product[];
  warehouses: Warehouse[];
  suppliers: Supplier[];
  distributors: Distributor[];
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    order_type: 'purchase' as 'purchase' | 'sales',
    product_id: '',
    warehouse_id: '',
    supplier_id: '',
    distributor_id: '',
    quantity: 0,
    unit_price: 0,
    order_date: new Date().toISOString().split('T')[0],
    expected_delivery: '',
    notes: '',
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      await api.post('/orders', {
        ...data,
        supplier_id: data.order_type === 'purchase' ? data.supplier_id : undefined,
        distributor_id: data.order_type === 'sales' ? data.distributor_id : undefined,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      onOpenChange(false);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="rounded-none border-border bg-background sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-lg">Create Order</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-6 mt-4">
          <div className="space-y-2">
            <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Order Type *</label>
            <Select
              value={formData.order_type}
              onChange={(e) => setFormData({ ...formData, order_type: e.target.value as 'purchase' | 'sales' })}
              className="w-full rounded-none border-border font-mono text-sm uppercase"
            >
              <option value="purchase">Purchase Order</option>
              <option value="sales">Sales Order</option>
            </Select>
          </div>

          {formData.order_type === 'purchase' && (
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Supplier *</label>
              <Select
                value={formData.supplier_id}
                onChange={(e) => setFormData({ ...formData, supplier_id: e.target.value })}
                className="w-full rounded-none border-border font-mono text-sm"
              >
                <option value="">SELECT SUPPLIER</option>
                {suppliers.map((supplier) => (
                  <option key={supplier.id} value={supplier.id}>{supplier.name}</option>
                ))}
              </Select>
            </div>
          )}

          {formData.order_type === 'sales' && (
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Distributor *</label>
              <Select
                value={formData.distributor_id}
                onChange={(e) => setFormData({ ...formData, distributor_id: e.target.value })}
                className="w-full rounded-none border-border font-mono text-sm"
              >
                <option value="">SELECT DISTRIBUTOR</option>
                {distributors.map((distributor) => (
                  <option key={distributor.id} value={distributor.id}>{distributor.name}</option>
                ))}
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Product *</label>
            <Select
              value={formData.product_id}
              onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
              className="w-full rounded-none border-border font-mono text-sm"
            >
              <option value="">SELECT PRODUCT</option>
              {products.map((product) => (
                <option key={product.id} value={product.id}>{product.name}</option>
              ))}
            </Select>
          </div>

          <div className="space-y-2">
            <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Warehouse *</label>
            <Select
              value={formData.warehouse_id}
              onChange={(e) => setFormData({ ...formData, warehouse_id: e.target.value })}
              className="w-full rounded-none border-border font-mono text-sm"
            >
              <option value="">SELECT WAREHOUSE</option>
              {warehouses.map((warehouse) => (
                <option key={warehouse.id} value={warehouse.id}>{warehouse.name}</option>
              ))}
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Quantity *</label>
              <Input
                required
                type="number"
                value={formData.quantity}
                onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) })}
                className="rounded-none border-border font-mono text-sm"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Unit Price *</label>
              <Input
                required
                type="number"
                step="0.01"
                value={formData.unit_price}
                onChange={(e) => setFormData({ ...formData, unit_price: parseFloat(e.target.value) })}
                className="rounded-none border-border font-mono text-sm"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Order Date *</label>
              <Input
                required
                type="date"
                value={formData.order_date}
                onChange={(e) => setFormData({ ...formData, order_date: e.target.value })}
                className="rounded-none border-border font-mono text-sm uppercase"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Expected Delivery</label>
              <Input
                type="date"
                value={formData.expected_delivery}
                onChange={(e) => setFormData({ ...formData, expected_delivery: e.target.value })}
                className="rounded-none border-border font-mono text-sm uppercase"
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground">Notes</label>
            <Input
              value={formData.notes}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              placeholder="OPTIONAL NOTES"
              className="rounded-none border-border font-mono text-sm placeholder:text-muted-foreground/50"
            />
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} className="rounded-none font-mono uppercase">
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending} className="rounded-none font-mono uppercase">
              {saveMutation.isPending ? 'CREATING...' : 'CREATE ORDER'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function OrderDetailsDialog({
  open,
  onOpenChange,
  order,
  products,
  warehouses,
  suppliers,
  distributors,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  order: Order | null;
  products: Product[];
  warehouses: Warehouse[];
  suppliers: Supplier[];
  distributors: Distributor[];
}) {
  if (!order) return null;

  const product = products.find(p => p.id === order.product_id);
  const warehouse = warehouses.find(w => w.id === order.warehouse_id);
  const supplier = suppliers.find(s => s.id === order.supplier_id);
  const distributor = distributors.find(d => d.id === order.distributor_id);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="rounded-none border-border bg-background sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="font-mono uppercase tracking-widest text-lg">Order Details</DialogTitle>
        </DialogHeader>
        <div className="space-y-6 mt-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div>
              <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Order ID</span>
              <div className="font-mono text-sm font-medium">{order.id}</div>
            </div>
            <div>
              <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Order Type</span>
              <div className="flex items-center">
                <span className={`px-2 py-0.5 text-xs font-mono uppercase tracking-wider border ${order.order_type === 'sales'
                    ? 'border-emerald-500/50 text-emerald-500 bg-emerald-500/10'
                    : 'border-purple-500/50 text-purple-500 bg-purple-500/10'
                  }`}>
                  {order.order_type}
                </span>
              </div>
            </div>
            <div>
              <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Status</span>
              <OrderStatusBadge status={order.status} />
            </div>
            <div>
              <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Date</span>
              <div className="font-mono text-sm">{format(new Date(order.created_at), 'PPP')}</div>
            </div>
          </div>

          <div className="border-t border-border pt-4">
            <h4 className="text-xs font-mono uppercase tracking-widest text-muted-foreground mb-4">Product Information</h4>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <div>
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Product</span>
                <div className="font-mono text-sm">{product?.name || order.product_id}</div>
              </div>
              <div>
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Warehouse</span>
                <div className="font-mono text-sm">{warehouse?.name || order.warehouse_id}</div>
              </div>
              <div>
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Quantity</span>
                <div className="font-mono text-sm">{order.quantity}</div>
              </div>
              <div>
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Unit Price</span>
                <div className="font-mono text-sm">${Number(order.unit_price).toFixed(2)}</div>
              </div>
              <div className="col-span-2">
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Total Amount</span>
                <div className="font-mono text-xl font-bold text-primary">${(Number(order.unit_price) * Number(order.quantity)).toFixed(2)}</div>
              </div>
            </div>
          </div>

          <div className="border-t border-border pt-4">
            <h4 className="text-xs font-mono uppercase tracking-widest text-muted-foreground mb-4">Logistics</h4>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
              {order.order_type === 'purchase' ? (
                <div>
                  <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Supplier</span>
                  <div className="font-mono text-sm">{supplier?.name || order.supplier_id || '-'}</div>
                </div>
              ) : (
                <div>
                  <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Distributor</span>
                  <div className="font-mono text-sm">{distributor?.name || order.distributor_id || '-'}</div>
                </div>
              )}
              <div>
                <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Expected Delivery</span>
                <div className="font-mono text-sm">
                  {order.expected_delivery ? format(new Date(order.expected_delivery), 'PPP') : '-'}
                </div>
              </div>
            </div>
          </div>

          {order.notes && (
            <div className="border-t border-border pt-4">
              <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground block mb-1">Notes</span>
              <div className="font-mono text-sm text-muted-foreground bg-muted/10 p-3 border border-border">
                {order.notes}
              </div>
            </div>
          )}

          <div className="flex justify-end pt-4">
            <Button variant="outline" onClick={() => onOpenChange(false)} className="rounded-none font-mono uppercase">
              Close
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
