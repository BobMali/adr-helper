<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { fetchConfig, fetchTemplateSections, fetchScopes, addScope } from '../api'
import { useCreateADR } from '../composables/useCreateADR'
import type { TemplateSectionDef } from '../types'

const router = useRouter()
const { title, sections, submitting, submitError, sectionErrors, submit } = useCreateADR()

const templateName = ref('')
const sectionDefs = ref<TemplateSectionDef[]>([])
const configLoading = ref(true)
const configError = ref('')
const titleInputRef = ref<HTMLInputElement | null>(null)

// Above this many scopes, the checkbox list gets a scroll cap and a filter input.
const SCOPE_FILTER_THRESHOLD = 8

// Scope vocabulary state (used by fields with def.vocabulary).
const scopeOptions = ref<string[]>([])
const scopeSelection = ref<Record<string, string[]>>({})
const newScope = ref('')
const scopeAddError = ref('')
const addingScope = ref(false)
// scopeFilter/scopeOptions are shared across all vocabulary fields; only one
// exists today (see def.vocabulary).
const scopeFilter = ref('')
const filteredScopeOptions = computed(() => {
  const q = scopeFilter.value.trim().toLowerCase()
  return q ? scopeOptions.value.filter((s) => s.toLowerCase().includes(q)) : scopeOptions.value
})

async function loadForm() {
  const [config, templateSections, scopes] = await Promise.all([
    fetchConfig(),
    fetchTemplateSections(),
    // A scopes failure must not break the form; treat as empty vocabulary.
    fetchScopes().catch(() => [] as string[]),
  ])
  templateName.value = config.template
  sectionDefs.value = templateSections
  scopeOptions.value = scopes
}

onMounted(async () => {
  try {
    await loadForm()
  } catch (e) {
    configError.value = e instanceof Error ? e.message : 'Failed to load config'
  } finally {
    configLoading.value = false
    await nextTick()
    titleInputRef.value?.focus()
  }
})

function setScopeSelection(key: string, next: string[]) {
  scopeSelection.value[key] = next
  sections.value[key] = next.join(', ')
}

function toggleScope(key: string, value: string, checked: boolean) {
  const current = scopeSelection.value[key] ?? []
  setScopeSelection(key, checked ? [...current, value] : current.filter((v) => v !== value))
}

function clearScopes(key: string) {
  setScopeSelection(key, [])
}

// Selected scopes not currently visible under the active filter.
function hiddenSelectedCount(key: string): number {
  const visible = new Set(filteredScopeOptions.value)
  return (scopeSelection.value[key] ?? []).filter((s) => !visible.has(s)).length
}

async function handleAddScope(key: string) {
  const value = newScope.value.trim()
  if (!value) return
  scopeAddError.value = ''
  addingScope.value = true
  try {
    const updated = await addScope(value)
    scopeOptions.value = updated
    // Select the newly added value using the server's canonical spelling.
    const canonical = updated.find((s) => s.toLowerCase() === value.toLowerCase()) ?? value
    if (!(scopeSelection.value[key] ?? []).includes(canonical)) {
      toggleScope(key, canonical, true)
    }
    newScope.value = ''
    // Clear any active filter so the just-added (auto-checked) scope is visible.
    scopeFilter.value = ''
  } catch (e) {
    scopeAddError.value = e instanceof Error ? e.message : 'Failed to add scope'
  } finally {
    addingScope.value = false
  }
}

async function handleSubmit() {
  const result = await submit(sectionDefs.value)
  if (!result) {
    // Focus first error field
    await nextTick()
    if (submitError.value) {
      titleInputRef.value?.focus()
    } else {
      const firstErrorKey = Object.keys(sectionErrors.value)[0]
      if (firstErrorKey) {
        const el = document.getElementById(`section-${firstErrorKey}`)
        el?.focus()
      }
    }
    return
  }
  router.push({ name: 'detail', params: { number: result.number }, query: { created: 'true' } })
}

function retryLoad() {
  configError.value = ''
  configLoading.value = true
  loadForm()
    .catch((e) => {
      configError.value = e instanceof Error ? e.message : 'Failed to load config'
    })
    .finally(() => {
      configLoading.value = false
    })
}
</script>

<template>
  <nav class="mb-6">
    <RouterLink
      to="/"
      class="text-sm text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    >
      &larr; Back to list
    </RouterLink>
  </nav>

  <!-- Loading config -->
  <div v-if="configLoading" role="status" class="text-center py-16 text-gray-500 dark:text-gray-400">
    Loading&hellip;
  </div>

  <!-- Config error -->
  <div v-else-if="configError" class="text-center py-16">
    <p class="text-red-600 dark:text-red-400">{{ configError }}</p>
    <button
      class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
      @click="retryLoad"
    >
      Retry
    </button>
    <RouterLink
      to="/"
      class="mt-4 ml-4 inline-block text-blue-600 dark:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
    >
      &larr; Back to list
    </RouterLink>
  </div>

  <!-- Create form -->
  <div v-else>
    <header class="mb-6">
      <h1 class="text-2xl font-semibold tracking-tight">New Architecture Decision Record</h1>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Template: <span class="font-medium">{{ templateName }}</span>
        <span class="ml-1 text-xs text-gray-400 dark:text-gray-500" title="Set in project config">(project config)</span>
      </p>
    </header>

    <form @submit.prevent="handleSubmit" class="max-w-2xl">
      <div class="mb-6">
        <label for="adr-title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Title <span class="text-red-500" aria-hidden="true">*</span>
        </label>
        <input
          id="adr-title"
          ref="titleInputRef"
          v-model="title"
          type="text"
          required
          aria-required="true"
          :disabled="submitting"
          class="w-full py-2.5 px-4 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          placeholder="e.g. Use PostgreSQL for persistence"
        />
        <p
          v-if="submitError"
          role="alert"
          class="mt-1 text-sm text-red-600 dark:text-red-400"
        >
          {{ submitError }}
        </p>
      </div>

      <!-- Section fields -->
      <fieldset
        v-for="def in sectionDefs"
        :key="def.key"
        class="mb-6"
      >
        <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ def.heading }}
          <span v-if="!def.optional" class="text-red-500" aria-hidden="true">*</span>
          <span v-if="def.optional" class="ml-1 text-xs text-gray-400 dark:text-gray-500">optional</span>
        </legend>
        <p
          :id="`section-help-${def.key}`"
          class="text-xs text-gray-400 dark:text-gray-500 mb-1"
        >
          {{ def.placeholder }}
        </p>

        <!-- Vocabulary field: checkbox group of selectable values + inline add -->
        <div
          v-if="def.vocabulary"
          :id="`section-${def.key}`"
          tabindex="-1"
          :aria-describedby="`section-help-${def.key}`"
          class="focus:outline-none"
        >
          <p v-if="scopeOptions.length === 0" class="text-sm text-gray-500 dark:text-gray-400 mb-2">
            No scopes yet &mdash; add one below.
          </p>
          <template v-else>
            <!-- Filter (large vocabularies only) -->
            <input
              v-if="scopeOptions.length > SCOPE_FILTER_THRESHOLD"
              v-model="scopeFilter"
              type="text"
              :aria-label="`Filter ${def.heading.toLowerCase()}s`"
              placeholder="Filter scopes&hellip;"
              :disabled="submitting"
              class="w-full mb-2 py-2 px-3 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              @keydown.enter.prevent
            />
            <div v-if="scopeOptions.length > SCOPE_FILTER_THRESHOLD" class="sr-only" aria-live="polite">
              {{ scopeFilter.trim() ? `${filteredScopeOptions.length} of ${scopeOptions.length} scopes shown` : '' }}
            </div>

            <!-- Checkbox list (scroll-capped above the threshold) -->
            <p
              v-if="filteredScopeOptions.length === 0"
              class="text-sm text-gray-500 dark:text-gray-400 mb-3"
            >
              No scopes match "{{ scopeFilter }}".
            </p>
            <div
              v-else
              class="flex flex-wrap gap-x-4 gap-y-2 mb-3"
              :class="scopeOptions.length > SCOPE_FILTER_THRESHOLD
                ? 'max-h-40 overflow-y-auto rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-2'
                : ''"
            >
              <label
                v-for="opt in filteredScopeOptions"
                :key="opt"
                class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"
              >
                <input
                  type="checkbox"
                  :value="opt"
                  :checked="(scopeSelection[def.key] ?? []).includes(opt)"
                  :disabled="submitting"
                  class="rounded border-gray-300 dark:border-gray-700 text-blue-600 focus:ring-blue-500"
                  @change="toggleScope(def.key, opt, ($event.target as HTMLInputElement).checked)"
                />
                {{ opt }}
              </label>
            </div>

            <!-- Selection summary + clear-all -->
            <div
              v-if="(scopeSelection[def.key] ?? []).length > 0"
              class="flex items-center gap-2 mb-3 text-sm text-gray-500 dark:text-gray-400"
            >
              <span>
                {{ (scopeSelection[def.key] ?? []).length }} selected<template
                  v-if="hiddenSelectedCount(def.key) > 0"
                > &middot; {{ hiddenSelectedCount(def.key) }} hidden by filter</template>
              </span>
              <button
                type="button"
                :disabled="submitting"
                class="px-2 py-1 text-xs rounded border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
                @click="clearScopes(def.key)"
              >
                Clear all
              </button>
            </div>
          </template>
          <div class="flex items-center gap-2">
            <input
              v-model="newScope"
              type="text"
              :disabled="addingScope || submitting"
              :aria-label="`Add a new ${def.heading.toLowerCase()}`"
              placeholder="Add a new scope&hellip;"
              class="flex-1 py-2 px-3 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              @keydown.enter.prevent="handleAddScope(def.key)"
            />
            <button
              type="button"
              :disabled="addingScope || submitting || !newScope.trim()"
              class="px-3 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
              @click="handleAddScope(def.key)"
            >
              Add
            </button>
          </div>
          <p
            v-if="scopeAddError"
            role="alert"
            class="mt-1 text-sm text-red-600 dark:text-red-400"
          >
            {{ scopeAddError }}
          </p>
        </div>

        <textarea
          v-else
          :id="`section-${def.key}`"
          v-model="sections[def.key]"
          :aria-required="!def.optional || undefined"
          :aria-describedby="`section-help-${def.key}`"
          :disabled="submitting"
          rows="4"
          class="w-full py-2.5 px-4 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 resize-y min-h-[6rem]"
        />
        <p
          v-if="sectionErrors[def.key]"
          role="alert"
          class="mt-1 text-sm text-red-600 dark:text-red-400"
        >
          {{ sectionErrors[def.key] }}
        </p>
      </fieldset>

      <div class="sticky bottom-0 bg-white dark:bg-gray-950 py-4 border-t border-gray-200 dark:border-gray-800 flex items-center gap-3">
        <RouterLink
          to="/"
          class="px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
        >
          Cancel
        </RouterLink>
        <button
          type="submit"
          :disabled="submitting"
          :aria-busy="submitting"
          class="px-4 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 disabled:opacity-50"
        >
          {{ submitting ? 'Creating\u2026' : 'Create ADR' }}
        </button>
      </div>
    </form>
  </div>
</template>
