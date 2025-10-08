'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { RefreshCw, PlayCircle, XCircle, Clock, CheckCircle2, AlertCircle, Eye, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import api from '@/lib/api';
import { formatDate } from '@/lib/utils';
import { showSuccess, showError, getErrorMessage } from '@/lib/toast';

interface Job {
  id: string;
  type: string;
  payload: any;
  state: string;
  max_retry: number;
  retried: number;
  last_err: string | null;
  queue: string;
  created_at: string;
  processed_at?: string;
}

export default function JobsPage() {
  const [selectedJob, setSelectedJob] = useState<Job | null>(null);
  const [viewLogsJob, setViewLogsJob] = useState<Job | null>(null);
  const queryClient = useQueryClient();

  const { data: jobs, isLoading, refetch } = useQuery<{ jobs: Job[]; stats: any }>({
    queryKey: ['background-jobs'],
    queryFn: async () => {
      const response = await api.get('/jobs');
      return response.data;
    },
    refetchInterval: 5000, // Auto-refresh every 5 seconds
  });

  const retryJob = useMutation({
    mutationFn: async (jobId: string) => {
      await api.post(`/jobs/${jobId}/retry`);
    },
    onSuccess: () => {
      showSuccess('Job queued for retry');
      queryClient.invalidateQueries({ queryKey: ['background-jobs'] });
    },
    onError: (error) => {
      showError(getErrorMessage(error));
    },
  });

  const cancelJob = useMutation({
    mutationFn: async (jobId: string) => {
      await api.post(`/jobs/${jobId}/cancel`);
    },
    onSuccess: () => {
      showSuccess('Job cancelled successfully');
      queryClient.invalidateQueries({ queryKey: ['background-jobs'] });
    },
    onError: (error) => {
      showError(getErrorMessage(error));
    },
  });

  const getStatusBadge = (state: string) => {
    const variants: Record<string, any> = {
      pending: { variant: 'warning' as const, icon: Clock },
      active: { variant: 'default' as const, icon: Loader2 },
      completed: { variant: 'success' as const, icon: CheckCircle2 },
      failed: { variant: 'danger' as const, icon: AlertCircle },
      cancelled: { variant: 'secondary' as const, icon: XCircle },
    };

    const config = variants[state] || variants.pending;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant} className="flex items-center gap-1 w-fit">
        <Icon className={`h-3 w-3 ${state === 'active' ? 'animate-spin' : ''}`} />
        {state}
      </Badge>
    );
  };

  const stats = jobs?.stats || {
    total: 0,
    pending: 0,
    active: 0,
    completed: 0,
    failed: 0,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Background Jobs</h1>
          <p className="text-gray-500 mt-1">Monitor and manage asynchronous tasks</p>
        </div>
        <Button onClick={() => refetch()} variant="outline">
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        {[
          { label: 'Total Jobs', value: stats.total, color: 'text-gray-600', bgColor: 'bg-gray-50' },
          { label: 'Pending', value: stats.pending, color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
          { label: 'Active', value: stats.active, color: 'text-blue-600', bgColor: 'bg-blue-50' },
          { label: 'Completed', value: stats.completed, color: 'text-green-600', bgColor: 'bg-green-50' },
          { label: 'Failed', value: stats.failed, color: 'text-red-600', bgColor: 'bg-red-50' },
        ].map((stat) => (
          <Card key={stat.label}>
            <CardContent className="pt-6">
              <div className="text-sm font-medium text-gray-600">{stat.label}</div>
              <div className={`text-3xl font-bold ${stat.color} mt-2`}>{stat.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Jobs Table */}
      <Card>
        <CardHeader>
          <CardTitle>Job Queue</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading jobs...</div>
          ) : jobs?.jobs && jobs.jobs.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Queue</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Retries</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {jobs.jobs.map((job) => (
                  <TableRow key={job.id}>
                    <TableCell className="font-medium">{job.type}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{job.queue}</Badge>
                    </TableCell>
                    <TableCell>{getStatusBadge(job.state)}</TableCell>
                    <TableCell>
                      {job.retried}/{job.max_retry}
                    </TableCell>
                    <TableCell>{formatDate(job.created_at)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setViewLogsJob(job)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        {job.state === 'failed' && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => retryJob.mutate(job.id)}
                            disabled={retryJob.isPending}
                          >
                            <PlayCircle className="h-4 w-4 mr-1" />
                            Retry
                          </Button>
                        )}
                        {['pending', 'active'].includes(job.state) && (
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => cancelJob.mutate(job.id)}
                            disabled={cancelJob.isPending}
                          >
                            <XCircle className="h-4 w-4 mr-1" />
                            Cancel
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-12 text-gray-500">
              <Loader2 className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>No background jobs found</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Job Details Dialog */}
      {viewLogsJob && (
        <Dialog open={!!viewLogsJob} onOpenChange={() => setViewLogsJob(null)}>
          <DialogContent className="max-w-3xl">
            <DialogHeader>
              <DialogTitle>Job Details: {viewLogsJob.type}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <h4 className="font-semibold mb-2">Status</h4>
                {getStatusBadge(viewLogsJob.state)}
              </div>
              <div>
                <h4 className="font-semibold mb-2">Queue</h4>
                <p>{viewLogsJob.queue}</p>
              </div>
              <div>
                <h4 className="font-semibold mb-2">Payload</h4>
                <pre className="bg-gray-100 p-4 rounded-lg overflow-auto text-sm">
                  {JSON.stringify(viewLogsJob.payload, null, 2)}
                </pre>
              </div>
              {viewLogsJob.last_err && (
                <div>
                  <h4 className="font-semibold mb-2 text-red-600">Last Error</h4>
                  <pre className="bg-red-50 p-4 rounded-lg overflow-auto text-sm text-red-900">
                    {viewLogsJob.last_err}
                  </pre>
                </div>
              )}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="font-semibold mb-2">Created At</h4>
                  <p>{formatDate(viewLogsJob.created_at)}</p>
                </div>
                {viewLogsJob.processed_at && (
                  <div>
                    <h4 className="font-semibold mb-2">Processed At</h4>
                    <p>{formatDate(viewLogsJob.processed_at)}</p>
                  </div>
                )}
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
