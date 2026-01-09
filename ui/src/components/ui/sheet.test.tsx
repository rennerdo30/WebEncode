import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import {
  Sheet,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetFooter,
  SheetTitle,
  SheetDescription,
  SheetClose,
  SheetOverlay,
  SheetPortal,
} from './sheet'

describe('Sheet', () => {
  describe('Sheet Root', () => {
    it('should render sheet trigger', () => {
      render(
        <Sheet>
          <SheetTrigger>Open Sheet</SheetTrigger>
          <SheetContent>
            <SheetTitle>Test Sheet</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      expect(screen.getByText('Open Sheet')).toBeInTheDocument()
    })

    it('should open sheet when trigger is clicked', async () => {
      render(
        <Sheet>
          <SheetTrigger>Open Sheet</SheetTrigger>
          <SheetContent>
            <SheetTitle>Test Sheet</SheetTitle>
            <p>Sheet content</p>
          </SheetContent>
        </Sheet>
      )

      fireEvent.click(screen.getByText('Open Sheet'))
      await waitFor(() => {
        expect(screen.getByText('Sheet content')).toBeInTheDocument()
      })
    })

    it('should support controlled mode', async () => {
      const { rerender } = render(
        <Sheet open={false}>
          <SheetTrigger>Open</SheetTrigger>
          <SheetContent>
            <SheetTitle>Test</SheetTitle>
            <p>Content</p>
          </SheetContent>
        </Sheet>
      )
      expect(screen.queryByText('Content')).not.toBeInTheDocument()

      rerender(
        <Sheet open={true}>
          <SheetTrigger>Open</SheetTrigger>
          <SheetContent>
            <SheetTitle>Test</SheetTitle>
            <p>Content</p>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })
    })

    it('should call onOpenChange when state changes', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Sheet onOpenChange={handleOpenChange}>
          <SheetTrigger>Open Sheet</SheetTrigger>
          <SheetContent>
            <SheetTitle>Test Sheet</SheetTitle>
          </SheetContent>
        </Sheet>
      )

      fireEvent.click(screen.getByText('Open Sheet'))
      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(true)
      })
    })
  })

  describe('SheetContent', () => {
    it('should render content with children', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <p>Test content</p>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Test content')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Sheet open>
          <SheetContent className="custom-sheet">
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toHaveClass('custom-sheet')
      })
    })

    it('should render close button', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Close')).toBeInTheDocument()
      })
    })

    it('should have proper styling classes', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const sheet = screen.getByRole('dialog')
        expect(sheet).toHaveClass('fixed')
        expect(sheet).toHaveClass('z-50')
        expect(sheet).toHaveClass('bg-background')
      })
    })

    describe('side variants', () => {
      it('should render right side by default', async () => {
        render(
          <Sheet open>
            <SheetContent>
              <SheetTitle>Title</SheetTitle>
            </SheetContent>
          </Sheet>
        )
        await waitFor(() => {
          const sheet = screen.getByRole('dialog')
          expect(sheet).toHaveClass('inset-y-0')
          expect(sheet).toHaveClass('right-0')
        })
      })

      it('should render left side', async () => {
        render(
          <Sheet open>
            <SheetContent side="left">
              <SheetTitle>Title</SheetTitle>
            </SheetContent>
          </Sheet>
        )
        await waitFor(() => {
          const sheet = screen.getByRole('dialog')
          expect(sheet).toHaveClass('inset-y-0')
          expect(sheet).toHaveClass('left-0')
        })
      })

      it('should render top side', async () => {
        render(
          <Sheet open>
            <SheetContent side="top">
              <SheetTitle>Title</SheetTitle>
            </SheetContent>
          </Sheet>
        )
        await waitFor(() => {
          const sheet = screen.getByRole('dialog')
          expect(sheet).toHaveClass('inset-x-0')
          expect(sheet).toHaveClass('top-0')
        })
      })

      it('should render bottom side', async () => {
        render(
          <Sheet open>
            <SheetContent side="bottom">
              <SheetTitle>Title</SheetTitle>
            </SheetContent>
          </Sheet>
        )
        await waitFor(() => {
          const sheet = screen.getByRole('dialog')
          expect(sheet).toHaveClass('inset-x-0')
          expect(sheet).toHaveClass('bottom-0')
        })
      })
    })
  })

  describe('SheetHeader', () => {
    it('should render header with children', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>Header Title</SheetTitle>
            </SheetHeader>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Header Title')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetHeader className="custom-header" data-testid="header">
              <SheetTitle>Title</SheetTitle>
            </SheetHeader>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByTestId('header')).toHaveClass('custom-header')
      })
    })

    it('should have flex column layout', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetHeader data-testid="header">
              <SheetTitle>Title</SheetTitle>
            </SheetHeader>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByTestId('header')).toHaveClass('flex')
        expect(screen.getByTestId('header')).toHaveClass('flex-col')
      })
    })

    it('should have correct displayName', () => {
      expect(SheetHeader.displayName).toBe('SheetHeader')
    })
  })

  describe('SheetFooter', () => {
    it('should render footer with children', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetFooter>
              <button>Cancel</button>
              <button>Confirm</button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Cancel')).toBeInTheDocument()
        expect(screen.getByText('Confirm')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetFooter className="custom-footer" data-testid="footer">
              <button>OK</button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByTestId('footer')).toHaveClass('custom-footer')
      })
    })

    it('should have responsive layout classes', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetFooter data-testid="footer">
              <button>OK</button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const footer = screen.getByTestId('footer')
        expect(footer).toHaveClass('flex-col-reverse')
        expect(footer).toHaveClass('sm:flex-row')
        expect(footer).toHaveClass('sm:justify-end')
      })
    })

    it('should have correct displayName', () => {
      expect(SheetFooter.displayName).toBe('SheetFooter')
    })
  })

  describe('SheetTitle', () => {
    it('should render title text', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>My Sheet Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('My Sheet Title')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle className="custom-title">Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Title')).toHaveClass('custom-title')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const title = screen.getByText('Title')
        expect(title).toHaveClass('text-lg')
        expect(title).toHaveClass('font-semibold')
        expect(title).toHaveClass('text-foreground')
      })
    })

    it('should forward ref correctly', async () => {
      const ref = { current: null }
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle ref={ref}>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLHeadingElement)
      })
    })
  })

  describe('SheetDescription', () => {
    it('should render description text', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription>This is a description</SheetDescription>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('This is a description')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription className="custom-desc">Description</SheetDescription>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Description')).toHaveClass('custom-desc')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription>Description</SheetDescription>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const desc = screen.getByText('Description')
        expect(desc).toHaveClass('text-sm')
        expect(desc).toHaveClass('text-muted-foreground')
      })
    })

    it('should forward ref correctly', async () => {
      const ref = { current: null }
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription ref={ref}>Description</SheetDescription>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLParagraphElement)
      })
    })
  })

  describe('SheetClose', () => {
    it('should close sheet when clicked', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Sheet open onOpenChange={handleOpenChange}>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetClose data-testid="close-btn">Close Me</SheetClose>
          </SheetContent>
        </Sheet>
      )

      await waitFor(() => {
        fireEvent.click(screen.getByTestId('close-btn'))
      })
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })
  })

  describe('SheetOverlay', () => {
    it('should render overlay', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const overlay = document.querySelector('[data-state="open"][class*="bg-black"]')
        expect(overlay).toBeInTheDocument()
      })
    })

    it('should forward ref to SheetOverlay', async () => {
      const ref = { current: null }
      const TestComponent = () => (
        <Sheet open>
          <SheetPortal>
            <SheetOverlay ref={ref} data-testid="overlay" />
          </SheetPortal>
        </Sheet>
      )
      render(<TestComponent />)
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLDivElement)
      })
    })
  })

  describe('accessibility', () => {
    it('should have role="dialog"', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Accessible Sheet</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })

    it('should be labeled by title', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Sheet Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveAttribute('aria-labelledby')
      })
    })

    it('should be described by description', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription>Description text</SheetDescription>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveAttribute('aria-describedby')
      })
    })

    it('should close on escape key', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Sheet open onOpenChange={handleOpenChange}>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )

      await waitFor(() => {
        fireEvent.keyDown(document, { key: 'Escape' })
      })
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })

    it('should have sr-only text for close button', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetTitle>Title</SheetTitle>
          </SheetContent>
        </Sheet>
      )
      await waitFor(() => {
        expect(screen.getByText('Close')).toHaveClass('sr-only')
      })
    })
  })

  describe('complete sheet composition', () => {
    it('should render a complete sheet with all parts', async () => {
      render(
        <Sheet open>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>Complete Sheet</SheetTitle>
              <SheetDescription>This is a complete sheet example</SheetDescription>
            </SheetHeader>
            <div data-testid="body">Sheet body content</div>
            <SheetFooter>
              <button type="button">Cancel</button>
              <button type="button">Save</button>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      )

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText('Complete Sheet')).toBeInTheDocument()
        expect(screen.getByText('This is a complete sheet example')).toBeInTheDocument()
        expect(screen.getByTestId('body')).toBeInTheDocument()
        expect(screen.getByText('Cancel')).toBeInTheDocument()
        expect(screen.getByText('Save')).toBeInTheDocument()
      })
    })
  })
})
