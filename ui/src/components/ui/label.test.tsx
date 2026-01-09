import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Label } from './label'

describe('Label', () => {
  describe('rendering', () => {
    it('should render label with text content', () => {
      render(<Label>Email</Label>)
      expect(screen.getByText('Email')).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      render(<Label className="custom-class">Test</Label>)
      expect(screen.getByText('Test')).toHaveClass('custom-class')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(<Label ref={ref}>Test</Label>)
      expect(ref.current).toBeInstanceOf(HTMLLabelElement)
    })

    it('should have correct displayName', () => {
      expect(Label.displayName).toBe('Label')
    })
  })

  describe('styling', () => {
    it('should have base styling classes', () => {
      render(<Label>Test</Label>)
      const label = screen.getByText('Test')
      expect(label).toHaveClass('text-sm')
      expect(label).toHaveClass('font-medium')
      expect(label).toHaveClass('leading-none')
    })

    it('should have peer-disabled styles', () => {
      render(<Label>Test</Label>)
      const label = screen.getByText('Test')
      expect(label).toHaveClass('peer-disabled:cursor-not-allowed')
      expect(label).toHaveClass('peer-disabled:opacity-70')
    })

    it('should merge custom className with default styles', () => {
      render(<Label className="text-red-500">Test</Label>)
      const label = screen.getByText('Test')
      expect(label).toHaveClass('text-sm')
      expect(label).toHaveClass('text-red-500')
    })
  })

  describe('association with form elements', () => {
    it('should associate with input via htmlFor', () => {
      render(
        <>
          <Label htmlFor="email-input">Email</Label>
          <input id="email-input" type="email" />
        </>
      )
      const label = screen.getByText('Email')
      expect(label).toHaveAttribute('for', 'email-input')
    })

    it('should wrap form element when used as container', () => {
      render(
        <Label>
          <span>Username</span>
          <input type="text" data-testid="wrapped-input" />
        </Label>
      )
      expect(screen.getByText('Username')).toBeInTheDocument()
      expect(screen.getByTestId('wrapped-input')).toBeInTheDocument()
    })
  })

  describe('HTML attributes', () => {
    it('should pass through data attributes', () => {
      render(<Label data-testid="my-label">Test</Label>)
      expect(screen.getByTestId('my-label')).toBeInTheDocument()
    })

    it('should support id attribute', () => {
      render(<Label id="form-label">Test</Label>)
      expect(screen.getByText('Test')).toHaveAttribute('id', 'form-label')
    })

    it('should support aria attributes', () => {
      render(<Label aria-describedby="hint">Test</Label>)
      expect(screen.getByText('Test')).toHaveAttribute('aria-describedby', 'hint')
    })
  })

  describe('children', () => {
    it('should render string children', () => {
      render(<Label>Simple text</Label>)
      expect(screen.getByText('Simple text')).toBeInTheDocument()
    })

    it('should render React element children', () => {
      render(
        <Label>
          <span data-testid="child">Child element</span>
        </Label>
      )
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })

    it('should render multiple children', () => {
      render(
        <Label>
          <span>Required</span>
          <span className="text-red-500">*</span>
        </Label>
      )
      expect(screen.getByText('Required')).toBeInTheDocument()
      expect(screen.getByText('*')).toBeInTheDocument()
    })
  })
})
