'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useDebounce } from '@/hooks/useDebounce';
import { Plus, Search, Edit, Trash2, Shield, Users as UsersIcon, Key, CheckCircle, Clock, XCircle, Mail, RotateCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import api from '@/lib/api';
import { tenantService, invitationService } from '@/lib/services';
import { User, Role, Permission, Tenant, Invitation } from '@/types';
import { formatDate } from '@/lib/utils';
import { useAuth } from '@/hooks/useAuth';

type TabType = 'users' | 'pending' | 'roles' | 'permissions' | 'invitations';


export default function UsersPage() {
  const [activeTab, setActiveTab] = useState<TabType>('users');
  const [searchQuery, setSearchQuery] = useState('');

  // Debounce search query to reduce API calls
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users & Roles</h1>
          <p className="text-muted-foreground">Manage users, roles, and permissions</p>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('users')}
            className={`${activeTab === 'users'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 transition-colors`}
          >
            <UsersIcon className="h-5 w-5" />
            Users
          </button>
          <button
            onClick={() => setActiveTab('pending')}
            className={`${activeTab === 'pending'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 transition-colors`}
          >
            <Clock className="h-5 w-5" />
            Pending Approval
          </button>
          <button
            onClick={() => setActiveTab('invitations')}
            className={`${activeTab === 'invitations'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 transition-colors`}
          >
            <Mail className="h-5 w-5" />
            Invitations
          </button>

          <button
            onClick={() => setActiveTab('roles')}
            className={`${activeTab === 'roles'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 transition-colors`}
          >
            <Shield className="h-5 w-5" />
            Roles
          </button>
          <button
            onClick={() => setActiveTab('permissions')}
            className={`${activeTab === 'permissions'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 transition-colors`}
          >
            <Key className="h-5 w-5" />
            Permissions
          </button>
        </nav>
      </div>

      {/* Search Bar */}
      <div className="flex items-center gap-4">
        <div className="flex-1 relative max-w-sm">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={`Search ${activeTab}...`}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'users' && <UsersTab searchQuery={debouncedSearchQuery} />}
      {activeTab === 'pending' && <PendingApprovalTab searchQuery={debouncedSearchQuery} />}
      {activeTab === 'invitations' && <InvitationsTab searchQuery={debouncedSearchQuery} />}
      {activeTab === 'roles' && <RolesTab searchQuery={debouncedSearchQuery} />}
      {activeTab === 'permissions' && <PermissionsTab searchQuery={debouncedSearchQuery} />}

    </div>
  );
}

function UsersTab({ searchQuery }: { searchQuery: string }) {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [assigningRoles, setAssigningRoles] = useState<User | null>(null);
  const queryClient = useQueryClient();

  const { data: users, isLoading } = useQuery<{ users: User[] }>({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await api.get('/users');
      return response.data;
    },
  });

  const { data: roles } = useQuery<{ roles: Role[] }>({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await api.get('/roles');
      return response.data;
    },
  });

  const deleteUser = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/users/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const approveUser = useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/users/${id}/approve`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['pending-users'] });
    },
  });

  const filteredUsers = users?.users?.filter(user =>
    user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.first_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.last_name.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <>
      <div className="flex justify-end mb-4 gap-2">
        <Button variant="outline" onClick={() => setIsInviteDialogOpen(true)}>
          <Mail className="h-4 w-4 mr-2" />
          Invite User
        </Button>
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add User
        </Button>
      </div>


      <Card>
        <CardContent className="pt-6">
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading users...</div>
          ) : filteredUsers.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No users found. Add your first user to get started.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">
                      {user.first_name} {user.last_name}
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={user.status === 'active' ? 'success' : user.status === 'pending_approval' ? 'warning' : 'secondary'}>
                        {user.status === 'pending_approval' ? 'Pending Approval' : user.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(user.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {user.status === 'pending_approval' && (
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => {
                              if (confirm(`Approve ${user.first_name} ${user.last_name}?`)) {
                                approveUser.mutate(user.id);
                              }
                            }}
                            disabled={approveUser.isPending}
                          >
                            <CheckCircle className="h-4 w-4 mr-1" />
                            Approve
                          </Button>
                        )}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setAssigningRoles(user)}
                        >
                          <Shield className="h-4 w-4 mr-1" />
                          Roles
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingUser(user)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm('Are you sure you want to delete this user?')) {
                              deleteUser.mutate(user.id);
                            }
                          }}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
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

      <UserFormDialog
        open={isAddDialogOpen || !!editingUser}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingUser(null);
        }}
        user={editingUser}
      />

      <InviteUserDialog
        open={isInviteDialogOpen}
        onOpenChange={setIsInviteDialogOpen}
        roles={roles?.roles || []}
      />

      <AssignRolesDialog
        open={!!assigningRoles}
        onOpenChange={(open) => {
          if (!open) setAssigningRoles(null);
        }}
        user={assigningRoles}
        roles={roles?.roles || []}
      />
    </>
  );
}

// Pending Approval Tab Component
function PendingApprovalTab({ searchQuery }: { searchQuery: string }) {
  const queryClient = useQueryClient();

  const { data: pendingUsers, isLoading, error } = useQuery<{ users: User[] }>({
    queryKey: ['pending-users'],
    queryFn: async () => {
      const response = await api.get('/users/pending');
      return response.data;
    },
  });

  const approveUser = useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/users/${id}/approve`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-users'] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const rejectUser = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/users/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-users'] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const filteredUsers = pendingUsers?.users?.filter(user =>
    user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.first_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.last_name.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="mb-4">
          <h3 className="text-lg font-semibold">Pending User Approvals</h3>
          <p className="text-sm text-muted-foreground">
            Review and approve new user registrations
          </p>
        </div>

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading pending users...</div>
        ) : error ? (
          <div className="text-center py-8 text-destructive">
            Failed to load pending users. You may not have permission to view this.
          </div>
        ) : filteredUsers.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <CheckCircle className="h-12 w-12 mx-auto mb-4 text-green-500" />
            <p className="font-medium">No pending approvals</p>
            <p className="text-sm">All user registrations have been processed.</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Registered</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredUsers.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">
                    {user.first_name} {user.last_name}
                  </TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>{formatDate(user.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="default"
                        size="sm"
                        onClick={() => {
                          if (confirm(`Approve ${user.first_name} ${user.last_name}?`)) {
                            approveUser.mutate(user.id);
                          }
                        }}
                        disabled={approveUser.isPending}
                      >
                        <CheckCircle className="h-4 w-4 mr-1" />
                        Approve
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => {
                          if (confirm(`Reject and delete ${user.first_name} ${user.last_name}? This action cannot be undone.`)) {
                            rejectUser.mutate(user.id);
                          }
                        }}
                        disabled={rejectUser.isPending}
                      >
                        <XCircle className="h-4 w-4 mr-1" />
                        Reject
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
  );
}

function RolesTab({ searchQuery }: { searchQuery: string }) {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [managingPermissions, setManagingPermissions] = useState<Role | null>(null);
  const queryClient = useQueryClient();

  const { data: roles, isLoading } = useQuery<{ roles: Role[] }>({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await api.get('/roles');
      return response.data;
    },
  });

  const { data: permissions } = useQuery<{ permissions: Permission[] }>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const response = await api.get('/permissions');
      return response.data;
    },
  });

  const deleteRole = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/roles/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
    },
  });

  const filteredRoles = roles?.roles?.filter(role =>
    role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    role.description?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <>
      <div className="flex justify-end mb-4">
        <Button onClick={() => setIsAddDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Role
        </Button>
      </div>

      <Card>
        <CardContent className="pt-6">
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading roles...</div>
          ) : filteredRoles.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No roles found. Add your first role to get started.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRoles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium">{role.name}</TableCell>
                    <TableCell>{role.description || '-'}</TableCell>
                    <TableCell>{formatDate(role.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setManagingPermissions(role)}
                        >
                          <Key className="h-4 w-4 mr-1" />
                          Permissions
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingRole(role)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm('Are you sure you want to delete this role?')) {
                              deleteRole.mutate(role.id);
                            }
                          }}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
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

      <RoleFormDialog
        open={isAddDialogOpen || !!editingRole}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingRole(null);
        }}
        role={editingRole}
      />

      <ManagePermissionsDialog
        open={!!managingPermissions}
        onOpenChange={(open) => {
          if (!open) setManagingPermissions(null);
        }}
        role={managingPermissions}
        permissions={permissions?.permissions || []}
      />
    </>
  );
}

function PermissionsTab({ searchQuery }: { searchQuery: string }) {
  const { data: permissions, isLoading } = useQuery<{ permissions: Permission[] }>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const response = await api.get('/permissions');
      return response.data;
    },
  });

  const filteredPermissions = permissions?.permissions?.filter(permission =>
    permission.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    permission.resource.toLowerCase().includes(searchQuery.toLowerCase()) ||
    permission.action.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  // Group permissions by resource
  const groupedPermissions = filteredPermissions.reduce((acc, permission) => {
    if (!acc[permission.resource]) {
      acc[permission.resource] = [];
    }
    acc[permission.resource].push(permission);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <Card>
      <CardContent className="pt-6">
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading permissions...</div>
        ) : filteredPermissions.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No permissions found.
          </div>
        ) : (
          <div className="space-y-6">
            {Object.entries(groupedPermissions).map(([resource, perms]) => (
              <div key={resource}>
                <h3 className="text-lg font-semibold mb-3 capitalize">
                  {resource}
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                  {perms.map((permission) => (
                    <div
                      key={permission.id}
                      className="border rounded-lg p-4 hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <Badge variant="secondary">{permission.action}</Badge>
                        <Key className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <p className="mt-2 text-sm font-medium">
                        {permission.name}
                      </p>
                      {permission.description && (
                        <p className="mt-1 text-xs text-muted-foreground">
                          {permission.description}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// User Form Dialog
function UserFormDialog({ open, onOpenChange, user }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user?: User | null;
}) {
  const queryClient = useQueryClient();
  const { user: currentUser } = useAuth();
  const isSuperAdmin = currentUser?.role === 'super_admin';

  const [formData, setFormData] = useState({
    email: user?.email || '',
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    tenant_id: user?.tenant_id || currentUser?.tenant_id || '',
    password: '',
    status: user?.status || 'active',
  });

  const { data: tenantsData } = useQuery<{ tenants: Tenant[] }>({
    queryKey: ['tenants'],
    queryFn: async () => {
      const response = await tenantService.list();
      return response.data;
    },
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (user) {
        await api.put(`/users/${user.id}`, data);
      } else {
        await api.post('/users', data);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
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
          <DialogTitle>{user ? 'Edit User' : 'Add User'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">First Name *</label>
              <Input
                required
                value={formData.first_name}
                onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Last Name *</label>
              <Input
                required
                value={formData.last_name}
                onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Email *</label>
            <Input
              required
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            />
          </div>
          {isSuperAdmin && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Tenant *</label>
              <select
                required
                value={formData.tenant_id}
                onChange={(e) => setFormData({ ...formData, tenant_id: e.target.value })}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                disabled={!!user}
              >
                <option value="">Select a tenant</option>
                {tenantsData?.tenants?.map((tenant) => (
                  <option key={tenant.id} value={tenant.id}>
                    {tenant.name} ({tenant.subdomain})
                  </option>
                ))}
              </select>
              {user && (
                <p className="text-xs text-muted-foreground">Tenant cannot be changed after user creation</p>
              )}
            </div>
          )}
          {!user && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Password *</label>
              <Input
                required
                type="password"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                placeholder="Minimum 8 characters"
              />
            </div>
          )}
          <div className="space-y-2">
            <label className="text-sm font-medium">Status *</label>
            <select
              required
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save User'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Role Form Dialog
function RoleFormDialog({ open, onOpenChange, role }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role?: Role | null;
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: role?.name || '',
    description: role?.description || '',
  });

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      if (role) {
        await api.put(`/roles/${role.id}`, data);
      } else {
        await api.post('/roles', data);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
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
          <DialogTitle>{role ? 'Edit Role' : 'Add Role'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Role Name *</label>
            <Input
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g., Manager, Operator"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Input
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Role description"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save Role'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Assign Roles Dialog
function AssignRolesDialog({ open, onOpenChange, user, roles }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user: User | null;
  roles: Role[];
}) {
  const queryClient = useQueryClient();
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch user's current roles
  useQuery({
    queryKey: ['user-roles', user?.id],
    queryFn: async () => {
      if (!user) return null;
      const response = await api.get(`/users/${user.id}/roles`);
      setSelectedRoles(response.data.roles?.map((r: Role) => r.id) || []);
      return response.data;
    },
    enabled: !!user && open,
  });

  const handleToggleRole = (roleId: string) => {
    setSelectedRoles(prev =>
      prev.includes(roleId)
        ? prev.filter(id => id !== roleId)
        : [...prev, roleId]
    );
  };

  const handleSave = async () => {
    if (!user) return;
    setLoading(true);
    try {
      await api.post(`/users/${user.id}/roles`, { role_ids: selectedRoles });
      queryClient.invalidateQueries({ queryKey: ['user-roles', user.id] });
      onOpenChange(false);
    } catch (error) {
      alert('Failed to assign roles');
    } finally {
      setLoading(false);
    }
  };

  if (!user) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Assign Roles to {user.first_name} {user.last_name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="text-sm text-muted-foreground">
            Select roles to assign to this user
          </div>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {roles.map((role) => (
              <label
                key={role.id}
                className="flex items-center p-3 border rounded-lg hover:bg-muted/50 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={selectedRoles.includes(role.id)}
                  onChange={() => handleToggleRole(role.id)}
                  className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
                />
                <div className="ml-3 flex-1">
                  <div className="font-medium text-sm">{role.name}</div>
                  {role.description && (
                    <div className="text-xs text-muted-foreground">{role.description}</div>
                  )}
                </div>
              </label>
            ))}
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={loading}>
              {loading ? 'Saving...' : 'Save Roles'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Manage Permissions Dialog
function ManagePermissionsDialog({ open, onOpenChange, role, permissions }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role: Role | null;
  permissions: Permission[];
}) {
  const queryClient = useQueryClient();
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch role's current permissions
  useQuery({
    queryKey: ['role-permissions', role?.id],
    queryFn: async () => {
      if (!role) return null;
      const response = await api.get(`/roles/${role.id}/permissions`);
      setSelectedPermissions(response.data.permissions?.map((p: Permission) => p.id) || []);
      return response.data;
    },
    enabled: !!role && open,
  });

  const handleTogglePermission = (permissionId: string) => {
    setSelectedPermissions(prev =>
      prev.includes(permissionId)
        ? prev.filter(id => id !== permissionId)
        : [...prev, permissionId]
    );
  };

  const handleSave = async () => {
    if (!role) return;
    setLoading(true);
    try {
      await api.post(`/roles/${role.id}/permissions`, { permission_ids: selectedPermissions });
      queryClient.invalidateQueries({ queryKey: ['role-permissions', role.id] });
      onOpenChange(false);
    } catch (error) {
      alert('Failed to assign permissions');
    } finally {
      setLoading(false);
    }
  };

  if (!role) return null;

  // Group permissions by resource
  const groupedPermissions = permissions.reduce((acc, permission) => {
    if (!acc[permission.resource]) {
      acc[permission.resource] = [];
    }
    acc[permission.resource].push(permission);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Manage Permissions for {role.name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="text-sm text-muted-foreground">
            Select permissions to grant to this role
          </div>
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {Object.entries(groupedPermissions).map(([resource, perms]) => (
              <div key={resource} className="border rounded-lg p-4">
                <h4 className="font-semibold text-sm mb-3 capitalize">
                  {resource}
                </h4>
                <div className="grid grid-cols-2 gap-2">
                  {perms.map((permission) => (
                    <label
                      key={permission.id}
                      className="flex items-center p-2 hover:bg-muted/50 rounded cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={selectedPermissions.includes(permission.id)}
                        onChange={() => handleTogglePermission(permission.id)}
                        className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
                      />
                      <span className="ml-2 text-sm">{permission.name}</span>
                    </label>
                  ))}
                </div>
              </div>
            ))}
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={loading}>
              {loading ? 'Saving...' : 'Save Permissions'}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Invitations Tab Component
function InvitationsTab({ searchQuery }: { searchQuery: string }) {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery<{ invitations: Invitation[], count: number }>({
    queryKey: ['invitations'],
    queryFn: async () => {
      // invitationService.list() returns Promise<AxiosResponse> in the existing service
      // We need to unwrap the data
      const response = await invitationService.list();
      // Ensure we return object with invitations array. 
      // Existing service returns AxiosResponse.
      // If backend returns { invitations: [...] }, then response.data is that object.
      return response.data;
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => invitationService.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invitations'] });
    },
  });

  const filteredInvitations = data?.invitations?.filter(inv =>
    inv.email.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="mb-4">
          <h3 className="text-lg font-semibold">Pending Invitations</h3>
          <p className="text-sm text-muted-foreground">
            Manage outstanding invitations to join the tenant.
          </p>
        </div>

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading invitations...</div>
        ) : filteredInvitations.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No pending invitations found.
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Expires</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredInvitations.map((inv) => (
                <TableRow key={inv.id}>
                  <TableCell>{inv.email}</TableCell>
                  <TableCell>
                    <Badge variant={inv.status === 'pending' ? 'warning' : 'secondary'}>
                      {inv.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{formatDate(inv.expires_at)}</TableCell>
                  <TableCell>{formatDate(inv.created_at)}</TableCell>
                  <TableCell>
                    {inv.status === 'pending' && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          if (confirm(`Revoke invitation for ${inv.email}?`)) {
                            revokeMutation.mutate(inv.id);
                          }
                        }}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}

// Invite User Dialog
function InviteUserDialog({ open, onOpenChange, roles }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  roles: Role[];
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    email: '',
    role_id: '',
  });

  const inviteMutation = useMutation({
    mutationFn: (data: typeof formData) => invitationService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invitations'] });
      onOpenChange(false);
      setFormData({ email: '', role_id: '' });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    inviteMutation.mutate(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite User</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Email *</label>
            <Input
              required
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              placeholder="user@example.com"
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Role *</label>
            <select
              required
              value={formData.role_id}
              onChange={(e) => setFormData({ ...formData, role_id: e.target.value })}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            >
              <option value="">Select a role</option>
              {roles.map((role) => (
                <option key={role.id} value={role.id}>
                  {role.name}
                </option>
              ))}
            </select>
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={inviteMutation.isPending}>
              {inviteMutation.isPending ? 'Sending...' : 'Send Invitation'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
