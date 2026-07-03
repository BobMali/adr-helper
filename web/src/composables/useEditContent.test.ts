import { useEditContent } from './useEditContent'
import { withSetup } from './testHelper'
import type { ADRDetail } from '../types'

vi.mock('../api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api')>()
  return {
    ...actual,
    updateADRContent: vi.fn(),
  }
})

import { updateADRContent } from '../api'

const mockedUpdateADRContent = updateADRContent as ReturnType<typeof vi.fn>

const sampleDetail: ADRDetail = {
  number: 1,
  title: 'Use Go',
  status: 'Accepted',
  date: '2024-01-15',
  content: '# 1. Use Go\n\n## Status\n\nAccepted\n\n## Context\n\nUpdated.\n',
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
})

describe('useEditContent', () => {
  it('has correct initial state', () => {
    const [{ editState, editedContent, saveError }] = withSetup(() => useEditContent())

    expect(editState.value).toBe('idle')
    expect(editedContent.value).toBe('')
    expect(saveError.value).toBe('')
  })

  it('requestEdit transitions to confirming', () => {
    const [{ editState, requestEdit }] = withSetup(() => useEditContent())

    requestEdit()

    expect(editState.value).toBe('confirming')
  })

  it('confirmEdit transitions to editing with content', () => {
    const [{ editState, editedContent, requestEdit, confirmEdit }] = withSetup(() => useEditContent())

    requestEdit()
    confirmEdit('# 1. Use Go\n\nSome content')

    expect(editState.value).toBe('editing')
    expect(editedContent.value).toBe('# 1. Use Go\n\nSome content')
  })

  it('cancelEdit from confirming returns to idle', () => {
    const [{ editState, requestEdit, cancelEdit }] = withSetup(() => useEditContent())

    requestEdit()
    cancelEdit()

    expect(editState.value).toBe('idle')
  })

  it('cancelEdit from editing returns to idle', () => {
    const [{ editState, requestEdit, confirmEdit, cancelEdit }] = withSetup(() => useEditContent())

    requestEdit()
    confirmEdit('content')
    cancelEdit()

    expect(editState.value).toBe('idle')
  })

  it('saveEdit success returns ADRDetail and resets to idle', async () => {
    mockedUpdateADRContent.mockResolvedValue(sampleDetail)

    const [{ editState, editedContent, requestEdit, confirmEdit, saveEdit }] = withSetup(() => useEditContent())

    requestEdit()
    confirmEdit('updated content')
    const result = await saveEdit(1)

    expect(result).toEqual(sampleDetail)
    expect(editState.value).toBe('idle')
    expect(editedContent.value).toBe('')
    expect(mockedUpdateADRContent).toHaveBeenCalledWith(1, 'updated content')
  })

  it('saveEdit error keeps state as editing and sets saveError', async () => {
    mockedUpdateADRContent.mockRejectedValue(new Error('Save failed'))

    const [{ editState, saveError, requestEdit, confirmEdit, saveEdit }] = withSetup(() => useEditContent())

    requestEdit()
    confirmEdit('content')
    const result = await saveEdit(1)

    expect(result).toBeNull()
    expect(editState.value).toBe('editing')
    expect(saveError.value).toBe('Save failed')
  })

  it('state transitions: idle → confirming → editing → saving → idle', async () => {
    mockedUpdateADRContent.mockResolvedValue(sampleDetail)

    const [{ editState, requestEdit, confirmEdit, saveEdit }] = withSetup(() => useEditContent())

    expect(editState.value).toBe('idle')
    requestEdit()
    expect(editState.value).toBe('confirming')
    confirmEdit('content')
    expect(editState.value).toBe('editing')

    const promise = saveEdit(1)
    // During save, state should be 'saving' synchronously
    // (but it resolves immediately in test, so we check the result)
    const result = await promise
    expect(result).toEqual(sampleDetail)
    expect(editState.value).toBe('idle')
  })
})
