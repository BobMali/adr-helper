import { ref } from 'vue'
import type { ADRDetail } from '../types'
import { updateADRContent } from '../api'

export type EditState = 'idle' | 'confirming' | 'editing' | 'saving'

export function useEditContent() {
  const editState = ref<EditState>('idle')
  const editedContent = ref('')
  const saveError = ref('')

  function requestEdit() {
    editState.value = 'confirming'
    saveError.value = ''
  }

  function confirmEdit(currentContent: string) {
    editedContent.value = currentContent
    editState.value = 'editing'
    saveError.value = ''
  }

  function cancelEdit() {
    editState.value = 'idle'
    editedContent.value = ''
    saveError.value = ''
  }

  async function saveEdit(number: number): Promise<ADRDetail | null> {
    editState.value = 'saving'
    saveError.value = ''
    try {
      const result = await updateADRContent(number, editedContent.value)
      editState.value = 'idle'
      editedContent.value = ''
      return result
    } catch (e) {
      saveError.value = e instanceof Error ? e.message : 'Failed to save'
      editState.value = 'editing'
      return null
    }
  }

  return {
    editState,
    editedContent,
    saveError,
    requestEdit,
    confirmEdit,
    cancelEdit,
    saveEdit,
  }
}
