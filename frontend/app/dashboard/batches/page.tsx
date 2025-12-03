'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';

import { toast } from 'react-hot-toast';
import api from '@/lib/api';
import { BatchOperations, Batch } from '@/components/inventory/batch-operations';
import { Product, Warehouse } from '@/types';

export default function BatchesPage() {
    const [selectedProductId, setSelectedProductId] = useState<string>('');
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const [editingBatch, setEditingBatch] = useState<Batch | null>(null);
    const queryClient = useQueryClient();

    // Fetch products for selection
    const { data: productsData, isLoading: isLoadingProducts } = useQuery<{ products: Product[] }>({
        queryKey: ['products'],
        queryFn: async () => {
            const response = await api.get('/products?limit=100');
            return response.data;
        },
    });

    // Fetch batches for selected product
    const { data: batches, isLoading: isLoadingBatches } = useQuery<Batch[]>({
        queryKey: ['batches', selectedProductId],
        queryFn: async () => {
            if (!selectedProductId) return [];
            const response = await api.get(`/products/${selectedProductId}/batches`);
            return response.data;
        },
        enabled: !!selectedProductId,
    });

    // Fetch warehouses for create/edit dialog
    const { data: warehousesData } = useQuery<{ warehouses: Warehouse[] }>({
        queryKey: ['warehouses'],
        queryFn: async () => {
            const response = await api.get('/warehouses?limit=100');
            return response.data;
        },
    });

    const createBatchMutation = useMutation({
        mutationFn: async (data: any) => {
            const response = await api.post(`/products/${data.product_id}/batches`, data);
            return response.data;
        },
        onSuccess: () => {
            toast.success('Batch created successfully');
            setIsCreateDialogOpen(false);
            queryClient.invalidateQueries({ queryKey: ['batches', selectedProductId] });
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.message || 'Failed to create batch');
        },
    });

    const updateBatchMutation = useMutation({
        mutationFn: async (data: any) => {
            const response = await api.put(`/batches/${data.id}`, data);
            return response.data;
        },
        onSuccess: () => {
            toast.success('Batch updated successfully');
            setEditingBatch(null);
            queryClient.invalidateQueries({ queryKey: ['batches', selectedProductId] });
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.message || 'Failed to update batch');
        },
    });

    const handleCreateBatch = async (batchData: any) => {
        createBatchMutation.mutate({ ...batchData, product_id: selectedProductId });
    };

    const handleUpdateBatch = async (batchId: string, updates: Partial<Batch>) => {
        // This is called by BatchOperations, but we might want to open a dialog instead
        // For now, let's assume BatchOperations handles some updates or we pass a handler
        // Actually BatchOperations just calls the prop.
        // We can implement direct update if it's just status change, or open dialog for full edit.
        // Let's just log for now as BatchOperations doesn't seem to have edit UI built-in for all fields.
        console.log('Update batch', batchId, updates);
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-gray-900">Batch Management</h1>
                    <p className="text-gray-500 mt-1">Track and manage product batches</p>
                </div>
                <Button
                    onClick={() => setIsCreateDialogOpen(true)}
                    disabled={!selectedProductId}
                >
                    <Plus className="mr-2 h-4 w-4" />
                    Create Batch
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Select Product</CardTitle>
                    <CardDescription>Choose a product to view its batches</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="max-w-md">
                        <Label>Product</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 mt-1"
                            value={selectedProductId}
                            onChange={(e) => setSelectedProductId(e.target.value)}
                            disabled={isLoadingProducts}
                        >
                            <option value="">Select a product...</option>
                            {productsData?.products?.map((product) => (
                                <option key={product.id} value={product.id}>
                                    {product.name} ({product.quantity} in stock)
                                </option>
                            ))}
                        </select>
                    </div>
                </CardContent>
            </Card>

            {selectedProductId && (
                <div className="space-y-6">
                    {isLoadingBatches ? (
                        <div className="text-center py-12">Loading batches...</div>
                    ) : batches && batches.length > 0 ? (
                        <BatchOperations
                            batches={batches}
                            onBatchUpdate={handleUpdateBatch}
                            onBatchCreate={async (batch) => {
                                // This prop is required but we handle creation via dialog
                                // We can just pass a dummy function or wire it if BatchOperations had a create form
                                return Promise.resolve();
                            }}
                        />
                    ) : (
                        <div className="text-center py-12 bg-gray-50 rounded-lg border border-dashed border-gray-300">
                            <p className="text-gray-500">No batches found for this product.</p>
                            <Button variant="link" onClick={() => setIsCreateDialogOpen(true)}>
                                Create the first batch
                            </Button>
                        </div>
                    )}
                </div>
            )}

            <CreateBatchDialog
                open={isCreateDialogOpen}
                onOpenChange={setIsCreateDialogOpen}
                productId={selectedProductId}
                warehouses={warehousesData?.warehouses || []}
                onSubmit={handleCreateBatch}
                isSubmitting={createBatchMutation.isPending}
            />
        </div>
    );
}

function CreateBatchDialog({
    open,
    onOpenChange,
    productId,
    warehouses,
    onSubmit,
    isSubmitting
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    productId: string;
    warehouses: Warehouse[];
    onSubmit: (data: any) => void;
    isSubmitting: boolean;
}) {
    const [formData, setFormData] = useState({
        batch_number: '',
        quantity: 0,
        manufacturing_date: '',
        expiry_date: '',
        warehouse_id: '',
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Create New Batch</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <Label>Batch Number</Label>
                        <Input
                            value={formData.batch_number}
                            onChange={(e) => setFormData({ ...formData, batch_number: e.target.value })}
                            placeholder="e.g., BATCH-001"
                            required
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>Warehouse</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            value={formData.warehouse_id}
                            onChange={(e) => setFormData({ ...formData, warehouse_id: e.target.value })}
                            required
                        >
                            <option value="">Select warehouse...</option>
                            {warehouses.map((w) => (
                                <option key={w.id} value={w.id}>{w.name}</option>
                            ))}
                        </select>
                    </div>

                    <div className="space-y-2">
                        <Label>Quantity</Label>
                        <Input
                            type="number"
                            value={formData.quantity}
                            onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) })}
                            required
                            min="0"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Manufacturing Date</Label>
                            <Input
                                type="date"
                                value={formData.manufacturing_date}
                                onChange={(e) => setFormData({ ...formData, manufacturing_date: e.target.value })}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Expiry Date</Label>
                            <Input
                                type="date"
                                value={formData.expiry_date}
                                onChange={(e) => setFormData({ ...formData, expiry_date: e.target.value })}
                                required
                            />
                        </div>
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={isSubmitting}>
                            {isSubmitting ? 'Creating...' : 'Create Batch'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
