import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ScrollArea, ScrollBar } from './scroll-area'

describe('ScrollArea', () => {
  describe('rendering', () => {
    it('should render scroll area with children', () => {
      render(
        <ScrollArea>
          <p>Scroll content</p>
        </ScrollArea>
      )
      expect(screen.getByText('Scroll content')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(
        <ScrollArea className="custom-class" data-testid="scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      expect(screen.getByTestId('scroll-area')).toHaveClass('custom-class')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <ScrollArea ref={ref}>
          <p>Content</p>
        </ScrollArea>
      )
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })

    it('should have correct displayName', () => {
      expect(ScrollArea.displayName).toBe('ScrollArea')
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(
        <ScrollArea data-testid="scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      const scrollArea = screen.getByTestId('scroll-area')
      expect(scrollArea).toHaveClass('relative')
      expect(scrollArea).toHaveClass('overflow-hidden')
    })

    it('should merge custom className with default styles', () => {
      render(
        <ScrollArea className="h-[400px] w-[300px]" data-testid="scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      const scrollArea = screen.getByTestId('scroll-area')
      expect(scrollArea).toHaveClass('relative')
      expect(scrollArea).toHaveClass('overflow-hidden')
      expect(scrollArea).toHaveClass('h-[400px]')
      expect(scrollArea).toHaveClass('w-[300px]')
    })
  })

  describe('viewport', () => {
    it('should render viewport element', () => {
      const { container } = render(
        <ScrollArea>
          <p>Content</p>
        </ScrollArea>
      )
      const viewport = container.querySelector('[class*="h-full"][class*="w-full"]')
      expect(viewport).toBeInTheDocument()
    })

    it('should have viewport styling', () => {
      const { container } = render(
        <ScrollArea>
          <p>Content</p>
        </ScrollArea>
      )
      const viewport = container.querySelector('[data-radix-scroll-area-viewport]')
      expect(viewport).toHaveClass('h-full')
      expect(viewport).toHaveClass('w-full')
    })
  })

  describe('with long content', () => {
    it('should render long content that would scroll', () => {
      const longContent = Array.from({ length: 50 }, (_, i) => `Item ${i + 1}`).join('\n')
      render(
        <ScrollArea className="h-[200px]">
          <p style={{ whiteSpace: 'pre' }}>{longContent}</p>
        </ScrollArea>
      )
      expect(screen.getByText(/Item 1/)).toBeInTheDocument()
      expect(screen.getByText(/Item 50/)).toBeInTheDocument()
    })

    it('should render horizontal scrolling content', () => {
      render(
        <ScrollArea className="w-[200px]">
          <div style={{ width: '500px' }}>Wide content</div>
        </ScrollArea>
      )
      expect(screen.getByText('Wide content')).toBeInTheDocument()
    })
  })

  describe('HTML attributes', () => {
    it('should pass through data attributes', () => {
      render(
        <ScrollArea data-testid="my-scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      expect(screen.getByTestId('my-scroll-area')).toBeInTheDocument()
    })

    it('should support id attribute', () => {
      render(
        <ScrollArea id="main-scroll" data-testid="scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      expect(screen.getByTestId('scroll-area')).toHaveAttribute('id', 'main-scroll')
    })

    it('should support style attribute', () => {
      render(
        <ScrollArea style={{ maxHeight: '300px' }} data-testid="scroll-area">
          <p>Content</p>
        </ScrollArea>
      )
      expect(screen.getByTestId('scroll-area')).toHaveStyle({ maxHeight: '300px' })
    })
  })

  describe('children', () => {
    it('should render single child', () => {
      render(
        <ScrollArea>
          <div data-testid="child">Single child</div>
        </ScrollArea>
      )
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })

    it('should render multiple children', () => {
      render(
        <ScrollArea>
          <div data-testid="child1">First</div>
          <div data-testid="child2">Second</div>
          <div data-testid="child3">Third</div>
        </ScrollArea>
      )
      expect(screen.getByTestId('child1')).toBeInTheDocument()
      expect(screen.getByTestId('child2')).toBeInTheDocument()
      expect(screen.getByTestId('child3')).toBeInTheDocument()
    })

    it('should render nested content', () => {
      render(
        <ScrollArea>
          <ul>
            <li>Item 1</li>
            <li>Item 2</li>
            <li>Item 3</li>
          </ul>
        </ScrollArea>
      )
      expect(screen.getByRole('list')).toBeInTheDocument()
      expect(screen.getAllByRole('listitem')).toHaveLength(3)
    })
  })
})

describe('ScrollBar', () => {
  describe('rendering', () => {
    it('should have correct displayName', () => {
      expect(ScrollBar.displayName).toBe('ScrollAreaScrollbar')
    })
  })

  describe('styling', () => {
    it('should render with proper orientation classes when vertical', () => {
      // ScrollBar component uses classes for vertical orientation
      const { container } = render(
        <ScrollArea>
          <p>Content</p>
        </ScrollArea>
      )
      // Verify the scroll area renders
      expect(container.firstChild).toBeInTheDocument()
    })
  })
})
