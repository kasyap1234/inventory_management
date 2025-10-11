'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, Trash2, Building2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import api from '@/lib/api';
import { Warehouse } from '@/types';
import { formatDate } from '@/lib/utils';

export default function WarehousesPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingWarehouse, setEditingWarehouse] = useState<Warehouse | null>(null);
  const queryClient = useQueryClient();

  const { data: warehouses, isLoading } = useQuery<{ warehouses: Warehouse[] }>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const response = await api.get('/warehouses?limit=100');
      return response.data;
    },
  });

  const deleteWarehouse = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/warehouses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['warehouses'] });
    },
  });

  const filteredWarehouses = warehouses?.warehouses?.filter(warehouse =>
    warehouse.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    warehouse.address?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Warehouses</h1>
          <p className="text-muted-foreground mt-1">Manage storage locations</p>
        </div>
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Warehouse
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search warehouses..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading warehouses...</div>
          ) : filteredWarehouses.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Building2 className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
              <p>No warehouses found. Add your first warehouse to get started.</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Address</TableHead>
                  <TableHead>Capacity</TableHead>
                  <TableHead>License Number</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredWarehouses.map((warehouse) => (
                  <TableRow key={warehouse.id}>
                    <TableCell className="font-medium">{warehouse.name}</TableCell>
                    <TableCell>{warehouse.address || '-'}</TableCell>
                    <TableCell>{warehouse.capacity || '-'}</TableCell>
                    <TableCell>{warehouse.license_number || '-'}</TableCell>
                    <TableCell>{formatDate(warehouse.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingWarehouse(warehouse)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm('Are you sure you want to delete this warehouse?')) {
                              deleteWarehouse.mutate(warehouse.id);
                            }
                          }}
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <WarehouseFormDialog
        open={isAddDialogOpen || !!editingWarehouse}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingWarehouse(null);
        }}
        warehouse={editingWarehouse}
      />
    </div>
  );
}

function WarehouseFormDialog({ open, onOpenChange, warehouse }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  warehouse?: Warehouse | null;
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: warehouse?.name || '',
    address: warehouse?.address || '',
    capacity: warehouse?.capacity || 0,
    license_number: warehouse?.license_number || '',
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (warehouse) {
        await api.put(`/warehouses/${warehouse.id}`, data);
      } else {
        await api.post('/warehouses', data);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['warehouses'] });
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
          <DialogTitle>{warehouse ? 'Edit Warehouse' : 'Add Warehouse'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Warehouse Name *</label>
            <Input
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Main Warehouse"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Address</label>
            <Input
              value={formData.address}
              onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              placeholder="123 Industrial Area, City"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Capacity</label>
            <Input
              type="number"
              value={formData.capacity}
              onChange={(e) => setFormData({ ...formData, capacity: parseInt(e.target.value) })}
              placeholder="10000"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">License Number</label>
            <Input
              value={formData.license_number}
              onChange={(e) => setFormData({ ...formData, license_number: e.target.value })}
              placeholder="WH001"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save Warehouse'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
