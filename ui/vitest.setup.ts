import '@testing-library/jest-dom'

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
  }),
  usePathname: () => '/',
  useSearchParams: () => new URLSearchParams(),
}))

// Mock next-intl
vi.mock('next-intl', () => ({
  useTranslations: () => (key: string) => key,
  useLocale: () => 'en',
}))

// Mock next-intl/navigation for i18n routing
vi.mock('next-intl/routing', () => ({
  defineRouting: vi.fn(() => ({
    locales: ['en'],
    defaultLocale: 'en',
    localePrefix: 'always',
  })),
}))

vi.mock('next-intl/navigation', () => ({
  createNavigation: vi.fn(() => ({
    Link: ({ children, ...props }: { children: React.ReactNode; href: string }) => {
      const React = require('react')
      return React.createElement('a', props, children)
    },
    redirect: vi.fn(),
    usePathname: () => '/',
    useRouter: () => ({
      push: vi.fn(),
      replace: vi.fn(),
    }),
    getPathname: vi.fn(),
  })),
}))

// Mock i18n/routing module
vi.mock('@/i18n/routing', () => ({
  Link: ({ children, href }: { children: React.ReactNode; href: string }) => {
    const React = require('react')
    return React.createElement('a', { href }, children)
  },
  redirect: vi.fn(),
  usePathname: () => '/',
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
  }),
  getPathname: vi.fn(),
  routing: {
    locales: ['en'],
    defaultLocale: 'en',
    localePrefix: 'always',
  },
}))

// Mock next/image
vi.mock('next/image', () => ({
  default: ({ src, alt, ...props }: { src: string; alt: string }) => {
    const React = require('react')
    return React.createElement('img', { src, alt, ...props })
  },
}))
