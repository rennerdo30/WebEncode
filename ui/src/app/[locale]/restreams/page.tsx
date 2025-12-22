"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchRestreams, createRestream, startRestream, stopRestream, RestreamJob, fetchPlugins } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useState } from "react";
import {
    Loader2,
    Plus,
    Play,
    Square,
    Trash2,
    Radio,
    RefreshCw,
    Share2,
    Twitch,
    Youtube,
    Tv,
    FileVideo,
    LayoutDashboard,
    Check
} from "lucide-react";
import { FileBrowser } from "@/components/file-browser";

export default function RestreamsPage() {
    const { data: restreams, isLoading, refetch, isRefetching } = useQuery({
        queryKey: ["restreams"],
        queryFn: fetchRestreams,
        refetchInterval: 5000,
    });

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Restreams</h1>
                    <p className="text-muted-foreground mt-1">
                        Manage file-based simulcasting and scheduled streams
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        onClick={() => refetch()}
                        disabled={isRefetching}
                        className="border-border"
                    >
                        <RefreshCw className={`h-4 w-4 mr-2 ${isRefetching ? 'animate-spin' : ''}`} />
                        Refresh
                    </Button>
                    <CreateRestreamDialog />
                </div>
            </div>

            {isLoading ? (
                <div className="flex justify-center py-16">
                    <div className="flex flex-col items-center gap-3">
                        <Loader2 className="h-10 w-10 animate-spin text-violet-400" />
                        <span className="text-sm text-muted-foreground">Loading restreams...</span>
                    </div>
                </div>
            ) : restreams && restreams.length > 0 ? (
                <div className="rounded-lg border bg-card">
                    <Table>
                        <TableHeader>
                            <TableRow className="hover:bg-transparent">
                                <TableHead>Title</TableHead>
                                <TableHead>Input Source</TableHead>
                                <TableHead>Destinations</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {restreams.map((restream, index) => (
                                <RestreamRow
                                    key={restream.id}
                                    restream={restream}
                                    style={{ animationDelay: `${index * 50}ms` }}
                                />
                            ))}
                        </TableBody>
                    </Table>
                </div>
            ) : (
                <Card className="card-glow">
                    <CardContent className="flex flex-col items-center justify-center py-20">
                        <Share2 className="h-16 w-16 text-muted-foreground/30 mb-4" />
                        <h3 className="text-xl font-semibold mb-2">No Restreams Configured</h3>
                        <p className="text-muted-foreground text-center max-w-md mb-6">
                            Create a restream job to broadcast pre-recorded files or URLs to multiple platforms simultaneously.
                        </p>
                        <CreateRestreamDialog />
                    </CardContent>
                </Card>
            )}
        </div>
    );
}

interface RestreamRowProps {
    restream: RestreamJob;
    style?: React.CSSProperties;
}

function RestreamRow({ restream, style }: RestreamRowProps) {
    const queryClient = useQueryClient();

    const startMutation = useMutation({
        mutationFn: () => startRestream(restream.id),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["restreams"] }),
    });

    const stopMutation = useMutation({
        mutationFn: () => stopRestream(restream.id),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["restreams"] }),
    });

    const statusColors: Record<string, string> = {
        streaming: "bg-green-500/10 text-green-500 border-green-500/30",
        stopped: "bg-slate-500/10 text-slate-500 border-slate-500/30",
        error: "bg-red-500/10 text-red-500 border-red-500/30",
        queued: "bg-blue-500/10 text-blue-500 border-blue-500/30",
    };

    const StatusIcon = {
        streaming: Radio,
        stopped: Square,
        error: Trash2, // Or alert icon
        queued: Loader2,
    }[restream.status] || Radio;

    const getInputIcon = (type: string | null) => {
        if (type === 'file') return <FileVideo className="h-4 w-4" />;
        return <LayoutDashboard className="h-4 w-4" />;
    };

    return (
        <TableRow className="group animate-[slide-up_0.3s_ease-out]" style={style}>
            <TableCell className="font-medium">
                <div className="flex items-center gap-2">
                    <div className={`p-1.5 rounded-md ${restream.status === 'streaming' ? 'bg-green-500/10' : 'bg-muted'
                        }`}>
                        <StatusIcon className={`h-4 w-4 ${restream.status === 'streaming' ? 'text-green-500 animate-pulse' : 'text-muted-foreground'
                            }`} />
                    </div>
                    {restream.title || "Untitled Restream"}
                </div>
            </TableCell>
            <TableCell>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    {getInputIcon(restream.input_type)}
                    <span className="truncate max-w-[200px]" title={restream.input_url || ""}>
                        {restream.input_type === 'file'
                            ? restream.input_url?.split('/').pop()
                            : restream.input_url}
                    </span>
                </div>
            </TableCell>
            <TableCell>
                <div className="flex flex-wrap gap-1">
                    {restream.output_destinations?.map((dest, i) => (
                        <Badge
                            key={i}
                            variant="outline"
                            className={`text-xs gap-1 ${dest.enabled ? 'bg-muted/50' : 'opacity-50'
                                }`}
                        >
                            {dest.platform.includes('twitch') && <Twitch className="h-3 w-3 text-purple-400" />}
                            {dest.platform.includes('youtube') && <Youtube className="h-3 w-3 text-red-400" />}
                            {!dest.platform.includes('twitch') && !dest.platform.includes('youtube') && (
                                <Tv className="h-3 w-3 text-muted-foreground" />
                            )}
                            {dest.platform.replace('publisher-', '')}
                        </Badge>
                    ))}
                </div>
            </TableCell>
            <TableCell>
                <Badge variant="outline" className={statusColors[restream.status] || ""}>
                    {restream.status.charAt(0).toUpperCase() + restream.status.slice(1)}
                </Badge>
            </TableCell>
            <TableCell className="text-right">
                <div className="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    {restream.status === "streaming" ? (
                        <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => stopMutation.mutate()}
                            disabled={stopMutation.isPending}
                            className="h-8"
                        >
                            <Square className="h-3.5 w-3.5 mr-1.5" />
                            Stop
                        </Button>
                    ) : (
                        <Button
                            variant="default"
                            size="sm"
                            onClick={() => startMutation.mutate()}
                            disabled={startMutation.isPending}
                            className="h-8 bg-green-600 hover:bg-green-700 text-white"
                        >
                            <Play className="h-3.5 w-3.5 mr-1.5" />
                            Start
                        </Button>
                    )}
                </div>
            </TableCell>
        </TableRow>
    );
}

function CreateRestreamDialog() {
    const [open, setOpen] = useState(false);
    const [title, setTitle] = useState("");
    const [inputType, setInputType] = useState("file");
    const [inputUrl, setInputUrl] = useState("");
    const [destinations, setDestinations] = useState<{ platform: string; url: string; enabled: boolean }[]>([]);

    // New destination state
    const [newDestPlatform, setNewDestPlatform] = useState("");
    const [newDestUrl, setNewDestUrl] = useState("");

    const queryClient = useQueryClient();

    // Fetch plugins to populate platform list
    const { data: plugins } = useQuery({
        queryKey: ["plugins"],
        queryFn: fetchPlugins,
        enabled: open,
    });

    const publisherPlugins = plugins?.filter(p => p.type === "publisher" && p.is_enabled) || [];

    const mutation = useMutation({
        mutationFn: () => createRestream({
            title,
            input_type: inputType,
            input_url: inputUrl,
            output_destinations: destinations,
        }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["restreams"] });
            setOpen(false);
            resetForm();
        },
    });

    const resetForm = () => {
        setTitle("");
        setInputType("file");
        setInputUrl("");
        setDestinations([]);
        setNewDestPlatform("");
        setNewDestUrl("");
    };

    const addDestination = () => {
        if (!newDestPlatform || !newDestUrl) return;
        setDestinations([...destinations, { platform: newDestPlatform, url: newDestUrl, enabled: true }]);
        setNewDestPlatform("");
        setNewDestUrl("");
    };

    const removeDestination = (index: number) => {
        setDestinations(destinations.filter((_, i) => i !== index));
    };

    const getPlatformIcon = (pluginId: string) => {
        if (pluginId.includes("twitch")) return <Twitch className="h-4 w-4 text-purple-400" />;
        if (pluginId.includes("youtube")) return <Youtube className="h-4 w-4 text-red-400" />;
        return <Tv className="h-4 w-4 text-muted-foreground" />;
    };

    const getPlatformName = (pluginId: string) => {
        return pluginId.replace("publisher-", "").replace("-", " ").replace(/\b\w/g, c => c.toUpperCase());
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button className="btn-gradient text-white">
                    <Plus className="mr-2 h-4 w-4" /> New Restream
                </Button>
            </DialogTrigger>
            <DialogContent className="glass max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create Restream Job</DialogTitle>
                    <DialogDescription>
                        Stream a file or URL to multiple destinations simultaneously.
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-6 pt-4">
                    {/* Basic Info */}
                    <div className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="title">Job Title</Label>
                            <Input
                                id="title"
                                placeholder="My Scheduled Stream"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                className="bg-muted/30"
                            />
                        </div>

                        {/* Input Source */}
                        <div className="space-y-2">
                            <Label>Input Source</Label>
                            <Tabs defaultValue="file" value={inputType} onValueChange={setInputType} className="w-full">
                                <TabsList className="grid w-full grid-cols-3">
                                    <TabsTrigger value="file">File Browser</TabsTrigger>
                                    <TabsTrigger value="upload">Upload</TabsTrigger>
                                    <TabsTrigger value="rtmp">RTMP Pull</TabsTrigger>
                                </TabsList>
                                <TabsContent value="file" className="mt-4 border rounded-md p-4 bg-muted/10">
                                    <Label className="text-xs text-muted-foreground mb-2 block">Select Video File</Label>
                                    <div className="flex flex-col">
                                        <FileBrowser
                                            onSelect={(path) => setInputUrl(path)}
                                            mediaOnly={true}
                                        />
                                    </div>
                                    {inputUrl && (
                                        <div className="mt-2 text-sm flex items-center gap-2 text-green-400">
                                            <FileVideo className="h-4 w-4" />
                                            Selected: <span className="font-mono text-xs">{inputUrl}</span>
                                        </div>
                                    )}
                                </TabsContent>
                                <TabsContent value="upload" className="mt-4 border rounded-md p-6 bg-muted/10">
                                    <div className="flex flex-col items-center justify-center space-y-4">
                                        <div className="p-4 rounded-full bg-violet-500/10">
                                            <Share2 className="h-8 w-8 text-violet-400" />
                                        </div>
                                        <div className="text-center">
                                            <h3 className="text-sm font-medium">Upload Video File</h3>
                                            <p className="text-xs text-muted-foreground mt-1">
                                                Drag & drop or click to upload (up to 10GB)
                                            </p>
                                        </div>
                                        <Input
                                            type="file"
                                            className="hidden"
                                            id="file-upload"
                                            accept="video/*"
                                            onChange={async (e) => {
                                                const file = e.target.files?.[0];
                                                if (!file) return;

                                                const formData = new FormData();
                                                formData.append("file", file);
                                                // Use S3 plugin for uploads if generic, or let backend decide
                                                // Default backend logic prefers S3

                                                try {
                                                    const res = await fetch("/api/v1/files/upload", {
                                                        method: "POST",
                                                        body: formData,
                                                    });
                                                    if (!res.ok) throw new Error("Upload failed");
                                                    const data = await res.json();
                                                    setInputUrl(data.url);
                                                } catch (err) {
                                                    console.error(err);
                                                    // TODO: Show error toast
                                                }
                                            }}
                                        />
                                        <Button variant="outline" onClick={() => document.getElementById("file-upload")?.click()}>
                                            Select File
                                        </Button>
                                    </div>
                                    {inputUrl && inputUrl.startsWith("s3://") && (
                                        <div className="mt-4 pt-4 border-t border-border w-full">
                                            <div className="text-sm flex items-center gap-2 text-green-400">
                                                <Check className="h-4 w-4" />
                                                Uploaded: <span className="font-mono text-xs truncate max-w-[200px]">{inputUrl}</span>
                                            </div>
                                        </div>
                                    )}
                                </TabsContent>
                                <TabsContent value="rtmp" className="mt-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="rtmp-url">RTMP Source URL</Label>
                                        <Input
                                            id="rtmp-url"
                                            placeholder="rtmp://server.com/live/stream"
                                            value={inputUrl}
                                            onChange={(e) => setInputUrl(e.target.value)}
                                            className="bg-muted/30 font-mono"
                                        />
                                    </div>
                                </TabsContent>
                            </Tabs>
                        </div>
                    </div>

                    {/* Destinations */}
                    <div className="space-y-3 pt-4 border-t border-border">
                        <Label>Destinations</Label>

                        {/* List of added destinations */}
                        {destinations.length > 0 && (
                            <div className="space-y-2 mb-4">
                                {destinations.map((dest, colIndex) => (
                                    <div key={colIndex} className="flex items-center justify-between p-3 rounded-lg border bg-muted/20">
                                        <div className="flex items-center gap-3">
                                            {getPlatformIcon(dest.platform)}
                                            <div>
                                                <div className="font-medium text-sm">{getPlatformName(dest.platform)}</div>
                                                <div className="text-xs text-muted-foreground max-w-[250px] truncate font-mono">
                                                    {dest.url}
                                                </div>
                                            </div>
                                        </div>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-red-400 hover:text-red-300"
                                            onClick={() => removeDestination(colIndex)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        )}

                        {/* Add new destination */}
                        <div className="grid gap-3 sm:grid-cols-[1fr_2fr_auto] items-end">
                            <div className="space-y-2">
                                <Label className="text-xs text-muted-foreground">Platform</Label>
                                <Select value={newDestPlatform} onValueChange={setNewDestPlatform}>
                                    <SelectTrigger className="bg-muted/30">
                                        <SelectValue placeholder="Select..." />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {publisherPlugins.map(p => (
                                            <SelectItem key={p.id} value={p.id}>
                                                <div className="flex items-center gap-2">
                                                    {getPlatformIcon(p.id)}
                                                    {getPlatformName(p.id)}
                                                </div>
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label className="text-xs text-muted-foreground">RTMP URL / Key</Label>
                                <Input
                                    placeholder="rtmp://..."
                                    value={newDestUrl}
                                    onChange={(e) => setNewDestUrl(e.target.value)}
                                    className="bg-muted/30 font-mono"
                                />
                            </div>
                            <Button
                                onClick={addDestination}
                                disabled={!newDestPlatform || !newDestUrl}
                                variant="secondary"
                            >
                                <Plus className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                </div>

                <DialogFooter className="mt-6">
                    <Button variant="outline" onClick={() => setOpen(false)}>
                        Cancel
                    </Button>
                    <Button
                        onClick={() => mutation.mutate()}
                        disabled={mutation.isPending || !title || !inputUrl || destinations.length === 0}
                        className="btn-gradient text-white"
                    >
                        {mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        Create Restream
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
