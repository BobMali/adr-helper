import type { ADRSummary, ADRDetail } from './types'

async function apiFetch(url: string, init?: RequestInit): Promise<Response> {
  try {
    return init ? await fetch(url, init) : await fetch(url)
  } catch (err) {
    if (err instanceof TypeError) {
      throw new Error('Network error: unable to reach server')
    }
    throw err
  }
}

export async function fetchADRs(): Promise<ADRSummary[]> {
  const res = await apiFetch('/api/adr')
  if (!res.ok) {
    throw new Error(`Failed to fetch ADRs: ${res.status}`)
  }
  return res.json()
}

export async function fetchADR(number: number): Promise<ADRDetail> {
  const res = await apiFetch(`/api/adr/${number}`)
  if (res.status === 404) {
    throw new NotFoundError(`ADR #${number} not found`)
  }
  if (!res.ok) {
    throw new Error(`Failed to fetch ADR: ${res.status}`)
  }
  return res.json()
}

export async function fetchStatuses(): Promise<string[]> {
  const res = await apiFetch('/api/adr/statuses')
  if (!res.ok) {
    throw new Error(`Failed to fetch statuses: ${res.status}`)
  }
  return res.json()
}

interface UpdateStatusPayload {
  status: string
  supersededBy?: number
}

export async function updateADRStatus(
  number: number,
  status: string,
  options?: { supersededBy?: number },
): Promise<ADRDetail> {
  const payload: UpdateStatusPayload = { status }
  if (options?.supersededBy != null) {
    payload.supersededBy = options.supersededBy
  }
  const res = await apiFetch(`/api/adr/${number}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
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
