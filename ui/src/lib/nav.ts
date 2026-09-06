import {
    AlertCircle,
    Film,
    LayoutDashboard,
    Radio,
    Repeat,
    Server,
    Settings,
    Sliders,
} from "lucide-react";

/**
 * Primary navigation, shared by the desktop sidebar and the mobile drawer so
 * both stay in sync. `labelKey` is resolved against the `common` message
 * namespace by the rendering component.
 */
export interface NavItem {
    href: string;
    labelKey: string;
    icon: React.ElementType;
}

export const NAV_ITEMS: readonly NavItem[] = [
    { href: "/", labelKey: "dashboard", icon: LayoutDashboard },
    { href: "/jobs", labelKey: "jobs", icon: Film },
    { href: "/streams", labelKey: "streams", icon: Radio },
    { href: "/restreams", labelKey: "restreams", icon: Repeat },
    { href: "/workers", labelKey: "workers", icon: Server },
    { href: "/errors", labelKey: "errors", icon: AlertCircle },
    { href: "/profiles", labelKey: "profiles", icon: Sliders },
    { href: "/settings", labelKey: "settings", icon: Settings },
];

/** Width of the fixed desktop sidebar; the main column offsets by the same amount. */
export const SIDEBAR_WIDTH_CLASS = "lg:w-64";
export const SIDEBAR_OFFSET_CLASS = "lg:pl-64";

/** True when `href` is the section the current pathname belongs to. */
export function isActiveNavHref(href: string, pathname: string): boolean {
    if (href === "/") {
        return pathname === "/";
    }
    return pathname === href || pathname.startsWith(`${href}/`);
}
