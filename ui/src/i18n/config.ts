export const locales = ['en', 'de', 'ja', 'es', 'fr'] as const;
export const defaultLocale = 'en';

export type Locale = (typeof locales)[number];

export const localeNames: Record<Locale, string> = {
    en: 'English',
    de: 'Deutsch',
    ja: '日本語',
    es: 'Español',
    fr: 'Français',
};
