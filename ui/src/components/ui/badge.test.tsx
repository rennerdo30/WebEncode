import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from './badge'

describe('Badge', () => {
  describe('rendering', () => {
    it('should render badge with children', () => {
      render(<Badge>Status</Badge>)
      expect(screen.getByText('Status')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(<Badge className="custom-class">Test</Badge>)
      expect(screen.getByText('Test')).toHaveClass('custom-class')
    })

    it('should have rounded-full class for pill shape', () => {
      render(<Badge>Pill</Badge>)
      expect(screen.getByText('Pill')).toHaveClass('rounded-full')
    })
  })

  describe('variants', () => {
    it('should render default variant', () => {
      render(<Badge>Default</Badge>)
      const badge = screen.getByText('Default')
      expect(badge).toHaveClass('bg-primary')
      expect(badge).toHaveClass('text-primary-foreground')
    })

    it('should render secondary variant', () => {
      render(<Badge variant="secondary">Secondary</Badge>)
      const badge = screen.getByText('Secondary')
      expect(badge).toHaveClass('bg-secondary')
      expect(badge).toHaveClass('text-secondary-foreground')
    })

    it('should render destructive variant', () => {
      render(<Badge variant="destructive">Destructive</Badge>)
      const badge = screen.getByText('Destructive')
      expect(badge).toHaveClass('bg-destructive')
      expect(badge).toHaveClass('text-destructive-foreground')
    })

    it('should render outline variant', () => {
      render(<Badge variant="outline">Outline</Badge>)
      const badge = screen.getByText('Outline')
      expect(badge).toHaveClass('text-foreground')
      expect(badge).not.toHaveClass('bg-primary')
    })
  })

  describe('styling', () => {
    it('should have correct base styles', () => {
      render(<Badge>Test</Badge>)
      const badge = screen.getByText('Test')
      expect(badge).toHaveClass('inline-flex')
      expect(badge).toHaveClass('items-center')
      expect(badge).toHaveClass('text-xs')
      expect(badge).toHaveClass('font-semibold')
    })

    it('should have padding styles', () => {
      render(<Badge>Padded</Badge>)
      const badge = screen.getByText('Padded')
      expect(badge).toHaveClass('px-2.5')
      expect(badge).toHaveClass('py-0.5')
    })

    it('should have focus styles', () => {
      render(<Badge>Focus</Badge>)
      const badge = screen.getByText('Focus')
      expect(badge).toHaveClass('focus:outline-none')
      expect(badge).toHaveClass('focus:ring-2')
    })
  })

  describe('HTML attributes', () => {
    it('should pass through HTML attributes', () => {
      render(<Badge data-testid="test-badge" role="status">Status</Badge>)
      const badge = screen.getByTestId('test-badge')
      expect(badge).toHaveAttribute('role', 'status')
    })

    it('should support onClick handler', () => {
      const handleClick = vi.fn()
      render(<Badge onClick={handleClick}>Clickable</Badge>)
      screen.getByText('Clickable').click()
      expect(handleClick).toHaveBeenCalledTimes(1)
    })
  })
})
