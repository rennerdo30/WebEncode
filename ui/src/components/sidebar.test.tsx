import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Sidebar } from './sidebar'

// Mock the usePathname hook to return different paths
let mockPathname = '/'
vi.mock('@/i18n/routing', () => ({
  Link: ({ children, href, className }: { children: React.ReactNode; href: string; className?: string }) => {
    const React = require('react')
    return React.createElement('a', { href, className }, children)
  },
  usePathname: () => mockPathname,
}))

describe('Sidebar', () => {
  beforeEach(() => {
    mockPathname = '/'
  })

  describe('rendering', () => {
    it('should render the sidebar', () => {
      render(<Sidebar />)
      expect(screen.getByRole('complementary')).toBeInTheDocument()
    })

    it('should render the logo', () => {
      render(<Sidebar />)
      const logo = screen.getByAltText('WebEncode')
      expect(logo).toBeInTheDocument()
    })

    it('should render WebEncode text', () => {
      render(<Sidebar />)
      expect(screen.getByText('WebEncode')).toBeInTheDocument()
    })
  })

  describe('navigation items', () => {
    it('should render dashboard link', () => {
      render(<Sidebar />)
      expect(screen.getByText('dashboard')).toBeInTheDocument()
    })

    it('should render jobs link', () => {
      render(<Sidebar />)
      expect(screen.getByText('jobs')).toBeInTheDocument()
    })

    it('should render streams link', () => {
      render(<Sidebar />)
      expect(screen.getByText('streams')).toBeInTheDocument()
    })

    it('should render restreams link', () => {
      render(<Sidebar />)
      expect(screen.getByText('restreams')).toBeInTheDocument()
    })

    it('should render workers link', () => {
      render(<Sidebar />)
      expect(screen.getByText('workers')).toBeInTheDocument()
    })

    it('should render errors link', () => {
      render(<Sidebar />)
      expect(screen.getByText('errors')).toBeInTheDocument()
    })

    it('should render profiles link', () => {
      render(<Sidebar />)
      expect(screen.getByText('profiles')).toBeInTheDocument()
    })

    it('should render settings link', () => {
      render(<Sidebar />)
      expect(screen.getByText('settings')).toBeInTheDocument()
    })
  })

  describe('navigation links', () => {
    it('should have correct href for dashboard', () => {
      render(<Sidebar />)
      const dashboardLink = screen.getByText('dashboard').closest('a')
      expect(dashboardLink).toHaveAttribute('href', '/')
    })

    it('should have correct href for jobs', () => {
      render(<Sidebar />)
      const jobsLink = screen.getByText('jobs').closest('a')
      expect(jobsLink).toHaveAttribute('href', '/jobs')
    })

    it('should have correct href for streams', () => {
      render(<Sidebar />)
      const streamsLink = screen.getByText('streams').closest('a')
      expect(streamsLink).toHaveAttribute('href', '/streams')
    })

    it('should have correct href for restreams', () => {
      render(<Sidebar />)
      const restreamsLink = screen.getByText('restreams').closest('a')
      expect(restreamsLink).toHaveAttribute('href', '/restreams')
    })

    it('should have correct href for workers', () => {
      render(<Sidebar />)
      const workersLink = screen.getByText('workers').closest('a')
      expect(workersLink).toHaveAttribute('href', '/workers')
    })

    it('should have correct href for errors', () => {
      render(<Sidebar />)
      const errorsLink = screen.getByText('errors').closest('a')
      expect(errorsLink).toHaveAttribute('href', '/errors')
    })

    it('should have correct href for profiles', () => {
      render(<Sidebar />)
      const profilesLink = screen.getByText('profiles').closest('a')
      expect(profilesLink).toHaveAttribute('href', '/profiles')
    })

    it('should have correct href for settings', () => {
      render(<Sidebar />)
      const settingsLink = screen.getByText('settings').closest('a')
      expect(settingsLink).toHaveAttribute('href', '/settings')
    })
  })

  describe('footer', () => {
    it('should display system health status', () => {
      render(<Sidebar />)
      expect(screen.getByText('System Healthy')).toBeInTheDocument()
    })

    it('should display version info', () => {
      render(<Sidebar />)
      expect(screen.getByText(/v1\.0\.0/)).toBeInTheDocument()
    })

    it('should display license info', () => {
      render(<Sidebar />)
      expect(screen.getByText(/MIT License/)).toBeInTheDocument()
    })
  })

  describe('styling', () => {
    it('should be hidden on smaller screens', () => {
      render(<Sidebar />)
      const aside = screen.getByRole('complementary')
      expect(aside).toHaveClass('hidden')
      expect(aside).toHaveClass('lg:flex')
    })

    it('should have fixed width on large screens', () => {
      render(<Sidebar />)
      const aside = screen.getByRole('complementary')
      expect(aside).toHaveClass('lg:w-64')
    })

    it('should be fixed position', () => {
      render(<Sidebar />)
      const aside = screen.getByRole('complementary')
      expect(aside).toHaveClass('lg:fixed')
    })
  })
})
