'use client';

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { FileText, Upload, Download, RefreshCw, CheckCircle, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'react-hot-toast';
import api from '@/lib/api';

export default function TallyPage() {
    const [activeTab, setActiveTab] = useState<'export' | 'import'>('export');

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold text-gray-900">Tally Integration</h1>
                <p className="text-gray-500 mt-1">Export data to Tally or import data from Tally</p>
            </div>

            <div className="flex gap-4 border-b border-gray-200">
                <button
                    onClick={() => setActiveTab('export')}
                    className={`px-4 py-2 font-medium transition-colors ${activeTab === 'export'
                            ? 'border-b-2 border-blue-600 text-blue-600'
                            : 'text-gray-600 hover:text-gray-900'
                        }`}
                >
                    <Download className="inline h-4 w-4 mr-2" />
                    Export to Tally
                </button>
                <button
                    onClick={() => setActiveTab('import')}
                    className={`px-4 py-2 font-medium transition-colors ${activeTab === 'import'
                            ? 'border-b-2 border-blue-600 text-blue-600'
                            : 'text-gray-600 hover:text-gray-900'
                        }`}
                >
                    <Upload className="inline h-4 w-4 mr-2" />
                    Import from Tally
                </button>
            </div>

            {activeTab === 'export' ? <ExportTab /> : <ImportTab />}
        </div>
    );
}

function ExportTab() {
    const [formData, setFormData] = useState({
        startDate: '',
        endDate: '',
        format: 'xml',
        dataType: 'masters',
    });

    const exportMutation = useMutation({
        mutationFn: async (data: typeof formData) => {
            const response = await api.post('/tally/export', {
                start_date: data.startDate,
                end_date: data.endDate,
                format: data.format,
                data_type: data.dataType,
            });
            return response.data;
        },
        onSuccess: () => {
            toast.success('Export job queued successfully');
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.message || 'Failed to queue export job');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        exportMutation.mutate(formData);
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>Export Data</CardTitle>
                <CardDescription>Generate Tally-compatible XML/CSV files</CardDescription>
            </CardHeader>
            <CardContent>
                <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
                    <div className="space-y-2">
                        <Label>Data Type</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            value={formData.dataType}
                            onChange={(e) => setFormData({ ...formData, dataType: e.target.value })}
                        >
                            <option value="masters">Masters (Ledgers, Items)</option>
                            <option value="vouchers">Vouchers (Sales, Purchases)</option>
                        </select>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Start Date</Label>
                            <Input
                                type="date"
                                value={formData.startDate}
                                onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                                required={formData.dataType === 'vouchers'}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>End Date</Label>
                            <Input
                                type="date"
                                value={formData.endDate}
                                onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                                required={formData.dataType === 'vouchers'}
                            />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>Format</Label>
                        <div className="flex gap-4">
                            <label className="flex items-center space-x-2">
                                <input
                                    type="radio"
                                    value="xml"
                                    checked={formData.format === 'xml'}
                                    onChange={(e) => setFormData({ ...formData, format: e.target.value })}
                                    className="h-4 w-4 text-primary"
                                />
                                <span>XML (Tally Default)</span>
                            </label>
                            <label className="flex items-center space-x-2">
                                <input
                                    type="radio"
                                    value="csv"
                                    checked={formData.format === 'csv'}
                                    onChange={(e) => setFormData({ ...formData, format: e.target.value })}
                                    className="h-4 w-4 text-primary"
                                />
                                <span>CSV</span>
                            </label>
                        </div>
                    </div>

                    <Button type="submit" disabled={exportMutation.isPending}>
                        {exportMutation.isPending ? (
                            <>
                                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                                Exporting...
                            </>
                        ) : (
                            <>
                                <Download className="mr-2 h-4 w-4" />
                                Start Export
                            </>
                        )}
                    </Button>
                </form>
            </CardContent>
        </Card>
    );
}

function ImportTab() {
    const [dataType, setDataType] = useState('orders');
    const [fileContent, setFileContent] = useState('');
    const [dragActive, setDragActive] = useState(false);

    const importMutation = useMutation({
        mutationFn: async () => {
            const response = await api.post('/tally/import', {
                data_type: dataType,
                data: fileContent,
            });
            return response.data;
        },
        onSuccess: (data) => {
            toast.success(data.message || 'Import completed successfully');
            setFileContent('');
        },
        onError: (error: any) => {
            toast.error(error.response?.data?.message || 'Failed to import data');
        },
    });

    const handleDrag = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        if (e.type === 'dragenter' || e.type === 'dragover') {
            setDragActive(true);
        } else if (e.type === 'dragleave') {
            setDragActive(false);
        }
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setDragActive(false);

        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            const file = e.dataTransfer.files[0];
            const reader = new FileReader();
            reader.onload = (e) => {
                if (e.target?.result) {
                    setFileContent(e.target.result as string);
                }
            };
            reader.readAsText(file);
        }
    };

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            const file = e.target.files[0];
            const reader = new FileReader();
            reader.onload = (e) => {
                if (e.target?.result) {
                    setFileContent(e.target.result as string);
                }
            };
            reader.readAsText(file);
        }
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>Import Data</CardTitle>
                <CardDescription>Import Tally XML/CSV data into the system</CardDescription>
            </CardHeader>
            <CardContent>
                <div className="space-y-4 max-w-xl">
                    <div className="space-y-2">
                        <Label>Data Type</Label>
                        <select
                            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                            value={dataType}
                            onChange={(e) => setDataType(e.target.value)}
                        >
                            <option value="orders">Orders</option>
                            <option value="invoices">Invoices</option>
                        </select>
                    </div>

                    <div
                        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${dragActive ? 'border-primary bg-primary/5' : 'border-gray-300'
                            }`}
                        onDragEnter={handleDrag}
                        onDragLeave={handleDrag}
                        onDragOver={handleDrag}
                        onDrop={handleDrop}
                    >
                        <Upload className="mx-auto h-12 w-12 text-gray-400" />
                        <p className="mt-2 text-sm text-gray-600">
                            Drag and drop your file here, or{' '}
                            <label className="text-primary hover:underline cursor-pointer">
                                browse
                                <input type="file" className="hidden" onChange={handleFileChange} accept=".xml,.csv,.txt" />
                            </label>
                        </p>
                        <p className="text-xs text-gray-500 mt-1">Supports XML and CSV</p>
                    </div>

                    {fileContent && (
                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <Label>File Content Preview</Label>
                                <Button variant="ghost" size="sm" onClick={() => setFileContent('')} className="text-red-500 h-auto p-0 hover:text-red-700">
                                    Clear
                                </Button>
                            </div>
                            <textarea
                                className="flex min-h-[200px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 font-mono"
                                value={fileContent}
                                onChange={(e) => setFileContent(e.target.value)}
                                placeholder="Paste XML/CSV content here..."
                            />
                        </div>
                    )}

                    <Button
                        onClick={() => importMutation.mutate()}
                        disabled={importMutation.isPending || !fileContent}
                        className="w-full"
                    >
                        {importMutation.isPending ? (
                            <>
                                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                                Importing...
                            </>
                        ) : (
                            <>
                                <Upload className="mr-2 h-4 w-4" />
                                Import Data
                            </>
                        )}
                    </Button>

                    {importMutation.isError && (
                        <div className="p-4 bg-red-50 text-red-700 rounded-md flex items-start gap-2">
                            <AlertCircle className="h-5 w-5 mt-0.5" />
                            <div>
                                <p className="font-medium">Import Failed</p>
                                <p className="text-sm">{(importMutation.error as any)?.response?.data?.message || 'Unknown error'}</p>
                            </div>
                        </div>
                    )}

                    {importMutation.isSuccess && (
                        <div className="p-4 bg-green-50 text-green-700 rounded-md flex items-start gap-2">
                            <CheckCircle className="h-5 w-5 mt-0.5" />
                            <div>
                                <p className="font-medium">Import Successful</p>
                                <p className="text-sm">Data has been processed successfully.</p>
                            </div>
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}
