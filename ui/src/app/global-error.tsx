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
            <body className="font-sans antialiased text-white bg-zinc-950 flex flex-col items-center justify-center min-h-screen">
                <div className="flex flex-col items-center justify-center p-8 text-center max-w-md">
                    <div className="bg-red-500/10 p-4 rounded-full mb-6">
                        <AlertCircle className="h-12 w-12 text-red-500" />
                    </div>
                    <h2 className="text-3xl font-bold mb-4">Critical Error</h2>
                    <p className="text-zinc-400 mb-8">
                        A critical error occurred that crashed the entire application. We apologize for the inconvenience.
                    </p>
                    <Button onClick={() => reset()} variant="default" className="bg-violet-600 hover:bg-violet-700">
                        Reload Application
                    </Button>
                    {process.env.NODE_ENV === 'development' && (
                        <div className="mt-8 p-4 bg-zinc-900 rounded-lg text-left w-full overflow-auto text-xs font-mono">
                            <p className="font-bold text-red-400 mb-2">{error.toString()}</p>
                        </div>
                    )}
                </div>
            </body>
        </html>
    );
}
