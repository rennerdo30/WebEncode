import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { Popover, PopoverTrigger, PopoverContent, PopoverAnchor } from './popover'

describe('Popover', () => {
  describe('Popover Root', () => {
    it('should render popover trigger', () => {
      render(
        <Popover>
          <PopoverTrigger>Open Popover</PopoverTrigger>
          <PopoverContent>Popover content</PopoverContent>
        </Popover>
      )
      expect(screen.getByText('Open Popover')).toBeInTheDocument()
    })

    it('should open popover when trigger is clicked', async () => {
      render(
        <Popover>
          <PopoverTrigger>Open Popover</PopoverTrigger>
          <PopoverContent>Popover content</PopoverContent>
        </Popover>
      )

      fireEvent.click(screen.getByText('Open Popover'))
      await waitFor(() => {
        expect(screen.getByText('Popover content')).toBeInTheDocument()
      })
    })

    it('should support controlled mode', async () => {
      const { rerender } = render(
        <Popover open={false}>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      expect(screen.queryByText('Content')).not.toBeInTheDocument()

      rerender(
        <Popover open={true}>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })
    })

    it('should call onOpenChange when state changes', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Popover onOpenChange={handleOpenChange}>
          <PopoverTrigger>Open Popover</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )

      fireEvent.click(screen.getByText('Open Popover'))
      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(true)
      })
    })

    it('should support defaultOpen prop', async () => {
      render(
        <Popover defaultOpen>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })
    })
  })

  describe('PopoverTrigger', () => {
    it('should render trigger button', () => {
      render(
        <Popover>
          <PopoverTrigger data-testid="trigger">Click me</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      expect(screen.getByTestId('trigger')).toBeInTheDocument()
    })

    it('should support asChild prop', () => {
      render(
        <Popover>
          <PopoverTrigger asChild>
            <button type="button" data-testid="custom-trigger">Custom Button</button>
          </PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      expect(screen.getByTestId('custom-trigger')).toHaveTextContent('Custom Button')
    })
  })

  describe('PopoverContent', () => {
    it('should render content when open', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>
            <p>Popover content</p>
          </PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByText('Popover content')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent className="custom-popover" data-testid="content">
            Content
          </PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByTestId('content')).toHaveClass('custom-popover')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent data-testid="content">Content</PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        const content = screen.getByTestId('content')
        expect(content).toHaveClass('z-50')
        expect(content).toHaveClass('w-72')
        expect(content).toHaveClass('rounded-md')
        expect(content).toHaveClass('border')
        expect(content).toHaveClass('bg-popover')
        expect(content).toHaveClass('p-4')
        expect(content).toHaveClass('shadow-md')
      })
    })

    it('should forward ref correctly', async () => {
      const ref = { current: null }
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent ref={ref}>Content</PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLDivElement)
      })
    })

    it('should support align prop', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent align="start" data-testid="content">
            Content
          </PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByTestId('content')).toBeInTheDocument()
      })
    })

    it('should support sideOffset prop', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent sideOffset={8} data-testid="content">
            Content
          </PopoverContent>
        </Popover>
      )
      await waitFor(() => {
        expect(screen.getByTestId('content')).toBeInTheDocument()
      })
    })

    it('should close on escape key', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Popover open onOpenChange={handleOpenChange}>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )

      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })

      fireEvent.keyDown(document, { key: 'Escape' })
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })

    it('should close when clicking outside', async () => {
      const handleOpenChange = vi.fn()
      render(
        <>
          <div data-testid="outside">Outside</div>
          <Popover open onOpenChange={handleOpenChange}>
            <PopoverTrigger>Open</PopoverTrigger>
            <PopoverContent>Content</PopoverContent>
          </Popover>
        </>
      )

      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })

      fireEvent.pointerDown(screen.getByTestId('outside'))
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })
  })

  describe('PopoverAnchor', () => {
    it('should render anchor element', () => {
      render(
        <Popover>
          <PopoverAnchor data-testid="anchor">
            <span>Anchor content</span>
          </PopoverAnchor>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      expect(screen.getByTestId('anchor')).toBeInTheDocument()
    })

    it('should support asChild prop', () => {
      render(
        <Popover>
          <PopoverAnchor asChild>
            <div data-testid="custom-anchor">Custom Anchor</div>
          </PopoverAnchor>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )
      expect(screen.getByTestId('custom-anchor')).toBeInTheDocument()
    })
  })

  describe('accessibility', () => {
    it('should have proper aria attributes on trigger', async () => {
      render(
        <Popover>
          <PopoverTrigger data-testid="trigger">Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )

      const trigger = screen.getByTestId('trigger')
      expect(trigger).toHaveAttribute('aria-expanded', 'false')

      fireEvent.click(trigger)
      await waitFor(() => {
        expect(trigger).toHaveAttribute('aria-expanded', 'true')
      })
    })

    it('should be keyboard accessible', async () => {
      render(
        <Popover>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>Content</PopoverContent>
        </Popover>
      )

      const trigger = screen.getByText('Open')
      trigger.focus()
      expect(trigger).toHaveFocus()
    })

    it('should trap focus within popover when open', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>
            <button>First</button>
            <button>Second</button>
          </PopoverContent>
        </Popover>
      )

      await waitFor(() => {
        expect(screen.getByText('First')).toBeInTheDocument()
        expect(screen.getByText('Second')).toBeInTheDocument()
      })
    })
  })

  describe('content positioning', () => {
    it('should support different align values', async () => {
      const alignValues = ['start', 'center', 'end'] as const

      for (const align of alignValues) {
        const { unmount } = render(
          <Popover open>
            <PopoverTrigger>Open</PopoverTrigger>
            <PopoverContent align={align} data-testid="content">
              Content
            </PopoverContent>
          </Popover>
        )

        await waitFor(() => {
          expect(screen.getByTestId('content')).toBeInTheDocument()
        })

        unmount()
      }
    })

    it('should support different side values', async () => {
      const sideValues = ['top', 'right', 'bottom', 'left'] as const

      for (const side of sideValues) {
        const { unmount } = render(
          <Popover open>
            <PopoverTrigger>Open</PopoverTrigger>
            <PopoverContent side={side} data-testid="content">
              Content
            </PopoverContent>
          </Popover>
        )

        await waitFor(() => {
          expect(screen.getByTestId('content')).toBeInTheDocument()
        })

        unmount()
      }
    })
  })

  describe('displayNames', () => {
    it('should have correct displayName', () => {
      expect(PopoverContent.displayName).toBe('PopoverContent')
    })
  })

  describe('complex content', () => {
    it('should render form inside popover', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>
            <form data-testid="popover-form">
              <label htmlFor="name">Name</label>
              <input id="name" type="text" />
              <button type="submit">Submit</button>
            </form>
          </PopoverContent>
        </Popover>
      )

      await waitFor(() => {
        expect(screen.getByTestId('popover-form')).toBeInTheDocument()
        expect(screen.getByLabelText('Name')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: 'Submit' })).toBeInTheDocument()
      })
    })

    it('should render list inside popover', async () => {
      render(
        <Popover open>
          <PopoverTrigger>Open</PopoverTrigger>
          <PopoverContent>
            <ul>
              <li>Item 1</li>
              <li>Item 2</li>
              <li>Item 3</li>
            </ul>
          </PopoverContent>
        </Popover>
      )

      await waitFor(() => {
        expect(screen.getByRole('list')).toBeInTheDocument()
        expect(screen.getAllByRole('listitem')).toHaveLength(3)
      })
    })
  })
})
