'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Calendar, Package2, CheckCircle, XCircle, Clock, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { toast } from 'react-hot-toast';
import api from '@/lib/api';
import { formatDate } from '@/lib/utils';

interface Reservation {
    id: string;
    product_id: string;
    warehouse_id?: string;
    reservation_id: string;
    quantity: number;
    reserved_by: string;
    reserved_at: string;
    expires_at?: string;
    status: string; // active, expired, released, committed
    order_id?: string;
    notes?: string;
}

const STATUS_ICONS = {
    active: Clock,
    expired: XCircle,
    released: CheckCircle,
    committed: CheckCircle,
};

const STATUS_COLORS = {
    active: 'bg-blue-100 text-blue-800',
    expired: 'bg-red-100 text-red-800',
    released: 'bg-gray-100 text-gray-800',
    committed: 'bg-green-100 text-green-800',
};

export default function ReservationsPage() {
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const [statusFilter, setStatusFilter] = useState('active');
    const queryClient = useQueryClient();

    const { data: reservationsData, isLoading } = useQuery({
        queryKey: ['reservations', statusFilter],
        queryFn: async () => {
            const response = await api.get(`/inventory-reservations?status=${statusFilter}`);
            return response.data;
        },
    });

    const updateStatusMutation = useMutation({
        mutationFn: async ({ id, status }: { id: string; status: string }) => {
            await api.put(`/inventory-reservations/${id}/status`, { status });
        },
        onSuccess: () => {
            toast.success('Reservation status updated');
            queryClient.invalidateQueries({ queryKey: ['reservations'] });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to update status');
        },
    });

    const deleteMutation = useMutation({
        mutationFn: async (id: string) => {
            await api.delete(`/inventory-reservations/${id}`);
        },
        onSuccess: () => {
            toast.success('Reservation deleted');
            queryClient.invalidateQueries({ queryKey: ['reservations'] });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to delete reservation');
        },
    });

    const reservations: Reservation[] = reservationsData?.reservations || [];

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Inventory Reservations</h1>
                    <p className="text-muted-foreground mt-1">Manage inventory reservations for orders and transfers</p>
                </div>
                <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    Create Reservation
                </Button>
            </div>

            {/* Status Filter */}
            <Card>
                <CardHeader>
                    <CardTitle>Filter by Status</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex gap-2">
                        {['active', 'expired', 'released', 'committed'].map((status) => (
                            <Button
                                key={status}
                                variant={statusFilter === status ? 'default' : 'outline'}
                                size="sm"
                                onClick={() => setStatusFilter(status)}
                            >
                                {status.charAt(0).toUpperCase() + status.slice(1)}
                            </Button>
                        ))}
                    </div>
                </CardContent>
            </Card>

            {/* Reservations Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Reservations</CardTitle>
                    <CardDescription>{reservations.length} reservation{reservations.length !== 1 ? 's' : ''} found</CardDescription>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">Loading reservations...</div>
                    ) : reservations.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            <Package2 className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
                            <p>No {statusFilter} reservations found</p>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Reservation ID</TableHead>
                                    <TableHead>Quantity</TableHead>
                                    <TableHead>Reserved At</TableHead>
                                    <TableHead>Expires At</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Notes</TableHead>
                                    <TableHead>Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {reservations.map((reservation) => {
                                    const StatusIcon = STATUS_ICONS[reservation.status as keyof typeof STATUS_ICONS] || Clock;
                                    const statusColor = STATUS_COLORS[reservation.status as keyof typeof STATUS_COLORS] || 'bg-gray-100 text-gray-800';

                                    return (
                                        <TableRow key={reservation.id}>
                                            <TableCell className="font-medium">{reservation.reservation_id}</TableCell>
                                            <TableCell>
                                                <span className="font-semibold">{reservation.quantity}</span> units
                                            </TableCell>
                                            <TableCell>{formatDate(reservation.reserved_at)}</TableCell>
                                            <TableCell>
                                                {reservation.expires_at ? formatDate(reservation.expires_at) : '—'}
                                            </TableCell>
                                            <TableCell>
                                                <Badge className={statusColor}>
                                                    <StatusIcon className="h-3 w-3 mr-1" />
                                                    {reservation.status}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="max-w-xs truncate">{reservation.notes || '—'}</TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    {reservation.status === 'active' && (
                                                        <>
                                                            <Button
                                                                variant="outline"
                                                                size="sm"
                                                                onClick={() => updateStatusMutation.mutate({ id: reservation.id, status: 'committed' })}
                                                            >
                                                                Commit
                                                            </Button>
                                                            <Button
                                                                variant="outline"
                                                                size="sm"
                                                                onClick={() => updateStatusMutation.mutate({ id: reservation.id, status: 'released' })}
                                                            >
                                                                Release
                                                            </Button>
                                                        </>
                                                    )}
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => {
                                                            if (confirm('Delete this reservation?')) {
                                                                deleteMutation.mutate(reservation.id);
                                                            }
                                                        }}
                                                    >
                                                        <Trash2 className="h-4 w-4 text-destructive" />
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

            <CreateReservationDialog
                open={isCreateDialogOpen}
                onOpenChange={setIsCreateDialogOpen}
            />
        </div>
    );
}

function CreateReservationDialog({ open, onOpenChange }: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
}) {
    const queryClient = useQueryClient();
    const [formData, setFormData] = useState({
        product_id: '',
        warehouse_id: '',
        reservation_id: '',
        quantity: 0,
        expires_at: '',
        notes: '',
    });

    // Fetch products for dropdown
    const { data: productsData } = useQuery({
        queryKey: ['products'],
        queryFn: async () => {
            const response = await api.get('/products?limit=100');
            return response.data;
        },
        enabled: open,
    });

    const createMutation = useMutation({
        mutationFn: async (data: typeof formData) => {
            const payload = {
                ...data,
                product_id: data.product_id || undefined,
                warehouse_id: data.warehouse_id || undefined,
                expires_at: data.expires_at || undefined,
            };
            await api.post('/inventory-reservations', payload);
        },
        onSuccess: () => {
            toast.success('Reservation created successfully');
            queryClient.invalidateQueries({ queryKey: ['reservations'] });
            onOpenChange(false);
            setFormData({
                product_id: '',
                warehouse_id: '',
                reservation_id: '',
                quantity: 0,
                expires_at: '',
                notes: '',
            });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to create reservation');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        createMutation.mutate(formData);
    };

    const products = productsData?.products || [];

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>Create Reservation</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label>Reservation ID *</Label>
                        <Input
                            required
                            value={formData.reservation_id}
                            onChange={(e) => setFormData({ ...formData, reservation_id: e.target.value })}
                            placeholder="RES-001"
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>Product *</Label>
                        <select
                            required
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.product_id}
                            onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
                        >
                            <option value="">Select a product</option>
                            {products.map((product: { id: string; name: string }) => (
                                <option key={product.id} value={product.id}>
                                    {product.name}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-2">
                        <Label>Quantity *</Label>
                        <Input
                            required
                            type="number"
                            min="1"
                            value={formData.quantity || ''}
                            onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) || 0 })}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>Expires At (Optional)</Label>
                        <Input
                            type="datetime-local"
                            value={formData.expires_at}
                            onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>Notes</Label>
                        <textarea
                            className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.notes}
                            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                            placeholder="Optional notes about this reservation"
                        />
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createMutation.isPending}>
                            {createMutation.isPending ? 'Creating...' : 'Create Reservation'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
