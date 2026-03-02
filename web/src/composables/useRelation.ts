import { ref, onUnmounted } from 'vue'
import type { ADRDetail } from '../types'
import { addRelation } from '../api'

export function useRelation(adrNumber: number) {
  const adding = ref(false)
  const feedback = ref('')
  const feedbackType = ref<'success' | 'error' | ''>('')

  let feedbackTimer: ReturnType<typeof setTimeout> | null = null

  function clearFeedbackTimer() {
    if (feedbackTimer !== null) {
      clearTimeout(feedbackTimer)
      feedbackTimer = null
    }
  }

  function setFeedback(msg: string, type: 'success' | 'error') {
    clearFeedbackTimer()
    feedback.value = msg
    feedbackType.value = type
    feedbackTimer = setTimeout(() => {
      feedback.value = ''
      feedbackType.value = ''
    }, 4000)
  }

  async function confirmRelation(targetNumber: number): Promise<ADRDetail | null> {
    if (adding.value) return null

    adding.value = true
    try {
      const result = await addRelation(adrNumber, targetNumber)
      setFeedback('Relation added successfully', 'success')
      return result
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Failed to add relation'
      setFeedback(msg, 'error')
      return null
    } finally {
      adding.value = false
    }
  }

  onUnmounted(() => {
    clearFeedbackTimer()
  })

  return {
    adding,
    feedback,
    feedbackType,
    confirmRelation,
  }
}
