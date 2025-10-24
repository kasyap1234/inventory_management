'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Clock, User, FileText } from 'lucide-react';
import { format } from 'date-fns';

interface TimelineEvent {
  id: string;
  timestamp: string;
  action: string;
  entity: string;
  record_id: string;
  changed_by?: string;
  summary: string;
  details?: Record<string, any>;
}

interface TimelineGroup {
  period: string;
  events: TimelineEvent[];
}

interface AuditTimelineProps {
  timeline: TimelineGroup[];
  isLoading?: boolean;
}

const getActionColor = (action: string) => {
  switch (action.toLowerCase()) {
    case 'insert':
    case 'create':
      return 'bg-green-100 text-green-800';
    case 'update':
      return 'bg-blue-100 text-blue-800';
    case 'delete':
      return 'bg-red-100 text-red-800';
    case 'soft_delete':
      return 'bg-yellow-100 text-yellow-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

export function AuditTimeline({ timeline, isLoading }: AuditTimelineProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Audit Timeline</CardTitle>
          <CardDescription>Loading audit history...</CardDescription>
        </CardHeader>
        <CardContent className="h-96 flex items-center justify-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  if (!timeline || timeline.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Audit Timeline</CardTitle>
          <CardDescription>No audit events found</CardDescription>
        </CardHeader>
        <CardContent className="h-96 flex items-center justify-center text-gray-500">
          No audit history available
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit Timeline</CardTitle>
        <CardDescription>Chronological view of all system activities</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-8">
          {timeline.map((group, groupIdx) => (
            <div key={groupIdx}>
              <h3 className="text-sm font-semibold text-gray-700 mb-4">{group.period}</h3>
              <div className="space-y-4">
                {group.events.map((event, eventIdx) => (
                  <div key={eventIdx} className="flex gap-4">
                    <div className="flex flex-col items-center">
                      <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white">
                        <Clock className="w-4 h-4" />
                      </div>
                      {eventIdx < group.events.length - 1 && (
                        <div className="w-0.5 h-12 bg-gray-300 my-2" />
                      )}
                    </div>
                    <div className="flex-1 pb-4">
                      <div className="flex items-start justify-between">
                        <div>
                          <p className="font-medium text-gray-900">{event.summary}</p>
                          <p className="text-sm text-gray-600 mt-1">
                            {event.entity} • {event.record_id}
                          </p>
                        </div>
                        <Badge className={getActionColor(event.action)}>
                          {event.action}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-4 mt-3 text-xs text-gray-500">
                        <div className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          {format(new Date(event.timestamp), 'MMM d, yyyy HH:mm:ss')}
                        </div>
                        {event.changed_by && (
                          <div className="flex items-center gap-1">
                            <User className="w-3 h-3" />
                            {event.changed_by}
                          </div>
                        )}
                      </div>
                      {event.details && Object.keys(event.details).length > 0 && (
                        <details className="mt-2">
                          <summary className="cursor-pointer text-xs text-blue-600 hover:text-blue-800">
                            View details
                          </summary>
                          <div className="mt-2 p-2 bg-gray-50 rounded text-xs font-mono text-gray-700 overflow-auto max-h-32">
                            <pre>{JSON.stringify(event.details, null, 2)}</pre>
                          </div>
                        </details>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
