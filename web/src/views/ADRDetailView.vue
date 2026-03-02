<script setup lang="ts">
import { ref, onMounted, nextTick, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import type { ADRDetail } from '../types'
import { fetchADR, fetchStatuses, NotFoundError } from '../api'
import { useStatusUpdate } from '../composables/useStatusUpdate'
import { useSupersede } from '../composables/useSupersede'
import { useRelation } from '../composables/useRelation'
import { useADRSearch } from '../composables/useADRSearch'
import SupersedeSelector from '../components/SupersedeSelector.vue'
import RelationInput from '../components/RelationInput.vue'

const props = defineProps<{ number: number }>()

const adr = ref<ADRDetail | null>(null)
const statuses = ref<string[]>([])
const loading = ref(true)
const error = ref('')
const notFound = ref(false)
const titleRef = ref<HTMLHeadingElement | null>(null)

const selectedStatus = ref('')

const showRelationInput = ref(false)
const addRelationBtnRef = ref<HTMLButtonElement | null>(null)

const {
  updating,
  feedback: statusFeedback,
  feedbackType: statusFeedbackType,
  doStatusUpdate,
  setPreviousStatus,
  getPreviousStatus,
} = useStatusUpdate(props.number)

const {
  pendingSuperseded,
  supersededBy,
  availableADRs,
  loadingADRs,
  startSupersede,
  cancelSupersede,
} = useSupersede(props.number)

const {
  adding,
  feedback: relationFeedback,
  feedbackType: relationFeedbackType,
  confirmRelation,
} = useRelation(props.number)

const search = useADRSearch()

const filteredRelationResults = computed(() =>
  search.adrs.value.filter(a => a.number !== props.number),
)

const relationSearching = computed(() =>
  search.loading.value && search.hasSearchQuery.value,
)

// Most-recent-wins feedback: show whichever was set most recently
const activeFeedback = computed(() => {
  if (relationFeedback.value) return relationFeedback.value
  if (statusFeedback.value) return statusFeedback.value
  return ''
})
const activeFeedbackType = computed(() => {
  if (relationFeedback.value) return relationFeedbackType.value
  if (statusFeedback.value) return statusFeedbackType.value
  return ''
})

const renderedContent = computed(() => {
  if (!adr.value?.content) return ''
  const raw = marked(adr.value.content) as string
  return DOMPurify.sanitize(raw)
})

const formattedDate = computed(() => {
  if (!adr.value?.date) return ''
  const d = new Date(adr.value.date + 'T00:00:00')
  return new Intl.DateTimeFormat(undefined, {
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
    selectedStatus.value = adrData.status
    setPreviousStatus(adrData.status)

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
  if (newStatus === getPreviousStatus()) {
    cancelSupersede()
    selectedStatus.value = getPreviousStatus()
    return
  }

  if (newStatus === 'Superseded') {
    showRelationInput.value = false
    await startSupersede()
    return
  }

  cancelSupersede()
  const updated = await doStatusUpdate(newStatus)
  if (updated) {
    adr.value = updated
    selectedStatus.value = updated.status
  } else {
    selectedStatus.value = getPreviousStatus()
  }
})

async function confirmSupersede() {
  if (supersededBy.value == null) return
  const updated = await doStatusUpdate('Superseded', { supersededBy: supersededBy.value })
  if (updated) {
    adr.value = updated
    selectedStatus.value = updated.status
    pendingSuperseded.value = false
  } else {
    selectedStatus.value = getPreviousStatus()
  }
}

function handleCancelSupersede() {
  cancelSupersede()
  selectedStatus.value = getPreviousStatus()
}

function onSelectorKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    handleCancelSupersede()
  } else if (event.key === 'Enter' && supersededBy.value != null) {
    event.preventDefault()
    confirmSupersede()
  }
}

function openRelationPanel() {
  cancelSupersede()
  selectedStatus.value = getPreviousStatus()
  showRelationInput.value = true
}

function closeRelationPanel() {
  showRelationInput.value = false
  search.searchQuery.value = ''
  nextTick(() => {
    addRelationBtnRef.value?.focus()
  })
}

function onRelationSearch(query: string) {
  search.searchQuery.value = query
  search.onSearchInput()
}

async function handleRelationSelect(targetNumber: number) {
  const updated = await confirmRelation(targetNumber)
  if (updated) {
    adr.value = updated
    closeRelationPanel()
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

        <button
          ref="addRelationBtnRef"
          v-if="!showRelationInput && !pendingSuperseded"
          :disabled="updating || adding"
          :aria-expanded="showRelationInput"
          class="text-sm text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 disabled:opacity-50"
          @click="openRelationPanel"
        >
          + Add relation
        </button>
      </div>

      <!-- Supersede selector -->
      <SupersedeSelector
        v-if="pendingSuperseded"
        :available-a-d-rs="availableADRs"
        :loading-a-d-rs="loadingADRs"
        :model-value="supersededBy"
        :disabled="updating"
        @confirm="confirmSupersede"
        @cancel="handleCancelSupersede"
        @update:model-value="supersededBy = $event"
        @keydown="onSelectorKeydown"
      />

      <!-- Relation input -->
      <RelationInput
        v-if="showRelationInput"
        :search-results="filteredRelationResults"
        :searching="relationSearching"
        :disabled="adding"
        @search="onRelationSearch"
        @select="handleRelationSelect"
        @cancel="closeRelationPanel"
      />

      <div
        aria-live="polite"
        aria-atomic="true"
        class="mt-2 text-sm min-h-[1.25rem]"
      >
        <span
          v-if="activeFeedback"
          :class="activeFeedbackType === 'success' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
        >
          {{ activeFeedback }}
        </span>
      </div>
    </header>

    <section
      class="prose dark:prose-invert max-w-none"
      v-html="renderedContent"
    ></section>
  </article>
</template>
