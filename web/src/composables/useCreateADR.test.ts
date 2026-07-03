import { useCreateADR } from './useCreateADR'
import { withSetup } from './testHelper'
import type { ADRDetail, TemplateSectionDef } from '../types'

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

const sampleSectionDefs: TemplateSectionDef[] = [
  { key: 'context', heading: 'Context', kind: 'h2', optional: false, placeholder: 'Some text' },
  { key: 'decision', heading: 'Decision', kind: 'h2', optional: false, placeholder: 'Some text' },
  { key: 'consequences', heading: 'Consequences', kind: 'h2', optional: true, placeholder: 'Some text' },
]

afterEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
})

describe('useCreateADR', () => {
  it('has correct initial state', () => {
    const [{ title, submitting, submitError, sections, sectionErrors }] = withSetup(() => useCreateADR())

    expect(title.value).toBe('')
    expect(submitting.value).toBe(false)
    expect(submitError.value).toBe('')
    expect(sections.value).toEqual({})
    expect(sectionErrors.value).toEqual({})
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

  it('includes sections in createADR call', async () => {
    mockedCreateADR.mockResolvedValue(sampleDetail)

    const [{ title, sections, submit }] = withSetup(() => useCreateADR())

    title.value = 'Use PostgreSQL'
    sections.value = { context: 'We need a DB.', decision: 'Use PostgreSQL.' }
    const result = await submit(sampleSectionDefs)

    expect(mockedCreateADR).toHaveBeenCalledWith({
      title: 'Use PostgreSQL',
      sections: { context: 'We need a DB.', decision: 'Use PostgreSQL.' },
    })
    expect(result).toEqual(sampleDetail)
  })

  it('required section empty sets sectionErrors and returns null', async () => {
    const [{ title, sections, submit, sectionErrors }] = withSetup(() => useCreateADR())

    title.value = 'Test'
    sections.value = { context: 'Some context' } // decision is missing (required)
    const result = await submit(sampleSectionDefs)

    expect(result).toBeNull()
    expect(sectionErrors.value['decision']).toBe('Decision is required')
    expect(mockedCreateADR).not.toHaveBeenCalled()
  })

  it('optional section empty is not included in payload and no error', async () => {
    mockedCreateADR.mockResolvedValue(sampleDetail)

    const [{ title, sections, submit, sectionErrors }] = withSetup(() => useCreateADR())

    title.value = 'Test'
    sections.value = { context: 'Some context', decision: 'A decision' }
    // consequences is optional and empty
    const result = await submit(sampleSectionDefs)

    expect(result).toEqual(sampleDetail)
    expect(sectionErrors.value).toEqual({})
    expect(mockedCreateADR).toHaveBeenCalledWith({
      title: 'Test',
      sections: { context: 'Some context', decision: 'A decision' },
    })
  })
})
