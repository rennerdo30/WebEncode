"use client";

import { useState, useCallback, useRef } from "react";
import { useMutation } from "@tanstack/react-query";
import { uploadFile, reportError, UploadProgress, UploadResponse } from "@/lib/api";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import {
    Upload,
    File,
    Film,
    Music,
    Image,
    X,
    CheckCircle2,
    AlertCircle,
    Loader2,
    CloudUpload,
} from "lucide-react";

interface FileUploadProps {
    /** Called when upload completes successfully */
    onUploadComplete: (url: string, response: UploadResponse) => void;
    /** Accepted file types (default: video/*,audio/*) */
    accept?: string;
    /** Maximum file size in bytes (default: 10GB) */
    maxSize?: number;
}

type UploadState = "idle" | "selected" | "uploading" | "complete" | "error";

/**
 * FileUpload component for uploading video/audio files to storage
 * Features drag-and-drop, progress tracking, and file validation
 */
export function FileUpload({
    onUploadComplete,
    accept = "video/*,audio/*,.mp4,.mkv,.avi,.mov,.wmv,.flv,.webm,.m4v,.mpeg,.mpg,.mp3,.wav,.flac,.ogg,.aac,.wma,.m4a,.opus",
    maxSize = 10 * 1024 * 1024 * 1024, // 10GB
}: FileUploadProps) {
    const [state, setState] = useState<UploadState>("idle");
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [progress, setProgress] = useState<UploadProgress>({ loaded: 0, total: 0, percentage: 0 });
    const [error, setError] = useState<string | null>(null);
    const [isDragging, setIsDragging] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const uploadMutation = useMutation({
        mutationFn: (file: File) =>
            uploadFile(file, {
                onProgress: setProgress,
            }),
        onSuccess: (response) => {
            setState("complete");
            if (selectedFile) {
                onUploadComplete(response.url, response);
            }
        },
        onError: (err) => {
            setState("error");
            const msg = err instanceof Error ? err.message : "Upload failed";
            setError(msg);
            reportError(msg, "component:file-upload", err instanceof Error ? err.stack : undefined);
        },
    });

    const handleFileSelect = useCallback(
        (file: File) => {
            // Validate file size
            if (file.size > maxSize) {
                setError(`File size exceeds maximum of ${formatSize(maxSize)}`);
                setState("error");
                return;
            }

            // Validate file type
            const isVideo = file.type.startsWith("video/");
            const isAudio = file.type.startsWith("audio/");
            const videoExtensions = ["mp4", "mkv", "avi", "mov", "wmv", "flv", "webm", "m4v", "mpeg", "mpg"];
            const audioExtensions = ["mp3", "wav", "flac", "ogg", "aac", "wma", "m4a", "opus"];
            const ext = file.name.split(".").pop()?.toLowerCase() || "";
            const isValidExtension = videoExtensions.includes(ext) || audioExtensions.includes(ext);

            if (!isVideo && !isAudio && !isValidExtension) {
                setError("Please select a video or audio file");
                setState("error");
                return;
            }

            setSelectedFile(file);
            setError(null);
            setState("selected");
        },
        [maxSize]
    );

    const handleDrop = useCallback(
        (e: React.DragEvent<HTMLDivElement>) => {
            e.preventDefault();
            setIsDragging(false);

            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleFileSelect(files[0]);
            }
        },
        [handleFileSelect]
    );

    const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragging(false);
    }, []);

    const handleInputChange = useCallback(
        (e: React.ChangeEvent<HTMLInputElement>) => {
            const files = e.target.files;
            if (files && files.length > 0) {
                handleFileSelect(files[0]);
            }
        },
        [handleFileSelect]
    );

    const handleBrowseClick = () => {
        fileInputRef.current?.click();
    };

    const handleUpload = () => {
        if (selectedFile) {
            setState("uploading");
            setProgress({ loaded: 0, total: selectedFile.size, percentage: 0 });
            uploadMutation.mutate(selectedFile);
        }
    };

    const handleReset = () => {
        setState("idle");
        setSelectedFile(null);
        setProgress({ loaded: 0, total: 0, percentage: 0 });
        setError(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
    };

    const getFileIcon = (file: File) => {
        if (file.type.startsWith("video/")) {
            return <Film className="h-8 w-8 text-violet-400" />;
        }
        if (file.type.startsWith("audio/")) {
            return <Music className="h-8 w-8 text-cyan-400" />;
        }
        if (file.type.startsWith("image/")) {
            return <Image className="h-8 w-8 text-emerald-400" />;
        }
        return <File className="h-8 w-8 text-muted-foreground" />;
    };

    const formatSize = (bytes: number): string => {
        if (bytes === 0) return "0 B";
        const units = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
    };

    return (
        <div className="space-y-4">
            <input
                ref={fileInputRef}
                type="file"
                accept={accept}
                onChange={handleInputChange}
                className="hidden"
                id="file-upload-input"
            />

            {/* Drop Zone */}
            {state === "idle" && (
                <div
                    onDrop={handleDrop}
                    onDragOver={handleDragOver}
                    onDragLeave={handleDragLeave}
                    onClick={handleBrowseClick}
                    className={`
                        relative border-2 border-dashed rounded-lg p-8 text-center cursor-pointer
                        transition-all duration-200
                        ${isDragging
                            ? "border-violet-500 bg-violet-500/10"
                            : "border-border hover:border-violet-500/50 hover:bg-muted/50"
                        }
                    `}
                >
                    <div className="flex flex-col items-center gap-4">
                        <div
                            className={`
                                p-4 rounded-full transition-colors
                                ${isDragging ? "bg-violet-500/20" : "bg-muted"}
                            `}
                        >
                            <CloudUpload
                                className={`h-10 w-10 ${isDragging ? "text-violet-400" : "text-muted-foreground"}`}
                            />
                        </div>
                        <div>
                            <p className="text-lg font-medium">
                                {isDragging ? "Drop your file here" : "Drag and drop your video file"}
                            </p>
                            <p className="text-sm text-muted-foreground mt-1">
                                or click to browse • Max {formatSize(maxSize)}
                            </p>
                        </div>
                        <div className="flex gap-2 flex-wrap justify-center text-xs text-muted-foreground">
                            <span className="px-2 py-1 bg-muted rounded">MP4</span>
                            <span className="px-2 py-1 bg-muted rounded">MKV</span>
                            <span className="px-2 py-1 bg-muted rounded">AVI</span>
                            <span className="px-2 py-1 bg-muted rounded">MOV</span>
                            <span className="px-2 py-1 bg-muted rounded">WebM</span>
                            <span className="px-2 py-1 bg-muted rounded">MP3</span>
                            <span className="px-2 py-1 bg-muted rounded">WAV</span>
                        </div>
                    </div>
                </div>
            )}

            {/* Selected File */}
            {state === "selected" && selectedFile && (
                <div className="border border-border rounded-lg p-4 bg-muted/30">
                    <div className="flex items-start gap-4">
                        <div className="p-3 bg-muted rounded-lg">
                            {getFileIcon(selectedFile)}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="font-medium truncate">{selectedFile.name}</p>
                            <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
                                <span>{formatSize(selectedFile.size)}</span>
                                <span>•</span>
                                <span>{selectedFile.type || "Unknown type"}</span>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={handleReset}
                            className="text-muted-foreground hover:text-foreground"
                        >
                            <X className="h-4 w-4" />
                        </Button>
                    </div>
                    <div className="flex gap-2 mt-4">
                        <Button onClick={handleUpload} className="flex-1">
                            <Upload className="h-4 w-4 mr-2" />
                            Upload File
                        </Button>
                        <Button variant="outline" onClick={handleBrowseClick}>
                            Choose Different File
                        </Button>
                    </div>
                </div>
            )}

            {/* Uploading State */}
            {state === "uploading" && selectedFile && (
                <div className="border border-border rounded-lg p-4 bg-muted/30">
                    <div className="flex items-start gap-4">
                        <div className="p-3 bg-violet-500/20 rounded-lg">
                            <Loader2 className="h-8 w-8 text-violet-400 animate-spin" />
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="font-medium truncate">{selectedFile.name}</p>
                            <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
                                <span>
                                    {formatSize(progress.loaded)} / {formatSize(progress.total)}
                                </span>
                                <span>•</span>
                                <span className="text-violet-400 font-medium">{progress.percentage}%</span>
                            </div>
                            <div className="mt-3">
                                <Progress value={progress.percentage} className="h-2" />
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Complete State */}
            {state === "complete" && selectedFile && (
                <div className="border border-emerald-500/30 rounded-lg p-4 bg-emerald-500/10">
                    <div className="flex items-start gap-4">
                        <div className="p-3 bg-emerald-500/20 rounded-lg">
                            <CheckCircle2 className="h-8 w-8 text-emerald-400" />
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="font-medium text-emerald-400">Upload Complete</p>
                            <p className="text-sm text-muted-foreground mt-1 truncate">
                                {selectedFile.name}
                            </p>
                        </div>
                        <Button variant="outline" size="sm" onClick={handleReset}>
                            Upload Another
                        </Button>
                    </div>
                </div>
            )}

            {/* Error State */}
            {state === "error" && (
                <div className="border border-red-500/30 rounded-lg p-4 bg-red-500/10">
                    <div className="flex items-start gap-4">
                        <div className="p-3 bg-red-500/20 rounded-lg">
                            <AlertCircle className="h-8 w-8 text-red-400" />
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="font-medium text-red-400">Upload Failed</p>
                            <p className="text-sm text-red-400/80 mt-1">{error}</p>
                        </div>
                        <Button variant="outline" size="sm" onClick={handleReset}>
                            Try Again
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
}
