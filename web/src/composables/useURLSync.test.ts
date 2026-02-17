import { ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { defineComponent } from 'vue'
import { useURLSync } from './useURLSync'
import type { SortField, SortDirection } from '../types'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
    ],
  })
}

function mountWithURLSync(initialPath = '/') {
  const router = makeRouter()
  router.push(initialPath)

  const searchQuery = ref('')
  const selectedStatuses = ref<Set<string>>(new Set())
  const sortField = ref<SortField>('number')
  const sortDirection = ref<SortDirection>('asc')

  let urlSync!: ReturnType<typeof useURLSync>

  const wrapper = mount(
    defineComponent({
      setup() {
        urlSync = useURLSync(searchQuery, selectedStatuses, sortField, sortDirection)
        return () => null
      },
    }),
    { global: { plugins: [router] } },
  )

  return { wrapper, router, searchQuery, selectedStatuses, sortField, sortDirection, urlSync: () => urlSync }
}

async function mountReady(initialPath = '/') {
  const result = mountWithURLSync(initialPath)
  const router = result.router
  router.push(initialPath)
  await router.isReady()

  // Re-mount with ready router
  const searchQuery = ref('')
  const selectedStatuses = ref<Set<string>>(new Set())
  const sortField = ref<SortField>('number')
  const sortDirection = ref<SortDirection>('asc')

  let urlSync!: ReturnType<typeof useURLSync>

  const wrapper = mount(
    defineComponent({
      setup() {
        urlSync = useURLSync(searchQuery, selectedStatuses, sortField, sortDirection)
        return () => null
      },
    }),
    { global: { plugins: [router] } },
  )

  return { wrapper, router, searchQuery, selectedStatuses, sortField, sortDirection, urlSync: () => urlSync }
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
  })
})
