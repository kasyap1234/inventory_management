'use client';

import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, RefreshCcw, Link, Shield, Trash2, RotateCcw, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { webhookService } from '@/lib/services';
import { formatDateTime } from '@/lib/utils';

type WebhookSubscription = {
  id: string;
  name: string;
  description?: string | null;
  url: string;
  secret: string;
  events: string[];
  is_active: boolean;
  last_used_at?: string | null;
  created_at?: string;
  updated_at?: string;
};

type WebhookFormState = {
  name: string;
  description: string;
  url: string;
  secret: string;
  events: string[];
  is_active: boolean;
};

const EVENT_OPTIONS = [
  { value: 'order.created', label: 'Order Created' },
  { value: 'order.updated', label: 'Order Updated' },
  { value: 'order.delivered', label: 'Order Delivered' },
  { value: 'invoice.created', label: 'Invoice Created' },
  { value: 'inventory.low_stock', label: 'Low Stock Alert' },
  { value: 'subscription.renewed', label: 'Subscription Renewed' },
];

export default function WebhooksPage() {
  const queryClient = useQueryClient();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<WebhookSubscription | null>(null);

  const { data, isLoading, isFetching } = useQuery<{ webhook_subscriptions: WebhookSubscription[] }>({
    queryKey: ['webhooks', 'subscriptions'],
    queryFn: async () => {
      const response = await webhookService.list();
      return response.data;
    },
    staleTime: 30_000,
  });

  const subscriptions = useMemo(() => data?.webhook_subscriptions ?? [], [data]);

  const createWebhook = useMutation({
    mutationFn: (payload: WebhookFormState) => webhookService.create({
      name: payload.name,
      description: payload.description || undefined,
      url: payload.url,
      secret: payload.secret,
      events: payload.events,
      is_active: payload.is_active,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', 'subscriptions'] });
      setIsDialogOpen(false);
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to create webhook');
    },
  });

  const updateWebhook = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<WebhookFormState> }) =>
      webhookService.update(id, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', 'subscriptions'] });
      setEditingWebhook(null);
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to update webhook');
    },
  });

  const deleteWebhook = useMutation({
    mutationFn: (id: string) => webhookService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', 'subscriptions'] });
    },
    onError: (error: any) => {
      alert(error.response?.data?.message || 'Failed to delete webhook');
    },
  });

  const handleCreate = (form: WebhookFormState) => {
    createWebhook.mutate(form);
  };

  const handleUpdate = (id: string, form: WebhookFormState) => {
    updateWebhook.mutate({
      id,
      payload: {
        name: form.name,
        description: form.description,
        url: form.url,
        secret: form.secret,
        events: form.events,
        is_active: form.is_active,
      },
    });
  };

  const handleToggleActive = (subscription: WebhookSubscription) => {
    updateWebhook.mutate({
      id: subscription.id,
      payload: { is_active: !subscription.is_active },
    });
  };

  const handleRotateSecret = (subscription: WebhookSubscription) => {
    const nextSecret = crypto.randomUUID().replace(/-/g, '');
    if (confirm('Rotate secret for this webhook subscription?')) {
      updateWebhook.mutate({
        id: subscription.id,
        payload: { secret: nextSecret },
      });
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Webhook Subscriptions</h1>
          <p className="text-gray-500 mt-1">Manage outbound integrations and event delivery</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries({ queryKey: ['webhooks', 'subscriptions'] })}
            disabled={isFetching}
          >
            <RefreshCcw className={`h-4 w-4 mr-2 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button onClick={() => setIsDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Webhook
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-semibold text-gray-900 flex items-center gap-2">
            <Link className="h-5 w-5 text-blue-600" />
            Active Integrations
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-10">Loading webhook subscriptions...</div>
          ) : subscriptions.length === 0 ? (
            <div className="text-center py-16 text-gray-500">
              <Shield className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>No webhook subscriptions configured</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Endpoint</TableHead>
                  <TableHead>Events</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead className="w-40">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {subscriptions.map((subscription) => (
                  <TableRow key={subscription.id}>
                    <TableCell>
                      <div className="font-semibold text-gray-900">{subscription.name}</div>
                      {subscription.description && (
                        <div className="text-sm text-gray-500">{subscription.description}</div>
                      )}
                    </TableCell>
                    <TableCell>
                      <a
                        href={subscription.url}
                        target="_blank"
                        rel="noreferrer"
                        className="text-sm text-blue-600 hover:underline"
                      >
                        {subscription.url}
                      </a>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {subscription.events.map((event) => (
                          <Badge key={event} variant="secondary" className="text-xs">
                            {event}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={subscription.is_active ? 'success' : 'secondary'}>
                        {subscription.is_active ? 'Active' : 'Disabled'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {subscription.last_used_at ? formatDateTime(subscription.last_used_at) : '—'}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setEditingWebhook(subscription)}
                        >
                          <Shield className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleToggleActive(subscription)}
                          disabled={updateWebhook.isPending}
                        >
                          {subscription.is_active ? (
                            <>
                              <XCircle className="h-4 w-4 mr-1" /> Disable
                            </>
                          ) : (
                            <>
                              <CheckCircle className="h-4 w-4 mr-1" /> Enable
                            </>
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRotateSecret(subscription)}
                          disabled={updateWebhook.isPending}
                        >
                          <RotateCcw className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            if (confirm('Delete this webhook subscription?')) {
                              deleteWebhook.mutate(subscription.id);
                            }
                          }}
                          disabled={deleteWebhook.isPending}
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <WebhookFormDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        title="Create Webhook Subscription"
        onSubmit={handleCreate}
        isSubmitting={createWebhook.isPending}
      />

      <WebhookFormDialog
        open={!!editingWebhook}
        onOpenChange={(open) => {
          if (!open) setEditingWebhook(null);
        }}
        title={editingWebhook ? `Update ${editingWebhook.name}` : 'Update Webhook'}
        webhook={editingWebhook ?? undefined}
        onSubmit={(form) => {
          if (editingWebhook) {
            handleUpdate(editingWebhook.id, form);
          }
        }}
        isSubmitting={updateWebhook.isPending}
        mode="edit"
      />
    </div>
  );
}

function WebhookFormDialog({
  open,
  onOpenChange,
  title,
  webhook,
  onSubmit,
  isSubmitting,
  mode = 'create',
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  webhook?: WebhookSubscription;
  onSubmit: (form: WebhookFormState) => void;
  isSubmitting: boolean;
  mode?: 'create' | 'edit';
}) {
  const [formState, setFormState] = useState<WebhookFormState>(() => ({
    name: webhook?.name ?? '',
    description: webhook?.description ?? '',
    url: webhook?.url ?? '',
    secret: webhook?.secret ?? generateSecret(),
    events: webhook?.events ?? [],
    is_active: webhook?.is_active ?? true,
  }));

  useEffect(() => {
    setFormState({
      name: webhook?.name ?? '',
      description: webhook?.description ?? '',
      url: webhook?.url ?? '',
      secret: webhook?.secret ?? generateSecret(),
      events: webhook?.events ?? [],
      is_active: webhook?.is_active ?? true,
    });
  }, [webhook]);

  const toggleEvent = (event: string) => {
    setFormState((prev) => {
      const exists = prev.events.includes(event);
      return {
        ...prev,
        events: exists ? prev.events.filter((item) => item !== event) : [...prev.events, event],
      };
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(event) => {
            event.preventDefault();
            if (formState.events.length === 0) {
              alert('Select at least one event');
              return;
            }
            onSubmit(formState);
          }}
        >
          <div className="space-y-2">
            <label className="text-sm font-medium">Name *</label>
            <Input
              required
              value={formState.name}
              onChange={(e) => setFormState((prev) => ({ ...prev, name: e.target.value }))}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Input
              value={formState.description}
              onChange={(e) => setFormState((prev) => ({ ...prev, description: e.target.value }))}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Endpoint URL *</label>
            <Input
              required
              type="url"
              placeholder="https://example.com/webhooks"
              value={formState.url}
              onChange={(e) => setFormState((prev) => ({ ...prev, url: e.target.value }))}
            />
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium">Secret *</label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setFormState((prev) => ({ ...prev, secret: generateSecret() }))}
              >
                <RotateCcw className="h-4 w-4 mr-1" />
                Regenerate
              </Button>
            </div>
            <Input
              required
              value={formState.secret}
              onChange={(e) => setFormState((prev) => ({ ...prev, secret: e.target.value }))}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Subscribed Events *</label>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
              {EVENT_OPTIONS.map((option) => (
                <label
                  key={option.value}
                  className="flex items-center gap-2 rounded border border-gray-200 px-3 py-2 text-sm hover:bg-gray-50"
                >
                  <input
                    type="checkbox"
                    checked={formState.events.includes(option.value)}
                    onChange={() => toggleEvent(option.value)}
                    className="h-4 w-4 rounded border-gray-300"
                  />
                  {option.label}
                </label>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Status</label>
            <select
              value={formState.is_active ? 'active' : 'inactive'}
              onChange={(e) => setFormState((prev) => ({ ...prev, is_active: e.target.value === 'active' }))}
              className="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
            >
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : mode === 'create' ? 'Create' : 'Save'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function generateSecret() {
  return crypto.randomUUID().replace(/-/g, '');
}
