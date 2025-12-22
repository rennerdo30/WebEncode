"use client";

import { useEffect, useState } from "react";
import { AlertCircle, CheckCircle, Filter, RefreshCw, XCircle, Copy, Check } from "lucide-react";
import { format } from "date-fns";

type ErrorEvent = {
    id: string;
    source_component: string;
    severity: "warning" | "error" | "critical" | "fatal";
    message: string;
    stack_trace?: string;
    context_data?: any;
    resolved: boolean;
    created_at: string;
};

export default function ErrorsPage() {
    const [errors, setErrors] = useState<ErrorEvent[]>([]);
    const [loading, setLoading] = useState(true);
    const [filterSource, setFilterSource] = useState("all");

    const fetchErrors = async () => {
        setLoading(true);
        try {
            const url = filterSource !== "all"
                ? `/api/v1/errors?source=${filterSource}`
                : `/api/v1/errors`;

            const res = await fetch(url);
            if (res.ok) {
                const data = await res.json();
                setErrors(data || []);
            }
        } catch (e) {
            console.error("Failed to fetch errors", e);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchErrors();
    }, [filterSource]);

    const handleResolve = async (id: string) => {
        try {
            const res = await fetch(`/api/v1/errors/${id}/resolve`, { method: "PATCH" });
            if (res.ok) {
                fetchErrors();
            }
        } catch (e) {
            console.error("Failed to resolve error", e);
        }
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
                        onClick={fetchErrors}
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
