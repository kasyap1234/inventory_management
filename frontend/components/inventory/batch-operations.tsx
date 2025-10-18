'use client';

import { useState } from 'react';
import { QrCode, Package, Calendar, MapPin, Scan, Download, Upload } from 'lucide-react';
import toast from 'react-hot-toast';
import { QRCodeGenerator } from './qr-code-generator';
import { QRCodeScanner } from './qr-code-scanner';

export interface Batch {
  id: string;
  batch_number: string;
  product_id: string;
  product_name: string;
  quantity: number;
  manufacturing_date: string;
  expiry_date: string;
  warehouse_id: string;
  warehouse_name: string;
  status: 'active' | 'expired' | 'recalled';
}

interface BatchOperationsProps {
  batches: Batch[];
  onBatchUpdate: (batchId: string, updates: Partial<Batch>) => Promise<void>;
  onBatchCreate: (batch: Omit<Batch, 'id'>) => Promise<void>;
}

export function BatchOperations({ batches, onBatchUpdate, onBatchCreate }: BatchOperationsProps) {
  const [selectedBatch, setSelectedBatch] = useState<Batch | null>(null);
  const [showQRGenerator, setShowQRGenerator] = useState(false);
  const [showQRScanner, setShowQRScanner] = useState(false);
  const [scanMode, setScanMode] = useState<'receive' | 'ship' | 'verify'>('verify');

  const handleQRScan = async (data: string) => {
    try {
      const batchData = JSON.parse(data);
      
      if (scanMode === 'verify') {
        const batch = batches.find(b => b.batch_number === batchData.batch_number);
        if (batch) {
          setSelectedBatch(batch);
          toast.success(`Batch verified: ${batch.batch_number}`);
        } else {
          toast.error('Batch not found');
        }
      } else if (scanMode === 'receive') {
        toast.success(`Receiving batch: ${batchData.batch_number}`);
        // Handle receive logic
      } else if (scanMode === 'ship') {
        toast.success(`Shipping batch: ${batchData.batch_number}`);
        // Handle ship logic
      }
      
      setShowQRScanner(false);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Invalid QR code';
      toast.error(message);
    }
  };

  const generateBatchQRData = (batch: Batch) => {
    return JSON.stringify({
      batch_number: batch.batch_number,
      product_id: batch.product_id,
      product_name: batch.product_name,
      manufacturing_date: batch.manufacturing_date,
      expiry_date: batch.expiry_date,
      warehouse_id: batch.warehouse_id,
    });
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'expired':
        return 'bg-red-100 text-red-800';
      case 'recalled':
        return 'bg-orange-100 text-orange-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Actions */}
      <div className="flex flex-wrap gap-3">
        <button
          onClick={() => setShowQRScanner(true)}
          className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          <Scan className="h-5 w-5" />
          Scan QR Code
        </button>

        <div className="flex gap-2 border border-gray-300 rounded-lg overflow-hidden">
          <button
            onClick={() => setScanMode('verify')}
            className={`px-4 py-2 text-sm transition-colors ${
              scanMode === 'verify' ? 'bg-green-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Verify
          </button>
          <button
            onClick={() => setScanMode('receive')}
            className={`px-4 py-2 text-sm transition-colors ${
              scanMode === 'receive' ? 'bg-green-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Receive
          </button>
          <button
            onClick={() => setScanMode('ship')}
            className={`px-4 py-2 text-sm transition-colors ${
              scanMode === 'ship' ? 'bg-green-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Ship
          </button>
        </div>
      </div>

      {/* Batch List */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {batches.map((batch) => (
          <div
            key={batch.id}
            className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-lg transition-shadow"
          >
            <div className="flex justify-between items-start mb-3">
              <div className="flex-1">
                <h3 className="font-semibold text-gray-900">{batch.product_name}</h3>
                <p className="text-sm text-gray-600">Batch: {batch.batch_number}</p>
              </div>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(batch.status)}`}>
                {batch.status}
              </span>
            </div>

            <div className="space-y-2 mb-4">
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <Package className="h-4 w-4" />
                <span>Qty: {batch.quantity}</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <Calendar className="h-4 w-4" />
                <span>Exp: {new Date(batch.expiry_date).toLocaleDateString()}</span>
              </div>
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <MapPin className="h-4 w-4" />
                <span>{batch.warehouse_name}</span>
              </div>
            </div>

            <button
              onClick={() => {
                setSelectedBatch(batch);
                setShowQRGenerator(true);
              }}
              className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
            >
              <QrCode className="h-4 w-4" />
              Generate QR
            </button>
          </div>
        ))}
      </div>

      {/* QR Code Generator Modal */}
      {showQRGenerator && selectedBatch && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold">Batch QR Code</h3>
              <button
                onClick={() => setShowQRGenerator(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            </div>
            
            <QRCodeGenerator
              data={generateBatchQRData(selectedBatch)}
              filename={`batch_${selectedBatch.batch_number}`}
              title={selectedBatch.product_name}
              subtitle={`Batch: ${selectedBatch.batch_number}`}
            />
          </div>
        </div>
      )}

      {/* QR Code Scanner Modal */}
      {showQRScanner && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold">Scan Batch QR Code</h3>
              <button
                onClick={() => setShowQRScanner(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            </div>
            
            <QRCodeScanner
              onScan={handleQRScan}
              onError={(error) => toast.error(error)}
            />
          </div>
        </div>
      )}
    </div>
  );
}
