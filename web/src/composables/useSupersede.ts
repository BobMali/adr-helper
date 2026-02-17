import { ref, onUnmounted } from 'vue'
import type { ADRSummary } from '../types'
import { fetchADRs } from '../api'

export function useSupersede(adrNumber: number) {
  const pendingSuperseded = ref(false)
  const supersededBy = ref<number | null>(null)
  const availableADRs = ref<ADRSummary[]>([])
  const loadingADRs = ref(false)

  let supersedeFetchController: AbortController | null = null

  async function startSupersede() {
    supersedeFetchController?.abort()
    supersedeFetchController = new AbortController()
    const currentController = supersedeFetchController

    pendingSuperseded.value = true
    supersededBy.value = null
    loadingADRs.value = true
    try {
      const allADRs = await fetchADRs(undefined, currentController.signal)
      if (currentController !== supersedeFetchController) return
      availableADRs.value = allADRs.filter(a => a.number !== adrNumber)
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      if (currentController !== supersedeFetchController) return
      availableADRs.value = []
    } finally {
      if (currentController === supersedeFetchController) {
        loadingADRs.value = false
      }
    }
  }

  function cancelSupersede() {
    supersedeFetchController?.abort()
    supersedeFetchController = null
    pendingSuperseded.value = false
    supersededBy.value = null
    availableADRs.value = []
  }

  function onSelectorKeydown(
    event: KeyboardEvent,
    onConfirm: () => void,
  ) {
    if (event.key === 'Escape') {
      event.preventDefault()
      cancelSupersede()
    } else if (event.key === 'Enter' && supersededBy.value != null) {
      event.preventDefault()
      onConfirm()
    }
  }

  onUnmounted(() => {
    supersedeFetchController?.abort()
  })

  return {
    pendingSuperseded,
    supersededBy,
    availableADRs,
    loadingADRs,
    startSupersede,
    cancelSupersede,
    onSelectorKeydown,
  }
}
