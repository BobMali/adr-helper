import { fetchADRs, fetchADR, fetchStatuses, updateADRStatus, NotFoundError } from './api'

function mockFetchOk(body: unknown, status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      status,
      json: () => Promise.resolve(body),
    }),
  )
}

function mockFetchFail(status: number) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: false,
      status,
      json: () => Promise.resolve({}),
    }),
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NotFoundError', () => {
  it('is an instance of Error', () => {
    const err = new NotFoundError('gone')
    expect(err).toBeInstanceOf(Error)
  })

  it('has name "NotFoundError"', () => {
    const err = new NotFoundError('gone')
    expect(err.name).toBe('NotFoundError')
  })
})

describe('fetchADRs', () => {
  it('GETs /api/adr and returns ADRSummary[]', async () => {
    const data = [{ number: 1, title: 'Use X', status: 'Accepted', date: '2025-01-01' }]
    mockFetchOk(data)

    const result = await fetchADRs()

    expect(fetch).toHaveBeenCalledWith('/api/adr')
    expect(result).toEqual(data)
  })

  it('throws on non-ok response', async () => {
    mockFetchFail(500)

    await expect(fetchADRs()).rejects.toThrow('Failed to fetch ADRs: 500')
  })

  it('wraps network TypeError with user-friendly message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(fetchADRs()).rejects.toThrow('Network error: unable to reach server')
  })
})

describe('fetchADR', () => {
  it('GETs /api/adr/{n} and returns ADRDetail', async () => {
    const data = { number: 3, title: 'Use Y', status: 'Proposed', date: '2025-02-01', content: '# Y' }
    mockFetchOk(data)

    const result = await fetchADR(3)

    expect(fetch).toHaveBeenCalledWith('/api/adr/3')
    expect(result).toEqual(data)
  })

  it('throws NotFoundError on 404', async () => {
    mockFetchFail(404)

    await expect(fetchADR(99)).rejects.toThrow(NotFoundError)
    await expect(fetchADR(99)).rejects.toThrow('ADR #99 not found')
  })

  it('throws generic Error on other failures', async () => {
    mockFetchFail(503)

    await expect(fetchADR(1)).rejects.toThrow('Failed to fetch ADR: 503')
  })

  it('wraps network TypeError with user-friendly message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(fetchADR(3)).rejects.toThrow('Network error: unable to reach server')
  })
})

describe('fetchStatuses', () => {
  it('GETs /api/adr/statuses and returns string[]', async () => {
    const data = ['Proposed', 'Accepted', 'Superseded']
    mockFetchOk(data)

    const result = await fetchStatuses()

    expect(fetch).toHaveBeenCalledWith('/api/adr/statuses')
    expect(result).toEqual(data)
  })

  it('throws on non-ok response', async () => {
    mockFetchFail(500)

    await expect(fetchStatuses()).rejects.toThrow('Failed to fetch statuses: 500')
  })

  it('wraps network TypeError with user-friendly message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(fetchStatuses()).rejects.toThrow('Network error: unable to reach server')
  })
})

describe('updateADRStatus', () => {
  it('PATCHes with JSON body containing status', async () => {
    const data = { number: 1, title: 'Use X', status: 'Accepted', date: '2025-01-01', content: '# X' }
    mockFetchOk(data)

    await updateADRStatus(1, 'Accepted')

    expect(fetch).toHaveBeenCalledWith('/api/adr/1/status', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: 'Accepted' }),
    })
  })

  it('includes supersededBy in body when provided', async () => {
    const data = { number: 1, title: 'Use X', status: 'Superseded', date: '2025-01-01', content: '# X' }
    mockFetchOk(data)

    await updateADRStatus(1, 'Superseded', { supersededBy: 5 })

    expect(fetch).toHaveBeenCalledWith('/api/adr/1/status', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: 'Superseded', supersededBy: 5 }),
    })
  })

  it('omits supersededBy when absent', async () => {
    const data = { number: 1, title: 'Use X', status: 'Accepted', date: '2025-01-01', content: '# X' }
    mockFetchOk(data)

    await updateADRStatus(1, 'Accepted')

    const body = JSON.parse((fetch as ReturnType<typeof vi.fn>).mock.calls[0]![1].body)
    expect(body).not.toHaveProperty('supersededBy')
  })

  it('throws NotFoundError on 404', async () => {
    mockFetchFail(404)

    await expect(updateADRStatus(99, 'Accepted')).rejects.toThrow(NotFoundError)
  })

  it('throws generic Error on other failures', async () => {
    mockFetchFail(503)

    await expect(updateADRStatus(1, 'Accepted')).rejects.toThrow('Failed to update status: 503')
  })

  it('wraps network TypeError with user-friendly message', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(updateADRStatus(1, 'Accepted')).rejects.toThrow('Network error: unable to reach server')
  })
})
