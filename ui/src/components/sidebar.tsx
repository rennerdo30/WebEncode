"use client";

import Image from "next/image";
import { useTranslations } from "next-intl";

import { Link, usePathname } from "@/i18n/routing";
import { APP_LICENSE, APP_NAME, APP_VERSION } from "@/lib/app-meta";
import { NAV_ITEMS, SIDEBAR_WIDTH_CLASS, isActiveNavHref } from "@/lib/nav";
import { cn } from "@/lib/utils";

export function Sidebar() {
    const t = useTranslations('common');

    return (
        <aside
            className={cn(
                "hidden lg:flex lg:flex-col lg:fixed lg:inset-y-0 border-r border-border bg-card/50",
                SIDEBAR_WIDTH_CLASS,
            )}
        >
            {/* Logo */}
            <div className="flex items-center h-16 px-6 border-b border-border">
                <Link
                    href="/"
                    className="flex items-center gap-3 group rounded-md focus-ring"
                >
                    <div className="relative">
                        <Image
                            src="/logo.png"
                            alt={APP_NAME}
                            width={36}
                            height={36}
                            className="rounded-lg transition-transform group-hover:scale-105 motion-reduce:transition-none"
                        />
                        <span className="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full bg-success border-2 border-card" />
                    </div>
                    <span className="text-xl font-bold tracking-tight text-gradient">
                        {APP_NAME}
                    </span>
                </Link>
            </div>

            {/* Navigation */}
            <nav
                aria-label={t('mainNavigation')}
                className="flex-1 px-3 py-4 space-y-1 overflow-y-auto scrollbar-thin"
            >
                {NAV_ITEMS.map((item) => (
                    <NavLink key={item.href} href={item.href} icon={item.icon}>
                        {t(item.labelKey)}
                    </NavLink>
                ))}
            </nav>

            {/* Footer */}
            <div className="p-4 border-t border-border">
                <div className="px-3 py-2 rounded-lg bg-muted/40">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span className="h-2 w-2 rounded-full bg-success animate-pulse motion-reduce:animate-none" />
                        <span>System Healthy</span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">{APP_VERSION} • {APP_LICENSE}</p>
                </div>
            </div>
        </aside>
    );
}

interface NavLinkProps {
    href: string;
    children: React.ReactNode;
    icon: React.ElementType;
}

function NavLink({ href, children, icon: Icon }: NavLinkProps) {
    const pathname = usePathname();
    const isActive = isActiveNavHref(href, pathname);

    return (
        <Link
            href={href}
            aria-current={isActive ? "page" : undefined}
            className={cn(
                "nav-item group flex items-center gap-3 px-3 py-2.5 rounded-lg",
                "focus-ring",
                isActive
                    ? "bg-accent text-brand"
                    : "text-muted-foreground hover:text-foreground",
            )}
        >
            <Icon
                aria-hidden="true"
                className={cn(
                    "h-5 w-5 transition-colors",
                    isActive ? "text-brand" : "group-hover:text-brand",
                )}
            />
            <span className="text-sm font-medium">{children}</span>
            {isActive && (
                <span aria-hidden="true" className="ml-auto h-2 w-2 rounded-full bg-brand" />
            )}
        </Link>
    );
}
