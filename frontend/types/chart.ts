/**
 * Chart tooltip payload item
 */
export interface ChartPayloadItem {
  name: string;
  value: number;
  color?: string;
  dataKey?: string;
  fill?: string;
  payload?: Record<string, unknown>;
}

/**
 * Chart tooltip props
 */
export interface ChartTooltipProps {
  active?: boolean;
  payload?: ChartPayloadItem[];
  label?: string | number;
}

/**
 * Chart legend payload item
 */
export interface ChartLegendItem {
  value: string;
  type?: string;
  id?: string;
  color?: string;
  dataKey?: string;
  payload?: Record<string, unknown>;
}

/**
 * Chart legend props
 */
export interface ChartLegendProps {
  payload?: ChartLegendItem[];
}
