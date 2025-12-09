'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { userService, invitationService, roleService, userManagementService } from '@/lib/services';
import { Loader2, Mail, UserPlus, Trash2, Pencil, Check, X, Shield, UserCheck } from 'lucide-react';
import { getErrorMessage } from '@/lib/toast';
import toast from 'react-hot-toast';
import { useAuth } from '@/hooks/useAuth';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';

type User = {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role?: string;
    status: string;
    tenant_id?: string;
};

type Invitation = {
    id: string;
    email: string;
    role_id: string;
    status: string;
    created_at: string;
};

type Role = {
    id: string;
    name: string;
    description?: string;
};

export default function UsersPage() {
    const [users, setUsers] = useState<User[]>([]);
    const [pendingUsers, setPendingUsers] = useState<User[]>([]);
    const [invitations, setInvitations] = useState<Invitation[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [roles, setRoles] = useState<Role[]>([]);
    const { user: currentUser } = useAuth();

    // Invite form state
    const [inviteEmail, setInviteEmail] = useState('');
    const [selectedRoleId, setSelectedRoleId] = useState('');

    // Create user form state
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [createFormData, setCreateFormData] = useState({
        email: '',
        first_name: '',
        last_name: '',
        role_id: '',
    });

    // Edit user state
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [editingUser, setEditingUser] = useState<User | null>(null);
    const [editFormData, setEditFormData] = useState({
        first_name: '',
        last_name: '',
        status: '',
    });

    // Delete confirmation state
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    const [userToDelete, setUserToDelete] = useState<User | null>(null);

    // Role assignment state
    const [showRoleDialog, setShowRoleDialog] = useState(false);
    const [userForRoles, setUserForRoles] = useState<User | null>(null);
    const [userRoles, setUserRoles] = useState<Role[]>([]);
    const [selectedRoleForAssign, setSelectedRoleForAssign] = useState('');

    const fetchData = async () => {
        try {
            const [usersRes, invitationsRes, rolesRes] = await Promise.all([
                userService.list(),
                invitationService.list(),
                roleService.list()
            ]);
            setUsers(usersRes.data.users || []);
            setInvitations(invitationsRes.data.invitations || []);
            const fetchedRoles: Role[] = rolesRes.data?.roles || [];
            setRoles(fetchedRoles);
            if (!selectedRoleId && fetchedRoles.length > 0) {
                setSelectedRoleId(fetchedRoles[0].id);
            }
            if (!createFormData.role_id && fetchedRoles.length > 0) {
                setCreateFormData(prev => ({ ...prev, role_id: fetchedRoles[0].id }));
            }

            // Fetch pending users separately
            try {
                const pendingRes = await userManagementService.getPendingUsers();
                setPendingUsers(pendingRes.data?.users || []);
            } catch {
                // User may not have permission to view pending users
                setPendingUsers([]);
            }
        } catch (error) {
            console.error('Failed to fetch data', error);
            toast.error('Failed to load users');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Invite user handler
    const handleInvite = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inviteEmail || !selectedRoleId) {
            toast.error('Please fill in all fields');
            return;
        }

        setIsSubmitting(true);
        try {
            await invitationService.create({ email: inviteEmail, role_id: selectedRoleId });
            toast.success(`Invitation sent to ${inviteEmail}`);
            setInviteEmail('');
            fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Create user handler
    const handleCreateUser = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!createFormData.email || !createFormData.first_name || !createFormData.last_name) {
            toast.error('Please fill in all required fields');
            return;
        }

        setIsSubmitting(true);
        try {
            await userService.create({
                email: createFormData.email,
                first_name: createFormData.first_name,
                last_name: createFormData.last_name,
                tenant_id: currentUser?.tenant_id,
                role_id: createFormData.role_id || undefined,
            });
            toast.success('User created successfully');
            setShowCreateDialog(false);
            setCreateFormData({ email: '', first_name: '', last_name: '', role_id: roles[0]?.id || '' });
            fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Edit user handlers
    const openEditDialog = (user: User) => {
        setEditingUser(user);
        setEditFormData({
            first_name: user.first_name,
            last_name: user.last_name,
            status: user.status,
        });
        setShowEditDialog(true);
    };

    const handleUpdateUser = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!editingUser) return;

        setIsSubmitting(true);
        try {
            await userService.update(editingUser.id, {
                first_name: editFormData.first_name,
                last_name: editFormData.last_name,
                status: editFormData.status,
            });
            toast.success('User updated successfully');
            setShowEditDialog(false);
            setEditingUser(null);
            fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Delete user handlers
    const openDeleteDialog = (user: User) => {
        setUserToDelete(user);
        setShowDeleteDialog(true);
    };

    const handleDeleteUser = async () => {
        if (!userToDelete) return;

        setIsSubmitting(true);
        try {
            await userService.delete(userToDelete.id);
            toast.success('User deleted successfully');
            setShowDeleteDialog(false);
            setUserToDelete(null);
            fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Approve user handler
    const handleApproveUser = async (userId: string) => {
        setIsSubmitting(true);
        try {
            await userManagementService.approveUser(userId);
            toast.success('User approved successfully');
            fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Role management handlers
    const openRoleDialog = async (user: User) => {
        setUserForRoles(user);
        setShowRoleDialog(true);
        try {
            const res = await roleService.getUserRoles(user.id);
            setUserRoles(res.data?.roles || []);
        } catch {
            setUserRoles([]);
        }
    };

    const handleAssignRole = async () => {
        if (!userForRoles || !selectedRoleForAssign) return;

        setIsSubmitting(true);
        try {
            const currentRoleIds = userRoles.map(r => r.id);
            await roleService.assignRolesToUser(userForRoles.id, [...currentRoleIds, selectedRoleForAssign]);
            toast.success('Role assigned successfully');
            const res = await roleService.getUserRoles(userForRoles.id);
            setUserRoles(res.data?.roles || []);
            setSelectedRoleForAssign('');
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleRemoveRole = async (roleId: string) => {
        if (!userForRoles) return;

        setIsSubmitting(true);
        try {
            await roleService.removeRoleFromUser(userForRoles.id, roleId);
            toast.success('Role removed successfully');
            const res = await roleService.getUserRoles(userForRoles.id);
            setUserRoles(res.data?.roles || []);
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsSubmitting(false);
        }
    };

    // Revoke invitation handler
    const handleRevoke = async (id: string) => {
        try {
            await invitationService.revoke(id);
            toast.success('Invitation revoked');
            fetchData();
        } catch {
            toast.error('Failed to revoke invitation');
        }
    };

    const getStatusBadge = (status: string) => {
        const statusStyles: Record<string, string> = {
            active: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
            inactive: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
            pending_approval: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
            pending_verification: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
        };
        return (
            <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusStyles[status] || statusStyles.inactive}`}>
                {status.replace('_', ' ')}
            </span>
        );
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    return (
        <div className="container mx-auto py-10 space-y-8">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold">User Management</h1>
                <Button onClick={() => setShowCreateDialog(true)}>
                    <UserPlus className="mr-2 h-4 w-4" />
                    Create User
                </Button>
            </div>

            <div className="grid gap-6 lg:grid-cols-3">
                {/* Invite User Form */}
                <Card>
                    <CardHeader>
                        <CardTitle>Invite User</CardTitle>
                        <CardDescription>
                            Send an invitation email to a new user.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleInvite} className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="email">Email Address</Label>
                                <div className="relative">
                                    <Mail className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="email"
                                        type="email"
                                        placeholder="colleague@company.com"
                                        className="pl-9"
                                        value={inviteEmail}
                                        onChange={(e) => setInviteEmail(e.target.value)}
                                        required
                                    />
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="role">Role</Label>
                                <select
                                    id="role"
                                    className="w-full rounded-md border px-3 py-2 text-sm"
                                    value={selectedRoleId}
                                    onChange={(e) => setSelectedRoleId(e.target.value)}
                                >
                                    {roles.length === 0 && <option value="">No roles available</option>}
                                    {roles.map((role) => (
                                        <option key={role.id} value={role.id}>
                                            {role.name}
                                        </option>
                                    ))}
                                </select>
                            </div>
                            <Button type="submit" className="w-full" disabled={isSubmitting}>
                                {isSubmitting ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Sending...
                                    </>
                                ) : (
                                    <>
                                        <Mail className="mr-2 h-4 w-4" />
                                        Send Invitation
                                    </>
                                )}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Stats Cards */}
                <Card>
                    <CardHeader>
                        <CardTitle>Total Users</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-4xl font-bold">{users.length}</div>
                        <p className="text-muted-foreground text-sm mt-2">
                            Active members in your organization
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Pending</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-4xl font-bold text-yellow-600">
                            {pendingUsers.length + invitations.filter(i => i.status === 'pending').length}
                        </div>
                        <p className="text-muted-foreground text-sm mt-2">
                            Invitations and approvals pending
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* Pending Users Section */}
            {pendingUsers.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <UserCheck className="h-5 w-5 text-yellow-600" />
                            Users Pending Approval
                        </CardTitle>
                        <CardDescription>
                            These users have registered and are waiting for admin approval.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Email</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {pendingUsers.map((user) => (
                                    <TableRow key={user.id}>
                                        <TableCell>{user.first_name} {user.last_name}</TableCell>
                                        <TableCell>{user.email}</TableCell>
                                        <TableCell>{getStatusBadge(user.status)}</TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                size="sm"
                                                onClick={() => handleApproveUser(user.id)}
                                                disabled={isSubmitting}
                                            >
                                                <Check className="mr-2 h-4 w-4" />
                                                Approve
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </CardContent>
                </Card>
            )}

            {/* Pending Invitations */}
            <Card>
                <CardHeader>
                    <CardTitle>Pending Invitations</CardTitle>
                </CardHeader>
                <CardContent>
                    {invitations.length === 0 ? (
                        <p className="text-muted-foreground text-sm">No pending invitations.</p>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Email</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Sent</TableHead>
                                    <TableHead className="text-right">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {invitations.map((invitation) => (
                                    <TableRow key={invitation.id}>
                                        <TableCell>{invitation.email}</TableCell>
                                        <TableCell className="capitalize">{invitation.status}</TableCell>
                                        <TableCell>{new Date(invitation.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => handleRevoke(invitation.id)}
                                                className="text-destructive hover:text-destructive/90"
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            {/* Active Users */}
            <Card>
                <CardHeader>
                    <CardTitle>All Users</CardTitle>
                    <CardDescription>Manage users in your organization</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Email</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Role</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {users.map((user) => (
                                <TableRow key={user.id}>
                                    <TableCell>{user.first_name} {user.last_name}</TableCell>
                                    <TableCell>{user.email}</TableCell>
                                    <TableCell>{getStatusBadge(user.status)}</TableCell>
                                    <TableCell className="capitalize">{user.role || '-'}</TableCell>
                                    <TableCell className="text-right space-x-1">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => openRoleDialog(user)}
                                            title="Manage Roles"
                                        >
                                            <Shield className="h-4 w-4" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => openEditDialog(user)}
                                            title="Edit User"
                                        >
                                            <Pencil className="h-4 w-4" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => openDeleteDialog(user)}
                                            className="text-destructive hover:text-destructive/90"
                                            title="Delete User"
                                            disabled={user.id === currentUser?.id}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            {/* Create User Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create New User</DialogTitle>
                        <DialogDescription>
                            Create a user account directly. The user will need to set their password later.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleCreateUser} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="create-email">Email</Label>
                            <Input
                                id="create-email"
                                type="email"
                                value={createFormData.email}
                                onChange={(e) => setCreateFormData({ ...createFormData, email: e.target.value })}
                                required
                            />
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="create-first-name">First Name</Label>
                                <Input
                                    id="create-first-name"
                                    value={createFormData.first_name}
                                    onChange={(e) => setCreateFormData({ ...createFormData, first_name: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="create-last-name">Last Name</Label>
                                <Input
                                    id="create-last-name"
                                    value={createFormData.last_name}
                                    onChange={(e) => setCreateFormData({ ...createFormData, last_name: e.target.value })}
                                    required
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="create-role">Role</Label>
                            <select
                                id="create-role"
                                className="w-full rounded-md border px-3 py-2 text-sm"
                                value={createFormData.role_id}
                                onChange={(e) => setCreateFormData({ ...createFormData, role_id: e.target.value })}
                            >
                                {roles.map((role) => (
                                    <option key={role.id} value={role.id}>
                                        {role.name}
                                    </option>
                                ))}
                            </select>
                        </div>
                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => setShowCreateDialog(false)}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Create User'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>

            {/* Edit User Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Edit User</DialogTitle>
                        <DialogDescription>
                            Update user information.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleUpdateUser} className="space-y-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="edit-first-name">First Name</Label>
                                <Input
                                    id="edit-first-name"
                                    value={editFormData.first_name}
                                    onChange={(e) => setEditFormData({ ...editFormData, first_name: e.target.value })}
                                    required
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="edit-last-name">Last Name</Label>
                                <Input
                                    id="edit-last-name"
                                    value={editFormData.last_name}
                                    onChange={(e) => setEditFormData({ ...editFormData, last_name: e.target.value })}
                                    required
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="edit-status">Status</Label>
                            <select
                                id="edit-status"
                                className="w-full rounded-md border px-3 py-2 text-sm"
                                value={editFormData.status}
                                onChange={(e) => setEditFormData({ ...editFormData, status: e.target.value })}
                            >
                                <option value="active">Active</option>
                                <option value="inactive">Inactive</option>
                            </select>
                        </div>
                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => setShowEditDialog(false)}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={isSubmitting}>
                                {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Save Changes'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>

            {/* Delete Confirmation Dialog */}
            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                            This will permanently delete the user <strong>{userToDelete?.email}</strong>.
                            This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleDeleteUser}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Delete'}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            {/* Role Management Dialog */}
            <Dialog open={showRoleDialog} onOpenChange={setShowRoleDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Manage Roles</DialogTitle>
                        <DialogDescription>
                            Manage roles for {userForRoles?.first_name} {userForRoles?.last_name}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        <div>
                            <Label className="text-sm font-medium">Current Roles</Label>
                            <div className="mt-2 space-y-2">
                                {userRoles.length === 0 ? (
                                    <p className="text-sm text-muted-foreground">No roles assigned</p>
                                ) : (
                                    userRoles.map((role) => (
                                        <div key={role.id} className="flex items-center justify-between bg-muted p-2 rounded-md">
                                            <span className="text-sm font-medium">{role.name}</span>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => handleRemoveRole(role.id)}
                                                disabled={isSubmitting}
                                            >
                                                <X className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                        <div className="border-t pt-4">
                            <Label className="text-sm font-medium">Add Role</Label>
                            <div className="flex gap-2 mt-2">
                                <select
                                    className="flex-1 rounded-md border px-3 py-2 text-sm"
                                    value={selectedRoleForAssign}
                                    onChange={(e) => setSelectedRoleForAssign(e.target.value)}
                                >
                                    <option value="">Select a role...</option>
                                    {roles
                                        .filter(r => !userRoles.find(ur => ur.id === r.id))
                                        .map((role) => (
                                            <option key={role.id} value={role.id}>
                                                {role.name}
                                            </option>
                                        ))}
                                </select>
                                <Button
                                    onClick={handleAssignRole}
                                    disabled={!selectedRoleForAssign || isSubmitting}
                                >
                                    Add
                                </Button>
                            </div>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowRoleDialog(false)}>
                            Done
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
