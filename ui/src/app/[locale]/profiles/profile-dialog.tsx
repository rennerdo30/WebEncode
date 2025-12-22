"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Profile, createProfile, updateProfile } from "@/lib/api";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Loader2, Plus, Edit2, Settings2 } from "lucide-react";

const profileSchema = z.object({
    id: z.string().min(2, "ID must be at least 2 characters").max(50).regex(/^[a-z0-9_-]+$/, "ID must be lowercase alphanumeric with dashes or underscores"),
    name: z.string().min(2, "Name must be at least 2 characters").max(100),
    description: z.string().optional(),
    video_codec: z.string().min(1, "Video codec is required"),
    audio_codec: z.string().optional(),
    width: z.coerce.number().int().min(0).optional(),
    height: z.coerce.number().int().min(0).optional(),
    bitrate_kbps: z.coerce.number().int().min(0).optional(),
    preset: z.string().optional(),
    container: z.string().optional(),
    config_json: z.string().optional().refine((val) => {
        if (!val) return true;
        try {
            JSON.parse(val);
            return true;
        } catch {
            return false;
        }
    }, "Invalid JSON format"),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

interface ProfileDialogProps {
    profile?: Profile;
    trigger?: React.ReactNode;
    open?: boolean;
    onOpenChange?: (open: boolean) => void;
}

export function ProfileDialog({ profile, trigger, open, onOpenChange }: ProfileDialogProps) {
    const [isOpen, setIsOpen] = useState(false);
    const queryClient = useQueryClient();
    const isEditing = !!profile;

    const form = useForm<ProfileFormValues>({
        resolver: zodResolver(profileSchema) as any,
        defaultValues: {
            id: profile?.id || "",
            name: profile?.name || "",
            description: profile?.description || "",
            video_codec: profile?.video_codec || "libx264",
            audio_codec: profile?.audio_codec || "aac",
            width: profile?.width || 1920,
            height: profile?.height || 1080,
            bitrate_kbps: profile?.bitrate_kbps || 5000,
            preset: profile?.preset || "fast",
            container: profile?.container || "mp4",
            config_json: profile?.config ? JSON.stringify(profile.config, null, 2) : "{}",
        },
    });

    // Reset form when profile changes
    useEffect(() => {
        if (profile) {
            form.reset({
                id: profile.id,
                name: profile.name,
                description: profile.description || "",
                video_codec: profile.video_codec,
                audio_codec: profile.audio_codec || "",
                width: profile.width,
                height: profile.height,
                bitrate_kbps: profile.bitrate_kbps,
                preset: profile.preset || "",
                container: profile.container || "",
                config_json: profile.config ? JSON.stringify(profile.config, null, 2) : "{}",
            });
        } else {
            form.reset({
                id: "",
                name: "",
                description: "",
                video_codec: "libx264",
                audio_codec: "aac",
                width: 1920,
                height: 1080,
                bitrate_kbps: 5000,
                preset: "fast",
                container: "mp4",
                config_json: "{}",
            });
        }
    }, [profile, form, isOpen]);

    const mutation = useMutation({
        mutationFn: async (values: ProfileFormValues) => {
            const config = values.config_json ? JSON.parse(values.config_json) : {};
            const payload: Partial<Profile> = {
                ...values,
                config,
            };

            // Remove config_json from payload as it's not in Profile interface directly in the way we want (it's mapped to config)
            delete (payload as any).config_json;

            if (isEditing && profile) {
                await updateProfile(profile.id, payload);
            } else {
                await createProfile(payload);
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["profiles"] });
            toast.success(isEditing ? "Profile updated successfully" : "Profile created successfully");
            setIsOpen(false);
            onOpenChange?.(false);
            if (!isEditing) form.reset();
        },
        onError: (error) => {
            toast.error(`Failed to ${isEditing ? "update" : "create"} profile: ${error.message}`);
        },
    });

    const onSubmit = (values: ProfileFormValues) => {
        mutation.mutate(values);
    };

    const handleOpenChange = (newOpen: boolean) => {
        setIsOpen(newOpen);
        onOpenChange?.(newOpen);
    };

    return (
        <Dialog open={open !== undefined ? open : isOpen} onOpenChange={handleOpenChange}>
            <DialogTrigger asChild>
                {trigger || (
                    <Button>
                        <Plus className="mr-2 h-4 w-4" />
                        Create Profile
                    </Button>
                )}
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{isEditing ? "Edit Profile" : "Create New Profile"}</DialogTitle>
                    <DialogDescription>
                        Configure encoding settings for your transcoding jobs.
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {/* Basic Info */}
                            <div className="space-y-4 md:col-span-2">
                                <h3 className="text-sm font-medium text-muted-foreground border-b pb-2 flex items-center gap-2">
                                    <Settings2 className="h-4 w-4" /> General
                                </h3>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="id"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>ID</FormLabel>
                                                <FormControl>
                                                    <Input placeholder="my_custom_profile" {...field} disabled={isEditing} />
                                                </FormControl>
                                                <FormDescription>Unique identifier for the profile.</FormDescription>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="name"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Name</FormLabel>
                                                <FormControl>
                                                    <Input placeholder="My Custom Profile" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                                <FormField
                                    control={form.control}
                                    name="description"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Description</FormLabel>
                                            <FormControl>
                                                <Input placeholder="Optimization for high motion content..." {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            {/* Video Settings */}
                            <div className="space-y-4 md:col-span-2">
                                <h3 className="text-sm font-medium text-muted-foreground border-b pb-2 flex items-center gap-2">
                                    <Settings2 className="h-4 w-4" /> Video Settings
                                </h3>
                                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="video_codec"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Codec</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder="Select codec" />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="libx264">H.264 (libx264)</SelectItem>
                                                        <SelectItem value="libx265">H.265 (libx265)</SelectItem>
                                                        <SelectItem value="libvpx-vp9">VP9 (libvpx-vp9)</SelectItem>
                                                        <SelectItem value="av1">AV1</SelectItem>
                                                        <SelectItem value="copy">Copy (No Transcode)</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="preset"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Preset</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder="Select preset" />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="ultrafast">Ultrafast</SelectItem>
                                                        <SelectItem value="superfast">Superfast</SelectItem>
                                                        <SelectItem value="veryfast">Veryfast</SelectItem>
                                                        <SelectItem value="faster">Faster</SelectItem>
                                                        <SelectItem value="fast">Fast</SelectItem>
                                                        <SelectItem value="medium">Medium</SelectItem>
                                                        <SelectItem value="slow">Slow</SelectItem>
                                                        <SelectItem value="slower">Slower</SelectItem>
                                                        <SelectItem value="veryslow">Veryslow</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="bitrate_kbps"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Bitrate (Kbps)</FormLabel>
                                                <FormControl>
                                                    <Input type="number" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="width"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Width</FormLabel>
                                                <FormControl>
                                                    <Input type="number" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="height"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Height</FormLabel>
                                                <FormControl>
                                                    <Input type="number" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                            </div>

                            {/* Audio & Container */}
                            <div className="space-y-4 md:col-span-2">
                                <h3 className="text-sm font-medium text-muted-foreground border-b pb-2 flex items-center gap-2">
                                    <Settings2 className="h-4 w-4" /> Audio & Output
                                </h3>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="audio_codec"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Audio Codec</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder="Select audio codec" />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="aac">AAC</SelectItem>
                                                        <SelectItem value="libopus">Opus</SelectItem>
                                                        <SelectItem value="libmp3lame">MP3</SelectItem>
                                                        <SelectItem value="copy">Copy</SelectItem>
                                                        <SelectItem value="none">None</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="container"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Container</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl>
                                                        <SelectTrigger>
                                                            <SelectValue placeholder="Select container" />
                                                        </SelectTrigger>
                                                    </FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="mp4">MP4</SelectItem>
                                                        <SelectItem value="mkv">MKV</SelectItem>
                                                        <SelectItem value="webm">WebM</SelectItem>
                                                        <SelectItem value="mov">MOV</SelectItem>
                                                        <SelectItem value="ts">MPEG-TS</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                            </div>

                            {/* Advanced Config */}
                            <div className="md:col-span-2">
                                <FormField
                                    control={form.control}
                                    name="config_json"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Advanced Configuration (JSON)</FormLabel>
                                            <FormControl>
                                                <Textarea
                                                    placeholder="{}"
                                                    className="font-mono text-xs h-32"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormDescription>
                                                Pass arbitrary FFmpeg flags or filter chains here.
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>
                        </div>

                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
                                Cancel
                            </Button>
                            <Button type="submit" disabled={mutation.isPending}>
                                {mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                {isEditing ? "Save Changes" : "Create Profile"}
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
