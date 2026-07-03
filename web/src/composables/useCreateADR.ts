import { ref } from 'vue'
import type { Ref } from 'vue'
import type { ADRDetail, TemplateSectionDef } from '../types'
import { createADR } from '../api'

export function useCreateADR() {
  const title = ref('')
  const sections: Ref<Record<string, string>> = ref({})
  const submitting = ref(false)
  const submitError = ref('')
  const sectionErrors: Ref<Record<string, string>> = ref({})

  async function submit(sectionDefs?: TemplateSectionDef[]): Promise<ADRDetail | null> {
    submitError.value = ''
    sectionErrors.value = {}

    const trimmed = title.value.trim()
    if (!trimmed) {
      submitError.value = 'Title is required'
      return null
    }

    // Validate required sections
    if (sectionDefs) {
      let hasError = false
      for (const def of sectionDefs) {
        if (!def.optional) {
          const val = (sections.value[def.key] || '').trim()
          if (!val) {
            sectionErrors.value[def.key] = `${def.heading} is required`
            hasError = true
          }
        }
      }
      if (hasError) {
        return null
      }
    }

    // Build sections payload — only include non-empty values
    const sectionPayload: Record<string, string> = {}
    for (const [key, value] of Object.entries(sections.value)) {
      const trimmedVal = value.trim()
      if (trimmedVal) {
        sectionPayload[key] = trimmedVal
      }
    }

    submitting.value = true
    try {
      const payload = { title: trimmed, ...(Object.keys(sectionPayload).length > 0 ? { sections: sectionPayload } : {}) }
      const result = await createADR(payload)
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
    sections,
    submitting,
    submitError,
    sectionErrors,
    submit,
  }
}
