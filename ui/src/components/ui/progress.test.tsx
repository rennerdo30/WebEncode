import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Progress } from './progress'

describe('Progress', () => {
  describe('rendering', () => {
    it('should render progress bar', () => {
      render(<Progress value={50} />)
      const progressbar = screen.getByRole('progressbar')
      expect(progressbar).toBeInTheDocument()
    })

    it('should pass value prop to root component', () => {
      const { container } = render(<Progress value={75} />)
      const progressbar = screen.getByRole('progressbar')
      // Verify the progressbar element exists and has expected structure
      expect(progressbar).toBeInTheDocument()
      // The indicator should have the correct transform based on value
      const indicator = container.querySelector('[class*="bg-primary"]')
      expect(indicator).toHaveStyle({ transform: 'translateX(-25%)' })
    })

    it('should have default max of 100', () => {
      render(<Progress value={50} />)
      const progressbar = screen.getByRole('progressbar')
      expect(progressbar).toHaveAttribute('aria-valuemax', '100')
    })
  })

  describe('indicator position', () => {
    it('should show 0% progress when value is 0', () => {
      render(<Progress value={0} />)
      const progressbar = screen.getByRole('progressbar')
      const indicator = progressbar.querySelector('[class*="bg-primary"]')
      expect(indicator).toHaveStyle({ transform: 'translateX(-100%)' })
    })

    it('should show 50% progress when value is 50', () => {
      render(<Progress value={50} />)
      const progressbar = screen.getByRole('progressbar')
      const indicator = progressbar.querySelector('[class*="bg-primary"]')
      expect(indicator).toHaveStyle({ transform: 'translateX(-50%)' })
    })

    it('should show 100% progress when value is 100', () => {
      render(<Progress value={100} />)
      const progressbar = screen.getByRole('progressbar')
      const indicator = progressbar.querySelector('[class*="bg-primary"]')
      expect(indicator).toHaveStyle({ transform: 'translateX(-0%)' })
    })

    it('should handle undefined value as 0', () => {
      render(<Progress />)
      const progressbar = screen.getByRole('progressbar')
      const indicator = progressbar.querySelector('[class*="bg-primary"]')
      expect(indicator).toHaveStyle({ transform: 'translateX(-100%)' })
    })
  })

  describe('styling', () => {
    it('should have base styles', () => {
      render(<Progress value={50} />)
      const progressbar = screen.getByRole('progressbar')
      expect(progressbar).toHaveClass('relative')
      expect(progressbar).toHaveClass('h-4')
      expect(progressbar).toHaveClass('w-full')
      expect(progressbar).toHaveClass('overflow-hidden')
      expect(progressbar).toHaveClass('rounded-full')
    })

    it('should apply custom className', () => {
      render(<Progress value={50} className="custom-class" />)
      const progressbar = screen.getByRole('progressbar')
      expect(progressbar).toHaveClass('custom-class')
    })

    it('should have indicator with primary background', () => {
      render(<Progress value={50} />)
      const progressbar = screen.getByRole('progressbar')
      const indicator = progressbar.querySelector('[class*="bg-primary"]')
      expect(indicator).toBeInTheDocument()
    })
  })

  describe('ref forwarding', () => {
    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Progress value={50} ref={ref} />)
      expect(ref.current).not.toBeNull()
    })
  })
})
