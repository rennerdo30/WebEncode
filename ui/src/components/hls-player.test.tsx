import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'

// Mock hls.js before imports
vi.mock('hls.js', () => {
  const mockHlsInstance = {
    loadSource: vi.fn(),
    attachMedia: vi.fn(),
    on: vi.fn(),
    destroy: vi.fn(),
    recoverMediaError: vi.fn(),
  }

  const MockHls = vi.fn(() => mockHlsInstance) as unknown as {
    isSupported: () => boolean
    Events: Record<string, string>
    ErrorTypes: Record<string, string>
    new(): typeof mockHlsInstance
  }

  MockHls.isSupported = vi.fn().mockReturnValue(true)
  MockHls.Events = {
    MANIFEST_PARSED: 'hlsManifestParsed',
    ERROR: 'hlsError',
  }
  MockHls.ErrorTypes = {
    NETWORK_ERROR: 'networkError',
    MEDIA_ERROR: 'mediaError',
  }

  return {
    default: MockHls,
  }
})

import { HLSPlayer } from './hls-player'

describe('HLSPlayer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render video element', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video')
      expect(video).toBeInTheDocument()
    })

    it('should render with custom className', () => {
      const { container } = render(
        <HLSPlayer src="http://example.com/stream.m3u8" className="custom-class" />
      )

      expect(container.firstChild).toHaveClass('custom-class')
    })

    it('should display stream URL at bottom', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      expect(screen.getByText('http://example.com/stream.m3u8')).toBeInTheDocument()
    })

    it('should render video controls', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video')
      expect(video).toHaveAttribute('controls')
    })

    it('should render muted by default', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video') as HTMLVideoElement
      expect(video.muted).toBe(true)
    })

    it('should render not muted when muted=false', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" muted={false} />)

      const video = document.querySelector('video') as HTMLVideoElement
      expect(video.muted).toBe(false)
    })

    it('should have playsInline attribute', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video')
      expect(video).toHaveAttribute('playsinline')
    })
  })

  describe('loading state', () => {
    it('should show loading state initially', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      expect(screen.getByText('Connecting to stream...')).toBeInTheDocument()
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })
  })

  describe('base styling', () => {
    it('should have relative position', () => {
      const { container } = render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      expect(container.firstChild).toHaveClass('relative')
    })

    it('should have black background', () => {
      const { container } = render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      expect(container.firstChild).toHaveClass('bg-black')
    })

    it('should have full width and height', () => {
      const { container } = render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      expect(container.firstChild).toHaveClass('w-full')
      expect(container.firstChild).toHaveClass('h-full')
    })
  })

  describe('video styling', () => {
    it('should have video with object-contain', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video')
      expect(video).toHaveClass('object-contain')
    })

    it('should have video with full width and height', () => {
      render(<HLSPlayer src="http://example.com/stream.m3u8" />)

      const video = document.querySelector('video')
      expect(video).toHaveClass('w-full')
      expect(video).toHaveClass('h-full')
    })
  })
})
