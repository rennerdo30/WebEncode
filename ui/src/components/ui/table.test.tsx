import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from './table'

describe('Table Components', () => {
  describe('Table', () => {
    it('should render table with children', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('table')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table className="custom-table">
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('table')).toHaveClass('custom-table')
    })

    it('should have base styling classes', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      const table = screen.getByRole('table')
      expect(table).toHaveClass('w-full')
      expect(table).toHaveClass('caption-bottom')
      expect(table).toHaveClass('text-sm')
    })

    it('should be wrapped in scrollable container', () => {
      render(
        <Table data-testid="table">
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      const table = screen.getByRole('table')
      const wrapper = table.parentElement
      expect(wrapper).toHaveClass('relative')
      expect(wrapper).toHaveClass('w-full')
      expect(wrapper).toHaveClass('overflow-auto')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table ref={ref}>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableElement)
    })
  })

  describe('TableHeader', () => {
    it('should render table header', () => {
      render(
        <Table>
          <TableHeader data-testid="table-header">
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('table-header')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableHeader className="custom-header" data-testid="header">
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('header')).toHaveClass('custom-header')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableHeader ref={ref}>
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement)
    })
  })

  describe('TableBody', () => {
    it('should render table body', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('rowgroup')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody className="custom-body" data-testid="body">
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('body')).toHaveClass('custom-body')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableBody ref={ref}>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement)
    })
  })

  describe('TableFooter', () => {
    it('should render table footer', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter data-testid="footer">
            <TableRow>
              <TableCell>Footer</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )
      expect(screen.getByTestId('footer')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter className="custom-footer" data-testid="footer">
            <TableRow>
              <TableCell>Footer</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )
      expect(screen.getByTestId('footer')).toHaveClass('custom-footer')
    })

    it('should have footer styling', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter data-testid="footer">
            <TableRow>
              <TableCell>Footer</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )
      const footer = screen.getByTestId('footer')
      expect(footer).toHaveClass('border-t')
      expect(footer).toHaveClass('font-medium')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter ref={ref}>
            <TableRow>
              <TableCell>Footer</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableSectionElement)
    })
  })

  describe('TableRow', () => {
    it('should render table row', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('row')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow className="custom-row" data-testid="row">
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('row')).toHaveClass('custom-row')
    })

    it('should have hover styles', () => {
      render(
        <Table>
          <TableBody>
            <TableRow data-testid="row">
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('row')).toHaveClass('hover:bg-muted/50')
    })

    it('should have border', () => {
      render(
        <Table>
          <TableBody>
            <TableRow data-testid="row">
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('row')).toHaveClass('border-b')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableBody>
            <TableRow ref={ref}>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableRowElement)
    })
  })

  describe('TableHead', () => {
    it('should render table head cell', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('columnheader')).toHaveTextContent('Header')
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="custom-head" data-testid="head">Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('head')).toHaveClass('custom-head')
    })

    it('should have header cell styling', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead data-testid="head">Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      const head = screen.getByTestId('head')
      expect(head).toHaveClass('h-12')
      expect(head).toHaveClass('px-4')
      expect(head).toHaveClass('text-left')
      expect(head).toHaveClass('font-medium')
      expect(head).toHaveClass('text-muted-foreground')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead ref={ref}>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableCellElement)
    })
  })

  describe('TableCell', () => {
    it('should render table cell', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell content</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByRole('cell')).toHaveTextContent('Cell content')
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell className="custom-cell" data-testid="cell">Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('cell')).toHaveClass('custom-cell')
    })

    it('should have cell styling', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell data-testid="cell">Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      const cell = screen.getByTestId('cell')
      expect(cell).toHaveClass('p-4')
      expect(cell).toHaveClass('align-middle')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell ref={ref}>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableCellElement)
    })
  })

  describe('TableCaption', () => {
    it('should render table caption', () => {
      render(
        <Table>
          <TableCaption>Table caption</TableCaption>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByText('Table caption')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <Table>
          <TableCaption className="custom-caption" data-testid="caption">Caption</TableCaption>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(screen.getByTestId('caption')).toHaveClass('custom-caption')
    })

    it('should have caption styling', () => {
      render(
        <Table>
          <TableCaption data-testid="caption">Caption</TableCaption>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      const caption = screen.getByTestId('caption')
      expect(caption).toHaveClass('mt-4')
      expect(caption).toHaveClass('text-sm')
      expect(caption).toHaveClass('text-muted-foreground')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <Table>
          <TableCaption ref={ref}>Caption</TableCaption>
          <TableBody>
            <TableRow>
              <TableCell>Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )
      expect(ref.current).toBeInstanceOf(HTMLTableCaptionElement)
    })
  })

  describe('Complete Table composition', () => {
    it('should render a complete data table', () => {
      render(
        <Table>
          <TableCaption>A list of recent transactions</TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead>Invoice</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Amount</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>INV001</TableCell>
              <TableCell>Paid</TableCell>
              <TableCell>$250.00</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>INV002</TableCell>
              <TableCell>Pending</TableCell>
              <TableCell>$150.00</TableCell>
            </TableRow>
          </TableBody>
          <TableFooter>
            <TableRow>
              <TableCell colSpan={2}>Total</TableCell>
              <TableCell>$400.00</TableCell>
            </TableRow>
          </TableFooter>
        </Table>
      )

      expect(screen.getByRole('table')).toBeInTheDocument()
      expect(screen.getByText('A list of recent transactions')).toBeInTheDocument()
      expect(screen.getByText('Invoice')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
      expect(screen.getByText('Amount')).toBeInTheDocument()
      expect(screen.getByText('INV001')).toBeInTheDocument()
      expect(screen.getByText('Paid')).toBeInTheDocument()
      expect(screen.getByText('$400.00')).toBeInTheDocument()
    })

    it('should have correct number of rows', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Header</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Row 1</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Row 2</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Row 3</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      )

      // 1 header row + 3 body rows = 4 total
      expect(screen.getAllByRole('row')).toHaveLength(4)
    })
  })
})
