"use client";

import { Link, usePathname } from "@/i18n/routing";
import Image from "next/image";
import {
    LayoutDashboard,
    Film,
    Radio,
    Repeat,
    Server,
    Sliders,
    Settings,
    AlertCircle,
} from "lucide-react";
import { useTranslations } from "next-intl";

export function Sidebar() {
    const t = useTranslations('common');

    // Note: next-intl useTranslations call must be inside the component
    // We cannot define navItems outside with translations

    const navItems = [
        { href: "/", label: t('dashboard'), icon: LayoutDashboard },
        { href: "/jobs", label: t('jobs'), icon: Film },
        { href: "/streams", label: t('streams'), icon: Radio },
        { href: "/restreams", label: t('restreams'), icon: Repeat },
        { href: "/workers", label: t('workers'), icon: Server },
        { href: "/errors", label: t('errors'), icon: AlertCircle },
        { href: "/profiles", label: t('profiles'), icon: Sliders },
        { href: "/settings", label: t('settings'), icon: Settings },
    ];

    return (
        <aside className="hidden lg:flex lg:flex-col lg:w-64 lg:fixed lg:inset-y-0 border-r border-border bg-card/50">
            {/* Logo */}
            <div className="flex items-center h-16 px-6 border-b border-border">
                <Link href="/" className="flex items-center gap-3 group">
                    <div className="relative">
                        <Image
                            src="/logo.png"
                            alt="WebEncode"
                            width={36}
                            height={36}
                            className="rounded-lg transition-transform group-hover:scale-105"
                        />
                        <div className="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full bg-emerald-400 border-2 border-card" />
                    </div>
                    <span className="text-xl font-bold text-gradient">
                        WebEncode
                    </span>
                </Link>
            </div>

            {/* Navigation */}
            <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto scrollbar-thin">
                {navItems.map((item) => (
                    <NavLink key={item.href} href={item.href} icon={item.icon}>
                        {item.label}
                    </NavLink>
                ))}
            </nav>

            {/* Footer */}
            <div className="p-4 border-t border-border">
                <div className="px-3 py-2 rounded-lg bg-muted/30">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <div className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
                        <span>System Healthy</span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">v1.0.0 • MIT License</p>
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
    const isActive = href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(href + "/");

    return (
        <Link
            href={href}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 group nav-item ${isActive
                ? "bg-muted/50 text-violet-400"
                : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
                }`}
        >
            <Icon
                className={`h-5 w-5 transition-colors ${isActive ? "text-violet-400" : "group-hover:text-violet-400"
                    }`}
            />
            <span className="text-sm font-medium">{children}</span>
            {isActive && (
                <span className="ml-auto h-2 w-2 rounded-full bg-violet-400" />
            )}
        </Link>
    );
}
