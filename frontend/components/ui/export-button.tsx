'use client';

import { useState } from 'react';
import { Download, FileSpreadsheet, FileText } from 'lucide-react';
import { exportData, type ExportColumn, type ExportFormat } from '@/lib/export';
import { api } from '@/lib/api';
import toast from 'react-hot-toast';

interface ExportButtonProps {
  data: any[];
  columns: ExportColumn[];
  filename: string;
  label?: string;
  className?: string;
  showFormatSelector?: boolean;
}

export function ExportButton({
  data,
  columns,
  filename,
  label = 'Export',
  className = '',
  showFormatSelector = true,
}: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false);
  const [showMenu, setShowMenu] = useState(false);

  const handleExport = async (format: ExportFormat) => {
    setIsExporting(true);
    setShowMenu(false);

    try {
      await exportData(
        {
          filename,
          format,
          columns,
          data,
        },
        api
      );
      
      toast.success(`Exported ${data.length} records as ${format.toUpperCase()}`);
    } catch (error) {
      console.error('Export error:', error);
      toast.error(`Failed to export data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsExporting(false);
    }
  };

  if (!showFormatSelector) {
    return (
      <button
        onClick={() => handleExport('csv')}
        disabled={isExporting || data.length === 0}
        className={`inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors ${className}`}
      >
        <Download className="h-4 w-4" />
        {isExporting ? 'Exporting...' : label}
      </button>
    );
  }

  return (
    <div className="relative">
      <button
        onClick={() => setShowMenu(!showMenu)}
        disabled={isExporting || data.length === 0}
        className={`inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors ${className}`}
      >
        <Download className="h-4 w-4" />
        {isExporting ? 'Exporting...' : label}
      </button>

      {showMenu && !isExporting && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-10"
            onClick={() => setShowMenu(false)}
          />
          
          {/* Menu */}
          <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-20">
            <button
              onClick={() => handleExport('csv')}
              className="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <FileText className="h-4 w-4 text-gray-500" />
              <span>Export as CSV</span>
            </button>
            <button
              onClick={() => handleExport('excel')}
              className="w-full flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <FileSpreadsheet className="h-4 w-4 text-green-600" />
              <span>Export as Excel</span>
            </button>
          </div>
        </>
      )}
    </div>
  );
}
