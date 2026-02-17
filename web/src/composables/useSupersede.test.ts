import { withSetup } from './testHelper'
import { useSupersede } from './useSupersede'
import { fetchADRs } from '../api'

vi.mock('../api', () => ({
  fetchADRs: vi.fn(),
}))

const mockedFetchADRs = fetchADRs as ReturnType<typeof vi.fn>

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useSupersede', () => {
  it('initializes with idle state', () => {
    const [result] = withSetup(() => useSupersede(5))
    expect(result.pendingSuperseded.value).toBe(false)
    expect(result.supersededBy.value).toBeNull()
    expect(result.availableADRs.value).toEqual([])
    expect(result.loadingADRs.value).toBe(false)
  })

  it('startSupersede fetches ADRs and filters out current', async () => {
    const allADRs = [
      { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
      { number: 5, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      { number: 7, title: 'Use SQLite', status: 'Proposed', date: '2025-02-01' },
    ]
    mockedFetchADRs.mockResolvedValue(allADRs)

    const [result] = withSetup(() => useSupersede(5))
    await result.startSupersede()

    expect(result.pendingSuperseded.value).toBe(true)
    expect(result.availableADRs.value).toHaveLength(2)
    expect(result.availableADRs.value.map(a => a.number)).toEqual([3, 7])
    expect(result.loadingADRs.value).toBe(false)
  })

  it('startSupersede sets loadingADRs during fetch', async () => {
    mockedFetchADRs.mockReturnValue(new Promise(() => {}))

    const [result] = withSetup(() => useSupersede(5))
    result.startSupersede()

    expect(result.loadingADRs.value).toBe(true)
    expect(result.pendingSuperseded.value).toBe(true)
  })

  it('startSupersede handles fetch error gracefully', async () => {
    mockedFetchADRs.mockRejectedValue(new Error('Network error'))

    const [result] = withSetup(() => useSupersede(5))
    await result.startSupersede()

    expect(result.availableADRs.value).toEqual([])
    expect(result.loadingADRs.value).toBe(false)
  })

  it('startSupersede ignores AbortError', async () => {
    mockedFetchADRs.mockRejectedValue(new DOMException('aborted', 'AbortError'))

    const [result] = withSetup(() => useSupersede(5))
    await result.startSupersede()

    // Should not set loadingADRs to false (abort controller was replaced)
    // But in this case the controller is still the same, so loadingADRs won't be set
    // The important thing is no error is shown
    expect(result.availableADRs.value).toEqual([])
  })

  it('cancelSupersede resets state', async () => {
    mockedFetchADRs.mockResolvedValue([
      { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
    ])

    const [result] = withSetup(() => useSupersede(5))
    await result.startSupersede()
    result.cancelSupersede()

    expect(result.pendingSuperseded.value).toBe(false)
    expect(result.supersededBy.value).toBeNull()
    expect(result.availableADRs.value).toEqual([])
  })

  it('onSelectorKeydown calls cancel on Escape', () => {
    const [result] = withSetup(() => useSupersede(5))
    const event = new KeyboardEvent('keydown', { key: 'Escape' })
    const preventDefault = vi.spyOn(event, 'preventDefault')
    const onConfirm = vi.fn()

    result.onSelectorKeydown(event, onConfirm)

    expect(preventDefault).toHaveBeenCalled()
    expect(result.pendingSuperseded.value).toBe(false)
    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('onSelectorKeydown calls confirm on Enter when ADR selected', () => {
    const [result] = withSetup(() => useSupersede(5))
    result.supersededBy.value = 3
    const event = new KeyboardEvent('keydown', { key: 'Enter' })
    const preventDefault = vi.spyOn(event, 'preventDefault')
    const onConfirm = vi.fn()

    result.onSelectorKeydown(event, onConfirm)

    expect(preventDefault).toHaveBeenCalled()
    expect(onConfirm).toHaveBeenCalled()
  })

  it('onSelectorKeydown ignores Enter when no ADR selected', () => {
    const [result] = withSetup(() => useSupersede(5))
    const event = new KeyboardEvent('keydown', { key: 'Enter' })
    const onConfirm = vi.fn()

    result.onSelectorKeydown(event, onConfirm)

    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('cancels inflight fetch when cancelSupersede is called', async () => {
    let resolveFirst!: (v: unknown) => void
    mockedFetchADRs.mockImplementationOnce(
      () => new Promise(r => { resolveFirst = r }),
    )

    const [result] = withSetup(() => useSupersede(5))
    result.startSupersede()

    expect(result.loadingADRs.value).toBe(true)

    result.cancelSupersede()

    resolveFirst([
      { number: 3, title: 'Stale ADR', status: 'Accepted', date: '2025-01-01' },
    ])
    await new Promise(r => setTimeout(r, 0))

    expect(result.pendingSuperseded.value).toBe(false)
    expect(result.availableADRs.value).toEqual([])
  })
})
