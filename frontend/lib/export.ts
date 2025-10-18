/**
 * Export utilities for CSV and Excel formats
 */

export type ExportFormat = 'csv' | 'excel';

export interface ExportColumn {
  key: string;
  label: string;
  format?: (value: any) => string;
}

export interface ExportOptions {
  filename: string;
  format: ExportFormat;
  columns: ExportColumn[];
  data: any[];
}

/**
 * Convert data to CSV format
 */
function convertToCSV(columns: ExportColumn[], data: any[]): string {
  // Create header row
  const headers = columns.map(col => `"${col.label}"`).join(',');
  
  // Create data rows
  const rows = data.map(row => {
    return columns.map(col => {
      let value = row[col.key];
      
      // Apply custom formatter if provided
      if (col.format && value !== null && value !== undefined) {
        value = col.format(value);
      }
      
      // Handle null/undefined
      if (value === null || value === undefined) {
        return '""';
      }
      
      // Convert to string and escape quotes
      const stringValue = String(value).replace(/"/g, '""');
      return `"${stringValue}"`;
    }).join(',');
  });
  
  return [headers, ...rows].join('\n');
}

/**
 * Download CSV file
 */
function downloadCSV(filename: string, csvContent: string) {
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  
  link.setAttribute('href', url);
  link.setAttribute('download', filename);
  link.style.visibility = 'hidden';
  
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  
  URL.revokeObjectURL(url);
}

/**
 * Export data to CSV
 */
export function exportToCSV(options: Omit<ExportOptions, 'format'>) {
  const csvContent = convertToCSV(options.columns, options.data);
  const filename = options.filename.endsWith('.csv') 
    ? options.filename 
    : `${options.filename}.csv`;
  
  downloadCSV(filename, csvContent);
}

/**
 * Export data to Excel (via backend API)
 */
export async function exportToExcel(
  options: Omit<ExportOptions, 'format'>,
  apiEndpoint: string,
  apiClient: any
) {
  try {
    const response = await apiClient.post(
      apiEndpoint,
      {
        filename: options.filename,
        columns: options.columns,
        data: options.data,
      },
      {
        responseType: 'blob',
      }
    );
    
    // Create download link
    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    
    const filename = options.filename.endsWith('.xlsx')
      ? options.filename
      : `${options.filename}.xlsx`;
    
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.remove();
    
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Error exporting to Excel:', error);
    throw error;
  }
}

/**
 * Generic export function that handles both CSV and Excel
 */
export async function exportData(options: ExportOptions, apiClient?: any) {
  if (options.format === 'csv') {
    exportToCSV(options);
  } else if (options.format === 'excel') {
    if (!apiClient) {
      throw new Error('API client is required for Excel export');
    }
    await exportToExcel(options, '/export/excel', apiClient);
  }
}

/**
 * Common formatters for export
 */
export const formatters = {
  date: (value: string | Date) => {
    if (!value) return '';
    const date = new Date(value);
    return date.toLocaleDateString();
  },
  
  datetime: (value: string | Date) => {
    if (!value) return '';
    const date = new Date(value);
    return date.toLocaleString();
  },
  
  currency: (value: number, currency = 'USD') => {
    if (value === null || value === undefined) return '';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
    }).format(value);
  },
  
  number: (value: number, decimals = 2) => {
    if (value === null || value === undefined) return '';
    return value.toFixed(decimals);
  },
  
  boolean: (value: boolean) => {
    return value ? 'Yes' : 'No';
  },
  
  array: (value: any[]) => {
    if (!Array.isArray(value)) return '';
    return value.join(', ');
  },
};
