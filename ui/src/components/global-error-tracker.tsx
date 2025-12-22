"use client";

import { useEffect } from "react";

export function GlobalErrorTracker() {
    useEffect(() => {
        const reportError = async (
            message: string,
            source: string,
            stack?: string,
            context?: any
        ) => {
            try {
                // Debounce or filter distinct errors if needed, but for now send all
                await fetch("/api/v1/errors", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        source: "frontend",
                        severity: "error", // Default to error
                        message: message,
                        stack_trace: stack,
                        context_data: {
                            ...context,
                            url: window.location.href,
                            userAgent: navigator.userAgent,
                        },
                    }),
                });
            } catch (err) {
                // Prevent infinite loops if reporting fails
                console.warn("Failed to report error to backend", err);
            }
        };

        // 1. Capture Global JS Errors
        const handleError = (event: ErrorEvent) => {
            reportError(
                event.message,
                "frontend:js",
                event.error?.stack,
                {
                    filename: event.filename,
                    lineno: event.lineno,
                    colno: event.colno,
                }
            );
        };

        // 2. Capture Unhandled Promise Rejections
        const handleRejection = (event: PromiseRejectionEvent) => {
            let message = "Unhandled Promise Rejection";
            let stack = "";

            if (event.reason instanceof Error) {
                message = event.reason.message;
                stack = event.reason.stack || "";
            } else if (typeof event.reason === "string") {
                message = event.reason;
            }

            reportError(
                message,
                "frontend:promise",
                stack,
                { reason: event.reason }
            );
        };

        // 3. Capture Resource Loading Errors (img, script, css)
        // using capture=true
        const handleResourceError = (event: Event) => {
            // Filter out window error events which are duplicate of Step 1
            if (event instanceof ErrorEvent) return;

            const target = event.target as HTMLElement;
            if (target) {
                const url = (target as any).src || (target as any).href;
                if (url) {
                    reportError(
                        `Failed to load resource: ${url}`,
                        "frontend:resource",
                        `<${target.tagName.toLowerCase()} src="${url}">`,
                        { tagName: target.tagName, url }
                    );
                }
            }
        };

        window.addEventListener("error", handleError);
        window.addEventListener("unhandledrejection", handleRejection);
        // Capture phase for resource errors
        window.addEventListener("error", handleResourceError, true);

        return () => {
            window.removeEventListener("error", handleError);
            window.removeEventListener("unhandledrejection", handleRejection);
            window.removeEventListener("error", handleResourceError, true);
        };
    }, []);

    return null;
}
