"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchProfiles, deleteProfile, Profile } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Loader2, Film, Lock, Trash2, Pencil } from "lucide-react";
import { ProfileDialog } from "./profile-dialog";
import { toast } from "sonner";

export default function ProfilesPage() {
    const { data: profiles, isLoading } = useQuery({
        queryKey: ["profiles"],
        queryFn: fetchProfiles,
    });

    if (isLoading) {
        return (
            <div className="flex justify-center p-16">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    // Sort profiles to ensure consistent order
    const sortedProfiles = profiles ? [...profiles].sort((a, b) => a.name.localeCompare(b.name)) : [];

    const systemProfiles = sortedProfiles.filter(p => p.is_system);
    const customProfiles = sortedProfiles.filter(p => !p.is_system);

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold tracking-tight">Encoding Profiles</h1>
                <ProfileDialog />
            </div>

            <div className="space-y-6">
                <section>
                    <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                        <Lock className="h-5 w-5" />
                        System Profiles
                    </h2>
                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {systemProfiles.map((profile) => (
                            <ProfileCard key={profile.id} profile={profile} />
                        ))}
                    </div>
                </section>

                <section>
                    <div className="flex items-center gap-2 mb-4">
                        <h2 className="text-xl font-semibold">Custom Profiles</h2>
                        {customProfiles.length === 0 && (
                            <span className="text-sm text-muted-foreground">(None created yet)</span>
                        )}
                    </div>

                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {customProfiles.map((profile) => (
                            <ProfileCard key={profile.id} profile={profile} />
                        ))}
                    </div>
                </section>
            </div>
        </div>
    );
}

function ProfileCard({ profile }: { profile: Profile }) {
    const queryClient = useQueryClient();

    const deleteMutation = useMutation({
        mutationFn: deleteProfile,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["profiles"] });
            toast.success("Profile deleted");
        },
        onError: (error) => {
            toast.error("Failed to delete profile: " + error.message);
        }
    });

    const handleDelete = () => {
        if (confirm("Are you sure you want to delete this profile?")) {
            deleteMutation.mutate(profile.id);
        }
    };

    return (
        <Card className="flex flex-col h-full">
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <Film className="h-5 w-5 text-muted-foreground" />
                        <CardTitle className="text-base truncate" title={profile.name}>{profile.name}</CardTitle>
                    </div>
                    {profile.is_system && (
                        <Badge variant="secondary" className="text-xs">System</Badge>
                    )}
                </div>
                {profile.description && (
                    <p className="text-xs text-muted-foreground line-clamp-2" title={profile.description}>
                        {profile.description}
                    </p>
                )}
            </CardHeader>
            <CardContent className="space-y-2 text-sm flex-1">
                <div className="grid grid-cols-2 gap-2">
                    <div>
                        <span className="text-muted-foreground">Video:</span>
                        <span className="ml-1 font-mono">{profile.video_codec}</span>
                    </div>
                    {profile.audio_codec && (
                        <div>
                            <span className="text-muted-foreground">Audio:</span>
                            <span className="ml-1 font-mono">{profile.audio_codec}</span>
                        </div>
                    )}
                </div>
                {(profile.width || profile.height) && (
                    <div>
                        <span className="text-muted-foreground">Resolution:</span>
                        <span className="ml-1">{profile.width}x{profile.height}</span>
                    </div>
                )}
                {profile.bitrate_kbps && (
                    <div>
                        <span className="text-muted-foreground">Bitrate:</span>
                        <span className="ml-1">{profile.bitrate_kbps} kbps</span>
                    </div>
                )}
                {profile.container && (
                    <div>
                        <span className="text-muted-foreground">Container:</span>
                        <span className="ml-1 uppercase">{profile.container}</span>
                    </div>
                )}
            </CardContent>
            {!profile.is_system && (
                <CardFooter className="pt-0 flex gap-2 justify-end border-t mt-4 p-4">
                    <ProfileDialog
                        profile={profile}
                        trigger={
                            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                                <Pencil className="h-4 w-4" />
                                <span className="sr-only">Edit</span>
                            </Button>
                        }
                    />
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                        onClick={handleDelete}
                        disabled={deleteMutation.isPending}
                    >
                        {deleteMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                            <Trash2 className="h-4 w-4" />
                        )}
                        <span className="sr-only">Delete</span>
                    </Button>
                </CardFooter>
            )}
        </Card>
    );
}
