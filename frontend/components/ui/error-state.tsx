import { AlertCircle, AlertTriangle, Info, CheckCircle2 } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface ErrorStateProps {
    title?: string;
    message: string;
    type?: 'error' | 'warning' | 'info' | 'success';
    action?: {
        label: string;
        onClick: () => void;
    };
}

export function ErrorState({
    title,
    message,
    type = 'error',
    action
}: ErrorStateProps) {
    const icons = {
        error: AlertCircle,
        warning: AlertTriangle,
        info: Info,
        success: CheckCircle2,
    };

    const variants = {
        error: 'destructive',
        warning: 'default',
        info: 'default',
        success: 'default',
    } as const;

    const Icon = icons[type];

    return (
        <div className="py-8">
            <Alert variant={variants[type]}>
                <Icon className="h-4 w-4" />
                {title && <AlertTitle>{title}</AlertTitle>}
                <AlertDescription className="space-y-2">
                    <p>{message}</p>
                    {action && (
                        <button
                            onClick={action.onClick}
                            className="text-sm font-medium underline underline-offset-4 hover:no-underline"
                        >
                            {action.label}
                        </button>
                    )}
                </AlertDescription>
            </Alert>
        </div>
    );
}

interface ApiErrorStateProps {
    error: unknown;
    retry?: () => void;
}

export function ApiErrorState({ error, retry }: ApiErrorStateProps) {
    const err = error as { response?: { data?: { message?: string } }; message?: string };
    const message = err.response?.data?.message || err.message || 'An unexpected error occurred';

    return (
        <ErrorState
            title="Error Loading Data"
            message={message}
            type="error"
            action={retry ? { label: 'Try Again', onClick: retry } : undefined}
        />
    );
}
