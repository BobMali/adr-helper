import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ADRListView from './ADRListView.vue'
import { fetchADRs, fetchStatuses } from '../api'

vi.mock('../api', () => ({
  fetchADRs: vi.fn(),
  fetchStatuses: vi.fn(),
}))

const mockedFetchADRs = fetchADRs as ReturnType<typeof vi.fn>
const mockedFetchStatuses = fetchStatuses as ReturnType<typeof vi.fn>

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: ADRListView },
      { path: '/adr/:number', name: 'detail', component: { template: '<div />' } },
    ],
  })
}

async function mountView(initialPath = '/') {
  const router = makeRouter()
  router.push(initialPath)
  await router.isReady()
  const wrapper = mount(ADRListView, { global: { plugins: [router] } })
  return { wrapper, router }
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ADRListView', () => {
  beforeEach(() => {
    mockedFetchStatuses.mockResolvedValue(['Accepted', 'Proposed', 'Rejected', 'Deprecated', 'Superseded'])
  })

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

      const statusEls = wrapper.findAll('[role="status"]')
      const loadingEl = statusEls.find(el => el.text().includes('Loading'))
      expect(loadingEl).toBeDefined()
      expect(loadingEl!.text()).toContain('Loading')
    })
  })

  describe('error state', () => {
    it('shows error message on fetch rejection', async () => {
      mockedFetchADRs.mockRejectedValue(new Error('Network down'))
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Network down')
    })

    it('shows "Unknown error" for non-Error rejection', async () => {
      mockedFetchADRs.mockRejectedValue('something weird')
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Unknown error')
    })

    it('shows a retry button when fetch fails', async () => {
      mockedFetchADRs.mockRejectedValue(new Error('Network down'))
      const { wrapper } = await mountView()
      await flushPromises()

      const buttons = wrapper.findAll('button')
      const retryBtn = buttons.find(b => b.text() === 'Retry')
      expect(retryBtn).toBeDefined()
      expect(retryBtn!.text()).toBe('Retry')
    })

    it('clicking retry re-fetches and shows data on success', async () => {
      mockedFetchADRs.mockRejectedValueOnce(new Error('Network down'))
      const { wrapper } = await mountView()
      await flushPromises()
      expect(wrapper.text()).toContain('Network down')

      mockedFetchADRs.mockResolvedValueOnce([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      const retryBtn = wrapper.findAll('button').find(b => b.text() === 'Retry')!
      await retryBtn.trigger('click')
      await flushPromises()

      expect(wrapper.text()).not.toContain('Network down')
      expect(wrapper.text()).toContain('Use PostgreSQL')
    })

    it('clicking retry shows loading state during re-fetch', async () => {
      mockedFetchADRs.mockRejectedValueOnce(new Error('Network down'))
      const { wrapper } = await mountView()
      await flushPromises()

      mockedFetchADRs.mockReturnValueOnce(new Promise(() => {}))
      const retryBtn = wrapper.findAll('button').find(b => b.text() === 'Retry')!
      await retryBtn.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Loading')
      expect(wrapper.text()).not.toContain('Network down')
    })
  })

  describe('empty state', () => {
    it('shows "No ADRs yet" when array is empty', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
      await flushPromises()

      const links = wrapper.findAll('a')
      const link1 = links.find(l => l.attributes('href') === '/adr/1')!
      const link2 = links.find(l => l.attributes('href') === '/adr/2')!

      expect(link1.attributes('aria-label')).toBe('ADR #1: Use PostgreSQL')
      expect(link2.attributes('aria-label')).toBe('ADR #2: Use Redis')
    })

    it('links to /adr/{number} for each ADR', async () => {
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
      await flushPromises()

      const input = wrapper.find('input[type="search"]')
      expect(input.exists()).toBe(true)
      expect(input.attributes('aria-label')).toBe('Search ADRs')
    })

    it('does NOT fetch with query when input is 1 character', async () => {
      mockedFetchADRs.mockResolvedValue([])
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
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
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-green-500').exists()).toBe(true)
      expect(wrapper.find('.text-green-600').exists()).toBe(true)
    })

    it('amber dot/text for "Proposed"', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Proposed', date: '2025-01-01' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-amber-500').exists()).toBe(true)
      expect(wrapper.find('.text-amber-600').exists()).toBe(true)
    })

    it('red dot/text for other statuses', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Superseded', date: '2025-01-01' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.find('.bg-red-500').exists()).toBe(true)
      expect(wrapper.find('.text-red-600').exists()).toBe(true)
    })
  })

  describe('status filtering', () => {
    const mixedADRs = [
      { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
      { number: 3, title: 'Use MongoDB', status: 'Rejected', date: '2025-03-01' },
    ]

    beforeEach(() => {
      mockedFetchADRs.mockResolvedValue(mixedADRs)
    })

    it('shows all ADRs when no status chip is selected', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).toContain('Use Redis')
      expect(wrapper.text()).toContain('Use MongoDB')
    })

    it('clicking a chip filters list to only that status', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      await acceptedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).not.toContain('Use Redis')
      expect(wrapper.text()).not.toContain('Use MongoDB')
    })

    it('multiple chips use OR logic (shows ADRs matching any selected status)', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      const proposedChip = chips.find(b => b.text().includes('Proposed'))!
      await acceptedChip.trigger('click')
      await proposedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).toContain('Use Redis')
      expect(wrapper.text()).not.toContain('Use MongoDB')
    })

    it('deselecting all chips shows all ADRs again', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!

      // Select then deselect
      await acceptedChip.trigger('click')
      await flushPromises()
      expect(wrapper.text()).not.toContain('Use Redis')

      await acceptedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).toContain('Use Redis')
      expect(wrapper.text()).toContain('Use MongoDB')
    })

    it('shows filter empty state when chips exclude all results', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      mockedFetchStatuses.mockResolvedValue(['Accepted', 'Proposed', 'Rejected'])
      const { wrapper } = await mountView()
      await flushPromises()

      // Select 'Proposed' — no ADRs have this status
      const chips = wrapper.find('[role="group"]').findAll('button')
      const proposedChip = chips.find(b => b.text().includes('Proposed'))!
      await proposedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('No ADRs match the selected filters')
      const statusEl = wrapper.find('[role="status"]')
      expect(statusEl.exists()).toBe(true)
    })
  })

  describe('filter + search composition', () => {
    beforeEach(() => {
      vi.useFakeTimers()
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('search narrows via API, chip filters client-side (AND logic)', async () => {
      // Initial load
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
        { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      // Search returns both (server-side)
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
        { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
      ])
      const input = wrapper.find('input[type="search"]')
      await input.setValue('Use')
      await input.trigger('input')
      await vi.advanceTimersByTimeAsync(300)
      await flushPromises()

      // Now filter by Accepted chip
      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      await acceptedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).not.toContain('Use Redis')
    })

    it('search returns results but filter excludes all → shows filter empty state', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      // Select Proposed — no Proposed ADRs in results
      const chips = wrapper.find('[role="group"]').findAll('button')
      const proposedChip = chips.find(b => b.text().includes('Proposed'))!
      await proposedChip.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('No ADRs match the selected filters')
      expect(wrapper.text()).not.toContain('No ADRs match "')
    })
  })

  describe('sorting', () => {
    const sortTestADRs = [
      { number: 3, title: 'Use MongoDB', status: 'Rejected', date: '2025-03-01' },
      { number: 1, title: 'Adopt TypeScript', status: 'Accepted', date: '2025-01-15' },
      { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
    ]

    beforeEach(() => {
      mockedFetchADRs.mockResolvedValue(sortTestADRs)
    })

    it('default sort is number ascending with ID button active', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      expect(sortGroup.exists()).toBe(true)

      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!
      expect(idBtn.attributes('aria-pressed')).toBe('true')

      // Items should be in number order: 1, 2, 3
      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('#1')
      expect(items[1]!.text()).toContain('#2')
      expect(items[2]!.text()).toContain('#3')
    })

    it('sort by title ascending', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      await titleBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('Adopt TypeScript')
      expect(items[1]!.text()).toContain('Use MongoDB')
      expect(items[2]!.text()).toContain('Use Redis')
    })

    it('sort by title descending', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      // Click once for asc, click again for desc
      await titleBtn.trigger('click')
      await titleBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('Use Redis')
      expect(items[1]!.text()).toContain('Use MongoDB')
      expect(items[2]!.text()).toContain('Adopt TypeScript')
    })

    it('sort by status uses lifecycle order (Proposed → Accepted → Deprecated → Superseded → Rejected)', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const statusBtn = sortGroup.findAll('button').find(b => b.text().includes('Status'))!
      await statusBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      // Proposed(2) → Accepted(1) → Rejected(3)
      expect(items[0]!.text()).toContain('Proposed')
      expect(items[1]!.text()).toContain('Accepted')
      expect(items[2]!.text()).toContain('Rejected')
    })

    it('clicking active field toggles direction', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!

      // Click ID (already active, asc) → desc
      await idBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('#3')
      expect(items[1]!.text()).toContain('#2')
      expect(items[2]!.text()).toContain('#1')
    })

    it('clicking new field resets direction to asc', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!

      // Toggle ID to desc
      await idBtn.trigger('click')
      await flushPromises()

      // Switch to Title → should reset to asc
      await titleBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('Adopt TypeScript')
      expect(items[2]!.text()).toContain('Use Redis')
    })

    it('sort applies after status filtering', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 3, title: 'Use MongoDB', status: 'Accepted', date: '2025-03-01' },
        { number: 1, title: 'Adopt TypeScript', status: 'Accepted', date: '2025-01-15' },
        { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      // Filter to Accepted only
      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      await acceptedChip.trigger('click')
      await flushPromises()

      // Sort by title
      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      await titleBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items).toHaveLength(2)
      expect(items[0]!.text()).toContain('Adopt TypeScript')
      expect(items[1]!.text()).toContain('Use MongoDB')
    })

    it('toggling a status chip does not reset sort state', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      // Sort by title first
      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      await titleBtn.trigger('click')
      await flushPromises()

      // Toggle a status chip
      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      await acceptedChip.trigger('click')
      await flushPromises()

      // Title button should still be active
      expect(titleBtn.attributes('aria-pressed')).toBe('true')
    })

    it('unknown status sorts to end without crashing', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'A', status: 'Accepted', date: '2025-01-01' },
        { number: 2, title: 'B', status: 'UnknownStatus', date: '2025-01-02' },
        { number: 3, title: 'C', status: 'Proposed', date: '2025-01-03' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const statusBtn = sortGroup.findAll('button').find(b => b.text().includes('Status'))!
      await statusBtn.trigger('click')
      await flushPromises()

      const items = wrapper.findAll('li')
      // Proposed, Accepted, then UnknownStatus at end
      expect(items[0]!.text()).toContain('Proposed')
      expect(items[1]!.text()).toContain('Accepted')
      expect(items[2]!.text()).toContain('UnknownStatus')
    })

    it('sr-only count reflects correct count after sort change', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      await titleBtn.trigger('click')
      await flushPromises()

      const srOnly = wrapper.find('.sr-only[role="status"]')
      expect(srOnly.text()).toContain('3 records shown')
    })

    it('inactive button has aria-pressed="false"', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      expect(titleBtn.attributes('aria-pressed')).toBe('false')
    })

    it('active button has descriptive aria-label', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!
      expect(idBtn.attributes('aria-label')).toContain('ascending')
      expect(idBtn.attributes('aria-label')).toContain('descending')
    })

    it('shows direction arrow on active button', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!

      // Should have an arrow (↑ for asc)
      expect(idBtn.text()).toMatch(/[↑↓]/)
    })
  })

  describe('sorting URL sync', () => {
    const sortTestADRs = [
      { number: 3, title: 'Use MongoDB', status: 'Rejected', date: '2025-03-01' },
      { number: 1, title: 'Adopt TypeScript', status: 'Accepted', date: '2025-01-15' },
      { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
    ]

    beforeEach(() => {
      mockedFetchADRs.mockResolvedValue(sortTestADRs)
    })

    it('sort change updates URL params', async () => {
      const { wrapper, router } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      await titleBtn.trigger('click')
      await flushPromises()

      expect(router.currentRoute.value.query.sort).toBe('title')
    })

    it('defaults produce no extra sort/dir params', async () => {
      const { router } = await mountView()
      await flushPromises()

      expect(router.currentRoute.value.query.sort).toBeUndefined()
      expect(router.currentRoute.value.query.dir).toBeUndefined()
    })

    it('desc direction adds dir param', async () => {
      const { wrapper, router } = await mountView()
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!
      await idBtn.trigger('click') // toggle to desc
      await flushPromises()

      expect(router.currentRoute.value.query.dir).toBe('desc')
    })

    it('mount with ?sort=title&dir=desc restores sort state', async () => {
      const { wrapper } = await mountView('/?sort=title&dir=desc')
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const titleBtn = sortGroup.findAll('button').find(b => b.text().includes('Title'))!
      expect(titleBtn.attributes('aria-pressed')).toBe('true')

      // Items should be title desc: Use Redis, Use MongoDB, Adopt TypeScript
      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('Use Redis')
      expect(items[2]!.text()).toContain('Adopt TypeScript')
    })

    it('mount with ?sort=title but no dir defaults to asc', async () => {
      const { wrapper } = await mountView('/?sort=title')
      await flushPromises()

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('Adopt TypeScript')
      expect(items[2]!.text()).toContain('Use Redis')
    })

    it('invalid ?sort=bogus falls back to default (number asc)', async () => {
      const { wrapper } = await mountView('/?sort=bogus')
      await flushPromises()

      const sortGroup = wrapper.find('[aria-label="Sort options"]')
      const idBtn = sortGroup.findAll('button').find(b => b.text().includes('ID'))!
      expect(idBtn.attributes('aria-pressed')).toBe('true')

      const items = wrapper.findAll('li')
      expect(items[0]!.text()).toContain('#1')
    })
  })

  describe('URL state', () => {
    it('on mount with ?status=Accepted, chip is pre-selected and list is filtered', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
        { number: 2, title: 'Use Redis', status: 'Proposed', date: '2025-02-01' },
      ])
      const { wrapper } = await mountView('/?status=Accepted')
      await flushPromises()

      // Accepted chip should be pressed
      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      expect(acceptedChip.attributes('aria-pressed')).toBe('true')

      // Only Accepted ADR shown
      expect(wrapper.text()).toContain('Use PostgreSQL')
      expect(wrapper.text()).not.toContain('Use Redis')
    })

    it('toggling a chip updates route query params', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      const { wrapper, router } = await mountView()
      await flushPromises()

      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      await acceptedChip.trigger('click')
      await flushPromises()

      expect(router.currentRoute.value.query.status).toBe('Accepted')
    })

    it('on mount with ?q=database&status=Accepted, both search and filter are initialized', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use database', status: 'Accepted', date: '2025-01-15' },
      ])
      const { wrapper } = await mountView('/?q=database&status=Accepted')
      await flushPromises()

      const input = wrapper.find('input[type="search"]')
      expect((input.element as HTMLInputElement).value).toBe('database')

      const chips = wrapper.find('[role="group"]').findAll('button')
      const acceptedChip = chips.find(b => b.text().includes('Accepted'))!
      expect(acceptedChip.attributes('aria-pressed')).toBe('true')

      // fetchADRs should have been called with the query
      expect(mockedFetchADRs).toHaveBeenCalledWith('database', expect.any(AbortSignal))
    })

    it('invalid status in URL is ignored gracefully', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 1, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      const { wrapper } = await mountView('/?status=Bogus')
      await flushPromises()

      // Should not crash — no chip rendered for Bogus, filter empty state shown
      expect(wrapper.text()).toContain('No ADRs match the selected filters')
    })
  })
})
