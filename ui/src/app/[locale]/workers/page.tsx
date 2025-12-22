"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchWorkers, Worker } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Loader2,
    Server,
    CheckCircle2,
    XCircle,
    Clock,
    Cpu,
    Activity,
    RefreshCw,
    Zap,
    HardDrive
} from "lucide-react";

export default function WorkersPage() {
    const { data: workers, isLoading, refetch, isRefetching } = useQuery({
        queryKey: ["workers"],
        queryFn: fetchWorkers,
        refetchInterval: 5000,
    });

    const healthyCount = workers?.filter(w => w.is_healthy).length || 0;
    const unhealthyCount = workers?.filter(w => !w.is_healthy).length || 0;
    const busyCount = workers?.filter(w => w.current_task_id).length || 0;

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Workers</h1>
                    <p className="text-muted-foreground mt-1">
                        Monitor your encoding worker fleet
                    </p>
                </div>
                <Button
                    variant="outline"
                    onClick={() => refetch()}
                    disabled={isRefetching}
                    className="border-border"
                >
                    <RefreshCw className={`h-4 w-4 mr-2 ${isRefetching ? 'animate-spin' : ''}`} />
                    Refresh
                </Button>
            </div>

            {/* Stats */}
            <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
                <StatCard
                    label="Total Workers"
                    value={workers?.length || 0}
                    icon={Server}
                    color="violet"
                />
                <StatCard
                    label="Healthy"
                    value={healthyCount}
                    icon={CheckCircle2}
                    color="emerald"
                />
                <StatCard
                    label="Processing"
                    value={busyCount}
                    icon={Activity}
                    color="blue"
                    pulse={busyCount > 0}
                />
                <StatCard
                    label="Unhealthy"
                    value={unhealthyCount}
                    icon={XCircle}
                    color="red"
                />
            </div>

            {/* Workers Grid */}
            {isLoading ? (
                <div className="flex justify-center py-16">
                    <div className="flex flex-col items-center gap-3">
                        <Loader2 className="h-10 w-10 animate-spin text-violet-400" />
                        <span className="text-sm text-muted-foreground">Loading workers...</span>
                    </div>
                </div>
            ) : workers && workers.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {workers.map((worker, index) => (
                        <WorkerCard
                            key={worker.id}
                            worker={worker}
                            style={{ animationDelay: `${index * 50}ms` }}
                        />
                    ))}
                </div>
            ) : (
                <Card className="card-glow">
                    <CardContent className="flex flex-col items-center justify-center py-20">
                        <Server className="h-16 w-16 text-muted-foreground/30 mb-4" />
                        <h3 className="text-xl font-semibold mb-2">No Workers Registered</h3>
                        <p className="text-muted-foreground text-center max-w-md mb-6">
                            Start a worker to begin processing encoding jobs. Workers automatically register with the kernel.
                        </p>
                        <div className="bg-muted/30 rounded-lg p-4 border border-border">
                            <p className="text-sm text-muted-foreground mb-2">Start a worker with:</p>
                            <code className="mono-highlight text-sm">docker compose up worker</code>
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}

interface WorkerCardProps {
    worker: Worker;
    style?: React.CSSProperties;
}

function WorkerCard({ worker, style }: WorkerCardProps) {
    const lastSeen = new Date(worker.last_heartbeat);
    const timeSince = Date.now() - lastSeen.getTime();
    const isRecent = timeSince < 60000;
    const isProcessing = !!worker.current_task_id;

    return (
        <Card
            className={`card-glow animate-[slide-up_0.3s_ease-out] ${!worker.is_healthy ? 'border-red-500/30' :
                    isProcessing ? 'border-blue-500/30' : ''
                }`}
            style={style}
        >
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${worker.is_healthy
                                ? isProcessing
                                    ? 'bg-blue-500/10'
                                    : 'bg-emerald-500/10'
                                : 'bg-red-500/10'
                            }`}>
                            {worker.is_healthy ? (
                                isProcessing ? (
                                    <Activity className="h-5 w-5 text-blue-400 animate-pulse" />
                                ) : (
                                    <CheckCircle2 className="h-5 w-5 text-emerald-400" />
                                )
                            ) : (
                                <XCircle className="h-5 w-5 text-red-400" />
                            )}
                        </div>
                        <div>
                            <CardTitle className="text-base font-mono">{worker.id}</CardTitle>
                            <CardDescription className="text-xs">
                                Worker Node
                            </CardDescription>
                        </div>
                    </div>
                    <Badge
                        variant="outline"
                        className={`${worker.is_healthy
                                ? isProcessing
                                    ? 'badge-info'
                                    : 'badge-success'
                                : 'badge-error'
                            }`}
                    >
                        {worker.is_healthy
                            ? isProcessing
                                ? 'Processing'
                                : 'Healthy'
                            : 'Unhealthy'
                        }
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="space-y-4">
                {/* Heartbeat */}
                <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2 text-muted-foreground">
                        <Clock className="h-4 w-4" />
                        <span>Last heartbeat</span>
                    </div>
                    <span className={isRecent ? 'text-emerald-400' : 'text-amber-400'}>
                        {isRecent ? 'Just now' : formatTimeSince(timeSince)}
                    </span>
                </div>

                {/* Current Task */}
                {isProcessing && (
                    <div className="p-3 rounded-lg bg-blue-500/10 border border-blue-500/20">
                        <div className="flex items-center gap-2">
                            <Zap className="h-4 w-4 text-blue-400" />
                            <span className="text-sm font-medium text-blue-400">Active Task</span>
                        </div>
                        <code className="text-xs mono-highlight mt-2 block">
                            {worker.current_task_id?.slice(0, 16)}...
                        </code>
                    </div>
                )}

                {/* Capabilities */}
                {worker.capabilities && Object.keys(worker.capabilities).length > 0 && (
                    <div className="space-y-2">
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                            <Cpu className="h-3.5 w-3.5" />
                            <span>Capabilities</span>
                        </div>
                        <div className="flex flex-wrap gap-1.5">
                            {Object.entries(worker.capabilities).map(([key, value]) => (
                                <Badge key={key} variant="outline" className="text-xs bg-muted/30">
                                    {key}: {String(value)}
                                </Badge>
                            ))}
                        </div>
                    </div>
                )}

                {/* Status Bar */}
                <div className="pt-3 border-t border-border flex items-center justify-between text-xs text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                        <HardDrive className="h-3.5 w-3.5" />
                        <span>Ready for tasks</span>
                    </div>
                    <span className={`h-2 w-2 rounded-full ${worker.is_healthy ? 'bg-emerald-400' : 'bg-red-400'
                        } ${worker.is_healthy && 'animate-pulse'}`} />
                </div>
            </CardContent>
        </Card>
    );
}

interface StatCardProps {
    label: string;
    value: number;
    icon: React.ElementType;
    color: "violet" | "emerald" | "blue" | "red";
    pulse?: boolean;
}

function StatCard({ label, value, icon: Icon, color, pulse }: StatCardProps) {
    const colors = {
        violet: "text-violet-400 bg-violet-500/10 border-violet-500/20",
        emerald: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
        blue: "text-blue-400 bg-blue-500/10 border-blue-500/20",
        red: "text-red-400 bg-red-500/10 border-red-500/20",
    };

    return (
        <div className={`flex items-center gap-3 p-4 rounded-lg border ${colors[color]}`}>
            <div className="relative">
                <Icon className={`h-8 w-8 ${colors[color].split(' ')[0]}`} />
                {pulse && (
                    <span className="absolute -top-1 -right-1 h-3 w-3">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                        <span className="relative inline-flex rounded-full h-3 w-3 bg-blue-500" />
                    </span>
                )}
            </div>
            <div>
                <div className={`text-2xl font-bold ${colors[color].split(' ')[0]}`}>{value}</div>
                <div className="text-sm text-muted-foreground">{label}</div>
            </div>
        </div>
    );
}

function formatTimeSince(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    return `${hours}h ago`;
}
