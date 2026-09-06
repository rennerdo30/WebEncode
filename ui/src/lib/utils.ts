import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Unit suffixes used when rendering byte counts. */
const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const
/** Binary step between two consecutive byte units. */
const BYTES_PER_UNIT = 1024
/** Number of decimals kept when rendering byte counts. */
const BYTE_FRACTION_DIGITS = 1
/** Rendered when the byte count is zero and no override is supplied. */
const DEFAULT_ZERO_LABEL = "0 B"

/**
 * Format a byte count using binary units (B, KB, MB, GB, TB).
 *
 * @param bytes Byte count to format. Non-positive and non-finite values render as `zeroLabel`.
 * @param zeroLabel Text used for a zero/unknown size.
 */
export function formatBytes(bytes: number, zeroLabel: string = DEFAULT_ZERO_LABEL): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return zeroLabel
  const exponent = Math.min(
    Math.floor(Math.log(bytes) / Math.log(BYTES_PER_UNIT)),
    BYTE_UNITS.length - 1,
  )
  const value = bytes / Math.pow(BYTES_PER_UNIT, exponent)
  return `${value.toFixed(BYTE_FRACTION_DIGITS)} ${BYTE_UNITS[exponent]}`
}
