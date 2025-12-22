"use client";

import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
    browseFiles,
    fetchBrowseRoots,
    BrowseEntry,
    BrowseRoot,
    BrowseResponse,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
    Folder,
    File,
    Film,
    Music,
    Image,
    ChevronLeft,
    ChevronRight,
    Search,
    Home,
    RefreshCw,
    Loader2,
    HardDrive,
    Check,
} from "lucide-react";

interface FileBrowserProps {
    onSelect: (path: string) => void;
    selectedPath?: string;
    mediaOnly?: boolean;
}

export function FileBrowser({ onSelect, selectedPath, mediaOnly = true }: FileBrowserProps) {
    const [currentPath, setCurrentPath] = useState<string>("");
    const [currentPlugin, setCurrentPlugin] = useState<string>("");
    const [searchQuery, setSearchQuery] = useState("");
    const [showHidden, setShowHidden] = useState(false);

    // Fetch browse roots
    const { data: roots, isLoading: rootsLoading } = useQuery({
        queryKey: ["browse-roots"],
        queryFn: fetchBrowseRoots,
    });

    // Fetch directory contents
    const { data: browseData, isLoading: browseLoading, refetch } = useQuery({
        queryKey: ["browse-files", currentPlugin, currentPath, searchQuery, showHidden, mediaOnly],
        queryFn: () =>
            browseFiles({
                plugin: currentPlugin,
                path: currentPath,
                search: searchQuery,
                showHidden,
                mediaOnly,
            }),
        enabled: !!currentPath || !!currentPlugin,
    });

    // Auto-select first root if none selected
    useEffect(() => {
        if (roots && roots.length > 0 && !currentPath && !currentPlugin) {
            setCurrentPlugin(roots[0].plugin_id);
            setCurrentPath(roots[0].path);
        }
    }, [roots, currentPath, currentPlugin]);

    const handleNavigate = (entry: BrowseEntry) => {
        if (entry.is_directory) {
            setCurrentPath(entry.path);
        } else {
            // Select the file
            onSelect(`file://${entry.path}`);
        }
    };

    const handleGoUp = () => {
        if (browseData?.parent_path) {
            setCurrentPath(browseData.parent_path);
        }
    };

    const handleSelectRoot = (root: BrowseRoot) => {
        setCurrentPlugin(root.plugin_id);
        setCurrentPath(root.path);
        setSearchQuery("");
    };

    const formatSize = (bytes: number): string => {
        if (bytes === 0) return "—";
        const units = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
    };

    const formatDate = (timestamp: number): string => {
        if (!timestamp) return "—";
        return new Date(timestamp * 1000).toLocaleDateString();
    };

    const getFileIcon = (entry: BrowseEntry) => {
        if (entry.is_directory) {
            return <Folder className="h-5 w-5 text-amber-400" />;
        }
        if (entry.is_video) {
            return <Film className="h-5 w-5 text-violet-400" />;
        }
        if (entry.is_audio) {
            return <Music className="h-5 w-5 text-cyan-400" />;
        }
        if (entry.is_image) {
            return <Image className="h-5 w-5 text-emerald-400" />;
        }
        return <File className="h-5 w-5 text-muted-foreground" />;
    };

    const isSelected = (entry: BrowseEntry) => {
        return selectedPath === `file://${entry.path}`;
    };

    if (rootsLoading) {
        return (
            <div className="flex items-center justify-center p-8">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (!roots || roots.length === 0) {
        return (
            <div className="text-center p-8 text-muted-foreground">
                <HardDrive className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>No storage plugins with browse support available.</p>
                <p className="text-sm mt-2">Make sure storage-fs plugin is loaded.</p>
            </div>
        );
    }

    return (
        <div className="border border-border rounded-lg overflow-hidden bg-card">
            {/* Header */}
            <div className="border-b border-border p-3 bg-muted/30">
                <div className="flex items-center gap-2 mb-3">
                    {/* Navigation buttons */}
                    <Button
                        variant="outline"
                        size="icon"
                        onClick={handleGoUp}
                        disabled={!browseData?.parent_path}
                        className="h-8 w-8"
                    >
                        <ChevronLeft className="h-4 w-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="icon"
                        onClick={() => refetch()}
                        disabled={browseLoading}
                        className="h-8 w-8"
                    >
                        <RefreshCw className={`h-4 w-4 ${browseLoading ? "animate-spin" : ""}`} />
                    </Button>

                    {/* Path breadcrumb */}
                    <div className="flex-1 px-3 py-1.5 bg-background rounded-md text-sm font-mono text-muted-foreground truncate">
                        {browseData?.current_path || currentPath || "/"}
                    </div>
                </div>

                {/* Search */}
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search files..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-9 h-8"
                    />
                </div>
            </div>

            {/* Sidebar + Content */}
            <div className="flex" style={{ height: "350px" }}>
                {/* Sidebar - Roots grouped by Plugin */}
                <div className="w-48 border-r border-border bg-muted/20 overflow-y-auto">
                    {/* Group roots by plugin */}
                    {(() => {
                        // Get unique plugin IDs
                        const plugins = [...new Set(roots.map(r => r.plugin_id))];
                        return plugins.map((pluginId) => {
                            const pluginRoots = roots.filter(r => r.plugin_id === pluginId);
                            // Determine plugin type from the first root
                            const storageType = pluginRoots[0]?.storage_type || "filesystem";
                            const isActivePlugin = currentPlugin === pluginId;

                            return (
                                <div key={pluginId} className="mb-2">
                                    {/* Plugin header */}
                                    <div className={`px-2 py-1.5 text-xs font-medium uppercase tracking-wide flex items-center gap-1.5 ${isActivePlugin ? "text-violet-400" : "text-muted-foreground"
                                        }`}>
                                        <HardDrive className="h-3 w-3" />
                                        <span className="truncate">{pluginId.replace('storage-', '')}</span>
                                        {isActivePlugin && (
                                            <span className="ml-auto h-1.5 w-1.5 rounded-full bg-violet-400" />
                                        )}
                                    </div>
                                    {/* Plugin roots */}
                                    {pluginRoots.map((root, index) => (
                                        <button
                                            key={`${root.plugin_id}-${root.path}-${index}`}
                                            onClick={() => handleSelectRoot(root)}
                                            className={`w-full text-left px-3 py-2 text-sm flex items-center gap-2 hover:bg-muted/50 transition-colors ${currentPath === root.path && currentPlugin === root.plugin_id
                                                    ? "bg-violet-500/10 text-violet-400 border-l-2 border-violet-400"
                                                    : "text-muted-foreground"
                                                }`}
                                        >
                                            <Folder className="h-4 w-4" />
                                            <span className="truncate">{root.name}</span>
                                        </button>
                                    ))}
                                </div>
                            );
                        });
                    })()}
                </div>

                {/* File list */}
                <div className="flex-1 overflow-y-auto">
                    {browseLoading ? (
                        <div className="flex items-center justify-center h-full">
                            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                        </div>
                    ) : browseData?.entries && browseData.entries.length > 0 ? (
                        <div className="divide-y divide-border/50">
                            {browseData.entries.map((entry, index) => (
                                <button
                                    key={`${entry.path}-${index}`}
                                    onClick={() => handleNavigate(entry)}
                                    onDoubleClick={() => {
                                        if (!entry.is_directory) {
                                            onSelect(`file://${entry.path}`);
                                        }
                                    }}
                                    className={`w-full text-left px-4 py-2.5 flex items-center gap-3 hover:bg-muted/50 transition-colors group ${isSelected(entry) ? "bg-violet-500/10 border-l-2 border-violet-500" : ""
                                        }`}
                                >
                                    {getFileIcon(entry)}
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2">
                                            <span className={`truncate ${isSelected(entry) ? "text-violet-400 font-medium" : ""}`}>
                                                {entry.name}
                                            </span>
                                            {isSelected(entry) && (
                                                <Check className="h-4 w-4 text-violet-400 flex-shrink-0" />
                                            )}
                                        </div>
                                        {!entry.is_directory && (
                                            <div className="text-xs text-muted-foreground flex items-center gap-2">
                                                <span>{formatSize(entry.size)}</span>
                                                <span>•</span>
                                                <span>{formatDate(entry.mod_time)}</span>
                                                {entry.extension && (
                                                    <>
                                                        <span>•</span>
                                                        <Badge variant="outline" className="text-[10px] px-1 py-0">
                                                            {entry.extension.toUpperCase()}
                                                        </Badge>
                                                    </>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                    {entry.is_directory && (
                                        <ChevronRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                                    )}
                                </button>
                            ))}
                        </div>
                    ) : (
                        <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                            <Folder className="h-12 w-12 mb-3 opacity-50" />
                            <p>No files found</p>
                            {mediaOnly && (
                                <p className="text-sm mt-1">Only showing media files</p>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Footer */}
            {selectedPath && (
                <div className="border-t border-border p-3 bg-muted/30">
                    <div className="flex items-center gap-2 text-sm">
                        <Film className="h-4 w-4 text-violet-400" />
                        <span className="text-muted-foreground">Selected:</span>
                        <span className="font-mono text-xs truncate flex-1">
                            {selectedPath.replace("file://", "")}
                        </span>
                    </div>
                </div>
            )}
        </div>
    );
}
