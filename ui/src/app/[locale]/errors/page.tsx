"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle, RefreshCw, Copy } from "lucide-react";
import { format } from "date-fns";

/** Base path of the error-tracking endpoints. */
const ERRORS_ENDPOINT = "/api/v1/errors";
/** React Query cache key for the error list. */
const ERRORS_QUERY_KEY = "errors";
/** Filter value that disables source filtering. */
const ALL_SOURCES = "all";

type ErrorEvent = {
    id: string;
    source_component: string;
    severity: "warning" | "error" | "critical" | "fatal";
    message: string;
    stack_trace?: string;
    context_data?: Record<string, unknown>;
    resolved: boolean;
    created_at: string;
};

async function fetchErrorEvents(source: string): Promise<ErrorEvent[]> {
    const url = source !== ALL_SOURCES
        ? `${ERRORS_ENDPOINT}?source=${encodeURIComponent(source)}`
        : ERRORS_ENDPOINT;

    const res = await fetch(url);
    if (!res.ok) {
        throw new Error(`Failed to fetch errors: ${res.status}`);
    }
    const data = await res.json();
    return data ?? [];
}

async function resolveErrorEvent(id: string): Promise<void> {
    const res = await fetch(`${ERRORS_ENDPOINT}/${encodeURIComponent(id)}/resolve`, {
        method: "PATCH",
    });
    if (!res.ok) {
        throw new Error(`Failed to resolve error: ${res.status}`);
    }
}

export default function ErrorsPage() {
    const [filterSource, setFilterSource] = useState(ALL_SOURCES);
    const queryClient = useQueryClient();

    const {
        data: errors = [],
        isFetching: loading,
        refetch,
    } = useQuery({
        queryKey: [ERRORS_QUERY_KEY, filterSource],
        queryFn: () => fetchErrorEvents(filterSource),
    });

    const resolveMutation = useMutation({
        mutationFn: resolveErrorEvent,
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: [ERRORS_QUERY_KEY] });
        },
        onError: (e) => {
            console.error("Failed to resolve error", e);
        },
    });

    const handleResolve = (id: string) => {
        resolveMutation.mutate(id);
    };

    const getSeverityColor = (severity: string) => {
        switch (severity) {
            case "fatal": return "bg-red-900/50 text-red-200 border-red-700";
            case "critical": return "bg-red-500/20 text-red-300 border-red-500/50";
            case "error": return "bg-orange-500/20 text-orange-300 border-orange-500/50";
            case "warning": return "bg-yellow-500/20 text-yellow-300 border-yellow-500/50";
            default: return "bg-slate-800 text-slate-300";
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
    };

    return (
        <div className="p-8 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-violet-400">
                        System Errors
                    </h1>
                    <p className="text-muted-foreground mt-1">Global error tracking and alerts</p>
                </div>
                <div className="flex gap-2">
                    <select
                        className="bg-secondary/50 border border-border rounded-md px-3 py-2 text-sm"
                        value={filterSource}
                        onChange={(e) => setFilterSource(e.target.value)}
                    >
                        <option value="all">All Sources</option>
                        <option value="kernel">Kernel</option>
                        <option value="frontend">Frontend</option>
                        <option value="worker">Worker</option>
                    </select>
                    <button
                        onClick={() => void refetch()}
                        className="p-2 hover:bg-white/10 rounded-md transition-colors"
                    >
                        <RefreshCw className={`w-5 h-5 ${loading ? "animate-spin" : ""}`} />
                    </button>
                </div>
            </div>

            <div className="space-y-4">
                {errors.length === 0 && !loading ? (
                    <div className="text-center py-20 text-muted-foreground">
                        <CheckCircle className="w-12 h-12 mx-auto mb-4 text-green-500/50" />
                        <p className="text-lg">No errors detected</p>
                        <p className="text-sm">System is healthy</p>
                    </div>
                ) : (
                    errors.map((error) => (
                        <div
                            key={error.id}
                            className={`relative group border rounded-lg p-4 transition-all hover:bg-white/5 ${error.resolved ? 'opacity-50 grayscale' : 'bg-slate-900/40'}`}
                        >
                            <div className="flex items-start justify-between gap-4">
                                <div className="space-y-1">
                                    <div className="flex items-center gap-2">
                                        <span className={`px-2 py-0.5 rounded text-xs font-medium uppercase border ${getSeverityColor(error.severity)}`}>
                                            {error.severity}
                                        </span>
                                        <span className="text-xs text-muted-foreground bg-white/5 px-2 py-0.5 rounded">
                                            {error.source_component}
                                        </span>
                                        <span className="text-xs text-muted-foreground">
                                            {format(new Date(error.created_at), "yyyy-MM-dd HH:mm:ss")}
                                        </span>
                                    </div>
                                    <p className="font-mono text-sm text-foreground/90">{error.message}</p>
                                </div>

                                <div className="flex gap-1">
                                    {!error.resolved && (
                                        <button
                                            onClick={() => handleResolve(error.id)}
                                            className="opacity-0 group-hover:opacity-100 transition-opacity p-2 hover:bg-green-500/20 text-green-400 rounded-md"
                                            title="Mark as Resolved"
                                        >
                                            <CheckCircle className="w-4 h-4" />
                                        </button>
                                    )}
                                    <button
                                        onClick={() => copyToClipboard(`${error.message}\n\n${error.stack_trace || ""}`)}
                                        className="opacity-0 group-hover:opacity-100 transition-opacity p-2 hover:bg-white/10 text-muted-foreground hover:text-foreground rounded-md"
                                        title="Copy Error Details"
                                    >
                                        <Copy className="w-4 h-4" />
                                    </button>
                                </div>
                            </div>

                            {error.stack_trace && (
                                <details className="mt-3">
                                    <summary className="text-xs text-muted-foreground cursor-pointer hover:text-foreground">
                                        Show Stack Trace
                                    </summary>
                                    <pre className="mt-2 p-3 bg-black/50 rounded text-xs text-red-200/70 overflow-x-auto">
                                        {error.stack_trace}
                                    </pre>
                                </details>
                            )}
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
