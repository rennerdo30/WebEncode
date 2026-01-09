"use client";

import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchStreamChat, sendStreamChat, ChatMessage } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    MessageCircle,
    Send,
    Loader2,
    Twitch,
    Youtube,
    Tv,
    RefreshCw,
} from "lucide-react";

interface ChatWidgetProps {
    streamId: string;
    className?: string;
}

function getPlatformIcon(platform: string) {
    const lower = platform.toLowerCase();
    if (lower.includes("twitch")) return <Twitch className="h-3 w-3 text-purple-400" />;
    if (lower.includes("youtube")) return <Youtube className="h-3 w-3 text-red-400" />;
    if (lower.includes("kick")) return <Tv className="h-3 w-3 text-green-400" />;
    if (lower.includes("rumble")) return <Tv className="h-3 w-3 text-emerald-400" />;
    return <MessageCircle className="h-3 w-3 text-muted-foreground" />;
}

function getPlatformColor(platform: string): string {
    const lower = platform.toLowerCase();
    if (lower.includes("twitch")) return "bg-purple-500/10 text-purple-400 border-purple-500/30";
    if (lower.includes("youtube")) return "bg-red-500/10 text-red-400 border-red-500/30";
    if (lower.includes("kick")) return "bg-green-500/10 text-green-400 border-green-500/30";
    if (lower.includes("rumble")) return "bg-emerald-500/10 text-emerald-400 border-emerald-500/30";
    return "bg-muted text-muted-foreground";
}

function formatTimestamp(timestamp: number): string {
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export function ChatWidget({ streamId, className }: ChatWidgetProps) {
    const [message, setMessage] = useState("");
    const scrollRef = useRef<HTMLDivElement>(null);
    const queryClient = useQueryClient();

    const { data: messages, isLoading, refetch, isRefetching } = useQuery({
        queryKey: ["stream-chat", streamId],
        queryFn: () => fetchStreamChat(streamId),
        refetchInterval: 3000, // Poll every 3 seconds
    });

    const sendMutation = useMutation({
        mutationFn: (msg: string) => sendStreamChat(streamId, { message: msg }),
        onSuccess: () => {
            setMessage("");
            queryClient.invalidateQueries({ queryKey: ["stream-chat", streamId] });
        },
    });

    // Auto-scroll to bottom when new messages arrive
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages]);

    const handleSend = () => {
        if (message.trim() && !sendMutation.isPending) {
            sendMutation.mutate(message.trim());
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    return (
        <Card className={`flex flex-col ${className || ""}`}>
            <CardHeader className="pb-3 flex flex-row items-center justify-between space-y-0">
                <CardTitle className="text-base flex items-center gap-2">
                    <MessageCircle className="h-4 w-4 text-violet-400" />
                    Live Chat
                </CardTitle>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => refetch()}
                    disabled={isRefetching}
                >
                    <RefreshCw className={`h-3.5 w-3.5 ${isRefetching ? "animate-spin" : ""}`} />
                </Button>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col p-0">
                {/* Messages Area */}
                <ScrollArea className="flex-1 px-4" ref={scrollRef}>
                    <div className="space-y-3 py-2">
                        {isLoading ? (
                            <div className="flex justify-center py-8">
                                <Loader2 className="h-6 w-6 animate-spin text-violet-400" />
                            </div>
                        ) : messages && messages.length > 0 ? (
                            messages.map((msg) => (
                                <ChatMessageItem key={msg.id} message={msg} />
                            ))
                        ) : (
                            <div className="text-center py-8 text-muted-foreground">
                                <MessageCircle className="h-10 w-10 mx-auto mb-2 opacity-30" />
                                <p className="text-sm">No messages yet</p>
                                <p className="text-xs mt-1">
                                    Messages from connected platforms will appear here
                                </p>
                            </div>
                        )}
                    </div>
                </ScrollArea>

                {/* Send Message Input */}
                <div className="p-3 border-t border-border">
                    <div className="flex gap-2">
                        <Input
                            placeholder="Send a message..."
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            onKeyDown={handleKeyDown}
                            disabled={sendMutation.isPending}
                            className="bg-muted/30 text-sm"
                        />
                        <Button
                            size="icon"
                            onClick={handleSend}
                            disabled={!message.trim() || sendMutation.isPending}
                            className="btn-gradient text-white shrink-0"
                        >
                            {sendMutation.isPending ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                                <Send className="h-4 w-4" />
                            )}
                        </Button>
                    </div>
                    {sendMutation.isError && (
                        <p className="text-xs text-red-400 mt-1">
                            Failed to send message
                        </p>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}

interface ChatMessageItemProps {
    message: ChatMessage;
}

function ChatMessageItem({ message }: ChatMessageItemProps) {
    return (
        <div className="flex gap-2 text-sm group animate-[slide-up_0.2s_ease-out]">
            <div className="shrink-0 pt-0.5">
                {getPlatformIcon(message.platform)}
            </div>
            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-0.5">
                    <span className="font-medium text-foreground truncate">
                        {message.author_name}
                    </span>
                    <Badge
                        variant="outline"
                        className={`text-[10px] px-1 py-0 ${getPlatformColor(message.platform)}`}
                    >
                        {message.platform}
                    </Badge>
                    <span className="text-[10px] text-muted-foreground ml-auto shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                        {formatTimestamp(message.timestamp)}
                    </span>
                </div>
                <p className="text-muted-foreground break-words">{message.content}</p>
            </div>
        </div>
    );
}

export default ChatWidget;
