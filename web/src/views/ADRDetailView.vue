<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { marked } from 'marked'
import type { ADRDetail } from '../types'
import { fetchADR, fetchStatuses, updateADRStatus, NotFoundError } from '../api'

const props = defineProps<{ number: number }>()

const adr = ref<ADRDetail | null>(null)
const statuses = ref<string[]>([])
const loading = ref(true)
const error = ref('')
const notFound = ref(false)
const updating = ref(false)
const feedback = ref('')
const feedbackType = ref<'success' | 'error'>('success')
const titleRef = ref<HTMLHeadingElement | null>(null)

let previousStatus = ''
let feedbackTimer: ReturnType<typeof setTimeout> | undefined

const renderedContent = computed(() => {
  if (!adr.value?.content) return ''
  return marked(adr.value.content) as string
})

const formattedDate = computed(() => {
  if (!adr.value?.date) return ''
  const d = new Date(adr.value.date + 'T00:00:00')
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(d)
})

onMounted(async () => {
  try {
    const [adrData, statusData] = await Promise.all([
      fetchADR(props.number),
      fetchStatuses(),
    ])
    adr.value = adrData
    statuses.value = statusData
    previousStatus = adrData.status

    await nextTick()
    titleRef.value?.focus()
  } catch (e) {
    if (e instanceof NotFoundError) {
      notFound.value = true
    } else {
      error.value = e instanceof Error ? e.message : 'Unknown error'
    }
  } finally {
    loading.value = false
  }
})

async function onStatusChange(event: Event) {
  const select = event.target as HTMLSelectElement
  const newStatus = select.value

  if (newStatus === previousStatus) return

  updating.value = true
  feedback.value = ''

  if (feedbackTimer) {
    clearTimeout(feedbackTimer)
    feedbackTimer = undefined
  }

  try {
    const updated = await updateADRStatus(props.number, newStatus)
    adr.value = updated
    previousStatus = updated.status
    feedbackType.value = 'success'
    feedback.value = `Status updated to ${updated.status}`
    feedbackTimer = setTimeout(() => { feedback.value = '' }, 4000)
  } catch (e) {
    select.value = previousStatus
    feedbackType.value = 'error'
    feedback.value = e instanceof Error ? e.message : 'Failed to update status'
  } finally {
    updating.value = false
  }
}
</script>

<template>
  <!-- Loading -->
  <div v-if="loading" class="text-center py-16 text-gray-500 dark:text-gray-400">
    Loading…
  </div>

  <!-- Not found -->
  <div v-else-if="notFound" class="text-center py-16">
    <p class="text-lg font-medium text-gray-500 dark:text-gray-400">ADR #{{ number }} not found</p>
    <RouterLink
      to="/"
      class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    >
      ← Back to list
    </RouterLink>
  </div>

  <!-- Error -->
  <div v-else-if="error" class="text-center py-16">
    <p class="text-red-600 dark:text-red-400">{{ error }}</p>
    <RouterLink
      to="/"
      class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    >
      ← Back to list
    </RouterLink>
  </div>

  <!-- Detail -->
  <article v-else-if="adr" aria-labelledby="adr-title">
    <nav class="mb-6">
      <RouterLink
        to="/"
        class="text-sm text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      >
        ← Back to list
      </RouterLink>
    </nav>

    <header class="mb-6">
      <h1
        id="adr-title"
        ref="titleRef"
        tabindex="-1"
        class="text-2xl font-semibold tracking-tight focus:outline-none"
      >
        ADR #{{ adr.number }}: {{ adr.title }}
      </h1>

      <div class="mt-2 flex flex-wrap items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
        <time v-if="adr.date" :datetime="adr.date">{{ formattedDate }}</time>

        <div class="flex items-center gap-2">
          <label for="status-select" class="text-sm font-medium text-gray-700 dark:text-gray-300">Status:</label>
          <select
            id="status-select"
            :value="adr.status"
            :aria-busy="updating"
            @change="onStatusChange"
            class="rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm px-2 py-1 text-gray-900 dark:text-gray-100 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none"
          >
            <option v-for="s in statuses" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
      </div>

      <div
        aria-live="polite"
        aria-atomic="true"
        class="mt-2 text-sm min-h-[1.25rem]"
      >
        <span
          v-if="feedback"
          :class="feedbackType === 'success' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
        >
          {{ feedback }}
        </span>
      </div>
    </header>

    <section
      class="prose dark:prose-invert max-w-none"
      v-html="renderedContent"
    ></section>
  </article>
</template>
