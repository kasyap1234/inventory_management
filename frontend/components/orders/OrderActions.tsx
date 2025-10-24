'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { CheckCircle, Package, Truck, Home, XCircle } from 'lucide-react';
import api from '@/lib/api';
import { showSuccess, showError } from '@/lib/toast';

interface OrderActionsProps {
  orderId: string;
  status: string;
  onConfirm?: () => void;
  onProcess?: () => void;
  onShip?: () => void;
  onDeliver?: () => void;
  onCancel?: () => void;
  onSuccess?: () => void;
  disabled?: boolean;
}

export function OrderActions({
  orderId,
  status,
  onConfirm,
  onProcess,
  onShip,
  onDeliver,
  onCancel,
  onSuccess,
  disabled = false,
}: OrderActionsProps) {
  const [loading, setLoading] = useState(false);

  const handleAction = async (action: string, callback?: () => void) => {
    setLoading(true);
    try {
      await api.post(`/orders/${orderId}/${action}`);
      showSuccess(`Order ${action}ed successfully`);
      if (callback) callback();
      if (onSuccess) onSuccess();
    } catch (error: any) {
      showError(error.response?.data?.message || `Failed to ${action} order`);
    } finally {
      setLoading(false);
    }
  };

  const isDisabled = disabled || loading;

  return (
    <div className="flex items-center gap-2">
      {status === 'pending' && (
        <Button
          onClick={() => handleAction('approve', onConfirm)}
          disabled={isDisabled}
          className="bg-blue-600 hover:bg-blue-700 text-white"
        >
          <CheckCircle className="h-4 w-4 mr-1" />
          Approve
        </Button>
      )}

      {status === 'approved' && (
        <Button
          onClick={() => handleAction('process', onProcess)}
          disabled={isDisabled}
          className="bg-purple-600 hover:bg-purple-700 text-white"
        >
          <Package className="h-4 w-4 mr-1" />
          Process
        </Button>
      )}

      {status === 'processing' && (
        <Button
          onClick={() => handleAction('ship', onShip)}
          disabled={isDisabled}
          className="bg-indigo-600 hover:bg-indigo-700 text-white"
        >
          <Truck className="h-4 w-4 mr-1" />
          Ship
        </Button>
      )}

      {status === 'shipped' && (
        <Button
          onClick={() => handleAction('deliver', onDeliver)}
          disabled={isDisabled}
          className="bg-green-600 hover:bg-green-700 text-white"
        >
          <Home className="h-4 w-4 mr-1" />
          Deliver
        </Button>
      )}

      {['pending', 'approved', 'processing'].includes(status) && (
        <Button
          variant="outline"
          onClick={() => handleAction('cancel', onCancel)}
          disabled={isDisabled}
          className="text-red-600 hover:text-red-700 hover:bg-red-50"
        >
          <XCircle className="h-4 w-4 mr-1" />
          Cancel
        </Button>
      )}
    </div>
  );
}
