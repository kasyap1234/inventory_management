'use client';

import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, Trash2, RefreshCcw, ShieldCheck, AlertCircle, Users } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { tenantService } from '@/lib/services';
import type { Tenant } from '@/types';
import { formatDate } from '@/lib/utils';

type TenantFormState = {
  name: string;
  subdomain: string;
  license: string;
  admin_email: string;
  status: string;
};

export default function TenantsPage() {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingTenant, setEditingTenant] = useState<Tenant | null>(null);

  const { data, isLoading, isFetching } = useQuery<{ tenants: Tenant[] }>({
    queryKey: ['tenants'],
    queryFn: async () => {
      const response = await tenantService.list();
      return response.data;
    },
    staleTime: 30_000,
  });

  const tenants = data?.tenants ?? [];

  const filteredTenants = useMemo(() => {
    if (!searchQuery) {
      return tenants;
    }
    const query = searchQuery.toLowerCase();
    return tenants.filter((tenant) =>
      tenant.name.toLowerCase().includes(query) ||
      tenant.subdomain.toLowerCase().includes(query) ||
      tenant.license_number?.toLowerCase().includes(query)
    );
  }, [tenants, searchQuery]);

  const createTenant = useMutation({
    mutationFn: async (payload: { name: string; subdomain: string; license: string; admin_email: string }) => {
      // @ts-ignore - tenantService.create definition might need update in services.ts or we cast
      await tenantService.create(payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
      setIsCreateDialogOpen(false);
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to create tenant');
    },
  });

  const updateTenant = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Partial<TenantFormState> }) => {
      await tenantService.update(id, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
      setEditingTenant(null);
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to update tenant');
    },
  });

  const deleteTenant = useMutation({
    mutationFn: async (id: string) => {
      await tenantService.delete(id);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to delete tenant');
    },
  });

  const handleCreateTenant = (form: TenantFormState) => {
    createTenant.mutate({
      name: form.name.trim(),
      subdomain: form.subdomain.trim().toLowerCase(),
      license: form.license.trim(),
      admin_email: form.admin_email.trim(),
    });
  };

  const handleUpdateTenant = (id: string, form: TenantFormState) => {
    updateTenant.mutate({
      id,
      data: {
        name: form.name.trim(),
        subdomain: form.subdomain.trim().toLowerCase(),
        license: form.license.trim(),
        status: form.status,
      },
    });
  };

  const statusBadgeVariant = (status?: string | null) => {
    switch ((status ?? '').toLowerCase()) {
      case 'active':
        return 'success';
      case 'suspended':
      case 'inactive':
        return 'warning';
      case 'cancelled':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Tenants</h1>
          <p className="text-gray-500 mt-1">Manage customer accounts and licensing</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries({ queryKey: ['tenants'] })}
            disabled={isFetching}
          >
            <RefreshCcw className={`h-4 w-4 mr-2 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            New Tenant
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-semibold text-foreground flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-blue-600" />
            Tenant Directory
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search by name, subdomain, or license..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>

          {isLoading ? (
            <div className="text-center py-10">Loading tenants...</div>
          ) : filteredTenants.length === 0 ? (
            <div className="text-center py-16 text-gray-500">
              <AlertCircle className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>No tenants found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Subdomain</TableHead>
                  <TableHead>License</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="w-48">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTenants.map((tenant) => (
                  <TableRow key={tenant.id}>
                    <TableCell className="font-medium">{tenant.name}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{tenant.subdomain}.agromart.example</Badge>
                    </TableCell>
                    <TableCell>{tenant.license_number || '—'}</TableCell>
                    <TableCell>
                      <Badge variant={statusBadgeVariant(tenant.status)}>
                        {tenant.status || 'unknown'}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(tenant.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            // Navigate to users page with tenant filter
                            window.location.href = `/dashboard/users?tenant=${tenant.id}`;
                          }}
                          title="View users in this tenant"
                        >
                          <Users className="h-4 w-4 mr-1" />
                          Users
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingTenant(tenant)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm(`Delete tenant ${tenant.name}?`)) {
                              deleteTenant.mutate(tenant.id);
                            }
                          }}
                          disabled={deleteTenant.isPending}
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

      <TenantFormDialog
        title="Create Tenant"
        open={isCreateDialogOpen}
        onOpenChange={(open) => {
          setIsCreateDialogOpen(open);
        }}
        onSubmit={handleCreateTenant}
        isSubmitting={createTenant.isPending}
      />

      <TenantFormDialog
        title={editingTenant ? `Update ${editingTenant.name}` : 'Update Tenant'}
        open={!!editingTenant}
        onOpenChange={(open) => {
          if (!open) setEditingTenant(null);
        }}
        onSubmit={(form) => {
          if (editingTenant) {
            handleUpdateTenant(editingTenant.id, form);
          }
        }}
        tenant={editingTenant}
        isSubmitting={updateTenant.isPending}
        mode="edit"
      />
    </div>
  );
}

function TenantFormDialog({
  open,
  onOpenChange,
  title,
  onSubmit,
  tenant,
  isSubmitting,
  mode = 'create',
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  onSubmit: (form: TenantFormState) => void;
  tenant?: Tenant | null;
  isSubmitting: boolean;
  mode?: 'create' | 'edit';
}) {
  const [formState, setFormState] = useState<TenantFormState>({
    name: tenant?.name ?? '',
    subdomain: tenant?.subdomain ?? '',
    license: tenant?.license_number ?? '',
    admin_email: '',
    status: tenant?.status ?? 'active',
  });

  useEffect(() => {
    setFormState({
      name: tenant?.name ?? '',
      subdomain: tenant?.subdomain ?? '',
      license: tenant?.license_number ?? '',
      admin_email: '',
      status: tenant?.status ?? 'active',
    });
  }, [tenant]);

  const resetState = () => {
    setFormState({
      name: tenant?.name ?? '',
      subdomain: tenant?.subdomain ?? '',
      license: tenant?.license_number ?? '',
      admin_email: '',
      status: tenant?.status ?? 'active',
    });
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        onOpenChange(next);
        if (!next) {
          resetState();
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault();
            onSubmit(formState);
          }}
        >
          <div className="space-y-2">
            <label className="text-sm font-medium">Tenant Name *</label>
            <Input
              required
              value={formState.name}
              onChange={(e) => setFormState((prev) => ({ ...prev, name: e.target.value }))}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Subdomain *</label>
            <Input
              required
              value={formState.subdomain}
              onChange={(e) => setFormState((prev) => ({ ...prev, subdomain: e.target.value }))}
              placeholder="acme"
            />
            <p className="text-xs text-gray-500">Final URL: https://{formState.subdomain || 'subdomain'}.agromart.example</p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">License Number *</label>
            <Input
              required
              value={formState.license}
              onChange={(e) => setFormState((prev) => ({ ...prev, license: e.target.value }))}
              placeholder="LIC-12345"
            />
          </div>

          {mode === 'create' && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Admin Email *</label>
              <Input
                required
                type="email"
                value={formState.admin_email}
                onChange={(e) => setFormState((prev) => ({ ...prev, admin_email: e.target.value }))}
                placeholder="admin@example.com"
              />
            </div>
          )}

          {mode === 'edit' && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Status</label>
              <select
                value={formState.status}
                onChange={(e) => setFormState((prev) => ({ ...prev, status: e.target.value }))}
                className="flex h-10 w-full rounded-md border border-input bg-background text-foreground px-3 py-2 text-sm"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="suspended">Suspended</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
