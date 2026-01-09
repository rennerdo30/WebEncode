import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  DialogClose,
  DialogOverlay,
  DialogPortal,
} from './dialog'

describe('Dialog', () => {
  describe('Dialog Root', () => {
    it('should render dialog trigger', () => {
      render(
        <Dialog>
          <DialogTrigger>Open Dialog</DialogTrigger>
          <DialogContent>
            <DialogTitle>Test Dialog</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      expect(screen.getByText('Open Dialog')).toBeInTheDocument()
    })

    it('should open dialog when trigger is clicked', async () => {
      render(
        <Dialog>
          <DialogTrigger>Open Dialog</DialogTrigger>
          <DialogContent>
            <DialogTitle>Test Dialog</DialogTitle>
            <p>Dialog content</p>
          </DialogContent>
        </Dialog>
      )

      fireEvent.click(screen.getByText('Open Dialog'))
      await waitFor(() => {
        expect(screen.getByText('Dialog content')).toBeInTheDocument()
      })
    })

    it('should support controlled mode', async () => {
      const { rerender } = render(
        <Dialog open={false}>
          <DialogTrigger>Open</DialogTrigger>
          <DialogContent>
            <DialogTitle>Test</DialogTitle>
            <p>Content</p>
          </DialogContent>
        </Dialog>
      )
      expect(screen.queryByText('Content')).not.toBeInTheDocument()

      rerender(
        <Dialog open={true}>
          <DialogTrigger>Open</DialogTrigger>
          <DialogContent>
            <DialogTitle>Test</DialogTitle>
            <p>Content</p>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Content')).toBeInTheDocument()
      })
    })

    it('should call onOpenChange when state changes', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Dialog onOpenChange={handleOpenChange}>
          <DialogTrigger>Open Dialog</DialogTrigger>
          <DialogContent>
            <DialogTitle>Test Dialog</DialogTitle>
          </DialogContent>
        </Dialog>
      )

      fireEvent.click(screen.getByText('Open Dialog'))
      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(true)
      })
    })
  })

  describe('DialogContent', () => {
    it('should render content with children', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <p>Test content</p>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Test content')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Dialog open>
          <DialogContent className="custom-dialog">
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toHaveClass('custom-dialog')
      })
    })

    it('should render close button', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Close')).toBeInTheDocument()
      })
    })

    it('should have proper styling classes', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveClass('fixed')
        expect(dialog).toHaveClass('z-50')
        expect(dialog).toHaveClass('border')
        expect(dialog).toHaveClass('bg-background')
      })
    })
  })

  describe('DialogHeader', () => {
    it('should render header with children', () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Header Title</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      )
      expect(screen.getByText('Header Title')).toBeInTheDocument()
    })

    it('should apply custom className', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogHeader className="custom-header" data-testid="header">
              <DialogTitle>Title</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByTestId('header')).toHaveClass('custom-header')
      })
    })

    it('should have flex column layout', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogHeader data-testid="header">
              <DialogTitle>Title</DialogTitle>
            </DialogHeader>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByTestId('header')).toHaveClass('flex')
        expect(screen.getByTestId('header')).toHaveClass('flex-col')
      })
    })

    it('should have correct displayName', () => {
      expect(DialogHeader.displayName).toBe('DialogHeader')
    })
  })

  describe('DialogFooter', () => {
    it('should render footer with children', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogFooter>
              <button>Cancel</button>
              <button>Confirm</button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Cancel')).toBeInTheDocument()
        expect(screen.getByText('Confirm')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogFooter className="custom-footer" data-testid="footer">
              <button>OK</button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByTestId('footer')).toHaveClass('custom-footer')
      })
    })

    it('should have responsive layout classes', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogFooter data-testid="footer">
              <button>OK</button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const footer = screen.getByTestId('footer')
        expect(footer).toHaveClass('flex-col-reverse')
        expect(footer).toHaveClass('sm:flex-row')
        expect(footer).toHaveClass('sm:justify-end')
      })
    })

    it('should have correct displayName', () => {
      expect(DialogFooter.displayName).toBe('DialogFooter')
    })
  })

  describe('DialogTitle', () => {
    it('should render title text', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>My Dialog Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('My Dialog Title')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle className="custom-title">Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Title')).toHaveClass('custom-title')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const title = screen.getByText('Title')
        expect(title).toHaveClass('text-lg')
        expect(title).toHaveClass('font-semibold')
        expect(title).toHaveClass('leading-none')
      })
    })

    it('should forward ref correctly', async () => {
      const ref = { current: null }
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle ref={ref}>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLHeadingElement)
      })
    })
  })

  describe('DialogDescription', () => {
    it('should render description text', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription>This is a description</DialogDescription>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('This is a description')).toBeInTheDocument()
      })
    })

    it('should apply custom className', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription className="custom-desc">Description</DialogDescription>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Description')).toHaveClass('custom-desc')
      })
    })

    it('should have styling classes', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription>Description</DialogDescription>
          </DialogContent>
        </Dialog>
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
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription ref={ref}>Description</DialogDescription>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(ref.current).toBeInstanceOf(HTMLParagraphElement)
      })
    })
  })

  describe('DialogClose', () => {
    it('should close dialog when clicked', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Dialog open onOpenChange={handleOpenChange}>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogClose data-testid="close-btn">Close Me</DialogClose>
          </DialogContent>
        </Dialog>
      )

      await waitFor(() => {
        fireEvent.click(screen.getByTestId('close-btn'))
      })
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })
  })

  describe('DialogOverlay', () => {
    it('should render overlay with proper classes', async () => {
      const ref = { current: null }
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const overlay = document.querySelector('[data-state="open"]')
        expect(overlay).toBeInTheDocument()
      })
    })

    it('should forward ref to DialogOverlay', async () => {
      const ref = { current: null }
      const TestComponent = () => (
        <Dialog open>
          <DialogPortal>
            <DialogOverlay ref={ref} data-testid="overlay" />
          </DialogPortal>
        </Dialog>
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
        <Dialog open>
          <DialogContent>
            <DialogTitle>Accessible Dialog</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })

    it('should be labeled by title', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Dialog Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveAttribute('aria-labelledby')
      })
    })

    it('should be described by description', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription>Description text</DialogDescription>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveAttribute('aria-describedby')
      })
    })

    it('should close on escape key', async () => {
      const handleOpenChange = vi.fn()
      render(
        <Dialog open onOpenChange={handleOpenChange}>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )

      await waitFor(() => {
        fireEvent.keyDown(document, { key: 'Escape' })
      })
      expect(handleOpenChange).toHaveBeenCalledWith(false)
    })

    it('should have sr-only text for close button', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogTitle>Title</DialogTitle>
          </DialogContent>
        </Dialog>
      )
      await waitFor(() => {
        expect(screen.getByText('Close')).toHaveClass('sr-only')
      })
    })
  })

  describe('complete dialog composition', () => {
    it('should render a complete dialog with all parts', async () => {
      render(
        <Dialog open>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Complete Dialog</DialogTitle>
              <DialogDescription>This is a complete dialog example</DialogDescription>
            </DialogHeader>
            <div data-testid="body">Dialog body content</div>
            <DialogFooter>
              <button type="button">Cancel</button>
              <button type="button">Save</button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText('Complete Dialog')).toBeInTheDocument()
        expect(screen.getByText('This is a complete dialog example')).toBeInTheDocument()
        expect(screen.getByTestId('body')).toBeInTheDocument()
        expect(screen.getByText('Cancel')).toBeInTheDocument()
        expect(screen.getByText('Save')).toBeInTheDocument()
      })
    })
  })
})
