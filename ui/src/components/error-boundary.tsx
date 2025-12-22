'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { reportError } from '@/lib/api';
import { AlertCircle, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
    source?: string;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        reportError(
            error.message,
            this.props.source || 'frontend:error-boundary',
            error.stack,
            { componentStack: errorInfo.componentStack }
        );
    }

    handleReset = () => {
        this.setState({ hasError: false, error: null });
    };

    render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div className="p-6 border border-red-200 bg-red-50/10 rounded-lg flex flex-col items-start gap-4 my-4">
                    <div className="flex items-center gap-3 text-red-500">
                        <AlertCircle className="h-5 w-5" />
                        <h3 className="font-medium">Something went wrong in this component</h3>
                    </div>
                    {process.env.NODE_ENV === 'development' && this.state.error && (
                        <pre className="text-xs bg-black/50 p-4 rounded overflow-auto w-full max-h-40 font-mono text-red-300">
                            {this.state.error.toString()}
                        </pre>
                    )}
                    <Button variant="outline" size="sm" onClick={this.handleReset} className="gap-2">
                        <RefreshCw className="h-4 w-4" />
                        Try Again
                    </Button>
                </div>
            );
        }

        return this.props.children;
    }
}
