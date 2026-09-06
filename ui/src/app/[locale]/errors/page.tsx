"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle, RefreshCw, Copy } from "lucide-react";
import { format } from "date-fns";
import { useTranslations } from "next-intl";

/** Timestamp format for log rows: sortable and locale-independent on purpose. */
const LOG_TIMESTAMP_FORMAT = "yyyy-MM-dd HH:mm:ss";

/**
 * Badge treatment per severity. Each level gets its own tint so `error` and
 * `warning` stay distinguishable at a glance, and `fatal` reads as the loudest.
 */
const SEVERITY_STYLES: Record<string, string> = {
    fatal: "bg-danger text-danger-foreground border-danger",
    critical: "bg-danger/20 text-danger border-danger/60",
    error: "bg-danger/10 text-danger border-danger/30",
    warning: "bg-warning/15 text-warning border-warning/40",
};

const SEVERITY_FALLBACK_STYLE = "badge-neutral";

/** Source components the API can filter on. */
const SOURCE_FILTERS = ["kernel", "frontend", "worker"] as const;

/** Shared affordance for the row actions that only appear on hover/focus. */
const ROW_ACTION_CLASS =
    "rounded-md p-2 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100 focus-ring";

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
    const t = useTranslations('errors');
    const commonT = useTranslations('common');
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

    const getSeverityColor = (severity: string) =>
        SEVERITY_STYLES[severity] ?? SEVERITY_FALLBACK_STYLE;

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
    };

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
                    <p className="text-muted-foreground mt-1">{t('subtitle')}</p>
                </div>
                <div className="flex items-center gap-2">
                    <select
                        aria-label={t('filterBySource')}
                        className="bg-secondary/50 border border-border rounded-md px-3 py-2 text-sm transition-colors hover:border-primary/40 focus-ring"
                        value={filterSource}
                        onChange={(e) => setFilterSource(e.target.value)}
                    >
                        <option value={ALL_SOURCES}>{t('allSources')}</option>
                        {SOURCE_FILTERS.map((source) => (
                            <option key={source} value={source} className="capitalize">
                                {source}
                            </option>
                        ))}
                    </select>
                    <button
                        type="button"
                        onClick={() => void refetch()}
                        disabled={loading}
                        aria-label={t('refresh')}
                        title={t('refresh')}
                        className="rounded-md p-2 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-60 focus-ring"

                    >
                        <RefreshCw
                            aria-hidden="true"
                            className={`w-5 h-5 ${loading ? "animate-spin motion-reduce:animate-none" : ""}`}
                        />
                    </button>
                </div>
            </div>

            <div className="space-y-4">
                {loading && errors.length === 0 ? (
                    <div className="space-y-3" role="status" aria-label={commonT('loading')}>
                        <div className="skeleton h-24" />
                        <div className="skeleton h-24" />
                        <div className="skeleton h-24" />
                    </div>
                ) : errors.length === 0 ? (
                    <div className="text-center py-20 text-muted-foreground">
                        <CheckCircle className="w-12 h-12 mx-auto mb-4 text-success/50" aria-hidden="true" />
                        <p className="text-lg">{t('noErrors')}</p>
                        <p className="text-sm">{t('systemHealthy')}</p>
                    </div>
                ) : (
                    errors.map((error) => (
                        <div
                            key={error.id}
                            className={`relative group border rounded-lg p-4 transition-all hover:bg-muted/60 ${error.resolved ? 'opacity-50 grayscale' : 'bg-muted/40'}`}
                        >
                            <div className="flex items-start justify-between gap-4">
                                <div className="space-y-1">
                                    <div className="flex items-center gap-2">
                                        <span className={`px-2 py-0.5 rounded text-xs font-medium uppercase border ${getSeverityColor(error.severity)}`}>
                                            {error.severity}
                                        </span>
                                        <span className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded">
                                            {error.source_component}
                                        </span>
                                        <time dateTime={error.created_at} className="text-xs text-muted-foreground">
                                            {format(new Date(error.created_at), LOG_TIMESTAMP_FORMAT)}
                                        </time>
                                    </div>
                                    <p className="font-mono text-sm text-foreground/90">{error.message}</p>
                                </div>

                                <div className="flex gap-1">
                                    {!error.resolved && (
                                        <button
                                            type="button"
                                            onClick={() => handleResolve(error.id)}
                                            className={`${ROW_ACTION_CLASS} text-success hover:bg-success/20`}
                                            aria-label={t('markResolved')}
                                            title={t('markResolved')}
                                        >
                                            <CheckCircle className="w-4 h-4" aria-hidden="true" />
                                        </button>
                                    )}
                                    <button
                                        type="button"
                                        onClick={() => copyToClipboard(`${error.message}\n\n${error.stack_trace || ""}`)}
                                        className={`${ROW_ACTION_CLASS} text-muted-foreground hover:bg-muted hover:text-foreground`}
                                        aria-label={t('copyDetails')}
                                        title={t('copyDetails')}
                                    >
                                        <Copy className="w-4 h-4" aria-hidden="true" />
                                    </button>
                                </div>
                            </div>

                            {error.stack_trace && (
                                <details className="mt-3">
                                    <summary className="inline-flex rounded-sm text-xs text-muted-foreground cursor-pointer transition-colors hover:text-foreground focus-ring">
                                        {t('showStackTrace')}
                                    </summary>
                                    <pre className="mt-2 p-3 bg-muted rounded text-xs text-danger overflow-x-auto">
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
