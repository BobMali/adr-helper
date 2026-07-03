import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ADRCreateView from './ADRCreateView.vue'
import { fetchConfig, createADR, fetchTemplateSections, fetchScopes, addScope } from '../api'
import type { TemplateSectionDef } from '../types'

vi.mock('../api', () => ({
  fetchConfig: vi.fn(),
  createADR: vi.fn(),
  fetchTemplateSections: vi.fn(),
  fetchScopes: vi.fn(),
  addScope: vi.fn(),
}))

const mockedFetchConfig = fetchConfig as ReturnType<typeof vi.fn>
const mockedCreateADR = createADR as ReturnType<typeof vi.fn>
const mockedFetchTemplateSections = fetchTemplateSections as ReturnType<typeof vi.fn>
const mockedFetchScopes = fetchScopes as ReturnType<typeof vi.fn>
const mockedAddScope = addScope as ReturnType<typeof vi.fn>

beforeEach(() => {
  // Sensible defaults so non-scoped tests don't need to wire the vocabulary.
  mockedFetchScopes.mockResolvedValue([])
})

const nygardSections: TemplateSectionDef[] = [
  { key: 'context', heading: 'Context', kind: 'h2', optional: false, placeholder: 'What is the issue?' },
  { key: 'decision', heading: 'Decision', kind: 'h2', optional: false, placeholder: 'What is the change?' },
  { key: 'consequences', heading: 'Consequences', kind: 'h2', optional: true, placeholder: 'What becomes easier?' },
]

const scopedSections: TemplateSectionDef[] = [
  { key: 'scope', heading: 'Scope', kind: 'meta', optional: false, placeholder: 'Which part(s)?', vocabulary: true },
  { key: 'context', heading: 'Context', kind: 'h2', optional: false, placeholder: 'What is the issue?' },
  { key: 'decision', heading: 'Decision', kind: 'h2', optional: false, placeholder: 'What is the change?' },
  { key: 'consequences', heading: 'Consequences', kind: 'h2', optional: false, placeholder: 'What becomes harder?' },
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

  describe('scope vocabulary (nygard-scoped)', () => {
    beforeEach(() => {
      mockedFetchConfig.mockResolvedValue({ template: 'nygard-scoped' })
      mockedFetchTemplateSections.mockResolvedValue(scopedSections)
    })

    it('renders a checkbox per vocabulary value', async () => {
      mockedFetchScopes.mockResolvedValue(['Backend', 'Frontend'])
      const { wrapper } = await mountView()
      await flushPromises()

      const boxes = wrapper.findAll('input[type="checkbox"]')
      expect(boxes).toHaveLength(2)
      expect(wrapper.text()).toContain('Backend')
      expect(wrapper.text()).toContain('Frontend')
      // Scope field is not a textarea
      expect(wrapper.find('#section-scope').element.tagName).not.toBe('TEXTAREA')
    })

    it('shows empty-state hint when there are no scopes', async () => {
      mockedFetchScopes.mockResolvedValue([])
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('No scopes yet')
    })

    it('adding a scope calls addScope and selects the new value', async () => {
      mockedFetchScopes.mockResolvedValue(['Backend'])
      mockedAddScope.mockResolvedValue(['Backend', 'Frontend'])
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('input[type="text"][aria-label="Add a new scope"]').setValue('Frontend')
      await wrapper.findAll('button').find((b) => b.text() === 'Add')!.trigger('click')
      await flushPromises()

      expect(mockedAddScope).toHaveBeenCalledWith('Frontend')
      const boxes = wrapper.findAll('input[type="checkbox"]')
      expect(boxes).toHaveLength(2)
      // Newly added value is auto-checked
      const frontend = boxes.find((b) => (b.element as HTMLInputElement).value === 'Frontend')!
      expect((frontend.element as HTMLInputElement).checked).toBe(true)
    })

    it('surfaces add-scope validation errors', async () => {
      mockedFetchScopes.mockResolvedValue([])
      mockedAddScope.mockRejectedValue(new Error('scope "a,b" must not contain commas'))
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('input[type="text"][aria-label="Add a new scope"]').setValue('a,b')
      await wrapper.findAll('button').find((b) => b.text() === 'Add')!.trigger('click')
      await flushPromises()

      expect(wrapper.text()).toContain('must not contain commas')
    })

    it('joins selected scopes into the create payload', async () => {
      mockedFetchScopes.mockResolvedValue(['Backend', 'Frontend'])
      mockedCreateADR.mockResolvedValue({ number: 1, title: 'T', status: 'Proposed', date: '2026-03-02', content: '' })
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('#adr-title').setValue('T')
      const boxes = wrapper.findAll('input[type="checkbox"]')
      await boxes[0].setValue(true) // Backend
      await boxes[1].setValue(true) // Frontend
      await wrapper.find('#section-context').setValue('ctx')
      await wrapper.find('#section-decision').setValue('dec')
      await wrapper.find('#section-consequences').setValue('con')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(mockedCreateADR).toHaveBeenCalledWith({
        title: 'T',
        sections: { scope: 'Backend, Frontend', context: 'ctx', decision: 'dec', consequences: 'con' },
      })
    })

    it('blocks submission when no scope is selected (required)', async () => {
      mockedFetchScopes.mockResolvedValue(['Backend'])
      const { wrapper } = await mountView()
      await flushPromises()

      mockedCreateADR.mockClear() // shared factory mock; reset accumulated calls
      await wrapper.find('#adr-title').setValue('T')
      await wrapper.find('#section-context').setValue('ctx')
      await wrapper.find('#section-decision').setValue('dec')
      await wrapper.find('#section-consequences').setValue('con')
      await wrapper.find('form').trigger('submit')
      await flushPromises()

      expect(mockedCreateADR).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain('Scope is required')
    })
  })
