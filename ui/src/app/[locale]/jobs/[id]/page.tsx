"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useRouter } from "@/i18n/routing";
import { fetchJob, cancelJob, deleteJob, retryJob, fetchJobLogs, fetchJobOutputs, fetchPlugins, publishJob, Task, JobOutput, PublishJobRequest } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
    ArrowLeft,
    CheckCircle2,
    Circle,
    Loader2,
    XCircle,
    Ban,
    Trash2,
    RefreshCw,
    Copy,
    FileJson,
    Check,
    Download,
    Upload,
    ExternalLink,
    FileVideo
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

export default function JobDetailsPage() {
    const { id } = useParams() as { id: string };
    const router = useRouter();
    const queryClient = useQueryClient();

    const { data, isLoading, error } = useQuery({
        queryKey: ["job", id],
        queryFn: () => fetchJob(id),
        refetchInterval: 5000,
    });

    const { data: logs } = useQuery({
        queryKey: ["jobLogs", id],
        queryFn: () => fetchJobLogs(id),
        refetchInterval: 5000,
    });

    const cancelMutation = useMutation({
        mutationFn: () => cancelJob(id),
        onSuccess: () => {
            toast.success("Job cancelled successfully");
            queryClient.invalidateQueries({ queryKey: ["job", id] });
        },
        onError: (error) => {
            toast.error("Failed to cancel job", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    const deleteMutation = useMutation({
        mutationFn: () => deleteJob(id),
        onSuccess: () => {
            toast.success("Job deleted successfully");
            router.push("/jobs");
        },
        onError: (error) => {
            toast.error("Failed to delete job", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    const retryMutation = useMutation({
        mutationFn: () => retryJob(id),
        onSuccess: (newJob) => {
            toast.success("Job restarted successfully", {
                description: newJob?.id ? `New job ID: ${newJob.id.slice(0, 8)}...` : undefined
            });
            if (newJob && newJob.id && newJob.id !== id) {
                router.push(`/jobs/${newJob.id}`);
            } else {
                queryClient.invalidateQueries({ queryKey: ["job", id] });
            }
        },
        onError: (error) => {
            toast.error("Failed to retry job", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    // State for publish dialog
    const [publishDialogOpen, setPublishDialogOpen] = useState(false);
    const [publishPlatform, setPublishPlatform] = useState("");
    const [publishTitle, setPublishTitle] = useState("");
    const [publishDescription, setPublishDescription] = useState("");
    const [publishAccessToken, setPublishAccessToken] = useState("");

    const isCompleted = data?.job?.status === "completed";

    // Fetch outputs for completed jobs
    const { data: outputsData } = useQuery({
        queryKey: ["jobOutputs", id],
        queryFn: () => fetchJobOutputs(id),
        enabled: isCompleted,
    });

    // Fetch available publisher plugins
    const { data: plugins } = useQuery({
        queryKey: ["plugins"],
        queryFn: fetchPlugins,
        enabled: publishDialogOpen,
    });

    const publisherPlugins = plugins?.filter(p => p.type === "publisher" && p.is_enabled) || [];

    // Publish mutation
    const publishMutation = useMutation({
        mutationFn: (request: PublishJobRequest) => publishJob(id, request),
        onSuccess: (result) => {
            toast.success("Published successfully!", {
                description: result.platform_url ? `View at: ${result.platform_url}` : result.message
            });
            setPublishDialogOpen(false);
            resetPublishForm();
        },
        onError: (error) => {
            toast.error("Failed to publish", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    const resetPublishForm = () => {
        setPublishPlatform("");
        setPublishTitle("");
        setPublishDescription("");
        setPublishAccessToken("");
    };

    const handlePublish = () => {
        if (!publishPlatform || !publishAccessToken) {
            toast.error("Please fill in all required fields");
            return;
        }
        publishMutation.mutate({
            platform: publishPlatform,
            title: publishTitle || `Video ${id.slice(0, 8)}`,
            description: publishDescription,
            access_token: publishAccessToken,
        });
    };

    const handleDownload = (url: string, filename: string) => {
        // Obfuscate the parameters to avoid showing internal paths in the browser URL bar
        const payload = encodeURIComponent(btoa(JSON.stringify({ url, filename })));
        const proxyUrl = `/api/download?data=${payload}`;
        window.open(proxyUrl, '_blank');
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center p-16">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    if (error || !data) {
        return (
            <div className="p-8 text-center">
                <p className="text-muted-foreground">Job not found or failed to load.</p>
                <Link href="/jobs" className="text-primary hover:underline">
                    Back to Jobs
                </Link>
            </div>
        );
    }

    const { job, tasks } = data;
    const progress = job.progress_pct ?? 0;
    const isRunning = job.status === "processing" || job.status === "queued" || job.status === "stitching";
    const isFailed = job.status === "failed" || job.status === "cancelled";

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-4">
                <Link href="/jobs" className="p-2 hover:bg-muted rounded-md transition-colors">
                    <ArrowLeft className="h-5 w-5" />
                </Link>
                <div className="flex-1">
                    <div className="flex items-center justify-between">
                        <div className="flex flex-col">
                            <h1 className="text-3xl font-bold tracking-tight">
                                Job {id.slice(0, 8)}
                            </h1>
                            <p className="text-muted-foreground truncate max-w-xl text-sm mt-1">{job.source_url}</p>
                        </div>
                        <div className="flex items-center gap-3">
                            <StatusBadge status={job.status} />

                            {isFailed && (
                                <Button
                                    variant="default"
                                    size="sm"
                                    className="h-8 px-3"
                                    onClick={() => retryMutation.mutate()}
                                    disabled={retryMutation.isPending}
                                >
                                    {retryMutation.isPending ? (
                                        <Loader2 className="h-4 w-4 animate-spin mr-1.5" />
                                    ) : (
                                        <RefreshCw className="h-4 w-4 mr-1.5" />
                                    )}
                                    Retry
                                </Button>
                            )}

                            {isRunning && (
                                <Button
                                    variant="destructive"
                                    size="sm"
                                    className="h-8 px-3"
                                    onClick={() => cancelMutation.mutate()}
                                    disabled={cancelMutation.isPending}
                                >
                                    <Ban className="h-4 w-4 mr-1.5" />
                                    Cancel
                                </Button>
                            )}

                            {isCompleted && (
                                <Button
                                    variant="default"
                                    size="sm"
                                    className="h-8 px-3"
                                    onClick={() => setPublishDialogOpen(true)}
                                >
                                    <Upload className="h-4 w-4 mr-1.5" />
                                    Publish
                                </Button>
                            )}

                            <Button
                                variant="outline"
                                size="sm"
                                className="h-8 px-3"
                                onClick={() => deleteMutation.mutate()}
                                disabled={deleteMutation.isPending || isRunning}
                            >
                                <Trash2 className="h-4 w-4 mr-1.5" />
                                Delete
                            </Button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader><CardTitle>Progress</CardTitle></CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                                <span>Overall</span>
                                <span>{Math.round(progress)}%</span>
                            </div>
                            <Progress value={progress} className="h-3 relative overflow-hidden transition-all" />
                        </div>
                        {job.status === "processing" && job.eta_seconds !== null && job.eta_seconds > 0 && (
                            <p className="text-sm text-muted-foreground">
                                ETA: {formatETA(job.eta_seconds)}
                            </p>
                        )}
                        {job.error_message && (
                            <div className="group relative p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-500 text-sm">
                                <p>{job.error_message}</p>
                                <CopyButton text={job.error_message} className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity" />
                            </div>
                        )}
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader><CardTitle>Details</CardTitle></CardHeader>
                    <CardContent className="space-y-2 text-sm">
                        <DetailRow label="ID" value={job.id} mono copyable />
                        <DetailRow label="Status" value={job.status} />
                        <DetailRow label="Created" value={new Date(job.created_at).toLocaleString()} />
                        {job.started_at && (
                            <DetailRow label="Started" value={new Date(job.started_at).toLocaleString()} />
                        )}
                        {job.finished_at && (
                            <DetailRow label="Finished" value={new Date(job.finished_at).toLocaleString()} />
                        )}
                        <DetailRow label="Profiles" value={job.profiles.join(", ")} />
                    </CardContent>
                </Card>
            </div>

            {/* Outputs Section - Only shown for completed jobs */}
            {isCompleted && outputsData && outputsData.outputs.length > 0 && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <FileVideo className="h-5 w-5" />
                            Output Files
                        </CardTitle>
                        <CardDescription>
                            Download or publish the encoded video files
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3">
                            {outputsData.outputs.map((output, index) => (
                                <div
                                    key={index}
                                    className="flex items-center justify-between p-3 bg-muted/30 rounded-lg border border-border/50"
                                >
                                    <div className="flex items-center gap-3">
                                        <div className="h-10 w-10 rounded-md bg-primary/10 flex items-center justify-center">
                                            <FileVideo className="h-5 w-5 text-primary" />
                                        </div>
                                        <div>
                                            <p className="font-medium text-sm">{output.name}</p>
                                            <div className="flex items-center gap-2">
                                                <Badge variant="secondary" className="text-xs">
                                                    {output.type === 'final' ? 'Final Output' : output.type}
                                                </Badge>
                                                {output.profile && (
                                                    <span className="text-xs text-muted-foreground">{output.profile}</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        {output.download_url && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleDownload(output.download_url!, output.name)}
                                            >
                                                <Download className="h-4 w-4 mr-1.5" />
                                                Download
                                            </Button>
                                        )}
                                        {output.type === 'final' && (
                                            <Button
                                                variant="default"
                                                size="sm"
                                                onClick={() => setPublishDialogOpen(true)}
                                            >
                                                <Upload className="h-4 w-4 mr-1.5" />
                                                Publish
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            )}

            <Tabs defaultValue="graph" className="w-full">
                <div className="flex items-center justify-between mb-2">
                    <TabsList>
                        <TabsTrigger value="graph">Pipeline Graph</TabsTrigger>
                        <TabsTrigger value="list">Task List</TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="graph" className="mt-0">
                    <Card>
                        <CardHeader className="pb-3 border-b border-border/50">
                            <CardTitle className="text-base font-medium">Execution Pipeline</CardTitle>
                        </CardHeader>
                        <CardContent className="p-6 overflow-x-auto">
                            <JobPipelineGraph tasks={tasks} />
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="list" className="mt-0">
                    <Card>
                        <CardContent className="p-0">
                            <div className="divide-y divide-border/50">
                                {tasks && tasks.length > 0 ? (
                                    tasks.sort((a, b) => {
                                        // Sort by type priority then index
                                        const typePrio: Record<string, number> = { probe: 0, transcode: 1, stitch: 2, upload: 3 };
                                        if (typePrio[a.type] !== typePrio[b.type]) return typePrio[a.type] - typePrio[b.type];
                                        return (a.sequence_index || 0) - (b.sequence_index || 0);
                                    }).map((task) => (
                                        <div key={task.id} className="p-4 hover:bg-muted/30 transition-colors">
                                            <TaskRow task={task} />
                                        </div>
                                    ))
                                ) : (
                                    <div className="p-8 text-center text-muted-foreground">No tasks yet.</div>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>


            <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle>System Logs</CardTitle>
                    <Button
                        variant="outline"
                        size="sm"
                        className="h-7 px-2 text-xs gap-1.5"
                        onClick={() => navigator.clipboard.writeText(logs?.map(l => `[${l.level}] ${l.message}`).join('\n') || '')}
                    >
                        <Copy className="h-3 w-3" />
                        Copy All
                    </Button>
                </CardHeader>
                <CardContent>
                    <div className="bg-black/90 text-zinc-400 p-4 rounded-md font-mono text-xs h-64 overflow-y-auto whitespace-pre-wrap">
                        {logs && logs.length > 0 ? (
                            logs.map((log) => (
                                <div key={log.id} className="group flex gap-2 border-b border-zinc-800/50 pb-1 mb-1 last:border-0 last:mb-0 last:pb-0 hover:bg-white/5 px-1 -mx-1 rounded">
                                    <span className="text-zinc-600 shrink-0 w-20">{new Date(log.created_at).toLocaleTimeString()}</span>
                                    <span className={`w-12 shrink-0 font-bold ${log.level === "error" ? "text-red-400" : log.level === "warn" ? "text-yellow-400" : "text-zinc-300"}`}>
                                        {log.level.toUpperCase()}
                                    </span>
                                    <span className="break-all">{log.message}</span>
                                </div>
                            ))
                        ) : (
                            <span className="text-zinc-600">No logs available...</span>
                        )}
                    </div>
                </CardContent>
            </Card>

            {/* Publish Dialog */}
            <Dialog open={publishDialogOpen} onOpenChange={setPublishDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>Publish Video</DialogTitle>
                        <DialogDescription>
                            Upload your encoded video to an external platform.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="platform">Platform</Label>
                            <Select value={publishPlatform} onValueChange={setPublishPlatform}>
                                <SelectTrigger>
                                    <SelectValue placeholder="Select a platform" />
                                </SelectTrigger>
                                <SelectContent>
                                    {publisherPlugins.length > 0 ? (
                                        publisherPlugins.map((plugin) => (
                                            <SelectItem key={plugin.id} value={plugin.id.replace('publisher-', '')}>
                                                {plugin.id.replace('publisher-', '').charAt(0).toUpperCase() +
                                                    plugin.id.replace('publisher-', '').slice(1)}
                                            </SelectItem>
                                        ))
                                    ) : (
                                        <>
                                            <SelectItem value="twitch">Twitch</SelectItem>
                                            <SelectItem value="youtube">YouTube</SelectItem>
                                            <SelectItem value="kick">Kick</SelectItem>
                                            <SelectItem value="rumble">Rumble</SelectItem>
                                        </>
                                    )}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="title">Title</Label>
                            <Input
                                id="title"
                                value={publishTitle}
                                onChange={(e) => setPublishTitle(e.target.value)}
                                placeholder={`Video ${id.slice(0, 8)}`}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="description">Description (Optional)</Label>
                            <Textarea
                                id="description"
                                value={publishDescription}
                                onChange={(e) => setPublishDescription(e.target.value)}
                                placeholder="Enter a description for your video..."
                                rows={3}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="accessToken">Access Token</Label>
                            <Input
                                id="accessToken"
                                type="password"
                                value={publishAccessToken}
                                onChange={(e) => setPublishAccessToken(e.target.value)}
                                placeholder="Enter your platform OAuth token"
                            />
                            <p className="text-xs text-muted-foreground">
                                You can obtain this token from your platform&apos;s developer settings or OAuth flow.
                            </p>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setPublishDialogOpen(false)}>
                            Cancel
                        </Button>
                        <Button
                            onClick={handlePublish}
                            disabled={publishMutation.isPending || !publishPlatform || !publishAccessToken}
                        >
                            {publishMutation.isPending ? (
                                <>
                                    <Loader2 className="h-4 w-4 animate-spin mr-1.5" />
                                    Publishing...
                                </>
                            ) : (
                                <>
                                    <Upload className="h-4 w-4 mr-1.5" />
                                    Publish
                                </>
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

// ---- Sub Components ----

function JobPipelineGraph({ tasks }: { tasks: Task[] }) {
    // Group tasks by type
    const probeTask = tasks?.find(t => t.type === 'probe');
    const transcodeTasks = tasks?.filter(t => t.type === 'transcode').sort((a, b) => a.sequence_index - b.sequence_index) || [];
    const stitchTask = tasks?.find(t => t.type === 'stitch');

    // Determine workflow state
    const probeComplete = probeTask?.status === 'completed';
    const allTranscodeComplete = transcodeTasks.length > 0 && transcodeTasks.every(t => t.status === 'completed');

    return (
        <div className="flex flex-row items-center justify-center py-6 gap-2 overflow-x-auto">
            {/* Stage 1: Probe */}
            <div className="flex flex-col items-center">
                <div className="text-[10px] text-muted-foreground mb-1 uppercase tracking-wider">Step 1</div>
                <PipelineNode task={probeTask} label="Media Probe" pending={!probeTask} />
            </div>

            {/* Connector */}
            <div className="h-px w-8 bg-border mt-4" />

            {/* Stage 2: Transcode */}
            <div className="flex flex-col items-center">
                <div className="text-[10px] text-muted-foreground mb-1 uppercase tracking-wider">Step 2</div>
                {transcodeTasks.length > 0 ? (
                    <div className="flex flex-col gap-2 max-h-[400px] flex-wrap justify-center">
                        {transcodeTasks.map((t) => (
                            <PipelineNode key={t.id} task={t} label={`Segment #${t.sequence_index}`} small />
                        ))}
                    </div>
                ) : (
                    <PipelineNode label="Transcode" pending={!probeComplete} waiting={probeComplete} />
                )}
            </div>

            {/* Connector */}
            <div className="h-px w-8 bg-border mt-4" />

            {/* Stage 3: Stitch */}
            <div className="flex flex-col items-center">
                <div className="text-[10px] text-muted-foreground mb-1 uppercase tracking-wider">Step 3</div>
                <PipelineNode
                    task={stitchTask}
                    label="Stitch & Finalize"
                    pending={!allTranscodeComplete && !stitchTask}
                    waiting={allTranscodeComplete && !stitchTask}
                />
            </div>
        </div>
    );
}

function PipelineNode({ task, label, small, pending, waiting }: {
    task?: Task;
    label: string;
    small?: boolean;
    pending?: boolean;
    waiting?: boolean;
}) {
    // Pending = greyed out, not yet available
    // Waiting = dashed border, ready to start
    if (!task) {
        const baseStyle = small ? 'px-3 py-1.5 text-xs' : 'px-5 py-2.5 text-sm';
        if (pending) {
            return (
                <div className={`border border-dashed border-zinc-700 rounded-lg bg-zinc-800/30 text-zinc-500 ${baseStyle}`}>
                    {label}
                </div>
            );
        }
        if (waiting) {
            return (
                <div className={`border-2 border-dashed border-yellow-500/40 rounded-lg bg-yellow-500/5 text-yellow-500/70 ${baseStyle} animate-pulse`}>
                    ⏳ {label}
                </div>
            );
        }
        return (
            <div className={`border border-dashed border-muted rounded-lg bg-muted/20 text-muted-foreground ${baseStyle}`}>
                {label}
            </div>
        );
    }

    const statusStyles: Record<string, string> = {
        pending: "border-zinc-500/30 bg-zinc-900/50 text-zinc-400",
        assigned: "border-blue-500/50 bg-blue-500/10 text-blue-400 shadow-blue-500/10 shadow-md",
        completed: "border-emerald-500/50 bg-emerald-500/10 text-emerald-400",
        failed: "border-red-500/50 bg-red-500/15 text-red-400",
    };

    const StatusIcon = {
        pending: <Circle className="h-3.5 w-3.5 opacity-50" />,
        assigned: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
        completed: <CheckCircle2 className="h-3.5 w-3.5" />,
        failed: <XCircle className="h-3.5 w-3.5" />,
    }[task.status] || <Circle className="h-3.5 w-3.5" />;

    return (
        <div className={`border rounded-lg ${statusStyles[task.status] || statusStyles.pending} ${small ? 'px-3 py-1.5' : 'px-5 py-2.5'} transition-all`}>
            <div className="flex items-center gap-2">
                {StatusIcon}
                <div className="flex flex-col">
                    <span className={`font-medium ${small ? 'text-xs' : 'text-sm'}`}>{label}</span>
                    {!small && <span className="text-[10px] opacity-50 font-mono">{task.id.slice(0, 8)}</span>}
                </div>
            </div>
        </div>
    );
}

// ... existing helpers updated ...

function StatusBadge({ status }: { status: string }) {
    const styles: Record<string, string> = {
        queued: "bg-yellow-500/10 text-yellow-600 border-yellow-500/30",
        processing: "bg-blue-500/10 text-blue-500 border-blue-500/30",
        stitching: "bg-purple-500/10 text-purple-500 border-purple-500/30",
        uploading: "bg-cyan-500/10 text-cyan-500 border-cyan-500/30",
        completed: "bg-emerald-500/10 text-emerald-500 border-emerald-500/30",
        failed: "bg-red-500/15 text-red-500 border-red-500/40 font-medium",
        cancelled: "bg-zinc-500/10 text-zinc-400 border-zinc-500/30",
    };
    return (
        <Badge variant="outline" className={`h-6 px-2.5 text-xs capitalize ${styles[status] || ""}`}>
            {status}
        </Badge>
    );
}

function DetailRow({ label, value, mono, copyable }: { label: string; value: string; mono?: boolean; copyable?: boolean }) {
    return (
        <div className="flex justify-between items-center py-1 border-b border-border/50 group h-8">
            <span className="text-muted-foreground">{label}</span>
            <div className="flex items-center gap-2">
                <span className={mono ? "font-mono text-xs" : ""}>{value}</span>
                {copyable && <CopyButton text={value} className="opacity-0 group-hover:opacity-100 transition-opacity h-6 w-6" />}
            </div>
        </div>
    );
}

function TaskRow({ task }: { task: Task }) {
    const [showJson, setShowJson] = useState(false);

    const statusIcons: Record<string, React.ReactNode> = {
        pending: <Circle className="h-4 w-4 text-muted-foreground" />,
        assigned: <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />,
        completed: <CheckCircle2 className="h-4 w-4 text-green-500" />,
        failed: <XCircle className="h-4 w-4 text-red-500" />,
    };

    const typeLabels: Record<string, string> = {
        probe: "Probe Media",
        transcode: "Transcode",
        stitch: "Stitch Segments",
        upload: "Upload",
    };

    // Label for sequence
    let seqLabel = "";
    if (task.type === 'stitch' || task.type === 'probe' || task.type === 'upload') {
        // Don't show segment label for non-segment types
    } else if (task.sequence_index !== null && task.sequence_index >= 0) {
        seqLabel = `Segment #${task.sequence_index}`;
    }

    // Try to parse result for Probe tasks
    let probeResult = null;
    if (task.type === 'probe' && task.result) {
        try {
            probeResult = typeof task.result === 'string' ? JSON.parse(task.result) : task.result;
        } catch { }
    }

    return (
        <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    {statusIcons[task.status] || <Circle className="h-4 w-4" />}
                    <div>
                        <div className="flex items-center gap-2">
                            <p className="font-medium">{typeLabels[task.type] || task.type}</p>
                            {seqLabel && <Badge variant="secondary" className="text-[10px] h-4 px-1">{seqLabel}</Badge>}
                        </div>
                        <p className="text-xs text-muted-foreground font-mono">
                            {task.id.slice(0, 8)}
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-3">
                    <span className="text-xs text-muted-foreground capitalize">{task.status}</span>
                    {task.result && <CopyButton text={typeof task.result === 'string' ? task.result : JSON.stringify(task.result)} />}
                </div>
            </div>

            {/* Custom UI for Probe Result */}
            {probeResult && task.type === 'probe' ? (
                <div className="mt-2 border rounded-md overflow-hidden bg-background">
                    <div className="border-b px-3 py-1.5 flex items-center justify-between bg-muted/40">
                        <span className="text-xs font-semibold flex items-center gap-1"><FileJson className="h-3 w-3" /> Media Metadata</span>
                        <div className="flex items-center gap-2">
                            <span className="text-xs text-muted-foreground">{probeResult.Format}</span>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 text-[10px] px-2"
                                onClick={() => setShowJson(!showJson)}
                            >
                                {showJson ? "View GUI" : "View JSON"}
                            </Button>
                        </div>
                    </div>
                    {showJson ? (
                        <div className="text-xs bg-black/80 text-green-400 p-2 overflow-auto font-mono w-full max-h-40 whitespace-pre-wrap relative group">
                            {typeof task.result === 'string' ? task.result : JSON.stringify(task.result, null, 2)}
                        </div>
                    ) : (
                        <div className="p-3 space-y-3">
                            {/* Main Stats Row */}
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                                <div>
                                    <span className="text-muted-foreground block mb-0.5">Duration</span>
                                    <span className="font-mono font-medium">{formatDuration(probeResult.Duration)}</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground block mb-0.5">Bitrate</span>
                                    <span className="font-mono font-medium">{probeResult.Bitrate ? formatBitrate(probeResult.Bitrate) : 'N/A'}</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground block mb-0.5">Resolution</span>
                                    <span className="font-mono font-medium">{probeResult.Width}×{probeResult.Height}</span>
                                </div>
                                <div>
                                    <span className="text-muted-foreground block mb-0.5">Keyframes</span>
                                    <span className="font-mono font-medium">{probeResult.Keyframes?.length || 0}</span>
                                </div>
                            </div>

                            {/* Streams Table */}
                            {probeResult.Streams && probeResult.Streams.length > 0 && (
                                <div className="border rounded-md">
                                    <div className="px-2 py-1 bg-muted/30 text-xs font-medium border-b">
                                        Streams ({probeResult.Streams.length})
                                    </div>
                                    <div className="divide-y divide-border/50">
                                        {probeResult.Streams.map((stream: { Index: number; CodecType: string; CodecName: string }) => (
                                            <div key={stream.Index} className="px-2 py-1.5 flex items-center gap-3 text-xs">
                                                <Badge variant="outline" className={`text-[10px] h-5 ${stream.CodecType === 'video' ? 'border-blue-500/50 text-blue-400' : 'border-green-500/50 text-green-400'}`}>
                                                    {stream.CodecType}
                                                </Badge>
                                                <span className="font-mono">{stream.CodecName.toUpperCase()}</span>
                                                <span className="text-muted-foreground">Stream #{stream.Index}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            ) : task.result && (
                // Default raw JSON view for others
                <div className="mt-1 text-xs bg-black/80 text-green-400 p-2 rounded overflow-auto font-mono w-full max-h-40 whitespace-pre-wrap relative group">
                    {typeof task.result === 'string' ? task.result : JSON.stringify(task.result, null, 2)}
                    <CopyButton
                        text={typeof task.result === 'string' ? task.result : JSON.stringify(task.result, null, 2)}
                        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity bg-zinc-800 hover:bg-zinc-700 text-white"
                    />
                </div>
            )}
        </div>
    );
}

function CopyButton({ text, className }: { text: string; className?: string }) {
    const [copied, setCopied] = useState(false);

    const onCopy = () => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <Button
            variant="ghost"
            size="icon"
            className={`h-6 w-6 ${className || ''}`}
            onClick={onCopy}
            title="Copy to clipboard"
        >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
        </Button>
    );
}

function formatETA(seconds: number): string {
    if (seconds < 60) return `${seconds}s`;
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    if (mins < 60) return `${mins}m ${secs}s`;
    const hours = Math.floor(mins / 60);
    const remainMins = mins % 60;
    return `${hours}h ${remainMins}m`;
}

function formatDuration(seconds: number | undefined): string {
    if (!seconds) return 'N/A';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    if (mins < 60) return `${mins}:${secs.toString().padStart(2, '0')}`;
    const hours = Math.floor(mins / 60);
    const remainMins = mins % 60;
    return `${hours}:${remainMins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

function formatBitrate(bps: number): string {
    if (bps >= 1000000) return `${(bps / 1000000).toFixed(1)} Mbps`;
    if (bps >= 1000) return `${(bps / 1000).toFixed(0)} kbps`;
    return `${bps} bps`;
}
