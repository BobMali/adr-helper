import type { ADRSummary } from './types'

export async function fetchADRs(): Promise<ADRSummary[]> {
  const res = await fetch('/api/adr')
  if (!res.ok) {
    throw new Error(`Failed to fetch ADRs: ${res.status}`)
  }
  return res.json()
}
