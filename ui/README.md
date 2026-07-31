# WebEncode Dashboard

The web UI for [WebEncode](../README.md): a Next.js 16 App Router app that
monitors jobs, workers, live streams and restreams against the kernel REST API.

## Getting Started

```bash
npm install
npm run dev        # http://localhost:3000
```

Requests to `/api/v1/*` are rewritten to the kernel (see `next.config.ts`). Set
`INTERNAL_API_URL` if the kernel is not listening on `http://localhost:8090`.

## Scripts

| Script | Purpose |
| --- | --- |
| `npm run dev` | Development server |
| `npm run build` | Production build (`output: "standalone"`) |
| `npm run start` | Serve a production build |
| `npm run lint` | ESLint (`eslint-config-next`) |
| `npm run test:run` | Vitest suite, single run |
| `npm run test:coverage` | Vitest with V8 coverage |

## Structure

```
src/
├── app/[locale]/      # Localised routes (dashboard, jobs, streams, ...)
├── components/        # App components; components/ui holds the shadcn primitives
├── i18n/              # next-intl routing and request config
├── lib/               # API client, providers, nav/theme/app metadata constants
└── app/globals.css    # Theme tokens, component classes, utilities
messages/              # Translation bundles: en, de, es, fr, ja
```

## Theming

Colours are CSS custom properties in `src/app/globals.css`. Both a dark (default)
and a light theme are defined; the active one is the presence of the `dark` class
on `<html>`, set before first paint by the init script in `src/lib/theme.ts` and
toggled from the header. Prefer the semantic tokens (`primary`, `brand`,
`success`, `warning`, `danger`, `info`, `muted`, `border`, ...) over fixed
palette utilities such as `text-violet-400`, which only work on one theme.

## Internationalisation

Text lives in `messages/<locale>.json` and is read with `next-intl`. Add new keys
to every bundle, and format dates and numbers with next-intl's `useFormatter`
rather than hardcoding a locale.
