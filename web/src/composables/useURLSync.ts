import { watch, nextTick, type Ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { SortField, SortDirection } from '../types'

const VALID_SORT_FIELDS = new Set<string>(['number', 'title', 'status'])

export function useURLSync(
  searchQuery: Ref<string>,
  selectedStatuses: Ref<Set<string>>,
  sortField: Ref<SortField>,
  sortDirection: Ref<SortDirection>,
) {
  const route = useRoute()
  const router = useRouter()

  let internalNavigationCount = 0

  function buildQuery(): Record<string, string | string[]> {
    const q: Record<string, string | string[]> = {}
    const searchTrimmed = searchQuery.value.trim()
    if (searchTrimmed.length >= 2) {
      q.q = searchTrimmed
    }
    if (selectedStatuses.value.size > 0) {
      const arr = [...selectedStatuses.value]
      q.status = arr.length === 1 ? arr[0]! : arr
    }
    if (sortField.value !== 'number') {
      q.sort = sortField.value
    }
    if (sortDirection.value !== 'asc') {
      q.dir = sortDirection.value
    }
    return q
  }

  function syncToURL() {
    internalNavigationCount++
    router.replace({ query: buildQuery() }).finally(() => {
      nextTick(() => { internalNavigationCount-- })
    })
  }

  function readStateFromQuery(query: Record<string, unknown>) {
    const statusParam = query.status
    if (statusParam) {
      const arr = Array.isArray(statusParam) ? statusParam : [statusParam]
      selectedStatuses.value = new Set(arr.filter((s): s is string => typeof s === 'string'))
    } else {
      selectedStatuses.value = new Set()
    }

    if (typeof query.q === 'string') {
      searchQuery.value = query.q
    } else {
      searchQuery.value = ''
    }

    const sortParam = typeof query.sort === 'string' && VALID_SORT_FIELDS.has(query.sort)
      ? query.sort as SortField
      : 'number'
    const dirParam = query.dir === 'desc' ? 'desc' as const : 'asc' as const

    if (sortField.value !== sortParam) sortField.value = sortParam
    if (sortDirection.value !== dirParam) sortDirection.value = dirParam
  }

  function initFromURL() {
    readStateFromQuery(route.query as Record<string, unknown>)
  }

  watch(() => route.query, (newQuery) => {
    if (internalNavigationCount > 0) return
    readStateFromQuery(newQuery as Record<string, unknown>)
  }, { deep: true })

  return {
    syncToURL,
    initFromURL,
  }
}
