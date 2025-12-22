"use client";

import { useRef, useEffect, useState } from "react";
import Hls from "hls.js";
import { Loader2, AlertCircle, Video } from "lucide-react";

interface HLSPlayerProps {
    src: string;
    autoPlay?: boolean;
    muted?: boolean;
    className?: string;
}

export function HLSPlayer({ src, autoPlay = true, muted = true, className = "" }: HLSPlayerProps) {
    const videoRef = useRef<HTMLVideoElement>(null);
    const hlsRef = useRef<Hls | null>(null);
    const [status, setStatus] = useState<"loading" | "playing" | "error" | "waiting">("loading");
    const [errorMessage, setErrorMessage] = useState<string>("");

    useEffect(() => {
        const video = videoRef.current;
        if (!video || !src) return;

        // Check if HLS is natively supported (Safari)
        if (video.canPlayType("application/vnd.apple.mpegurl")) {
            video.src = src;
            video.addEventListener("loadedmetadata", () => {
                setStatus("playing");
                if (autoPlay) video.play().catch(() => { });
            });
            video.addEventListener("error", () => {
                setStatus("error");
                setErrorMessage("Failed to load stream");
            });
        } else if (Hls.isSupported()) {
            const hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true,
                backBufferLength: 30,
            });

            hlsRef.current = hls;

            hls.loadSource(src);
            hls.attachMedia(video);

            hls.on(Hls.Events.MANIFEST_PARSED, () => {
                setStatus("playing");
                if (autoPlay) video.play().catch(() => { });
            });

            hls.on(Hls.Events.ERROR, (_, data) => {
                if (data.fatal) {
                    switch (data.type) {
                        case Hls.ErrorTypes.NETWORK_ERROR:
                            setStatus("waiting");
                            setErrorMessage("Stream not available yet. Waiting...");
                            // Try to recover after delay
                            setTimeout(() => {
                                hls.loadSource(src);
                            }, 3000);
                            break;
                        case Hls.ErrorTypes.MEDIA_ERROR:
                            setStatus("error");
                            setErrorMessage("Media error - trying to recover");
                            hls.recoverMediaError();
                            break;
                        default:
                            setStatus("error");
                            setErrorMessage(`Fatal error: ${data.details}`);
                            break;
                    }
                }
            });

            return () => {
                hls.destroy();
                hlsRef.current = null;
            };
        } else {
            setStatus("error");
            setErrorMessage("HLS is not supported in this browser");
        }
    }, [src, autoPlay]);

    return (
        <div className={`relative bg-black w-full h-full ${className}`}>
            <video
                ref={videoRef}
                className="w-full h-full object-contain"
                controls
                muted={muted}
                playsInline
            />

            {status === "loading" && (
                <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/80 text-white gap-3">
                    <Loader2 className="h-8 w-8 animate-spin text-violet-400" />
                    <span className="text-sm text-muted-foreground">Connecting to stream...</span>
                </div>
            )}

            {status === "waiting" && (
                <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/80 text-white gap-3">
                    <Video className="h-12 w-12 text-amber-400 animate-pulse" />
                    <span className="text-sm text-amber-400">{errorMessage}</span>
                    <span className="text-xs text-muted-foreground">The stream will appear when ready</span>
                </div>
            )}

            {status === "error" && (
                <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/80 text-white gap-3">
                    <AlertCircle className="h-8 w-8 text-red-400" />
                    <span className="text-sm text-red-400">{errorMessage}</span>
                </div>
            )}

            {/* Stream URL indicator */}
            <div className="absolute bottom-2 left-2 text-xs font-mono bg-black/50 px-2 py-1 rounded text-white/70 max-w-[80%] truncate">
                {src}
            </div>
        </div>
    );
}
