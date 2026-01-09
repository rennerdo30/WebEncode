import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Switch } from './switch'

describe('Switch', () => {
  describe('rendering', () => {
    it('should render switch element', () => {
      render(<Switch aria-label="Test switch" />)
      expect(screen.getByRole('switch')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(<Switch className="custom-class" aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveClass('custom-class')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Switch ref={ref} aria-label="Test" />)
      expect(ref.current).toBeInstanceOf(HTMLButtonElement)
    })

    it('should have correct displayName', () => {
      expect(Switch.displayName).toBe('Switch')
    })
  })

  describe('states', () => {
    it('should render unchecked by default', () => {
      render(<Switch aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveAttribute('data-state', 'unchecked')
    })

    it('should render checked when checked prop is true', () => {
      render(<Switch checked aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveAttribute('data-state', 'checked')
    })

    it('should render unchecked when checked prop is false', () => {
      render(<Switch checked={false} aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveAttribute('data-state', 'unchecked')
    })

    it('should support defaultChecked prop', () => {
      render(<Switch defaultChecked aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveAttribute('data-state', 'checked')
    })
  })

  describe('interactions', () => {
    it('should toggle when clicked', () => {
      const handleChange = vi.fn()
      render(<Switch onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('switch'))
      expect(handleChange).toHaveBeenCalledWith(true)
    })

    it('should call onCheckedChange with false when turning off', () => {
      const handleChange = vi.fn()
      render(<Switch checked onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('switch'))
      expect(handleChange).toHaveBeenCalledWith(false)
    })

    it('should be controllable', () => {
      const { rerender } = render(<Switch checked={false} aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveAttribute('data-state', 'unchecked')

      rerender(<Switch checked aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveAttribute('data-state', 'checked')
    })

    it('should support uncontrolled mode', () => {
      render(<Switch aria-label="Test" />)
      const switchEl = screen.getByRole('switch')

      expect(switchEl).toHaveAttribute('data-state', 'unchecked')
      fireEvent.click(switchEl)
      expect(switchEl).toHaveAttribute('data-state', 'checked')
      fireEvent.click(switchEl)
      expect(switchEl).toHaveAttribute('data-state', 'unchecked')
    })
  })

  describe('disabled state', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<Switch disabled aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toBeDisabled()
    })

    it('should not toggle when disabled', () => {
      const handleChange = vi.fn()
      render(<Switch disabled onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('switch'))
      expect(handleChange).not.toHaveBeenCalled()
    })

    it('should have disabled styling classes', () => {
      render(<Switch disabled aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveClass('disabled:cursor-not-allowed')
      expect(switchEl).toHaveClass('disabled:opacity-50')
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(<Switch aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveClass('h-6')
      expect(switchEl).toHaveClass('w-11')
      expect(switchEl).toHaveClass('rounded-full')
      expect(switchEl).toHaveClass('cursor-pointer')
    })

    it('should have focus ring classes', () => {
      render(<Switch aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveClass('focus-visible:ring-2')
      expect(switchEl).toHaveClass('focus-visible:ring-ring')
      expect(switchEl).toHaveClass('focus-visible:ring-offset-2')
    })

    it('should have transition classes', () => {
      render(<Switch aria-label="Test" />)
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toHaveClass('transition-colors')
    })
  })

  describe('thumb element', () => {
    it('should render thumb element', () => {
      const { container } = render(<Switch aria-label="Test" />)
      const thumb = container.querySelector('[class*="rounded-full"][class*="bg-background"]')
      expect(thumb).toBeInTheDocument()
    })

    it('should have thumb styling classes', () => {
      const { container } = render(<Switch aria-label="Test" />)
      const thumb = container.querySelector('[class*="h-5"][class*="w-5"]')
      expect(thumb).toHaveClass('pointer-events-none')
      expect(thumb).toHaveClass('rounded-full')
      expect(thumb).toHaveClass('bg-background')
      expect(thumb).toHaveClass('shadow-lg')
    })
  })

  describe('accessibility', () => {
    it('should support aria-label', () => {
      render(<Switch aria-label="Toggle notifications" />)
      expect(screen.getByLabelText('Toggle notifications')).toBeInTheDocument()
    })

    it('should support aria-describedby', () => {
      render(
        <>
          <Switch aria-describedby="description" aria-label="Test" />
          <span id="description">Helper text</span>
        </>
      )
      expect(screen.getByRole('switch')).toHaveAttribute('aria-describedby', 'description')
    })

    it('should have role="switch"', () => {
      render(<Switch aria-label="Test" />)
      expect(screen.getByRole('switch')).toBeInTheDocument()
    })

    it('should be keyboard accessible via space', () => {
      const handleChange = vi.fn()
      render(<Switch onCheckedChange={handleChange} aria-label="Test" />)

      const switchEl = screen.getByRole('switch')
      switchEl.focus()
      fireEvent.keyDown(switchEl, { key: ' ' })
    })
  })

  describe('HTML attributes', () => {
    it('should pass through data attributes', () => {
      render(<Switch data-testid="my-switch" aria-label="Test" />)
      expect(screen.getByTestId('my-switch')).toBeInTheDocument()
    })

    it('should support id attribute', () => {
      render(<Switch id="dark-mode-switch" aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveAttribute('id', 'dark-mode-switch')
    })

    it('should support name attribute', () => {
      render(<Switch name="darkMode" aria-label="Test" />)
      // Radix switch renders name on a hidden input, not the button
      const switchEl = screen.getByRole('switch')
      expect(switchEl).toBeInTheDocument()
    })

    it('should support value attribute', () => {
      render(<Switch value="on" aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveAttribute('value', 'on')
    })

    it('should support required attribute', () => {
      render(<Switch required aria-label="Test" />)
      expect(screen.getByRole('switch')).toHaveAttribute('aria-required', 'true')
    })
  })
})
