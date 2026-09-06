'use client';

import { useEffect } from 'react';
import { reportError } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { AlertCircle } from 'lucide-react';

export default function GlobalError({
    error,
    reset,
}: {
    error: Error & { digest?: string };
    reset: () => void;
}) {
    useEffect(() => {
        reportError(
            error.message,
            'frontend:global-error',
            error.stack,
            { digest: error.digest }
        );
    }, [error]);

    return (
        <html>
            <body className="font-sans antialiased text-foreground bg-background flex flex-col items-center justify-center min-h-screen">
                <div className="flex flex-col items-center justify-center p-8 text-center max-w-md">
                    <div className="bg-danger/10 p-4 rounded-full mb-6">
                        <AlertCircle className="h-12 w-12 text-danger" />
                    </div>
                    <h2 className="text-3xl font-bold mb-4">Critical Error</h2>
                    <p className="text-muted-foreground mb-8">
                        A critical error occurred that crashed the entire application. We apologize for the inconvenience.
                    </p>
                    <Button onClick={() => reset()} variant="default" className="bg-primary text-primary-foreground hover:bg-primary/90">
                        Reload Application
                    </Button>
                    {process.env.NODE_ENV === 'development' && (
                        <div className="mt-8 p-4 bg-card rounded-lg text-left w-full overflow-auto text-xs font-mono">
                            <p className="font-bold text-danger mb-2">{error.toString()}</p>
                        </div>
                    )}
                </div>
            </body>
        </html>
    );
}
