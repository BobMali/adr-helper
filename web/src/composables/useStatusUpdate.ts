import { ref, onUnmounted } from 'vue'
import type { ADRDetail } from '../types'
import { updateADRStatus } from '../api'

export function useStatusUpdate(adrNumber: number) {
  const updating = ref(false)
  const feedback = ref('')
  const feedbackType = ref<'success' | 'error'>('success')

  let feedbackTimer: ReturnType<typeof setTimeout> | undefined
  let previousStatus = ''

  function setPreviousStatus(status: string) {
    previousStatus = status
  }

  function getPreviousStatus(): string {
    return previousStatus
  }

  async function doStatusUpdate(
    newStatus: string,
    options?: { supersededBy?: number },
  ): Promise<ADRDetail | null> {
    updating.value = true
    feedback.value = ''

    if (feedbackTimer) {
      clearTimeout(feedbackTimer)
      feedbackTimer = undefined
    }

    try {
      const updated = await updateADRStatus(adrNumber, newStatus, options)
      previousStatus = updated.status
      feedbackType.value = 'success'
      feedback.value = `Status updated to ${updated.status}`
      feedbackTimer = setTimeout(() => { feedback.value = '' }, 4000)
      return updated
    } catch (e) {
      feedbackType.value = 'error'
      feedback.value = e instanceof Error ? e.message : 'Failed to update status'
      return null
    } finally {
      updating.value = false
    }
  }

  onUnmounted(() => {
    if (feedbackTimer) {
      clearTimeout(feedbackTimer)
    }
  })

  return {
    updating,
    feedback,
    feedbackType,
    doStatusUpdate,
    setPreviousStatus,
    getPreviousStatus,
  }
}
