import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ChatWidget } from './chat-widget'
import * as api from '@/lib/api'

// Mock the API module
vi.mock('@/lib/api', () => ({
  fetchStreamChat: vi.fn(),
  sendStreamChat: vi.fn(),
}))

const mockFetchStreamChat = vi.mocked(api.fetchStreamChat)
const mockSendStreamChat = vi.mocked(api.sendStreamChat)

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

describe('ChatWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render chat widget with title', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      expect(screen.getByText('Live Chat')).toBeInTheDocument()
    })

    it('should render loading state while fetching messages', () => {
      mockFetchStreamChat.mockImplementation(() => new Promise(() => {})) // Never resolves

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      // Check for loading spinner (Loader2 component with animate-spin class)
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })

    it('should render empty state when no messages', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        expect(screen.getByText('No messages yet')).toBeInTheDocument()
      })
      expect(screen.getByText('Messages from connected platforms will appear here')).toBeInTheDocument()
    })

    it('should render with custom className', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      const { container } = renderWithQueryClient(
        <ChatWidget streamId="test-stream-123" className="custom-class" />
      )

      await waitFor(() => {
        expect(container.firstChild).toHaveClass('custom-class')
      })
    })
  })

  describe('messages', () => {
    it('should render chat messages', async () => {
      const mockMessages: api.ChatMessage[] = [
        {
          id: '1',
          stream_id: 'test-stream',
          platform: 'twitch',
          author_name: 'TestUser',
          content: 'Hello world!',
          timestamp: Math.floor(Date.now() / 1000),
        },
        {
          id: '2',
          stream_id: 'test-stream',
          platform: 'youtube',
          author_name: 'AnotherUser',
          content: 'Hi there!',
          timestamp: Math.floor(Date.now() / 1000),
        },
      ]
      mockFetchStreamChat.mockResolvedValue(mockMessages)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        expect(screen.getByText('Hello world!')).toBeInTheDocument()
        expect(screen.getByText('Hi there!')).toBeInTheDocument()
      })
      expect(screen.getByText('TestUser')).toBeInTheDocument()
      expect(screen.getByText('AnotherUser')).toBeInTheDocument()
    })

    it('should display platform badges', async () => {
      const mockMessages: api.ChatMessage[] = [
        {
          id: '1',
          stream_id: 'test-stream',
          platform: 'twitch',
          author_name: 'TwitchUser',
          content: 'Twitch message',
          timestamp: Math.floor(Date.now() / 1000),
        },
      ]
      mockFetchStreamChat.mockResolvedValue(mockMessages)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        expect(screen.getByText('twitch')).toBeInTheDocument()
      })
    })
  })

  describe('sending messages', () => {
    it('should render message input and send button', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Send a message...')).toBeInTheDocument()
      })
      // Multiple buttons exist (refresh + send), check that buttons are present
      const buttons = document.querySelectorAll('button')
      expect(buttons.length).toBeGreaterThanOrEqual(2)
    })

    it('should disable send button when input is empty', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const sendButton = document.querySelector('button.btn-gradient')
        expect(sendButton).toBeDisabled()
      })
    })

    it('should enable send button when input has text', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        fireEvent.change(input, { target: { value: 'Hello' } })
      })

      const sendButton = document.querySelector('button.btn-gradient')
      expect(sendButton).not.toBeDisabled()
    })

    it('should call sendStreamChat when send button is clicked', async () => {
      mockFetchStreamChat.mockResolvedValue([])
      mockSendStreamChat.mockResolvedValue({} as api.ChatMessage)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        fireEvent.change(input, { target: { value: 'Test message' } })
      })

      const sendButton = document.querySelector('button.btn-gradient')!
      fireEvent.click(sendButton)

      await waitFor(() => {
        expect(mockSendStreamChat).toHaveBeenCalledWith('test-stream-123', { message: 'Test message' })
      })
    })

    it('should send message on Enter key press', async () => {
      mockFetchStreamChat.mockResolvedValue([])
      mockSendStreamChat.mockResolvedValue({} as api.ChatMessage)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        fireEvent.change(input, { target: { value: 'Test message' } })
        fireEvent.keyDown(input, { key: 'Enter' })
      })

      await waitFor(() => {
        expect(mockSendStreamChat).toHaveBeenCalled()
      })
    })

    it('should not send message on Shift+Enter', async () => {
      mockFetchStreamChat.mockResolvedValue([])
      mockSendStreamChat.mockResolvedValue({} as api.ChatMessage)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        fireEvent.change(input, { target: { value: 'Test message' } })
        fireEvent.keyDown(input, { key: 'Enter', shiftKey: true })
      })

      expect(mockSendStreamChat).not.toHaveBeenCalled()
    })

    it('should clear input after successful send', async () => {
      mockFetchStreamChat.mockResolvedValue([])
      mockSendStreamChat.mockResolvedValue({} as api.ChatMessage)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        fireEvent.change(input, { target: { value: 'Test message' } })
      })

      const sendButton = document.querySelector('button.btn-gradient')!
      fireEvent.click(sendButton)

      await waitFor(() => {
        const input = screen.getByPlaceholderText('Send a message...')
        expect(input).toHaveValue('')
      })
    })
  })

  describe('refresh', () => {
    it('should render refresh button', async () => {
      mockFetchStreamChat.mockResolvedValue([])

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        expect(screen.getByText('Live Chat')).toBeInTheDocument()
      })

      // Refresh button exists near the title
      const buttons = document.querySelectorAll('button')
      expect(buttons.length).toBeGreaterThan(0)
    })
  })

  describe('platform icons', () => {
    it('should render Twitch icon for twitch messages', async () => {
      const mockMessages: api.ChatMessage[] = [
        {
          id: '1',
          stream_id: 'test-stream',
          platform: 'twitch',
          author_name: 'User',
          content: 'Message',
          timestamp: Math.floor(Date.now() / 1000),
        },
      ]
      mockFetchStreamChat.mockResolvedValue(mockMessages)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        // Check for purple text class that indicates Twitch
        expect(document.querySelector('.text-purple-400')).toBeInTheDocument()
      })
    })

    it('should render YouTube icon for youtube messages', async () => {
      const mockMessages: api.ChatMessage[] = [
        {
          id: '1',
          stream_id: 'test-stream',
          platform: 'youtube',
          author_name: 'User',
          content: 'Message',
          timestamp: Math.floor(Date.now() / 1000),
        },
      ]
      mockFetchStreamChat.mockResolvedValue(mockMessages)

      renderWithQueryClient(<ChatWidget streamId="test-stream-123" />)

      await waitFor(() => {
        // Check for red text class that indicates YouTube
        expect(document.querySelector('.text-red-400')).toBeInTheDocument()
      })
    })
  })
})
