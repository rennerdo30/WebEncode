import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ErrorBoundary } from './error-boundary'

// Mock the reportError function
vi.mock('@/lib/api', () => ({
  reportError: vi.fn(),
}))

// Component that throws an error
const ThrowError = ({ shouldThrow = true }: { shouldThrow?: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test error message')
  }
  return <div>No error</div>
}

describe('ErrorBoundary', () => {
  // Suppress console.error for error boundary tests
  const originalError = console.error
  beforeEach(() => {
    console.error = vi.fn()
  })
  afterEach(() => {
    console.error = originalError
  })

  describe('normal rendering', () => {
    it('should render children when no error occurs', () => {
      render(
        <ErrorBoundary>
          <div>Child content</div>
        </ErrorBoundary>
      )
      expect(screen.getByText('Child content')).toBeInTheDocument()
    })

    it('should render multiple children', () => {
      render(
        <ErrorBoundary>
          <div>First child</div>
          <div>Second child</div>
        </ErrorBoundary>
      )
      expect(screen.getByText('First child')).toBeInTheDocument()
      expect(screen.getByText('Second child')).toBeInTheDocument()
    })
  })

  describe('error handling', () => {
    it('should catch errors and display error UI', () => {
      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )
      expect(screen.getByText('Something went wrong in this component')).toBeInTheDocument()
    })

    it('should display Try Again button', () => {
      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )
      expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
    })

    it('should report error to API with default source', async () => {
      const { reportError } = await import('@/lib/api')

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      expect(reportError).toHaveBeenCalledWith(
        'Test error message',
        'frontend:error-boundary',
        expect.any(String),
        expect.objectContaining({
          componentStack: expect.any(String),
        })
      )
    })

    it('should report error with custom source', async () => {
      const { reportError } = await import('@/lib/api')
      vi.mocked(reportError).mockClear()

      render(
        <ErrorBoundary source="custom-component">
          <ThrowError />
        </ErrorBoundary>
      )

      expect(reportError).toHaveBeenCalledWith(
        'Test error message',
        'custom-component',
        expect.any(String),
        expect.any(Object)
      )
    })
  })

  describe('custom fallback', () => {
    it('should render custom fallback when provided', () => {
      render(
        <ErrorBoundary fallback={<div>Custom error message</div>}>
          <ThrowError />
        </ErrorBoundary>
      )
      expect(screen.getByText('Custom error message')).toBeInTheDocument()
      expect(screen.queryByText('Something went wrong in this component')).not.toBeInTheDocument()
    })

    it('should not show Try Again button when using custom fallback', () => {
      render(
        <ErrorBoundary fallback={<div>Custom fallback</div>}>
          <ThrowError />
        </ErrorBoundary>
      )
      expect(screen.queryByRole('button', { name: /try again/i })).not.toBeInTheDocument()
    })
  })

  describe('reset functionality', () => {
    it('should reset error state when Try Again is clicked', () => {
      let shouldThrow = true

      const TestComponent = () => {
        if (shouldThrow) {
          throw new Error('Test error')
        }
        return <div>Recovered content</div>
      }

      const { rerender } = render(
        <ErrorBoundary>
          <TestComponent />
        </ErrorBoundary>
      )

      expect(screen.getByText('Something went wrong in this component')).toBeInTheDocument()

      // Fix the component before clicking reset
      shouldThrow = false

      // Click the reset button
      fireEvent.click(screen.getByRole('button', { name: /try again/i }))

      // Re-render with fixed component
      rerender(
        <ErrorBoundary>
          <TestComponent />
        </ErrorBoundary>
      )

      // The error UI should no longer be displayed
      expect(screen.queryByText('Something went wrong in this component')).not.toBeInTheDocument()
    })
  })

  describe('development mode', () => {
    it('should not show error details in production', () => {
      const originalNodeEnv = process.env.NODE_ENV

      // Mock production environment
      vi.stubEnv('NODE_ENV', 'production')

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      // Error details should not be visible in production
      expect(screen.queryByText(/Test error message/)).not.toBeInTheDocument()

      // Restore environment
      vi.stubEnv('NODE_ENV', originalNodeEnv || 'test')
    })
  })

  describe('styling', () => {
    it('should have error styling classes', () => {
      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )

      const errorContainer = screen.getByText('Something went wrong in this component').closest('div')?.parentElement
      expect(errorContainer).toHaveClass('p-6')
      expect(errorContainer).toHaveClass('border')
      expect(errorContainer).toHaveClass('rounded-lg')
    })
  })
})
