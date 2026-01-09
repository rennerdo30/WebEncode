import '@testing-library/jest-dom'

// Mock scrollIntoView for Radix UI components
Element.prototype.scrollIntoView = vi.fn()

// Mock ResizeObserver for Radix UI components
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock PointerEvent for Radix UI components
class MockPointerEvent extends Event {
  button: number
  ctrlKey: boolean
  pointerType: string

  constructor(type: string, props: PointerEventInit = {}) {
    super(type, props)
    this.button = props.button ?? 0
    this.ctrlKey = props.ctrlKey ?? false
    this.pointerType = props.pointerType ?? 'mouse'
  }
}
global.PointerEvent = MockPointerEvent as typeof PointerEvent

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
