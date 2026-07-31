import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/app/globals.css";
import { Providers } from "@/lib/providers";
import { Link } from "@/i18n/routing";
import Image from "next/image";
import { User } from "lucide-react";
import { NotificationsDropdown } from "@/components/notifications-dropdown";
import { Sidebar } from "@/components/sidebar";
import { MobileNav } from "@/components/mobile-nav";
import { ThemeToggle } from "@/components/theme-toggle";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { GlobalErrorTracker } from "@/components/global-error-tracker";
import { Toaster } from "sonner";
import { NextIntlClientProvider } from 'next-intl';
import { getMessages, getTranslations, setRequestLocale } from 'next-intl/server';
import { locales, type Locale } from '@/i18n/config';
import { notFound } from 'next/navigation';
import { APP_NAME, API_DOCS_URL, DOCS_URL, REPO_URL } from "@/lib/app-meta";
import { SIDEBAR_OFFSET_CLASS } from "@/lib/nav";
import { THEME_INIT_SCRIPT } from "@/lib/theme";

const inter = Inter({
    subsets: ["latin"],
    variable: "--font-inter",
});

export const metadata: Metadata = {
    title: "WebEncode Dashboard",
    description: "Distributed Video Transcoding Platform - Self-hosted, open-source video processing",
    keywords: ["video encoding", "transcoding", "ffmpeg", "live streaming", "distributed"],
    authors: [{ name: "WebEncode Team" }],
};

/** Anchor for the "skip to content" link. */
const MAIN_CONTENT_ID = "main-content";

/** Page gutters and the ceiling that keeps content readable on wide screens. */
const SHELL_WIDTH_CLASS = "w-full max-w-content-max mx-auto";
const SHELL_PADDING_CLASS = "p-4 sm:p-6 lg:p-8";

const FOOTER_LINK_CLASS = "rounded-sm transition-colors hover:text-foreground focus-ring";

export function generateStaticParams() {
    return locales.map((locale) => ({ locale }));
}

interface LayoutProps {
    children: React.ReactNode;
    params: Promise<{ locale: string }>;
}

export default async function LocaleLayout({
    children,
    params
}: LayoutProps) {
    const { locale } = await params;

    if (!locales.includes(locale as Locale)) {
        notFound();
    }

    setRequestLocale(locale);
    const messages = await getMessages();
    const t = await getTranslations({ locale, namespace: 'common' });
    const tFooter = await getTranslations({ locale, namespace: 'footer' });
    const year = new Intl.NumberFormat(locale, { useGrouping: false }).format(new Date().getFullYear());

    return (
        // The theme class is rewritten before paint by THEME_INIT_SCRIPT.
        <html lang={locale} className="dark" suppressHydrationWarning>
            <head>
                <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
            </head>
            <body className={`${inter.variable} font-sans`}>
                <NextIntlClientProvider messages={messages}>
                    <Providers>
                        <Toaster position="top-right" richColors closeButton />
                        <GlobalErrorTracker />
                        <a
                            href={`#${MAIN_CONTENT_ID}`}
                            className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground"
                        >
                            {t('skipToContent')}
                        </a>
                        <div className="min-h-screen bg-background text-foreground flex">
                            {/* Sidebar */}
                            <Sidebar />

                            {/* Main Content */}
                            <div className={`flex flex-col min-h-screen flex-1 min-w-0 ${SIDEBAR_OFFSET_CLASS}`}>
                                {/* Top Header */}
                                <header className="sticky top-0 z-40 h-16 border-b border-border bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/70">
                                    <div className="flex items-center justify-between h-full gap-2 px-4 lg:px-8">
                                        <div className="flex items-center gap-2 min-w-0">
                                            <MobileNav />

                                            {/* Mobile logo */}
                                            <Link
                                                href="/"
                                                className="lg:hidden flex items-center gap-2 rounded-md focus-ring"
                                            >
                                                <Image
                                                    src="/logo.png"
                                                    alt=""
                                                    width={28}
                                                    height={28}
                                                    className="rounded"
                                                />
                                                <span className="font-bold text-gradient">{APP_NAME}</span>
                                            </Link>
                                        </div>

                                        {/* Right side actions */}
                                        <div className="flex items-center gap-1">
                                            <LocaleSwitcher />
                                            <ThemeToggle />
                                            <NotificationsDropdown />

                                            {/* User */}
                                            <button
                                                type="button"
                                                className="flex items-center gap-2 p-1.5 pr-3 rounded-lg transition-colors hover:bg-accent focus-ring"
                                            >
                                                <span className="h-8 w-8 rounded-full bg-primary flex items-center justify-center">
                                                    <User className="h-4 w-4 text-primary-foreground" aria-hidden="true" />
                                                </span>
                                                {/* Visually hidden below sm, but always part of the accessible name. */}
                                                <span className="sr-only sm:not-sr-only text-sm font-medium">{t('admin')}</span>
                                            </button>
                                        </div>
                                    </div>
                                </header>

                                {/* Page Content */}
                                <main id={MAIN_CONTENT_ID} className={`flex-1 ${SHELL_WIDTH_CLASS} ${SHELL_PADDING_CLASS}`}>
                                    {children}
                                </main>

                                {/* Footer */}
                                <footer className="border-t border-border">
                                    <div className={`${SHELL_WIDTH_CLASS} ${SHELL_PADDING_CLASS} flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground`}>
                                        <p>{tFooter('copyright', { year, name: APP_NAME })}</p>
                                        <nav aria-label={tFooter('label')}>
                                            <ul className="flex items-center gap-4">
                                                <li>
                                                    <a
                                                        href={REPO_URL}
                                                        target="_blank"
                                                        rel="noreferrer noopener"
                                                        className={FOOTER_LINK_CLASS}
                                                    >
                                                        GitHub
                                                    </a>
                                                </li>
                                                <li>
                                                    <a
                                                        href={DOCS_URL}
                                                        target="_blank"
                                                        rel="noreferrer noopener"
                                                        className={FOOTER_LINK_CLASS}
                                                    >
                                                        {tFooter('documentation')}
                                                    </a>
                                                </li>
                                                <li>
                                                    <a
                                                        href={API_DOCS_URL}
                                                        target="_blank"
                                                        rel="noreferrer noopener"
                                                        className={FOOTER_LINK_CLASS}
                                                    >
                                                        {tFooter('api')}
                                                    </a>
                                                </li>
                                            </ul>
                                        </nav>
                                    </div>
                                </footer>
                            </div>
                        </div>
                    </Providers>
                </NextIntlClientProvider>
            </body>
        </html>
    );
}
