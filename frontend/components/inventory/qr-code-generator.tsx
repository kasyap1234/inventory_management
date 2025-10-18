'use client';

import { useEffect, useRef } from 'react';
import { Download } from 'lucide-react';
import QRCode from 'qrcode';

interface QRCodeGeneratorProps {
  data: string;
  filename?: string;
  title?: string;
  subtitle?: string;
  size?: number;
}

export function QRCodeGenerator({
  data,
  filename = 'qrcode',
  title,
  subtitle,
  size = 256,
}: QRCodeGeneratorProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (canvasRef.current) {
      QRCode.toCanvas(
        canvasRef.current,
        data,
        {
          width: size,
          margin: 2,
          color: {
            dark: '#000000',
            light: '#FFFFFF',
          },
        },
        (error: Error | null | undefined) => {
          if (error) console.error('QR Code generation error:', error);
        }
      );
    }
  }, [data, size]);

  const handleDownload = () => {
    if (canvasRef.current) {
      const url = canvasRef.current.toDataURL('image/png');
      const link = document.createElement('a');
      link.download = `${filename}.png`;
      link.href = url;
      link.click();
    }
  };

  const handlePrint = () => {
    const printWindow = window.open('', '_blank');
    if (printWindow && canvasRef.current) {
      const imageUrl = canvasRef.current.toDataURL('image/png');
      printWindow.document.write(`
        <html>
          <head>
            <title>Print QR Code</title>
            <style>
              body {
                display: flex;
                flex-direction: column;
                align-items: center;
                justify-content: center;
                min-height: 100vh;
                margin: 0;
                font-family: Arial, sans-serif;
              }
              .container {
                text-align: center;
                padding: 20px;
              }
              h1 { margin: 10px 0; font-size: 24px; }
              h2 { margin: 5px 0; font-size: 18px; color: #666; }
              img { margin: 20px 0; }
              @media print {
                body { margin: 0; }
              }
            </style>
          </head>
          <body>
            <div class="container">
              ${title ? `<h1>${title}</h1>` : ''}
              ${subtitle ? `<h2>${subtitle}</h2>` : ''}
              <img src="${imageUrl}" alt="QR Code" />
            </div>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.focus();
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 250);
    }
  };

  return (
    <div className="flex flex-col items-center space-y-4">
      {title && <h3 className="text-lg font-semibold text-gray-900">{title}</h3>}
      {subtitle && <p className="text-sm text-gray-600">{subtitle}</p>}
      
      <div className="bg-white p-4 rounded-lg border-2 border-gray-200">
        <canvas ref={canvasRef} />
      </div>

      <div className="flex gap-3">
        <button
          onClick={handleDownload}
          className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          <Download className="h-4 w-4" />
          Download
        </button>
        <button
          onClick={handlePrint}
          className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
        >
          Print
        </button>
      </div>
    </div>
  );
}
