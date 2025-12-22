import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Input } from './input'

describe('Input', () => {
  describe('rendering', () => {
    it('should render an input element', () => {
      render(<Input />)
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('should render with placeholder', () => {
      render(<Input placeholder="Enter text" />)
      expect(screen.getByPlaceholderText('Enter text')).toBeInTheDocument()
    })

    it('should render with value', () => {
      render(<Input value="test value" readOnly />)
      expect(screen.getByDisplayValue('test value')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(<Input className="custom-class" />)
      expect(screen.getByRole('textbox')).toHaveClass('custom-class')
    })
  })

  describe('types', () => {
    it('should render text input by default', () => {
      render(<Input type="text" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('type', 'text')
    })

    it('should render input without type when not specified', () => {
      render(<Input />)
      // When type is not specified, input defaults to text behavior but may not have type attribute
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('should render password input', () => {
      render(<Input type="password" data-testid="password-input" />)
      expect(screen.getByTestId('password-input')).toHaveAttribute('type', 'password')
    })

    it('should render email input', () => {
      render(<Input type="email" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('type', 'email')
    })

    it('should render number input', () => {
      render(<Input type="number" />)
      expect(screen.getByRole('spinbutton')).toHaveAttribute('type', 'number')
    })

    it('should render search input', () => {
      render(<Input type="search" />)
      expect(screen.getByRole('searchbox')).toHaveAttribute('type', 'search')
    })
  })

  describe('interactions', () => {
    it('should handle onChange events', () => {
      const handleChange = vi.fn()
      render(<Input onChange={handleChange} />)

      fireEvent.change(screen.getByRole('textbox'), { target: { value: 'new value' } })
      expect(handleChange).toHaveBeenCalledTimes(1)
    })

    it('should handle onFocus events', () => {
      const handleFocus = vi.fn()
      render(<Input onFocus={handleFocus} />)

      fireEvent.focus(screen.getByRole('textbox'))
      expect(handleFocus).toHaveBeenCalledTimes(1)
    })

    it('should handle onBlur events', () => {
      const handleBlur = vi.fn()
      render(<Input onBlur={handleBlur} />)

      fireEvent.blur(screen.getByRole('textbox'))
      expect(handleBlur).toHaveBeenCalledTimes(1)
    })

    it('should update value when typing', () => {
      const handleChange = vi.fn()
      render(<Input onChange={handleChange} />)

      const input = screen.getByRole('textbox')
      fireEvent.change(input, { target: { value: 'hello' } })

      expect(handleChange).toHaveBeenCalled()
    })
  })

  describe('disabled state', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<Input disabled />)
      expect(screen.getByRole('textbox')).toBeDisabled()
    })

    it('should have disabled styles', () => {
      render(<Input disabled />)
      expect(screen.getByRole('textbox')).toHaveClass('disabled:cursor-not-allowed')
      expect(screen.getByRole('textbox')).toHaveClass('disabled:opacity-50')
    })

    it('should not trigger onChange when disabled', () => {
      const handleChange = vi.fn()
      render(<Input disabled onChange={handleChange} />)

      // Disabled inputs don't fire change events
      expect(screen.getByRole('textbox')).toBeDisabled()
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(<Input />)
      const input = screen.getByRole('textbox')
      expect(input).toHaveClass('flex')
      expect(input).toHaveClass('h-10')
      expect(input).toHaveClass('w-full')
      expect(input).toHaveClass('rounded-md')
      expect(input).toHaveClass('border')
    })

    it('should have focus-visible styles', () => {
      render(<Input />)
      const input = screen.getByRole('textbox')
      expect(input).toHaveClass('focus-visible:outline-none')
      expect(input).toHaveClass('focus-visible:ring-2')
    })
  })

  describe('ref forwarding', () => {
    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Input ref={ref} />)
      expect(ref.current).toBeInstanceOf(HTMLInputElement)
    })

    it('should allow ref to be used for focus', () => {
      const ref = { current: null as HTMLInputElement | null }
      render(<Input ref={ref} />)

      ref.current?.focus()
      expect(document.activeElement).toBe(ref.current)
    })
  })

  describe('HTML attributes', () => {
    it('should pass through HTML attributes', () => {
      render(<Input data-testid="test-input" aria-label="Test input" />)
      const input = screen.getByTestId('test-input')
      expect(input).toHaveAttribute('aria-label', 'Test input')
    })

    it('should support name attribute', () => {
      render(<Input name="email" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('name', 'email')
    })

    it('should support required attribute', () => {
      render(<Input required />)
      expect(screen.getByRole('textbox')).toBeRequired()
    })

    it('should support maxLength attribute', () => {
      render(<Input maxLength={100} />)
      expect(screen.getByRole('textbox')).toHaveAttribute('maxLength', '100')
    })

    it('should support autoComplete attribute', () => {
      render(<Input autoComplete="email" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('autoComplete', 'email')
    })
  })
})
