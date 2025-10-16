'use client';

import { useRef, useState } from 'react';
import { Upload, X, Image as ImageIcon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/toast';
import { useProducts } from '@/hooks/useProducts';

interface ProductImageUploadProps {
  productId: string;
  onUploadComplete?: () => void;
}

export function ProductImageUpload({ productId, onUploadComplete }: ProductImageUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const { uploadProductImage } = useProducts();
  const { addToast } = useToast();

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 5 * 1024 * 1024) {
        addToast('File size must be less than 5MB', 'error');
        return;
      }

      if (!file.type.startsWith('image/')) {
        addToast('Please select an image file', 'error');
        return;
      }

      setSelectedFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    try {
      await uploadProductImage.mutateAsync({ productId, file: selectedFile });
      addToast('Image uploaded successfully', 'success');
      setSelectedFile(null);
      setPreview(null);
      onUploadComplete?.();
    } catch (error) {
      addToast('Failed to upload image', 'error');
    }
  };

  const handleClear = () => {
    setSelectedFile(null);
    setPreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileSelect}
          className="hidden"
        />
        <Button
          type="button"
          variant="outline"
          onClick={() => fileInputRef.current?.click()}
        >
          <Upload className="h-4 w-4 mr-2" />
          Select Image
        </Button>
        {selectedFile && (
          <>
            <Button
              type="button"
              onClick={handleUpload}
              disabled={uploadProductImage.isPending}
            >
              {uploadProductImage.isPending ? 'Uploading...' : 'Upload'}
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={handleClear}
            >
              <X className="h-4 w-4" />
            </Button>
          </>
        )}
      </div>

      {preview ? (
        <div className="relative w-full h-48 border border-gray-200 rounded-lg overflow-hidden">
          <img
            src={preview}
            alt="Preview"
            className="w-full h-full object-cover"
          />
        </div>
      ) : (
        <div className="w-full h-48 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center">
          <div className="text-center">
            <ImageIcon className="h-12 w-12 mx-auto text-gray-400 mb-2" />
            <p className="text-sm text-gray-500">No image selected</p>
          </div>
        </div>
      )}

      <p className="text-xs text-gray-500">
        Supported formats: JPG, PNG, GIF. Max size: 5MB
      </p>
    </div>
  );
}
