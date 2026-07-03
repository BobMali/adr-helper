import type { ADRSummary, ADRDetail, CreateADRPayload, TemplateSectionDef } from './types'

async function apiFetch(url: string, init?: RequestInit): Promise<Response> {
  try {
    return init ? await fetch(url, init) : await fetch(url)
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw err
    }
    if (err instanceof TypeError) {
      throw new Error('Network error: unable to reach server')
    }
    throw err
  }
}

export async function fetchADRs(query?: string, signal?: AbortSignal): Promise<ADRSummary[]> {
  let url = '/api/adr'
  if (query) {
    url += `?q=${encodeURIComponent(query)}`
  }
  const init: RequestInit | undefined = signal ? { signal } : undefined
  const res = await apiFetch(url, init)
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

export async function addRelation(number: number, relatedTo: number): Promise<ADRDetail> {
  const res = await apiFetch(`/api/adr/${number}/relations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ relatedTo }),
  })
  if (res.status === 404) {
    throw new NotFoundError(`ADR #${number} not found`)
  }
  if (!res.ok) {
    throw new Error(`Failed to add relation: ${res.status}`)
  }
  return res.json()
}

export async function fetchTemplateSections(): Promise<TemplateSectionDef[]> {
  const res = await apiFetch('/api/template-sections')
  if (!res.ok) {
    throw new Error(`Failed to fetch template sections: ${res.status}`)
  }
  return res.json()
}

export async function updateADRContent(number: number, content: string): Promise<ADRDetail> {
  const res = await apiFetch(`/api/adr/${number}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content }),
  })
  if (res.status === 404) {
    throw new NotFoundError(`ADR #${number} not found`)
  }
  if (!res.ok) {
    throw new Error(`Failed to update content: ${res.status}`)
  }
  return res.json()
}

export async function fetchConfig(): Promise<{ template: string }> {
  const res = await apiFetch('/api/config')
  if (!res.ok) {
    throw new Error(`Failed to fetch config: ${res.status}`)
  }
  return res.json()
}

export async function createADR(payload: CreateADRPayload): Promise<ADRDetail> {
  const res = await apiFetch('/api/adr', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  if (res.status === 409) {
    throw new ConflictError('ADR already exists')
  }
  if (!res.ok) {
    throw new Error(`Failed to create ADR: ${res.status}`)
  }
  return res.json()
}

export class NotFoundError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'NotFoundError'
  }
}

export class ConflictError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'ConflictError'
  }
}
