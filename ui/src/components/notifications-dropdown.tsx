"use client"

import { useState } from "react"
import { Bell, Check, Info, AlertTriangle, XCircle, CheckCircle2 } from "lucide-react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { formatDistanceToNow } from "date-fns"

import { Button } from "@/components/ui/button"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
    fetchNotifications,
    markNotificationRead,
    clearAllNotifications,
    Notification
} from "@/lib/api"
import { cn } from "@/lib/utils"

export function NotificationsDropdown() {
    const [isOpen, setIsOpen] = useState(false)
    const queryClient = useQueryClient()

    // Fetch notifications
    const { data: notifications = [], isLoading } = useQuery<Notification[]>({
        queryKey: ["notifications"],
        queryFn: async () => {
            return await fetchNotifications()
        },
        refetchInterval: 30000, // Poll every 30s
    })

    // Mark as read mutation
    const markReadMutation = useMutation({
        mutationFn: async (id: string) => {
            await markNotificationRead(id)
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["notifications"] })
        },
    })

    // Clear all mutation
    const clearAllMutation = useMutation({
        mutationFn: async () => {
            await clearAllNotifications()
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["notifications"] })
        },
    })

    const unreadCount = notifications.filter((n) => !n.is_read).length

    const getIcon = (type: string) => {
        switch (type) {
            case "success":
                return <CheckCircle2 className="h-4 w-4 text-green-500" />
            case "warning":
                return <AlertTriangle className="h-4 w-4 text-yellow-500" />
            case "error":
                return <XCircle className="h-4 w-4 text-red-500" />
            default:
                return <Info className="h-4 w-4 text-blue-500" />
        }
    }

    return (
        <Popover open={isOpen} onOpenChange={setIsOpen}>
            <PopoverTrigger asChild>
                <Button variant="ghost" size="icon" className="relative text-muted-foreground hover:text-foreground">
                    <Bell className="h-5 w-5" />
                    {unreadCount > 0 && (
                        <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-violet-500 ring-2 ring-background" />
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-0" align="end">
                <div className="flex items-center justify-between p-4 border-b">
                    <h4 className="font-semibold text-sm">Notifications</h4>
                    {unreadCount > 0 && (
                        <Button
                            variant="ghost"
                            size="sm"
                            className="text-xs h-auto py-1 px-2"
                            onClick={() => clearAllMutation.mutate()}
                            disabled={clearAllMutation.isPending}
                        >
                            Mark all read
                        </Button>
                    )}
                </div>
                <ScrollArea className="h-[400px]">
                    {isLoading ? (
                        <div className="p-4 text-center text-sm text-muted-foreground">
                            Loading...
                        </div>
                    ) : notifications.length === 0 ? (
                        <div className="p-8 text-center text-sm text-muted-foreground">
                            No notifications
                        </div>
                    ) : (
                        <div className="divide-y divide-border/50">
                            {notifications.map((notification) => (
                                <div
                                    key={notification.id}
                                    className={cn(
                                        "p-4 hover:bg-muted/50 transition-colors relative group",
                                        !notification.is_read && "bg-muted/20"
                                    )}
                                >
                                    <div className="flex gap-3">
                                        <div className="mt-1 flex-shrink-0">{getIcon(notification.type)}</div>
                                        <div className="flex-1 space-y-1">
                                            <p className={cn("text-sm leading-none", !notification.is_read && "font-medium")}>
                                                {notification.title}
                                            </p>
                                            <p className="text-sm text-muted-foreground line-clamp-2">
                                                {notification.message}
                                            </p>
                                            <p className="text-[10px] text-muted-foreground pt-1">
                                                {formatDistanceToNow(new Date(notification.created_at), {
                                                    addSuffix: true,
                                                })}
                                            </p>
                                        </div>
                                        {!notification.is_read && (
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity -mr-2 -mt-1"
                                                onClick={(e) => {
                                                    e.stopPropagation()
                                                    markReadMutation.mutate(notification.id)
                                                }}
                                                disabled={markReadMutation.isPending}
                                            >
                                                <Check className="h-3 w-3" />
                                                <span className="sr-only">Mark as read</span>
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </ScrollArea>
            </PopoverContent>
        </Popover>
    )
}
