'use client';

import React, { useRef, useState, useCallback, useEffect } from 'react';
import Webcam from 'react-webcam';
import jsQR from 'jsqr';
import { X, Camera, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Select } from '@/components/ui/select';

interface BarcodeScannerProps {
    isOpen: boolean;
    onClose: () => void;
    onScan: (code: string) => void;
}

export function BarcodeScanner({ isOpen, onClose, onScan }: BarcodeScannerProps) {
    const webcamRef = useRef<Webcam>(null);
    const [isScanning, setIsScanning] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [devices, setDevices] = useState<MediaDeviceInfo[]>([]);
    const [selectedDeviceId, setSelectedDeviceId] = useState<string | undefined>(undefined);

    const handleDevices = useCallback(
        (mediaDevices: MediaDeviceInfo[]) => {
            const videoDevices = mediaDevices.filter(({ kind }) => kind === 'videoinput');
            setDevices(videoDevices);
            // Prefer back camera if available
            const backCamera = videoDevices.find(device =>
                device.label.toLowerCase().includes('back') ||
                device.label.toLowerCase().includes('environment')
            );
            if (backCamera) {
                setSelectedDeviceId(backCamera.deviceId);
            } else if (videoDevices.length > 0) {
                setSelectedDeviceId(videoDevices[0].deviceId);
            }
        },
        []
    );

    useEffect(() => {
        navigator.mediaDevices.enumerateDevices().then(handleDevices);
    }, [handleDevices]);

    const capture = useCallback(() => {
        if (!webcamRef.current || !isScanning) return;

        const imageSrc = webcamRef.current.getScreenshot();
        if (imageSrc) {
            const image = new Image();
            image.src = imageSrc;
            image.onload = () => {
                const canvas = document.createElement('canvas');
                canvas.width = image.width;
                canvas.height = image.height;
                const context = canvas.getContext('2d');
                if (context) {
                    context.drawImage(image, 0, 0);
                    const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
                    const code = jsQR(imageData.data, imageData.width, imageData.height);

                    if (code) {
                        setIsScanning(false);
                        // Play beep sound
                        const audio = new Audio('/sounds/beep.mp3'); // We might need to add this file or use a data URI
                        audio.play().catch(() => { }); // Ignore auto-play errors
                        onScan(code.data);
                    }
                }
            };
        }
    }, [isScanning, onScan]);

    useEffect(() => {
        let interval: NodeJS.Timeout;
        if (isOpen && isScanning) {
            interval = setInterval(capture, 500); // Scan every 500ms
        }
        return () => clearInterval(interval);
    }, [isOpen, isScanning, capture]);

    // Reset scanning state when dialog opens
    useEffect(() => {
        if (isOpen) {
            setIsScanning(true);
            setError(null);
        }
    }, [isOpen]);

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <Camera className="h-5 w-5" />
                        Scan Barcode
                    </DialogTitle>
                </DialogHeader>

                <div className="relative aspect-video bg-black rounded-lg overflow-hidden">
                    {error ? (
                        <div className="absolute inset-0 flex items-center justify-center text-white p-4 text-center">
                            <p>{error}</p>
                        </div>
                    ) : (
                        <>
                            <Webcam
                                audio={false}
                                ref={webcamRef}
                                screenshotFormat="image/jpeg"
                                videoConstraints={{
                                    deviceId: selectedDeviceId,
                                    facingMode: selectedDeviceId ? undefined : 'environment'
                                }}
                                className="w-full h-full object-cover"
                                onUserMediaError={(err) => setError('Could not access camera. Please ensure permissions are granted.')}
                            />
                            <div className="absolute inset-0 border-2 border-white/50 m-8 rounded-lg pointer-events-none">
                                <div className="absolute top-0 left-0 w-4 h-4 border-t-2 border-l-2 border-primary"></div>
                                <div className="absolute top-0 right-0 w-4 h-4 border-t-2 border-r-2 border-primary"></div>
                                <div className="absolute bottom-0 left-0 w-4 h-4 border-b-2 border-l-2 border-primary"></div>
                                <div className="absolute bottom-0 right-0 w-4 h-4 border-b-2 border-r-2 border-primary"></div>
                            </div>
                            {/* Scanning line animation */}
                            <div className="absolute inset-x-8 top-8 h-0.5 bg-primary/80 shadow-[0_0_8px_rgba(var(--primary),0.8)] animate-scan"></div>
                        </>
                    )}
                </div>

                <div className="flex items-center justify-between gap-4 mt-4">
                    {devices.length > 1 && (
                        <select
                            value={selectedDeviceId}
                            onChange={(e) => setSelectedDeviceId(e.target.value)}
                            className="w-[200px]"
                        >
                            <option value="" disabled>Select Camera</option>
                            {devices.map((device) => (
                                <option key={device.deviceId} value={device.deviceId}>
                                    {device.label || `Camera ${devices.indexOf(device) + 1}`}
                                </option>
                            ))}
                        </select>
                    )}

                    <Button
                        variant="outline"
                        size="icon"
                        onClick={() => setIsScanning(true)}
                        title="Reset Scanner"
                    >
                        <RefreshCw className={`h-4 w-4 ${isScanning ? 'animate-spin' : ''}`} />
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    );
}
