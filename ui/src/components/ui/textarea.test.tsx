import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Textarea } from './textarea'

describe('Textarea', () => {
  describe('rendering', () => {
    it('should render textarea element', () => {
      render(<Textarea aria-label="Test textarea" />)
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(<Textarea className="custom-class" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveClass('custom-class')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Textarea ref={ref} aria-label="Test" />)
      expect(ref.current).toBeInstanceOf(HTMLTextAreaElement)
    })

    it('should have correct displayName', () => {
      expect(Textarea.displayName).toBe('Textarea')
    })
  })

  describe('value handling', () => {
    it('should render with initial value', () => {
      render(<Textarea defaultValue="Initial text" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveValue('Initial text')
    })

    it('should be controllable', () => {
      const { rerender } = render(<Textarea value="First" onChange={() => {}} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveValue('First')

      rerender(<Textarea value="Second" onChange={() => {}} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveValue('Second')
    })

    it('should call onChange when value changes', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<Textarea onChange={handleChange} aria-label="Test" />)

      await user.type(screen.getByRole('textbox'), 'Hello')
      expect(handleChange).toHaveBeenCalled()
    })

    it('should support uncontrolled mode', async () => {
      const user = userEvent.setup()
      render(<Textarea aria-label="Test" />)

      const textarea = screen.getByRole('textbox')
      await user.type(textarea, 'Hello World')
      expect(textarea).toHaveValue('Hello World')
    })
  })

  describe('placeholder', () => {
    it('should display placeholder text', () => {
      render(<Textarea placeholder="Enter your message..." aria-label="Test" />)
      expect(screen.getByPlaceholderText('Enter your message...')).toBeInTheDocument()
    })

    it('should have placeholder styling', () => {
      render(<Textarea placeholder="Test" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveClass('placeholder:text-muted-foreground')
    })
  })

  describe('disabled state', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<Textarea disabled aria-label="Test" />)
      expect(screen.getByRole('textbox')).toBeDisabled()
    })

    it('should not allow typing when disabled', async () => {
      const user = userEvent.setup()
      render(<Textarea disabled defaultValue="Original" aria-label="Test" />)

      const textarea = screen.getByRole('textbox')
      await user.type(textarea, 'New text')
      expect(textarea).toHaveValue('Original')
    })

    it('should have disabled styling classes', () => {
      render(<Textarea disabled aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('disabled:cursor-not-allowed')
      expect(textarea).toHaveClass('disabled:opacity-50')
    })
  })

  describe('read-only state', () => {
    it('should be read-only when readOnly prop is true', () => {
      render(<Textarea readOnly aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('readonly')
    })

    it('should not allow editing when read-only', async () => {
      const user = userEvent.setup()
      render(<Textarea readOnly defaultValue="Read only text" aria-label="Test" />)

      const textarea = screen.getByRole('textbox')
      await user.type(textarea, 'New text')
      expect(textarea).toHaveValue('Read only text')
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(<Textarea aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('min-h-[80px]')
      expect(textarea).toHaveClass('w-full')
      expect(textarea).toHaveClass('rounded-md')
      expect(textarea).toHaveClass('border')
      expect(textarea).toHaveClass('border-input')
    })

    it('should have focus ring classes', () => {
      render(<Textarea aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('focus-visible:ring-2')
      expect(textarea).toHaveClass('focus-visible:ring-ring')
      expect(textarea).toHaveClass('focus-visible:ring-offset-2')
    })

    it('should have padding classes', () => {
      render(<Textarea aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('px-3')
      expect(textarea).toHaveClass('py-2')
    })

    it('should have responsive text size classes', () => {
      render(<Textarea aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('text-base')
      expect(textarea).toHaveClass('md:text-sm')
    })

    it('should have background styling', () => {
      render(<Textarea aria-label="Test" />)
      const textarea = screen.getByRole('textbox')
      expect(textarea).toHaveClass('bg-background')
    })
  })

  describe('interactions', () => {
    it('should handle focus event', () => {
      const handleFocus = vi.fn()
      render(<Textarea onFocus={handleFocus} aria-label="Test" />)

      fireEvent.focus(screen.getByRole('textbox'))
      expect(handleFocus).toHaveBeenCalled()
    })

    it('should handle blur event', () => {
      const handleBlur = vi.fn()
      render(<Textarea onBlur={handleBlur} aria-label="Test" />)

      const textarea = screen.getByRole('textbox')
      fireEvent.focus(textarea)
      fireEvent.blur(textarea)
      expect(handleBlur).toHaveBeenCalled()
    })

    it('should handle keyDown event', () => {
      const handleKeyDown = vi.fn()
      render(<Textarea onKeyDown={handleKeyDown} aria-label="Test" />)

      fireEvent.keyDown(screen.getByRole('textbox'), { key: 'Enter' })
      expect(handleKeyDown).toHaveBeenCalled()
    })

    it('should handle paste event', () => {
      const handlePaste = vi.fn()
      render(<Textarea onPaste={handlePaste} aria-label="Test" />)

      fireEvent.paste(screen.getByRole('textbox'), {
        clipboardData: { getData: () => 'Pasted text' }
      })
      expect(handlePaste).toHaveBeenCalled()
    })
  })

  describe('accessibility', () => {
    it('should support aria-label', () => {
      render(<Textarea aria-label="Message content" />)
      expect(screen.getByLabelText('Message content')).toBeInTheDocument()
    })

    it('should support aria-describedby', () => {
      render(
        <>
          <Textarea aria-describedby="description" aria-label="Test" />
          <span id="description">Max 500 characters</span>
        </>
      )
      expect(screen.getByRole('textbox')).toHaveAttribute('aria-describedby', 'description')
    })

    it('should support aria-invalid', () => {
      render(<Textarea aria-invalid="true" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('aria-invalid', 'true')
    })

    it('should support aria-required', () => {
      render(<Textarea aria-required="true" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('aria-required', 'true')
    })
  })

  describe('HTML attributes', () => {
    it('should pass through data attributes', () => {
      render(<Textarea data-testid="my-textarea" aria-label="Test" />)
      expect(screen.getByTestId('my-textarea')).toBeInTheDocument()
    })

    it('should support id attribute', () => {
      render(<Textarea id="message-textarea" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('id', 'message-textarea')
    })

    it('should support name attribute', () => {
      render(<Textarea name="message" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('name', 'message')
    })

    it('should support rows attribute', () => {
      render(<Textarea rows={5} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('rows', '5')
    })

    it('should support cols attribute', () => {
      render(<Textarea cols={50} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('cols', '50')
    })

    it('should support maxLength attribute', () => {
      render(<Textarea maxLength={500} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('maxlength', '500')
    })

    it('should support minLength attribute', () => {
      render(<Textarea minLength={10} aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('minlength', '10')
    })

    it('should support required attribute', () => {
      render(<Textarea required aria-label="Test" />)
      expect(screen.getByRole('textbox')).toBeRequired()
    })

    it('should support autoComplete attribute', () => {
      render(<Textarea autoComplete="off" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('autocomplete', 'off')
    })

    it('should support autoFocus attribute', () => {
      render(<Textarea autoFocus aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveFocus()
    })

    it('should support wrap attribute', () => {
      render(<Textarea wrap="hard" aria-label="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('wrap', 'hard')
    })
  })
})
