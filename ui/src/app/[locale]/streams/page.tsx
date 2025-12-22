"use client";

import { Link } from "@/i18n/routing";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchStreams, createStream, Stream, fetchStreamDestinations, updateStreamDestinations, RestreamDestination, fetchPlugins, Plugin } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState, useEffect } from "react";
import {
    Loader2,
    Plus,
    Radio,
    Copy,
    Check,
    Users,
    Signal,
    Clock,
    Eye,
    RefreshCw,
    ExternalLink,
    Share2,
    Trash2,
    Twitch,
    Youtube,
    Tv
} from "lucide-react";
import { useTranslations } from "next-intl";

export default function StreamsPage() {
    const t = useTranslations('streams');
    const commonT = useTranslations('common');
    const { data: streams, isLoading, refetch, isRefetching } = useQuery({
        queryKey: ["streams"],
        queryFn: fetchStreams,
        refetchInterval: 10000,
    });

    const liveCount = streams?.filter(s => s.is_live).length || 0;
    const offlineCount = streams?.filter(s => !s.is_live).length || 0;
    const totalViewers = streams?.reduce((acc, s) => acc + (s.current_viewers || 0), 0) || 0;

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">{t('title')}</h1>
                    <p className="text-muted-foreground mt-1">
                        {t('subtitle')}
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
                    <CreateStreamDialog />
                </div>
            </div>

            {/* Stats */}
            <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
                <StatCard
                    label={t('title')}
                    value={streams?.length || 0}
                    icon={Radio}
                    color="violet"
                />
                <StatCard
                    label={commonT('live')}
                    value={liveCount}
                    icon={Signal}
                    color="rose"
                    pulse={liveCount > 0}
                />
                <StatCard
                    label={commonT('offline')}
                    value={offlineCount}
                    icon={Clock}
                    color="slate"
                />
                <StatCard
                    label={t('viewers')}
                    value={totalViewers}
                    icon={Users}
                    color="cyan"
                />
            </div>

            {/* Streams Grid */}
            {isLoading ? (
                <div className="flex justify-center py-16">
                    <div className="flex flex-col items-center gap-3">
                        <Loader2 className="h-10 w-10 animate-spin text-violet-400" />
                        <span className="text-sm text-muted-foreground">{commonT('loading')}</span>
                    </div>
                </div>
            ) : streams && streams.length > 0 ? (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {streams.map((stream, index) => (
                        <StreamCard
                            key={stream.id}
                            stream={stream}
                            style={{ animationDelay: `${index * 50}ms` }}
                        />
                    ))}
                </div>
            ) : (
                <Card className="card-glow">
                    <CardContent className="flex flex-col items-center justify-center py-20">
                        <Radio className="h-16 w-16 text-muted-foreground/30 mb-4" />
                        <h3 className="text-xl font-semibold mb-2">No Streams Yet</h3>
                        <p className="text-muted-foreground text-center max-w-md mb-6">
                            Create your first stream to start broadcasting. You&apos;ll receive a unique stream key to use with OBS or other streaming software.
                        </p>
                        <CreateStreamDialog />
                    </CardContent>
                </Card>
            )}
        </div>
    );
}

interface StreamCardProps {
    stream: Stream;
    style?: React.CSSProperties;
}

function StreamCard({ stream, style }: StreamCardProps) {
    const t = useTranslations('streams');
    const commonT = useTranslations('common');
    const [copied, setCopied] = useState(false);

    const copyStreamKey = () => {
        navigator.clipboard.writeText(stream.stream_key);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <Card
            className={`card-glow animate-[slide-up_0.3s_ease-out] ${stream.is_live ? 'border-rose-500/30' : ''
                }`}
            style={style}
        >
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${stream.is_live ? 'bg-rose-500/10' : 'bg-muted/50'
                            }`}>
                            {stream.is_live ? (
                                <Signal className="h-5 w-5 text-rose-400 animate-pulse" />
                            ) : (
                                <Radio className="h-5 w-5 text-muted-foreground" />
                            )}
                        </div>
                        <div className="flex-1 min-w-0">
                            <CardTitle className="text-base truncate">
                                {stream.title || "Untitled Stream"}
                            </CardTitle>
                            <CardDescription className="text-xs truncate">
                                {stream.description || "No description"}
                            </CardDescription>
                        </div>
                    </div>
                    <Badge
                        variant="outline"
                        className={stream.is_live ? 'badge-error' : 'bg-muted text-muted-foreground'}
                    >
                        {stream.is_live ? (
                            <span className="flex items-center gap-1.5">
                                <span className="h-2 w-2 rounded-full bg-rose-500 animate-pulse" />
                                {commonT('live').toUpperCase()}
                            </span>
                        ) : (
                            commonT('offline')
                        )}
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="space-y-3">
                    {/* Ingest URL */}
                    <div className="space-y-1.5">
                        <Label className="text-xs text-muted-foreground ml-1">{t('rtmpUrl')}</Label>
                        <div className="flex items-center gap-2">
                            <div className="relative flex-1">
                                <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none">
                                    <Signal className="h-3.5 w-3.5 text-muted-foreground/50" />
                                </div>
                                <Input
                                    readOnly
                                    value={stream.ingest_url || "rtmp://localhost/live"}
                                    className="h-8 text-xs font-mono pl-9 bg-muted/30"
                                />
                            </div>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-muted"
                                onClick={() => {
                                    navigator.clipboard.writeText(stream.ingest_url || "rtmp://localhost/live");
                                }}
                            >
                                <Copy className="h-3.5 w-3.5" />
                            </Button>
                        </div>
                    </div>

                    {/* Stream Key */}
                    <div className="space-y-1.5">
                        <Label className="text-xs text-muted-foreground ml-1">{t('streamKey')}</Label>
                        <div className="flex items-center gap-2">
                            <div className="relative flex-1">
                                <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none">
                                    <Radio className="h-3.5 w-3.5 text-muted-foreground/50" />
                                </div>
                                <Input
                                    readOnly
                                    type="password"
                                    value={stream.stream_key}
                                    className="h-8 text-xs font-mono pl-9 bg-muted/30"
                                />
                            </div>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 hover:bg-muted"
                                onClick={copyStreamKey}
                            >
                                {copied ? (
                                    <Check className="h-3.5 w-3.5 text-emerald-400" />
                                ) : (
                                    <Copy className="h-3.5 w-3.5" />
                                )}
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Live Stats */}
                {stream.is_live && (
                    <div className="grid grid-cols-2 gap-3 mt-4">
                        <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20">
                            <div className="flex items-center gap-2">
                                <Eye className="h-4 w-4 text-rose-400" />
                                <span className="text-xs text-muted-foreground">{t('viewers')}</span>
                            </div>
                            <div className="text-xl font-bold text-rose-400 mt-1">
                                {stream.current_viewers || 0}
                            </div>
                        </div>
                        <div className="p-3 rounded-lg bg-cyan-500/10 border border-cyan-500/20">
                            <div className="flex items-center gap-2">
                                <Signal className="h-4 w-4 text-cyan-400" />
                                <span className="text-xs text-muted-foreground">{commonT('status')}</span>
                            </div>
                            <div className="text-xl font-bold text-cyan-400 mt-1">
                                {commonT('live')}
                            </div>
                        </div>
                    </div>
                )}

                {/* Footer */}
                <div className="pt-3 border-t border-border flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">
                        Created {new Date(stream.created_at).toLocaleDateString()}
                    </span>
                    <div className="flex items-center gap-1">
                        <DestinationsDialog streamId={stream.id} />
                        <Link href={`/streams/${stream.id}/studio`}>
                            <Button variant="ghost" size="sm" className="h-7 text-xs">
                                <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                                Studio
                            </Button>
                        </Link>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function CreateStreamDialog() {
    const t = useTranslations('streams');
    const commonT = useTranslations('common');
    const [open, setOpen] = useState(false);
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const queryClient = useQueryClient();

    const mutation = useMutation({
        mutationFn: () => createStream(title, description),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["streams"] });
            setOpen(false);
            setTitle("");
            setDescription("");
        },
    });

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button className="btn-gradient text-white">
                    <Plus className="mr-2 h-4 w-4" /> {t('createStream')}
                </Button>
            </DialogTrigger>
            <DialogContent className="glass">
                <DialogHeader>
                    <DialogTitle>{t('createStream')}</DialogTitle>
                    <DialogDescription>
                        Set up a new live streaming channel. You&apos;ll receive a unique stream key.
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 pt-4">
                    <div className="space-y-2">
                        <Label htmlFor="title">Title</Label>
                        <Input
                            id="title"
                            placeholder="My Live Stream"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="bg-muted/30"
                        />
                    </div>
                    <div className="space-y-2">
                        <Label htmlFor="description">Description</Label>
                        <Textarea
                            id="description"
                            placeholder="What will you be streaming?"
                            value={description}
                            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setDescription(e.target.value)}
                            className="bg-muted/30 min-h-[100px]"
                        />
                    </div>
                    <div className="flex justify-end gap-2 pt-2">
                        <Button variant="outline" onClick={() => setOpen(false)}>
                            {commonT('cancel')}
                        </Button>
                        <Button
                            onClick={() => mutation.mutate()}
                            disabled={mutation.isPending || !title}
                            className="btn-gradient text-white"
                        >
                            {mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            {t('createStream')}
                        </Button>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}

interface StatCardProps {
    label: string;
    value: number;
    icon: React.ElementType;
    color: "violet" | "rose" | "slate" | "cyan";
    pulse?: boolean;
}

function StatCard({ label, value, icon: Icon, color, pulse }: StatCardProps) {
    const colors = {
        violet: "text-violet-400 bg-violet-500/10 border-violet-500/20",
        rose: "text-rose-400 bg-rose-500/10 border-rose-500/20",
        slate: "text-slate-400 bg-slate-500/10 border-slate-500/20",
        cyan: "text-cyan-400 bg-cyan-500/10 border-cyan-500/20",
    };

    return (
        <div className={`flex items-center gap-3 p-4 rounded-lg border ${colors[color]}`}>
            <div className="relative">
                <Icon className={`h-8 w-8 ${colors[color].split(' ')[0]}`} />
                {pulse && (
                    <span className="absolute -top-1 -right-1 h-3 w-3">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-rose-400 opacity-75" />
                        <span className="relative inline-flex rounded-full h-3 w-3 bg-rose-500" />
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

// Destinations Dialog for managing restream targets
interface DestinationsDialogProps {
    streamId: string;
}

function DestinationsDialog({ streamId }: DestinationsDialogProps) {
    const t = useTranslations('common'); // Simplified for now
    const [open, setOpen] = useState(false);
    const [destinations, setDestinations] = useState<RestreamDestination[]>([]);
    const [newPluginId, setNewPluginId] = useState("");
    const [newAccessToken, setNewAccessToken] = useState("");
    const queryClient = useQueryClient();

    // Fetch current destinations
    const { data: currentDestinations, isLoading: loadingDests } = useQuery({
        queryKey: ["stream-destinations", streamId],
        queryFn: () => fetchStreamDestinations(streamId),
        enabled: open,
    });

    // Fetch available publisher plugins
    const { data: plugins } = useQuery({
        queryKey: ["plugins"],
        queryFn: fetchPlugins,
        enabled: open,
    });

    const publisherPlugins = plugins?.filter(p => p.type === "publisher" && p.is_enabled) || [];

    // Sync destinations when data loads
    useEffect(() => {
        if (currentDestinations) {
            setDestinations(currentDestinations);
        }
    }, [currentDestinations]);

    // Update mutation
    const updateMutation = useMutation({
        mutationFn: () => updateStreamDestinations(streamId, destinations),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["stream-destinations", streamId] });
            setOpen(false);
        },
    });

    const addDestination = () => {
        if (!newPluginId) return;
        setDestinations([
            ...destinations,
            { plugin_id: newPluginId, access_token: newAccessToken, enabled: true }
        ]);
        setNewPluginId("");
        setNewAccessToken("");
    };

    const removeDestination = (index: number) => {
        setDestinations(destinations.filter((_, i) => i !== index));
    };

    const toggleDestination = (index: number) => {
        setDestinations(destinations.map((d, i) =>
            i === index ? { ...d, enabled: !d.enabled } : d
        ));
    };

    const getPlatformIcon = (pluginId: string) => {
        if (pluginId.includes("twitch")) return <Twitch className="h-4 w-4 text-purple-400" />;
        if (pluginId.includes("youtube")) return <Youtube className="h-4 w-4 text-red-400" />;
        return <Tv className="h-4 w-4 text-muted-foreground" />;
    };

    const getPlatformName = (pluginId: string) => {
        if (pluginId.includes("twitch")) return "Twitch";
        if (pluginId.includes("youtube")) return "YouTube";
        if (pluginId.includes("kick")) return "Kick";
        if (pluginId.includes("rumble")) return "Rumble";
        return pluginId.replace("publisher-", "").replace("-", " ");
    };

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
                <Button variant="ghost" size="sm" className="h-7 text-xs">
                    <Share2 className="h-3.5 w-3.5 mr-1.5" />
                    Restream
                </Button>
            </DialogTrigger>
            <DialogContent className="glass max-w-lg">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <Share2 className="h-5 w-5 text-violet-400" />
                        Restream Destinations
                    </DialogTitle>
                    <DialogDescription>
                        Configure platforms to automatically restream to when you go live.
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 py-4">
                    {/* Current Destinations */}
                    {loadingDests ? (
                        <div className="flex justify-center py-4">
                            <Loader2 className="h-6 w-6 animate-spin text-violet-400" />
                        </div>
                    ) : destinations.length > 0 ? (
                        <div className="space-y-2">
                            <Label className="text-xs text-muted-foreground">Active Destinations</Label>
                            {destinations.map((dest, index) => (
                                <div
                                    key={index}
                                    className={`flex items-center justify-between p-3 rounded-lg border ${dest.enabled ? 'bg-muted/30 border-violet-500/30' : 'bg-muted/10 border-border opacity-60'
                                        }`}
                                >
                                    <div className="flex items-center gap-3">
                                        {getPlatformIcon(dest.plugin_id)}
                                        <div>
                                            <div className="font-medium text-sm">{getPlatformName(dest.plugin_id)}</div>
                                            <div className="text-xs text-muted-foreground">
                                                {dest.access_token ? "Token configured" : "No token"}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Switch
                                            checked={dest.enabled}
                                            onCheckedChange={() => toggleDestination(index)}
                                        />
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-8 w-8 text-red-400 hover:text-red-300"
                                            onClick={() => removeDestination(index)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-6 text-muted-foreground">
                            <Share2 className="h-12 w-12 mx-auto mb-3 opacity-30" />
                            <p className="text-sm">No restream destinations configured</p>
                        </div>
                    )}

                    {/* Add New Destination */}
                    <div className="pt-4 border-t border-border space-y-3">
                        <Label className="text-xs text-muted-foreground">Add Destination</Label>
                        <div className="flex gap-2">
                            <Select value={newPluginId} onValueChange={setNewPluginId}>
                                <SelectTrigger className="flex-1 bg-muted/30">
                                    <SelectValue placeholder="Select platform..." />
                                </SelectTrigger>
                                <SelectContent>
                                    {publisherPlugins.length > 0 ? (
                                        publisherPlugins.map(p => (
                                            <SelectItem key={p.id} value={p.id}>
                                                <span className="flex items-center gap-2">
                                                    {getPlatformIcon(p.id)}
                                                    {getPlatformName(p.id)}
                                                </span>
                                            </SelectItem>
                                        ))
                                    ) : (
                                        <SelectItem value="_none" disabled>
                                            No publisher plugins available
                                        </SelectItem>
                                    )}
                                </SelectContent>
                            </Select>
                        </div>
                        {newPluginId && (
                            <div className="space-y-2">
                                <Input
                                    type="password"
                                    placeholder="Access Token / Stream Key"
                                    value={newAccessToken}
                                    onChange={(e) => setNewAccessToken(e.target.value)}
                                    className="bg-muted/30"
                                />
                                <Button
                                    onClick={addDestination}
                                    className="w-full btn-gradient text-white"
                                    disabled={!newPluginId}
                                >
                                    <Plus className="h-4 w-4 mr-2" />
                                    Add Destination
                                </Button>
                            </div>
                        )}
                    </div>
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => setOpen(false)}>
                        {t('cancel')}
                    </Button>
                    <Button
                        onClick={() => updateMutation.mutate()}
                        disabled={updateMutation.isPending}
                        className="btn-gradient text-white"
                    >
                        {updateMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {t('save')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}

// ... StatCard remains same
