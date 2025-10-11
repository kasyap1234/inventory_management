'use client';

import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { FileText, User, Table, Activity, Clock, GitCompare, List } from 'lucide-react';
import {
  auditLogsService,
  type AuditLogEntry,
  type AuditSummary,
  type AuditLogListResult,
} from '@/lib/services';
import { formatDistance, format } from 'date-fns';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export default function AuditLogsPage() {
  const [filters, setFilters] = useState({
    table_name: '',
    action: '',
    page: 1,
    limit: 50,
  });
  const [viewMode, setViewMode] = useState<'list' | 'timeline'>('list');

  const { data: auditLogsData, isLoading } = useQuery<AuditLogListResult>({
    queryKey: ['audit-logs', filters],
    queryFn: () => auditLogsService.list(filters),
  });

  const { data: summary } = useQuery<AuditSummary>({
    queryKey: ['audit-logs-summary'],
    queryFn: () => auditLogsService.getSummary(),
  });

  const auditLogs: AuditLogEntry[] = auditLogsData?.items ?? [];

  const getActionColor = (action: string) => {
    switch (action?.toUpperCase()) {
      case 'INSERT':
      case 'CREATE':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'UPDATE':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      case 'DELETE':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-gray-100 text-foreground border-gray-200';
    }
  };

  const actionBreakdown = useMemo((): Array<[string, number]> => {
    if (!summary?.action_breakdown) {
      return [];
    }
    return Object.entries(summary.action_breakdown) as Array<[string, number]>;
  }, [summary?.action_breakdown]);

  const mostFrequentAction = useMemo(() => {
    if (!actionBreakdown.length) {
      return null;
    }
    return actionBreakdown.reduce((prev, curr) => (curr[1] > prev[1] ? curr : prev));
  }, [actionBreakdown]);

  const statCards = [
    {
      title: 'Total Logs',
      value: summary?.total_logs ?? 0,
      icon: Activity,
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
    },
    {
      title: 'Tables Tracked',
      value: summary?.table_breakdown ? Object.keys(summary.table_breakdown).length : 0,
      icon: Table,
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
    },
    {
      title: 'Active Users',
      value: summary?.user_activity ? Object.keys(summary.user_activity).length : 0,
      icon: User,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
    },
    {
      title: 'Top Action',
      value: mostFrequentAction ? `${mostFrequentAction[0]} (${mostFrequentAction[1]})` : 'N/A',
      icon: FileText,
      color: 'text-orange-600',
      bgColor: 'bg-orange-50',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
            Audit Logs
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Track all changes and activities in your system
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4 mr-2" />
            List View
          </Button>
          <Button
            variant={viewMode === 'timeline' ? 'default' : 'outline'}
            onClick={() => setViewMode('timeline')}
          >
            <Clock className="h-4 w-4 mr-2" />
            Timeline View
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <Card
            key={stat.title}
            className="card-hover border-0 shadow-md bg-white overflow-hidden"
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-3 pt-6">
              <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
                {stat.title}
              </CardTitle>
              <div className={`p-3 rounded-xl ${stat.bgColor} shadow-sm`}>
                <stat.icon className={`h-6 w-6 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-foreground">{stat.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <Card className="border-0 shadow-md">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold text-foreground">Filters</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Table Name
              </label>
              <Input
                placeholder="Filter by table..."
                value={filters.table_name}
                onChange={(e) => setFilters({ ...filters, table_name: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Action
              </label>
              <Input
                placeholder="Filter by action..."
                value={filters.action}
                onChange={(e) => setFilters({ ...filters, action: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Rows per page
              </label>
              <Input
                type="number"
                value={filters.limit}
                onChange={(e) => setFilters({ ...filters, limit: parseInt(e.target.value) })}
                min={10}
                max={100}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Audit Logs - Dynamic View */}
      <Card className="border-0 shadow-md">
        <CardHeader className="border-b border-gray-100">
          <CardTitle className="text-lg font-bold text-foreground">Activity Log</CardTitle>
        </CardHeader>
        <CardContent className="pt-6">
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-20 bg-gray-200 rounded-lg animate-pulse"></div>
              ))}
            </div>
          ) : auditLogs && auditLogs.length > 0 ? (
            viewMode === 'list' ? (
              <ListViewLogs logs={auditLogs} getActionColor={getActionColor} />
            ) : (
              <TimelineViewLogs logs={auditLogs} getActionColor={getActionColor} />
            )
          ) : (
            <div className="text-center py-16">
              <FileText className="h-16 w-16 text-muted-foreground/50 mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-foreground mb-2">No audit logs found</h3>
              <p className="text-muted-foreground">Audit logs will appear here when actions are performed</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

// List View Component
function ListViewLogs({ logs, getActionColor }: { logs: AuditLogEntry[]; getActionColor: (action: string) => string }) {
  return (
    <div className="space-y-3">
      {logs.map((log) => (
        <div
          key={log.id}
          className="flex items-start justify-between p-4 bg-gray-50 hover:bg-gray-100 rounded-lg border border-gray-200 transition-colors"
        >
          <div className="flex items-start space-x-4 flex-1">
            <div className="flex-shrink-0 mt-1">
              <Activity className="h-5 w-5 text-muted-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 mb-2">
                <span
                  className={`px-3 py-1 text-xs font-semibold rounded-full border ${getActionColor(
                    log.action
                  )}`}
                >
                  {log.action}
                </span>
                <span className="text-sm font-medium text-foreground">
                  {log.table_name}
                </span>
                {log.record_id && (
                  <span className="text-sm text-muted-foreground">
                    ID: {log.record_id.substring(0, 8)}...
                  </span>
                )}
              </div>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center">
                  <User className="h-4 w-4 mr-1" />
                  User: {log.changed_by ? `${log.changed_by.substring(0, 8)}...` : 'System'}
                </span>
                <span>•</span>
                <span>
                  {formatDistance(new Date(log.created_at), new Date(), {
                    addSuffix: true,
                  })}
                </span>
              </div>
              {(log.old_values || log.new_values) && (
                <DiffView oldValues={log.old_values} newValues={log.new_values} />
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

// Timeline View Component
function TimelineViewLogs({
  logs,
  getActionColor,
}: {
  logs: AuditLogEntry[];
  getActionColor: (action: string) => string;
}) {
  return (
    <div className="relative">
      {/* Timeline line */}
      <div className="absolute left-8 top-0 bottom-0 w-0.5 bg-gray-300"></div>
      
      <div className="space-y-6">
        {logs.map((log) => (
          <div key={log.id} className="relative pl-16">
            {/* Timeline dot */}
            <div className={`absolute left-6 w-5 h-5 rounded-full border-4 border-white shadow-md ${
              log.action === 'INSERT' || log.action === 'CREATE' ? 'bg-green-500' :
              log.action === 'UPDATE' ? 'bg-blue-500' :
              log.action === 'DELETE' ? 'bg-red-500' : 'bg-gray-500'
            }`}></div>
            
            <div className="bg-white rounded-lg border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow">
              <div className="flex items-start justify-between mb-2">
                <div className="flex items-center gap-3">
                  <Badge className={getActionColor(log.action)}>
                    {log.action}
                  </Badge>
                  <span className="text-sm font-semibold text-foreground">
                    {log.table_name}
                  </span>
                </div>
                <span className="text-xs text-muted-foreground">
                  {format(new Date(log.created_at), 'MMM dd, yyyy HH:mm')}
                </span>
              </div>
              
              <div className="flex items-center gap-3 text-sm text-muted-foreground mb-3">
                <span className="flex items-center">
                  <User className="h-4 w-4 mr-1" />
                  {log.changed_by ? `${log.changed_by.substring(0, 8)}...` : 'System'}
                </span>
                {log.record_id && (
                  <>
                    <span>•</span>
                    <span>Record: {log.record_id.substring(0, 8)}...</span>
                  </>
                )}
              </div>
              
              {(log.old_values || log.new_values) && (
                <DiffView oldValues={log.old_values} newValues={log.new_values} />
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Diff View Component
function DiffView({
  oldValues,
  newValues,
}: {
  oldValues: Record<string, unknown> | null | undefined;
  newValues: Record<string, unknown> | null | undefined;
}) {
  const [isExpanded, setIsExpanded] = useState(false);
  
  if (!oldValues && !newValues) return null;

  const changes = getChangedFields(
    (oldValues ?? {}) as Record<string, unknown>,
    (newValues ?? {}) as Record<string, unknown>
  );
  
  if (changes.length === 0) return null;
  
  return (
    <div className="mt-3">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700 font-medium"
      >
        <GitCompare className="h-4 w-4" />
        {isExpanded ? 'Hide' : 'View'} {changes.length} change(s)
      </button>
      
      {isExpanded && (
        <div className="mt-3 space-y-2">
          {changes.map((change, index) => (
            <div key={index} className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-foreground">{change.field}</span>
                <span className="text-xs text-muted-foreground uppercase">{change.type}</span>
              </div>
              <div className="grid grid-cols-2 gap-4 text-sm">
                {change.oldValue !== undefined && (
                  <div>
                    <div className="text-xs text-muted-foreground mb-1">Old:</div>
                    <div className="bg-red-50 border border-red-200 rounded px-2 py-1 text-red-900 font-mono text-xs break-all">
                      {formatValue(change.oldValue)}
                    </div>
                  </div>
                )}
                {change.newValue !== undefined && (
                  <div>
                    <div className="text-xs text-muted-foreground mb-1">New:</div>
                    <div className="bg-green-50 border border-green-200 rounded px-2 py-1 text-green-900 font-mono text-xs break-all">
                      {formatValue(change.newValue)}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

type AuditValue = unknown;

type ChangeDescriptor = {
  field: string;
  oldValue: AuditValue;
  newValue: AuditValue;
  type: 'added' | 'removed' | 'modified';
};

function getChangedFields(
  oldValues: Record<string, AuditValue>,
  newValues: Record<string, AuditValue>
): ChangeDescriptor[] {
  const changes: ChangeDescriptor[] = [];
  const allKeys = new Set([...Object.keys(oldValues), ...Object.keys(newValues)]);

  allKeys.forEach((key) => {
    const oldVal = oldValues[key];
    const newVal = newValues[key];

    if (oldVal !== newVal) {
      const changeType: ChangeDescriptor['type'] =
        oldVal === undefined ? 'added' : newVal === undefined ? 'removed' : 'modified';
      changes.push({
        field: key,
        oldValue: oldVal,
        newValue: newVal,
        type: changeType,
      });
    }
  });

  return changes;
}

function formatValue(value: AuditValue): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'object') return JSON.stringify(value, null, 2);
  return String(value);
}
