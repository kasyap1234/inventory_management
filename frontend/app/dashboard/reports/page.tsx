'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { FileText, Download, Calendar, Package, TrendingUp, List, BarChart3, Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import api from '@/lib/api';
import { formatCurrency, formatDate } from '@/lib/utils';
import { toast } from 'react-hot-toast';

type ReportType = 'inventory' | 'sales' | 'stock-movements' | 'low-stock';

export default function ReportsPage() {
    const [activeReport, setActiveReport] = useState<ReportType>('inventory');
    const [dateRange, setDateRange] = useState({
        start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        end: new Date().toISOString().split('T')[0],
    });
    const [exporting, setExporting] = useState<string | null>(null);

    const handleExport = async (format: 'csv' | 'pdf' | 'excel', reportType: ReportType) => {
        setExporting(`${reportType}-${format}`);
        try {
            const response = await api.post(`/reports/${reportType}/export`, {
                format,
                start_date: dateRange.start,
                end_date: dateRange.end,
            }, {
                responseType: 'blob',
            });

            // Create download link
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            const extension = format === 'excel' ? 'xlsx' : format;
            link.setAttribute('download', `${reportType}-report-${new Date().toISOString().split('T')[0]}.${extension}`);
            document.body.appendChild(link);
            link.click();
            link.parentNode?.removeChild(link);

            toast.success(`${reportType} report exported successfully`);
        } catch (error) {
            console.error('Export failed:', error);
            toast.error('Failed to export report. Feature may not be fully implemented yet.');
        } finally {
            setExporting(null);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Reports & Export</h1>
                    <p className="text-muted-foreground mt-1">Generate and export business reports</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline">
                        <Calendar className="h-4 w-4 mr-2" />
                        Date Range
                    </Button>
                </div>
            </div>

            {/* Date Range Filter */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Filter className="h-5 w-5" />
                        Report Filters
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Start Date</Label>
                            <Input
                                type="date"
                                value={dateRange.start}
                                onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>End Date</Label>
                            <Input
                                type="date"
                                value={dateRange.end}
                                onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })}
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Reports Tabs */}
            <Tabs value={activeReport} onValueChange={(value: string) => setActiveReport(value as ReportType)}>
                <TabsList className="grid grid-cols-4 w-full">
                    <TabsTrigger value="inventory">
                        <Package className="h-4 w-4 mr-2" />
                        Inventory
                    </TabsTrigger>
                    <TabsTrigger value="sales">
                        <TrendingUp className="h-4 w-4 mr-2" />
                        Sales
                    </TabsTrigger>
                    <TabsTrigger value="stock-movements">
                        <List className="h-4 w-4 mr-2" />
                        Stock Movements
                    </TabsTrigger>
                    <TabsTrigger value="low-stock">
                        <BarChart3 className="h-4 w-4 mr-2" />
                        Low Stock
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="inventory" className="space-y-4">
                    <InventoryReport dateRange={dateRange} onExport={handleExport} exporting={exporting} />
                </TabsContent>

                <TabsContent value="sales" className="space-y-4">
                    <SalesReport dateRange={dateRange} onExport={handleExport} exporting={exporting} />
                </TabsContent>

                <TabsContent value="stock-movements" className="space-y-4">
                    <StockMovementsReport dateRange={dateRange} onExport={handleExport} exporting={exporting} />
                </TabsContent>

                <TabsContent value="low-stock" className="space-y-4">
                    <LowStockReport dateRange={dateRange} onExport={handleExport} exporting={exporting} />
                </TabsContent>
            </Tabs>
        </div>
    );
}

function InventoryReport({ dateRange, onExport, exporting }: {
    dateRange: { start: string; end: string };
    onExport: (format: 'csv' | 'pdf' | 'excel', reportType: ReportType) => void;
    exporting: string | null;
}) {
    const { data, isLoading } = useQuery({
        queryKey: ['inventory-report', dateRange],
        queryFn: async () => {
            const response = await api.get('/inventory?limit=1000');
            return response.data;
        },
    });

    const inventory = data?.inventory || [];
    const totalValue = inventory.reduce((sum: number, item: { quantity: number; product?: { unit_price?: number } }) =>
        sum + (item.quantity * (item.product?.unit_price || 0)), 0
    );

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <div>
                        <CardTitle>Inventory Report</CardTitle>
                        <CardDescription>Current stock levels and valuation</CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('csv', 'inventory')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'inventory-csv' ? 'Exporting...' : 'CSV'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('excel', 'inventory')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'inventory-excel' ? 'Exporting...' : 'Excel'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('pdf', 'inventory')}
                            disabled={!!exporting}
                        >
                            <FileText className="h-4 w-4 mr-2" />
                            {exporting === 'inventory-pdf' ? 'Exporting...' : 'PDF'}
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                <div className="mb-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-primary/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Total Items</p>
                        <p className="text-2xl font-bold">{inventory.length}</p>
                    </div>
                    <div className="bg-green-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Total Stock Value</p>
                        <p className="text-2xl font-bold">{formatCurrency(totalValue)}</p>
                    </div>
                    <div className="bg-orange-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Low Stock Items</p>
                        <p className="text-2xl font-bold">
                            {inventory.filter((item: { quantity: number }) => item.quantity < 10).length}
                        </p>
                    </div>
                </div>

                {isLoading ? (
                    <div className="text-center py-8 text-muted-foreground">Loading inventory data...</div>
                ) : inventory.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No inventory data available</div>
                ) : (
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Product</TableHead>
                                <TableHead>Warehouse</TableHead>
                                <TableHead>Quantity</TableHead>
                                <TableHead>Unit Price</TableHead>
                                <TableHead>Total Value</TableHead>
                                <TableHead>Status</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {inventory.slice(0, 10).map((item: {
                                id: string;
                                product?: { name?: string; unit_price?: number };
                                warehouse?: { name?: string };
                                quantity: number;
                            }) => (
                                <TableRow key={item.id}>
                                    <TableCell className="font-medium">{item.product?.name || 'N/A'}</TableCell>
                                    <TableCell>{item.warehouse?.name || 'N/A'}</TableCell>
                                    <TableCell>{item.quantity}</TableCell>
                                    <TableCell>{formatCurrency(item.product?.unit_price || 0)}</TableCell>
                                    <TableCell>{formatCurrency(item.quantity * (item.product?.unit_price || 0))}</TableCell>
                                    <TableCell>
                                        <Badge variant={item.quantity < 10 ? 'destructive' : 'success'}>
                                            {item.quantity < 10 ? 'Low' : 'OK'}
                                        </Badge>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
                {inventory.length > 10 && (
                    <p className="text-sm text-muted-foreground mt-4 text-center">
                        Showing 10 of {inventory.length} items. Export to see all.
                    </p>
                )}
            </CardContent>
        </Card>
    );
}

function SalesReport({ dateRange, onExport, exporting }: {
    dateRange: { start: string; end: string };
    onExport: (format: 'csv' | 'pdf' | 'excel', reportType: ReportType) => void;
    exporting: string | null;
}) {
    const { data, isLoading } = useQuery({
        queryKey: ['sales-report', dateRange],
        queryFn: async () => {
            const response = await api.get(`/orders?start_date=${dateRange.start}&end_date=${dateRange.end}`);
            return response.data;
        },
    });

    const orders = data?.orders || [];
    const totalSales = orders.reduce((sum: number, order: { total_amount?: number }) => sum + (order.total_amount || 0), 0);
    const completedOrders = orders.filter((order: { status?: string }) => order.status === 'delivered');

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <div>
                        <CardTitle>Sales Report</CardTitle>
                        <CardDescription>
                            Sales from {formatDate(dateRange.start)} to {formatDate(dateRange.end)}
                        </CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('csv', 'sales')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'sales-csv' ? 'Exporting...' : 'CSV'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('excel', 'sales')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'sales-excel' ? 'Exporting...' : 'Excel'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('pdf', 'sales')}
                            disabled={!!exporting}
                        >
                            <FileText className="h-4 w-4 mr-2" />
                            {exporting === 'sales-pdf' ? 'Exporting...' : 'PDF'}
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                <div className="mb-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-primary/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Total Orders</p>
                        <p className="text-2xl font-bold">{orders.length}</p>
                    </div>
                    <div className="bg-green-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Total Sales</p>
                        <p className="text-2xl font-bold">{formatCurrency(totalSales)}</p>
                    </div>
                    <div className="bg-blue-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Completed Orders</p>
                        <p className="text-2xl font-bold">{completedOrders.length}</p>
                    </div>
                </div>

                {isLoading ? (
                    <div className="text-center py-8 text-muted-foreground">Loading sales data...</div>
                ) : orders.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No sales data for selected period</div>
                ) : (
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Order ID</TableHead>
                                <TableHead>Date</TableHead>
                                <TableHead>Customer</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Amount</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {orders.slice(0, 10).map((order: {
                                id: string;
                                order_number?: string;
                                created_at?: string;
                                customer_name?: string;
                                status?: string;
                                total_amount?: number;
                            }) => (
                                <TableRow key={order.id}>
                                    <TableCell className="font-medium">{order.order_number || order.id.slice(0, 8)}</TableCell>
                                    <TableCell>{order.created_at ? formatDate(order.created_at) : 'N/A'}</TableCell>
                                    <TableCell>{order.customer_name || 'N/A'}</TableCell>
                                    <TableCell>
                                        <Badge>{order.status || 'pending'}</Badge>
                                    </TableCell>
                                    <TableCell>{formatCurrency(order.total_amount || 0)}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
                {orders.length > 10 && (
                    <p className="text-sm text-muted-foreground mt-4 text-center">
                        Showing 10 of {orders.length} orders. Export to see all.
                    </p>
                )}
            </CardContent>
        </Card>
    );
}

function StockMovementsReport({ dateRange, onExport, exporting }: {
    dateRange: { start: string; end: string };
    onExport: (format: 'csv' | 'pdf' | 'excel', reportType: ReportType) => void;
    exporting: string | null;
}) {
    const { data, isLoading } = useQuery({
        queryKey: ['stock-movements-report', dateRange],
        queryFn: async () => {
            const response = await api.get(`/stock-adjustments?start_date=${dateRange.start}&end_date=${dateRange.end}`);
            return response.data;
        },
    });

    const adjustments = data?.adjustments || [];

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <div>
                        <CardTitle>Stock Movements Report</CardTitle>
                        <CardDescription>
                            Stock adjustments from {formatDate(dateRange.start)} to {formatDate(dateRange.end)}
                        </CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('csv', 'stock-movements')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'stock-movements-csv' ? 'Exporting...' : 'CSV'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('excel', 'stock-movements')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'stock-movements-excel' ? 'Exporting...' : 'Excel'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('pdf', 'stock-movements')}
                            disabled={!!exporting}
                        >
                            <FileText className="h-4 w-4 mr-2" />
                            {exporting === 'stock-movements-pdf' ? 'Exporting...' : 'PDF'}
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                <div className="mb-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="bg-primary/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Total Movements</p>
                        <p className="text-2xl font-bold">{adjustments.length}</p>
                    </div>
                    <div className="bg-green-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Stock Increases</p>
                        <p className="text-2xl font-bold">
                            {adjustments.filter((a: { adjustment_type?: string }) =>
                                ['increase', 'return', 'transfer_in'].includes(a.adjustment_type || '')
                            ).length}
                        </p>
                    </div>
                    <div className="bg-red-500/10 p-4 rounded-lg">
                        <p className="text-sm text-muted-foreground">Stock Decreases</p>
                        <p className="text-2xl font-bold">
                            {adjustments.filter((a: { adjustment_type?: string }) =>
                                ['decrease', 'damage', 'transfer_out'].includes(a.adjustment_type || '')
                            ).length}
                        </p>
                    </div>
                </div>

                {isLoading ? (
                    <div className="text-center py-8 text-muted-foreground">Loading stock movements...</div>
                ) : adjustments.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">No stock movements for selected period</div>
                ) : (
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
                            {adjustments.slice(0, 10).map((adj: {
                                id: string;
                                adjusted_at?: string;
                                adjustment_type?: string;
                                quantity?: number;
                                previous_stock?: number;
                                new_stock?: number;
                                reason?: string;
                            }) => (
                                <TableRow key={adj.id}>
                                    <TableCell>{adj.adjusted_at ? formatDate(adj.adjusted_at) : 'N/A'}</TableCell>
                                    <TableCell>
                                        <Badge variant={adj.quantity && adj.quantity > 0 ? 'success' : 'destructive'}>
                                            {adj.adjustment_type || 'N/A'}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className={adj.quantity && adj.quantity >= 0 ? 'text-green-600' : 'text-red-600'}>
                                        {adj.quantity && adj.quantity >= 0 ? '+' : ''}{adj.quantity || 0}
                                    </TableCell>
                                    <TableCell>{adj.previous_stock || 0}</TableCell>
                                    <TableCell className="font-semibold">{adj.new_stock || 0}</TableCell>
                                    <TableCell className="text-muted-foreground text-sm">{adj.reason || '—'}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
                {adjustments.length > 10 && (
                    <p className="text-sm text-muted-foreground mt-4 text-center">
                        Showing 10 of {adjustments.length} movements. Export to see all.
                    </p>
                )}
            </CardContent>
        </Card>
    );
}

function LowStockReport({ dateRange, onExport, exporting }: {
    dateRange: { start: string; end: string };
    onExport: (format: 'csv' | 'pdf' | 'excel', reportType: ReportType) => void;
    exporting: string | null;
}) {
    const { data, isLoading } = useQuery({
        queryKey: ['low-stock-report'],
        queryFn: async () => {
            const response = await api.get('/inventory?limit=1000');
            return response.data;
        },
    });

    const inventory = data?.inventory || [];
    const lowStockItems = inventory.filter((item: { quantity: number }) => item.quantity < 10);

    return (
        <Card>
            <CardHeader>
                <div className="flex justify-between items-center">
                    <div>
                        <CardTitle>Low Stock Alert Report</CardTitle>
                        <CardDescription>Items below minimum stock threshold (10 units)</CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('csv', 'low-stock')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'low-stock-csv' ? 'Exporting...' : 'CSV'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('excel', 'low-stock')}
                            disabled={!!exporting}
                        >
                            <Download className="h-4 w-4 mr-2" />
                            {exporting === 'low-stock-excel' ? 'Exporting...' : 'Excel'}
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onExport('pdf', 'low-stock')}
                            disabled={!!exporting}
                        >
                            <FileText className="h-4 w-4 mr-2" />
                            {exporting === 'low-stock-pdf' ? 'Exporting...' : 'PDF'}
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent>
                <div className="mb-4 bg-orange-500/10 p-4 rounded-lg">
                    <p className="text-sm text-muted-foreground">Critical Items (Below Threshold)</p>
                    <p className="text-2xl font-bold text-orange-600">{lowStockItems.length}</p>
                </div>

                {isLoading ? (
                    <div className="text-center py-8 text-muted-foreground">Loading low stock data...</div>
                ) : lowStockItems.length === 0 ? (
                    <div className="text-center py-8 text-green-600">
                        <Package className="h-12 w-12 mx-auto mb-2" />
                        <p>All items are adequately stocked!</p>
                    </div>
                ) : (
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Product</TableHead>
                                <TableHead>Warehouse</TableHead>
                                <TableHead>Current Stock</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Reorder Urgency</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {lowStockItems.map((item: {
                                id: string;
                                product?: { name?: string };
                                warehouse?: { name?: string };
                                quantity: number;
                            }) => (
                                <TableRow key={item.id}>
                                    <TableCell className="font-medium">{item.product?.name || 'N/A'}</TableCell>
                                    <TableCell>{item.warehouse?.name || 'N/A'}</TableCell>
                                    <TableCell>
                                        <span className="text-orange-600 font-semibold">{item.quantity}</span>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={item.quantity === 0 ? 'destructive' : 'warning'}>
                                            {item.quantity === 0 ? 'Out of Stock' : 'Low Stock'}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={item.quantity < 5 ? 'destructive' : 'warning'}>
                                            {item.quantity < 5 ? 'High' : 'Medium'}
                                        </Badge>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
            </CardContent>
        </Card>
    );
}
