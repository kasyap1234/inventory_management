'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Edit, AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { toast } from 'react-hot-toast';
import api from '@/lib/api';

interface AlertRule {
    id: string;
    name: string;
    description?: string;
    event_type: string;
    conditions: Record<string, unknown>;
    actions: Array<Record<string, unknown>>;
    is_active: boolean;
    created_at: string;
}

export default function AlertRulesPage() {
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
    const queryClient = useQueryClient();

    const { data: rules, isLoading } = useQuery<AlertRule[]>({
        queryKey: ['alert-rules'],
        queryFn: async () => {
            const response = await api.get('/alert-rules');
            return response.data.alert_rules || [];
        },
    });

    const deleteMutation = useMutation({
        mutationFn: async (id: string) => {
            await api.delete(`/alert-rules/${id}`);
        },
        onSuccess: () => {
            toast.success('Alert rule deleted');
            queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to delete rule');
        },
    });

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Alert Rules</h1>
                    <p className="text-gray-500 mt-1">Configure automated alerts and notifications</p>
                </div>
                <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Rule
                </Button>
            </div>

            {isLoading ? (
                <div className="text-center py-12">Loading rules...</div>
            ) : rules && rules.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {rules.map((rule) => (
                        <Card key={rule.id} className="relative overflow-hidden">
                            <div className={`absolute top-0 left-0 w-1 h-full ${rule.is_active ? 'bg-green-500' : 'bg-gray-300'}`} />
                            <CardHeader className="pb-2">
                                <div className="flex justify-between items-start">
                                    <CardTitle className="text-lg">{rule.name}</CardTitle>
                                    <div className="flex gap-1">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 w-8 p-0"
                                            onClick={() => setEditingRule(rule)}
                                        >
                                            <Edit className="h-4 w-4 text-gray-500" />
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            className="h-8 w-8 p-0"
                                            onClick={() => {
                                                if (confirm('Are you sure you want to delete this rule?')) {
                                                    deleteMutation.mutate(rule.id);
                                                }
                                            }}
                                        >
                                            <Trash2 className="h-4 w-4 text-red-500" />
                                        </Button>
                                    </div>
                                </div>
                                <CardDescription>{rule.description || 'No description'}</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="space-y-2 text-sm">
                                    <div className="flex justify-between">
                                        <span className="text-gray-500">Event Type:</span>
                                        <span className="font-medium">{rule.event_type}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-gray-500">Status:</span>
                                        <span className={`font-medium ${rule.is_active ? 'text-green-600' : 'text-gray-500'}`}>
                                            {rule.is_active ? 'Active' : 'Inactive'}
                                        </span>
                                    </div>
                                    <div className="pt-2">
                                        <span className="text-gray-500 block mb-1">Conditions:</span>
                                        <pre className="bg-gray-50 p-2 rounded text-xs overflow-x-auto">
                                            {JSON.stringify(rule.conditions, null, 2)}
                                        </pre>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            ) : (
                <div className="text-center py-12 bg-gray-50 rounded-lg border border-dashed border-gray-300">
                    <AlertTriangle className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-semibold text-foreground">No alert rules</h3>
                    <p className="mt-1 text-sm text-gray-500">Get started by creating a new alert rule.</p>
                    <div className="mt-6">
                        <Button onClick={() => setIsCreateDialogOpen(true)}>
                            <Plus className="mr-2 h-4 w-4" />
                            Create Rule
                        </Button>
                    </div>
                </div>
            )}

            <RuleDialog
                open={isCreateDialogOpen || !!editingRule}
                onOpenChange={(open) => {
                    setIsCreateDialogOpen(open);
                    if (!open) setEditingRule(null);
                }}
                rule={editingRule}
            />
        </div>
    );
}

function RuleDialog({
    open,
    onOpenChange,
    rule,
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    rule: AlertRule | null;
}) {
    const queryClient = useQueryClient();
    const [formData, setFormData] = useState({
        name: rule?.name || '',
        description: rule?.description || '',
        event_type: rule?.event_type || 'inventory.low_stock',
        conditions: rule ? JSON.stringify(rule.conditions, null, 2) : '{\n  "threshold": 10\n}',
        actions: rule ? JSON.stringify(rule.actions, null, 2) : '[\n  {\n    "type": "email",\n    "recipient": "admin@example.com"\n  }\n]',
        is_active: rule?.is_active ?? true,
    });

    const saveMutation = useMutation({
        mutationFn: async (data: typeof formData) => {
            const payload = {
                ...data,
                conditions: JSON.parse(data.conditions),
                actions: JSON.parse(data.actions),
            };

            if (rule) {
                await api.put(`/alert-rules/${rule.id}`, payload);
            } else {
                await api.post('/alert-rules', payload);
            }
        },
        onSuccess: () => {
            toast.success(rule ? 'Rule updated' : 'Rule created');
            queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
            onOpenChange(false);
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to save rule');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        try {
            JSON.parse(formData.conditions);
            JSON.parse(formData.actions);
            saveMutation.mutate(formData);
        } catch {
            toast.error('Invalid JSON in conditions or actions');
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl">
                <DialogHeader>
                    <DialogTitle>{rule ? 'Edit Rule' : 'Create Rule'}</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Rule Name</Label>
                            <Input
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Event Type</Label>
                            <select
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                value={formData.event_type}
                                onChange={(e) => setFormData({ ...formData, event_type: e.target.value })}
                            >
                                <option value="inventory.low_stock">Low Stock</option>
                                <option value="inventory.out_of_stock">Out of Stock</option>
                                <option value="order.created">Order Created</option>
                                <option value="order.status_changed">Order Status Changed</option>
                            </select>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>Description</Label>
                        <Input
                            value={formData.description}
                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Conditions (JSON)</Label>
                            <Textarea
                                className="font-mono text-xs h-32"
                                value={formData.conditions}
                                onChange={(e) => setFormData({ ...formData, conditions: e.target.value })}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Actions (JSON)</Label>
                            <Textarea
                                className="font-mono text-xs h-32"
                                value={formData.actions}
                                onChange={(e) => setFormData({ ...formData, actions: e.target.value })}
                            />
                        </div>
                    </div>

                    <div className="flex items-center space-x-2">
                        <Checkbox
                            checked={formData.is_active}
                            onCheckedChange={(checked: boolean) => setFormData({ ...formData, is_active: checked })}
                        />
                        <Label>Active</Label>
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={saveMutation.isPending}>
                            {saveMutation.isPending ? 'Saving...' : 'Save Rule'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
