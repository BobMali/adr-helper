import { withSetup } from './testHelper'
import { useADRSearch } from './useADRSearch'
import { fetchADRs } from '../api'

vi.mock('../api', () => ({
  fetchADRs: vi.fn(),
}))

const mockedFetchADRs = fetchADRs as ReturnType<typeof vi.fn>

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useADRSearch', () => {
  it('initializes with loading=true and empty state', () => {
    const [result] = withSetup(() => useADRSearch())
    expect(result.loading.value).toBe(true)
    expect(result.adrs.value).toEqual([])
    expect(result.error.value).toBe('')
    expect(result.searchQuery.value).toBe('')
  })

  it('loadADRs fetches and populates adrs', async () => {
    const data = [{ number: 1, title: 'ADR 1', status: 'Accepted', date: '2025-01-01' }]
    mockedFetchADRs.mockResolvedValue(data)

    const [result] = withSetup(() => useADRSearch())
    await result.loadADRs()

    expect(result.adrs.value).toEqual(data)
    expect(result.loading.value).toBe(false)
    expect(result.error.value).toBe('')
  })

  it('loadADRs passes query to fetchADRs', async () => {
    mockedFetchADRs.mockResolvedValue([])

    const [result] = withSetup(() => useADRSearch())
    await result.loadADRs('test')

    expect(mockedFetchADRs).toHaveBeenCalledWith('test', expect.any(AbortSignal))
  })

  it('loadADRs sets error on failure', async () => {
    mockedFetchADRs.mockRejectedValue(new Error('Network down'))

    const [result] = withSetup(() => useADRSearch())
    await result.loadADRs()

    expect(result.error.value).toBe('Network down')
    expect(result.loading.value).toBe(false)
  })

  it('loadADRs ignores AbortError', async () => {
    mockedFetchADRs.mockRejectedValue(new DOMException('aborted', 'AbortError'))

    const [result] = withSetup(() => useADRSearch())
    await result.loadADRs()

    expect(result.error.value).toBe('')
  })

  it('hasSearchQuery is true when query has 1+ non-whitespace chars', () => {
    const [result] = withSetup(() => useADRSearch())
    expect(result.hasSearchQuery.value).toBe(false)

    result.searchQuery.value = '  '
    expect(result.hasSearchQuery.value).toBe(false)

    result.searchQuery.value = 'a'
    expect(result.hasSearchQuery.value).toBe(true)
  })

  describe('onSearchInput', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('debounces and calls loadADRs with query after 300ms', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result] = withSetup(() => useADRSearch())
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue([])

      result.searchQuery.value = 'test'
      result.onSearchInput()

      expect(mockedFetchADRs).not.toHaveBeenCalled()

      await vi.advanceTimersByTimeAsync(300)

      expect(mockedFetchADRs).toHaveBeenCalledWith('test', expect.any(AbortSignal))
    })

    it('calls loadADRs with undefined for short queries', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result] = withSetup(() => useADRSearch())
      result.searchQuery.value = 'a'
      result.onSearchInput()

      await vi.advanceTimersByTimeAsync(300)

      expect(mockedFetchADRs).toHaveBeenCalledWith(undefined, expect.any(AbortSignal))
    })

    it('clears error immediately on input', () => {
      const [result] = withSetup(() => useADRSearch())
      result.error.value = 'some error'
      result.onSearchInput()

      expect(result.error.value).toBe('')
    })

    it('debounces rapid input (only final value triggers fetch)', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result] = withSetup(() => useADRSearch())
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue([])

      result.searchQuery.value = 'ch'
      result.onSearchInput()
      await vi.advanceTimersByTimeAsync(100)

      result.searchQuery.value = 'chi'
      result.onSearchInput()
      await vi.advanceTimersByTimeAsync(300)

      expect(mockedFetchADRs).toHaveBeenCalledTimes(1)
      expect(mockedFetchADRs).toHaveBeenCalledWith('chi', expect.any(AbortSignal))
    })

    it('cancels previous request when new search is triggered', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result] = withSetup(() => useADRSearch())
      mockedFetchADRs.mockReset()

      const staleData = [{ number: 1, title: 'Stale', status: 'Accepted', date: '2025-01-01' }]
      let resolveFirst!: (v: unknown) => void
      mockedFetchADRs.mockImplementationOnce(() => new Promise(r => { resolveFirst = r }))

      result.searchQuery.value = 'abc'
      result.onSearchInput()
      await vi.advanceTimersByTimeAsync(300)

      const freshData = [{ number: 2, title: 'Fresh', status: 'Proposed', date: '2025-02-01' }]
      mockedFetchADRs.mockResolvedValueOnce(freshData)

      result.searchQuery.value = 'xyz'
      result.onSearchInput()
      await vi.advanceTimersByTimeAsync(300)

      resolveFirst(staleData)
      await vi.advanceTimersByTimeAsync(0)

      expect(result.adrs.value).toEqual(freshData)
    })

    it('does not fire request if unmounted before debounce completes', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result, app] = withSetup(() => useADRSearch())
      mockedFetchADRs.mockClear()

      result.searchQuery.value = 'abc'
      result.onSearchInput()

      app.unmount()

      await vi.advanceTimersByTimeAsync(300)

      expect(mockedFetchADRs).not.toHaveBeenCalled()
    })
  })

  describe('clearSearch', () => {
    it('resets query and reloads', async () => {
      mockedFetchADRs.mockResolvedValue([])

      const [result] = withSetup(() => useADRSearch())
      result.searchQuery.value = 'test'
      await result.clearSearch()

      expect(result.searchQuery.value).toBe('')
      expect(mockedFetchADRs).toHaveBeenCalledWith(undefined, expect.any(AbortSignal))
    })
  })
})
