import { mount } from '@vue/test-utils'
import SupersedeSelector from './SupersedeSelector.vue'
import type { ADRSummary } from '../types'

const sampleADRs: ADRSummary[] = [
  { number: 3, title: 'Use MySQL', status: 'Accepted', date: '2025-01-01' },
  { number: 7, title: 'Use SQLite', status: 'Proposed', date: '2025-02-01' },
]

function mountSelector(props: Partial<{
  availableADRs: ADRSummary[]
  loadingADRs: boolean
  modelValue: number | null
  disabled: boolean
}> = {}) {
  return mount(SupersedeSelector, {
    props: {
      availableADRs: props.availableADRs ?? sampleADRs,
      loadingADRs: props.loadingADRs ?? false,
      modelValue: props.modelValue ?? null,
      disabled: props.disabled ?? false,
    },
  })
}

describe('SupersedeSelector', () => {
  it('renders role="group" container', () => {
    const wrapper = mountSelector()
    expect(wrapper.find('[role="group"]').exists()).toBe(true)
  })

  it('shows loading state when loadingADRs is true', () => {
    const wrapper = mountSelector({ loadingADRs: true })
    expect(wrapper.text()).toContain('Loading ADRs')
  })

  it('shows empty state when no ADRs available', () => {
    const wrapper = mountSelector({ availableADRs: [], loadingADRs: false })
    expect(wrapper.text()).toContain('No other ADRs available')
  })

  it('renders select with ADR options', () => {
    const wrapper = mountSelector()
    const options = wrapper.findAll('select option')
    // 1 placeholder + 2 ADRs
    expect(options).toHaveLength(3)
    expect(options[1]!.text()).toContain('ADR-0003')
    expect(options[1]!.text()).toContain('Use MySQL')
    expect(options[2]!.text()).toContain('ADR-0007')
    expect(options[2]!.text()).toContain('Use SQLite')
  })

  it('confirm button disabled when modelValue is null', () => {
    const wrapper = mountSelector({ modelValue: null })
    const confirmBtn = wrapper.findAll('button').find(b => b.text() === 'Confirm')!
    expect(confirmBtn.element.disabled).toBe(true)
  })

  it('both buttons disabled when disabled prop is true', () => {
    const wrapper = mountSelector({ disabled: true, modelValue: 3 })
    const buttons = wrapper.findAll('button')
    const confirmBtn = buttons.find(b => b.text() === 'Confirm')!
    const cancelBtn = buttons.find(b => b.text() === 'Cancel')!
    expect(confirmBtn.element.disabled).toBe(true)
    expect(cancelBtn.element.disabled).toBe(true)
  })

  it('emits update:modelValue on select change', async () => {
    const wrapper = mountSelector()
    const select = wrapper.find('select')
    select.element.value = '3'
    await select.trigger('change')

    expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([3])
  })

  it('emits cancel on Escape keydown', async () => {
    const wrapper = mountSelector()
    const group = wrapper.find('[role="group"]')
    await group.trigger('keydown', { key: 'Escape' })

    expect(wrapper.emitted('cancel')).toBeTruthy()
  })

  it('emits confirm on Enter keydown when modelValue is set', async () => {
    const wrapper = mountSelector({ modelValue: 3 })
    const group = wrapper.find('[role="group"]')
    await group.trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('confirm')).toBeTruthy()
  })

  it('does not emit confirm on Enter when modelValue is null', async () => {
    const wrapper = mountSelector({ modelValue: null })
    const group = wrapper.find('[role="group"]')
    await group.trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('confirm')).toBeFalsy()
  })
})
