import { withSetup } from './testHelper'
import { useStatusUpdate } from './useStatusUpdate'
import { updateADRStatus } from '../api'

vi.mock('../api', () => ({
  updateADRStatus: vi.fn(),
}))

const mockedUpdateADRStatus = updateADRStatus as ReturnType<typeof vi.fn>

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useStatusUpdate', () => {
  it('initializes with idle state', () => {
    const [result] = withSetup(() => useStatusUpdate(5))
    expect(result.updating.value).toBe(false)
    expect(result.feedback.value).toBe('')
    expect(result.feedbackType.value).toBe('success')
  })

  it('doStatusUpdate calls API and returns updated ADR', async () => {
    const updated = { number: 5, title: 'Test', status: 'Deprecated', date: '2025-01-01', content: '' }
    mockedUpdateADRStatus.mockResolvedValue(updated)

    const [result] = withSetup(() => useStatusUpdate(5))
    const returned = await result.doStatusUpdate('Deprecated')

    expect(mockedUpdateADRStatus).toHaveBeenCalledWith(5, 'Deprecated', undefined)
    expect(returned).toEqual(updated)
    expect(result.feedback.value).toContain('Deprecated')
    expect(result.feedbackType.value).toBe('success')
    expect(result.updating.value).toBe(false)
  })

  it('doStatusUpdate passes options to API', async () => {
    const updated = { number: 5, title: 'Test', status: 'Superseded', date: '2025-01-01', content: '' }
    mockedUpdateADRStatus.mockResolvedValue(updated)

    const [result] = withSetup(() => useStatusUpdate(5))
    await result.doStatusUpdate('Superseded', { supersededBy: 3 })

    expect(mockedUpdateADRStatus).toHaveBeenCalledWith(5, 'Superseded', { supersededBy: 3 })
  })

  it('doStatusUpdate returns null on error', async () => {
    mockedUpdateADRStatus.mockRejectedValue(new Error('Update failed'))

    const [result] = withSetup(() => useStatusUpdate(5))
    const returned = await result.doStatusUpdate('Deprecated')

    expect(returned).toBeNull()
    expect(result.feedback.value).toBe('Update failed')
    expect(result.feedbackType.value).toBe('error')
    expect(result.updating.value).toBe(false)
  })

  it('sets updating=true during API call', async () => {
    let resolveUpdate!: (v: unknown) => void
    mockedUpdateADRStatus.mockReturnValue(new Promise(r => { resolveUpdate = r }))

    const [result] = withSetup(() => useStatusUpdate(5))
    const promise = result.doStatusUpdate('Deprecated')

    expect(result.updating.value).toBe(true)

    resolveUpdate({ number: 5, title: 'Test', status: 'Deprecated', date: '2025-01-01', content: '' })
    await promise

    expect(result.updating.value).toBe(false)
  })

  it('tracks previousStatus', () => {
    const [result] = withSetup(() => useStatusUpdate(5))
    result.setPreviousStatus('Accepted')
    expect(result.getPreviousStatus()).toBe('Accepted')
  })

  it('doStatusUpdate updates previousStatus on success', async () => {
    const updated = { number: 5, title: 'Test', status: 'Deprecated', date: '2025-01-01', content: '' }
    mockedUpdateADRStatus.mockResolvedValue(updated)

    const [result] = withSetup(() => useStatusUpdate(5))
    result.setPreviousStatus('Accepted')
    await result.doStatusUpdate('Deprecated')

    expect(result.getPreviousStatus()).toBe('Deprecated')
  })

  it('feedback clears after 4 seconds', async () => {
    vi.useFakeTimers()
    const updated = { number: 5, title: 'Test', status: 'Deprecated', date: '2025-01-01', content: '' }
    mockedUpdateADRStatus.mockResolvedValue(updated)

    const [result] = withSetup(() => useStatusUpdate(5))
    await result.doStatusUpdate('Deprecated')

    expect(result.feedback.value).toContain('Deprecated')

    vi.advanceTimersByTime(4000)

    expect(result.feedback.value).toBe('')
    vi.useRealTimers()
  })

  it('clears feedback timer on unmount', async () => {
    vi.useFakeTimers()
    const updated = { number: 5, title: 'Test', status: 'Deprecated', date: '2025-01-01', content: '' }
    mockedUpdateADRStatus.mockResolvedValue(updated)

    const [result, app] = withSetup(() => useStatusUpdate(5))
    await result.doStatusUpdate('Deprecated')

    expect(result.feedback.value).toContain('Deprecated')

    app.unmount()

    // Advance past the 4s timer — should not throw
    vi.advanceTimersByTime(4000)

    vi.useRealTimers()
  })
})
