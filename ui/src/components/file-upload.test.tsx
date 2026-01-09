import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { FileUpload } from './file-upload'
import * as api from '@/lib/api'

// Mock the API module
vi.mock('@/lib/api', () => ({
  uploadFile: vi.fn(),
  reportError: vi.fn(),
}))

const mockUploadFile = vi.mocked(api.uploadFile)
const mockReportError = vi.mocked(api.reportError)

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

function createMockFile(
  name: string,
  size: number,
  type: string
): File {
  const file = new File([''], name, { type })
  Object.defineProperty(file, 'size', { value: size })
  return file
}

describe('FileUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render drop zone in idle state', () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      expect(screen.getByText('Drag and drop your video file')).toBeInTheDocument()
      expect(screen.getByText(/or click to browse/)).toBeInTheDocument()
    })

    it('should display max file size', () => {
      renderWithQueryClient(
        <FileUpload onUploadComplete={vi.fn()} maxSize={1024 * 1024 * 1024} />
      )

      expect(screen.getByText(/Max 1.0 GB/)).toBeInTheDocument()
    })

    it('should display supported file types', () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      expect(screen.getByText('MP4')).toBeInTheDocument()
      expect(screen.getByText('MKV')).toBeInTheDocument()
      expect(screen.getByText('AVI')).toBeInTheDocument()
      expect(screen.getByText('MOV')).toBeInTheDocument()
      expect(screen.getByText('WebM')).toBeInTheDocument()
      expect(screen.getByText('MP3')).toBeInTheDocument()
      expect(screen.getByText('WAV')).toBeInTheDocument()
    })

    it('should render hidden file input', () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const input = document.getElementById('file-upload-input')
      expect(input).toBeInTheDocument()
      expect(input).toHaveClass('hidden')
    })
  })

  describe('file selection', () => {
    it('should show selected file info when file is selected', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('test-video.mp4')).toBeInTheDocument()
        expect(screen.getByText('1.0 MB')).toBeInTheDocument()
        expect(screen.getByText('video/mp4')).toBeInTheDocument()
      })
    })

    it('should show upload button after file selection', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('Upload File')).toBeInTheDocument()
      })
    })

    it('should show "Choose Different File" button', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('Choose Different File')).toBeInTheDocument()
      })
    })

    it('should show remove button after file selection', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        // X button to remove file
        const buttons = screen.getAllByRole('button')
        expect(buttons.length).toBeGreaterThan(0)
      })
    })
  })

  describe('file validation', () => {
    it('should reject files that exceed max size', async () => {
      const maxSize = 1024 * 1024 // 1MB
      renderWithQueryClient(
        <FileUpload onUploadComplete={vi.fn()} maxSize={maxSize} />
      )

      const file = createMockFile('large-video.mp4', maxSize + 1, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('Upload Failed')).toBeInTheDocument()
        expect(screen.getByText(/File size exceeds maximum/)).toBeInTheDocument()
      })
    })

    it('should reject non-media files', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('document.pdf', 1024, 'application/pdf')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('Upload Failed')).toBeInTheDocument()
        expect(screen.getByText('Please select a video or audio file')).toBeInTheDocument()
      })
    })

    it('should accept video files', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('video.mp4', 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('video.mp4')).toBeInTheDocument()
        expect(screen.queryByText('Upload Failed')).not.toBeInTheDocument()
      })
    })

    it('should accept audio files', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('audio.mp3', 1024, 'audio/mpeg')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        expect(screen.getByText('audio.mp3')).toBeInTheDocument()
        expect(screen.queryByText('Upload Failed')).not.toBeInTheDocument()
      })
    })
  })

  describe('uploading', () => {
    it('should call uploadFile when upload button is clicked', async () => {
      mockUploadFile.mockResolvedValue({
        url: 'file:///path/to/file.mp4',
        filename: 'file.mp4',
        size: 1024,
      })

      const handleComplete = vi.fn()
      renderWithQueryClient(<FileUpload onUploadComplete={handleComplete} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        expect(mockUploadFile).toHaveBeenCalled()
      })
    })

    it('should show uploading state with progress', async () => {
      mockUploadFile.mockImplementation(() => new Promise(() => {})) // Never resolves

      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        expect(screen.getByText('0%')).toBeInTheDocument()
        expect(document.querySelector('.animate-spin')).toBeInTheDocument()
      })
    })

    it('should show success state after upload completes', async () => {
      const uploadResponse = {
        url: 'file:///path/to/file.mp4',
        filename: 'test-video.mp4',
        size: 1024 * 1024,
      }
      mockUploadFile.mockResolvedValue(uploadResponse)

      const handleComplete = vi.fn()
      renderWithQueryClient(<FileUpload onUploadComplete={handleComplete} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        expect(screen.getByText('Upload Complete')).toBeInTheDocument()
        expect(handleComplete).toHaveBeenCalledWith(uploadResponse.url, uploadResponse)
      })
    })

    it('should show error state on upload failure', async () => {
      mockUploadFile.mockRejectedValue(new Error('Network error'))

      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        expect(screen.getByText('Upload Failed')).toBeInTheDocument()
        expect(screen.getByText('Network error')).toBeInTheDocument()
      })
    })

    it('should call reportError on upload failure', async () => {
      mockUploadFile.mockRejectedValue(new Error('Network error'))

      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        expect(mockReportError).toHaveBeenCalledWith(
          'Network error',
          'component:file-upload',
          expect.any(String)
        )
      })
    })
  })

  describe('reset', () => {
    it('should reset to idle state when clicking "Try Again"', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('document.pdf', 1024, 'application/pdf')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Try Again'))
      })

      await waitFor(() => {
        expect(screen.getByText('Drag and drop your video file')).toBeInTheDocument()
      })
    })

    it('should reset to idle state when clicking "Upload Another"', async () => {
      mockUploadFile.mockResolvedValue({
        url: 'file:///path/to/file.mp4',
        filename: 'file.mp4',
        size: 1024,
      })

      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('test-video.mp4', 1024 * 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload File'))
      })

      await waitFor(() => {
        fireEvent.click(screen.getByText('Upload Another'))
      })

      await waitFor(() => {
        expect(screen.getByText('Drag and drop your video file')).toBeInTheDocument()
      })
    })
  })

  describe('drag and drop', () => {
    it('should highlight drop zone on drag over', () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const dropZone = screen.getByText('Drag and drop your video file').closest('div')!
      fireEvent.dragOver(dropZone)

      expect(screen.getByText('Drop your file here')).toBeInTheDocument()
    })

    it('should remove highlight on drag leave', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const dropZone = screen.getByText('Drag and drop your video file').closest('div')!
      fireEvent.dragOver(dropZone)
      fireEvent.dragLeave(dropZone)

      await waitFor(() => {
        expect(screen.getByText('Drag and drop your video file')).toBeInTheDocument()
      })
    })
  })

  describe('file icons', () => {
    it('should show video icon for video files', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('video.mp4', 1024, 'video/mp4')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        // Video files use violet-400 color
        expect(document.querySelector('.text-violet-400')).toBeInTheDocument()
      })
    })

    it('should show audio icon for audio files', async () => {
      renderWithQueryClient(<FileUpload onUploadComplete={vi.fn()} />)

      const file = createMockFile('audio.mp3', 1024, 'audio/mpeg')
      const input = document.getElementById('file-upload-input') as HTMLInputElement

      Object.defineProperty(input, 'files', { value: [file] })
      fireEvent.change(input)

      await waitFor(() => {
        // Audio files use cyan-400 color
        expect(document.querySelector('.text-cyan-400')).toBeInTheDocument()
      })
    })
  })
})
