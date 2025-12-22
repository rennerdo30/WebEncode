"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@/i18n/routing";
import { createJob, fetchProfiles, Profile } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArrowLeft, ArrowRight, Film, Loader2, Upload, CheckCircle2, FolderOpen, Link as LinkIcon, CloudUpload } from "lucide-react";
import Link from "next/link";
import { FileBrowser } from "@/components/file-browser";
import { FileUpload } from "@/components/file-upload";

type Step = "source" | "profiles" | "review";

export default function NewJobPage() {
    const router = useRouter();
    const queryClient = useQueryClient();

    const [step, setStep] = useState<Step>("source");
    const [sourceUrl, setSourceUrl] = useState("");
    const [selectedProfiles, setSelectedProfiles] = useState<string[]>(["1080p"]);

    const { data: profiles, isLoading: profilesLoading } = useQuery({
        queryKey: ["profiles"],
        queryFn: fetchProfiles,
    });

    const createMutation = useMutation({
        mutationFn: () => createJob(sourceUrl, selectedProfiles),
        onSuccess: (job) => {
            queryClient.invalidateQueries({ queryKey: ["jobs"] });
            router.push(`/jobs/${job.id}`);
        },
    });

    const toggleProfile = (profileId: string) => {
        setSelectedProfiles(prev =>
            prev.includes(profileId)
                ? prev.filter(p => p !== profileId)
                : [...prev, profileId]
        );
    };

    const canProceed = () => {
        switch (step) {
            case "source": return sourceUrl.trim().length > 0;
            case "profiles": return selectedProfiles.length > 0;
            case "review": return true;
        }
    };

    const nextStep = () => {
        if (step === "source") setStep("profiles");
        else if (step === "profiles") setStep("review");
    };

    const prevStep = () => {
        if (step === "profiles") setStep("source");
        else if (step === "review") setStep("profiles");
    };

    return (
        <div className="max-w-3xl mx-auto space-y-6">
            <div className="flex items-center gap-4">
                <Link href="/jobs" className="p-2 hover:bg-muted rounded-md transition-colors">
                    <ArrowLeft className="h-5 w-5" />
                </Link>
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Create New Job</h1>
                    <p className="text-muted-foreground">Configure and submit an encoding job</p>
                </div>
            </div>

            {/* Progress Steps */}
            <div className="flex items-center justify-center gap-4">
                <StepIndicator label="Source" step={1} current={step === "source"} complete={step !== "source"} />
                <div className="h-px w-8 bg-border" />
                <StepIndicator label="Profiles" step={2} current={step === "profiles"} complete={step === "review"} />
                <div className="h-px w-8 bg-border" />
                <StepIndicator label="Review" step={3} current={step === "review"} complete={false} />
            </div>

            {/* Step Content */}
            {step === "source" && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Upload className="h-5 w-5" />
                            Source File
                        </CardTitle>
                        <CardDescription>
                            Choose a video file to encode. Upload from your device, browse storage, or enter a remote URL.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Tabs defaultValue="upload" className="w-full">
                            <TabsList className="grid w-full grid-cols-3 mb-4">
                                <TabsTrigger value="upload" className="flex items-center gap-2">
                                    <CloudUpload className="h-4 w-4" />
                                    Upload
                                </TabsTrigger>
                                <TabsTrigger value="browse" className="flex items-center gap-2">
                                    <FolderOpen className="h-4 w-4" />
                                    Browse
                                </TabsTrigger>
                                <TabsTrigger value="url" className="flex items-center gap-2">
                                    <LinkIcon className="h-4 w-4" />
                                    URL
                                </TabsTrigger>
                            </TabsList>

                            <TabsContent value="upload" className="mt-0">
                                <FileUpload
                                    onUploadComplete={(url) => setSourceUrl(url)}
                                />
                            </TabsContent>

                            <TabsContent value="browse" className="mt-0">
                                <FileBrowser
                                    onSelect={(path) => setSourceUrl(path)}
                                    selectedPath={sourceUrl}
                                    mediaOnly={true}
                                />
                            </TabsContent>

                            <TabsContent value="url" className="mt-0 space-y-4">
                                <div className="space-y-2">
                                    <Label htmlFor="source">Source URL</Label>
                                    <Input
                                        id="source"
                                        placeholder="https://example.com/video.mp4 or s3://bucket/video.mp4"
                                        value={sourceUrl}
                                        onChange={(e) => setSourceUrl(e.target.value)}
                                        className="font-mono"
                                    />
                                </div>
                                <div className="flex gap-2 text-xs text-muted-foreground">
                                    <Badge variant="outline">HTTP(S)</Badge>
                                    <Badge variant="outline">S3</Badge>
                                    <Badge variant="outline">SeaweedFS</Badge>
                                    <Badge variant="outline">file://</Badge>
                                </div>
                            </TabsContent>
                        </Tabs>

                        {sourceUrl && (
                            <div className="mt-4 p-3 rounded-lg bg-muted/50 border border-border">
                                <div className="flex items-center gap-2 text-sm">
                                    <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                                    <span className="text-muted-foreground">Selected:</span>
                                    <span className="font-mono text-xs truncate flex-1">{sourceUrl}</span>
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>
            )}

            {step === "profiles" && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Film className="h-5 w-5" />
                            Output Profiles
                        </CardTitle>
                        <CardDescription>
                            Select the encoding profiles to generate. Multiple profiles create an adaptive streaming set.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        {profilesLoading ? (
                            <div className="flex justify-center p-8">
                                <Loader2 className="h-6 w-6 animate-spin" />
                            </div>
                        ) : (
                            <div className="grid gap-3 sm:grid-cols-2">
                                {profiles?.map((profile) => (
                                    <ProfileCard
                                        key={profile.id}
                                        profile={profile}
                                        selected={selectedProfiles.includes(profile.id)}
                                        onToggle={() => toggleProfile(profile.id)}
                                    />
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            )}

            {step === "review" && (
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <CheckCircle2 className="h-5 w-5" />
                            Review & Submit
                        </CardTitle>
                        <CardDescription>
                            Confirm the job configuration before submitting.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="p-4 bg-muted rounded-lg space-y-3">
                            <div>
                                <span className="text-sm text-muted-foreground">Source URL</span>
                                <p className="font-mono text-sm truncate">{sourceUrl}</p>
                            </div>
                            <div>
                                <span className="text-sm text-muted-foreground">Output Profiles</span>
                                <div className="flex flex-wrap gap-2 mt-1">
                                    {selectedProfiles.map(p => (
                                        <Badge key={p} variant="secondary">{p}</Badge>
                                    ))}
                                </div>
                            </div>
                        </div>
                        {createMutation.isError && (
                            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-md text-red-500 text-sm">
                                Failed to create job. Please try again.
                            </div>
                        )}
                    </CardContent>
                </Card>
            )}

            {/* Navigation */}
            <div className="flex justify-between">
                <Button
                    variant="outline"
                    onClick={prevStep}
                    disabled={step === "source"}
                >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back
                </Button>

                {step === "review" ? (
                    <Button
                        onClick={() => createMutation.mutate()}
                        disabled={createMutation.isPending}
                    >
                        {createMutation.isPending && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                        Submit Job
                    </Button>
                ) : (
                    <Button onClick={nextStep} disabled={!canProceed()}>
                        Next
                        <ArrowRight className="h-4 w-4 ml-2" />
                    </Button>
                )}
            </div>
        </div>
    );
}

function StepIndicator({ label, step, current, complete }: {
    label: string;
    step: number;
    current: boolean;
    complete: boolean;
}) {
    return (
        <div className="flex items-center gap-2">
            <div className={`
                w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium
                ${current ? "bg-primary text-primary-foreground" : ""}
                ${complete ? "bg-green-500 text-white" : ""}
                ${!current && !complete ? "bg-muted text-muted-foreground" : ""}
            `}>
                {complete ? <CheckCircle2 className="h-4 w-4" /> : step}
            </div>
            <span className={`text-sm ${current ? "font-medium" : "text-muted-foreground"}`}>
                {label}
            </span>
        </div>
    );
}

function ProfileCard({ profile, selected, onToggle }: {
    profile: Profile;
    selected: boolean;
    onToggle: () => void;
}) {
    return (
        <div
            className={`
                p-4 rounded-lg border cursor-pointer transition-all
                ${selected ? "border-primary bg-primary/5" : "border-border hover:border-primary/50"}
            `}
            onClick={onToggle}
        >
            <div className="flex items-start gap-3">
                <Checkbox checked={selected} onChange={onToggle} />
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                        <span className="font-medium">{profile.name}</span>
                        {profile.is_system && (
                            <Badge variant="outline" className="text-xs">System</Badge>
                        )}
                    </div>
                    <div className="text-sm text-muted-foreground mt-1 space-x-2">
                        {profile.width && profile.height && (
                            <span>{profile.width}×{profile.height}</span>
                        )}
                        {profile.bitrate_kbps && (
                            <span>{profile.bitrate_kbps} kbps</span>
                        )}
                        <span className="font-mono">{profile.video_codec}</span>
                    </div>
                </div>
            </div>
        </div>
    );
}
