import { useCreateADR } from './useCreateADR'
import { withSetup } from './testHelper'
import type { ADRDetail } from '../types'

vi.mock('../api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api')>()
  return {
    ...actual,
    createADR: vi.fn(),
  }
})

import { createADR } from '../api'

const mockedCreateADR = createADR as ReturnType<typeof vi.fn>

const sampleDetail: ADRDetail = {
  number: 3,
  title: 'Use PostgreSQL',
  status: 'Proposed',
  date: '2026-03-02',
  content: '# 3. Use PostgreSQL\n\n## Status\n\nProposed\n',
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
})

describe('useCreateADR', () => {
  it('has correct initial state', () => {
    const [{ title, submitting, submitError }] = withSetup(() => useCreateADR())

    expect(title.value).toBe('')
    expect(submitting.value).toBe(false)
    expect(submitError.value).toBe('')
  })

  it('empty title sets error and returns null without calling API', async () => {
    const [{ title, submit, submitError }] = withSetup(() => useCreateADR())

    title.value = '   '
    const result = await submit()

    expect(result).toBeNull()
    expect(submitError.value).toBe('Title is required')
    expect(mockedCreateADR).not.toHaveBeenCalled()
  })

  it('valid title calls createADR with trimmed title and returns ADRDetail', async () => {
    mockedCreateADR.mockResolvedValue(sampleDetail)

    const [{ title, submit, submitError }] = withSetup(() => useCreateADR())

    title.value = '  Use PostgreSQL  '
    const result = await submit()

    expect(mockedCreateADR).toHaveBeenCalledWith({ title: 'Use PostgreSQL' })
    expect(result).toEqual(sampleDetail)
    expect(submitError.value).toBe('')
  })

  it('API error sets submitError and returns null', async () => {
    mockedCreateADR.mockRejectedValue(new Error('Server error'))

    const [{ title, submit, submitError }] = withSetup(() => useCreateADR())

    title.value = 'Something'
    const result = await submit()

    expect(result).toBeNull()
    expect(submitError.value).toBe('Server error')
  })

  it('submitting is true during API call', async () => {
    let resolve!: (value: ADRDetail) => void
    mockedCreateADR.mockReturnValue(new Promise<ADRDetail>((r) => { resolve = r }))

    const [{ title, submit, submitting }] = withSetup(() => useCreateADR())

    title.value = 'Test'
    const promise = submit()
    expect(submitting.value).toBe(true)

    resolve(sampleDetail)
    await promise
    expect(submitting.value).toBe(false)
  })
})
