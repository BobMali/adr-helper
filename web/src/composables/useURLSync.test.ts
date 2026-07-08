import { ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { defineComponent } from 'vue'
import { useURLSync } from './useURLSync'
import type { SortField, SortDirection, MetaMatchMode } from '../types'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
    ],
  })
}

async function mountReady(initialPath = '/') {
  const router = makeRouter()
  router.push(initialPath)
  await router.isReady()

  const searchQuery = ref('')
  const selectedStatuses = ref<Set<string>>(new Set())
  const sortField = ref<SortField>('number')
  const sortDirection = ref<SortDirection>('asc')
  const selectedMeta = ref<Record<string, Set<string>>>({})
  const matchMode = ref<MetaMatchMode>('any')

  let urlSync!: ReturnType<typeof useURLSync>

  const wrapper = mount(
    defineComponent({
      setup() {
        urlSync = useURLSync({ searchQuery, selectedStatuses, sortField, sortDirection, selectedMeta, matchMode })
        return () => null
      },
    }),
    { global: { plugins: [router] } },
  )

  return { wrapper, router, searchQuery, selectedStatuses, sortField, sortDirection, selectedMeta, matchMode, urlSync: () => urlSync }
}

describe('useURLSync', () => {
  describe('initFromURL', () => {
    it('reads search query from URL', async () => {
      const { searchQuery, urlSync } = await mountReady('/?q=test')
      urlSync().initFromURL()
      expect(searchQuery.value).toBe('test')
    })

    it('reads status filters from URL', async () => {
      const { selectedStatuses, urlSync } = await mountReady('/?status=Accepted')
      urlSync().initFromURL()
      expect(selectedStatuses.value.has('Accepted')).toBe(true)
    })

    it('reads sort field from URL', async () => {
      const { sortField, urlSync } = await mountReady('/?sort=title')
      urlSync().initFromURL()
      expect(sortField.value).toBe('title')
    })

    it('reads sort direction from URL', async () => {
      const { sortDirection, urlSync } = await mountReady('/?dir=desc')
      urlSync().initFromURL()
      expect(sortDirection.value).toBe('desc')
    })

    it('ignores invalid sort field', async () => {
      const { sortField, urlSync } = await mountReady('/?sort=bogus')
      urlSync().initFromURL()
      expect(sortField.value).toBe('number')
    })

    it('defaults direction to asc when not specified', async () => {
      const { sortDirection, urlSync } = await mountReady('/')
      urlSync().initFromURL()
      expect(sortDirection.value).toBe('asc')
    })

    it('reads a single meta facet value from URL', async () => {
      const { selectedMeta, urlSync } = await mountReady('/?meta_scope=backend')
      urlSync().initFromURL()
      expect([...(selectedMeta.value.scope ?? [])]).toEqual(['backend'])
    })

    it('reads multiple meta facet values from URL', async () => {
      const { selectedMeta, urlSync } = await mountReady('/?meta_scope=backend&meta_scope=api')
      urlSync().initFromURL()
      expect([...(selectedMeta.value.scope ?? [])].sort()).toEqual(['api', 'backend'])
    })

    it('reads match mode from URL', async () => {
      const { matchMode, urlSync } = await mountReady('/?match=all')
      urlSync().initFromURL()
      expect(matchMode.value).toBe('all')
    })

    it('defaults match mode to any', async () => {
      const { matchMode, urlSync } = await mountReady('/?meta_scope=backend')
      urlSync().initFromURL()
      expect(matchMode.value).toBe('any')
    })

    it('keeps bookmarked meta values without needing the facet list loaded', async () => {
      // initFromURL runs before fetchMetaFields resolves; values must not be dropped.
      const { selectedMeta, urlSync } = await mountReady('/?meta_scope=backend')
      urlSync().initFromURL()
      expect(selectedMeta.value.scope?.has('backend')).toBe(true)
    })
  })

  describe('syncToURL', () => {
    it('writes search query to URL', async () => {
      const { searchQuery, router, urlSync } = await mountReady('/')
      searchQuery.value = 'test'
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.q).toBe('test')
    })

    it('does not write short queries to URL', async () => {
      const { searchQuery, router, urlSync } = await mountReady('/')
      searchQuery.value = 'a'
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.q).toBeUndefined()
    })

    it('writes status to URL', async () => {
      const { selectedStatuses, router, urlSync } = await mountReady('/')
      selectedStatuses.value = new Set(['Accepted'])
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.status).toBe('Accepted')
    })

    it('writes sort field to URL', async () => {
      const { sortField, router, urlSync } = await mountReady('/')
      sortField.value = 'title'
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.sort).toBe('title')
    })

    it('omits default sort/dir from URL', async () => {
      const { router, urlSync } = await mountReady('/')
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.sort).toBeUndefined()
      expect(router.currentRoute.value.query.dir).toBeUndefined()
    })

    it('writes meta facet selection to URL as meta_<key>', async () => {
      const { selectedMeta, router, urlSync } = await mountReady('/')
      selectedMeta.value = { scope: new Set(['backend']) }
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.meta_scope).toBe('backend')
    })

    it('writes match=all only when mode is all', async () => {
      const { selectedMeta, matchMode, router, urlSync } = await mountReady('/')
      selectedMeta.value = { scope: new Set(['backend', 'api']) }
      matchMode.value = 'all'
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.match).toBe('all')
    })

    it('omits match from URL when mode is any', async () => {
      const { selectedMeta, router, urlSync } = await mountReady('/')
      selectedMeta.value = { scope: new Set(['backend']) }
      urlSync().syncToURL()
      await flushPromises()

      expect(router.currentRoute.value.query.match).toBeUndefined()
    })

    it('round-trips meta + match through the URL', async () => {
      const { selectedMeta, matchMode, urlSync } = await mountReady('/')
      selectedMeta.value = { scope: new Set(['backend', 'api']) }
      matchMode.value = 'all'
      urlSync().syncToURL()
      await flushPromises()

      urlSync().initFromURL()
      expect([...(selectedMeta.value.scope ?? [])].sort()).toEqual(['api', 'backend'])
      expect(matchMode.value).toBe('all')
    })
  })
})
