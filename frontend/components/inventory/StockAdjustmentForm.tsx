'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { AlertCircle, CheckCircle } from 'lucide-react';

interface StockAdjustmentFormProps {
  productId: string;
  productName: string;
  currentStock: number;
  onSubmit?: (adjustment: StockAdjustment) => Promise<void>;
  onCancel?: () => void;
}

interface StockAdjustment {
  product_id: string;
  adjustment_type: 'increase' | 'decrease';
  quantity: number;
  reason: string;
  notes?: string;
}

export function StockAdjustmentForm({
  productId,
  productName,
  currentStock,
  onSubmit,
  onCancel,
}: StockAdjustmentFormProps) {
  const [adjustmentType, setAdjustmentType] = useState<'increase' | 'decrease'>('increase');
  const [quantity, setQuantity] = useState('');
  const [reason, setReason] = useState('');
  const [notes, setNotes] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const projectedStock = adjustmentType === 'increase' 
    ? currentStock + (parseInt(quantity) || 0)
    : currentStock - (parseInt(quantity) || 0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);

    if (!quantity || parseInt(quantity) <= 0) {
      setError('Please enter a valid quantity');
      return;
    }

    if (!reason) {
      setError('Please select a reason for adjustment');
      return;
    }

    if (adjustmentType === 'decrease' && projectedStock < 0) {
      setError('Adjustment would result in negative stock');
      return;
    }

    setIsLoading(true);
    try {
      const adjustment: StockAdjustment = {
        product_id: productId,
        adjustment_type: adjustmentType,
        quantity: parseInt(quantity),
        reason,
        notes: notes || undefined,
      };

      if (onSubmit) {
        await onSubmit(adjustment);
      }

      setSuccess(true);
      setQuantity('');
      setNotes('');
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to adjust stock');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Stock Adjustment</CardTitle>
        <CardDescription>Adjust inventory for {productName}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
              <AlertCircle className="w-4 h-4" />
              {error}
            </div>
          )}

          {success && (
            <div className="flex items-center gap-2 p-3 bg-green-50 border border-green-200 rounded text-green-700 text-sm">
              <CheckCircle className="w-4 h-4" />
              Stock adjusted successfully
            </div>
          )}

          <div className="space-y-2">
            <Label>Current Stock</Label>
            <div className="p-3 bg-gray-50 rounded border border-gray-200">
              <p className="text-lg font-semibold">{currentStock} units</p>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Adjustment Type</Label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="increase"
                  checked={adjustmentType === 'increase'}
                  onChange={(e) => setAdjustmentType(e.target.value as 'increase' | 'decrease')}
                  className="w-4 h-4"
                />
                <span className="text-sm">Increase Stock</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="decrease"
                  checked={adjustmentType === 'decrease'}
                  onChange={(e) => setAdjustmentType(e.target.value as 'increase' | 'decrease')}
                  className="w-4 h-4"
                />
                <span className="text-sm">Decrease Stock</span>
              </label>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="quantity">Quantity</Label>
            <Input
              id="quantity"
              type="number"
              min="1"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              placeholder="Enter quantity"
              disabled={isLoading}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="reason">Reason</Label>
            <select
              id="reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              disabled={isLoading}
              className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">Select a reason</option>
              <option value="physical_count">Physical Count Variance</option>
              <option value="damage">Damaged Goods</option>
              <option value="return">Customer Return</option>
              <option value="correction">Inventory Correction</option>
              <option value="waste">Waste/Spoilage</option>
              <option value="transfer">Transfer</option>
              <option value="other">Other</option>
            </select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes (Optional)</Label>
            <textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add any additional notes"
              disabled={isLoading}
              className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md focus:outline-none focus:ring-2 focus:ring-ring resize-none"
              rows={3}
            />
          </div>

          <div className="p-3 bg-blue-50 border border-blue-200 rounded">
            <p className="text-sm text-gray-600">Projected Stock After Adjustment</p>
            <p className={`text-lg font-semibold ${projectedStock < 0 ? 'text-red-600' : 'text-blue-600'}`}>
              {projectedStock} units
            </p>
          </div>

          <div className="flex gap-3 justify-end">
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
                Cancel
              </Button>
            )}
            <Button type="submit" disabled={isLoading || !quantity || !reason}>
              {isLoading ? 'Adjusting...' : 'Adjust Stock'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
