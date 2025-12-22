"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Copy, Radio, Activity, ArrowLeft, Loader2, Eye, Signal } from "lucide-react";
import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { fetchStream } from "@/lib/api";
import Link from "next/link";
import { HLSPlayer } from "@/components/hls-player";

export default function StreamStudioPage() {
    const params = useParams();
    const id = params.id as string;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const router = useRouter();

    const { data: stream, isLoading, error } = useQuery({
        queryKey: ["stream", id],
        queryFn: () => fetchStream(id),
        refetchInterval: 5000,
    });

    if (isLoading) return (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
            <Loader2 className="animate-spin h-8 w-8 text-primary" />
            <p className="text-muted-foreground">Loading studio...</p>
        </div>
    );

    if (error) return (
        <div className="flex flex-col items-center justify-center py-20 gap-4 text-destructive">
            <p>Failed to load stream details</p>
            <Button variant="outline" onClick={() => window.location.reload()}>Retry</Button>
        </div>
    );

    if (!stream) return <div className="p-8 text-center">Stream not found</div>;

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        // In a real app, you'd show a toast here
    };

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            <div className="flex items-center gap-4">
                <Link href="/streams">
                    <Button variant="ghost" size="icon">
                        <ArrowLeft className="h-5 w-5" />
                    </Button>
                </Link>
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">{stream.title || "Untitled Stream"}</h1>
                    <p className="text-muted-foreground text-sm">Studio & Control</p>
                </div>
                {stream.is_live && (
                    <div className="ml-auto flex items-center gap-2 px-3 py-1 rounded-full bg-rose-500/10 text-rose-500 border border-rose-500/20 text-sm font-medium animate-pulse">
                        <Signal className="h-4 w-4" />
                        LIVE
                    </div>
                )}
            </div>

            <div className="grid gap-6 lg:grid-cols-3">
                <div className="lg:col-span-2 space-y-6">
                    {/* Preview Player */}
                    <div className="aspect-video rounded-lg overflow-hidden ring-1 ring-border">
                        {stream.is_live && stream.playback_url ? (
                            <HLSPlayer
                                src={stream.playback_url}
                                autoPlay={true}
                                muted={true}
                            />
                        ) : (
                            <div className="w-full h-full bg-black flex items-center justify-center">
                                <div className="text-muted-foreground flex flex-col items-center gap-2">
                                    <Radio className="h-12 w-12 opacity-20" />
                                    <span>Stream Offline</span>
                                    <p className="text-xs text-muted-foreground/50">Start your encoder to see preview</p>
                                </div>
                            </div>
                        )}
                    </div>

                    <Card>
                        <CardHeader><CardTitle>Connection Details</CardTitle></CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <Label>Ingest Server (RTMP)</Label>
                                <div className="flex gap-2">
                                    <Input value={stream.ingest_url || ""} readOnly className="font-mono bg-muted" />
                                    <Button variant="outline" size="icon" onClick={() => copyToClipboard(stream.ingest_url || "")}>
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label>Stream Key</Label>
                                <div className="flex gap-2">
                                    <Input type="password" value={stream.stream_key} readOnly className="font-mono bg-muted" />
                                    <Button variant="outline" size="icon" onClick={() => copyToClipboard(stream.stream_key)}>
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                                <p className="text-xs text-muted-foreground">
                                    Keep this key secret. Anyone with this key can stream to your channel.
                                </p>
                            </div>
                            <div className="space-y-2">
                                <Label>Playback URL (HLS)</Label>
                                <div className="flex gap-2">
                                    <Input value={stream.playback_url || "Not available"} readOnly className="font-mono bg-muted" />
                                    <Button variant="outline" size="icon" onClick={() => copyToClipboard(stream.playback_url || "")}>
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                <div className="space-y-6">
                    <Card>
                        <CardHeader><CardTitle>Status</CardTitle></CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                    <Eye className="h-4 w-4" /> Viewers
                                </div>
                                <span className="font-mono font-medium">{stream.current_viewers}</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                    <Activity className="h-4 w-4" /> Bitrate
                                </div>
                                <span className="font-mono font-medium">{stream.is_live ? "Unknown" : "0 kbps"}</span>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="bg-muted/30">
                        <CardContent className="p-4">
                            <h4 className="font-medium text-sm mb-2">Instructions</h4>
                            <ol className="text-xs text-muted-foreground list-decimal pl-4 space-y-2">
                                <li>Open your streaming software (OBS, vMix, etc.)</li>
                                <li>Set service to &quot;Custom&quot;</li>
                                <li>Enter the <strong>Server URL</strong> and <strong>Stream Key</strong></li>
                                <li>Start Streaming</li>
                            </ol>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
