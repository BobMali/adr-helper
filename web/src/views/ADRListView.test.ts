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

    it('links to /adr/{number} for each ADR', async () => {
      const wrapper = await mountView()
      await flushPromises()

      const links = wrapper.findAll('a')
      const hrefs = links.map(l => l.attributes('href'))
      expect(hrefs).toContain('/adr/1')
      expect(hrefs).toContain('/adr/2')
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
