'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, AlertTriangle, PlusCircle, MinusCircle, Package, Scan } from 'lucide-react';
import { useDebounce } from '@/hooks/useDebounce';
import { BarcodeScanner } from '@/components/inventory/BarcodeScanner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import { BulkAdjustmentDialog } from '@/components/inventory/BulkAdjustmentDialog';
import { toast } from 'react-hot-toast';
import AdvancedFilters, { ActiveFilterBadges } from '@/components/filters/AdvancedFilters';
import api from '@/lib/api';
import { Inventory, Product, Warehouse } from '@/types';
import { formatDateTime } from '@/lib/utils';

// InventoryWithDetails is now just an alias since we added product_name and warehouse_name to Inventory
type InventoryWithDetails = Inventory;

type InventoryHistoryEntry = {
  id: string;
  action: string;
  created_at: string;
  new_values?: Record<string, unknown>;
  old_values?: Record<string, unknown>;
  changed_by?: string | null;
};

export default function InventoryPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isScannerOpen, setIsScannerOpen] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingInventory, setEditingInventory] = useState<Inventory | null>(null);
  const [adjustingInventory, setAdjustingInventory] = useState<Inventory | null>(null);
  const [advancedFilters, setAdvancedFilters] = useState<Record<string, string | number | string[] | undefined>>({});
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());
  const [isBulkAdjustmentOpen, setIsBulkAdjustmentOpen] = useState(false);
  const queryClient = useQueryClient();

  // Debounce search query to reduce API calls
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  const handleSelectAll = (checked: boolean) => {
    if (checked && inventory?.inventories) {
      setSelectedItems(new Set(inventory.inventories.map((item) => item.id)));
    } else {
      setSelectedItems(new Set());
    }
  };

  const handleSelectItem = (id: string, checked: boolean) => {
    const newSelected = new Set(selectedItems);
    if (checked) {
      newSelected.add(id);
    } else {
      newSelected.delete(id);
    }
    setSelectedItems(newSelected);
  };

  const bulkDeleteMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      await api.post('/inventory/bulk-delete', { inventory_ids: ids });
    },
    onSuccess: () => {
      toast.success('Selected items deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      setSelectedItems(new Set());
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete items');
    },
  });

  // Build query params for server-side filtering
  const buildQueryParams = () => {
    const params: Record<string, string | number> = {
      limit: 100,
    };

    if (debouncedSearchQuery) {
      params.query = debouncedSearchQuery;
    }

    if (advancedFilters.warehouse_id) {
      params.warehouse_id = String(advancedFilters.warehouse_id);
    }

    if (advancedFilters.product_id) {
      params.product_id = String(advancedFilters.product_id);
    }

    if (advancedFilters.min_quantity) {
      params.min_quantity = Number(advancedFilters.min_quantity);
    }

    if (advancedFilters.max_quantity) {
      params.max_quantity = Number(advancedFilters.max_quantity);
    }

    return params;
  };

  const { data: inventory, isLoading } = useQuery<{ inventories: InventoryWithDetails[] }>({
    queryKey: ['inventory', debouncedSearchQuery, advancedFilters],
    queryFn: async () => {
      const params = buildQueryParams();
      const queryString = new URLSearchParams(
        Object.entries(params).reduce((acc, [key, value]) => {
          if (value !== undefined && value !== null && value !== '') {
            acc[key] = String(value);
          }
          return acc;
        }, {} as Record<string, string>)
      ).toString();
      const response = await api.get(`/inventory/search?${queryString}`);
      return response.data;
    },
    staleTime: 30 * 1000, // 30 seconds for inventory data
  });

  // Fetch products and warehouses only for dropdowns (keep these minimal)
  const { data: products } = useQuery<{ products: Product[] }>({
    queryKey: ['products-minimal'],
    queryFn: async () => {
      const response = await api.get('/products?limit=1000');
      return response.data;
    },
  });

  const { data: warehouses } = useQuery<{ warehouses: Warehouse[] }>({
    queryKey: ['warehouses-minimal'],
    queryFn: async () => {
      const response = await api.get('/warehouses?limit=1000');
      return response.data;
    },
  });

  // Apply client-side status filtering only (since backend doesn't support it yet)
  const filteredInventory = inventory?.inventories?.filter(item => {
    const statuses = advancedFilters.statuses as string[] | undefined;
    if (statuses && statuses.length > 0) {
      const status = item.quantity === 0
        ? 'out_of_stock'
        : item.quantity < 10
          ? 'low_stock'
          : 'in_stock';
      if (!statuses.includes(status)) {
        return false;
      }
    }
    return true;
  }) || [];

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between border-b border-border pb-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tighter text-foreground uppercase">Inventory</h1>
          <p className="text-xs font-mono text-muted-foreground mt-1 uppercase tracking-widest">STOCK CONTROL CENTER</p>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => setIsScannerOpen(true)} variant="outline">
            <Scan className="mr-2 h-4 w-4" />
            Scan Barcode
          </Button>
          {selectedItems.size > 0 && (
            <DropdownMenu trigger={
              <Button variant="outline">
                Bulk Actions ({selectedItems.size})
              </Button>
            }>
              <DropdownMenuItem onSelect={() => setIsBulkAdjustmentOpen(true)}>
                Adjust Stock
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-red-600"
                onSelect={() => {
                  if (confirm('Are you sure you want to delete selected items?')) {
                    bulkDeleteMutation.mutate(Array.from(selectedItems));
                  }
                }}
              >
                Delete Selected
              </DropdownMenuItem>
            </DropdownMenu>
          )}
          <Button onClick={() => setIsAddDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Stock
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="rounded-none border-l-4 border-l-primary">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Total Items</CardTitle>
            <Package className="h-4 w-4 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">{inventory?.inventories?.length || 0}</div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              UNIQUE SKUS
            </p>
          </CardContent>
        </Card>

        <Card className="rounded-none border-l-4 border-l-amber-500">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Low Stock</CardTitle>
            <AlertTriangle className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">
              {inventory?.inventories?.filter(i => i.quantity < 10).length || 0}
            </div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              NEEDS REORDER
            </p>
          </CardContent>
        </Card>

        <Card className="rounded-none border-l-4 border-l-destructive">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs font-mono uppercase tracking-widest text-muted-foreground">Out of Stock</CardTitle>
            <MinusCircle className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold font-mono">
              {inventory?.inventories?.filter(i => i.quantity === 0).length || 0}
            </div>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              CRITICAL
            </p>
          </CardContent>
        </Card>
      </div>

      <Card className="rounded-none border border-border">
        <CardHeader className="border-b border-border bg-muted/10">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="SEARCH INVENTORY..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 rounded-none border-border bg-background font-mono text-sm uppercase placeholder:text-muted-foreground/50"
              />
            </div>
            <AdvancedFilters
              config={{
                statuses: {
                  label: 'Stock Status',
                  options: [
                    { value: 'in_stock', label: 'In Stock' },
                    { value: 'low_stock', label: 'Low Stock' },
                    { value: 'out_of_stock', label: 'Out of Stock' },
                  ],
                },
                quantityRange: {
                  label: 'Quantity',
                  minKey: 'min_quantity',
                  maxKey: 'max_quantity',
                },
                customFilters: [
                  {
                    key: 'warehouse_id',
                    label: 'Warehouse',
                    type: 'select',
                    options: (warehouses?.warehouses || []).map((wh) => ({ value: wh.id, label: wh.name })),
                  },
                  {
                    key: 'product_id',
                    label: 'Product',
                    type: 'select',
                    options: (products?.products || []).map((prod) => ({ value: prod.id, label: prod.name })),
                  },
                ],
              }}
              activeFilters={advancedFilters}
              onApply={setAdvancedFilters}
              onReset={() => setAdvancedFilters({})}
            />
          </div>
          <div className="mt-4">
            <ActiveFilterBadges
              filters={advancedFilters}
              onRemove={(key) => {
                const next = { ...advancedFilters };
                delete next[key];
                setAdvancedFilters(next);
              }}
            />
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="text-center py-12 text-muted-foreground font-mono animate-pulse">LOADING INVENTORY DATA...</div>
          ) : filteredInventory.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground font-mono">
              NO INVENTORY RECORDS FOUND
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent border-border">
                  <TableHead className="w-[50px]">
                    <Checkbox
                      checked={inventory?.inventories?.length ? selectedItems.size === inventory.inventories.length : false}
                      onCheckedChange={(checked: boolean) => handleSelectAll(checked)}
                    />
                  </TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10">Product</TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10">Warehouse</TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10">Quantity</TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10">Status</TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10">Last Updated</TableHead>
                  <TableHead className="font-mono text-xs uppercase tracking-wider h-10 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredInventory.map((item) => {
                  const isLowStock = item.quantity < 10;
                  const isOutOfStock = item.quantity === 0;

                  return (
                    <TableRow key={item.id} className="border-border hover:bg-muted/20 transition-colors group">
                      <TableCell>
                        <Checkbox
                          checked={selectedItems.has(item.id)}
                          onCheckedChange={(checked: boolean) => handleSelectItem(item.id, checked)}
                        />
                      </TableCell>
                      <TableCell className="font-medium font-mono text-sm">{item.product_name || 'Unknown'}</TableCell>
                      <TableCell className="font-mono text-sm text-muted-foreground">{item.warehouse_name || 'Unknown'}</TableCell>
                      <TableCell>
                        <span className={`font-mono font-bold ${isOutOfStock ? 'text-destructive' : isLowStock ? 'text-amber-500' : 'text-primary'}`}>
                          {item.quantity}
                        </span>
                      </TableCell>
                      <TableCell>
                        {isOutOfStock ? (
                          <div className="flex items-center text-destructive text-xs font-mono uppercase tracking-wider">
                            <AlertTriangle className="h-3 w-3 mr-1" />
                            Out of Stock
                          </div>
                        ) : isLowStock ? (
                          <div className="flex items-center text-amber-500 text-xs font-mono uppercase tracking-wider">
                            <AlertTriangle className="h-3 w-3 mr-1" />
                            Low Stock
                          </div>
                        ) : (
                          <span className="text-primary text-xs font-mono uppercase tracking-wider">In Stock</span>
                        )}
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">{formatDateTime(item.last_updated)}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setAdjustingInventory(item)}
                            className="h-7 text-xs font-mono uppercase"
                          >
                            <PlusCircle className="h-3 w-3 mr-1" />
                            Adjust
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditingInventory(item)}
                            className="h-7 w-7 p-0"
                          >
                            <Edit className="h-3 w-3" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <InventoryFormDialog
        open={isAddDialogOpen || !!editingInventory}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingInventory(null);
        }}
        inventory={editingInventory}
        products={products?.products || []}
        warehouses={warehouses?.warehouses || []}
      />

      <StockAdjustmentDialog
        open={!!adjustingInventory}
        onOpenChange={(open) => {
          if (!open) setAdjustingInventory(null);
        }}
        inventory={adjustingInventory}
      />

      <BarcodeScanner
        isOpen={isScannerOpen}
        onClose={() => setIsScannerOpen(false)}
        onScan={(code) => {
          setSearchQuery(code);
          setIsScannerOpen(false);
        }}
      />

      <BulkAdjustmentDialog
        open={isBulkAdjustmentOpen}
        onOpenChange={setIsBulkAdjustmentOpen}
        selectedItems={inventory?.inventories?.filter(item => selectedItems.has(item.id)) || []}
        onSuccess={() => setSelectedItems(new Set())}
      />
    </div>
  );
}

function InventoryFormDialog({
  open,
  onOpenChange,
  inventory,
  products,
  warehouses,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  inventory?: Inventory | null;
  products: Product[];
  warehouses: Warehouse[];
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    product_id: inventory?.product_id || '',
    warehouse_id: inventory?.warehouse_id || '',
    quantity: inventory?.quantity || 0,
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (inventory) {
        await api.put(`/inventory/${inventory.id}`, data);
      } else {
        await api.post('/inventory', data);
      }
    },
    onMutate: async (data) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['inventory'] });

      // Snapshot previous value
      const previousInventory = queryClient.getQueryData<{ inventories: InventoryWithDetails[] }>(['inventory']);

      // Optimistically update
      if (inventory) {
        // Update existing
        queryClient.setQueryData<{ inventories: InventoryWithDetails[] }>(
          ['inventory'],
          (old) => ({
            ...old,
            inventories: old?.inventories?.map((item) =>
              item.id === inventory.id ? { ...item, ...data } : item
            ) || [],
          })
        );
      } else {
        // Add new
        queryClient.setQueryData<{ inventories: InventoryWithDetails[] }>(
          ['inventory'],
          (old) => ({
            ...old,
            inventories: [data as Inventory, ...(old?.inventories || [])],
          })
        );
      }

      return { previousInventory };
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousInventory) {
        queryClient.setQueryData(['inventory'], context.previousInventory);
      }
      alert('Failed to save inventory. Please try again.');
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['inventory'], refetchType: 'none' });
      onOpenChange(false);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{inventory ? 'Update Stock' : 'Add Stock'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Product *</label>
            <select
              required
              value={formData.product_id}
              onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!!inventory}
            >
              <option value="">Select product</option>
              {products.map((product) => (
                <option key={product.id} value={product.id}>
                  {product.name}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Warehouse *</label>
            <select
              required
              value={formData.warehouse_id}
              onChange={(e) => setFormData({ ...formData, warehouse_id: e.target.value })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!!inventory}
            >
              <option value="">Select warehouse</option>
              {warehouses.map((warehouse) => (
                <option key={warehouse.id} value={warehouse.id}>
                  {warehouse.name}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Quantity *</label>
            <Input
              required
              type="number"
              value={formData.quantity}
              onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) })}
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function StockAdjustmentDialog({
  open,
  onOpenChange,
  inventory,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  inventory: Inventory | null;
}) {
  const queryClient = useQueryClient();
  const [adjustment, setAdjustment] = useState(0);
  const [reason, setReason] = useState('');

  const historyQuery = useQuery<{ history: InventoryHistoryEntry[] }>({
    queryKey: ['inventory-history', inventory?.id],
    queryFn: async () => {
      if (!inventory) return { history: [] };
      const response = await api.get(`/inventory/${inventory.id}/history?limit=25`);
      return response.data;
    },
    enabled: open && !!inventory,
  });

  const adjustMutation = useMutation({
    mutationFn: async () => {
      if (!inventory) return;
      await api.post('/inventory/adjust', {
        warehouse_id: inventory.warehouse_id,
        product_id: inventory.product_id,
        quantity_change: adjustment,
        reason,
      });
    },
    onMutate: async () => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['inventory'] });

      // Snapshot previous value
      const previousInventory = queryClient.getQueryData<{ inventories: InventoryWithDetails[] }>(['inventory']);

      // Optimistically update quantity
      queryClient.setQueryData<{ inventories: InventoryWithDetails[] }>(
        ['inventory'],
        (old) => ({
          ...old,
          inventories: old?.inventories?.map((item) =>
            item.id === inventory?.id
              ? { ...item, quantity: inventory.quantity + adjustment }
              : item
          ) || [],
        })
      );

      return { previousInventory };
    },
    onError: (error, variables, context) => {
      // Rollback on error
      if (context?.previousInventory) {
        queryClient.setQueryData(['inventory'], context.previousInventory);
      }
      const err = error as { response?: { data?: { error?: { message?: string } } } };
      alert(err.response?.data?.error?.message || 'Failed to adjust stock');
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['inventory'], refetchType: 'none' });
      if (inventory) {
        queryClient.invalidateQueries({ queryKey: ['inventory-history', inventory.id], refetchType: 'none' });
      }
      onOpenChange(false);
      setAdjustment(0);
      setReason('');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inventory) return;

    const newQuantity = inventory.quantity + adjustment;
    if (newQuantity < 0) {
      alert('Adjustment would result in negative stock. Please adjust the amount.');
      return;
    }

    if (confirm(`Adjust stock from ${inventory.quantity} to ${newQuantity}?`)) {
      adjustMutation.mutate();
    }
  };

  if (!inventory) return null;

  const newQuantity = inventory.quantity + adjustment;
  const history = historyQuery.data?.history || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Adjust Stock</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">Product</label>
            <div className="text-base font-semibold">{inventory.product_name || 'Unknown Product'}</div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">Warehouse</label>
            <div className="text-base font-semibold">{inventory.warehouse_name || 'Unknown Warehouse'}</div>
          </div>

          <div className="bg-muted/50 p-4 rounded-md">
            <label className="text-sm font-medium text-muted-foreground">Current Quantity</label>
            <div className="text-3xl font-bold mt-1">{inventory.quantity}</div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Adjustment *</label>
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setAdjustment(Math.max(adjustment - 10, -inventory.quantity))}
              >
                <MinusCircle className="h-4 w-4 mr-1" />
                -10
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setAdjustment(Math.max(adjustment - 1, -inventory.quantity))}
              >
                <MinusCircle className="h-4 w-4 mr-1" />
                -1
              </Button>
              <Input
                required
                type="number"
                value={adjustment}
                onChange={(e) => setAdjustment(parseInt(e.target.value) || 0)}
                className="text-center font-mono text-lg"
                placeholder="0"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setAdjustment(adjustment + 1)}
              >
                <PlusCircle className="h-4 w-4 mr-1" />
                +1
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setAdjustment(adjustment + 10)}
              >
                <PlusCircle className="h-4 w-4 mr-1" />
                +10
              </Button>
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              {adjustment > 0 && `Adding ${adjustment} units`}
              {adjustment < 0 && `Removing ${Math.abs(adjustment)} units`}
              {adjustment === 0 && 'Enter adjustment amount'}
            </p>
          </div>

          <div className={`p-4 rounded-md ${newQuantity < 0 ? 'bg-destructive/10' :
            newQuantity === 0 ? 'bg-amber-500/10' :
              'bg-green-500/10'
            }`}>
            <label className="text-sm font-medium text-muted-foreground">New Quantity</label>
            <div className={`text-3xl font-bold mt-1 ${newQuantity < 0 ? 'text-destructive' :
              newQuantity === 0 ? 'text-amber-600' :
                'text-green-600'
              }`}>
              {newQuantity}
            </div>
            {newQuantity < 0 && (
              <p className="text-sm text-destructive mt-1">⚠️ Cannot have negative stock</p>
            )}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Reason *</label>
            <Input
              required
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="e.g., Damaged goods, Stock count correction, Return from customer"
            />
            <p className="text-xs text-muted-foreground">This will be recorded in the audit log</p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">Recent Adjustments</label>
            {historyQuery.isLoading ? (
              <div className="text-sm text-muted-foreground">Loading history...</div>
            ) : history.length === 0 ? (
              <div className="text-sm text-muted-foreground">No past adjustments recorded.</div>
            ) : (
              <div className="max-h-60 overflow-y-auto divide-y divide-border rounded-md border">
                {history.map((entry) => {
                  const newValues = entry.new_values || {};
                  const oldValues = entry.old_values || {};
                  const change = (newValues.quantity_change as number) ?? 0;
                  const oldQty = oldValues.quantity as number | undefined;
                  const newQty = newValues.quantity as number | undefined;
                  const reasonText = newValues.reason as string | undefined;

                  return (
                    <div key={entry.id} className="p-3 text-sm">
                      <div className="flex items-center justify-between">
                        <span className="font-medium">{formatDateTime(entry.created_at)}</span>
                        <Badge variant={change >= 0 ? 'success' : 'destructive'}>
                          {change >= 0 ? `+${change}` : change}
                        </Badge>
                      </div>
                      <div className="mt-1 text-muted-foreground">
                        {oldQty !== undefined && newQty !== undefined ? (
                          <span>
                            Quantity {oldQty} → {newQty}
                          </span>
                        ) : (
                          <span>Quantity updated</span>
                        )}
                      </div>
                      {reasonText && (
                        <div className="mt-1 text-muted-foreground">Reason: {reasonText}</div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                onOpenChange(false);
                setAdjustment(0);
                setReason('');
              }}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={adjustMutation.isPending || adjustment === 0 || !reason || newQuantity < 0}
            >
              {adjustMutation.isPending ? 'Adjusting...' : 'Adjust Stock'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
