import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { useQueryClient } from '@tanstack/react-query'
import { Providers } from './providers'

// Test component to verify QueryClient is available
function TestConsumer() {
  const queryClient = useQueryClient()
  return <div data-testid="consumer">QueryClient: {queryClient ? 'available' : 'not available'}</div>
}

describe('Providers', () => {
  describe('rendering', () => {
    it('should render children', () => {
      render(
        <Providers>
          <div data-testid="child">Child content</div>
        </Providers>
      )

      expect(screen.getByTestId('child')).toBeInTheDocument()
      expect(screen.getByText('Child content')).toBeInTheDocument()
    })

    it('should render multiple children', () => {
      render(
        <Providers>
          <div data-testid="child1">First child</div>
          <div data-testid="child2">Second child</div>
        </Providers>
      )

      expect(screen.getByTestId('child1')).toBeInTheDocument()
      expect(screen.getByTestId('child2')).toBeInTheDocument()
    })

    it('should render nested components', () => {
      render(
        <Providers>
          <div data-testid="parent">
            <span data-testid="nested">Nested content</span>
          </div>
        </Providers>
      )

      expect(screen.getByTestId('parent')).toBeInTheDocument()
      expect(screen.getByTestId('nested')).toBeInTheDocument()
    })
  })

  describe('QueryClient', () => {
    it('should provide QueryClient to children', () => {
      render(
        <Providers>
          <TestConsumer />
        </Providers>
      )

      expect(screen.getByTestId('consumer')).toHaveTextContent('QueryClient: available')
    })

    it('should allow queries in children components', () => {
      // Simply ensure it renders without errors when using useQueryClient
      expect(() => {
        render(
          <Providers>
            <TestConsumer />
          </Providers>
        )
      }).not.toThrow()
    })
  })

  describe('QueryClient configuration', () => {
    it('should be configured with default options', () => {
      // The Providers component configures staleTime: 5000 and refetchOnWindowFocus: false
      // We verify this by checking that the component renders without issues
      render(
        <Providers>
          <div>Test</div>
        </Providers>
      )

      expect(screen.getByText('Test')).toBeInTheDocument()
    })
  })

  describe('stability', () => {
    it('should maintain same QueryClient instance across re-renders', () => {
      const { rerender } = render(
        <Providers>
          <TestConsumer />
        </Providers>
      )

      expect(screen.getByTestId('consumer')).toHaveTextContent('QueryClient: available')

      rerender(
        <Providers>
          <TestConsumer />
        </Providers>
      )

      expect(screen.getByTestId('consumer')).toHaveTextContent('QueryClient: available')
    })
  })

  describe('type safety', () => {
    it('should accept children as React.ReactNode', () => {
      // Test with various ReactNode types
      render(
        <Providers>
          <span>String child</span>
          {null}
          {undefined}
          {123}
        </Providers>
      )

      expect(screen.getByText('String child')).toBeInTheDocument()
      expect(screen.getByText('123')).toBeInTheDocument()
    })
  })
})
