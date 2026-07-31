"use client";

import { useSyncExternalStore } from "react";
import { Moon, Sun } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import {
    getServerThemeSnapshot,
    getThemeSnapshot,
    setTheme,
    subscribeToTheme,
} from "@/lib/theme";

/**
 * Switches between the dark and light theme. The active theme is read straight
 * from `<html>`, which the pre-paint init script has already set, so the button
 * never disagrees with what is on screen.
 */
export function ThemeToggle() {
    const t = useTranslations("common");
    const theme = useSyncExternalStore(
        subscribeToTheme,
        getThemeSnapshot,
        getServerThemeSnapshot,
    );

    const isDark = theme === "dark";
    const label = isDark ? t("switchToLight") : t("switchToDark");

    return (
        <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => setTheme(isDark ? "light" : "dark")}
            aria-label={label}
            title={label}
            className="text-muted-foreground hover:text-foreground"
        >
            {isDark ? (
                <Moon className="h-5 w-5" aria-hidden="true" />
            ) : (
                <Sun className="h-5 w-5" aria-hidden="true" />
            )}
        </Button>
    );
}
