import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import {
  Select,
  SelectTrigger,
  SelectContent,
  SelectItem,
  SelectValue,
  SelectGroup,
  SelectLabel,
  SelectSeparator,
  SelectScrollUpButton,
  SelectScrollDownButton,
} from './select'

describe('Select', () => {
  describe('Select Root', () => {
    it('should render select trigger', () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByRole('combobox')).toBeInTheDocument()
    })

    it('should support controlled mode', async () => {
      const handleValueChange = vi.fn()
      const { rerender } = render(
        <Select value="option1" onValueChange={handleValueChange}>
          <SelectTrigger aria-label="Select option">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByText('Option 1')).toBeInTheDocument()

      rerender(
        <Select value="option2" onValueChange={handleValueChange}>
          <SelectTrigger aria-label="Select option">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByText('Option 2')).toBeInTheDocument()
    })

    it('should call onValueChange when selection changes', async () => {
      const handleValueChange = vi.fn()
      render(
        <Select onValueChange={handleValueChange}>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        fireEvent.click(screen.getByText('Option 1'))
      })
      expect(handleValueChange).toHaveBeenCalledWith('option1')
    })

    it('should support defaultValue', () => {
      render(
        <Select defaultValue="option2">
          <SelectTrigger aria-label="Select option">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByText('Option 2')).toBeInTheDocument()
    })

    it('should support disabled state', () => {
      render(
        <Select disabled>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByRole('combobox')).toBeDisabled()
    })
  })

  describe('SelectTrigger', () => {
    it('should render trigger with placeholder', () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Choose an option" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByText('Choose an option')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Select>
          <SelectTrigger className="custom-trigger" aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByRole('combobox')).toHaveClass('custom-trigger')
    })

    it('should have styling classes', () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveClass('flex')
      expect(trigger).toHaveClass('h-10')
      expect(trigger).toHaveClass('w-full')
      expect(trigger).toHaveClass('rounded-md')
      expect(trigger).toHaveClass('border')
      expect(trigger).toHaveClass('border-input')
    })

    it('should have focus ring classes', () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveClass('focus:ring-2')
      expect(trigger).toHaveClass('focus:ring-ring')
      expect(trigger).toHaveClass('focus:ring-offset-2')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Select>
          <SelectTrigger ref={ref} aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(ref.current).toBeInstanceOf(HTMLButtonElement)
    })

    it('should render chevron icon', () => {
      const { container } = render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(container.querySelector('svg')).toBeInTheDocument()
    })

    it('should have disabled styling when disabled', () => {
      render(
        <Select disabled>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )
      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveClass('disabled:cursor-not-allowed')
      expect(trigger).toHaveClass('disabled:opacity-50')
    })
  })

  describe('SelectContent', () => {
    it('should open content when trigger is clicked', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent className="custom-content">
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('listbox')).toHaveClass('custom-content')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const content = screen.getByRole('listbox')
        expect(content).toHaveClass('z-50')
        expect(content).toHaveClass('rounded-md')
        expect(content).toHaveClass('border')
        expect(content).toHaveClass('bg-popover')
      })
    })

    it('should close on escape key', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })

      fireEvent.keyDown(document, { key: 'Escape' })
      await waitFor(() => {
        expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
      })
    })
  })

  describe('SelectItem', () => {
    it('should render item', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Option 1' })).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1" className="custom-item">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('option')).toHaveClass('custom-item')
      })
    })

    it('should show check mark when selected', async () => {
      render(
        <Select defaultValue="option1">
          <SelectTrigger aria-label="Select option">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const selectedOption = screen.getByRole('option', { name: 'Option 1' })
        expect(selectedOption).toHaveAttribute('data-state', 'checked')
      })
    })

    it('should be disabled when disabled prop is true', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1" disabled>Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const disabledOption = screen.getByRole('option', { name: 'Option 1' })
        expect(disabledOption).toHaveAttribute('data-disabled')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const item = screen.getByRole('option')
        expect(item).toHaveClass('relative')
        expect(item).toHaveClass('flex')
        expect(item).toHaveClass('cursor-default')
        expect(item).toHaveClass('rounded-sm')
        expect(item).toHaveClass('py-1.5')
        expect(item).toHaveClass('text-sm')
      })
    })

    it('should forward ref correctly', async () => {
      const ref = { current: null }
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1" ref={ref}>Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLDivElement)
      })
    })
  })

  describe('SelectGroup and SelectLabel', () => {
    it('should render group with label', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select fruit">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectLabel>Fruits</SelectLabel>
              <SelectItem value="apple">Apple</SelectItem>
              <SelectItem value="banana">Banana</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByText('Fruits')).toBeInTheDocument()
        expect(screen.getByRole('option', { name: 'Apple' })).toBeInTheDocument()
        expect(screen.getByRole('option', { name: 'Banana' })).toBeInTheDocument()
      })
    })

    it('should apply custom className to SelectLabel', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectLabel className="custom-label">Category</SelectLabel>
              <SelectItem value="item1">Item 1</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByText('Category')).toHaveClass('custom-label')
      })
    })

    it('should have styling classes on SelectLabel', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectLabel>Label</SelectLabel>
              <SelectItem value="item1">Item 1</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const label = screen.getByText('Label')
        expect(label).toHaveClass('py-1.5')
        expect(label).toHaveClass('pl-8')
        expect(label).toHaveClass('text-sm')
        expect(label).toHaveClass('font-semibold')
      })
    })
  })

  describe('SelectSeparator', () => {
    it('should render separator', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="item1">Item 1</SelectItem>
            <SelectSeparator data-testid="separator" />
            <SelectItem value="item2">Item 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByTestId('separator')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="item1">Item 1</SelectItem>
            <SelectSeparator className="custom-separator" data-testid="separator" />
            <SelectItem value="item2">Item 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByTestId('separator')).toHaveClass('custom-separator')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="item1">Item 1</SelectItem>
            <SelectSeparator data-testid="separator" />
            <SelectItem value="item2">Item 2</SelectItem>
          </SelectContent>
        </Select>
      )

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        const separator = screen.getByTestId('separator')
        expect(separator).toHaveClass('h-px')
        expect(separator).toHaveClass('bg-muted')
        expect(separator).toHaveClass('my-1')
      })
    })
  })

  describe('accessibility', () => {
    it('should have correct ARIA roles', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
          </SelectContent>
        </Select>
      )

      expect(screen.getByRole('combobox')).toBeInTheDocument()

      fireEvent.click(screen.getByRole('combobox'))
      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
        expect(screen.getByRole('option')).toBeInTheDocument()
      })
    })

    it('should support aria-label on trigger', () => {
      render(
        <Select>
          <SelectTrigger aria-label="Choose a color">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="red">Red</SelectItem>
          </SelectContent>
        </Select>
      )
      expect(screen.getByLabelText('Choose a color')).toBeInTheDocument()
    })

    it('should be keyboard navigable', async () => {
      render(
        <Select>
          <SelectTrigger aria-label="Select option">
            <SelectValue placeholder="Select..." />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="option1">Option 1</SelectItem>
            <SelectItem value="option2">Option 2</SelectItem>
            <SelectItem value="option3">Option 3</SelectItem>
          </SelectContent>
        </Select>
      )

      const trigger = screen.getByRole('combobox')
      fireEvent.keyDown(trigger, { key: 'ArrowDown' })

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })
    })
  })

  describe('displayNames', () => {
    it('should have correct displayNames', () => {
      expect(SelectTrigger.displayName).toBe('SelectTrigger')
      expect(SelectContent.displayName).toBe('SelectContent')
      expect(SelectItem.displayName).toBe('SelectItem')
      expect(SelectLabel.displayName).toBe('SelectLabel')
      expect(SelectSeparator.displayName).toBe('SelectSeparator')
      expect(SelectScrollUpButton.displayName).toBe('SelectScrollUpButton')
      expect(SelectScrollDownButton.displayName).toBe('SelectScrollDownButton')
    })
  })
})
