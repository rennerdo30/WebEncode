"use client";

import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Activity, Server, Film, Radio, Loader2, TrendingUp, Clock, Zap } from "lucide-react";
import { fetchDashboardStats, fetchJobs } from "@/lib/api";
import { Link } from "@/i18n/routing";
import { Progress } from "@/components/ui/progress";
import { useTranslations } from "next-intl";

export default function Dashboard() {
    const t = useTranslations('dashboard');
    const commonT = useTranslations('common');
    const tJobs = useTranslations('jobs');
    const tWorkers = useTranslations('workers');
    const tStreams = useTranslations('streams');

    const { data: stats, isLoading: statsLoading } = useQuery({
        queryKey: ["dashboard-stats"],
        queryFn: fetchDashboardStats,
        refetchInterval: 10000,
    });

    const { data: recentJobs } = useQuery({
        queryKey: ["recent-jobs"],
        queryFn: () => fetchJobs(5, 0),
        refetchInterval: 10000,
    });

    return (
        <div className="space-y-8 animate-[fade-in_0.3s_ease-out]">
            {/* Hero Section */}
            <div className="relative overflow-hidden rounded-xl bg-violet-900/10 p-8 border border-violet-500/20">
                <div className="absolute top-0 right-0 w-96 h-96 bg-violet-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
                <div className="relative z-10">
                    <h1 className="text-4xl font-bold tracking-tight">
                        {t.rich('welcome', {
                            span: (chunks) => <span className="text-gradient">WebEncode</span>
                        })}
                    </h1>
                    <p className="mt-2 text-muted-foreground max-w-xl">
                        {tJobs('subtitle')}
                    </p>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <StatsCard
                    title={t('activeJobs')}
                    value={statsLoading ? "-" : String(stats?.activeJobs ?? 0)}
                    icon={Activity}
                    loading={statsLoading}
                    trend="+12%"
                    trendUp
                    color="violet"
                />
                <StatsCard
                    title={t('workersOnline')}
                    value={statsLoading ? "-" : String(stats?.workersOnline ?? 0)}
                    icon={Server}
                    loading={statsLoading}
                    subtitle="Ready to process"
                    color="cyan"
                />
                <StatsCard
                    title={t('completedJobs')}
                    value={statsLoading ? "-" : String(stats?.completedJobs ?? 0)}
                    icon={Film}
                    loading={statsLoading}
                    trend="+5 today"
                    color="emerald"
                />
                <StatsCard
                    title={t('liveStreams')}
                    value={statsLoading ? "-" : String(stats?.liveStreams ?? 0)}
                    icon={Radio}
                    loading={statsLoading}
                    subtitle="Broadcasting now"
                    color="rose"
                    pulse={(stats?.liveStreams ?? 0) > 0}
                />
            </div>

            {/* Main Content Grid */}
            <div className="grid gap-6 lg:grid-cols-3">
                {/* Recent Jobs - 2 columns */}
                <Card className="lg:col-span-2 card-glow">
                    <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="flex items-center gap-2">
                                    <Clock className="h-5 w-5 text-violet-400" />
                                    {t('recentJobs')}
                                </CardTitle>
                                <CardDescription className="mt-1">
                                    {t('recentJobs')}
                                </CardDescription>
                            </div>
                            <Link
                                href="/jobs"
                                className="text-sm font-medium text-violet-400 hover:text-violet-300 transition-colors"
                            >
                                {commonT('view')} All →
                            </Link>
                        </div>
                    </CardHeader>
                    <CardContent>
                        {recentJobs && recentJobs.length > 0 ? (
                            <div className="space-y-3">
                                {recentJobs.map((job, index) => (
                                    <Link
                                        key={job.id}
                                        href={`/jobs/${job.id}`}
                                        className="flex items-center justify-between p-4 rounded-lg bg-muted/30 hover:bg-muted/50 border border-border/50 hover:border-violet-500/30 transition-all duration-200 group"
                                        style={{ animationDelay: `${index * 50}ms` }}
                                    >
                                        <div className="flex flex-col gap-1">
                                            <div className="flex items-center gap-2">
                                                <span className="font-mono text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">
                                                    {job.id.slice(0, 8)}
                                                </span>
                                                <StatusBadge status={job.status} />
                                            </div>
                                            <span className="text-sm text-foreground/80 truncate max-w-[400px] group-hover:text-foreground transition-colors">
                                                {job.source_url}
                                            </span>
                                        </div>
                                        <div className="flex items-center gap-4">
                                            {job.progress_pct !== undefined && job.status === 'processing' && (
                                                <div className="w-24">
                                                    <Progress value={job.progress_pct} className="h-2" />
                                                    <span className="text-xs text-muted-foreground mt-1">
                                                        {job.progress_pct}%
                                                    </span>
                                                </div>
                                            )}
                                            <span className="text-xs text-muted-foreground">
                                                {new Date(job.created_at).toLocaleDateString()}
                                            </span>
                                        </div>
                                    </Link>
                                ))}
                            </div>
                        ) : (
                            <div className="empty-state py-12 text-center">
                                <Film className="h-12 w-12 text-muted-foreground/50 mx-auto mb-3" />
                                <p className="text-muted-foreground">{t('noJobs')}</p>
                                <Link
                                    href="/jobs/new"
                                    className="inline-flex items-center mt-4 px-4 py-2 rounded-lg bg-violet-600 text-white text-sm font-medium"
                                >
                                    {t('createFirstJob')}
                                </Link>
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* System Health - 1 column */}
                <Card className="card-glow">
                    <CardHeader className="pb-3">
                        <CardTitle className="flex items-center gap-2">
                            <Zap className="h-5 w-5 text-cyan-400" />
                            {t('systemHealth')}
                        </CardTitle>
                        <CardDescription>
                            {t('systemHealth')}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <HealthItem label={t('apiServer')} status="healthy" detail="8ms latency" />
                        <HealthItem label={t('database')} status="healthy" detail="PostgreSQL 17" />
                        <HealthItem label="Message Bus" status="healthy" detail="NATS JetStream" />
                        <HealthItem label={t('storage')} status="healthy" detail="SeaweedFS" />
                        <HealthItem label="Live Plugin" status="healthy" detail="MediaMTX" />

                        <div className="pt-4 border-t border-border">
                            <div className="flex items-center justify-between text-sm">
                                <span className="text-muted-foreground">{commonT('status')}</span>
                                <span className="flex items-center gap-2 text-emerald-400 font-medium">
                                    <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
                                    {t('operational')}
                                </span>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Quick Actions */}
            <div className="grid gap-4 md:grid-cols-3">
                <QuickActionCard
                    title={t('newJob')}
                    description={t('newJobDesc')}
                    href="/jobs/new"
                    icon={Film}
                    className="bg-violet-600"
                />
                <QuickActionCard
                    title={t('createStream')}
                    description={t('createStreamDesc')}
                    href="/streams/new"
                    icon={Radio}
                    className="bg-rose-600"
                />
                <QuickActionCard
                    title={t('manageProfiles')}
                    description={t('manageProfilesDesc')}
                    href="/profiles"
                    icon={TrendingUp}
                    className="bg-cyan-600"
                />
            </div>
        </div>
    );
}

interface StatsCardProps {
    title: string;
    value: string;
    icon: React.ElementType;
    loading?: boolean;
    trend?: string;
    trendUp?: boolean;
    subtitle?: string;
    color: "violet" | "cyan" | "emerald" | "rose";
    pulse?: boolean;
}

function StatsCard({ title, value, icon: Icon, loading, trend, trendUp, subtitle, color, pulse }: StatsCardProps) {
    const colorClasses = {
        violet: "bg-violet-500/10 border-violet-500/30 text-violet-400",
        cyan: "bg-cyan-500/10 border-cyan-500/30 text-cyan-400",
        emerald: "bg-emerald-500/10 border-emerald-500/30 text-emerald-400",
        rose: "bg-rose-500/10 border-rose-500/30 text-rose-400",
    };

    return (
        <Card className={`relative overflow-hidden ${colorClasses[color]} transition-all duration-300 hover:scale-[1.02]`}>
            {pulse && (
                <div className="absolute top-3 right-3">
                    <span className="flex h-3 w-3">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-rose-400 opacity-75" />
                        <span className="relative inline-flex rounded-full h-3 w-3 bg-rose-500" />
                    </span>
                </div>
            )}
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
                <Icon className={`h-5 w-5 ${colorClasses[color].split(' ').pop()}`} />
            </CardHeader>
            <CardContent>
                {loading ? (
                    <Loader2 className="h-8 w-8 animate-spin" />
                ) : (
                    <>
                        <div className="text-3xl font-bold tracking-tight">{value}</div>
                        {trend && (
                            <p className={`text-xs mt-1 ${trendUp ? 'text-emerald-400' : 'text-muted-foreground'}`}>
                                {trendUp && <TrendingUp className="inline h-3 w-3 mr-1" />}
                                {trend}
                            </p>
                        )}
                        {subtitle && (
                            <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
                        )}
                    </>
                )}
            </CardContent>
        </Card>
    );
}

function StatusBadge({ status }: { status: string }) {
    const styles: Record<string, string> = {
        queued: "badge-warning",
        processing: "badge-info",
        stitching: "badge-info",
        completed: "badge-success",
        failed: "badge-error",
        cancelled: "bg-muted text-muted-foreground border border-border",
    };

    return (
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${styles[status] || "bg-muted"}`}>
            {status}
        </span>
    );
}

function HealthItem({ label, status, detail }: { label: string; status: "healthy" | "degraded" | "down"; detail?: string }) {
    const statusConfig = {
        healthy: { color: "bg-emerald-400", label: "Healthy" },
        degraded: { color: "bg-amber-400", label: "Degraded" },
        down: { color: "bg-red-400", label: "Down" },
    };

    return (
        <div className="flex items-center justify-between py-2">
            <div className="flex flex-col">
                <span className="text-sm font-medium">{label}</span>
                {detail && <span className="text-xs text-muted-foreground">{detail}</span>}
            </div>
            <div className="flex items-center gap-2">
                <span className={`h-2.5 w-2.5 rounded-full ${statusConfig[status].color}`} />
                <span className="text-xs text-muted-foreground">{statusConfig[status].label}</span>
            </div>
        </div>
    );
}

interface QuickActionCardProps {
    title: string;
    description: string;
    href: string;
    icon: React.ElementType;
    className: string;
}

function QuickActionCard({ title, description, href, icon: Icon, className }: QuickActionCardProps) {
    return (
        <Link href={href}>
            <Card className="group cursor-pointer transition-all duration-300 hover:scale-[1.02] hover:shadow-lg hover:shadow-violet-500/10 border-border/50 hover:border-violet-500/30">
                <CardContent className="flex items-center gap-4 p-6">
                    <div className={`p-3 rounded-xl ${className}`}>
                        <Icon className="h-6 w-6 text-white" />
                    </div>
                    <div>
                        <h3 className="font-semibold group-hover:text-violet-400 transition-colors">{title}</h3>
                        <p className="text-sm text-muted-foreground">{description}</p>
                    </div>
                </CardContent>
            </Card>
        </Link>
    );
}
