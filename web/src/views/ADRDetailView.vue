<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import type { ADRDetail, ADRSummary } from '../types'
import { fetchADR, fetchADRs, fetchStatuses, updateADRStatus, NotFoundError } from '../api'

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

const selectedStatus = ref('')
let previousStatus = ''
let feedbackTimer: ReturnType<typeof setTimeout> | undefined

// Supersede flow state
const pendingSuperseded = ref(false)
const supersededBy = ref<number | null>(null)
const availableADRs = ref<ADRSummary[]>([])
const loadingADRs = ref(false)
const supersedingSelectRef = ref<HTMLSelectElement | null>(null)
let supersedeFetchController: AbortController | null = null

const renderedContent = computed(() => {
  if (!adr.value?.content) return ''
  const raw = marked(adr.value.content) as string
  return DOMPurify.sanitize(raw)
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

onUnmounted(() => {
  if (feedbackTimer) {
    clearTimeout(feedbackTimer)
  }
  supersedeFetchController?.abort()
})

onMounted(async () => {
  try {
    const [adrData, statusData] = await Promise.all([
      fetchADR(props.number),
      fetchStatuses(),
    ])
    adr.value = adrData
    statuses.value = statusData
    selectedStatus.value = adrData.status
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

function statusDisplayText(s: string): string {
  return s === 'Superseded' ? 'Superseded\u2026' : s
}

watch(selectedStatus, async (newStatus) => {
  if (newStatus === previousStatus) {
    cancelSupersede()
    return
  }

  if (newStatus === 'Superseded') {
    supersedeFetchController?.abort()
    supersedeFetchController = new AbortController()
    const currentController = supersedeFetchController

    pendingSuperseded.value = true
    supersededBy.value = null
    loadingADRs.value = true
    try {
      const allADRs = await fetchADRs(undefined, currentController.signal)
      if (currentController !== supersedeFetchController) return
      availableADRs.value = allADRs.filter(a => a.number !== props.number)
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      if (currentController !== supersedeFetchController) return
      availableADRs.value = []
    } finally {
      if (currentController === supersedeFetchController) {
        loadingADRs.value = false
      }
    }
    await nextTick()
    supersedingSelectRef.value?.focus()
    return
  }

  // Non-superseded status change
  cancelSupersede()
  await doStatusUpdate(newStatus)
})

async function doStatusUpdate(newStatus: string, options?: { supersededBy?: number }) {
  updating.value = true
  feedback.value = ''

  if (feedbackTimer) {
    clearTimeout(feedbackTimer)
    feedbackTimer = undefined
  }

  try {
    const updated = await updateADRStatus(props.number, newStatus, options)
    adr.value = updated
    selectedStatus.value = updated.status
    previousStatus = updated.status
    feedbackType.value = 'success'
    feedback.value = `Status updated to ${updated.status}`
    feedbackTimer = setTimeout(() => { feedback.value = '' }, 4000)
    pendingSuperseded.value = false
  } catch (e) {
    selectedStatus.value = previousStatus
    feedbackType.value = 'error'
    feedback.value = e instanceof Error ? e.message : 'Failed to update status'
  } finally {
    updating.value = false
  }
}

function cancelSupersede() {
  supersedeFetchController?.abort()
  supersedeFetchController = null
  pendingSuperseded.value = false
  supersededBy.value = null
  availableADRs.value = []
  selectedStatus.value = previousStatus
}

async function confirmSupersede() {
  if (supersededBy.value == null) return
  await doStatusUpdate('Superseded', { supersededBy: supersededBy.value })
}

function onSelectorKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    cancelSupersede()
  } else if (event.key === 'Enter' && supersededBy.value != null) {
    event.preventDefault()
    confirmSupersede()
  }
}
</script>

<template>
  <!-- Loading -->
  <div v-if="loading" role="status" class="text-center py-16 text-gray-500 dark:text-gray-400">
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
            v-model="selectedStatus"
            :disabled="pendingSuperseded || updating"
            :aria-busy="updating"
            :aria-disabled="pendingSuperseded || undefined"
            class="rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm px-2 py-1 text-gray-900 dark:text-gray-100 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none disabled:opacity-50"
          >
            <option v-for="s in statuses" :key="s" :value="s">{{ statusDisplayText(s) }}</option>
          </select>
          <span v-if="pendingSuperseded" class="sr-only">Status dropdown disabled while selecting superseding ADR</span>
        </div>
      </div>

      <!-- Supersede selector -->
      <div
        v-if="pendingSuperseded"
        role="group"
        aria-labelledby="supersede-label"
        class="mt-3 ml-4 p-3 border-l-2 border-blue-500 bg-blue-50 dark:bg-blue-900/10 rounded-r"
        @keydown="onSelectorKeydown"
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
            ref="supersedingSelectRef"
            v-model.number="supersededBy"
            class="rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm px-2 py-1 text-gray-900 dark:text-gray-100 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none w-full"
          >
            <option :value="null" disabled>— Choose an ADR —</option>
            <option v-for="a in availableADRs" :key="a.number" :value="a.number">
              ADR-{{ String(a.number).padStart(4, '0') }}: {{ a.title }} ({{ a.status }})
            </option>
          </select>
        </div>
        <div class="mt-2 flex gap-2">
          <button
            :disabled="supersededBy == null || updating"
            class="px-3 py-1 text-sm rounded bg-blue-600 text-white hover:bg-blue-700 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
            @click="confirmSupersede"
          >
            Confirm
          </button>
          <button
            :disabled="updating"
            class="px-3 py-1 text-sm rounded border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:outline-none"
            @click="cancelSupersede"
          >
            Cancel
          </button>
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
