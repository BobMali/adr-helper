import type { ADRSummary, ADRDetail } from './types'

export async function fetchADRs(): Promise<ADRSummary[]> {
  const res = await fetch('/api/adr')
  if (!res.ok) {
    throw new Error(`Failed to fetch ADRs: ${res.status}`)
  }
  return res.json()
}

export async function fetchADR(number: number): Promise<ADRDetail> {
  const res = await fetch(`/api/adr/${number}`)
  if (res.status === 404) {
    throw new NotFoundError(`ADR #${number} not found`)
  }
  if (!res.ok) {
    throw new Error(`Failed to fetch ADR: ${res.status}`)
  }
  return res.json()
}

export async function fetchStatuses(): Promise<string[]> {
  const res = await fetch('/api/adr/statuses')
  if (!res.ok) {
    throw new Error(`Failed to fetch statuses: ${res.status}`)
  }
  return res.json()
}

export async function updateADRStatus(number: number, status: string): Promise<ADRDetail> {
  const res = await fetch(`/api/adr/${number}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (res.status === 404) {
    throw new NotFoundError(`ADR #${number} not found`)
  }
  if (!res.ok) {
    throw new Error(`Failed to update status: ${res.status}`)
  }
  return res.json()
}

export class NotFoundError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'NotFoundError'
  }
}
