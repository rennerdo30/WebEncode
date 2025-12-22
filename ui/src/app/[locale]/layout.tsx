import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/app/globals.css";
import { Providers } from "@/lib/providers";
import { Link } from "@/i18n/routing";
import Image from "next/image";
import { Menu, User } from "lucide-react";
import { NotificationsDropdown } from "@/components/notifications-dropdown";
import { Sidebar } from "@/components/sidebar";
import { GlobalErrorTracker } from "@/components/global-error-tracker";
import { Toaster } from "sonner";
import { NextIntlClientProvider } from 'next-intl';
import { getMessages, setRequestLocale } from 'next-intl/server';
import { locales, type Locale } from '@/i18n/config';
import { notFound } from 'next/navigation';

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

    return (
        <html lang={locale} className="dark">
            <body className={`${inter.variable} font-sans`}>
                <NextIntlClientProvider messages={messages}>
                    <Providers>
                        <Toaster position="top-right" richColors closeButton />
                        <GlobalErrorTracker />
                        <div className="min-h-screen bg-background text-foreground flex">
                            {/* Sidebar */}
                            <Sidebar />

                            {/* Main Content */}
                            <div className="flex-1 lg:pl-64">
                                {/* Top Header */}
                                <header className="sticky top-0 z-40 h-16 border-b border-border bg-background/80 backdrop-blur-xl">
                                    <div className="flex items-center justify-between h-full px-4 lg:px-8">
                                        {/* Mobile menu button */}
                                        <button className="lg:hidden p-2 rounded-lg hover:bg-muted transition-colors">
                                            <Menu className="h-5 w-5" />
                                        </button>

                                        {/* Mobile logo */}
                                        <Link href="/" className="lg:hidden flex items-center gap-2">
                                            <Image
                                                src="/logo.png"
                                                alt="WebEncode"
                                                width={28}
                                                height={28}
                                                className="rounded"
                                            />
                                            <span className="font-bold text-gradient">WebEncode</span>
                                        </Link>

                                        {/* Spacer for desktop */}
                                        <div className="hidden lg:block" />

                                        {/* Right side actions */}
                                        <div className="flex items-center gap-2">
                                            {/* Notifications */}
                                            <NotificationsDropdown />

                                            {/* User */}
                                            <button className="flex items-center gap-2 p-1.5 pr-3 rounded-lg hover:bg-muted transition-colors">
                                                <div className="h-8 w-8 rounded-full bg-violet-600 flex items-center justify-center">
                                                    <User className="h-4 w-4 text-white" />
                                                </div>
                                                <span className="hidden sm:block text-sm font-medium">Admin</span>
                                            </button>
                                        </div>
                                    </div>
                                </header>

                                {/* Page Content */}
                                <main className="p-4 lg:p-8">
                                    {children}
                                </main>

                                {/* Footer */}
                                <footer className="border-t border-border p-4 lg:p-8 mt-auto">
                                    <div className="flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
                                        <p>© 2025 WebEncode. Open-source under MIT License.</p>
                                        <div className="flex items-center gap-4">
                                            <Link href="https://github.com/rennerdo30/webencode" className="hover:text-foreground transition-colors">
                                                GitHub
                                            </Link>
                                            <Link href="/docs" className="hover:text-foreground transition-colors">
                                                Documentation
                                            </Link>
                                            <Link href="/api" className="hover:text-foreground transition-colors">
                                                API
                                            </Link>
                                        </div>
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
