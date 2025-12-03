import { Loader2 } from 'lucide-react';

interface LoadingStateProps {
    message?: string;
    size?: 'sm' | 'md' | 'lg';
}

export function LoadingState({ message = 'Loading...', size = 'md' }: LoadingStateProps) {
    const sizeClasses = {
        sm: 'h-4 w-4',
        md: 'h-8 w-8',
        lg: 'h-12 w-12',
    };

    return (
        <div className="flex flex-col items-center justify-center py-12">
            <Loader2 className={`${sizeClasses[size]} animate-spin text-primary`} />
            <p className="mt-4 text-sm text-muted-foreground">{message}</p>
        </div>
    );
}

interface TableLoadingStateProps {
    columns: number;
    rows?: number;
}

export function TableLoadingState({ columns, rows = 5 }: TableLoadingStateProps) {
    return (
        <div className="space-y-2">
            {Array.from({ length: rows }).map((_, i) => (
                <div key={i} className="flex gap-4">
                    {Array.from({ length: columns }).map((_, j) => (
                        <div key={j} className="h-10 flex-1 animate-pulse rounded bg-muted" />
                    ))}
                </div>
            ))}
        </div>
    );
}
