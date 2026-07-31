/**
 * Theme plumbing.
 *
 * The dashboard ships a dark and a light theme. Both are defined as CSS custom
 * property sets in `globals.css`; the active one is selected by the presence of
 * the `dark` class on `<html>`. Dark stays the default (it is what the app was
 * designed around) but an explicit choice, and otherwise the OS preference,
 * wins.
 */

export const THEME_STORAGE_KEY = "webencode-theme";
export const DARK_CLASS = "dark";

/** Suppresses colour transitions for the duration of a theme swap. */
export const THEME_SWITCHING_CLASS = "theme-switching";
export const THEME_SWITCHING_MS = 120;

export const THEMES = ["light", "dark"] as const;
export type Theme = (typeof THEMES)[number];

export const DEFAULT_THEME: Theme = "dark";

export function isTheme(value: unknown): value is Theme {
    return typeof value === "string" && (THEMES as readonly string[]).includes(value);
}

export function applyTheme(theme: Theme): void {
    document.documentElement.classList.toggle(DARK_CLASS, theme === "dark");
    document.documentElement.style.colorScheme = theme;
}

/*
  The active theme lives on the DOM (the `dark` class), which makes `<html>` the
  single source of truth for both the pre-paint init script and React. It is
  exposed as an external store so components can read it with
  `useSyncExternalStore` instead of syncing it into state from an effect.
*/
type ThemeListener = () => void;
const themeListeners = new Set<ThemeListener>();

export function subscribeToTheme(listener: ThemeListener): () => void {
    themeListeners.add(listener);

    // Keep tabs in sync when the preference is changed in another one.
    const onStorage = (event: StorageEvent) => {
        if (event.key !== THEME_STORAGE_KEY) return;
        if (isTheme(event.newValue)) {
            applyTheme(event.newValue);
        }
        listener();
    };
    window.addEventListener("storage", onStorage);

    return () => {
        themeListeners.delete(listener);
        window.removeEventListener("storage", onStorage);
    };
}

export function getThemeSnapshot(): Theme {
    return document.documentElement.classList.contains(DARK_CLASS) ? "dark" : "light";
}

/** During SSR and hydration the markup always reflects the default theme. */
export function getServerThemeSnapshot(): Theme {
    return DEFAULT_THEME;
}

/**
 * Applies `theme`, persists it and notifies subscribers. Colour transitions are
 * suppressed for the duration of the swap so the whole page does not animate.
 */
export function setTheme(theme: Theme): void {
    const root = document.documentElement;
    root.classList.add(THEME_SWITCHING_CLASS);
    applyTheme(theme);
    try {
        window.localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
        // Storage disabled (private mode): the choice just will not persist.
    }
    themeListeners.forEach((listener) => listener());
    window.setTimeout(() => root.classList.remove(THEME_SWITCHING_CLASS), THEME_SWITCHING_MS);
}

/**
 * Runs before first paint so the stored theme is applied without a flash of the
 * server-rendered default. Kept as a string because it has to be inlined into
 * the document head.
 */
export const THEME_INIT_SCRIPT = `(function(){try{var s=localStorage.getItem('${THEME_STORAGE_KEY}');var t=(s==='light'||s==='dark')?s:(window.matchMedia('(prefers-color-scheme: light)').matches?'light':'${DEFAULT_THEME}');document.documentElement.classList.toggle('${DARK_CLASS}',t==='dark');document.documentElement.style.colorScheme=t;}catch(e){}})();`;
