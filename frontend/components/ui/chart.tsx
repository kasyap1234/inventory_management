'use client';

import * as React from 'react';
import type { LegendProps, TooltipProps } from 'recharts';
import { Legend as RechartsLegend, Tooltip as RechartsTooltip } from 'recharts';

import { cn } from '@/lib/utils';
import type { ChartLegendItem } from '@/types/chart';

const DEFAULT_COLORS = [
  'hsl(217 91% 60%)',
  'hsl(142 71% 45%)',
  'hsl(24 95% 53%)',
  'hsl(276 70% 60%)',
  'hsl(199 89% 48%)',
  'hsl(346 77% 57%)',
  'hsl(46 90% 55%)',
  'hsl(189 88% 45%)',
];

export type ChartConfig<TDataKey extends string = string> = Record<
  TDataKey,
  {
    label?: string;
    color?: string;
  }
>;

const ChartContext = React.createContext<ChartConfig<string> | null>(null);

type ChartContainerProps = React.HTMLAttributes<HTMLDivElement> & {
  config: ChartConfig;
};

export const ChartContainer = React.forwardRef<HTMLDivElement, ChartContainerProps>(
  ({ config, className, style, children, ...props }, ref) => {
    const cssVariables = React.useMemo(() => {
      const entries = Object.entries(config ?? {});
      if (!entries.length) {
        return {};
      }

      const styles: React.CSSProperties = {};
      entries.forEach(([key, value], index) => {
        const color = value?.color ?? DEFAULT_COLORS[index % DEFAULT_COLORS.length];
        (styles as Record<string, string>)[`--color-${key}`] = color;
        if (value?.label) {
          (styles as Record<string, string>)[`--label-${key}`] = value.label;
        }
      });

      return styles;
    }, [config]);

    return (
      <ChartContext.Provider value={config}>
        <div
          ref={ref}
          className={cn('relative w-full', className)}
          style={{ ...cssVariables, ...style }}
          {...props}
        >
          {children}
        </div>
      </ChartContext.Provider>
    );
  }
);
ChartContainer.displayName = 'ChartContainer';

export type ChartTooltipContentFormatter = (args: {
  value: any;
  name: string;
  item: any;
  index: number;
}) => React.ReactNode;

export interface ChartTooltipContentProps {
  active?: boolean;
  payload?: any[];
  label?: any;
  indicator?: 'dot' | 'line';
  className?: string;
  formatter?: ChartTooltipContentFormatter;
  labelFormatter?: (label?: any) => React.ReactNode;
}

export function ChartTooltipContent({
  active,
  payload,
  label,
  indicator = 'dot',
  className,
  formatter,
  labelFormatter,
}: ChartTooltipContentProps) {
  const config = React.useContext(ChartContext);

  if (!active || !payload?.length) {
    return null;
  }

  const renderedLabel = labelFormatter ? labelFormatter(label) : label;

  return (
    <div className={cn('rounded-lg border bg-background p-3 shadow-sm', className)}>
      {renderedLabel ? (
        <div className="mb-2 text-sm font-semibold text-foreground">{renderedLabel}</div>
      ) : null}
      <div className="grid gap-1 text-sm">
        {payload.map((item, index) => {
          const key = (item.dataKey ?? item.name ?? index).toString();
          const color = item.color ?? `var(--color-${key})`;
          const displayLabel = config?.[key]?.label ?? item.name ?? key;

          let valueNode: React.ReactNode = item.value;
          if (formatter) {
            valueNode = formatter({
              value: item.value,
              name: key,
              item,
              index,
            });
          }

          if (Array.isArray(valueNode)) {
            valueNode = valueNode[0];
          }

          return (
            <div key={key} className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    indicator === 'line' ? 'h-0.5 w-6 rounded-full' : 'h-2 w-2 rounded-full'
                  )}
                  style={{ backgroundColor: color as string }}
                />
                <span className="text-muted-foreground">{displayLabel}</span>
              </div>
              <span className="font-medium text-foreground">{valueNode}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export const ChartTooltip = RechartsTooltip;

export const ChartLegend = RechartsLegend;

export interface ChartLegendContentProps {
  className?: string;
  payload?: ChartLegendItem[];
}

export function ChartLegendContent({ payload, className }: ChartLegendContentProps) {
  const config = React.useContext(ChartContext);

  if (!payload?.length) {
    return null;
  }

  return (
    <div className={cn('flex flex-wrap items-center gap-3 text-xs text-muted-foreground', className)}>
      {payload.map((item) => {
        const key = (item.dataKey ?? item.value ?? '').toString();
        const label = config?.[key]?.label ?? item.value ?? key;

        return (
          <div key={key} className="flex items-center gap-2">
            <span
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: item.color ?? `var(--color-${key})` }}
            />
            <span className="capitalize">{label}</span>
          </div>
        );
      })}
    </div>
  );
}
