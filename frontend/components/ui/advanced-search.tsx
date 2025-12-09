'use client';

import { useState, useEffect, useCallback } from 'react';
import { Search, X, Filter, SlidersHorizontal } from 'lucide-react';
import { useDebounce } from '@/hooks/useDebounce';

export interface SearchFilter {
  key: string;
  label: string;
  type: 'text' | 'select' | 'date' | 'number' | 'boolean';
  options?: { label: string; value: string }[];
  placeholder?: string;
}

export interface AdvancedSearchProps {
  onSearch: (query: string, filters: Record<string, any>) => void;
  placeholder?: string;
  filters?: SearchFilter[];
  className?: string;
  debounceMs?: number;
}

export function AdvancedSearch({
  onSearch,
  placeholder = 'Search...',
  filters = [],
  className = '',
  debounceMs = 300,
}: AdvancedSearchProps) {
  const [query, setQuery] = useState('');
  const [showFilters, setShowFilters] = useState(false);
  const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
  const debouncedQuery = useDebounce(query, debounceMs);

  useEffect(() => {
    onSearch(debouncedQuery, activeFilters);
  }, [debouncedQuery, activeFilters, onSearch]);

  const handleFilterChange = (key: string, value: any) => {
    setActiveFilters((prev) => {
      if (value === '' || value === null || value === undefined) {
        const { [key]: _, ...rest } = prev;
        return rest;
      }
      return { ...prev, [key]: value };
    });
  };

  const clearFilters = () => {
    setActiveFilters({});
    setQuery('');
  };

  const activeFilterCount = Object.keys(activeFilters).length;

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Search Bar */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={placeholder}
            className="w-full pl-10 pr-10 py-2.5 border border-input bg-background text-foreground placeholder:text-muted-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
          />
          {query && (
            <button
              onClick={() => setQuery('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-5 w-5" />
            </button>
          )}
        </div>

        {filters.length > 0 && (
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-2 px-4 py-2 border rounded-lg transition-colors ${
              showFilters || activeFilterCount > 0
                ? 'surface-success'
                : 'border-border text-foreground hover-surface'
            }`}
          >
            <SlidersHorizontal className="h-5 w-5" />
            <span>Filters</span>
            {activeFilterCount > 0 && (
              <span className="bg-green-600 text-white text-xs px-2 py-0.5 rounded-full">
                {activeFilterCount}
              </span>
            )}
          </button>
        )}

        {(query || activeFilterCount > 0) && (
          <button
            onClick={clearFilters}
            className="px-4 py-2 text-muted-foreground hover:text-foreground hover-surface rounded-lg transition-colors"
          >
            Clear
          </button>
        )}
      </div>

      {/* Filters Panel */}
      {showFilters && filters.length > 0 && (
        <div className="surface-muted rounded-lg p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filters.map((filter) => (
              <div key={filter.key} className="space-y-1.5">
                <label className="block text-sm font-medium text-foreground">
                  {filter.label}
                </label>
                
                {filter.type === 'text' && (
                  <input
                    type="text"
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    placeholder={filter.placeholder}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground placeholder:text-muted-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  />
                )}

                {filter.type === 'number' && (
                  <input
                    type="number"
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    placeholder={filter.placeholder}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground placeholder:text-muted-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  />
                )}

                {filter.type === 'date' && (
                  <input
                    type="date"
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  />
                )}

                {filter.type === 'select' && (
                  <select
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  >
                    <option value="">All</option>
                    {filter.options?.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                )}

                {filter.type === 'boolean' && (
                  <select
                    value={activeFilters[filter.key] || ''}
                    onChange={(e) => handleFilterChange(filter.key, e.target.value)}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-lg focus:ring-2 focus:ring-green-500 focus:border-transparent"
                  >
                    <option value="">All</option>
                    <option value="true">Yes</option>
                    <option value="false">No</option>
                  </select>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
