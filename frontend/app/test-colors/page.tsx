'use client';

import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

export default function TestColorsPage() {
  return (
    <div className="min-h-screen p-8 bg-background">
      <h1 className="text-4xl font-bold text-foreground mb-8">Color Test Page</h1>
      
      <div className="space-y-8">
        {/* Test Buttons */}
        <Card className="p-6">
          <h2 className="text-2xl font-bold mb-4 text-foreground">Buttons</h2>
          <div className="flex flex-wrap gap-4">
            <Button variant="default">Primary Button</Button>
            <Button variant="destructive">Destructive Button</Button>
            <Button variant="outline">Outline Button</Button>
            <Button variant="secondary">Secondary Button</Button>
            <Button variant="ghost">Ghost Button</Button>
            <Button variant="link">Link Button</Button>
          </div>
        </Card>

        {/* Test Colors */}
        <Card className="p-6">
          <h2 className="text-2xl font-bold mb-4 text-foreground">Colors</h2>
          <div className="grid grid-cols-4 gap-4">
            <div className="space-y-2">
              <div className="h-20 bg-primary rounded flex items-center justify-center text-primary-foreground font-bold">
                Primary
              </div>
              <p className="text-sm text-muted-foreground text-center">bg-primary</p>
            </div>
            
            <div className="space-y-2">
              <div className="h-20 bg-secondary rounded flex items-center justify-center text-secondary-foreground font-bold">
                Secondary
              </div>
              <p className="text-sm text-muted-foreground text-center">bg-secondary</p>
            </div>
            
            <div className="space-y-2">
              <div className="h-20 bg-destructive rounded flex items-center justify-center text-destructive-foreground font-bold">
                Destructive
              </div>
              <p className="text-sm text-muted-foreground text-center">bg-destructive</p>
            </div>
            
            <div className="space-y-2">
              <div className="h-20 bg-muted rounded flex items-center justify-center text-muted-foreground font-bold">
                Muted
              </div>
              <p className="text-sm text-muted-foreground text-center">bg-muted</p>
            </div>
          </div>
        </Card>

        {/* Test Text Colors */}
        <Card className="p-6">
          <h2 className="text-2xl font-bold mb-4 text-foreground">Text Colors</h2>
          <div className="space-y-2">
            <p className="text-primary text-lg font-semibold">Primary Text (text-primary)</p>
            <p className="text-secondary-foreground text-lg">Secondary Foreground (text-secondary-foreground)</p>
            <p className="text-muted-foreground text-lg">Muted Foreground (text-muted-foreground)</p>
            <p className="text-destructive text-lg font-semibold">Destructive Text (text-destructive)</p>
            <p className="text-foreground text-lg">Regular Foreground (text-foreground)</p>
          </div>
        </Card>

        {/* CSS Variables Check */}
        <Card className="p-6">
          <h2 className="text-2xl font-bold mb-4 text-foreground">CSS Variables</h2>
          <div className="space-y-2 font-mono text-sm">
            <p>--color-primary: <span style={{ color: 'var(--color-primary)' }}>■ This should be green</span></p>
            <p>--color-destructive: <span style={{ color: 'var(--color-destructive)' }}>■ This should be red</span></p>
            <p>--color-background: <span style={{ backgroundColor: 'var(--color-background)', padding: '2px 8px', border: '1px solid var(--color-border)' }}>This background</span></p>
          </div>
        </Card>

        {/* Direct Tailwind Colors */}
        <Card className="p-6">
          <h2 className="text-2xl font-bold mb-4 text-foreground">Direct Tailwind Colors (for comparison)</h2>
          <div className="flex flex-wrap gap-4">
            <div className="h-20 w-32 bg-green-600 rounded flex items-center justify-center text-white font-bold">
              Green-600
            </div>
            <div className="h-20 w-32 bg-red-600 rounded flex items-center justify-center text-white font-bold">
              Red-600
            </div>
            <div className="h-20 w-32 bg-gray-200 rounded flex items-center justify-center text-gray-800 font-bold">
              Gray-200
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
