"use client";

import { useState } from "react";
import Image from "next/image";
import { Menu } from "lucide-react";
import { useTranslations } from "next-intl";

import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { Link, usePathname } from "@/i18n/routing";
import { APP_NAME } from "@/lib/app-meta";
import { NAV_ITEMS, isActiveNavHref } from "@/lib/nav";
import { cn } from "@/lib/utils";

/**
 * Drawer version of the primary navigation for screens below the `lg`
 * breakpoint, where the fixed sidebar is hidden.
 */
export function MobileNav() {
    const t = useTranslations("common");
    const pathname = usePathname();
    const [open, setOpen] = useState(false);

    return (
        <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger
                className={cn(
                    "lg:hidden inline-flex h-10 w-10 items-center justify-center rounded-lg",
                    "text-muted-foreground transition-colors hover:bg-accent hover:text-foreground",
                    "focus-ring",
                )}
                aria-label={t("openNavigation")}
            >
                <Menu className="h-5 w-5" aria-hidden="true" />
            </SheetTrigger>
            <SheetContent side="left" className="w-72 p-0">
                <div className="flex h-16 items-center gap-3 border-b border-border px-6">
                    <Image src="/logo.png" alt="" width={28} height={28} className="rounded-lg" />
                    <SheetTitle className="text-lg font-bold text-brand">{APP_NAME}</SheetTitle>
                </div>
                <nav aria-label={t("mainNavigation")} className="space-y-1 p-3">
                    {NAV_ITEMS.map(({ href, labelKey, icon: Icon }) => {
                        const isActive = isActiveNavHref(href, pathname);
                        return (
                            <Link
                                key={href}
                                href={href}
                                onClick={() => setOpen(false)}
                                aria-current={isActive ? "page" : undefined}
                                className={cn(
                                    "nav-item flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium",
                                    isActive
                                        ? "bg-accent text-brand"
                                        : "text-muted-foreground hover:text-foreground",
                                )}
                            >
                                <Icon className="h-5 w-5" aria-hidden="true" />
                                <span>{t(labelKey)}</span>
                            </Link>
                        );
                    })}
                </nav>
            </SheetContent>
        </Sheet>
    );
}
