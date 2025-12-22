"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchJobs } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Link } from "@/i18n/routing";
import {
    Loader2,
    Plus,
    Filter,
    Search,
    Film,
    Clock,
    CheckCircle2,
    XCircle,
    PlayCircle,
    RefreshCw
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { useTranslations } from "next-intl";

export default function JobsPage() {
    const t = useTranslations('jobs');
    const commonT = useTranslations('common');
    const { data: jobs, isLoading, refetch, isRefetching } = useQuery({
        queryKey: ["jobs"],
        queryFn: () => fetchJobs(),
        refetchInterval: 5000,
    });

    const [searchTerm, setSearchTerm] = useState("");

    const filteredJobs = jobs?.filter(job =>
        job.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
        job.source_url?.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const statusCounts = {
        processing: jobs?.filter(j => j.status === 'processing' || j.status === 'stitching').length || 0,
        queued: jobs?.filter(j => j.status === 'queued').length || 0,
        completed: jobs?.filter(j => j.status === 'completed').length || 0,
        failed: jobs?.filter(j => j.status === 'failed').length || 0,
    };

    return (
        <div className="space-y-6 animate-[fade-in_0.3s_ease-out]">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">{commonT('jobs')}</h1>
                    <p className="text-muted-foreground mt-1">
                        {t('subtitle')}
                    </p>
                </div>
                <Link href="/jobs/new">
                    <Button className="btn-gradient text-white">
                        <Plus className="mr-2 h-4 w-4" /> {t('createJob')}
                    </Button>
                </Link>
            </div>

            {/* Stats Cards */}
            <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
                <StatusStatCard
                    label="Processing"
                    count={statusCounts.processing}
                    icon={PlayCircle}
                    color="blue"
                    pulse={statusCounts.processing > 0}
                />
                <StatusStatCard
                    label="Queued"
                    count={statusCounts.queued}
                    icon={Clock}
                    color="amber"
                />
                <StatusStatCard
                    label="Completed"
                    count={statusCounts.completed}
                    icon={CheckCircle2}
                    color="emerald"
                />
                <StatusStatCard
                    label="Failed"
                    count={statusCounts.failed}
                    icon={XCircle}
                    color="red"
                />
            </div>

            {/* Main Table Card */}
            <Card className="card-glow">
                <CardHeader className="border-b border-border">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <Film className="h-5 w-5 text-violet-400" />
                                {t('title')}
                            </CardTitle>
                            <CardDescription>
                                {jobs?.length || 0} total jobs in your queue
                            </CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder={commonT('search')}
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-9 w-64 bg-muted/30"
                                />
                            </div>
                            <Button variant="outline" size="icon" className="border-border">
                                <Filter className="h-4 w-4" />
                            </Button>
                            <Button
                                variant="outline"
                                size="icon"
                                className="border-border"
                                onClick={() => refetch()}
                                disabled={isRefetching}
                            >
                                <RefreshCw className={`h-4 w-4 ${isRefetching ? 'animate-spin' : ''}`} />
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow className="hover:bg-transparent border-border">
                                <TableHead className="w-[120px]">{t('jobId')}</TableHead>
                                <TableHead>{t('input')}</TableHead>
                                <TableHead className="w-[120px]">{commonT('status')}</TableHead>
                                <TableHead className="w-[140px]">{t('progress')}</TableHead>
                                <TableHead className="w-[160px]">{t('time')}</TableHead>
                                <TableHead className="w-[80px] text-right">{commonT('actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {isLoading ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="h-32 text-center">
                                        <div className="flex flex-col items-center gap-2">
                                            <Loader2 className="h-8 w-8 animate-spin text-violet-400" />
                                            <span className="text-sm text-muted-foreground">{commonT('loading')}</span>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ) : filteredJobs && filteredJobs.length > 0 ? (
                                filteredJobs.map((job) => (
                                    <TableRow key={job.id} className="table-row-hover border-border group">
                                        <TableCell>
                                            <code className="mono-highlight text-xs">
                                                {job.id.slice(0, 8)}
                                            </code>
                                        </TableCell>
                                        <TableCell>
                                            <span className="text-sm truncate block max-w-[300px]" title={job.source_url}>
                                                {job.source_url || "N/A"}
                                            </span>
                                        </TableCell>
                                        <TableCell>
                                            <StatusBadge status={job.status} />
                                        </TableCell>
                                        <TableCell>
                                            <JobProgress status={job.status} progress={job.progress_pct || 0} />
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex flex-col">
                                                <span className="text-sm">
                                                    {new Date(job.created_at).toLocaleDateString()}
                                                </span>
                                                <span className="text-xs text-muted-foreground">
                                                    {new Date(job.created_at).toLocaleTimeString()}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                asChild
                                                className="opacity-0 group-hover:opacity-100 transition-opacity"
                                            >
                                                <Link href={`/jobs/${job.id}`}>{commonT('view')}</Link>
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))
                            ) : (
                                <TableRow>
                                    <TableCell colSpan={6}>
                                        <div className="empty-state py-16 text-center">
                                            <Film className="h-12 w-12 text-muted-foreground/50 mx-auto mb-3" />
                                            <p className="text-muted-foreground mb-1">No jobs found</p>
                                            <p className="text-sm text-muted-foreground/70 mb-4">
                                                {searchTerm ? "Try adjusting your search" : "Create your first encoding job to get started"}
                                            </p>
                                            {!searchTerm && (
                                                <Link href="/jobs/new">
                                                    <Button className="btn-gradient text-white">
                                                        <Plus className="mr-2 h-4 w-4" /> {t('createJob')}
                                                    </Button>
                                                </Link>
                                            )}
                                        </div>
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}

function StatusBadge({ status }: { status: string }) {
    const config: Record<string, { bg: string; text: string; dot: string }> = {
        queued: { bg: "bg-amber-500/10", text: "text-amber-400", dot: "bg-amber-400" },
        processing: { bg: "bg-blue-500/10", text: "text-blue-400", dot: "bg-blue-400" },
        stitching: { bg: "bg-indigo-500/10", text: "text-indigo-400", dot: "bg-indigo-400" },
        completed: { bg: "bg-emerald-500/10", text: "text-emerald-400", dot: "bg-emerald-400" },
        failed: { bg: "bg-red-500/10", text: "text-red-400", dot: "bg-red-400" },
        cancelled: { bg: "bg-muted", text: "text-muted-foreground", dot: "bg-muted-foreground" },
    };

    const { bg, text, dot } = config[status] || config.queued;

    return (
        <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${bg} ${text}`}>
            <span className={`h-1.5 w-1.5 rounded-full ${dot} ${status === 'processing' ? 'animate-pulse' : ''}`} />
            {status}
        </span>
    );
}

function JobProgress({ status, progress }: { status: string; progress: number }) {
    if (status === 'completed') {
        return <span className="text-sm text-emerald-400">100%</span>;
    }
    if (status === 'failed' || status === 'cancelled') {
        return <span className="text-sm text-muted-foreground">—</span>;
    }
    if (status === 'queued') {
        return <span className="text-sm text-muted-foreground">Waiting...</span>;
    }

    return (
        <div className="flex items-center gap-2">
            <Progress value={progress} className="h-2 w-16" />
            <span className="text-sm text-muted-foreground">{progress}%</span>
        </div>
    );
}

interface StatusStatCardProps {
    label: string;
    count: number;
    icon: React.ElementType;
    color: "blue" | "amber" | "emerald" | "red";
    pulse?: boolean;
}

function StatusStatCard({ label, count, icon: Icon, color, pulse }: StatusStatCardProps) {
    const colors = {
        blue: "text-blue-400 bg-blue-500/10 border-blue-500/20",
        amber: "text-amber-400 bg-amber-500/10 border-amber-500/20",
        emerald: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
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
                <div className={`text-2xl font-bold ${colors[color].split(' ')[0]}`}>{count}</div>
                <div className="text-sm text-muted-foreground">{label}</div>
            </div>
        </div>
    );
}
