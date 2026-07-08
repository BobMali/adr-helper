<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import type { SortField, SortDirection, MetaField, MetaMatchMode, ADRSummary } from '../types'
import { fetchStatuses, fetchMetaFields } from '../api'
import StatusFilterChips from '../components/StatusFilterChips.vue'
import MetadataFacetFilter from '../components/MetadataFacetFilter.vue'
import { statusDotClass, statusTextClass } from '../utils/statusColors'
import { useADRSearch } from '../composables/useADRSearch'
import { useURLSync } from '../composables/useURLSync'

// Number of scope badges shown on a row before collapsing the rest into "+N".
const MAX_ROW_BADGES = 3

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

// Metadata facets. Chip values come from the field's vocabulary (search-independent),
// like status chips come from the status enum.
const metaFields = ref<MetaField[]>([])
const selectedMeta = ref<Record<string, Set<string>>>({})
const matchMode = ref<MetaMatchMode>('any')
// The scope facet(s) live behind a collapsible "More filters" disclosure.
const filtersOpen = ref(false)

const { adrs, loading, error, searchQuery, hasSearchQuery, loadADRs, onSearchInput, clearSearch } = useADRSearch()
const { syncToURL, initFromURL } = useURLSync({
  searchQuery, selectedStatuses, sortField, sortDirection, selectedMeta, matchMode,
})

// Only vocabulary fields with values get chip filters (and row badges).
const vocabularyFacets = computed(() =>
  metaFields.value.filter(f => f.vocabulary && (f.values?.length ?? 0) > 0),
)
const vocabularyKeys = computed(() => new Set(vocabularyFacets.value.map(f => f.key)))

// Badge count: only selections for facets that actually render, so stale meta_* URL
// params (not validated against the vocabulary) don't inflate the count.
const activeFilterCount = computed(() =>
  vocabularyFacets.value.reduce((n, f) => n + (selectedMeta.value[f.key]?.size ?? 0), 0),
)

function facetSelection(key: string): Set<string> {
  return selectedMeta.value[key] ?? new Set<string>()
}

function setFacet(key: string, next: Set<string>) {
  // Rebuild the Record so the shallow watcher fires (URL sync).
  const record = { ...selectedMeta.value }
  if (next.size === 0) {
    delete record[key]
  } else {
    record[key] = next
  }
  selectedMeta.value = record
}

function facetMatches(adr: ADRSummary, key: string, selected: Set<string>): boolean {
  const present = new Set((adr.meta?.[key] ?? []).map(v => v.toLowerCase()))
  const wanted = [...selected].map(v => v.toLowerCase())
  return matchMode.value === 'all'
    ? wanted.every(w => present.has(w))
    : wanted.some(w => present.has(w))
}

function rowBadges(adr: ADRSummary): string[] {
  if (!adr.meta) return []
  const out: string[] = []
  for (const key of vocabularyKeys.value) {
    for (const v of adr.meta[key] ?? []) out.push(v)
  }
  return out
}

const filteredADRs = computed(() => {
  let list = adrs.value
  if (selectedStatuses.value.size > 0) {
    list = list.filter(adr => selectedStatuses.value.has(adr.status))
  }
  const activeFacets = Object.entries(selectedMeta.value).filter(([, set]) => set.size > 0)
  if (activeFacets.length > 0) {
    list = list.filter(adr => activeFacets.every(([key, set]) => facetMatches(adr, key, set)))
  }
  return list
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

watch([selectedMeta, matchMode], () => {
  syncToURL()
})

watch([sortField, sortDirection], () => {
  syncToURL()
})

onMounted(() => {
  initFromURL()

  // Auto-open the "More filters" panel if a scope filter is already active (e.g. a
  // bookmarked URL). initFromURL() is synchronous, so selectedMeta is populated here
  // even though the facet vocabulary (vocabularyFacets) loads async below.
  if (Object.values(selectedMeta.value).some(s => s.size > 0)) {
    filtersOpen.value = true
  }

  const q = searchQuery.value.trim()
  loadADRs(q.length >= 2 ? q : undefined)

  fetchStatuses()
    .then(s => { availableStatuses.value = s })
    .catch(() => {
      const fromADRs = [...new Set(adrs.value.map(a => a.status))]
      if (fromADRs.length > 0) availableStatuses.value = fromADRs
    })

  fetchMetaFields()
    .then(f => { metaFields.value = f })
    .catch(() => { /* facets are optional; leave empty on failure */ })
})
</script>

<template>
  <div class="flex flex-col h-full">
  <header class="mb-8 flex items-center justify-between">
    <h1 class="text-2xl font-semibold tracking-tight">Architecture Decision Records</h1>
    <RouterLink
      :to="{ name: 'create' }"
      class="px-4 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    >
      + New ADR
    </RouterLink>
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

  <!-- Metadata facet filters (e.g. Scope) behind a collapsible "More filters" disclosure.
       The body loops all vocabularyFacets, so future facets land here automatically. -->
  <div v-if="vocabularyFacets.length > 0" class="mb-6">
    <button
      type="button"
      :aria-expanded="filtersOpen"
      aria-controls="metadata-filters"
      class="inline-flex items-center gap-1.5 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      @click="filtersOpen = !filtersOpen"
    >
      <span aria-hidden="true" class="text-xs">{{ filtersOpen ? '▾' : '▸' }}</span>
      More filters
      <span
        v-if="activeFilterCount > 0"
        class="inline-flex items-center justify-center px-1.5 py-0.5 rounded-full text-xs bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300"
      >{{ activeFilterCount }}<span class="sr-only"> active filters</span></span>
    </button>

    <div v-show="filtersOpen" id="metadata-filters" class="mt-3 flex flex-col gap-4">
      <MetadataFacetFilter
        v-for="facet in vocabularyFacets"
        :key="facet.key"
        :heading="facet.heading"
        :values="facet.values ?? []"
        :model-value="facetSelection(facet.key)"
        :match-mode="matchMode"
        @update:model-value="(v) => setFacet(facet.key, v)"
        @update:match-mode="(m) => (matchMode = m)"
      />
    </div>
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
    <p class="mt-1 text-sm text-gray-400 dark:text-gray-500">
      <RouterLink
        :to="{ name: 'create' }"
        class="text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      >
        Create your first ADR
      </RouterLink>
      to get started.
    </p>
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
      {{ adrs.length }} ADR{{ adrs.length === 1 ? '' : 's' }} available — try selecting different statuses or scopes, or clearing your search.
    </p>
  </div>

  <!-- ADR list -->
  <ul v-else aria-live="polite" tabindex="0" role="region" aria-label="ADR list" class="flex-1 min-h-0 overflow-y-auto divide-y divide-gray-200 dark:divide-gray-800 border-t border-b border-gray-200 dark:border-gray-800">
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

        <span
          v-if="rowBadges(adr).length > 0"
          class="shrink-0 hidden sm:flex items-center gap-1"
          aria-hidden="true"
        >
          <span
            v-for="value in rowBadges(adr).slice(0, MAX_ROW_BADGES)"
            :key="value"
            class="inline-block px-2 py-0.5 rounded-full text-xs bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300"
          >
            {{ value }}
          </span>
          <span
            v-if="rowBadges(adr).length > MAX_ROW_BADGES"
            class="text-xs text-gray-400 dark:text-gray-500"
          >
            +{{ rowBadges(adr).length - MAX_ROW_BADGES }}
          </span>
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
  </div>
</template>
