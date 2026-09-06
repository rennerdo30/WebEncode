"use client";

import { Check, Languages } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Link, usePathname } from "@/i18n/routing";
import { localeNames, locales, type Locale } from "@/i18n/config";
import { cn } from "@/lib/utils";

/**
 * Lets users reach the translations the app already ships. Rendered as plain
 * locale-prefixed links so it keeps working without client-side routing.
 */
export function LocaleSwitcher() {
    const t = useTranslations("common");
    const activeLocale = useLocale();
    const pathname = usePathname();

    return (
        <Popover>
            <PopoverTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    aria-label={t("language")}
                    title={t("language")}
                    className="text-muted-foreground hover:text-foreground"
                >
                    <Languages className="h-5 w-5" aria-hidden="true" />
                </Button>
            </PopoverTrigger>
            <PopoverContent align="end" className="w-48 p-1">
                <ul className="space-y-0.5" aria-label={t("language")}>
                    {locales.map((locale) => {
                        const isActive = locale === activeLocale;
                        return (
                            <li key={locale}>
                                <Link
                                    href={pathname}
                                    locale={locale as Locale}
                                    lang={locale}
                                    aria-current={isActive ? "true" : undefined}
                                    className={cn(
                                        "flex items-center justify-between gap-2 rounded-md px-3 py-2 text-sm transition-colors",
                                        "hover:bg-accent hover:text-accent-foreground",
                                        isActive ? "font-medium text-brand" : "text-foreground",
                                    )}
                                >
                                    {localeNames[locale]}
                                    {isActive && <Check className="h-4 w-4" aria-hidden="true" />}
                                </Link>
                            </li>
                        );
                    })}
                </ul>
            </PopoverContent>
        </Popover>
    );
}
