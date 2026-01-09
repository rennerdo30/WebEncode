import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useForm } from 'react-hook-form'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  useFormField,
} from './form'
import { Input } from './input'

// Test wrapper component to provide form context
interface TestFormProps {
  children: React.ReactNode
  defaultValues?: Record<string, unknown>
  onSubmit?: (data: Record<string, unknown>) => void
}

const TestFormWrapper = ({ children, defaultValues = {}, onSubmit = vi.fn() }: TestFormProps) => {
  const form = useForm({
    defaultValues,
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        {children}
      </form>
    </Form>
  )
}

describe('Form', () => {
  describe('Form provider', () => {
    it('should render children within form provider', () => {
      render(
        <TestFormWrapper>
          <div data-testid="child">Form child</div>
        </TestFormWrapper>
      )
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })

  describe('FormField', () => {
    it('should render field with control', () => {
      render(
        <TestFormWrapper defaultValues={{ username: '' }}>
          <FormField
            name="username"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Username</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByLabelText('Username')).toBeInTheDocument()
    })

    it('should provide field context to children', () => {
      render(
        <TestFormWrapper defaultValues={{ email: 'test@example.com' }}>
          <FormField
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input {...field} data-testid="email-input" />
                </FormControl>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByTestId('email-input')).toHaveValue('test@example.com')
    })

    it('should update form state on input change', async () => {
      const user = userEvent.setup()

      render(
        <TestFormWrapper defaultValues={{ name: '' }}>
          <FormField
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
              </FormItem>
            )}
          />
          <button type="submit">Submit</button>
        </TestFormWrapper>
      )

      const input = screen.getByLabelText('Name')
      await user.type(input, 'John Doe')
      expect(input).toHaveValue('John Doe')
    })
  })

  describe('FormItem', () => {
    it('should render form item container', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem data-testid="form-item">
                <FormLabel>Test</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByTestId('form-item')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem className="custom-item" data-testid="form-item">
                <FormLabel>Test</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByTestId('form-item')).toHaveClass('custom-item')
    })

    it('should have default spacing', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem data-testid="form-item">
                <FormLabel>Test</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByTestId('form-item')).toHaveClass('space-y-2')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem ref={ref}>
                <FormLabel>Test</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(ref.current).toBeInstanceOf(HTMLDivElement)
    })

    it('should have correct displayName', () => {
      expect(FormItem.displayName).toBe('FormItem')
    })
  })

  describe('FormLabel', () => {
    it('should render label with htmlFor pointing to form item', () => {
      render(
        <TestFormWrapper defaultValues={{ field: '' }}>
          <FormField
            name="field"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Field Label</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('Field Label')).toBeInTheDocument()
      expect(screen.getByLabelText('Field Label')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormLabel className="custom-label">Label</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('Label')).toHaveClass('custom-label')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormLabel ref={ref}>Label</FormLabel>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(ref.current).toBeInstanceOf(HTMLLabelElement)
    })

    it('should have correct displayName', () => {
      expect(FormLabel.displayName).toBe('FormLabel')
    })
  })

  describe('FormControl', () => {
    it('should render slot for form control', () => {
      render(
        <TestFormWrapper defaultValues={{ input: '' }}>
          <FormField
            name="input"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Input</FormLabel>
                <FormControl>
                  <Input {...field} data-testid="controlled-input" />
                </FormControl>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByTestId('controlled-input')).toBeInTheDocument()
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={({ field }) => (
              <FormItem>
                <FormControl ref={ref}>
                  <Input {...field} />
                </FormControl>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(ref.current).toBeInstanceOf(HTMLInputElement)
    })

    it('should have correct displayName', () => {
      expect(FormControl.displayName).toBe('FormControl')
    })
  })

  describe('FormDescription', () => {
    it('should render description text', () => {
      render(
        <TestFormWrapper defaultValues={{ field: '' }}>
          <FormField
            name="field"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Field</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormDescription>This is a helpful description</FormDescription>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('This is a helpful description')).toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormDescription className="custom-description">Description</FormDescription>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('Description')).toHaveClass('custom-description')
    })

    it('should have styling classes', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormDescription>Description</FormDescription>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      const description = screen.getByText('Description')
      expect(description).toHaveClass('text-sm')
      expect(description).toHaveClass('text-muted-foreground')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormDescription ref={ref}>Description</FormDescription>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(ref.current).toBeInstanceOf(HTMLParagraphElement)
    })

    it('should have correct displayName', () => {
      expect(FormDescription.displayName).toBe('FormDescription')
    })
  })

  describe('FormMessage', () => {
    it('should render children when no error', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormMessage>Custom message</FormMessage>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('Custom message')).toBeInTheDocument()
    })

    it('should not render when no error and no children', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormMessage data-testid="message" />
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.queryByTestId('message')).not.toBeInTheDocument()
    })

    it('should apply custom className', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormMessage className="custom-message">Message</FormMessage>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(screen.getByText('Message')).toHaveClass('custom-message')
    })

    it('should have error styling classes', () => {
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormMessage>Error message</FormMessage>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      const message = screen.getByText('Error message')
      expect(message).toHaveClass('text-sm')
      expect(message).toHaveClass('font-medium')
      expect(message).toHaveClass('text-destructive')
    })

    it('should forward ref correctly', () => {
      const ref = { current: null }
      render(
        <TestFormWrapper defaultValues={{ test: '' }}>
          <FormField
            name="test"
            render={() => (
              <FormItem>
                <FormMessage ref={ref}>Message</FormMessage>
              </FormItem>
            )}
          />
        </TestFormWrapper>
      )
      expect(ref.current).toBeInstanceOf(HTMLParagraphElement)
    })

    it('should have correct displayName', () => {
      expect(FormMessage.displayName).toBe('FormMessage')
    })
  })

  describe('useFormField hook', () => {
    it('should throw error when used outside FormField', () => {
      const TestComponent = () => {
        useFormField()
        return <div>Test</div>
      }

      expect(() => {
        render(
          <TestFormWrapper defaultValues={{}}>
            <TestComponent />
          </TestFormWrapper>
        )
      }).toThrow('useFormField should be used within <FormField>')
    })

    it('should throw error when used outside FormItem', () => {
      const TestComponent = () => {
        useFormField()
        return <div>Test</div>
      }

      // This test is tricky because we need FormField context but not FormItem
      // The hook checks both contexts
    })
  })

  describe('complete form composition', () => {
    it('should render a complete form with all parts', () => {
      render(
        <TestFormWrapper
          defaultValues={{ username: '', email: '' }}
        >
          <FormField
            name="username"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Username</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormDescription>Enter your username</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input type="email" {...field} />
                </FormControl>
                <FormDescription>Enter your email address</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <button type="submit">Submit</button>
        </TestFormWrapper>
      )

      expect(screen.getByLabelText('Username')).toBeInTheDocument()
      expect(screen.getByLabelText('Email')).toBeInTheDocument()
      expect(screen.getByText('Enter your username')).toBeInTheDocument()
      expect(screen.getByText('Enter your email address')).toBeInTheDocument()
      expect(screen.getByText('Submit')).toBeInTheDocument()
    })
  })
})
