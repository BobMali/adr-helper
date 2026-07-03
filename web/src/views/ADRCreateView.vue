<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { fetchConfig, fetchTemplateSections } from '../api'
import { useCreateADR } from '../composables/useCreateADR'
import type { TemplateSectionDef } from '../types'

const router = useRouter()
const { title, sections, submitting, submitError, sectionErrors, submit } = useCreateADR()

const templateName = ref('')
const sectionDefs = ref<TemplateSectionDef[]>([])
const configLoading = ref(true)
const configError = ref('')
const titleInputRef = ref<HTMLInputElement | null>(null)

onMounted(async () => {
  try {
    const [config, templateSections] = await Promise.all([
      fetchConfig(),
      fetchTemplateSections(),
    ])
    templateName.value = config.template
    sectionDefs.value = templateSections
  } catch (e) {
    configError.value = e instanceof Error ? e.message : 'Failed to load config'
  } finally {
    configLoading.value = false
    await nextTick()
    titleInputRef.value?.focus()
  }
})

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
  Promise.all([fetchConfig(), fetchTemplateSections()])
    .then(([config, templateSections]) => {
      templateName.value = config.template
      sectionDefs.value = templateSections
    })
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
        <textarea
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
