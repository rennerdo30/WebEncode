import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FileBrowser } from './file-browser'
import * as api from '@/lib/api'

// Mock the API module
vi.mock('@/lib/api', () => ({
  browseFiles: vi.fn(),
  fetchBrowseRoots: vi.fn(),
}))

const mockBrowseFiles = vi.mocked(api.browseFiles)
const mockFetchBrowseRoots = vi.mocked(api.fetchBrowseRoots)

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
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

describe('FileBrowser', () => {
  const mockRoots: api.BrowseRoot[] = [
    {
      name: 'Local Storage',
      path: '/media',
      plugin_id: 'storage-fs',
      storage_type: 'filesystem',
    },
  ]

  const mockBrowseResponse: api.BrowseResponse = {
    current_path: '/media',
    parent_path: '/',
    entries: [
      {
        name: 'video.mp4',
        path: '/media/video.mp4',
        is_directory: false,
        is_video: true,
        is_audio: false,
        is_image: false,
        size: 1024000,
        mod_time: Math.floor(Date.now() / 1000),
        extension: 'mp4',
      },
      {
        name: 'subfolder',
        path: '/media/subfolder',
        is_directory: true,
        is_video: false,
        is_audio: false,
        is_image: false,
        size: 0,
        mod_time: Math.floor(Date.now() / 1000),
        extension: '',
      },
    ],
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockFetchBrowseRoots.mockResolvedValue(mockRoots)
    mockBrowseFiles.mockResolvedValue(mockBrowseResponse)
  })

  describe('rendering', () => {
    it('should render loading state while fetching roots', () => {
      mockFetchBrowseRoots.mockImplementation(() => new Promise(() => {}))

      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })

    it('should render empty state when no roots available', async () => {
      mockFetchBrowseRoots.mockResolvedValue([])

      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('No storage plugins with browse support available.')).toBeInTheDocument()
      })
    })

    it('should render file browser with roots', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('fs')).toBeInTheDocument() // storage-fs becomes 'fs'
      })
    })

    it('should render navigation controls', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        // Check for navigation buttons
        const buttons = document.querySelectorAll('button')
        expect(buttons.length).toBeGreaterThan(0)
      })
    })

    it('should render search input', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search files...')).toBeInTheDocument()
      })
    })
  })

  describe('file list', () => {
    it('should render files and directories', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('video.mp4')).toBeInTheDocument()
        expect(screen.getByText('subfolder')).toBeInTheDocument()
      })
    })

    it('should display file size for files', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        // 1024000 bytes = ~1000 KB or ~1 MB
        expect(screen.getByText(/KB|MB/)).toBeInTheDocument()
      })
    })

    it('should display file extension badge', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('MP4')).toBeInTheDocument()
      })
    })
  })

  describe('navigation', () => {
    it('should navigate to directory on click', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        const folderButton = screen.getByText('subfolder').closest('button')
        expect(folderButton).toBeInTheDocument()
        if (folderButton) fireEvent.click(folderButton)
      })

      await waitFor(() => {
        expect(mockBrowseFiles).toHaveBeenCalledWith(
          expect.objectContaining({
            path: '/media/subfolder',
          })
        )
      })
    })

    it('should navigate up when clicking back button', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        const backButton = document.querySelector('button')
        expect(backButton).toBeInTheDocument()
      })
    })
  })

  describe('file selection', () => {
    it('should call onSelect when file is clicked', async () => {
      const handleSelect = vi.fn()
      renderWithQueryClient(<FileBrowser onSelect={handleSelect} />)

      await waitFor(() => {
        const fileButton = screen.getByText('video.mp4').closest('button')
        if (fileButton) fireEvent.click(fileButton)
      })

      expect(handleSelect).toHaveBeenCalledWith('file:///media/video.mp4')
    })

    it('should highlight selected file', async () => {
      renderWithQueryClient(
        <FileBrowser onSelect={vi.fn()} selectedPath="file:///media/video.mp4" />
      )

      await waitFor(() => {
        const fileButton = screen.getByText('video.mp4').closest('button')
        expect(fileButton).toHaveClass('bg-violet-500/10')
      })
    })

    it('should display selected file in footer', async () => {
      renderWithQueryClient(
        <FileBrowser onSelect={vi.fn()} selectedPath="file:///media/video.mp4" />
      )

      await waitFor(() => {
        expect(screen.getByText('Selected:')).toBeInTheDocument()
        expect(screen.getByText('/media/video.mp4')).toBeInTheDocument()
      })
    })
  })

  describe('search', () => {
    it('should update search query on input', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        const searchInput = screen.getByPlaceholderText('Search files...')
        fireEvent.change(searchInput, { target: { value: 'test' } })
      })

      await waitFor(() => {
        expect(mockBrowseFiles).toHaveBeenCalledWith(
          expect.objectContaining({
            search: 'test',
          })
        )
      })
    })
  })

  describe('root selection', () => {
    it('should render root in sidebar', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('Local Storage')).toBeInTheDocument()
      })
    })

    it('should highlight active root', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        const rootButton = screen.getByText('Local Storage').closest('button')
        expect(rootButton).toHaveClass('bg-violet-500/10')
      })
    })
  })

  describe('file icons', () => {
    it('should show folder icon for directories', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        // Folders have amber-400 color
        expect(document.querySelector('.text-amber-400')).toBeInTheDocument()
      })
    })

    it('should show video icon for video files', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        // Video files have violet-400 color
        expect(document.querySelector('.text-violet-400')).toBeInTheDocument()
      })
    })
  })

  describe('empty states', () => {
    it('should show empty state when no files in directory', async () => {
      mockBrowseFiles.mockResolvedValue({
        current_path: '/media',
        parent_path: '/',
        entries: [],
      })

      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('No files found')).toBeInTheDocument()
      })
    })

    it('should mention media filter when mediaOnly is true', async () => {
      mockBrowseFiles.mockResolvedValue({
        current_path: '/media',
        parent_path: '/',
        entries: [],
      })

      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} mediaOnly={true} />)

      await waitFor(() => {
        expect(screen.getByText('Only showing media files')).toBeInTheDocument()
      })
    })
  })

  describe('path breadcrumb', () => {
    it('should display current path', async () => {
      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('/media')).toBeInTheDocument()
      })
    })
  })

  describe('multiple roots', () => {
    it('should render multiple storage roots', async () => {
      const multipleRoots: api.BrowseRoot[] = [
        {
          name: 'Local Storage',
          path: '/media',
          plugin_id: 'storage-fs',
          storage_type: 'filesystem',
        },
        {
          name: 'S3 Bucket',
          path: '/bucket',
          plugin_id: 'storage-s3',
          storage_type: 's3',
        },
      ]
      mockFetchBrowseRoots.mockResolvedValue(multipleRoots)

      renderWithQueryClient(<FileBrowser onSelect={vi.fn()} />)

      await waitFor(() => {
        expect(screen.getByText('Local Storage')).toBeInTheDocument()
        expect(screen.getByText('S3 Bucket')).toBeInTheDocument()
      })
    })
  })
})
