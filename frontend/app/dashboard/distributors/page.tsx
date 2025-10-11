'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, Trash2, Users } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import api from '@/lib/api';
import { Distributor } from '@/types';
import { formatDate } from '@/lib/utils';

export default function DistributorsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingDistributor, setEditingDistributor] = useState<Distributor | null>(null);
  const queryClient = useQueryClient();

  const { data: distributors, isLoading } = useQuery<{ distributors: Distributor[] }>({
    queryKey: ['distributors'],
    queryFn: async () => {
      const response = await api.get('/distributors?limit=100');
      return response.data;
    },
  });

  const deleteDistributor = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/distributors/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['distributors'] });
    },
  });

  const filteredDistributors = distributors?.distributors?.filter(distributor =>
    distributor.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    distributor.contact_email?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Distributors</h1>
          <p className="text-muted-foreground mt-1">Manage your distributor network</p>
        </div>
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Distributor
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search distributors..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading distributors...</div>
          ) : filteredDistributors.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Users className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
              <p>No distributors found. Add your first distributor to get started.</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Phone</TableHead>
                  <TableHead>Address</TableHead>
                  <TableHead>License</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDistributors.map((distributor) => (
                  <TableRow key={distributor.id}>
                    <TableCell className="font-medium">{distributor.name}</TableCell>
                    <TableCell>{distributor.contact_email || '-'}</TableCell>
                    <TableCell>{distributor.contact_phone || '-'}</TableCell>
                    <TableCell>{distributor.address || '-'}</TableCell>
                    <TableCell>{distributor.license_number || '-'}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingDistributor(distributor)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm('Are you sure you want to delete this distributor?')) {
                              deleteDistributor.mutate(distributor.id);
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

      <DistributorFormDialog
        open={isAddDialogOpen || !!editingDistributor}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingDistributor(null);
        }}
        distributor={editingDistributor}
      />
    </div>
  );
}

function DistributorFormDialog({ open, onOpenChange, distributor }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  distributor?: Distributor | null;
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: distributor?.name || '',
    contact_email: distributor?.contact_email || '',
    contact_phone: distributor?.contact_phone || '',
    address: distributor?.address || '',
    license_number: distributor?.license_number || '',
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (distributor) {
        await api.put(`/distributors/${distributor.id}`, data);
      } else {
        await api.post('/distributors', data);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['distributors'] });
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
          <DialogTitle>{distributor ? 'Edit Distributor' : 'Add Distributor'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Distributor Name *</label>
            <Input
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Farm Distribution Co"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Email</label>
              <Input
                type="email"
                value={formData.contact_email}
                onChange={(e) => setFormData({ ...formData, contact_email: e.target.value })}
                placeholder="sales@distributor.com"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Phone</label>
              <Input
                value={formData.contact_phone}
                onChange={(e) => setFormData({ ...formData, contact_phone: e.target.value })}
                placeholder="+91-9876543210"
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Address</label>
            <Input
              value={formData.address}
              onChange={(e) => setFormData({ ...formData, address: e.target.value })}
              placeholder="789 Distributor Road, City"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">License Number</label>
            <Input
              value={formData.license_number}
              onChange={(e) => setFormData({ ...formData, license_number: e.target.value })}
              placeholder="DIST001"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save Distributor'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
