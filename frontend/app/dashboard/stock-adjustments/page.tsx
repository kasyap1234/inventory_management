'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Filter, TrendingUp, TrendingDown, RotateCcw, Package } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { toast } from 'react-hot-toast';
import api from '@/lib/api';
import { format } from 'date-fns';

interface StockAdjustment {
    id: string;
    product_id: string;
    warehouse_id?: string;
    adjustment_type: string;
    quantity: number;
    previous_stock: number;
    new_stock: number;
    reason?: string;
    reference_type?: string;
    reference_id?: string;
    adjusted_by: string;
    adjusted_at: string;
    created_at: string;
}

interface Product {
    id: string;
    name: string;
    quantity: number;
}

interface Warehouse {
    id: string;
    name: string;
}

const ADJUSTMENT_TYPES = [
    { value: 'increase', label: 'Increase', icon: TrendingUp, color: 'text-green-600' },
    { value: 'decrease', label: 'Decrease', icon: TrendingDown, color: 'text-red-600' },
    { value: 'correction', label: 'Correction', icon: RotateCcw, color: 'text-blue-600' },
    { value: 'damage', label: 'Damage', icon: TrendingDown, color: 'text-orange-600' },
    { value: 'return', label: 'Return', icon: TrendingUp, color: 'text-green-600' },
    { value: 'transfer_in', label: 'Transfer In', icon: TrendingUp, color: 'text-purple-600' },
    { value: 'transfer_out', label: 'Transfer Out', icon: TrendingDown, color: 'text-purple-600' },
];

export default function StockAdjustmentsPage() {
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const [filters, setFilters] = useState({
        product_id: '',
        warehouse_id: '',
        adjustment_type: '',
        start_date: '',
        end_date: '',
    });
    const queryClient = useQueryClient();

    const { data: adjustmentsData, isLoading } = useQuery({
        queryKey: ['stock-adjustments', filters],
        queryFn: async () => {
            const params = new URLSearchParams();
            Object.entries(filters).forEach(([key, value]) => {
                if (value) params.append(key, value);
            });
            const response = await api.get(`/stock-adjustments?${params.toString()}`);
            return response.data;
        },
    });

    const { data: productsData } = useQuery({
        queryKey: ['products'],
        queryFn: async () => {
            const response = await api.get('/products?limit=1000');
            return response.data;
        },
    });

    const { data: warehousesData } = useQuery({
        queryKey: ['warehouses'],
        queryFn: async () => {
            const response = await api.get('/warehouses?limit=100');
            return response.data;
        },
    });

    const adjustments: StockAdjustment[] = adjustmentsData?.adjustments || [];
    const products: Product[] = productsData?.products || [];
    const warehouses: Warehouse[] = warehousesData?.warehouses || [];

    const getAdjustmentTypeConfig = (type: string) => {
        return ADJUSTMENT_TYPES.find(t => t.value === type) || ADJUSTMENT_TYPES[0];
    };

    const getAdjustmentBadge = (adjustment: StockAdjustment) => {
        const config = getAdjustmentTypeConfig(adjustment.adjustment_type);
        const Icon = config.icon;
        return (
            <Badge variant="secondary" className="flex items-center gap-1 w-fit">
                <Icon className={`h-3 w-3 ${config.color}`} />
                <span className={config.color}>{config.label}</span>
            </Badge>
        );
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Stock Adjustments</h1>
                    <p className="text-muted-foreground mt-1">Track all inventory stock level changes</p>
                </div>
                <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Adjustment
                </Button>
            </div>

            {/* Filters */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Filter className="h-5 w-5" />
                        Filters
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
                        <div className="space-y-2">
                            <Label>Product</Label>
                            <select
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                value={filters.product_id}
                                onChange={(e) => setFilters({ ...filters, product_id: e.target.value })}
                            >
                                <option value="">All Products</option>
                                {products.map((product) => (
                                    <option key={product.id} value={product.id}>
                                        {product.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="space-y-2">
                            <Label>Warehouse</Label>
                            <select
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                value={filters.warehouse_id}
                                onChange={(e) => setFilters({ ...filters, warehouse_id: e.target.value })}
                            >
                                <option value="">All Warehouses</option>
                                {warehouses.map((warehouse) => (
                                    <option key={warehouse.id} value={warehouse.id}>
                                        {warehouse.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="space-y-2">
                            <Label>Type</Label>
                            <select
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                value={filters.adjustment_type}
                                onChange={(e) => setFilters({ ...filters, adjustment_type: e.target.value })}
                            >
                                <option value="">All Types</option>
                                {ADJUSTMENT_TYPES.map((type) => (
                                    <option key={type.value} value={type.value}>
                                        {type.label}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="space-y-2">
                            <Label>Start Date</Label>
                            <Input
                                type="date"
                                value={filters.start_date}
                                onChange={(e) => setFilters({ ...filters, start_date: e.target.value })}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label>End Date</Label>
                            <Input
                                type="date"
                                value={filters.end_date}
                                onChange={(e) => setFilters({ ...filters, end_date: e.target.value })}
                            />
                        </div>
                    </div>
                    {(filters.product_id || filters.adjustment_type || filters.start_date) && (
                        <Button
                            variant="outline"
                            className="mt-4"
                            onClick={() => setFilters({ product_id: '', warehouse_id: '', adjustment_type: '', start_date: '', end_date: '' })}
                        >
                            Clear Filters
                        </Button>
                    )}
                </CardContent>
            </Card>

            {/* Adjustments Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Adjustment History</CardTitle>
                    <CardDescription>
                        {adjustments.length} adjustment{adjustments.length !== 1 ? 's' : ''} found
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-12 text-muted-foreground">Loading adjustments...</div>
                    ) : adjustments.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            <Package className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
                            <p>No stock adjustments found</p>
                            <Button variant="link" onClick={() => setIsCreateDialogOpen(true)}>
                                Create the first adjustment
                            </Button>
                        </div>
                    ) : (
                        <div className="overflow-x-auto">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Date</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead>Quantity</TableHead>
                                        <TableHead>Previous Stock</TableHead>
                                        <TableHead>New Stock</TableHead>
                                        <TableHead>Reason</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {adjustments.map((adjustment) => (
                                        <TableRow key={adjustment.id}>
                                            <TableCell className="font-medium">
                                                {format(new Date(adjustment.adjusted_at), 'MMM d, yyyy HH:mm')}
                                            </TableCell>
                                            <TableCell>{getAdjustmentBadge(adjustment)}</TableCell>
                                            <TableCell>
                                                <span className={adjustment.quantity >= 0 ? 'text-green-600' : 'text-red-600'}>
                                                    {adjustment.quantity >= 0 ? '+' : ''}{adjustment.quantity}
                                                </span>
                                            </TableCell>
                                            <TableCell>{adjustment.previous_stock}</TableCell>
                                            <TableCell className="font-semibold">{adjustment.new_stock}</TableCell>
                                            <TableCell className="text-muted-foreground text-sm">
                                                {adjustment.reason || '—'}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    )}
                </CardContent>
            </Card>

            <CreateAdjustmentDialog
                open={isCreateDialogOpen}
                onOpenChange={setIsCreateDialogOpen}
                products={products}
                warehouses={warehouses}
                onSuccess={() => queryClient.invalidateQueries({ queryKey: ['stock-adjustments'] })}
            />
        </div>
    );
}

function CreateAdjustmentDialog({
    open,
    onOpenChange,
    products,
    warehouses,
    onSuccess,
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    products: Product[];
    warehouses: Warehouse[];
    onSuccess: () => void;
}) {
    const [formData, setFormData] = useState({
        product_id: '',
        warehouse_id: '',
        adjustment_type: 'increase',
        quantity: 0,
        reason: '',
    });

    const createMutation = useMutation({
        mutationFn: async (data: typeof formData) => {
            const response = await api.post('/stock-adjustments', data);
            return response.data;
        },
        onSuccess: () => {
            toast.success('Stock adjustment created successfully');
            onSuccess();
            onOpenChange(false);
            setFormData({
                product_id: '',
                warehouse_id: '',
                adjustment_type: 'increase',
                quantity: 0,
                reason: '',
            });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to create stock adjustment');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!formData.product_id) {
            toast.error('Please select a product');
            return;
        }
        if (formData.quantity === 0) {
            toast.error('Quantity must be non-zero');
            return;
        }
        createMutation.mutate(formData);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>Create Stock Adjustment</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label>Product *</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.product_id}
                            onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
                            required
                        >
                            <option value="">Select product...</option>
                            {products.map((product) => (
                                <option key={product.id} value={product.id}>
                                    {product.name} (Current: {product.quantity})
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-2">
                        <Label>Warehouse (Optional)</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.warehouse_id}
                            onChange={(e) => setFormData({ ...formData, warehouse_id: e.target.value })}
                        >
                            <option value="">General Stock</option>
                            {warehouses.map((warehouse) => (
                                <option key={warehouse.id} value={warehouse.id}>
                                    {warehouse.name}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-2">
                        <Label>Adjustment Type *</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.adjustment_type}
                            onChange={(e) => setFormData({ ...formData, adjustment_type: e.target.value })}
                            required
                        >
                            {ADJUSTMENT_TYPES.map((type) => (
                                <option key={type.value} value={type.value}>
                                    {type.label}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-2">
                        <Label>Quantity *</Label>
                        <Input
                            type="number"
                            value={formData.quantity}
                            onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) || 0 })}
                            placeholder="Enter quantity"
                            required
                        />
                        <p className="text-xs text-muted-foreground">
                            {formData.adjustment_type.includes('decrease') || formData.adjustment_type === 'damage'
                                ? 'Enter positive number to decrease stock'
                                : 'Enter positive number to increase stock'}
                        </p>
                    </div>

                    <div className="space-y-2">
                        <Label>Reason (Optional)</Label>
                        <Input
                            value={formData.reason}
                            onChange={(e) => setFormData({ ...formData, reason: e.target.value })}
                            placeholder="e.g., Received new shipment, Damaged goods, etc."
                        />
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createMutation.isPending}>
                            {createMutation.isPending ? 'Creating...' : 'Create Adjustment'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
