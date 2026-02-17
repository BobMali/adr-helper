export type StatusColorFamily = 'green' | 'amber' | 'red'

export function statusColorFamily(status: string): StatusColorFamily {
  const s = status.toLowerCase()
  if (s === 'accepted') return 'green'
  if (s === 'proposed') return 'amber'
  return 'red'
}

const DOT_CLASS: Record<StatusColorFamily, string> = {
  green: 'bg-green-500',
  amber: 'bg-amber-500',
  red: 'bg-red-500',
}

const TEXT_CLASS: Record<StatusColorFamily, string> = {
  green: 'text-green-600 dark:text-green-400',
  amber: 'text-amber-600 dark:text-amber-400',
  red: 'text-red-600 dark:text-red-400',
}

const CHIP_BG_CLASS: Record<StatusColorFamily, string> = {
  green: 'bg-green-600 hover:bg-green-700',
  amber: 'bg-amber-600 hover:bg-amber-700',
  red: 'bg-red-700 hover:bg-red-800',
}

export function statusDotClass(status: string): string {
  return DOT_CLASS[statusColorFamily(status)]
}

export function statusTextClass(status: string): string {
  return TEXT_CLASS[statusColorFamily(status)]
}

export function chipBgClass(status: string): string {
  return CHIP_BG_CLASS[statusColorFamily(status)]
}
