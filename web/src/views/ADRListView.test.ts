import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ADRListView from './ADRListView.vue'
import { fetchADRs } from '../api'

vi.mock('../api', () => ({
  fetchADRs: vi.fn(),
}))

const mockedFetchADRs = fetchADRs as ReturnType<typeof vi.fn>

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: ADRListView },
      { path: '/adr/:number', name: 'detail', component: { template: '<div />' } },
    ],
  })
}

async function mountView() {
  const router = makeRouter()
  router.push('/')
  await router.isReady()
  const wrapper = mount(ADRListView, { global: { plugins: [router] } })
  return wrapper
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ADRListView', () => {
  describe('loading state', () => {
    it('shows "Loading" before fetch resolves', () => {
      mockedFetchADRs.mockReturnValue(new Promise(() => {})) // never resolves
      const router = makeRouter()
      router.push('/')
      const wrapper = mount(ADRListView, { global: { plugins: [router] } })
      expect(wrapper.text()).toContain('Loading')
    })

    it('loading indicator has role="status" for screen readers', () => {
      mockedFetchADRs.mockReturnValue(new Promise(() => {}))
      const router = makeRouter()
      router.push('/')
      const wrapper = mount(ADRListView, { global: { plugins: [router] } })

      const loadingEl = wrapper.find('[role="status"]')
      expect(loadingEl.exists()).toBe(true)
      expect(loadingEl.text()).toContain('Loading')
    })
  })

  describe('error state', () => {
    it('shows error message on fetch rejection', async () => {
      mockedFetchADRs.mockRejectedValue(new Error('Network down'))
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Network down')
    })

    it('shows "Unknown error" for non-Error rejection', async () => {
      mockedFetchADRs.mockRejectedValue('something weird')
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Unknown error')
    })

    it('shows a retry button when fetch fails', async () => {
      mockedFetchADRs.mockRejectedValue(new Error('Network down'))
      const wrapper = await mountView()
      await flushPromises()

      const retryBtn = wrapper.find('button')
      expect(retryBtn.exists()).toBe(true)
      expect(retryBtn.text()).toBe('Retry')
    })

    it('clicking retry re-fetches and shows data on success', async () => {
      mockedFetchADRs.mockRejectedValueOnce(new Error('Network down'))
      const wrapper = await mountView()
      await flushPromises()
      expect(wrapper.text()).toContain('Network down')

      mockedFetchADRs.mockResolvedValueOnce([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      await wrapper.find('button').trigger('click')
      await flushPromises()

      expect(wrapper.text()).not.toContain('Network down')
      expect(wrapper.text()).toContain('Use PostgreSQL')
    })

    it('clicking retry shows loading state during re-fetch', async () => {
      mockedFetchADRs.mockRejectedValueOnce(new Error('Network down'))
      const wrapper = await mountView()
      await flushPromises()

      mockedFetchADRs.mockReturnValueOnce(new Promise(() => {}))
      await wrapper.find('button').trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Loading')
      expect(wrapper.text()).not.toContain('Network down')
    })
  })

  describe('empty state', () => {
    it('shows "No ADRs yet" when array is empty', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('No ADRs yet')
    })
  })

  describe('populated list', () => {
    const adrs = [
      { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
    ]

    beforeEach(() => {
      mockedFetchADRs.mockResolvedValue(adrs)
    })

    it('renders number, title, status, date for each ADR', async () => {
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('#1')
      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).toContain('Accepted')
      expect(wrapper.text()).toContain('2025-01-15')

      expect(wrapper.text()).toContain('#2')
      expect(wrapper.text()).toContain('Use Redis')
      expect(wrapper.text()).toContain('Proposed')
      expect(wrapper.text()).toContain('2025-02-01')
    })

    it('each link has an aria-label with ADR number and title', async () => {
      const wrapper = await mountView()
      await flushPromises()

      const links = wrapper.findAll('a')
      const link1 = links.find(l => l.attributes('href') === '/adr/1')!
      const link2 = links.find(l => l.attributes('href') === '/adr/2')!

      expect(link1.attributes('aria-label')).toBe('ADR #1: Use PostgreSQL')
      expect(link2.attributes('aria-label')).toBe('ADR #2: Use Redis')
    })

    it('links to /adr/{number} for each ADR', async () => {
      const wrapper = await mountView()
      await flushPromises()

      const links = wrapper.findAll('a')
      const hrefs = links.map(l => l.attributes('href'))
      expect(hrefs).toContain('/adr/1')
      expect(hrefs).toContain('/adr/2')
    })
  })

  describe('search', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('renders a search input', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()

      const input = wrapper.find('input[type="search"]')
      expect(input.exists()).toBe(true)
      expect(input.attributes('aria-label')).toBe('Search ADRs')
    })

    it('does NOT fetch with query when input is 1 character', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()

      const input = wrapper.find('input[type="search"]')
      await input.setValue('a')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(mockedFetchADRs).toHaveBeenCalledWith(undefined, expect.any(AbortSignal))
    })

    it('fetches with query param after 2+ chars and 300ms debounce', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue([])

      const input = wrapper.find('input[type="search"]')
      await input.setValue('chi')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(mockedFetchADRs).toHaveBeenCalledWith('chi', expect.any(AbortSignal))
    })

    it('debounces rapid input (only final value triggers fetch)', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue([])

      const input = wrapper.find('input[type="search"]')
      await input.setValue('ch')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(100)

      await input.setValue('chi')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      // Should only have fetched once, with the final value
      expect(mockedFetchADRs).toHaveBeenCalledTimes(1)
      expect(mockedFetchADRs).toHaveBeenCalledWith('chi', expect.any(AbortSignal))
    })

    it('shows "No matching ADRs" when search returns empty', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use Go', status: 'Accepted', date: '2025-01-01' },
      ])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue([])

      const input = wrapper.find('input[type="search"]')
      await input.setValue('zzz')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(wrapper.text()).toContain('No ADRs match')
      expect(wrapper.text()).toContain('zzz')
    })

    it('clearing search reloads all ADRs', async () => {
      const allADRs = [
        { number: 1, title: 'Use Go', status: 'Accepted', date: '2025-01-01' },
      ]
      mockedFetchADRs.mockResolvedValue(allADRs)
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue(allADRs)

      const input = wrapper.find('input[type="search"]')
      await input.setValue('')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(mockedFetchADRs).toHaveBeenCalledWith(undefined, expect.any(AbortSignal))
    })

    it('escape key clears search and reloads all ADRs', async () => {
      const allADRs = [
        { number: 1, title: 'Use Go', status: 'Accepted', date: '2025-01-01' },
      ]
      mockedFetchADRs.mockResolvedValue(allADRs)
      const wrapper = await mountView()
      await flushPromises()

      const input = wrapper.find('input[type="search"]')
      await input.setValue('chi')
      await input.trigger('input')
      mockedFetchADRs.mockClear()
      mockedFetchADRs.mockResolvedValue(allADRs)

      await input.trigger('keydown.esc')
      await flushPromises()

      expect((input.element as HTMLInputElement).value).toBe('')
      expect(mockedFetchADRs).toHaveBeenCalledWith(undefined, expect.any(AbortSignal))
    })

    it('error clears immediately when user starts typing new search', async () => {
      mockedFetchADRs.mockRejectedValue(new Error('Server error'))
      const wrapper = await mountView()
      await flushPromises()
      expect(wrapper.text()).toContain('Server error')

      mockedFetchADRs.mockResolvedValue([])
      const input = wrapper.find('input[type="search"]')
      await input.setValue('ch')
      await input.trigger('input')

      // Error should clear immediately, not after debounce
      expect(wrapper.text()).not.toContain('Server error')
    })
  })

  describe('search race condition prevention', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('cancels previous request when new search is triggered', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockReset()

      // First search: slow response
      const staleResult = [{ number: 1, title: 'Stale Result', status: 'Accepted', date: '2025-01-01' }]
      let resolveFirst!: (v: unknown) => void
      mockedFetchADRs.mockImplementationOnce(() => new Promise(r => { resolveFirst = r }))

      const input = wrapper.find('input[type="search"]')
      await input.setValue('abc')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      // Second search: fast response
      const freshResult = [{ number: 2, title: 'Fresh Result', status: 'Proposed', date: '2025-02-01' }]
      mockedFetchADRs.mockResolvedValueOnce(freshResult)

      await input.setValue('xyz')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      // Now the first resolves (stale)
      resolveFirst(staleResult)
      await flushPromises()

      // Only fresh result should be displayed
      expect(wrapper.text()).toContain('Fresh Result')
      expect(wrapper.text()).not.toContain('Stale Result')
    })

    it('does NOT show AbortError to user', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockReset()

      mockedFetchADRs.mockRejectedValueOnce(
        new DOMException('The operation was aborted', 'AbortError'),
      )

      const input = wrapper.find('input[type="search"]')
      await input.setValue('abc')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      expect(wrapper.text()).not.toContain('aborted')
      expect(wrapper.text()).not.toContain('AbortError')
      // Should not show any red error text
      expect(wrapper.find('.text-red-600').exists()).toBe(false)
    })

    it('does NOT fire request if unmounted before debounce completes', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const wrapper = await mountView()
      await flushPromises()
      mockedFetchADRs.mockClear()

      const input = wrapper.find('input[type="search"]')
      await input.setValue('abc')
      await input.trigger('input')

      // Unmount before debounce fires
      wrapper.unmount()

      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      // fetchADRs should NOT have been called after mount
      expect(mockedFetchADRs).not.toHaveBeenCalled()
    })
  })

  describe('status coloring', () => {
    it('green dot/text for "Accepted"', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Accepted', date: '2025-01-01' },
      ])
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-green-500').exists()).toBe(true)
      expect(wrapper.find('.text-green-600').exists()).toBe(true)
    })

    it('amber dot/text for "Proposed"', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Proposed', date: '2025-01-01' },
      ])
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-amber-500').exists()).toBe(true)
      expect(wrapper.find('.text-amber-600').exists()).toBe(true)
    })

    it('red dot/text for other statuses', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Superseded', date: '2025-01-01' },
      ])
      const wrapper = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-red-500').exists()).toBe(true)
      expect(wrapper.find('.text-red-600').exists()).toBe(true)
    })
  })
})
