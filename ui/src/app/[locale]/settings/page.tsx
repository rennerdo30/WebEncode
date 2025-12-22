"use client";

import React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchPlugins, enablePlugin, disablePlugin, updatePluginConfig, Plugin, fetchProfiles, createProfile, updateProfile, deleteProfile, Profile } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Loader2, Plug, Settings, CheckCircle2, XCircle, AlertTriangle, Video, Pencil } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";


export default function SettingsPage() {
    return (
        <div className="space-y-6">
            <h1 className="text-3xl font-bold tracking-tight">Settings</h1>

            <Tabs defaultValue="plugins" className="w-full">
                <TabsList>
                    <TabsTrigger value="plugins">Plugins</TabsTrigger>
                    <TabsTrigger value="profiles">Profiles</TabsTrigger>
                    <TabsTrigger value="storage">Storage</TabsTrigger>
                    <TabsTrigger value="auth">Authentication</TabsTrigger>
                </TabsList>

                <TabsContent value="plugins" className="mt-6">
                    <PluginsSection />
                </TabsContent>

                <TabsContent value="profiles" className="mt-6">
                    <ProfilesSection />
                </TabsContent>

                <TabsContent value="storage" className="mt-6">
                    <StorageSection />
                </TabsContent>

                <TabsContent value="auth" className="mt-6">
                    <AuthSection />
                </TabsContent>
            </Tabs>
        </div>
    );
}

function PluginsSection() {
    const { data: plugins, isLoading, isError } = useQuery({
        queryKey: ["plugins"],
        queryFn: fetchPlugins,
    });

    if (isLoading) {
        return (
            <div className="flex justify-center p-8">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    if (isError) {
        return (
            <Card className="border-red-500/50">
                <CardContent className="pt-6">
                    <div className="flex items-center gap-3 text-red-500">
                        <XCircle className="h-5 w-5" />
                        <span>Failed to load plugins. Make sure the API is running.</span>
                    </div>
                </CardContent>
            </Card>
        );
    }

    const pluginTypes = ["encoder", "storage", "auth", "live", "publisher"];
    const hasPlugins = plugins && plugins.length > 0;

    if (!hasPlugins) {
        return (
            <Card>
                <CardContent className="pt-6">
                    <div className="text-center py-8">
                        <Plug className="h-12 w-12 text-muted-foreground/50 mx-auto mb-4" />
                        <h3 className="text-lg font-medium mb-2">No Plugins Registered</h3>
                        <p className="text-muted-foreground mb-4 max-w-md mx-auto">
                            Plugins need to be registered in the database. Run the database migrations to seed default plugin configurations.
                        </p>
                        <div className="font-mono text-sm bg-muted p-3 rounded-md inline-block">
                            make migrate
                        </div>
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <div className="space-y-6">
            {pluginTypes.map((type) => {
                const typePlugins = plugins?.filter((p) => p.type === type) || [];
                if (typePlugins.length === 0) return null;

                return (
                    <div key={type}>
                        <h2 className="text-lg font-semibold mb-3 capitalize">{type} Plugins</h2>
                        <div className="grid gap-4 md:grid-cols-2">
                            {typePlugins.map((plugin) => (
                                <PluginCard key={plugin.id} plugin={plugin} />
                            ))}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}

function PluginCard({ plugin }: { plugin: Plugin }) {
    const queryClient = useQueryClient();
    const [configDialogOpen, setConfigDialogOpen] = React.useState(false);
    const [configJson, setConfigJson] = React.useState("");

    const toggleMutation = useMutation({
        mutationFn: () => plugin.is_enabled ? disablePlugin(plugin.id) : enablePlugin(plugin.id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["plugins"] });
            toast.success(plugin.is_enabled ? "Plugin disabled" : "Plugin enabled");
        },
        onError: (error) => {
            toast.error("Failed to toggle plugin", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    const updateConfigMutation = useMutation({
        mutationFn: () => {
            try {
                const config = JSON.parse(configJson);
                return updatePluginConfig(plugin.id, { config });
            } catch {
                throw new Error("Invalid JSON configuration");
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["plugins"] });
            setConfigDialogOpen(false);
            toast.success("Plugin configuration updated");
        },
        onError: (error) => {
            toast.error("Failed to update configuration", {
                description: error instanceof Error ? error.message : "Unknown error"
            });
        }
    });

    const openConfigDialog = () => {
        setConfigJson(JSON.stringify(plugin.config || {}, null, 2));
        setConfigDialogOpen(true);
    };

    const healthIcons: Record<string, React.ReactNode> = {
        healthy: <CheckCircle2 className="h-4 w-4 text-green-500" />,
        degraded: <AlertTriangle className="h-4 w-4 text-yellow-500" />,
        failed: <XCircle className="h-4 w-4 text-red-500" />,
        disabled: <XCircle className="h-4 w-4 text-gray-400" />,
    };

    // Define known config fields for common plugin types
    const getPluginConfigFields = (pluginId: string, pluginType: string) => {
        // Publisher plugins
        if (pluginType === "publisher") {
            return [
                { key: "client_id", label: "Client ID", type: "text", description: "OAuth Client ID" },
                { key: "client_secret", label: "Client Secret", type: "password", description: "OAuth Client Secret" },
                { key: "redirect_uri", label: "Redirect URI", type: "text", description: "OAuth Redirect URI" },
            ];
        }
        // Storage plugins
        if (pluginType === "storage") {
            return [
                { key: "endpoint", label: "Endpoint", type: "text", description: "Storage endpoint URL" },
                { key: "bucket", label: "Bucket", type: "text", description: "Default bucket name" },
                { key: "access_key", label: "Access Key", type: "text", description: "Access key ID" },
                { key: "secret_key", label: "Secret Key", type: "password", description: "Secret access key" },
                { key: "region", label: "Region", type: "text", description: "Storage region" },
            ];
        }
        // Default: no known fields, just show JSON editor
        return [];
    };

    const configFields = getPluginConfigFields(plugin.id, plugin.type);

    return (
        <>
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <Plug className="h-5 w-5 text-muted-foreground" />
                            <CardTitle className="text-base">{plugin.id}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={openConfigDialog}
                                title="Configure plugin"
                            >
                                <Pencil className="h-4 w-4" />
                            </Button>
                            <Switch
                                checked={plugin.is_enabled}
                                onCheckedChange={() => toggleMutation.mutate()}
                                disabled={toggleMutation.isPending}
                            />
                        </div>
                    </div>
                    <CardDescription>Type: {plugin.type}</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            {healthIcons[plugin.health]}
                            <span className="text-sm capitalize">{plugin.health}</span>
                        </div>
                        <div className="flex items-center gap-2">
                            {plugin.config && Object.keys(plugin.config).length > 0 && (
                                <Badge variant="secondary" className="text-xs">Configured</Badge>
                            )}
                            {plugin.version && (
                                <Badge variant="outline" className="text-xs">v{plugin.version}</Badge>
                            )}
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Plugin Configuration Dialog */}
            <Dialog open={configDialogOpen} onOpenChange={setConfigDialogOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>Configure {plugin.id}</DialogTitle>
                        <DialogDescription>
                            Update the configuration for this {plugin.type} plugin.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        {configFields.length > 0 ? (
                            // Show structured form for known plugin types
                            <>
                                {configFields.map((field) => {
                                    const currentConfig = plugin.config || {};
                                    const value = (currentConfig as Record<string, unknown>)[field.key] as string || "";
                                    return (
                                        <div key={field.key} className="space-y-2">
                                            <Label htmlFor={field.key}>{field.label}</Label>
                                            <Input
                                                id={field.key}
                                                type={field.type}
                                                defaultValue={value}
                                                placeholder={field.description}
                                                onChange={(e) => {
                                                    try {
                                                        const config = JSON.parse(configJson);
                                                        config[field.key] = e.target.value;
                                                        setConfigJson(JSON.stringify(config, null, 2));
                                                    } catch {
                                                        // If JSON is invalid, just update the field
                                                    }
                                                }}
                                            />
                                            <p className="text-xs text-muted-foreground">{field.description}</p>
                                        </div>
                                    );
                                })}
                                <div className="pt-4 border-t">
                                    <details className="text-sm">
                                        <summary className="cursor-pointer text-muted-foreground hover:text-foreground">
                                            Advanced: Edit raw JSON
                                        </summary>
                                        <Textarea
                                            className="mt-2 font-mono text-xs"
                                            value={configJson}
                                            onChange={(e) => setConfigJson(e.target.value)}
                                            rows={8}
                                        />
                                    </details>
                                </div>
                            </>
                        ) : (
                            // Show JSON editor for unknown plugin types
                            <div className="space-y-2">
                                <Label>Configuration (JSON)</Label>
                                <Textarea
                                    className="font-mono text-sm"
                                    value={configJson}
                                    onChange={(e) => setConfigJson(e.target.value)}
                                    rows={12}
                                    placeholder='{"key": "value"}'
                                />
                                <p className="text-xs text-muted-foreground">
                                    Enter the configuration as valid JSON. Check the plugin documentation for available options.
                                </p>
                            </div>
                        )}
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setConfigDialogOpen(false)}>
                            Cancel
                        </Button>
                        <Button
                            onClick={() => updateConfigMutation.mutate()}
                            disabled={updateConfigMutation.isPending}
                        >
                            {updateConfigMutation.isPending ? (
                                <>
                                    <Loader2 className="h-4 w-4 animate-spin mr-1.5" />
                                    Saving...
                                </>
                            ) : (
                                "Save Configuration"
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}

function ProfilesSection() {
    const queryClient = useQueryClient();
    const [isDialogOpen, setIsDialogOpen] = React.useState(false);
    const [editingProfile, setEditingProfile] = React.useState<Partial<Profile> | null>(null);

    const { data: profiles, isLoading, isError } = useQuery({
        queryKey: ["profiles"],
        queryFn: fetchProfiles,
    });

    const createMutation = useMutation({
        mutationFn: createProfile,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["profiles"] });
            setIsDialogOpen(false);
        },
    });

    const updateMutation = useMutation({
        mutationFn: (data: Partial<Profile>) => updateProfile(data.id!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["profiles"] });
            setIsDialogOpen(false);
        },
    });

    const deleteMutation = useMutation({
        mutationFn: deleteProfile,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["profiles"] }),
    });

    if (isLoading) {
        return (
            <div className="flex justify-center p-8">
                <Loader2 className="h-8 w-8 animate-spin" />
            </div>
        );
    }

    if (isError) {
        return (
            <div className="p-4 text-red-500 bg-red-500/10 rounded-lg">
                Failed to load encoding profiles.
            </div>
        );
    }

    const handleEdit = (profile: Profile) => {
        setEditingProfile(profile);
        setIsDialogOpen(true);
    };

    const handleCreate = () => {
        setEditingProfile({
            id: "",
            name: "",
            video_codec: "libx264",
            audio_codec: "aac",
            preset: "fast",
            container: "mp4",
            width: 1920,
            height: 1080,
            bitrate_kbps: 5000,
        });
        setIsDialogOpen(true);
    };

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const data = Object.fromEntries(formData.entries());

        const profileData: Partial<Profile> = {
            ...editingProfile,
            id: data.id as string,
            name: data.name as string,
            description: data.description as string,
            video_codec: data.video_codec as string,
            audio_codec: data.audio_codec as string,
            preset: data.preset as string,
            container: data.container as string,
            width: parseInt(data.width as string),
            height: parseInt(data.height as string),
            bitrate_kbps: parseInt(data.bitrate_kbps as string),
        };

        if (editingProfile?.is_system) {
            // Should not happen as UI disables edit for system profiles or we handle it
            return;
        }

        const isNew = !profiles?.find(p => p.id === profileData.id);

        if (isNew) {
            createMutation.mutate(profileData);
        } else {
            updateMutation.mutate(profileData);
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-xl font-semibold">Encoding Profiles</h2>
                    <p className="text-sm text-muted-foreground">Manage video encoding presets and ABR ladders</p>
                </div>
                <Button onClick={handleCreate} className="gap-2">
                    <Settings className="h-4 w-4" />
                    New Profile
                </Button>
            </div>

            <div className="grid gap-4">
                {profiles?.map((profile) => (
                    <Card key={profile.id} className="overflow-hidden border-border/50 bg-card/50 backdrop-blur-sm hover:bg-card/80 transition-colors">
                        <div className="p-6 flex items-center justify-between">
                            <div className="flex items-center gap-4">
                                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
                                    <Video className="h-6 w-6 text-primary" />
                                </div>
                                <div>
                                    <div className="flex items-center gap-2">
                                        <h3 className="font-semibold text-lg">{profile.name}</h3>
                                        {profile.is_system && (
                                            <Badge variant="secondary" className="text-[10px] uppercase font-bold tracking-wider">System</Badge>
                                        )}
                                    </div>
                                    <p className="text-sm text-muted-foreground line-clamp-1">
                                        {profile.video_codec} • {profile.width}x{profile.height} • {profile.bitrate_kbps}kbps
                                    </p>
                                </div>
                            </div>
                            <div className="flex items-center gap-2">
                                {!profile.is_system && (
                                    <>
                                        <Button variant="ghost" size="sm" onClick={() => handleEdit(profile)}>Edit</Button>
                                        <Button variant="ghost" size="sm" className="text-red-500 hover:text-red-600 hover:bg-red-500/10" onClick={() => deleteMutation.mutate(profile.id)}>Delete</Button>
                                    </>
                                )}
                                {profile.is_system && (
                                    <Button variant="ghost" size="sm" disabled>View Only</Button>
                                )}
                            </div>
                        </div>
                    </Card>
                ))}
            </div>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[600px] bg-background/95 backdrop-blur-md border-border/50">
                    <DialogHeader>
                        <DialogTitle>{editingProfile?.id ? "Edit Profile" : "Create New Profile"}</DialogTitle>
                        <DialogDescription>
                            Configure the video and audio settings for this encoding preset.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleSubmit} className="space-y-4 py-4">
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="id">Profile ID</Label>
                                <Input id="id" name="id" defaultValue={editingProfile?.id} disabled={!!editingProfile?.id} placeholder="e.g. 1080p_h264_high" required />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="name">Display Name</Label>
                                <Input id="name" name="name" defaultValue={editingProfile?.name} placeholder="e.g. 1080p High Quality" required />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="description">Description (Optional)</Label>
                            <Input id="description" name="description" defaultValue={editingProfile?.description} placeholder="Short description of this profile" />
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="video_codec">Video Codec</Label>
                                <Select name="video_codec" defaultValue={editingProfile?.video_codec}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select codec" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="libx264">H.264 (libx264)</SelectItem>
                                        <SelectItem value="hevc_nvenc">H.265 (HEVC NVENC)</SelectItem>
                                        <SelectItem value="libvpx-vp9">VP9 (WebM)</SelectItem>
                                        <SelectItem value="libaom-av1">AV1 (libaom)</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="audio_codec">Audio Codec</Label>
                                <Select name="audio_codec" defaultValue={editingProfile?.audio_codec}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select codec" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="aac">AAC</SelectItem>
                                        <SelectItem value="libopus">Opus</SelectItem>
                                        <SelectItem value="mp3">MP3</SelectItem>
                                        <SelectItem value="copy">Copy (Passthrough)</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        <div className="grid grid-cols-3 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="width">Width</Label>
                                <Input id="width" name="width" type="number" defaultValue={editingProfile?.width} required />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="height">Height</Label>
                                <Input id="height" name="height" type="number" defaultValue={editingProfile?.height} required />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="bitrate_kbps">Bitrate (kbps)</Label>
                                <Input id="bitrate_kbps" name="bitrate_kbps" type="number" defaultValue={editingProfile?.bitrate_kbps} required />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="preset">Encoder Preset</Label>
                                <Select name="preset" defaultValue={editingProfile?.preset}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select preset" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="ultrafast">Ultrafast</SelectItem>
                                        <SelectItem value="superfast">Superfast</SelectItem>
                                        <SelectItem value="veryfast">Veryfast</SelectItem>
                                        <SelectItem value="fast">Fast</SelectItem>
                                        <SelectItem value="medium">Medium</SelectItem>
                                        <SelectItem value="slow">Slow</SelectItem>
                                        <SelectItem value="veryslow">Verslow</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="container">Container</Label>
                                <Select name="container" defaultValue={editingProfile?.container}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select container" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="mp4">MP4</SelectItem>
                                        <SelectItem value="mkv">MKV</SelectItem>
                                        <SelectItem value="webm">WebM</SelectItem>
                                        <SelectItem value="mov">MOV</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        <DialogFooter className="pt-4">
                            <Button type="button" variant="ghost" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                                {(createMutation.isPending || updateMutation.isPending) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Save Profile
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}


function StorageSection() {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Storage Configuration</CardTitle>
                <CardDescription>Configure object storage backends</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                        <span className="font-medium">SeaweedFS</span>
                        <Badge variant="default">Active</Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                        Endpoint: seaweedfs-filer:8333
                    </p>
                </div>
                <div className="p-4 bg-muted rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                        <span className="font-medium">Local Filesystem</span>
                        <Badge variant="secondary">Available</Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                        Path: /data/webencode
                    </p>
                </div>
            </CardContent>
        </Card>
    );
}

function AuthSection() {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Authentication</CardTitle>
                <CardDescription>Configure identity providers</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="p-4 bg-muted rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                        <span className="font-medium">No Authentication</span>
                        <Badge variant="default">Active</Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                        All requests are allowed. Suitable for development only.
                    </p>
                </div>
                <p className="text-sm text-muted-foreground">
                    To enable OAuth2/OIDC, configure the auth-oidc plugin with your identity provider settings.
                </p>
            </CardContent>
        </Card>
    );
}
