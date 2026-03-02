import { useRelation } from './useRelation'
import { withSetup } from './testHelper'
import type { ADRDetail } from '../types'

vi.mock('../api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api')>()
  return {
    ...actual,
    addRelation: vi.fn(),
  }
})

import { addRelation } from '../api'

const mockedAddRelation = addRelation as ReturnType<typeof vi.fn>

const sampleDetail: ADRDetail = {
  number: 1,
  title: 'Use Go',
  status: 'Accepted',
  date: '2025-01-15',
  content: '## Relations\n\nRelates to [ADR-0003](0003-use-chi.md)',
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
  vi.useRealTimers()
})

describe('useRelation', () => {
  it('has correct initial state', () => {
    const [{ adding, feedback, feedbackType }] = withSetup(() => useRelation(1))

    expect(adding.value).toBe(false)
    expect(feedback.value).toBe('')
    expect(feedbackType.value).toBe('')
  })

  it('confirmRelation success calls API and returns updated ADR', async () => {
    mockedAddRelation.mockResolvedValue(sampleDetail)

    const [{ confirmRelation, adding, feedback, feedbackType }] = withSetup(() => useRelation(1))

    const result = await confirmRelation(3)

    expect(mockedAddRelation).toHaveBeenCalledWith(1, 3)
    expect(result).toEqual(sampleDetail)
    expect(adding.value).toBe(false)
    expect(feedback.value).toContain('Relation added')
    expect(feedbackType.value).toBe('success')
  })

  it('confirmRelation error returns null and sets error feedback', async () => {
    mockedAddRelation.mockRejectedValue(new Error('Network error'))

    const [{ confirmRelation, feedback, feedbackType }] = withSetup(() => useRelation(1))

    const result = await confirmRelation(3)

    expect(result).toBeNull()
    expect(feedback.value).toContain('Network error')
    expect(feedbackType.value).toBe('error')
  })

  it('feedback auto-clears after 4 seconds', async () => {
    vi.useFakeTimers()
    mockedAddRelation.mockResolvedValue(sampleDetail)

    const [{ confirmRelation, feedback }] = withSetup(() => useRelation(1))

    await confirmRelation(3)
    expect(feedback.value).not.toBe('')

    vi.advanceTimersByTime(4000)

    expect(feedback.value).toBe('')
  })

  it('concurrent call guard: calling while adding is true', async () => {
    let resolve!: (value: ADRDetail) => void
    mockedAddRelation.mockReturnValue(new Promise<ADRDetail>((r) => { resolve = r }))

    const [{ confirmRelation, adding }] = withSetup(() => useRelation(1))

    // First call
    const promise1 = confirmRelation(3)
    expect(adding.value).toBe(true)

    // Second call while first is in-flight
    const result2 = await confirmRelation(5)
    expect(result2).toBeNull()
    expect(mockedAddRelation).toHaveBeenCalledTimes(1)

    resolve(sampleDetail)
    await promise1
  })

  it('clears feedback timer on unmount', async () => {
    vi.useFakeTimers()
    mockedAddRelation.mockResolvedValue(sampleDetail)

    const [result, app] = withSetup(() => useRelation(1))
    await result.confirmRelation(3)

    const feedbackAfterCall = result.feedback.value
    expect(feedbackAfterCall).not.toBe('')

    app.unmount()
    vi.advanceTimersByTime(4000)

    // Timer was cancelled on unmount — setTimeout callback never ran,
    // so feedback stays frozen at its post-call value
    expect(result.feedback.value).toBe(feedbackAfterCall)
  })
})
