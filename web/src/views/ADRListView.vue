<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import type { SortField, SortDirection } from '../types'
import { fetchStatuses } from '../api'
import StatusFilterChips from '../components/StatusFilterChips.vue'
import { statusDotClass, statusTextClass } from '../utils/statusColors'
import { useADRSearch } from '../composables/useADRSearch'
import { useURLSync } from '../composables/useURLSync'

const STATUS_ORDER: Record<string, number> = {
  proposed: 0, accepted: 1, deprecated: 2, superseded: 3, rejected: 4,
}

function statusOrdinal(status: string): number {
  return STATUS_ORDER[status.toLowerCase()] ?? 999
}

const SORT_FIELDS = [
  { value: 'number', label: 'ID' },
  { value: 'title', label: 'Title' },
  { value: 'status', label: 'Status' },
] as const satisfies readonly { value: SortField; label: string }[]

const selectedStatuses = ref<Set<string>>(new Set())
const availableStatuses = ref<string[]>([])
const sortField = ref<SortField>('number')
const sortDirection = ref<SortDirection>('asc')

const { adrs, loading, error, searchQuery, hasSearchQuery, loadADRs, onSearchInput, clearSearch } = useADRSearch()
const { syncToURL, initFromURL } = useURLSync(searchQuery, selectedStatuses, sortField, sortDirection)

const filteredADRs = computed(() => {
  if (selectedStatuses.value.size === 0) return adrs.value
  return adrs.value.filter(adr => selectedStatuses.value.has(adr.status))
})

const sortedADRs = computed(() => {
  const list = [...filteredADRs.value]
  const dir = sortDirection.value === 'asc' ? 1 : -1
  return list.sort((a, b) => {
    switch (sortField.value) {
      case 'number': return (a.number - b.number) * dir
      case 'title':  return a.title.localeCompare(b.title) * dir
      case 'status': return (statusOrdinal(a.status) - statusOrdinal(b.status)) * dir
    }
  })
})

function setSort(field: SortField) {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

function getSortAriaLabel(field: { value: SortField; label: string }): string {
  if (sortField.value !== field.value) return `Sort by ${field.label}`
  const dir = sortDirection.value === 'asc' ? 'ascending' : 'descending'
  const rev = sortDirection.value === 'asc' ? 'descending' : 'ascending'
  return `Sort by ${field.label}, currently ${dir}, click to sort ${rev}`
}

function handleSearchInput() {
  onSearchInput()
  syncToURL()
}

function handleClearSearch() {
  clearSearch()
  syncToURL()
}

watch(selectedStatuses, () => {
  syncToURL()
})

watch([sortField, sortDirection], () => {
  syncToURL()
})

onMounted(() => {
  initFromURL()

  const q = searchQuery.value.trim()
  loadADRs(q.length >= 2 ? q : undefined)

  fetchStatuses()
    .then(s => { availableStatuses.value = s })
    .catch(() => {
      const fromADRs = [...new Set(adrs.value.map(a => a.status))]
      if (fromADRs.length > 0) availableStatuses.value = fromADRs
    })
})
</script>

<template>
  <header class="mb-8">
    <h1 class="text-2xl font-semibold tracking-tight">Architecture Decision Records</h1>
  </header>

  <input
    v-model="searchQuery"
    type="search"
    aria-label="Search ADRs"
    placeholder="Search ADRs…"
    class="w-full mb-4 py-2.5 px-4 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
    @input="handleSearchInput"
    @keydown.esc="handleClearSearch"
  />

  <div v-if="availableStatuses.length > 0" class="mb-6">
    <StatusFilterChips
      v-model="selectedStatuses"
      :statuses="availableStatuses"
    />
  </div>

  <!-- Sort controls -->
  <div class="mb-4 mt-4 flex items-center gap-2">
    <span class="sr-only sm:not-sr-only text-xs text-gray-500 dark:text-gray-400">Sort by:</span>
    <div role="group" aria-label="Sort options" class="grid grid-cols-3 gap-1 sm:flex sm:gap-1">
      <button
        v-for="field in SORT_FIELDS"
        :key="field.value"
        :aria-pressed="sortField === field.value ? 'true' : 'false'"
        :aria-label="getSortAriaLabel(field)"
        class="px-2 py-1 text-xs rounded border transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
        :class="sortField === field.value
          ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 border-blue-300 dark:border-blue-700'
          : 'bg-transparent text-gray-600 dark:text-gray-400 border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800'"
        @click="setSort(field.value)"
      >
        {{ field.label }}
        <span v-if="sortField === field.value" aria-hidden="true">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
      </button>
    </div>
  </div>

  <!-- Loading -->
  <div v-if="loading" role="status" class="text-center py-16 text-gray-500 dark:text-gray-400">
    Loading…
  </div>

  <!-- Error -->
  <div v-else-if="error" class="text-center py-16">
    <p class="text-red-600 dark:text-red-400">{{ error }}</p>
    <button
      class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      @click="loadADRs()"
    >
      Retry
    </button>
  </div>

  <!-- Empty state: no search -->
  <div v-else-if="adrs.length === 0 && !hasSearchQuery" class="text-center py-16">
    <p class="text-lg font-medium text-gray-500 dark:text-gray-400">No ADRs yet</p>
    <p class="mt-1 text-sm text-gray-400 dark:text-gray-500">Create your first Architecture Decision Record to get started.</p>
  </div>

  <!-- Empty state: search with no results -->
  <div v-else-if="adrs.length === 0 && hasSearchQuery" class="text-center py-16">
    <p class="text-lg font-medium text-gray-500 dark:text-gray-400">No ADRs match "{{ searchQuery.trim() }}"</p>
    <p class="mt-1 text-sm text-gray-400 dark:text-gray-500">Try a different search term.</p>
  </div>

  <!-- Status filter excludes all results -->
  <div v-else-if="filteredADRs.length === 0 && adrs.length > 0"
       class="text-center py-16" role="status" aria-live="polite">
    <p class="text-lg font-medium text-gray-500 dark:text-gray-400">
      No ADRs match the selected filters
    </p>
    <p class="mt-1 text-sm text-gray-400 dark:text-gray-500">
      {{ adrs.length }} ADR{{ adrs.length === 1 ? '' : 's' }} available — try selecting different statuses or clearing your search.
    </p>
  </div>

  <!-- ADR list -->
  <ul v-else aria-live="polite" class="divide-y divide-gray-200 dark:divide-gray-800 border-t border-b border-gray-200 dark:border-gray-800">
    <li
      v-for="adr in sortedADRs"
      :key="adr.number"
    >
      <RouterLink
        :to="{ name: 'detail', params: { number: adr.number } }"
        :aria-label="`ADR #${adr.number}: ${adr.title}`"
        class="flex items-center gap-4 py-3 px-2 sm:px-0 hover:bg-gray-50 dark:hover:bg-gray-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 transition-colors"
      >
        <span class="w-12 shrink-0 text-sm font-mono text-gray-500 dark:text-gray-400">
          #{{ adr.number }}
        </span>

        <span class="flex items-center gap-1.5 w-28 shrink-0 text-sm">
          <span
            class="inline-block w-2 h-2 rounded-full"
            :class="statusDotClass(adr.status)"
            aria-hidden="true"
          ></span>
          <span :class="statusTextClass(adr.status)">{{ adr.status }}</span>
        </span>

        <span class="flex-1 min-w-0 truncate text-sm font-medium">
          {{ adr.title }}
        </span>

        <time
          v-if="adr.date"
          :datetime="adr.date"
          class="shrink-0 text-sm text-gray-500 dark:text-gray-400 hidden sm:block"
        >
          {{ adr.date }}
        </time>
      </RouterLink>
    </li>
  </ul>

  <!-- Screen reader count announcement (outside v-if chain) -->
  <div v-if="!loading && !error && adrs.length > 0" class="sr-only" role="status" aria-live="polite" aria-atomic="true">
    {{ sortedADRs.length }} record{{ sortedADRs.length !== 1 ? 's' : '' }} shown
  </div>
</template>
