'use client';

import { Button } from '@/components/ui/button';
import { CheckCircle, Package, Truck, Home, XCircle } from 'lucide-react';

interface OrderActionsProps {
  status: string;
  onConfirm?: () => void;
  onProcess?: () => void;
  onShip?: () => void;
  onDeliver?: () => void;
  onCancel?: () => void;
  disabled?: boolean;
}

export function OrderActions({
  status,
  onConfirm,
  onProcess,
  onShip,
  onDeliver,
  onCancel,
  disabled = false,
}: OrderActionsProps) {
  return (
    <div className="flex items-center gap-2">
      {status === 'pending' && onConfirm && (
        <Button
          size="sm"
          onClick={onConfirm}
          disabled={disabled}
          className="bg-blue-600 hover:bg-blue-700"
        >
          <CheckCircle className="h-4 w-4 mr-1" />
          Confirm
        </Button>
      )}

      {status === 'confirmed' && onProcess && (
        <Button
          size="sm"
          onClick={onProcess}
          disabled={disabled}
          className="bg-purple-600 hover:bg-purple-700"
        >
          <Package className="h-4 w-4 mr-1" />
          Process
        </Button>
      )}

      {status === 'processing' && onShip && (
        <Button
          size="sm"
          onClick={onShip}
          disabled={disabled}
          className="bg-indigo-600 hover:bg-indigo-700"
        >
          <Truck className="h-4 w-4 mr-1" />
          Ship
        </Button>
      )}

      {status === 'shipped' && onDeliver && (
        <Button
          size="sm"
          onClick={onDeliver}
          disabled={disabled}
          className="bg-green-600 hover:bg-green-700"
        >
          <Home className="h-4 w-4 mr-1" />
          Deliver
        </Button>
      )}

      {['pending', 'confirmed'].includes(status) && onCancel && (
        <Button
          size="sm"
          variant="outline"
          onClick={onCancel}
          disabled={disabled}
          className="text-red-600 hover:text-red-700 hover:bg-red-50"
        >
          <XCircle className="h-4 w-4 mr-1" />
          Cancel
        </Button>
      )}
    </div>
  );
}
