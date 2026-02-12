<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import type { ADRSummary } from '../types'
import { fetchADRs } from '../api'

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
    const result = query
      ? await fetchADRs(query, currentController.signal)
      : await fetchADRs(undefined, currentController.signal)
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
    if (q.length >= 2) {
      loadADRs(q)
    } else {
      loadADRs()
    }
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

onMounted(loadADRs)

onUnmounted(() => {
  if (debounceTimer !== null) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
  abortController?.abort()
})

function statusDotClass(status: string): string {
  const s = status.toLowerCase()
  if (s === 'accepted') return 'bg-green-500'
  if (s === 'proposed') return 'bg-amber-500'
  return 'bg-red-500'
}

function statusTextClass(status: string): string {
  const s = status.toLowerCase()
  if (s === 'accepted') return 'text-green-600 dark:text-green-400'
  if (s === 'proposed') return 'text-amber-600 dark:text-amber-400'
  return 'text-red-600 dark:text-red-400'
}
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
    class="w-full mb-6 py-2.5 px-4 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
    @input="onSearchInput"
    @keydown.esc="clearSearch"
  />

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

  <!-- ADR list -->
  <ul v-else aria-live="polite" class="divide-y divide-gray-200 dark:divide-gray-800 border-t border-b border-gray-200 dark:border-gray-800">
    <li
      v-for="adr in adrs"
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
</template>
