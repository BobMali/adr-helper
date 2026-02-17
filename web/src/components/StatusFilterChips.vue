<script setup lang="ts">
import { statusDotClass, chipBgClass } from '../utils/statusColors'

const props = defineProps<{
  statuses: string[]
  modelValue: Set<string>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Set<string>]
}>()

function toggle(status: string) {
  const next = new Set(props.modelValue)
  if (next.has(status)) {
    next.delete(status)
  } else {
    next.add(status)
  }
  emit('update:modelValue', next)
}
</script>

<template>
  <div role="group" aria-label="Filter by status" class="flex flex-wrap gap-2">
    <button
      v-for="status in statuses"
      :key="status"
      type="button"
      :aria-pressed="modelValue.has(status)"
      class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      :class="modelValue.has(status)
        ? [chipBgClass(status), 'text-white']
        : 'border border-gray-300 dark:border-gray-600 bg-transparent text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'"
      @click="toggle(status)"
    >
      <span
        class="inline-block w-2 h-2 rounded-full"
        :class="modelValue.has(status) ? 'bg-white' : statusDotClass(status)"
        aria-hidden="true"
      ></span>
      {{ status }}
    </button>
  </div>
</template>
