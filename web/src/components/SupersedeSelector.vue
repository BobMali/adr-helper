<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ADRSummary } from '../types'

const props = defineProps<{
  availableADRs: ADRSummary[]
  loadingADRs: boolean
  modelValue: number | null
  disabled: boolean
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
  'update:modelValue': [value: number | null]
}>()

const selectRef = ref<HTMLSelectElement | null>(null)

watch(() => props.loadingADRs, (newVal, oldVal) => {
  if (oldVal && !newVal && props.availableADRs.length > 0) {
    setTimeout(() => selectRef.value?.focus(), 0)
  }
})

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('cancel')
  } else if (event.key === 'Enter' && props.modelValue != null) {
    event.preventDefault()
    emit('confirm')
  }
}
</script>

<template>
  <div
    role="group"
    aria-labelledby="supersede-label"
    class="mt-3 ml-4 p-3 border-l-2 border-blue-500 bg-blue-50 dark:bg-blue-900/10 rounded-r"
    @keydown="onKeydown"
  >
    <p id="supersede-label" class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
      Select the ADR that supersedes this one:
    </p>
    <div aria-live="polite">
      <div v-if="loadingADRs" class="text-sm text-gray-500 dark:text-gray-400">Loading ADRs…</div>
      <div v-else-if="availableADRs.length === 0" class="text-sm text-gray-500 dark:text-gray-400">
        No other ADRs available
      </div>
      <select
        v-else
        ref="selectRef"
        :value="modelValue"
        class="rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm px-2 py-1 text-gray-900 dark:text-gray-100 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none w-full"
        @change="emit('update:modelValue', Number(($event.target as HTMLSelectElement).value) || null)"
      >
        <option :value="null" disabled>— Choose an ADR —</option>
        <option v-for="a in availableADRs" :key="a.number" :value="a.number">
          ADR-{{ String(a.number).padStart(4, '0') }}: {{ a.title }} ({{ a.status }})
        </option>
      </select>
    </div>
    <div class="mt-2 flex gap-2">
      <button
        :disabled="modelValue == null || disabled"
        class="px-3 py-1 text-sm rounded bg-blue-600 text-white hover:bg-blue-700 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
        @click="emit('confirm')"
      >
        Confirm
      </button>
      <button
        :disabled="disabled"
        class="px-3 py-1 text-sm rounded border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none"
        @click="emit('cancel')"
      >
        Cancel
      </button>
    </div>
  </div>
</template>
