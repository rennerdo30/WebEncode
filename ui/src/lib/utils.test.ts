import { describe, it, expect } from 'vitest'
import { cn, formatBytes } from './utils'

describe('cn utility function', () => {
  it('should merge class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('should handle conditional classes', () => {
    expect(cn('base', true && 'active', false && 'hidden')).toBe('base active')
  })

  it('should merge tailwind classes correctly', () => {
    // twMerge should dedupe conflicting classes
    expect(cn('px-2', 'px-4')).toBe('px-4')
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
  })

  it('should handle arrays of classes', () => {
    expect(cn(['foo', 'bar'], 'baz')).toBe('foo bar baz')
  })

  it('should handle undefined and null', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })

  it('should handle empty input', () => {
    expect(cn()).toBe('')
  })

  it('should handle object syntax', () => {
    expect(cn({ active: true, disabled: false })).toBe('active')
  })
})

describe('formatBytes utility function', () => {
  it('should format byte counts using binary units', () => {
    expect(formatBytes(512)).toBe('512.0 B')
    expect(formatBytes(1024)).toBe('1.0 KB')
    expect(formatBytes(1024 * 1024)).toBe('1.0 MB')
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1.0 GB')
    expect(formatBytes(10 * 1024 ** 4)).toBe('10.0 TB')
  })

  it('should clamp values larger than the biggest known unit', () => {
    expect(formatBytes(1024 ** 5)).toBe('1024.0 TB')
  })

  it('should render the zero label for empty, negative and invalid sizes', () => {
    expect(formatBytes(0)).toBe('0 B')
    expect(formatBytes(-1)).toBe('0 B')
    expect(formatBytes(Number.NaN)).toBe('0 B')
    expect(formatBytes(0, '\u2014')).toBe('\u2014')
  })
})
