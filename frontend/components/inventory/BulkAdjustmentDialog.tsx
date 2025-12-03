import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { Inventory } from '@/types';
import { toast } from 'react-hot-toast';

interface BulkAdjustmentDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    selectedItems: Inventory[];
    onSuccess: () => void;
}

export function BulkAdjustmentDialog({
    open,
    onOpenChange,
    selectedItems,
    onSuccess,
}: BulkAdjustmentDialogProps) {
    const queryClient = useQueryClient();
    const [adjustments, setAdjustments] = useState<{ [key: string]: { adjustment: number; reason: string } }>({});

    const handleAdjustmentChange = (productId: string, field: 'adjustment' | 'reason', value: string | number) => {
        setAdjustments((prev) => ({
            ...prev,
            [productId]: {
                ...prev[productId],
                [field]: value,
            },
        }));
    };

    const bulkAdjustMutation = useMutation({
        mutationFn: async () => {
            const payload = {
                adjustments: selectedItems.map((item) => ({
                    product_id: item.product_id,
                    adjustment: adjustments[item.product_id]?.adjustment || 0,
                    reason: adjustments[item.product_id]?.reason || 'Bulk adjustment',
                })).filter(adj => adj.adjustment !== 0), // Only send non-zero adjustments
            };

            if (payload.adjustments.length === 0) {
                throw new Error("No adjustments entered");
            }

            await api.post('/inventory/bulk-adjust', payload);
        },
        onSuccess: () => {
            toast.success('Bulk adjustment successful');
            queryClient.invalidateQueries({ queryKey: ['inventory'] });
            onSuccess();
            onOpenChange(false);
            setAdjustments({});
        },
        onError: (error: any) => {
            toast.error(error.message || 'Failed to perform bulk adjustment');
        },
    });

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Bulk Stock Adjustment</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 py-4">
                    <div className="grid grid-cols-12 gap-4 font-medium text-sm text-muted-foreground mb-2">
                        <div className="col-span-4">Product</div>
                        <div className="col-span-2">Current</div>
                        <div className="col-span-2">Adjustment</div>
                        <div className="col-span-4">Reason</div>
                    </div>
                    {selectedItems.map((item) => (
                        <div key={item.id} className="grid grid-cols-12 gap-4 items-center">
                            <div className="col-span-4 text-sm truncate" title={item.product_name}>
                                {item.product_name}
                            </div>
                            <div className="col-span-2 text-sm">{item.quantity}</div>
                            <div className="col-span-2">
                                <Input
                                    type="number"
                                    placeholder="0"
                                    className="h-8"
                                    value={adjustments[item.product_id]?.adjustment || ''}
                                    onChange={(e) => handleAdjustmentChange(item.product_id, 'adjustment', parseInt(e.target.value) || 0)}
                                />
                            </div>
                            <div className="col-span-4">
                                <Input
                                    placeholder="Reason"
                                    className="h-8"
                                    value={adjustments[item.product_id]?.reason || ''}
                                    onChange={(e) => handleAdjustmentChange(item.product_id, 'reason', e.target.value)}
                                />
                            </div>
                        </div>
                    ))}
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        Cancel
                    </Button>
                    <Button
                        onClick={() => bulkAdjustMutation.mutate()}
                        disabled={bulkAdjustMutation.isPending}
                    >
                        {bulkAdjustMutation.isPending ? 'Saving...' : 'Save Adjustments'}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
