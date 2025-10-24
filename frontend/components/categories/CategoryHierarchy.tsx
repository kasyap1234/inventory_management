'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ChevronDown, ChevronRight, Plus, Edit2, Trash2 } from 'lucide-react';

interface Category {
  id: string;
  name: string;
  description?: string;
  level: number;
  subcategories?: Category[];
}

interface CategoryHierarchyProps {
  categories: Category[];
  isLoading?: boolean;
  onEdit?: (category: Category) => void;
  onDelete?: (categoryId: string) => void;
  onAddSubcategory?: (parentId: string) => void;
}

interface TreeNodeProps {
  category: Category;
  onEdit?: (category: Category) => void;
  onDelete?: (categoryId: string) => void;
  onAddSubcategory?: (parentId: string) => void;
}

function CategoryTreeNode({ category, onEdit, onDelete, onAddSubcategory }: TreeNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const hasSubcategories = category.subcategories && category.subcategories.length > 0;

  return (
    <div className="space-y-1">
      <div className="flex items-center gap-2 p-2 hover:bg-gray-100 rounded">
        {hasSubcategories ? (
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-0 hover:bg-gray-200 rounded"
          >
            {isExpanded ? (
              <ChevronDown className="w-4 h-4" />
            ) : (
              <ChevronRight className="w-4 h-4" />
            )}
          </button>
        ) : (
          <div className="w-4" />
        )}
        <div className="flex-1">
          <p className="font-medium text-sm">{category.name}</p>
          {category.description && (
            <p className="text-xs text-gray-600">{category.description}</p>
          )}
        </div>
        <div className="flex items-center gap-1">
          {onAddSubcategory && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onAddSubcategory(category.id)}
              className="h-6 w-6 p-0"
            >
              <Plus className="w-3 h-3" />
            </Button>
          )}
          {onEdit && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onEdit(category)}
              className="h-6 w-6 p-0"
            >
              <Edit2 className="w-3 h-3" />
            </Button>
          )}
          {onDelete && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onDelete(category.id)}
              className="h-6 w-6 p-0 text-red-600 hover:text-red-700"
            >
              <Trash2 className="w-3 h-3" />
            </Button>
          )}
        </div>
      </div>
      {isExpanded && hasSubcategories && (
        <div className="ml-4 border-l border-gray-300 pl-2 space-y-1">
          {category.subcategories!.map((subcat) => (
            <CategoryTreeNode
              key={subcat.id}
              category={subcat}
              onEdit={onEdit}
              onDelete={onDelete}
              onAddSubcategory={onAddSubcategory}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function CategoryHierarchy({
  categories,
  isLoading,
  onEdit,
  onDelete,
  onAddSubcategory,
}: CategoryHierarchyProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Category Hierarchy</CardTitle>
          <CardDescription>Loading categories...</CardDescription>
        </CardHeader>
        <CardContent className="h-96 flex items-center justify-center">
          <div className="animate-pulse">Loading...</div>
        </CardContent>
      </Card>
    );
  }

  if (!categories || categories.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Category Hierarchy</CardTitle>
          <CardDescription>No categories found</CardDescription>
        </CardHeader>
        <CardContent className="h-96 flex items-center justify-center text-gray-500">
          No categories available
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Category Hierarchy</CardTitle>
        <CardDescription>View and manage product categories</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-1">
          {categories.map((category) => (
            <CategoryTreeNode
              key={category.id}
              category={category}
              onEdit={onEdit}
              onDelete={onDelete}
              onAddSubcategory={onAddSubcategory}
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
