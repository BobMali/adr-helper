import { ref, computed, onUnmounted } from 'vue'
import type { ADRSummary } from '../types'
import { fetchADRs } from '../api'

export function useADRSearch() {
  const adrs = ref<ADRSummary[]>([])
  const loading = ref(true)
  const error = ref('')
  const searchQuery = ref('')

  let debounceTimer: ReturnType<typeof setTimeout> | null = null
  let abortController: AbortController | null = null

  const hasSearchQuery = computed(() => searchQuery.value.trim().length > 0)

  async function loadADRs(query?: string) {
    abortController?.abort()
    abortController = new AbortController()
    const currentController = abortController

    loading.value = true
    error.value = ''
    try {
      const result = await fetchADRs(query, currentController.signal)
      if (currentController !== abortController) return
      adrs.value = result
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      if (currentController !== abortController) return
      error.value = e instanceof Error ? e.message : 'Unknown error'
    } finally {
      if (currentController === abortController) {
        loading.value = false
      }
    }
  }

  function onSearchInput() {
    error.value = ''
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
    }
    debounceTimer = setTimeout(() => {
      const q = searchQuery.value.trim()
      loadADRs(q.length >= 2 ? q : undefined)
    }, 300)
  }

  function clearSearch() {
    searchQuery.value = ''
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    abortController?.abort()
    loadADRs()
  }

  onUnmounted(() => {
    if (debounceTimer !== null) {
      clearTimeout(debounceTimer)
      debounceTimer = null
    }
    abortController?.abort()
  })

  return {
    adrs,
    loading,
    error,
    searchQuery,
    hasSearchQuery,
    loadADRs,
    onSearchInput,
    clearSearch,
  }
}
