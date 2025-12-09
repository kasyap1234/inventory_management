'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Shield,
  Users,
  Key,
  Plus,
  Edit,
  Trash2,
  Check,
  X,
  Search,
  UserPlus,
  Settings,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import api from '@/lib/api';
import type { Role, Permission, User } from '@/types';

type TabType = 'roles' | 'users' | 'permissions';

export default function RBACManagementPage() {
  const [activeTab, setActiveTab] = useState<TabType>('roles');
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-foreground">RBAC Management</h1>
        <p className="text-gray-500 mt-1">Manage roles, permissions, and user access</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('roles')}
            className={`${
              activeTab === 'roles'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2`}
          >
            <Shield className="h-5 w-5" />
            Roles
          </button>
          <button
            onClick={() => setActiveTab('users')}
            className={`${
              activeTab === 'users'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2`}
          >
            <Users className="h-5 w-5" />
            User Roles
          </button>
          <button
            onClick={() => setActiveTab('permissions')}
            className={`${
              activeTab === 'permissions'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2`}
          >
            <Key className="h-5 w-5" />
            Permissions
          </button>
        </nav>
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder={`Search ${activeTab}...`}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'roles' && <RolesManagement searchQuery={searchQuery} />}
      {activeTab === 'users' && <UserRolesManagement searchQuery={searchQuery} />}
      {activeTab === 'permissions' && <PermissionsView searchQuery={searchQuery} />}
    </div>
  );
}

// Roles Management Component
function RolesManagement({ searchQuery }: { searchQuery: string }) {
  const queryClient = useQueryClient();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [managingPermissions, setManagingPermissions] = useState<Role | null>(null);

  const { data: roles, isLoading } = useQuery<{ roles: Role[] }>({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await api.get('/roles');
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
        <Button onClick={() => setIsCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Role
        </Button>
      </div>

      <Card>
        <CardContent className="pt-6">
          {isLoading ? (
            <div className="text-center py-8">Loading roles...</div>
          ) : filteredRoles.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No roles found</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Role Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRoles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium">{role.name}</TableCell>
                    <TableCell>{role.description || '—'}</TableCell>
                    <TableCell>
                      <Badge variant={role.is_active ? 'success' : 'secondary'}>
                        {role.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>{new Date(role.created_at).toLocaleDateString()}</TableCell>
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

      <RoleFormDialog
        open={isCreateDialogOpen || !!editingRole}
        onOpenChange={(open) => {
          setIsCreateDialogOpen(open);
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
      />
    </>
  );
}

// User Roles Management Component
function UserRolesManagement({ searchQuery }: { searchQuery: string }) {
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const { data: users, isLoading } = useQuery<{ users: User[] }>({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await api.get('/users');
      return response.data;
    },
  });

  const filteredUsers = users?.users?.filter(user =>
    user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.first_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    user.last_name.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>User Role Assignments</CardTitle>
          <CardDescription>Manage which roles are assigned to each user</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading users...</div>
          ) : filteredUsers.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No users found</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Status</TableHead>
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
                      <Badge variant={user.status === 'active' ? 'success' : 'secondary'}>
                        {user.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setSelectedUser(user)}
                      >
                        <UserPlus className="h-4 w-4 mr-1" />
                        Manage Roles
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ManageUserRolesDialog
        open={!!selectedUser}
        onOpenChange={(open) => {
          if (!open) setSelectedUser(null);
        }}
        user={selectedUser}
      />
    </>
  );
}

// Permissions View Component
function PermissionsView({ searchQuery }: { searchQuery: string }) {
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

  // Group by resource
  const groupedPermissions = filteredPermissions.reduce((acc, permission) => {
    if (!acc[permission.resource]) {
      acc[permission.resource] = [];
    }
    acc[permission.resource].push(permission);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <Card>
      <CardHeader>
        <CardTitle>System Permissions</CardTitle>
        <CardDescription>All available permissions in the system</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="text-center py-8">Loading permissions...</div>
        ) : filteredPermissions.length === 0 ? (
          <div className="text-center py-8 text-gray-500">No permissions found</div>
        ) : (
          <div className="space-y-6">
            {Object.entries(groupedPermissions).map(([resource, perms]) => (
              <div key={resource}>
                <h3 className="text-lg font-semibold text-foreground mb-3 capitalize">
                  {resource}
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                  {perms.map((permission) => (
                    <div
                      key={permission.id}
                      className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 transition-colors"
                    >
                      <div className="flex items-center justify-between mb-2">
                        <Badge variant="default">{permission.action}</Badge>
                        <Key className="h-4 w-4 text-gray-400" />
                      </div>
                      <p className="text-sm font-medium text-foreground">{permission.name}</p>
                      {permission.description && (
                        <p className="text-xs text-gray-500 mt-1">{permission.description}</p>
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
          <DialogTitle>{role ? 'Edit Role' : 'Create Role'}</DialogTitle>
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

// Manage Permissions Dialog
function ManagePermissionsDialog({ open, onOpenChange, role }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role: Role | null;
}) {
  const queryClient = useQueryClient();
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const { data: permissions } = useQuery<{ permissions: Permission[] }>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const response = await api.get('/permissions');
      return response.data;
    },
  });

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
  const groupedPermissions = (permissions?.permissions || []).reduce((acc, permission) => {
    if (!acc[permission.resource]) {
      acc[permission.resource] = [];
    }
    acc[permission.resource].push(permission);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Permissions for {role.name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="text-sm text-muted-foreground">
            Select permissions to grant to this role
          </div>
          <div className="space-y-4">
            {Object.entries(groupedPermissions).map(([resource, perms]) => (
              <div key={resource} className="border border-border rounded-lg p-4 bg-card/60">
                <h4 className="font-semibold text-sm text-foreground mb-3 capitalize">
                  {resource}
                </h4>
                <div className="grid grid-cols-2 gap-2">
                  {perms.map((permission) => (
                    <label
                      key={permission.id}
                      className="flex items-center p-2 rounded cursor-pointer hover-surface transition"
                    >
                      <input
                        type="checkbox"
                        checked={selectedPermissions.includes(permission.id)}
                        onChange={() => handleTogglePermission(permission.id)}
                        className="h-4 w-4 text-blue-600 rounded"
                      />
                      <div className="ml-2 text-sm">
                        <span className="font-medium">{permission.action}</span>
                        {permission.description && (
                          <span className="text-muted-foreground ml-1">({permission.description})</span>
                        )}
                      </div>
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

// Manage User Roles Dialog
function ManageUserRolesDialog({ open, onOpenChange, user }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user: User | null;
}) {
  const queryClient = useQueryClient();
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const { data: roles } = useQuery<{ roles: Role[] }>({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await api.get('/roles');
      return response.data;
    },
  });

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
            {roles?.roles?.map((role) => (
              <label
                key={role.id}
                className="flex items-center p-3 border border-border rounded-lg bg-card/60 hover-surface cursor-pointer transition"
              >
                <input
                  type="checkbox"
                  checked={selectedRoles.includes(role.id)}
                  onChange={() => handleToggleRole(role.id)}
                  className="h-4 w-4 text-blue-600 rounded"
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
