import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { NotificationsDropdown } from './notifications-dropdown'
import * as api from '@/lib/api'

// Mock the API module
vi.mock('@/lib/api', () => ({
  fetchNotifications: vi.fn(),
  markNotificationRead: vi.fn(),
  clearAllNotifications: vi.fn(),
}))

const mockFetchNotifications = vi.mocked(api.fetchNotifications)
const mockMarkNotificationRead = vi.mocked(api.markNotificationRead)
const mockClearAllNotifications = vi.mocked(api.clearAllNotifications)

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  })
}

function renderWithQueryClient(ui: React.ReactElement) {
  const queryClient = createTestQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  )
}

describe('NotificationsDropdown', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render bell icon button', async () => {
      mockFetchNotifications.mockResolvedValue([])

      renderWithQueryClient(<NotificationsDropdown />)

      // Bell icon button should be visible
      const button = screen.getByRole('button')
      expect(button).toBeInTheDocument()
    })

    it('should not show unread badge when no unread notifications', async () => {
      mockFetchNotifications.mockResolvedValue([])

      renderWithQueryClient(<NotificationsDropdown />)

      await waitFor(() => {
        // No unread badge (purple dot)
        expect(document.querySelector('.bg-violet-500')).not.toBeInTheDocument()
      })
    })

    it('should show unread badge when there are unread notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Test notification',
          message: 'Test message',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      await waitFor(() => {
        expect(document.querySelector('.bg-violet-500')).toBeInTheDocument()
      })
    })
  })

  describe('popover', () => {
    it('should open popover when button is clicked', async () => {
      mockFetchNotifications.mockResolvedValue([])

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.getByText('Notifications')).toBeInTheDocument()
      })
    })

    it('should show loading state while fetching', () => {
      mockFetchNotifications.mockImplementation(() => new Promise(() => {}))

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      expect(screen.getByText('Loading...')).toBeInTheDocument()
    })

    it('should show empty state when no notifications', async () => {
      mockFetchNotifications.mockResolvedValue([])

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.getByText('No notifications')).toBeInTheDocument()
      })
    })
  })

  describe('notifications list', () => {
    it('should render notification items', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Job completed',
          message: 'Your encoding job has finished',
          type: 'success',
          is_read: false,
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          title: 'Warning',
          message: 'Disk space running low',
          type: 'warning',
          is_read: true,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.getByText('Job completed')).toBeInTheDocument()
        expect(screen.getByText('Your encoding job has finished')).toBeInTheDocument()
        expect(screen.getByText('Warning')).toBeInTheDocument()
        expect(screen.getByText('Disk space running low')).toBeInTheDocument()
      })
    })

    it('should show success icon for success notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Success',
          message: 'Operation succeeded',
          type: 'success',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.text-green-500')).toBeInTheDocument()
      })
    })

    it('should show warning icon for warning notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Warning',
          message: 'Caution needed',
          type: 'warning',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.text-yellow-500')).toBeInTheDocument()
      })
    })

    it('should show error icon for error notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Error',
          message: 'Something went wrong',
          type: 'error',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.text-red-500')).toBeInTheDocument()
      })
    })

    it('should show info icon for info notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Info',
          message: 'For your information',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.text-blue-500')).toBeInTheDocument()
      })
    })

    it('should highlight unread notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unread',
          message: 'This is unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.bg-muted\\/20')).toBeInTheDocument()
      })
    })
  })

  describe('mark as read', () => {
    it('should show mark as read button on unread notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unread',
          message: 'This is unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.getByText('Mark as read')).toBeInTheDocument()
      })
    })

    it('should call markNotificationRead when mark as read button is clicked', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: 'notif-123',
          title: 'Unread',
          message: 'This is unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)
      mockMarkNotificationRead.mockResolvedValue(undefined)

      renderWithQueryClient(<NotificationsDropdown />)

      const triggerButton = screen.getByRole('button')
      fireEvent.click(triggerButton)

      await waitFor(() => {
        const markReadButton = screen.getByRole('button', { name: 'Mark as read' })
        fireEvent.click(markReadButton)
      })

      await waitFor(() => {
        expect(mockMarkNotificationRead).toHaveBeenCalledWith('notif-123')
      })
    })
  })

  describe('clear all', () => {
    it('should show mark all read button when there are unread notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unread',
          message: 'This is unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.getByText('Mark all read')).toBeInTheDocument()
      })
    })

    it('should not show mark all read button when no unread notifications', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Read',
          message: 'This is read',
          type: 'info',
          is_read: true,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(screen.queryByText('Mark all read')).not.toBeInTheDocument()
      })
    })

    it('should call clearAllNotifications when mark all read button is clicked', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unread',
          message: 'This is unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)
      mockClearAllNotifications.mockResolvedValue(undefined)

      renderWithQueryClient(<NotificationsDropdown />)

      const triggerButton = screen.getByRole('button')
      fireEvent.click(triggerButton)

      await waitFor(() => {
        const clearButton = screen.getByText('Mark all read')
        fireEvent.click(clearButton)
      })

      await waitFor(() => {
        expect(mockClearAllNotifications).toHaveBeenCalled()
      })
    })
  })

  describe('notification types styling', () => {
    it('should apply different styles for different notification types', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Success',
          message: 'Operation succeeded',
          type: 'success',
          is_read: false,
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          title: 'Error',
          message: 'Operation failed',
          type: 'error',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        expect(document.querySelector('.text-green-500')).toBeInTheDocument()
        expect(document.querySelector('.text-red-500')).toBeInTheDocument()
      })
    })

    it('should show default info icon for unknown notification types', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unknown',
          message: 'Unknown type',
          type: 'unknown' as api.Notification['type'],
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      const button = screen.getByRole('button')
      fireEvent.click(button)

      await waitFor(() => {
        // Default is blue info icon
        expect(document.querySelector('.text-blue-500')).toBeInTheDocument()
      })
    })
  })

  describe('unread count', () => {
    it('should calculate correct unread count', async () => {
      const mockNotifications: api.Notification[] = [
        {
          id: '1',
          title: 'Unread 1',
          message: 'Unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
        {
          id: '2',
          title: 'Read',
          message: 'Read',
          type: 'info',
          is_read: true,
          created_at: new Date().toISOString(),
        },
        {
          id: '3',
          title: 'Unread 2',
          message: 'Unread',
          type: 'info',
          is_read: false,
          created_at: new Date().toISOString(),
        },
      ]
      mockFetchNotifications.mockResolvedValue(mockNotifications)

      renderWithQueryClient(<NotificationsDropdown />)

      await waitFor(() => {
        // Badge should be visible (2 unread)
        expect(document.querySelector('.bg-violet-500')).toBeInTheDocument()
      })
    })
  })
})
