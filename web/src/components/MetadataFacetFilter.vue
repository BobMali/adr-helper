<script setup lang="ts">
import { computed } from 'vue'
import type { MetaMatchMode } from '../types'

// Number of values above which the chip list gets a scrollable cap.
const SCROLL_CAP_THRESHOLD = 8

const props = defineProps<{
  heading: string
  values: string[]
  modelValue: Set<string>
  matchMode: MetaMatchMode
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Set<string>]
  'update:matchMode': [value: MetaMatchMode]
}>()

const showMatchToggle = computed(() => props.modelValue.size >= 2)
const capped = computed(() => props.values.length > SCROLL_CAP_THRESHOLD)

function toggle(value: string) {
  // Always emit a NEW Set so the parent's shallow watcher (URL sync) fires.
  const next = new Set(props.modelValue)
  if (next.has(value)) {
    next.delete(value)
  } else {
    next.add(value)
  }
  emit('update:modelValue', next)
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between gap-2 mb-1.5">
      <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ heading }}</span>
      <div
        v-if="showMatchToggle"
        role="group"
        aria-label="Match any or all selected scopes"
        class="inline-flex rounded border border-gray-200 dark:border-gray-700 overflow-hidden text-xs"
      >
        <button
          type="button"
          :aria-pressed="matchMode === 'any'"
          class="px-2 py-0.5 transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
          :class="matchMode === 'any'
            ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300'
            : 'bg-transparent text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
          @click="emit('update:matchMode', 'any')"
        >
          Any (Union)
        </button>
        <button
          type="button"
          :aria-pressed="matchMode === 'all'"
          class="px-2 py-0.5 border-l border-gray-200 dark:border-gray-700 transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
          :class="matchMode === 'all'
            ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300'
            : 'bg-transparent text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
          @click="emit('update:matchMode', 'all')"
        >
          All (Intersect)
        </button>
      </div>
    </div>

    <div
      role="group"
      :aria-label="`Filter by ${heading}`"
      class="flex flex-wrap gap-2"
      :class="capped ? 'max-h-40 overflow-y-auto rounded border border-gray-200 dark:border-gray-700 p-2' : ''"
    >
      <button
        v-for="value in values"
        :key="value"
        type="button"
        :aria-pressed="modelValue.has(value)"
        class="inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
        :class="modelValue.has(value)
          ? 'bg-blue-600 text-white'
          : 'border border-gray-300 dark:border-gray-600 bg-transparent text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'"
        @click="toggle(value)"
      >
        {{ value }}
      </button>
    </div>
  </div>
</template>
