'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { superAdminService, tenantService } from '@/lib/services';
import { Loader2, Building2, Mail, RefreshCw } from 'lucide-react';
import { getErrorMessage } from '@/lib/toast';
import toast from 'react-hot-toast';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

interface Tenant {
    id: string;
    name: string;
    subdomain: string;
    status: string;
    license: string;
    created_at: string;
}

export default function AdminTenantsPage() {
    const [isInviting, setIsInviting] = useState(false);
    const [isLoadingTenants, setIsLoadingTenants] = useState(true);
    const [tenants, setTenants] = useState<Tenant[]>([]);
    const [formData, setFormData] = useState({
        email: '',
        tenantName: '',
    });

    const fetchTenants = async () => {
        setIsLoadingTenants(true);
        try {
            const response = await tenantService.list({ limit: 100 });
            // Backend returns { tenants: [], limit: ..., offset: ... }
            const data = response.data as { tenants: Tenant[] };
            setTenants(data.tenants || []);
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsLoadingTenants(false);
        }
    };

    useEffect(() => {
        fetchTenants();
    }, []);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData(prev => ({
            ...prev,
            [e.target.name]: e.target.value
        }));
    };

    const onSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsInviting(true);

        if (!formData.email || !formData.tenantName) {
            toast.error('Please fill in all fields');
            setIsInviting(false);
            return;
        }

        try {
            await superAdminService.inviteTenantAdmin(formData.email, formData.tenantName);
            toast.success(`Invitation sent to ${formData.email} for ${formData.tenantName}`);
            setFormData({ email: '', tenantName: '' });
            // Refresh list (though new tenant won't appear until admin accepts and tenant is created? 
            // Actually tenant is created upon invitation acceptance usually, or maybe before?
            // Checking tenant_service.go: CreateTenant creates tenant AND invites. 
            // But InviteTenantAdmin just creates an invitation. 
            // So the tenant might not exist yet in the tenants table until they accept?
            // Let's check invitationService.inviteTenantAdmin implementation...
            // It calls /admin/tenants/invite. 
            // Handler calls invitationService.InviteTenantAdmin.
            // This creates an invitation record. It does NOT create a tenant record yet.
            // The tenant is created when the invitation is accepted?
            // Wait, looking at `invitation_service.go` (not viewed yet but inferred), 
            // usually "Invite Tenant Admin" implies creating a tenant placeholder or just an invite.
            // If the tenant doesn't exist, it can't be listed.
            // However, the requirement is "view list of all tenants".
            // If I use "Create Tenant" endpoint, it creates tenant + invite.
            // The current UI uses "Invite Tenant Admin".
            // Let's stick to listing existing tenants for now.
        } catch (error: unknown) {
            toast.error(getErrorMessage(error));
        } finally {
            setIsInviting(false);
        }
    };

    return (
        <div className="container mx-auto py-10 space-y-8">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold">Tenant Management</h1>
                <Button variant="outline" onClick={fetchTenants} disabled={isLoadingTenants}>
                    <RefreshCw className={`mr-2 h-4 w-4 ${isLoadingTenants ? 'animate-spin' : ''}`} />
                    Refresh List
                </Button>
            </div>

            <div className="grid gap-6 md:grid-cols-1 lg:grid-cols-3">
                {/* Invite Form - Takes up 1/3 on large screens */}
                <Card className="lg:col-span-1 h-fit">
                    <CardHeader>
                        <CardTitle>Invite Tenant Admin</CardTitle>
                        <CardDescription>
                            Send an invitation to create a new tenant.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={onSubmit} className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="tenantName">Company Name</Label>
                                <div className="relative">
                                    <Building2 className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="tenantName"
                                        name="tenantName"
                                        placeholder="Acme Inc."
                                        className="pl-9"
                                        value={formData.tenantName}
                                        onChange={handleChange}
                                        required
                                    />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="email">Admin Email</Label>
                                <div className="relative">
                                    <Mail className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="email"
                                        name="email"
                                        type="email"
                                        placeholder="admin@acme.com"
                                        className="pl-9"
                                        value={formData.email}
                                        onChange={handleChange}
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
                                    'Send Invitation'
                                )}
                            </Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Tenant List - Takes up 2/3 on large screens */}
                <Card className="lg:col-span-2">
                    <CardHeader>
                        <CardTitle>Existing Tenants</CardTitle>
                        <CardDescription>List of all registered tenants in the system.</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {isLoadingTenants && tenants.length === 0 ? (
                            <div className="flex justify-center p-8">
                                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : tenants.length === 0 ? (
                            <div className="text-center p-8 text-muted-foreground">
                                No tenants found.
                            </div>
                        ) : (
                            <div className="rounded-md border">
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Name</TableHead>
                                            <TableHead>Subdomain</TableHead>
                                            <TableHead>Status</TableHead>
                                            <TableHead>License</TableHead>
                                            {/* <TableHead className="text-right">Actions</TableHead> */}
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {tenants.map((tenant) => (
                                            <TableRow key={tenant.id}>
                                                <TableCell className="font-medium">{tenant.name}</TableCell>
                                                <TableCell>{tenant.subdomain}</TableCell>
                                                <TableCell>
                                                    <Badge variant={tenant.status === 'active' ? 'default' : 'secondary'}>
                                                        {tenant.status}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell>{tenant.license || 'N/A'}</TableCell>
                                                {/* <TableCell className="text-right">
                                                    <Button variant="ghost" size="sm">Manage</Button>
                                                </TableCell> */}
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
