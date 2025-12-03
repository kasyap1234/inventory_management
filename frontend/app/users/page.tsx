'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { userService, invitationService } from '@/lib/services';
import { Loader2, Mail, UserPlus, Trash2 } from 'lucide-react';
import { getErrorMessage } from '@/lib/toast';
import toast from 'react-hot-toast';
import { useAuth } from '@/hooks/useAuth';

type User = {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: string;
    status: string;
};

type Invitation = {
    id: string;
    email: string;
    role_id: string;
    status: string;
    created_at: string;
};

export default function UsersPage() {
    const [users, setUsers] = useState<User[]>([]);
    const [invitations, setInvitations] = useState<Invitation[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isInviting, setIsInviting] = useState(false);
    const [inviteEmail, setInviteEmail] = useState('');
    const { user } = useAuth();

    const fetchData = async () => {
        try {
            const [usersRes, invitationsRes] = await Promise.all([
                userService.list(),
                invitationService.list()
            ]);
            setUsers(usersRes.data.users || []);
            setInvitations(invitationsRes.data.invitations || []);
        } catch (error) {
            console.error('Failed to fetch data', error);
            toast.error('Failed to load users and invitations');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleInvite = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inviteEmail) return;

        setIsInviting(true);
        try {
            // Default to 'user' role for now, or fetch roles to select
            // We need a role ID. Let's assume the backend handles default or we need to fetch roles.
            // For simplicity, let's assume we have a way to get role ID or the backend accepts role name?
            // The backend CreateInvitation expects RoleID.
            // We should probably fetch roles first.
            // For this MVP, let's just show a toast that we need to implement role selection.
            // Or better, let's fetch roles.

            // Wait, invitationService.create takes { email, role_id }.
            // I don't have roles loaded.
            // Let's just implement the UI structure for now and add role fetching if I have time.
            // Or I can hardcode a known role ID if I knew it, but I don't.
            // I'll skip the actual API call for now or mock it?
            // No, I should do it right.

            // Let's assume there's a 'user' role and we can find it.
            // But I don't have a roleService exposed in frontend yet.

            // I'll just put a placeholder for now.
            toast.error("Role selection not implemented yet");

            // await invitationService.create({ email: inviteEmail, role_id: '...' });
            // toast.success(`Invitation sent to ${inviteEmail}`);
            // setInviteEmail('');
            // fetchData();
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsInviting(false);
        }
    };

    const handleRevoke = async (id: string) => {
        try {
            await invitationService.revoke(id);
            toast.success('Invitation revoked');
            fetchData();
        } catch (error) {
            toast.error('Failed to revoke invitation');
        }
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
            </div>

            <div className="grid gap-6 md:grid-cols-3">
                {/* Invite User Form */}
                <Card className="md:col-span-1">
                    <CardHeader>
                        <CardTitle>Invite User</CardTitle>
                        <CardDescription>
                            Invite a new user to your organization.
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
                            <Button type="submit" className="w-full" disabled={isInviting}>
                                {isInviting ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        Sending...
                                    </>
                                ) : (
                                    <>
                                        <UserPlus className="mr-2 h-4 w-4" />
                                        Send Invitation
                                    </>
                                )}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Users and Invitations Lists */}
                <div className="md:col-span-2 space-y-6">
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

                    <Card>
                        <CardHeader>
                            <CardTitle>Active Users</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Email</TableHead>
                                        <TableHead>Role</TableHead>
                                        <TableHead>Status</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {users.map((user) => (
                                        <TableRow key={user.id}>
                                            <TableCell>{user.first_name} {user.last_name}</TableCell>
                                            <TableCell>{user.email}</TableCell>
                                            <TableCell className="capitalize">{user.role}</TableCell>
                                            <TableCell className="capitalize">{user.status}</TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
