'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Edit, Trash2, Send, Mail, MessageSquare, Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { toast } from 'react-hot-toast';
import api from '@/lib/api';
import { formatDate } from '@/lib/utils';

interface NotificationTemplate {
    id: string;
    name: string;
    description?: string;
    channel: string;
    subject?: string;
    body: string;
    variables?: string[];
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

const CHANNEL_ICONS = {
    email: Mail,
    sms: MessageSquare,
    in_app: Bell,
};

const CHANNEL_COLORS = {
    email: 'text-blue-600 bg-blue-100',
    sms: 'text-green-600 bg-green-100',
    in_app: 'text-purple-600 bg-purple-100',
};

export default function NotificationTemplatesPage() {
    const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
    const [editingTemplate, setEditingTemplate] = useState<NotificationTemplate | null>(null);
    const [testingTemplate, setTestingTemplate] = useState<NotificationTemplate | null>(null);
    const queryClient = useQueryClient();

    const { data: templatesData, isLoading } = useQuery({
        queryKey: ['notification-templates'],
        queryFn: async () => {
            const response = await api.get('/notification-templates');
            return response.data;
        },
    });

    const deleteMutation = useMutation({
        mutationFn: async (id: string) => {
            await api.delete(`/notification-templates/${id}`);
        },
        onSuccess: () => {
            toast.success('Template deleted successfully');
            queryClient.invalidateQueries({ queryKey: ['notification-templates'] });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to delete template');
        },
    });

    const templates: NotificationTemplate[] = templatesData?.templates || [];

    const getChannelIcon = (channel: string) => {
        const Icon = CHANNEL_ICONS[channel as keyof typeof CHANNEL_ICONS] || Bell;
        return Icon;
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">Notification Templates</h1>
                    <p className="text-muted-foreground mt-1">Manage email, SMS, and in-app notification templates</p>
                </div>
                <Button onClick={() => setIsCreateDialogOpen(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    Create Template
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Templates</CardTitle>
                    <CardDescription>{templates.length} template{templates.length !== 1 ? 's' : ''} configured</CardDescription>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <div className="text-center py-8 text-muted-foreground">Loading templates...</div>
                    ) : templates.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            <Bell className="h-12 w-12 mx-auto mb-4 text-muted-foreground/50" />
                            <p>No notification templates found</p>
                            <Button variant="link" onClick={() => setIsCreateDialogOpen(true)}>
                                Create your first template
                            </Button>
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Channel</TableHead>
                                    <TableHead>Subject</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Last Updated</TableHead>
                                    <TableHead>Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {templates.map((template) => {
                                    const Icon = getChannelIcon(template.channel);
                                    const colorClass = CHANNEL_COLORS[template.channel as keyof typeof CHANNEL_COLORS] || 'text-gray-600 bg-gray-100';

                                    return (
                                        <TableRow key={template.id}>
                                            <TableCell className="font-medium">{template.name}</TableCell>
                                            <TableCell>
                                                <div className={`inline-flex items-center gap-1 px-2 py-1 rounded ${colorClass}`}>
                                                    <Icon className="h-3 w-3" />
                                                    <span className="text-xs font-medium capitalize">{template.channel}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell className="max-w-xs truncate">{template.subject || '—'}</TableCell>
                                            <TableCell>
                                                <Badge variant={template.is_active ? 'success' : 'secondary'}>
                                                    {template.is_active ? 'Active' : 'Inactive'}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>{formatDate(template.updated_at)}</TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => setTestingTemplate(template)}
                                                    >
                                                        <Send className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => setEditingTemplate(template)}
                                                    >
                                                        <Edit className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => {
                                                            if (confirm('Delete this template?')) {
                                                                deleteMutation.mutate(template.id);
                                                            }
                                                        }}
                                                    >
                                                        <Trash2 className="h-4 w-4 text-destructive" />
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

            <TemplateFormDialog
                open={isCreateDialogOpen || !!editingTemplate}
                onOpenChange={(open) => {
                    setIsCreateDialogOpen(open);
                    if (!open) setEditingTemplate(null);
                }}
                template={editingTemplate}
            />

            <TestTemplateDialog
                open={!!testingTemplate}
                onOpenChange={(open) => !open && setTestingTemplate(null)}
                template={testingTemplate}
            />
        </div>
    );
}

function TemplateFormDialog({ open, onOpenChange, template }: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    template?: NotificationTemplate | null;
}) {
    const queryClient = useQueryClient();
    const [formData, setFormData] = useState({
        name: template?.name || '',
        description: template?.description || '',
        channel: template?.channel || 'email',
        subject: template?.subject || '',
        body: template?.body || '',
        is_active: template?.is_active ?? true,
    });

    const saveMutation = useMutation({
        mutationFn: async (data: typeof formData) => {
            if (template) {
                await api.put(`/notification-templates/${template.id}`, data);
            } else {
                await api.post('/notification-templates', data);
            }
        },
        onSuccess: () => {
            toast.success(template ? 'Template updated' : 'Template created');
            queryClient.invalidateQueries({ queryKey: ['notification-templates'] });
            onOpenChange(false);
            setFormData({
                name: '',
                description: '',
                channel: 'email',
                subject: '',
                body: '',
                is_active: true,
            });
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to save template');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        saveMutation.mutate(formData);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl">
                <DialogHeader>
                    <DialogTitle>{template ? 'Edit Template' : 'Create Template'}</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label>Template Name *</Label>
                            <Input
                                required
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                placeholder="Order Confirmation"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Channel *</Label>
                            <select
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                value={formData.channel}
                                onChange={(e) => setFormData({ ...formData, channel: e.target.value })}
                                required
                            >
                                <option value="email">Email</option>
                                <option value="sms">SMS</option>
                                <option value="in_app">In-App Notification</option>
                            </select>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <Label>Description</Label>
                        <Input
                            value={formData.description}
                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            placeholder="Brief description of this template"
                        />
                    </div>

                    {(formData.channel === 'email' || formData.channel === 'in_app') && (
                        <div className="space-y-2">
                            <Label>Subject {formData.channel === 'email' ? '*' : ''}</Label>
                            <Input
                                required={formData.channel === 'email'}
                                value={formData.subject}
                                onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
                                placeholder="Your order has been confirmed"
                            />
                        </div>
                    )}

                    <div className="space-y-2">
                        <Label>Message Body *</Label>
                        <textarea
                            required
                            className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                            value={formData.body}
                            onChange={(e) => setFormData({ ...formData, body: e.target.value })}
                            placeholder="Hello {{customer_name}}, your order #{{order_number}} has been confirmed..."
                        />
                        <p className="text-xs text-muted-foreground">
                            Use {'{{variable_name}}'} for dynamic content (e.g., {'{{customer_name}}'}, {'{{order_number}}'})
                        </p>
                    </div>

                    <div className="flex items-center gap-2">
                        <input
                            type="checkbox"
                            id="is_active"
                            checked={formData.is_active}
                            onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                            className="h-4 w-4 rounded border-gray-300"
                        />
                        <Label htmlFor="is_active" className="font-normal cursor-pointer">
                            Active (template can be used for notifications)
                        </Label>
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={saveMutation.isPending}>
                            {saveMutation.isPending ? 'Saving...' : template ? 'Update Template' : 'Create Template'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}

function TestTemplateDialog({ open, onOpenChange, template }: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    template: NotificationTemplate | null;
}) {
    const [recipient, setRecipient] = useState('');
    const [variables, setVariables] = useState('{}');

    const testMutation = useMutation({
        mutationFn: async () => {
            if (!template) return;

            let parsedVariables = {};
            try {
                parsedVariables = JSON.parse(variables);
            } catch {
                throw new Error('Invalid JSON for variables');
            }

            await api.post(`/notification-templates/${template.id}/test`, {
                recipient,
                variables: parsedVariables,
            });
        },
        onSuccess: () => {
            toast.success('Test notification sent successfully');
            onOpenChange(false);
            setRecipient('');
            setVariables('{}');
        },
        onError: (error: unknown) => {
            const err = error as { response?: { data?: { message?: string } } };
            toast.error(err.response?.data?.message || 'Failed to send test notification');
        },
    });

    const handleTest = (e: React.FormEvent) => {
        e.preventDefault();
        testMutation.mutate();
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Test Template - {template?.name}</DialogTitle>
                </DialogHeader>
                <form onSubmit={handleTest} className="space-y-4">
                    <div className="space-y-2">
                        <Label>
                            Recipient {template?.channel === 'email' ? 'Email' : template?.channel === 'sms' ? 'Phone Number' : 'User ID'}
                        </Label>
                        <Input
                            required
                            value={recipient}
                            onChange={(e) => setRecipient(e.target.value)}
                            placeholder={
                                template?.channel === 'email' ? 'test@example.com' :
                                    template?.channel === 'sms' ? '+1234567890' :
                                        'user-id'
                            }
                        />
                    </div>

                    <div className="space-y-2">
                        <Label>Variables (JSON)</Label>
                        <textarea
                            className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono"
                            value={variables}
                            onChange={(e) => setVariables(e.target.value)}
                            placeholder='{"customer_name": "John Doe", "order_number": "12345"}'
                        />
                        <p className="text-xs text-muted-foreground">
                            Provide test values for template variables in JSON format
                        </p>
                    </div>

                    <DialogFooter>
                        <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                            Cancel
                        </Button>
                        <Button type="submit" disabled={testMutation.isPending}>
                            {testMutation.isPending ? 'Sending...' : 'Send Test'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
