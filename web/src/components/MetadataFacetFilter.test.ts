import { mount } from '@vue/test-utils'
import MetadataFacetFilter from './MetadataFacetFilter.vue'

function mountFacet(props: {
  modelValue?: Set<string>
  matchMode?: 'any' | 'all'
  values?: string[]
} = {}) {
  return mount(MetadataFacetFilter, {
    props: {
      heading: 'Scope',
      values: props.values ?? ['backend', 'api', 'web'],
      modelValue: props.modelValue ?? new Set<string>(),
      matchMode: props.matchMode ?? 'any',
    },
  })
}

describe('MetadataFacetFilter', () => {
  it('renders a chip per value with an accessible group label', () => {
    const wrapper = mountFacet()
    const group = wrapper.find('[aria-label="Filter by Scope"]')
    expect(group.exists()).toBe(true)
    expect(group.findAll('button').map(b => b.text())).toEqual(['backend', 'api', 'web'])
  })

  it('marks selected chips with aria-pressed', () => {
    const wrapper = mountFacet({ modelValue: new Set(['api']) })
    const api = wrapper.find('[aria-label="Filter by Scope"]').findAll('button').find(b => b.text() === 'api')!
    expect(api.attributes('aria-pressed')).toBe('true')
  })

  it('emits a NEW Set on toggle (immutability)', async () => {
    const original = new Set(['backend'])
    const wrapper = mountFacet({ modelValue: original })

    await wrapper.find('[aria-label="Filter by Scope"]').findAll('button')
      .find(b => b.text() === 'api')!.trigger('click')

    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toHaveLength(1)
    const next = emitted![0]![0] as Set<string>
    expect(next).not.toBe(original) // new Set, not mutated in place
    expect(original.has('api')).toBe(false) // original untouched
    expect([...next].sort()).toEqual(['api', 'backend'])
  })

  it('removes a value when toggled off', async () => {
    const wrapper = mountFacet({ modelValue: new Set(['backend', 'api']) })
    await wrapper.find('[aria-label="Filter by Scope"]').findAll('button')
      .find(b => b.text() === 'backend')!.trigger('click')

    const next = wrapper.emitted('update:modelValue')![0]![0] as Set<string>
    expect([...next]).toEqual(['api'])
  })

  it('hides the match toggle until two values are selected', async () => {
    const one = mountFacet({ modelValue: new Set(['backend']) })
    expect(one.find('[aria-label="Match any or all selected scopes"]').exists()).toBe(false)

    const two = mountFacet({ modelValue: new Set(['backend', 'api']) })
    expect(two.find('[aria-label="Match any or all selected scopes"]').exists()).toBe(true)
  })

  it('emits update:matchMode when the toggle is used', async () => {
    const wrapper = mountFacet({ modelValue: new Set(['backend', 'api']), matchMode: 'any' })
    const allBtn = wrapper.find('[aria-label="Match any or all selected scopes"]')
      .findAll('button').find(b => b.text().includes('All'))!
    await allBtn.trigger('click')
    expect(wrapper.emitted('update:matchMode')![0]![0]).toBe('all')
  })

  it('applies a scroll cap only when the vocabulary is large', () => {
    const small = mountFacet({ values: ['a', 'b'] })
    expect(small.find('[aria-label="Filter by Scope"]').classes()).not.toContain('overflow-y-auto')

    const many = mountFacet({ values: Array.from({ length: 9 }, (_, i) => `v${i}`) })
    expect(many.find('[aria-label="Filter by Scope"]').classes()).toContain('overflow-y-auto')
  })
})
