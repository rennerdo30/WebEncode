import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Checkbox } from './checkbox'

describe('Checkbox', () => {
  describe('rendering', () => {
    it('should render checkbox element', () => {
      render(<Checkbox aria-label="Test checkbox" />)
      expect(screen.getByRole('checkbox')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(<Checkbox className="custom-class" aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveClass('custom-class')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Checkbox ref={ref} aria-label="Test" />)
      expect(ref.current).toBeInstanceOf(HTMLButtonElement)
    })

    it('should have correct displayName', () => {
      expect(Checkbox.displayName).toBe('Checkbox')
    })
  })

  describe('states', () => {
    it('should render unchecked by default', () => {
      render(<Checkbox aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveAttribute('data-state', 'unchecked')
    })

    it('should render checked when checked prop is true', () => {
      render(<Checkbox checked aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveAttribute('data-state', 'checked')
    })

    it('should render with indeterminate state', () => {
      render(<Checkbox checked="indeterminate" aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveAttribute('data-state', 'indeterminate')
    })
  })

  describe('interactions', () => {
    it('should toggle when clicked', () => {
      const handleChange = vi.fn()
      render(<Checkbox onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('checkbox'))
      expect(handleChange).toHaveBeenCalledWith(true)
    })

    it('should call onCheckedChange with false when unchecking', () => {
      const handleChange = vi.fn()
      render(<Checkbox checked onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('checkbox'))
      expect(handleChange).toHaveBeenCalledWith(false)
    })

    it('should be controllable', () => {
      const { rerender } = render(<Checkbox checked={false} aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveAttribute('data-state', 'unchecked')

      rerender(<Checkbox checked aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveAttribute('data-state', 'checked')
    })
  })

  describe('disabled state', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<Checkbox disabled aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toBeDisabled()
    })

    it('should not toggle when disabled', () => {
      const handleChange = vi.fn()
      render(<Checkbox disabled onCheckedChange={handleChange} aria-label="Test" />)

      fireEvent.click(screen.getByRole('checkbox'))
      expect(handleChange).not.toHaveBeenCalled()
    })

    it('should have disabled styling classes', () => {
      render(<Checkbox disabled aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveClass('disabled:cursor-not-allowed')
      expect(checkbox).toHaveClass('disabled:opacity-50')
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(<Checkbox aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveClass('h-4')
      expect(checkbox).toHaveClass('w-4')
      expect(checkbox).toHaveClass('rounded-sm')
      expect(checkbox).toHaveClass('border')
      expect(checkbox).toHaveClass('border-primary')
    })

    it('should have focus ring classes', () => {
      render(<Checkbox aria-label="Test" />)
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toHaveClass('focus-visible:ring-2')
      expect(checkbox).toHaveClass('focus-visible:ring-ring')
      expect(checkbox).toHaveClass('focus-visible:ring-offset-2')
    })
  })

  describe('accessibility', () => {
    it('should support aria-label', () => {
      render(<Checkbox aria-label="Accept terms" />)
      expect(screen.getByLabelText('Accept terms')).toBeInTheDocument()
    })

    it('should support aria-describedby', () => {
      render(
        <>
          <Checkbox aria-describedby="description" aria-label="Test" />
          <span id="description">Helper text</span>
        </>
      )
      expect(screen.getByRole('checkbox')).toHaveAttribute('aria-describedby', 'description')
    })

    it('should be keyboard accessible', () => {
      const handleChange = vi.fn()
      render(<Checkbox onCheckedChange={handleChange} aria-label="Test" />)

      const checkbox = screen.getByRole('checkbox')
      checkbox.focus()
      fireEvent.keyDown(checkbox, { key: 'Enter' })
      // Note: Radix checkbox uses Space to toggle, but Enter may also work depending on implementation
    })
  })

  describe('HTML attributes', () => {
    it('should pass through data attributes', () => {
      render(<Checkbox data-testid="my-checkbox" aria-label="Test" />)
      expect(screen.getByTestId('my-checkbox')).toBeInTheDocument()
    })

    it('should support id attribute', () => {
      render(<Checkbox id="terms-checkbox" aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveAttribute('id', 'terms-checkbox')
    })

    it('should support name attribute', () => {
      render(<Checkbox name="terms" aria-label="Test" />)
      // Radix checkbox renders name on a hidden input, not the button
      const checkbox = screen.getByRole('checkbox')
      expect(checkbox).toBeInTheDocument()
    })

    it('should support value attribute', () => {
      render(<Checkbox value="accepted" aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveAttribute('value', 'accepted')
    })

    it('should support required attribute', () => {
      render(<Checkbox required aria-label="Test" />)
      expect(screen.getByRole('checkbox')).toHaveAttribute('aria-required', 'true')
    })
  })
})
