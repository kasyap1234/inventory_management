'use client';

import { useState } from 'react';
import { X, Filter } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';

export interface FilterConfig {
  dateRange?: {
    label: string;
    startKey: string;
    endKey: string;
  };
  statuses?: {
    label: string;
    options: { value: string; label: string }[];
  };
  categories?: {
    label: string;
    options: { value: string; label: string }[];
  };
  priceRange?: {
    label: string;
    minKey: string;
    maxKey: string;
  };
  quantityRange?: {
    label: string;
    minKey: string;
    maxKey: string;
  };
  customFilters?: Array<{
    key: string;
    label: string;
    type: 'text' | 'select' | 'number' | 'date';
    options?: { value: string; label: string }[];
  }>;
}

interface AdvancedFiltersProps {
  config: FilterConfig;
  onApply: (filters: Record<string, any>) => void;
  onReset: () => void;
  activeFilters: Record<string, any>;
}

export default function AdvancedFilters({
  config,
  onApply,
  onReset,
  activeFilters,
}: AdvancedFiltersProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [filters, setFilters] = useState<Record<string, any>>(activeFilters);

  const handleApply = () => {
    onApply(filters);
    setIsOpen(false);
  };

  const handleReset = () => {
    setFilters({});
    onReset();
    setIsOpen(false);
  };

  const activeFilterCount = Object.keys(activeFilters).filter(
    key => activeFilters[key] !== undefined && activeFilters[key] !== '' && activeFilters[key] !== null
  ).length;

  return (
    <>
      <Button
        variant="outline"
        onClick={() => setIsOpen(true)}
        className="relative"
      >
        <Filter className="h-4 w-4 mr-2" />
        Filters
        {activeFilterCount > 0 && (
          <span className="ml-2 bg-blue-600 text-white text-xs rounded-full px-2 py-0.5">
            {activeFilterCount}
          </span>
        )}
      </Button>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Advanced Filters</DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            {/* Date Range Filter */}
            {config.dateRange && (
              <div className="space-y-2">
                <label className="text-sm font-medium">{config.dateRange.label}</label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-muted-foreground">From</label>
                    <Input
                      type="date"
                      value={filters[config.dateRange.startKey] || ''}
                      onChange={(e) =>
                        setFilters({ ...filters, [config.dateRange!.startKey]: e.target.value })
                      }
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">To</label>
                    <Input
                      type="date"
                      value={filters[config.dateRange.endKey] || ''}
                      onChange={(e) =>
                        setFilters({ ...filters, [config.dateRange!.endKey]: e.target.value })
                      }
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Status Filter */}
            {config.statuses && (
              <div className="space-y-2">
                <label className="text-sm font-medium">{config.statuses.label}</label>
                <div className="grid grid-cols-2 gap-2">
                  {config.statuses.options.map((option) => (
                    <label
                      key={option.value}
                      className="flex items-center p-2 border border-border rounded hover:bg-muted cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={(filters.statuses || []).includes(option.value)}
                        onChange={(e) => {
                          const currentStatuses = filters.statuses || [];
                          const newStatuses = e.target.checked
                            ? [...currentStatuses, option.value]
                            : currentStatuses.filter((s: string) => s !== option.value);
                          setFilters({ ...filters, statuses: newStatuses });
                        }}
                        className="mr-2 h-4 w-4 rounded"
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
            )}

            {/* Category Filter */}
            {config.categories && (
              <div className="space-y-2">
                <label className="text-sm font-medium">{config.categories.label}</label>
                <select
                  value={filters.category_id || ''}
                  onChange={(e) => setFilters({ ...filters, category_id: e.target.value })}
                  className="flex h-10 w-full rounded-md border border-input bg-background text-foreground px-3 py-2 text-sm"
                >
                  <option value="">All Categories</option>
                  {config.categories.options.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Price Range Filter */}
            {config.priceRange && (
              <div className="space-y-2">
                <label className="text-sm font-medium">{config.priceRange.label}</label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-muted-foreground">Min Price</label>
                    <Input
                      type="number"
                      step="0.01"
                      placeholder="0"
                      value={filters[config.priceRange.minKey] || ''}
                      onChange={(e) =>
                        setFilters({
                          ...filters,
                          [config.priceRange!.minKey]: e.target.value,
                        })
                      }
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">Max Price</label>
                    <Input
                      type="number"
                      step="0.01"
                      placeholder="∞"
                      value={filters[config.priceRange.maxKey] || ''}
                      onChange={(e) =>
                        setFilters({
                          ...filters,
                          [config.priceRange!.maxKey]: e.target.value,
                        })
                      }
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Quantity Range Filter */}
            {config.quantityRange && (
              <div className="space-y-2">
                <label className="text-sm font-medium">{config.quantityRange.label}</label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-muted-foreground">Min Quantity</label>
                    <Input
                      type="number"
                      placeholder="0"
                      value={filters[config.quantityRange.minKey] || ''}
                      onChange={(e) =>
                        setFilters({
                          ...filters,
                          [config.quantityRange!.minKey]: e.target.value,
                        })
                      }
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">Max Quantity</label>
                    <Input
                      type="number"
                      placeholder="∞"
                      value={filters[config.quantityRange.maxKey] || ''}
                      onChange={(e) =>
                        setFilters({
                          ...filters,
                          [config.quantityRange!.maxKey]: e.target.value,
                        })
                      }
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Custom Filters */}
            {config.customFilters?.map((customFilter) => (
              <div key={customFilter.key} className="space-y-2">
                <label className="text-sm font-medium">{customFilter.label}</label>
                {customFilter.type === 'select' && customFilter.options ? (
                  <select
                    value={filters[customFilter.key] || ''}
                    onChange={(e) =>
                      setFilters({ ...filters, [customFilter.key]: e.target.value })
                    }
                    className="flex h-10 w-full rounded-md border border-input bg-background text-foreground px-3 py-2 text-sm"
                  >
                    <option value="">All</option>
                    {customFilter.options.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                ) : (
                  <Input
                    type={customFilter.type}
                    value={filters[customFilter.key] || ''}
                    onChange={(e) =>
                      setFilters({ ...filters, [customFilter.key]: e.target.value })
                    }
                  />
                )}
              </div>
            ))}
          </div>

          {/* Actions */}
          <div className="flex justify-between pt-4">
            <Button variant="outline" onClick={handleReset}>
              <X className="h-4 w-4 mr-2" />
              Clear All
            </Button>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setIsOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleApply}>Apply Filters</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

// Active Filters Display Component
export function ActiveFilterBadges({
  filters,
  onRemove,
}: {
  filters: Record<string, any>;
  onRemove: (key: string) => void;
}) {
  const filterEntries = Object.entries(filters).filter(
    ([_, value]) => value !== undefined && value !== '' && value !== null
  );

  if (filterEntries.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {filterEntries.map(([key, value]) => {
        let displayValue = value;
        
        // Format display value based on key
        if (key === 'statuses' && Array.isArray(value)) {
          displayValue = value.join(', ');
        }
        
        return (
          <div
            key={key}
            className="inline-flex items-center gap-1 px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm"
          >
            <span className="font-medium">{key}:</span>
            <span>{displayValue}</span>
            <button
              onClick={() => onRemove(key)}
              className="ml-1 hover:bg-blue-200 rounded-full p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
        );
      })}
    </div>
  );
}
