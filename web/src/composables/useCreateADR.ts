import { ref } from 'vue'
import type { ADRDetail } from '../types'
import { createADR } from '../api'

export function useCreateADR() {
  const title = ref('')
  const submitting = ref(false)
  const submitError = ref('')

  async function submit(): Promise<ADRDetail | null> {
    submitError.value = ''

    const trimmed = title.value.trim()
    if (!trimmed) {
      submitError.value = 'Title is required'
      return null
    }

    submitting.value = true
    try {
      const result = await createADR({ title: trimmed })
      return result
    } catch (e) {
      submitError.value = e instanceof Error ? e.message : 'Failed to create ADR'
      return null
    } finally {
      submitting.value = false
    }
  }

  return {
    title,
    submitting,
    submitError,
    submit,
  }
}
