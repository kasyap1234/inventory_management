'use client';

import { useState, useMemo, useDeferredValue, useCallback, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, Edit, Trash2, CheckSquare, DollarSign, Trash, Package, AlertTriangle, Image as ImageIcon, Upload, X } from 'lucide-react';
import { format } from 'date-fns';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DatePicker } from '@/components/ui/date-picker';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import api from '@/lib/api';
import { Product } from '@/types';
import { formatCurrency, formatDate } from '@/lib/utils';
import { VirtualizedTable } from '@/components/ui/virtualized-table';

export default function ProductsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [selectedProducts, setSelectedProducts] = useState<string[]>([]);
  const [isBulkPriceDialogOpen, setIsBulkPriceDialogOpen] = useState(false);
  const [imagesProduct, setImagesProduct] = useState<Product | null>(null);
  const queryClient = useQueryClient();

  const { data: products, isLoading } = useQuery<{ products: Product[] }>({
    queryKey: ['products', 50],
    queryFn: async () => {
      const response = await api.get('/products?limit=50');
      return response.data;
    },
    staleTime: 5 * 60 * 1000,
  });

  const deleteProduct = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/products/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });

  const deferredSearch = useDeferredValue(searchQuery);

  const filteredProducts = useMemo(() => {
    const term = deferredSearch.trim().toLowerCase();
    const source = products?.products ?? [];
    if (!term) return source;
    return source.filter((product) =>
      product.name.toLowerCase().includes(term) ||
      product.barcode?.toLowerCase().includes(term)
    );
  }, [deferredSearch, products?.products]);

  const bulkDeleteProducts = useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map(id => api.delete(`/products/${id}`)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      setSelectedProducts([]);
    },
  });

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelectedProducts(filteredProducts.map((p) => p.id));
      return;
    }
    setSelectedProducts([]);
  }, [filteredProducts]);

  const handleSelectProduct = useCallback((productId: string, checked: boolean) => {
    setSelectedProducts((prev) => {
      if (checked) {
        return prev.includes(productId) ? prev : [...prev, productId];
      }
      return prev.filter((id) => id !== productId);
    });
  }, []);

  const handleBulkDelete = () => {
    if (confirm(`Are you sure you want to delete ${selectedProducts.length} product(s)?`)) {
      bulkDeleteProducts.mutate(selectedProducts);
    }
  };

  const columns = useMemo(() => {
    const allSelected = filteredProducts.length > 0 && selectedProducts.length === filteredProducts.length;

    return [
      {
        key: 'select',
        header: (
          <input
            type="checkbox"
            aria-label="Select all products"
            checked={allSelected}
            onChange={(e) => handleSelectAll(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300"
          />
        ),
        width: '56px',
        render: (product: Product) => (
          <input
            type="checkbox"
            aria-label={`Select ${product.name}`}
            checked={selectedProducts.includes(product.id)}
            onChange={(e) => handleSelectProduct(product.id, e.target.checked)}
            className="h-4 w-4 rounded border-gray-300"
          />
        ),
      },
      {
        key: 'name',
        header: 'Name',
        width: '1.2fr',
        render: (product: Product) => (
          <div className="font-medium">{product.name}</div>
        ),
      },
      {
        key: 'barcode',
        header: 'Barcode',
        width: '1fr',
        render: (product: Product) => product.barcode || '-',
      },
      {
        key: 'quantity',
        header: 'Quantity',
        width: '0.8fr',
        render: (product: Product) => (
          <Badge variant={product.quantity < 10 ? 'destructive' : 'success'}>
            {product.quantity}
          </Badge>
        ),
      },
      {
        key: 'unit_price',
        header: 'Unit Price',
        width: '1fr',
        render: (product: Product) => formatCurrency(product.unit_price),
      },
      {
        key: 'unit',
        header: 'Unit',
        width: '0.8fr',
        render: (product: Product) => product.unit_of_measure || '-',
      },
      {
        key: 'expiry_date',
        header: 'Expiry Date',
        width: '1fr',
        render: (product: Product) => product.expiry_date ? formatDate(product.expiry_date) : '-',
      },
      {
        key: 'actions',
        header: 'Actions',
        width: '170px',
        render: (product: Product) => (
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setImagesProduct(product)}
            >
              <ImageIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditingProduct(product)}
            >
              <Edit className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                if (confirm(`Delete ${product.name}?`)) {
                  deleteProduct.mutate(product.id);
                }
              }}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ),
      },
    ];
  }, [deleteProduct, filteredProducts.length, handleSelectAll, handleSelectProduct, selectedProducts, setEditingProduct, setImagesProduct]);

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Products</h1>
          <p className="text-muted-foreground">Manage your product catalog and inventory</p>
          {selectedProducts.length > 0 && (
            <div className="inline-flex items-center gap-2 mt-2 px-3 py-1 rounded-full bg-primary/10 text-primary text-sm font-medium">
              <CheckSquare className="w-4 h-4" />
              {selectedProducts.length} selected
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          {selectedProducts.length > 0 && (
            <>
              <Button
                variant="outline"
                onClick={() => setIsBulkPriceDialogOpen(true)}
              >
                <DollarSign className="h-4 w-4 mr-2" />
                Update Prices
              </Button>
              <Button
                variant="destructive"
                onClick={handleBulkDelete}
              >
                <Trash className="h-4 w-4 mr-2" />
                Delete Selected
              </Button>
            </>
          )}
          <Button onClick={() => setIsAddDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add Product
          </Button>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Products</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{products?.products?.length || 0}</div>
            <p className="text-xs text-muted-foreground">
              +12% from last month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
            <AlertTriangle className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {products?.products?.filter(p => p.quantity < 10).length || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Items need attention
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Value</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(products?.products?.reduce((acc, p) => acc + (p.quantity * p.unit_price), 0) || 0)}
            </div>
            <p className="text-xs text-muted-foreground">
              Inventory valuation
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search products..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading products...</div>
          ) : filteredProducts.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No products found. Add your first product to get started.
            </div>
          ) : (
            <VirtualizedTable
              data={filteredProducts}
              columns={columns}
              rowHeight={68}
              overscan={8}
              height={560}
            />
          )}
        </CardContent>
      </Card>

      <ProductFormDialog
        open={isAddDialogOpen || !!editingProduct}
        onOpenChange={(open) => {
          setIsAddDialogOpen(open);
          if (!open) setEditingProduct(null);
        }}
        product={editingProduct}
      />

      <BulkPriceUpdateDialog
        open={isBulkPriceDialogOpen}
        onOpenChange={setIsBulkPriceDialogOpen}
        selectedProductIds={selectedProducts}
        onSuccess={() => setSelectedProducts([])}
      />

      <ProductImagesDialog
        open={!!imagesProduct}
        onOpenChange={(open) => !open && setImagesProduct(null)}
        product={imagesProduct}
      />
    </div>
  );
}

function ProductFormDialog({ open, onOpenChange, product }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  product?: Product | null;
}) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState({
    name: product?.name || '',
    barcode: product?.barcode || '',
    quantity: product?.quantity || 0,
    unit_price: product?.unit_price || 0,
    unit_of_measure: product?.unit_of_measure || '',
    description: product?.description || '',
    batch_number: product?.batch_number || '',
  });
  const [expiryDate, setExpiryDate] = useState<Date | null>(
    product?.expiry_date ? new Date(product.expiry_date) : null
  );

  useEffect(() => {
    if (open) {
      setFormData({
        name: product?.name || '',
        barcode: product?.barcode || '',
        quantity: product?.quantity || 0,
        unit_price: product?.unit_price || 0,
        unit_of_measure: product?.unit_of_measure || '',
        description: product?.description || '',
        batch_number: product?.batch_number || '',
      });
      setExpiryDate(product?.expiry_date ? new Date(product.expiry_date) : null);
    }
  }, [open, product]);

  const saveMutation = useMutation({
    mutationFn: async (data: typeof formData) => {
      const payload = {
        ...data,
        expiry_date: expiryDate ? format(expiryDate, 'yyyy-MM-dd') : '',
      };
      if (product) {
        await api.put(`/products/${product.id}`, payload);
      } else {
        await api.post('/products', payload);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      onOpenChange(false);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate(formData);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{product ? 'Edit Product' : 'Add Product'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Product Name *</label>
            <Input
              required
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Urea Fertilizer"
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Barcode</label>
              <Input
                value={formData.barcode}
                onChange={(e) => setFormData({ ...formData, barcode: e.target.value })}
                placeholder="BAR123"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Batch Number</label>
              <Input
                value={formData.batch_number}
                onChange={(e) => setFormData({ ...formData, batch_number: e.target.value })}
                placeholder="BATCH001"
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Quantity *</label>
              <Input
                required
                type="number"
                value={isNaN(formData.quantity) ? '' : formData.quantity}
                onChange={(e) => setFormData({ ...formData, quantity: e.target.value === '' ? 0 : parseInt(e.target.value) || 0 })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Unit Price *</label>
              <Input
                required
                type="number"
                step="0.01"
                value={isNaN(formData.unit_price) ? '' : formData.unit_price}
                onChange={(e) => setFormData({ ...formData, unit_price: e.target.value === '' ? 0 : parseFloat(e.target.value) || 0 })}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Unit of Measure</label>
              <Input
                value={formData.unit_of_measure}
                onChange={(e) => setFormData({ ...formData, unit_of_measure: e.target.value })}
                placeholder="kg, liter, bag"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Expiry Date</label>
              <DatePicker
                date={expiryDate}
                onSelect={(date) => setExpiryDate(date || null)}
                placeholder="Select expiry date"
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Input
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Product description"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={saveMutation.isPending}>
              {saveMutation.isPending ? 'Saving...' : 'Save Product'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function BulkPriceUpdateDialog({ open, onOpenChange, selectedProductIds, onSuccess }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedProductIds: string[];
  onSuccess: () => void;
}) {
  const queryClient = useQueryClient();
  const [updateType, setUpdateType] = useState<'percentage' | 'fixed'>('percentage');
  const [value, setValue] = useState<number>(0);
  const [isIncrease, setIsIncrease] = useState(true);

  const bulkUpdateMutation = useMutation({
    mutationFn: async () => {
      const adjustment = updateType === 'percentage'
        ? { type: 'percentage', value: isIncrease ? value : -value }
        : { type: 'fixed', value: isIncrease ? value : -value };

      await api.post('/products/bulk-price-update', {
        product_ids: selectedProductIds,
        adjustment,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
      onOpenChange(false);
      onSuccess();
      setValue(0);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (value === 0) {
      alert('Please enter a value');
      return;
    }
    if (confirm(`This will update ${selectedProductIds.length} product(s). Continue?`)) {
      bulkUpdateMutation.mutate();
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Bulk Price Update</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Update Type</label>
            <div className="flex gap-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  checked={updateType === 'percentage'}
                  onChange={() => setUpdateType('percentage')}
                  className="mr-2"
                />
                Percentage
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  checked={updateType === 'fixed'}
                  onChange={() => setUpdateType('fixed')}
                  className="mr-2"
                />
                Fixed Amount
              </label>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Adjustment Type</label>
            <div className="flex gap-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  checked={isIncrease}
                  onChange={() => setIsIncrease(true)}
                  className="mr-2"
                />
                Increase
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  checked={!isIncrease}
                  onChange={() => setIsIncrease(false)}
                  className="mr-2"
                />
                Decrease
              </label>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">
              {updateType === 'percentage' ? 'Percentage (%)' : 'Amount (₹)'}
            </label>
            <Input
              required
              type="number"
              step="0.01"
              min="0"
              value={isNaN(value) ? '' : value}
              onChange={(e) => setValue(e.target.value === '' ? 0 : parseFloat(e.target.value) || 0)}
              placeholder={updateType === 'percentage' ? 'e.g., 10' : 'e.g., 50'}
            />
            <p className="text-xs text-muted-foreground">
              {updateType === 'percentage'
                ? `Prices will be ${isIncrease ? 'increased' : 'decreased'} by ${value}%`
                : `Prices will be ${isIncrease ? 'increased' : 'decreased'} by ₹${value}`}
            </p>
          </div>

          <div className="bg-primary/10 p-3 rounded-lg">
            <p className="text-sm text-primary">
              <strong>{selectedProductIds.length}</strong> product(s) will be updated
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={bulkUpdateMutation.isPending}>
              {bulkUpdateMutation.isPending ? 'Updating...' : 'Update Prices'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

interface ProductImage {
  id: string;
  product_id: string;
  image_url: string;
  alt_text?: string;
  is_primary: boolean;
  created_at: string;
}

function ProductImagesDialog({ open, onOpenChange, product }: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  product: Product | null;
}) {
  const queryClient = useQueryClient();
  const [uploading, setUploading] = useState(false);

  const { data: imagesData, isLoading } = useQuery<{ images: ProductImage[] }>({
    queryKey: ['product-images', product?.id],
    queryFn: async () => {
      if (!product) return { images: [] };
      const response = await api.get(`/products/${product.id}/images`);
      return response.data;
    },
    enabled: !!product && open,
  });

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      const presignResp = await api.post(`/products/${product!.id}/images/presign`, {
        filename: file.name,
        size: file.size,
        content_type: file.type,
      });
      const presign = presignResp.data as { upload_url: string; object_key: string };

      const putResp = await fetch(presign.upload_url, {
        method: 'PUT',
        headers: {
          'Content-Type': file.type,
        },
        body: file,
      });
      if (!putResp.ok) {
        throw new Error('Failed to upload to storage');
      }

      const finalizeResp = await api.post(`/products/${product!.id}/images/finalize`, {
        object_key: presign.object_key,
        alt_text: product?.name || '',
      });
      return finalizeResp.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['product-images', product?.id] });
      setUploading(false);
    },
    onError: (error: unknown) => {
      const err = error as { response?: { data?: { message?: string } } };
      alert(err.response?.data?.message || 'Failed to upload image');
      setUploading(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (imageId: string) => {
      await api.delete(`/products/${product!.id}/images/${imageId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['product-images', product?.id] });
    },
    onError: (error: unknown) => {
      const err = error as { response?: { data?: { message?: string } } };
      alert(err.response?.data?.message || 'Failed to delete image');
    },
  });

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      alert('File size must be less than 5MB');
      return;
    }

    // Validate file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      alert('Only JPEG, PNG, GIF, and WebP images are allowed');
      return;
    }

    setUploading(true);
    uploadMutation.mutate(file);
  };

  const images = imagesData?.images || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Product Images - {product?.name}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Upload Section */}
          <div className="border-2 border-dashed border-border rounded-lg p-6 text-center">
            <input
              type="file"
              id="image-upload"
              accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
              onChange={handleFileSelect}
              className="hidden"
              disabled={uploading}
            />
            <label htmlFor="image-upload" className="cursor-pointer">
              <div className="flex flex-col items-center gap-2">
                <Upload className="h-10 w-10 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">
                  {uploading ? 'Uploading...' : 'Click to upload or drag and drop'}
                </p>
                <p className="text-xs text-muted-foreground">
                  PNG, JPG, GIF, WebP up to 5MB
                </p>
              </div>
            </label>
          </div>

          {/* Images Gallery */}
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading images...</div>
          ) : images.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <ImageIcon className="h-12 w-12 mx-auto mb-2 text-muted-foreground/50" />
              <p>No images uploaded yet</p>
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {images.map((image) => (
                <div
                  key={image.id}
                  className="relative group rounded-lg overflow-hidden border border-border hover:border-primary transition-colors"
                >
                  <img
                    src={image.image_url}
                    alt={image.alt_text || product?.name}
                    className="w-full h-48 object-cover"
                  />
                  {image.is_primary && (
                    <Badge className="absolute top-2 left-2" variant="default">
                      Primary
                    </Badge>
                  )}
                  <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => {
                        if (confirm('Delete this image?')) {
                          deleteMutation.mutate(image.id);
                        }
                      }}
                      disabled={deleteMutation.isPending}
                    >
                      <X className="h-4 w-4 mr-1" />
                      Delete
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex justify-end">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
