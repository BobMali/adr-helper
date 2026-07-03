import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ADRCreateView from './ADRCreateView.vue'
import { fetchConfig, createADR, fetchTemplateSections } from '../api'
import type { TemplateSectionDef } from '../types'

vi.mock('../api', () => ({
  fetchConfig: vi.fn(),
  createADR: vi.fn(),
  fetchTemplateSections: vi.fn(),
}))

const mockedFetchConfig = fetchConfig as ReturnType<typeof vi.fn>
const mockedCreateADR = createADR as ReturnType<typeof vi.fn>
const mockedFetchTemplateSections = fetchTemplateSections as ReturnType<typeof vi.fn>

const nygardSections: TemplateSectionDef[] = [
  { key: 'context', heading: 'Context', kind: 'h2', optional: false, placeholder: 'What is the issue?' },
  { key: 'decision', heading: 'Decision', kind: 'h2', optional: false, placeholder: 'What is the change?' },
  { key: 'consequences', heading: 'Consequences', kind: 'h2', optional: true, placeholder: 'What becomes easier?' },
]

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
      mockedFetchTemplateSections.mockReturnValue(new Promise(() => {}))
      const router = makeRouter()
      router.push('/adr/new')
      const wrapper = mount(ADRCreateView, { global: { plugins: [router] } })

      expect(wrapper.text()).toContain('Loading')
    })
  })

  describe('config error', () => {
    it('shows config error if fetchConfig rejects', async () => {
      mockedFetchConfig.mockRejectedValue(new Error('Config unavailable'))
      mockedFetchTemplateSections.mockResolvedValue(nygardSections)
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Config unavailable')
    })

    it('shows error if fetchTemplateSections rejects', async () => {
      mockedFetchConfig.mockResolvedValue({ template: 'nygard' })
      mockedFetchTemplateSections.mockRejectedValue(new Error('Sections unavailable'))
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Sections unavailable')
    })

    it('shows retry button on error', async () => {
      mockedFetchConfig.mockRejectedValue(new Error('Config unavailable'))
      mockedFetchTemplateSections.mockResolvedValue(nygardSections)
      const { wrapper } = await mountView()
      await flushPromises()

      const retryBtn = wrapper.findAll('button').find(b => b.text().includes('Retry'))
      expect(retryBtn).toBeTruthy()
    })
  })

  describe('form rendering', () => {
    beforeEach(() => {
      mockedFetchConfig.mockResolvedValue({ template: 'nygard' })
      mockedFetchTemplateSections.mockResolvedValue(nygardSections)
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

    it('renders section textareas matching template sections', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      for (const def of nygardSections) {
        const textarea = wrapper.find(`#section-${def.key}`)
        expect(textarea.exists()).toBe(true)
      }
    })

    it('shows optional label on optional sections', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      // consequences is optional
      const fieldsets = wrapper.findAll('fieldset')
      const consequencesFieldset = fieldsets.find(f => f.text().includes('Consequences'))
      expect(consequencesFieldset?.text()).toContain('optional')
    })

    it('has aria-required on required textareas', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const contextTextarea = wrapper.find('#section-context')
      expect(contextTextarea.attributes('aria-required')).toBe('true')

      const consequencesTextarea = wrapper.find('#section-consequences')
      expect(consequencesTextarea.attributes('aria-required')).toBeUndefined()
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
      mockedFetchTemplateSections.mockResolvedValue(nygardSections)
    })

    it('submit with empty title shows error', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(wrapper.text()).toContain('Title is required')
      expect(mockedCreateADR).not.toHaveBeenCalled()
    })

    it('submit with missing required section shows field error', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Test ADR')
      await wrapper.find('#section-context').setValue('Some context')
      // decision is required but missing

      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(wrapper.text()).toContain('Decision is required')
      expect(mockedCreateADR).not.toHaveBeenCalled()
    })

    it('submit with valid title and sections navigates to detail view', async () => {
      const detail = { number: 5, title: 'Use Go', status: 'Proposed', date: '2026-03-02', content: '# 5. Use Go' }
      mockedCreateADR.mockResolvedValue(detail)

      const { wrapper, router } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Use Go')
      await wrapper.find('#section-context').setValue('We need a language')
      await wrapper.find('#section-decision').setValue('Use Go')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(mockedCreateADR).toHaveBeenCalledWith({
        title: 'Use Go',
        sections: { context: 'We need a language', decision: 'Use Go' },
      })
      expect(router.currentRoute.value.name).toBe('detail')
      expect(router.currentRoute.value.params.number).toBe('5')
      expect(router.currentRoute.value.query.created).toBe('true')
    })

    it('all fields populate correctly in API call', async () => {
      const detail = { number: 1, title: 'Test', status: 'Proposed', date: '2026-03-02', content: '# 1. Test' }
      mockedCreateADR.mockResolvedValue(detail)

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Test')
      await wrapper.find('#section-context').setValue('Context text')
      await wrapper.find('#section-decision').setValue('Decision text')
      await wrapper.find('#section-consequences').setValue('Consequences text')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(mockedCreateADR).toHaveBeenCalledWith({
        title: 'Test',
        sections: {
          context: 'Context text',
          decision: 'Decision text',
          consequences: 'Consequences text',
        },
      })
    })

    it('disables form during submission', async () => {
      let resolve!: (value: unknown) => void
      mockedCreateADR.mockReturnValue(new Promise((r) => { resolve = r }))

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('Test')
      await wrapper.find('#section-context').setValue('Context')
      await wrapper.find('#section-decision').setValue('Decision')
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
      await wrapper.find('#section-context').setValue('Context')
      await wrapper.find('#section-decision').setValue('Decision')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(wrapper.text()).toContain('Server error')
      const alert = wrapper.find('[role="alert"]')
      expect(alert.exists()).toBe(true)
    })
  })
})
