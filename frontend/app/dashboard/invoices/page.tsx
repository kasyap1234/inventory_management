'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Download, Eye, FileText, Trash } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import AdvancedFilters, { ActiveFilterBadges } from '@/components/filters/AdvancedFilters';
import api from '@/lib/api';
import { Invoice, Order } from '@/types';
import { formatCurrency, formatDate } from '@/lib/utils';

export default function InvoicesPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isBulkGenerateOpen, setIsBulkGenerateOpen] = useState(false);
  const [selectedInvoices, setSelectedInvoices] = useState<string[]>([]);
  const [advancedFilters, setAdvancedFilters] = useState<Record<string, any>>({});
  const queryClient = useQueryClient();

  const { data: invoices, isLoading } = useQuery<{ invoices: Invoice[] }>({
    queryKey: ['invoices'],
    queryFn: async () => {
      const response = await api.get('/invoices?limit=100');
      return response.data;
    },
  });

  const { data: orders } = useQuery<{ orders: Order[] }>({
    queryKey: ['orders'],
    queryFn: async () => {
      const response = await api.get('/orders?limit=100');
      return response.data;
    },
  });

  const generatePDF = useMutation({
    mutationFn: async (invoiceId: string) => {
      const response = await api.post(`/invoices/${invoiceId}/generate-pdf`, {}, {
        responseType: 'blob',
      });

      // Create blob and download
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `invoice-${invoiceId}.pdf`;
      link.click();
      window.URL.revokeObjectURL(url);
    },
  });

  const filteredInvoices = invoices?.invoices?.filter(invoice => {
    const matchesStatus = filterStatus === 'all' || invoice.status === filterStatus;

    const matchesSearch = !searchQuery
      || invoice.id.toLowerCase().includes(searchQuery.toLowerCase())
      || invoice.gstin?.toLowerCase().includes(searchQuery.toLowerCase());

    if (!matchesStatus || !matchesSearch) {
      return false;
    }

    if (advancedFilters.statuses && advancedFilters.statuses.length > 0) {
      if (!advancedFilters.statuses.includes(invoice.status)) {
        return false;
      }
    }

    if (advancedFilters.start_date) {
      if (!invoice.issued_date || invoice.issued_date < advancedFilters.start_date) {
        return false;
      }
    }

    if (advancedFilters.end_date) {
      if (!invoice.issued_date || invoice.issued_date > advancedFilters.end_date) {
        return false;
      }
    }

    if (advancedFilters.min_total) {
      if (invoice.total_amount < parseFloat(advancedFilters.min_total)) {
        return false;
      }
    }

    if (advancedFilters.max_total) {
      if (invoice.total_amount > parseFloat(advancedFilters.max_total)) {
        return false;
      }
    }

    if (advancedFilters.gstin && invoice.gstin) {
      if (!invoice.gstin.toLowerCase().includes(String(advancedFilters.gstin).toLowerCase())) {
        return false;
      }
    }

    return true;
  }) || [];

  const bulkDownloadPDFs = useMutation({
    mutationFn: async (invoiceIds: string[]) => {
      for (const id of invoiceIds) {
        const response = await api.post(`/invoices/${id}/generate-pdf`, {}, {
          responseType: 'blob',
        });
        const blob = new Blob([response.data], { type: 'application/pdf' });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `invoice-${id}.pdf`;
        link.click();
        window.URL.revokeObjectURL(url);
        // Small delay between downloads
        await new Promise(resolve => setTimeout(resolve, 500));
      }
    },
    onSuccess: () => {
      setSelectedInvoices([]);
    },
  });

  const bulkDeleteInvoices = useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map(id => api.delete(`/invoices/${id}`)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      setSelectedInvoices([]);
    },
  });

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedInvoices(filteredInvoices.map(i => i.id));
    } else {
      setSelectedInvoices([]);
    }
  };

  const handleSelectInvoice = (invoiceId: string, checked: boolean) => {
    if (checked) {
      setSelectedInvoices(prev => [...prev, invoiceId]);
    } else {
      setSelectedInvoices(prev => prev.filter(id => id !== invoiceId));
    }
  };

  const handleBulkDownload = () => {
    if (confirm(`Download ${selectedInvoices.length} PDF(s)?`)) {
      bulkDownloadPDFs.mutate(selectedInvoices);
    }
  };

  const handleBulkDelete = () => {
    if (confirm(`Are you sure you want to delete ${selectedInvoices.length} invoice(s)?`)) {
      bulkDeleteInvoices.mutate(selectedInvoices);
    }
  };

  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'paid': return 'success';
      case 'unpaid': return 'warning';
      case 'overdue': return 'destructive';
      case 'cancelled': return 'secondary';
      default: return 'default';
    }
  };

  const totalAmount = filteredInvoices.reduce((sum, inv) => sum + inv.total_amount, 0);
  const paidAmount = filteredInvoices.filter(inv => inv.status === 'paid').reduce((sum, inv) => sum + inv.total_amount, 0);
  const unpaidAmount = filteredInvoices.filter(inv => inv.status === 'unpaid').reduce((sum, inv) => sum + inv.total_amount, 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Invoices</h1>
          <p className="text-gray-500 mt-1">Manage billing and invoices</p>
          {selectedInvoices.length > 0 && (
            <p className="text-blue-600 text-sm mt-1">
              {selectedInvoices.length} invoice(s) selected
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {selectedInvoices.length > 0 && (
            <>
              <Button
                variant="outline"
                onClick={handleBulkDownload}
                disabled={bulkDownloadPDFs.isPending}
              >
                <Download className="h-4 w-4 mr-2" />
                {bulkDownloadPDFs.isPending ? 'Downloading...' : 'Download PDFs'}
              </Button>
              <Button
                variant="destructive"
                onClick={handleBulkDelete}
              >
                <Trash className="h-4 w-4 mr-2" />
                Delete Selected
              </Button>
            </>
          )}
          <Button variant="outline" onClick={() => setIsBulkGenerateOpen(true)}>
            <FileText className="h-4 w-4 mr-2" />
            Bulk Generate
          </Button>
          <Button onClick={() => setIsAddDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Invoice
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Total Invoices</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{invoices?.invoices?.length || 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Total Amount</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalAmount)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Paid</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{formatCurrency(paidAmount)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-600">Unpaid</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{formatCurrency(unpaidAmount)}</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search invoices..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="h-10 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
            >
              <option value="all">All Status</option>
              <option value="unpaid">Unpaid</option>
              <option value="paid">Paid</option>
              <option value="overdue">Overdue</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <AdvancedFilters
              config={{
                dateRange: {
                  label: 'Issued Date Range',
                  startKey: 'start_date',
                  endKey: 'end_date',
                },
                statuses: {
                  label: 'Invoice Status',
                  options: [
                    { value: 'unpaid', label: 'Unpaid' },
                    { value: 'paid', label: 'Paid' },
                    { value: 'overdue', label: 'Overdue' },
                    { value: 'cancelled', label: 'Cancelled' },
                  ],
                },
                priceRange: {
                  label: 'Total Amount',
                  minKey: 'min_total',
                  maxKey: 'max_total',
                },
                customFilters: [
                  {
                    key: 'gstin',
                    label: 'GSTIN Contains',
                    type: 'text',
                  },
                ],
              }}
              activeFilters={advancedFilters}
              onApply={setAdvancedFilters}
              onReset={() => setAdvancedFilters({})}
            />
          </div>
          <div className="mt-4">
            <ActiveFilterBadges
              filters={advancedFilters}
              onRemove={(key) => {
                const next = { ...advancedFilters };
                delete next[key];
                setAdvancedFilters(next);
              }}
            />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading invoices...</div>
          ) : filteredInvoices.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No invoices found. Create your first invoice to get started.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <input
                      type="checkbox"
                      checked={selectedInvoices.length === filteredInvoices.length && filteredInvoices.length > 0}
                      onChange={(e) => handleSelectAll(e.target.checked)}
                      className="h-4 w-4 rounded border-gray-300"
                    />
                  </TableHead>
                  <TableHead>Invoice ID</TableHead>
                  <TableHead>GSTIN</TableHead>
                  <TableHead>Total Amount</TableHead>
                  <TableHead>GST</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Issued Date</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredInvoices.map((invoice) => {
                  const gstTotal = (invoice.cgst || 0) + (invoice.sgst || 0) + (invoice.igst || 0);

                  return (
                    <TableRow key={invoice.id}>
                      <TableCell>
                        <input
                          type="checkbox"
                          checked={selectedInvoices.includes(invoice.id)}
                          onChange={(e) => handleSelectInvoice(invoice.id, e.target.checked)}
                          className="h-4 w-4 rounded border-gray-300"
                        />
                      </TableCell>
                      <TableCell className="font-mono text-sm">
                        {invoice.id.substring(0, 8)}...
                      </TableCell>
                      <TableCell>{invoice.gstin || '-'}</TableCell>
                      <TableCell className="font-semibold">
                        {formatCurrency(invoice.total_amount)}
                      </TableCell>
                      <TableCell>{formatCurrency(gstTotal)}</TableCell>
                      <TableCell>
                        <Badge variant={getStatusVariant(invoice.status)}>
                          {invoice.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatDate(invoice.issued_date)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            title="View Invoice"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => generatePDF.mutate(invoice.id)}
                            disabled={generatePDF.isPending}
                            title="Download PDF"
                          >
                            <Download className="h-4 w-4" />
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

      <InvoiceFormDialog
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        orders={orders?.orders || []}
      />

      <BulkInvoiceGenerateDialog
        open={isBulkGenerateOpen}
        onOpenChange={setIsBulkGenerateOpen}
        orders={orders?.orders || []}
      />
    </div>
  );
}

function InvoiceFormDialog({
  open,
  onOpenChange,
  orders,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orders: Order[];
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    order_id: '',
    gstin: '',
    hsn_sac: '',
    taxable_amount: 0,
    gst_rate: 18,
    cgst: 0,
    sgst: 0,
    igst: 0,
    total_amount: 0,
  });

  const calculateTax = () => {
    const taxable = formData.taxable_amount;
    const rate = formData.gst_rate / 100;
    const cgst = (taxable * rate) / 2;
    const sgst = (taxable * rate) / 2;
    const total = taxable + (taxable * rate);

    setFormData(prev => ({
      ...prev,
      cgst,
      sgst,
      total_amount: total,
    }));
  };

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      await api.post('/invoices', data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      onOpenChange(false);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate(formData);
  };

  const pendingOrders = orders.filter(o => o.status !== 'cancelled');

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Invoice</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Order *</label>
            <select
              required
              value={formData.order_id}
              onChange={(e) => setFormData({ ...formData, order_id: e.target.value })}
              className="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
            >
              <option value="">Select order</option>
              {pendingOrders.map((order) => (
                <option key={order.id} value={order.id}>
                  Order {order.id.substring(0, 8)}... - {formatCurrency(order.quantity * order.unit_price)}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">GSTIN</label>
              <Input
                value={formData.gstin}
                onChange={(e) => setFormData({ ...formData, gstin: e.target.value })}
                placeholder="22AAAAA0000A1Z5"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">HSN/SAC Code</label>
              <Input
                value={formData.hsn_sac}
                onChange={(e) => setFormData({ ...formData, hsn_sac: e.target.value })}
                placeholder="2301"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Taxable Amount *</label>
              <Input
                required
                type="number"
                step="0.01"
                value={formData.taxable_amount}
                onChange={(e) => setFormData({ ...formData, taxable_amount: parseFloat(e.target.value) })}
                onBlur={calculateTax}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">GST Rate (%) *</label>
              <Input
                required
                type="number"
                step="0.01"
                value={formData.gst_rate}
                onChange={(e) => setFormData({ ...formData, gst_rate: parseFloat(e.target.value) })}
                onBlur={calculateTax}
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">CGST</label>
              <Input
                type="number"
                step="0.01"
                value={formData.cgst}
                readOnly
                className="bg-gray-50"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">SGST</label>
              <Input
                type="number"
                step="0.01"
                value={formData.sgst}
                readOnly
                className="bg-gray-50"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">IGST</label>
              <Input
                type="number"
                step="0.01"
                value={formData.igst}
                onChange={(e) => setFormData({ ...formData, igst: parseFloat(e.target.value) })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Total Amount</label>
            <Input
              type="number"
              step="0.01"
              value={formData.total_amount}
              readOnly
              className="bg-gray-50 text-lg font-semibold"
            />
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Creating...' : 'Create Invoice'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function BulkInvoiceGenerateDialog({
  open,
  onOpenChange,
  orders,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orders: Order[];
}) {
  const queryClient = useQueryClient();
  const [selectedOrders, setSelectedOrders] = useState<string[]>([]);
  const [defaultGSTRate, setDefaultGSTRate] = useState(18);
  const [defaultGSTIN, setDefaultGSTIN] = useState('');
  const [defaultHSN, setDefaultHSN] = useState('');

  const bulkGenerateMutation = useMutation({
    mutationFn: async () => {
      const invoices = selectedOrders.map(orderId => {
        const order = orders.find(o => o.id === orderId);
        if (!order) return null;

        const taxableAmount = order.quantity * order.unit_price;
        const gstAmount = (taxableAmount * defaultGSTRate) / 100;

        return {
          order_id: orderId,
          gstin: defaultGSTIN,
          hsn_sac: defaultHSN,
          taxable_amount: taxableAmount,
          gst_rate: defaultGSTRate,
          cgst: gstAmount / 2,
          sgst: gstAmount / 2,
          igst: 0,
          total_amount: taxableAmount + gstAmount,
        };
      }).filter(Boolean);

      await api.post('/invoices/bulk-create', { invoices });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      onOpenChange(false);
      setSelectedOrders([]);
    },
  });

  const handleToggleOrder = (orderId: string) => {
    setSelectedOrders(prev =>
      prev.includes(orderId)
        ? prev.filter(id => id !== orderId)
        : [...prev, orderId]
    );
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedOrders(eligibleOrders.map(o => o.id));
    } else {
      setSelectedOrders([]);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedOrders.length === 0) {
      alert('Please select at least one order');
      return;
    }
    if (confirm(`Generate ${selectedOrders.length} invoice(s)?`)) {
      bulkGenerateMutation.mutate();
    }
  };

  // Filter orders that can have invoices generated (delivered orders without invoices)
  const eligibleOrders = orders.filter(o =>
    ['approved', 'delivered', 'shipped'].includes(o.status)
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Bulk Generate Invoices</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="bg-blue-50 p-3 rounded-lg">
            <p className="text-sm text-blue-800">
              Generate invoices for multiple orders at once with default GST settings
            </p>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Default GST Rate (%) *</label>
              <Input
                required
                type="number"
                step="0.01"
                value={defaultGSTRate}
                onChange={(e) => setDefaultGSTRate(parseFloat(e.target.value))}
                placeholder="18"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Default GSTIN</label>
              <Input
                value={defaultGSTIN}
                onChange={(e) => setDefaultGSTIN(e.target.value)}
                placeholder="22AAAAA0000A1Z5"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Default HSN/SAC</label>
              <Input
                value={defaultHSN}
                onChange={(e) => setDefaultHSN(e.target.value)}
                placeholder="2301"
              />
            </div>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium">Select Orders</label>
              <label className="flex items-center text-sm">
                <input
                  type="checkbox"
                  checked={selectedOrders.length === eligibleOrders.length && eligibleOrders.length > 0}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  className="mr-2 h-4 w-4 rounded border-gray-300"
                />
                Select All
              </label>
            </div>
            <div className="border border-gray-200 rounded-lg max-h-64 overflow-y-auto">
              {eligibleOrders.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  No eligible orders found. Orders must be approved, shipped, or delivered.
                </div>
              ) : (
                <div className="divide-y divide-gray-200">
                  {eligibleOrders.map((order) => {
                    const amount = order.quantity * order.unit_price;
                    const gstAmount = (amount * defaultGSTRate) / 100;
                    const total = amount + gstAmount;

                    return (
                      <label
                        key={order.id}
                        className="flex items-center p-3 hover:bg-gray-50 cursor-pointer"
                      >
                        <input
                          type="checkbox"
                          checked={selectedOrders.includes(order.id)}
                          onChange={() => handleToggleOrder(order.id)}
                          className="h-4 w-4 text-blue-600 rounded mr-3"
                        />
                        <div className="flex-1 grid grid-cols-4 gap-4 text-sm">
                          <div>
                            <div className="font-medium text-foreground">
                              Order #{order.id.substring(0, 8)}...
                            </div>
                            <div className="text-xs text-gray-500">
                              {order.order_type}
                            </div>
                          </div>
                          <div>
                            <div className="text-gray-600">Qty: {order.quantity}</div>
                            <div className="text-xs text-gray-500">@ {formatCurrency(order.unit_price)}</div>
                          </div>
                          <div>
                            <div className="text-gray-600">Taxable: {formatCurrency(amount)}</div>
                            <div className="text-xs text-gray-500">GST: {formatCurrency(gstAmount)}</div>
                          </div>
                          <div className="text-right">
                            <div className="font-semibold text-foreground">
                              {formatCurrency(total)}
                            </div>
                            <Badge variant={order.status === 'delivered' ? 'success' : 'warning'}>
                              {order.status}
                            </Badge>
                          </div>
                        </div>
                      </label>
                    );
                  })}
                </div>
              )}
            </div>
          </div>

          {selectedOrders.length > 0 && (
            <div className="bg-green-50 p-3 rounded-lg">
              <p className="text-sm text-green-800">
                <strong>{selectedOrders.length}</strong> invoice(s) will be generated
              </p>
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={bulkGenerateMutation.isPending || selectedOrders.length === 0}
            >
              {bulkGenerateMutation.isPending ? 'Generating...' : `Generate ${selectedOrders.length} Invoice(s)`}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
