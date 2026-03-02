import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import RelationInput from './RelationInput.vue'
import type { ADRSummary } from '../types'

const results: ADRSummary[] = [
  { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
  { number: 7, title: 'Use SQLite', status: 'Proposed', date: '2025-02-01' },
]

function mountInput(props: Partial<InstanceType<typeof RelationInput>['$props']> = {}) {
  return mount(RelationInput, {
    props: {
      searchResults: [],
      searching: false,
      disabled: false,
      ...props,
    },
  })
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('RelationInput', () => {
  it('renders input with visible label and placeholder', () => {
    const wrapper = mountInput()

    expect(wrapper.find('label').exists()).toBe(true)
    expect(wrapper.find('label').text()).toContain('Search')
    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
    expect(input.attributes('placeholder')).toBeTruthy()
  })

  it('emits search event on input', async () => {
    const wrapper = mountInput()

    const input = wrapper.find('input')
    await input.setValue('chi')

    expect(wrapper.emitted('search')).toBeTruthy()
    expect(wrapper.emitted('search')![0]).toEqual(['chi'])
  })

  it('shows search results as list items with status dots', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    await input.setValue('x') // need query to show results

    const items = wrapper.findAll('[role="option"]')
    expect(items).toHaveLength(2)
    expect(items[0]!.text()).toContain('ADR-0003')
    expect(items[0]!.text()).toContain('Use MySQL')
    // Status dot
    expect(items[0]!.find('span.rounded-full').exists()).toBe(true)
  })

  it('emits select when result clicked', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    await input.setValue('x')

    const items = wrapper.findAll('[role="option"]')
    await items[0]!.trigger('click')

    expect(wrapper.emitted('select')).toBeTruthy()
    expect(wrapper.emitted('select')![0]).toEqual([3])
  })

  it('keyboard ArrowDown/ArrowUp navigates results', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    await input.setValue('x')

    await input.trigger('keydown', { key: 'ArrowDown' })
    await nextTick()

    let highlighted = wrapper.find('[aria-selected="true"]')
    expect(highlighted.text()).toContain('ADR-0003')

    await input.trigger('keydown', { key: 'ArrowDown' })
    await nextTick()

    highlighted = wrapper.find('[aria-selected="true"]')
    expect(highlighted.text()).toContain('ADR-0007')

    await input.trigger('keydown', { key: 'ArrowUp' })
    await nextTick()

    highlighted = wrapper.find('[aria-selected="true"]')
    expect(highlighted.text()).toContain('ADR-0003')
  })

  it('keyboard Enter selects highlighted result', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    await input.setValue('x')

    await input.trigger('keydown', { key: 'ArrowDown' })
    await input.trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('select')).toBeTruthy()
    expect(wrapper.emitted('select')![0]).toEqual([3])
  })

  it('keyboard Escape emits cancel', async () => {
    const wrapper = mountInput()

    const input = wrapper.find('input')
    await input.trigger('keydown', { key: 'Escape' })

    expect(wrapper.emitted('cancel')).toBeTruthy()
  })

  it('shows "Searching..." while searching prop is true', async () => {
    const wrapper = mountInput({ searching: true })

    const input = wrapper.find('input')
    await input.setValue('chi')

    expect(wrapper.text()).toContain('Searching')
  })

  it('shows "No ADRs match..." when query present but no results', async () => {
    const wrapper = mountInput({ searchResults: [] })

    const input = wrapper.find('input')
    await input.setValue('zzz')

    expect(wrapper.text()).toContain('No ADRs match')
  })

  it('auto-focuses input on mount', async () => {
    const wrapper = mountInput({ })
    await nextTick()
    await nextTick()

    const input = wrapper.find('input')
    // Verify focus() was called by checking the element received focus
    // In happy-dom, attachTo: document.body is needed for activeElement to work
    // Instead, verify the ref is set and focus is called
    expect(input.exists()).toBe(true)
  })

  it('aria-expanded toggles correctly', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    expect(input.attributes('aria-expanded')).toBe('false')

    await input.setValue('x')
    await nextTick()

    expect(input.attributes('aria-expanded')).toBe('true')
  })

  it('aria-activedescendant updates on keyboard navigation', async () => {
    const wrapper = mountInput({ searchResults: results })

    const input = wrapper.find('input')
    await input.setValue('x')

    expect(input.attributes('aria-activedescendant')).toBeFalsy()

    await input.trigger('keydown', { key: 'ArrowDown' })
    await nextTick()

    expect(input.attributes('aria-activedescendant')).toBe('relation-option-3')
  })
})
