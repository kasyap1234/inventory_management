'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, AlertTriangle, PlusCircle, MinusCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import api from '@/lib/api';
import { Inventory, Product, Warehouse } from '@/types';
import { formatDateTime } from '@/lib/utils';

interface InventoryWithDetails extends Inventory {
  product?: Product;
  warehouse?: Warehouse;
}

export default function InventoryPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingInventory, setEditingInventory] = useState<Inventory | null>(null);
  const [adjustingInventory, setAdjustingInventory] = useState<Inventory | null>(null);
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
    const productName = products?.products?.find(p => p.id === item.product_id)?.name || '';
    const warehouseName = warehouses?.warehouses?.find(w => w.id === item.warehouse_id)?.name || '';
    const query = searchQuery.toLowerCase();
    return productName.toLowerCase().includes(query) || warehouseName.toLowerCase().includes(query);
  }) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Inventory</h1>
          <p className="text-gray-500 mt-1">Track stock levels across warehouses</p>
        </div>
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Stock
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">
              Total Items
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{inventory?.inventory?.length || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">
              Low Stock Items
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">
              {inventory?.inventory?.filter(i => i.quantity < 10).length || 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">
              Out of Stock
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {inventory?.inventory?.filter(i => i.quantity === 0).length || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search by product or warehouse..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading inventory...</div>
          ) : filteredInventory.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
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
                        <Badge variant={isOutOfStock ? 'danger' : isLowStock ? 'warning' : 'success'}>
                          {item.quantity}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {isOutOfStock ? (
                          <div className="flex items-center text-red-600">
                            <AlertTriangle className="h-4 w-4 mr-1" />
                            Out of Stock
                          </div>
                        ) : isLowStock ? (
                          <div className="flex items-center text-orange-600">
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
              className="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
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
              className="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
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

  const adjustMutation = useMutation({
    mutationFn: async () => {
      if (!inventory) return;
      await api.post(`/inventory/${inventory.id}/adjust`, {
        adjustment,
        reason,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Adjust Stock</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-600">Product</label>
            <div className="text-base font-semibold">{product?.name || 'Unknown Product'}</div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-600">Warehouse</label>
            <div className="text-base font-semibold">{warehouse?.name || 'Unknown Warehouse'}</div>
          </div>

          <div className="bg-gray-50 p-4 rounded-md">
            <label className="text-sm font-medium text-gray-600">Current Quantity</label>
            <div className="text-3xl font-bold text-gray-900 mt-1">{inventory.quantity}</div>
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
            <p className="text-sm text-gray-500 mt-1">
              {adjustment > 0 && `Adding ${adjustment} units`}
              {adjustment < 0 && `Removing ${Math.abs(adjustment)} units`}
              {adjustment === 0 && 'Enter adjustment amount'}
            </p>
          </div>

          <div className={`p-4 rounded-md ${
            newQuantity < 0 ? 'bg-red-50' : 
            newQuantity === 0 ? 'bg-orange-50' : 
            'bg-green-50'
          }`}>
            <label className="text-sm font-medium text-gray-600">New Quantity</label>
            <div className={`text-3xl font-bold mt-1 ${
              newQuantity < 0 ? 'text-red-600' : 
              newQuantity === 0 ? 'text-orange-600' : 
              'text-green-600'
            }`}>
              {newQuantity}
            </div>
            {newQuantity < 0 && (
              <p className="text-sm text-red-600 mt-1">⚠️ Cannot have negative stock</p>
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
            <p className="text-xs text-gray-500">This will be recorded in the audit log</p>
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
