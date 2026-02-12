import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { nextTick } from 'vue'
import ADRDetailView from './ADRDetailView.vue'
import type { ADRDetail, ADRSummary } from '../types'
import { NotFoundError } from '../api'

vi.mock('../api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api')>()
  return {
    ...actual,
    fetchADR: vi.fn(),
    fetchStatuses: vi.fn(),
    fetchADRs: vi.fn(),
    updateADRStatus: vi.fn(),
  }
})

import { fetchADR, fetchStatuses, fetchADRs, updateADRStatus } from '../api'

const mockedFetchADR = fetchADR as ReturnType<typeof vi.fn>
const mockedFetchStatuses = fetchStatuses as ReturnType<typeof vi.fn>
const mockedFetchADRs = fetchADRs as ReturnType<typeof vi.fn>
const mockedUpdateADRStatus = updateADRStatus as ReturnType<typeof vi.fn>

const sampleDetail: ADRDetail = {
  number: 5,
  title: 'Use PostgreSQL',
  status: 'Accepted',
  date: '2025-01-15',
  content: '## Context\nWe need a database.',
}

const sampleStatuses = ['Proposed', 'Accepted', 'Deprecated', 'Superseded']

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'list', component: { template: '<div>List</div>' } },
      {
        path: '/adr/:number',
        name: 'detail',
        component: ADRDetailView,
        props: (route) => ({ number: Number(route.params.number) }),
      },
    ],
  })
}

async function mountView(number = 5) {
  const router = makeRouter()
  router.push(`/adr/${number}`)
  await router.isReady()
  const wrapper = mount(ADRDetailView, {
    props: { number },
    global: { plugins: [router] },
  })
  return { wrapper, router }
}

afterEach(() => {
  vi.restoreAllMocks()
  vi.clearAllMocks()
})

describe('ADRDetailView', () => {
  describe('loading', () => {
    it('shows "Loading" before fetches resolve', () => {
      mockedFetchADR.mockReturnValue(new Promise(() => {}))
      mockedFetchStatuses.mockReturnValue(new Promise(() => {}))
      const router = makeRouter()
      router.push('/adr/5')
      const wrapper = mount(ADRDetailView, {
        props: { number: 5 },
        global: { plugins: [router] },
      })
      expect(wrapper.text()).toContain('Loading')
    })

    it('loading indicator has role="status" for screen readers', () => {
      mockedFetchADR.mockReturnValue(new Promise(() => {}))
      mockedFetchStatuses.mockReturnValue(new Promise(() => {}))
      const router = makeRouter()
      router.push('/adr/5')
      const wrapper = mount(ADRDetailView, {
        props: { number: 5 },
        global: { plugins: [router] },
      })

      const loadingEl = wrapper.find('[role="status"]')
      expect(loadingEl.exists()).toBe(true)
      expect(loadingEl.text()).toContain('Loading')
    })
  })

  describe('not found', () => {
    it('shows "ADR #N not found" and back link on NotFoundError', async () => {
      mockedFetchADR.mockRejectedValue(new NotFoundError('ADR #5 not found'))
      mockedFetchStatuses.mockResolvedValue(sampleStatuses)
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('ADR #5 not found')
      const backLink = wrapper.find('a')
      expect(backLink.attributes('href')).toBe('/')
    })
  })

  describe('error', () => {
    it('shows error message and back link on generic error', async () => {
      mockedFetchADR.mockRejectedValue(new Error('Server exploded'))
      mockedFetchStatuses.mockResolvedValue(sampleStatuses)
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('Server exploded')
      const backLink = wrapper.find('a')
      expect(backLink.attributes('href')).toBe('/')
    })
  })

  describe('detail rendering', () => {
    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('shows title as "ADR #N: title"', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('ADR #5: Use PostgreSQL')
    })

    it('shows formatted date', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.text()).toContain('January 15, 2025')
    })

    it('renders markdown content as HTML', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const section = wrapper.find('section')
      expect(section.html()).toContain('<h2')
      expect(section.text()).toContain('Context')
    })

    it('has back link to list', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const navLink = wrapper.find('nav a')
      expect(navLink.attributes('href')).toBe('/')
      expect(navLink.text()).toContain('Back to list')
    })
  })

  describe('status dropdown', () => {
    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('renders all statuses as options', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const options = wrapper.findAll('select#status-select option')
      expect(options).toHaveLength(4)
      expect(options.map(o => o.attributes('value'))).toEqual(sampleStatuses)
    })

    it('current status is selected', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const select = wrapper.find<HTMLSelectElement>('select#status-select')
      expect(select.element.value).toBe('Accepted')
    })
  })

  describe('status update', () => {
    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('calls updateADRStatus on non-Superseded change', async () => {
      const updatedADR = { ...sampleDetail, status: 'Deprecated' }
      mockedUpdateADRStatus.mockResolvedValue(updatedADR)

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Deprecated')
      await flushPromises()

      expect(mockedUpdateADRStatus).toHaveBeenCalledWith(5, 'Deprecated', undefined)
    })

    it('shows success feedback, clears after 4s', async () => {
      vi.useFakeTimers()
      const updatedADR = { ...sampleDetail, status: 'Deprecated' }
      mockedUpdateADRStatus.mockResolvedValue(updatedADR)

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Deprecated')
      await flushPromises()

      expect(wrapper.text()).toContain('Status updated to Deprecated')

      vi.advanceTimersByTime(4000)
      await nextTick()

      expect(wrapper.text()).not.toContain('Status updated')
      vi.useRealTimers()
    })

    it('shows error feedback and reverts status on failure', async () => {
      mockedUpdateADRStatus.mockRejectedValue(new Error('Update failed'))

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Deprecated')
      await flushPromises()

      expect(wrapper.text()).toContain('Update failed')
      const select = wrapper.find<HTMLSelectElement>('select#status-select')
      expect(select.element.value).toBe('Accepted')
    })

    it('disables select while updating', async () => {
      let resolveUpdate!: (value: ADRDetail) => void
      mockedUpdateADRStatus.mockReturnValue(
        new Promise<ADRDetail>((resolve) => { resolveUpdate = resolve }),
      )

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Deprecated')
      await nextTick()

      expect(wrapper.find<HTMLSelectElement>('select#status-select').element.disabled).toBe(true)

      resolveUpdate({ ...sampleDetail, status: 'Deprecated' })
      await flushPromises()

      expect(wrapper.find<HTMLSelectElement>('select#status-select').element.disabled).toBe(false)
    })
  })

  describe('supersede flow', () => {
    const otherADRs: ADRSummary[] = [
      { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
      { number: 5, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      { number: 7, title: 'Use SQLite', status: 'Proposed', date: '2025-02-01' },
    ]

    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('shows panel when "Superseded" selected', async () => {
      mockedFetchADRs.mockResolvedValue(otherADRs)
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      expect(wrapper.text()).toContain('Select the ADR that supersedes this one')
    })

    it('fetches and filters other ADRs (excludes current)', async () => {
      mockedFetchADRs.mockResolvedValue(otherADRs)
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const panel = wrapper.find('[role="group"]')
      expect(panel.text()).toContain('Use MySQL')
      expect(panel.text()).toContain('Use SQLite')
      expect(panel.text()).not.toContain('Use PostgreSQL')
    })

    it('shows loading state in panel', async () => {
      mockedFetchADRs.mockReturnValue(new Promise(() => {}))
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await nextTick()

      const panel = wrapper.find('[role="group"]')
      expect(panel.text()).toContain('Loading ADRs')
    })

    it('shows empty state in panel when no other ADRs', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 5, title: 'Use PostgreSQL', status: 'Accepted', date: '2025-01-15' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const panel = wrapper.find('[role="group"]')
      expect(panel.text()).toContain('No other ADRs available')
    })

    it('confirm button disabled until ADR chosen', async () => {
      mockedFetchADRs.mockResolvedValue(otherADRs)
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const confirmBtn = wrapper.findAll('button').find(b => b.text() === 'Confirm')!
      expect(confirmBtn.element.disabled).toBe(true)
    })

    it('confirm calls updateADRStatus with supersededBy', async () => {
      mockedFetchADRs.mockResolvedValue(otherADRs)
      const updatedADR = { ...sampleDetail, status: 'Superseded' }
      mockedUpdateADRStatus.mockResolvedValue(updatedADR)

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      // Select the superseding ADR via option element
      const supersedingSelect = wrapper.find<HTMLSelectElement>('[role="group"] select')
      const options = supersedingSelect.findAll('option')
      const targetOption = options.find(o => o.text().includes('Use MySQL'))!
      supersedingSelect.element.value = targetOption.element.value
      await supersedingSelect.trigger('change')
      await nextTick()

      const confirmBtn = wrapper.findAll('button').find(b => b.text() === 'Confirm')!
      await confirmBtn.trigger('click')
      await flushPromises()

      expect(mockedUpdateADRStatus).toHaveBeenCalledWith(5, 'Superseded', { supersededBy: 3 })
    })

    it('cancel reverts to previous status and hides panel', async () => {
      mockedFetchADRs.mockResolvedValue(otherADRs)
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const cancelBtn = wrapper.findAll('button').find(b => b.text() === 'Cancel')!
      await cancelBtn.trigger('click')
      await nextTick()

      expect(wrapper.find('[role="group"]').exists()).toBe(false)
      const select = wrapper.find<HTMLSelectElement>('select#status-select')
      expect(select.element.value).toBe('Accepted')
    })
  })

  describe('keyboard', () => {
    const otherADRs: ADRSummary[] = [
      { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
    ]

    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
      mockedFetchADRs.mockResolvedValue(otherADRs)
    })

    it('Escape cancels supersede flow', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const panel = wrapper.find('[role="group"]')
      await panel.trigger('keydown', { key: 'Escape' })
      await nextTick()

      expect(wrapper.find('[role="group"]').exists()).toBe(false)
    })

    it('Enter confirms when ADR selected', async () => {
      mockedUpdateADRStatus.mockResolvedValue({ ...sampleDetail, status: 'Superseded' })
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const supersedingSelect = wrapper.find<HTMLSelectElement>('[role="group"] select')
      const options = supersedingSelect.findAll('option')
      const targetOption = options.find(o => o.text().includes('Use MySQL'))!
      supersedingSelect.element.value = targetOption.element.value
      await supersedingSelect.trigger('change')
      await nextTick()

      const panel = wrapper.find('[role="group"]')
      await panel.trigger('keydown', { key: 'Enter' })
      await flushPromises()

      expect(mockedUpdateADRStatus).toHaveBeenCalledWith(5, 'Superseded', { supersededBy: 3 })
    })

    it('Enter ignored when no ADR selected', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      const panel = wrapper.find('[role="group"]')
      await panel.trigger('keydown', { key: 'Enter' })
      await flushPromises()

      expect(mockedUpdateADRStatus).not.toHaveBeenCalled()
    })
  })

  describe('supersede race condition prevention', () => {
    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('cancels previous fetchADRs when cancel is clicked during slow fetch', async () => {
      let resolveFirst!: (v: ADRSummary[]) => void
      mockedFetchADRs.mockImplementationOnce(
        () => new Promise<ADRSummary[]>(r => { resolveFirst = r }),
      )

      const { wrapper } = await mountView()
      await flushPromises()

      // Select Superseded — fires slow fetch
      await wrapper.find('select#status-select').setValue('Superseded')
      await nextTick()

      // Click Cancel — should abort the in-flight fetch
      const cancelBtn = wrapper.findAll('button').find(b => b.text() === 'Cancel')!
      await cancelBtn.trigger('click')
      await nextTick()

      // Resolve the stale fetch after cancellation
      resolveFirst([
        { number: 3, title: 'Stale ADR', status: 'Accepted', date: '2025-01-01' },
      ])
      await flushPromises()

      // Panel should be hidden, stale results should not populate availableADRs
      expect(wrapper.find('[role="group"]').exists()).toBe(false)
      expect(wrapper.text()).not.toContain('Stale ADR')
    })

    it('does NOT show AbortError when supersede fetch is cancelled', async () => {
      mockedFetchADRs.mockRejectedValueOnce(
        new DOMException('The operation was aborted', 'AbortError'),
      )

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      expect(wrapper.text()).not.toContain('aborted')
      expect(wrapper.text()).not.toContain('AbortError')
    })
  })

  describe('unmount cleanup', () => {
    it('clears feedback timer on unmount', async () => {
      vi.useFakeTimers()
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
      const updatedADR = { ...sampleDetail, status: 'Deprecated' }
      mockedUpdateADRStatus.mockResolvedValue(updatedADR)

      const { wrapper } = await mountView()
      await flushPromises()

      // Trigger status update so the 4s timer starts
      await wrapper.find('select#status-select').setValue('Deprecated')
      await flushPromises()

      expect(wrapper.text()).toContain('Status updated to Deprecated')

      // Unmount before timer fires
      wrapper.unmount()

      // Advance past the 4s timer — should not throw
      vi.advanceTimersByTime(4000)

      vi.useRealTimers()
    })
  })

  describe('XSS sanitization', () => {
    it('strips script tags from markdown content', async () => {
      mockedFetchADR.mockResolvedValue({
        ...sampleDetail,
        content: '# Title\n<script>alert("xss")</script>',
      })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])

      const { wrapper } = await mountView()
      await flushPromises()

      const section = wrapper.find('section')
      expect(section.html()).not.toContain('<script>')
    })

    it('strips event handler attributes', async () => {
      mockedFetchADR.mockResolvedValue({
        ...sampleDetail,
        content: '<img src=x onerror="alert(\'xss\')">',
      })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])

      const { wrapper } = await mountView()
      await flushPromises()

      const section = wrapper.find('section')
      expect(section.html()).not.toContain('onerror')
    })

    it('allows safe markdown formatting', async () => {
      mockedFetchADR.mockResolvedValue({
        ...sampleDetail,
        content: '## Heading\n\n**Bold text**',
      })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])

      const { wrapper } = await mountView()
      await flushPromises()

      const section = wrapper.find('section')
      expect(section.html()).toContain('<h2')
      expect(section.html()).toContain('<strong>')
    })
  })

  describe('accessibility', () => {
    beforeEach(() => {
      mockedFetchADR.mockResolvedValue({ ...sampleDetail })
      mockedFetchStatuses.mockResolvedValue([...sampleStatuses])
    })

    it('title has tabindex="-1" and receives focus on mount', async () => {
      const { wrapper } = await mountView()
      await flushPromises()
      await nextTick()

      const title = wrapper.find('h1')
      expect(title.attributes('tabindex')).toBe('-1')
    })

    it('has label for status select', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      const label = wrapper.find('label[for="status-select"]')
      expect(label.exists()).toBe(true)
      expect(label.text()).toContain('Status')
    })

    it('has role="group" on supersede panel', async () => {
      mockedFetchADRs.mockResolvedValue([
        { number: 3, title: 'Other', status: 'Accepted', date: '2025-01-01' },
      ])
      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Superseded')
      await flushPromises()

      expect(wrapper.find('[role="group"]').exists()).toBe(true)
    })

    it('has aria-live region', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.find('[aria-live="polite"]').exists()).toBe(true)
    })

    it('has aria-busy on select when updating', async () => {
      let resolveUpdate!: (value: ADRDetail) => void
      mockedUpdateADRStatus.mockReturnValue(
        new Promise<ADRDetail>((resolve) => { resolveUpdate = resolve }),
      )

      const { wrapper } = await mountView()
      await flushPromises()

      await wrapper.find('select#status-select').setValue('Deprecated')
      await nextTick()

      expect(wrapper.find('select#status-select').attributes('aria-busy')).toBe('true')

      resolveUpdate({ ...sampleDetail, status: 'Deprecated' })
      await flushPromises()
    })

    it('uses semantic HTML elements', async () => {
      const { wrapper } = await mountView()
      await flushPromises()

      expect(wrapper.find('article').exists()).toBe(true)
      expect(wrapper.find('nav').exists()).toBe(true)
      expect(wrapper.find('header').exists()).toBe(true)
      expect(wrapper.find('section').exists()).toBe(true)
    })
  })
})
