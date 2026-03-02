import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ADRCreateView from './ADRCreateView.vue'
import { fetchConfig, createADR } from '../api'

vi.mock('../api', () => ({
  fetchConfig: vi.fn(),
  createADR: vi.fn(),
}))

const mockedFetchConfig = fetchConfig as ReturnType<typeof vi.fn>
const mockedCreateADR = createADR as ReturnType<typeof vi.fn>

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'list', component: { template: '<div />' } },
      { path: '/adr/new', name: 'create', component: ADRCreateView },
      { path: '/adr/:number', name: 'detail', component: { template: '<div />' }, props: true },
    ],
  })
}

async function mountView(options?: { attachTo?: HTMLElement }) {
  const router = makeRouter()
  router.push('/adr/new')
  await router.isReady()
  const wrapper = mount(ADRCreateView, {
    global: { plugins: [router] },
    attachTo: options?.attachTo,
  })
  return { wrapper, router }
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ADRCreateView', () => {
  describe('loading state', () => {
    it('shows loading while config fetches', () => {
      mockedFetchConfig.mockReturnValue(new Promise(() => {}))
      const router = makeRouter()
      router.push('/adr/new')
      const wrapper = mount(ADRCreateView, { global: { plugins: [router] } })

      expect(wrapper.text()).toContain('Loading')
    })
  })

  describe('config error', () => {
    it('shows config error if fetchConfig rejects', async () => {
      mockedFetchConfig.mockRejectedValue(new Error('Config unavailable'))
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Config unavailable')
    })
  })

  describe('form rendering', () => {
    beforeEach(() => {
      mockedFetchConfig.mockResolvedValue({ template: 'nygard' })
    })

    it('shows template name after config loads', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('nygard')
    })

    it('focuses title input on mount', async () => {
      const container = document.createElement('div')
      document.body.appendChild(container)
      const { wrapper } = await mountView({ attachTo: container })
      await flushPromises()

      const input = wrapper.find('#adr-title')
      expect(input.exists()).toBe(true)
      expect(input.element).toBe(document.activeElement)

      wrapper.unmount()
      container.remove()
    })

    it('cancel links to /', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const cancelLink = wrapper.findAll('a').find(a => a.text() === 'Cancel')
      expect(cancelLink).toBeDefined()
      expect(cancelLink!.attributes('href')).toBe('/')
    })
  })

  describe('form submission', () => {
    beforeEach(() => {
      mockedFetchConfig.mockResolvedValue({ template: 'nygard' })
    })

    it('submit with empty title shows error', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(wrapper.text()).toContain('Title is required')
      expect(mockedCreateADR).not.toHaveBeenCalled()
    })

    it('submit with valid title navigates to detail view', async () => {
      const detail = { number: 5, title: 'Use Go', status: 'Proposed', date: '2026-03-02', content: '# 5. Use Go' }
      mockedCreateADR.mockResolvedValue(detail)

      const { wrapper, router } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Use Go')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(mockedCreateADR).toHaveBeenCalledWith({ title: 'Use Go' })
      expect(router.currentRoute.value.name).toBe('detail')
      expect(router.currentRoute.value.params.number).toBe('5')
      expect(router.currentRoute.value.query.created).toBe('true')
    })

    it('disables form during submission', async () => {
      let resolve!: (value: unknown) => void
      mockedCreateADR.mockReturnValue(new Promise((r) => { resolve = r }))

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Test')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      const submitBtn = wrapper.find('button[type="submit"]')
      expect(submitBtn.attributes('disabled')).toBeDefined()
      expect(submitBtn.text()).toContain('Creating')

      const input = wrapper.find('#adr-title')
      expect(input.attributes('disabled')).toBeDefined()

      resolve({ number: 1, title: 'Test', status: 'Proposed', date: '2026-03-02', content: '# Test' })
      await flushPromises()
    })

    it('shows API error on failure', async () => {
      mockedCreateADR.mockRejectedValue(new Error('Server error'))

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Test')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(wrapper.text()).toContain('Server error')
      const alert = wrapper.find('[role="alert"]')
      expect(alert.exists()).toBe(true)
    })
  })
})
