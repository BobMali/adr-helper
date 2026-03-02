<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import type { ADRSummary } from '../types'
import { statusDotClass } from '../utils/statusColors'

const props = defineProps<{
  searchResults: ADRSummary[]
  searching: boolean
  disabled: boolean
}>()

const emit = defineEmits<{
  search: [query: string]
  select: [number: number]
  cancel: []
}>()

const query = ref('')
const highlightedIndex = ref(-1)
const inputRef = ref<HTMLInputElement | null>(null)

const showResults = computed(() => query.value.trim().length > 0)
const hasResults = computed(() => props.searchResults.length > 0 && showResults.value)

const activeDescendant = computed(() => {
  const item = props.searchResults[highlightedIndex.value]
  if (!item) return undefined
  return `relation-option-${item.number}`
})

const resultCountText = computed(() => {
  if (!showResults.value) return ''
  const count = props.searchResults.length
  if (count === 0) return 'No results found'
  return `${count} result${count === 1 ? '' : 's'} available`
})

function onInput() {
  highlightedIndex.value = -1
  emit('search', query.value)
}

function selectResult(adr: ADRSummary) {
  emit('select', adr.number)
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (hasResults.value) {
      highlightedIndex.value = Math.min(highlightedIndex.value + 1, props.searchResults.length - 1)
    }
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (hasResults.value) {
      highlightedIndex.value = Math.max(highlightedIndex.value - 1, 0)
    }
  } else if (event.key === 'Enter') {
    event.preventDefault()
    const item = props.searchResults[highlightedIndex.value]
    if (item !== undefined) {
      selectResult(item)
    }
  } else if (event.key === 'Escape') {
    event.preventDefault()
    emit('cancel')
  }
}

function formatNumber(n: number): string {
  return `ADR-${String(n).padStart(4, '0')}`
}

onMounted(async () => {
  await nextTick()
  inputRef.value?.focus()
})
</script>

<template>
  <div class="mt-3 ml-4 p-3 border-l-2 border-gray-300 dark:border-gray-500 bg-gray-50 dark:bg-gray-900/20 rounded-r">
    <label for="relation-search-input" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
      Search for a related ADR
    </label>
    <input
      id="relation-search-input"
      ref="inputRef"
      v-model="query"
      role="combobox"
      :aria-expanded="hasResults ? 'true' : 'false'"
      aria-controls="relation-listbox"
      aria-autocomplete="list"
      :aria-activedescendant="activeDescendant"
      :disabled="disabled"
      placeholder="Type to search ADRs…"
      class="w-full rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm px-2 py-1 text-gray-900 dark:text-gray-100 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none disabled:opacity-50"
      @input="onInput"
      @keydown="onKeydown"
    />

    <!-- Screen-reader result count -->
    <div class="sr-only" aria-live="polite">{{ resultCountText }}</div>

    <!-- Searching indicator -->
    <p v-if="searching && showResults" class="mt-1 text-sm text-gray-500 dark:text-gray-400">
      Searching…
    </p>

    <!-- Results listbox -->
    <ul
      v-if="hasResults"
      id="relation-listbox"
      role="listbox"
      class="mt-1 max-h-40 overflow-y-auto rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900"
    >
      <li
        v-for="(adr, index) in searchResults"
        :id="`relation-option-${adr.number}`"
        :key="adr.number"
        role="option"
        :aria-selected="index === highlightedIndex"
        class="flex items-center gap-2 px-2 py-1.5 text-sm cursor-pointer"
        :class="index === highlightedIndex ? 'bg-blue-100 dark:bg-blue-900/40' : 'hover:bg-gray-100 dark:hover:bg-gray-800'"
        @click="selectResult(adr)"
      >
        <span class="inline-block w-2 h-2 rounded-full shrink-0" :class="statusDotClass(adr.status)"></span>
        <span class="text-gray-900 dark:text-gray-100">{{ formatNumber(adr.number) }}: {{ adr.title }}</span>
        <span class="text-xs text-gray-500 dark:text-gray-400 ml-auto">{{ adr.status }}</span>
      </li>
    </ul>

    <!-- No results -->
    <p
      v-if="showResults && !searching && searchResults.length === 0"
      class="mt-1 text-sm text-gray-500 dark:text-gray-400"
    >
      No ADRs match "{{ query }}"
    </p>
  </div>
</template>
