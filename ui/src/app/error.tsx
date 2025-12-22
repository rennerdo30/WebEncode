'use client';

import { useEffect } from 'react';
import { reportError } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { AlertCircle } from 'lucide-react';

export default function Error({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    useEffect(() => {
        // Log the error to our backend
        reportError(
            error.message,
            'frontend:nextjs-error',
            error.stack,
            { digest: error.digest }
        );
    }, [error]);

    return (
        <div className="flex flex-col items-center justify-center min-h-[50vh] p-8 text-center bg-background/50 backdrop-blur-sm rounded-xl border border-border mx-auto max-w-2xl mt-12">
            <div className="bg-red-500/10 p-4 rounded-full mb-6">
                <AlertCircle className="h-10 w-10 text-red-500" />
            </div>
            <h2 className="text-2xl font-bold mb-2">Something went wrong!</h2>
            <p className="text-muted-foreground mb-8 max-w-md">
                We've logged this error and notified our team. You can try refreshing the page or attempting the action again.
            </p>
            <div className="flex gap-4">
                <Button onClick={() => window.location.reload()} variant="outline">
                    Refresh Page
                </Button>
                <Button onClick={() => reset()}>Try Again</Button>
            </div>
            {process.env.NODE_ENV === 'development' && (
                <div className="mt-8 p-4 bg-muted/50 rounded-lg text-left w-full overflow-auto text-xs font-mono max-h-48">
                    <p className="font-bold text-red-400 mb-2">{error.toString()}</p>
                    <pre>{error.stack}</pre>
                </div>
            )}
        </div>
    );
}
