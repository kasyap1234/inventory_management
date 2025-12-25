'use client';

import React, { useCallback, useMemo } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';

interface Column<T> {
  key: string;
  header: React.ReactNode;
  render?: (item: T) => React.ReactNode;
  width?: string;
}

interface VirtualizedTableProps<T> {
  data: T[];
  columns: Column<T>[];
  rowHeight?: number;
  overscan?: number;
  onRowClick?: (item: T) => void;
  className?: string;
  height?: number | string;
}

// Memoized row component to prevent re-renders during scroll
const VirtualRow = React.memo(function VirtualRow<T extends Record<string, any>>({
  item,
  columns,
  gridTemplateColumns,
  rowHeight,
  translateY,
  onRowClick,
}: {
  item: T;
  columns: Column<T>[];
  gridTemplateColumns: string;
  rowHeight: number;
  translateY: number;
  onRowClick?: (item: T) => void;
}) {
  return (
    <div
      className={`grid grid-cols-1 border-b border-border hover-surface ${onRowClick ? 'cursor-pointer' : ''}`}
      style={{
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: `${rowHeight}px`,
        transform: `translateY(${translateY}px)`,
        gridTemplateColumns,
      }}
      onClick={() => onRowClick?.(item)}
    >
      {columns.map((column) => (
        <div
          key={column.key}
          className="px-4 py-3 text-sm text-foreground overflow-hidden text-ellipsis"
        >
          {column.render ? column.render(item) : item[column.key]}
        </div>
      ))}
    </div>
  );
}) as <T extends Record<string, any>>(props: {
  item: T;
  columns: Column<T>[];
  gridTemplateColumns: string;
  rowHeight: number;
  translateY: number;
  onRowClick?: (item: T) => void;
}) => React.ReactElement;

export function VirtualizedTable<T extends Record<string, any>>({
  data,
  columns,
  rowHeight = 48,
  overscan = 5,
  onRowClick,
  className = '',
  height,
}: VirtualizedTableProps<T>) {
  const parentRef = React.useRef<HTMLDivElement>(null);

  // Memoize column template to prevent recalculation
  const gridTemplateColumns = useMemo(
    () => columns.map(col => col.width || '1fr').join(' '),
    [columns]
  );

  const rowVirtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: useCallback(() => rowHeight, [rowHeight]),
    overscan,
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  return (
    <div
      className={`overflow-auto ${className}`}
      ref={parentRef}
      style={{ height: height ?? '600px' }}
    >
      <div className="min-w-full">
        <div className="grid grid-cols-1 bg-muted text-muted-foreground border-b border-border sticky top-0 z-10" style={{
          gridTemplateColumns,
        }}>
          {columns.map((column) => (
            <div
              key={column.key}
              className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider"
            >
              {column.header}
            </div>
          ))}
        </div>

        <div
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {virtualItems.map((virtualRow) => {
            const item = data[virtualRow.index];
            return (
              <VirtualRow
                key={virtualRow.key}
                item={item}
                columns={columns}
                gridTemplateColumns={gridTemplateColumns}
                rowHeight={virtualRow.size}
                translateY={virtualRow.start}
                onRowClick={onRowClick}
              />
            );
          })}
        </div>
      </div>
    </div>
  );
}
