'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, AlertTriangle, PlusCircle, MinusCircle, Package } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import AdvancedFilters, { ActiveFilterBadges } from '@/components/filters/AdvancedFilters';
import api from '@/lib/api';
import { Inventory, Product, Warehouse } from '@/types';
import { formatDateTime } from '@/lib/utils';

interface InventoryWithDetails extends Inventory {
  product?: Product;
  warehouse?: Warehouse;
}

type InventoryHistoryEntry = {
  id: string;
  action: string;
  created_at: string;
  new_values?: Record<string, any>;
  old_values?: Record<string, any>;
  changed_by?: string | null;
};

export default function InventoryPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingInventory, setEditingInventory] = useState<Inventory | null>(null);
  const [adjustingInventory, setAdjustingInventory] = useState<Inventory | null>(null);
  const [advancedFilters, setAdvancedFilters] = useState<Record<string, any>>({});
  const queryClient = useQueryClient();

  const { data: inventory, isLoading } = useQuery<{ inventory: InventoryWithDetails[] }>({
    queryKey: ['inventory'],
    queryFn: async () => {
      const response = await api.get('/inventory?limit=100');
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

  const filteredInventory = inventory?.inventory?.filter(item => {
    const product = products?.products?.find(p => p.id === item.product_id);
    const warehouse = warehouses?.warehouses?.find(w => w.id === item.warehouse_id);
    const productName = product?.name || '';
    const warehouseName = warehouse?.name || '';
    const query = searchQuery.toLowerCase();

    const matchesSearch =
      productName.toLowerCase().includes(query) ||
      warehouseName.toLowerCase().includes(query);

    if (!matchesSearch) {
      return false;
    }

    if (advancedFilters.warehouse_id && item.warehouse_id !== advancedFilters.warehouse_id) {
      return false;
    }

    if (advancedFilters.product_id && item.product_id !== advancedFilters.product_id) {
      return false;
    }

    if (advancedFilters.min_quantity) {
      const minQuantity = parseInt(advancedFilters.min_quantity, 10);
      if (!Number.isNaN(minQuantity) && item.quantity < minQuantity) {
        return false;
      }
    }

    if (advancedFilters.max_quantity) {
      const maxQuantity = parseInt(advancedFilters.max_quantity, 10);
      if (!Number.isNaN(maxQuantity) && item.quantity > maxQuantity) {
        return false;
      }
    }

    if (advancedFilters.statuses && advancedFilters.statuses.length > 0) {
      const status = item.quantity === 0
        ? 'out_of_stock'
        : item.quantity < 10
          ? 'low_stock'
          : 'in_stock';
      if (!advancedFilters.statuses.includes(status)) {
        return false;
      }
    }

    return true;
  }) || [];

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Inventory</h1>
          <p className="text-muted-foreground">Track stock levels across warehouses</p>
        </div>
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Stock
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Items</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{inventory?.inventory?.length || 0}</div>
            <p className="text-xs text-muted-foreground">
              Unique SKUs
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Low Stock Items</CardTitle>
            <AlertTriangle className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {inventory?.inventory?.filter(i => i.quantity < 10).length || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Needs reordering
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Out of Stock</CardTitle>
            <MinusCircle className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {inventory?.inventory?.filter(i => i.quantity === 0).length || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Critical status
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search by product or warehouse..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
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
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading inventory...</div>
          ) : filteredInventory.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No inventory records found. Add stock to get started.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Product</TableHead>
                  <TableHead>Warehouse</TableHead>
                  <TableHead>Quantity</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Updated</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredInventory.map((item) => {
                  const product = products?.products?.find(p => p.id === item.product_id);
                  const warehouse = warehouses?.warehouses?.find(w => w.id === item.warehouse_id);
                  const isLowStock = item.quantity < 10;
                  const isOutOfStock = item.quantity === 0;

                  return (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{product?.name || 'Unknown'}</TableCell>
                      <TableCell>{warehouse?.name || 'Unknown'}</TableCell>
                      <TableCell>
                        <Badge variant={isOutOfStock ? 'destructive' : isLowStock ? 'warning' : 'success'}>
                          {item.quantity}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {isOutOfStock ? (
                          <div className="flex items-center text-destructive">
                            <AlertTriangle className="h-4 w-4 mr-1" />
                            Out of Stock
                          </div>
                        ) : isLowStock ? (
                          <div className="flex items-center text-amber-600">
                            <AlertTriangle className="h-4 w-4 mr-1" />
                            Low Stock
                          </div>
                        ) : (
                          <span className="text-green-600">In Stock</span>
                        )}
                      </TableCell>
                      <TableCell>{formatDateTime(item.last_updated)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setAdjustingInventory(item)}
                          >
                            <PlusCircle className="h-4 w-4 mr-1" />
                            Adjust
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditingInventory(item)}
                          >
                            <Edit className="h-4 w-4" />
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
        product={products?.products?.find(p => p.id === adjustingInventory?.product_id)}
        warehouse={warehouses?.warehouses?.find(w => w.id === adjustingInventory?.warehouse_id)}
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
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
  product,
  warehouse,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  inventory: Inventory | null;
  product?: Product;
  warehouse?: Warehouse;
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      if (inventory) {
        queryClient.invalidateQueries({ queryKey: ['inventory-history', inventory.id] });
      }
      onOpenChange(false);
      setAdjustment(0);
      setReason('');
    },
    onError: (error: any) => {
      alert(error.response?.data?.error?.message || 'Failed to adjust stock');
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
            <div className="text-base font-semibold">{product?.name || 'Unknown Product'}</div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-muted-foreground">Warehouse</label>
            <div className="text-base font-semibold">{warehouse?.name || 'Unknown Warehouse'}</div>
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
                  const change = newValues.quantity_change ?? 0;
                  const oldQty = oldValues.quantity;
                  const newQty = newValues.quantity;
                  const reasonText = newValues.reason;

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
