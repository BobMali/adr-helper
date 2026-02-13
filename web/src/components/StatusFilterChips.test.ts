import { mount } from '@vue/test-utils'
import StatusFilterChips from './StatusFilterChips.vue'

describe('StatusFilterChips', () => {
  const statuses = ['Accepted', 'Proposed', 'Rejected']

  function mountChips(props?: { modelValue?: Set<string> }) {
    return mount(StatusFilterChips, {
      props: {
        statuses,
        modelValue: props?.modelValue ?? new Set<string>(),
      },
    })
  }

  it('renders a button for each status', () => {
    const wrapper = mountChips()
    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(3)
    expect(buttons[0]!.text()).toContain('Accepted')
    expect(buttons[1]!.text()).toContain('Proposed')
    expect(buttons[2]!.text()).toContain('Rejected')
  })

  it('container has role="group" with aria-label', () => {
    const wrapper = mountChips()
    const group = wrapper.find('[role="group"]')
    expect(group.exists()).toBe(true)
    expect(group.attributes('aria-label')).toBe('Filter by status')
  })

  it('unselected buttons have aria-pressed="false"', () => {
    const wrapper = mountChips()
    const buttons = wrapper.findAll('button')
    for (const btn of buttons) {
      expect(btn.attributes('aria-pressed')).toBe('false')
    }
  })

  it('selected buttons have aria-pressed="true"', () => {
    const wrapper = mountChips({ modelValue: new Set(['Accepted', 'Rejected']) })
    const buttons = wrapper.findAll('button')
    expect(buttons[0]!.attributes('aria-pressed')).toBe('true')  // Accepted
    expect(buttons[1]!.attributes('aria-pressed')).toBe('false') // Proposed
    expect(buttons[2]!.attributes('aria-pressed')).toBe('true')  // Rejected
  })

  it('clicking unselected chip emits update:modelValue with status added', async () => {
    const wrapper = mountChips()
    await wrapper.findAll('button')[0]!.trigger('click')

    const emitted = wrapper.emitted('update:modelValue')!
    expect(emitted).toHaveLength(1)
    const newSet = emitted[0]![0] as Set<string>
    expect(newSet).toBeInstanceOf(Set)
    expect(newSet.has('Accepted')).toBe(true)
    expect(newSet.size).toBe(1)
  })

  it('clicking selected chip emits update:modelValue with status removed', async () => {
    const wrapper = mountChips({ modelValue: new Set(['Accepted', 'Proposed']) })
    await wrapper.findAll('button')[0]!.trigger('click')

    const emitted = wrapper.emitted('update:modelValue')!
    expect(emitted).toHaveLength(1)
    const newSet = emitted[0]![0] as Set<string>
    expect(newSet).toBeInstanceOf(Set)
    expect(newSet.has('Accepted')).toBe(false)
    expect(newSet.has('Proposed')).toBe(true)
    expect(newSet.size).toBe(1)
  })

  it('selected chip gets filled background class', () => {
    const wrapper = mountChips({ modelValue: new Set(['Accepted']) })
    const acceptedBtn = wrapper.findAll('button')[0]!
    expect(acceptedBtn.classes()).toContain('bg-green-600')
  })

  it('unselected chip gets bg-transparent', () => {
    const wrapper = mountChips()
    const acceptedBtn = wrapper.findAll('button')[0]!
    expect(acceptedBtn.classes()).toContain('bg-transparent')
  })

  it('proposed chip gets amber background when selected', () => {
    const wrapper = mountChips({ modelValue: new Set(['Proposed']) })
    const proposedBtn = wrapper.findAll('button')[1]!
    expect(proposedBtn.classes()).toContain('bg-amber-600')
  })

  it('rejected chip gets red background when selected', () => {
    const wrapper = mountChips({ modelValue: new Set(['Rejected']) })
    const rejectedBtn = wrapper.findAll('button')[2]!
    expect(rejectedBtn.classes()).toContain('bg-red-700')
  })

  it('renders nothing when statuses array is empty', () => {
    const wrapper = mount(StatusFilterChips, {
      props: { statuses: [], modelValue: new Set<string>() },
    })
    expect(wrapper.findAll('button')).toHaveLength(0)
  })

  it('emits a new Set instance, not a mutated one', async () => {
    const original = new Set(['Accepted'])
    const wrapper = mountChips({ modelValue: original })
    await wrapper.findAll('button')[1]!.trigger('click') // click Proposed

    const emitted = wrapper.emitted('update:modelValue')!
    const newSet = emitted[0]![0] as Set<string>
    expect(newSet).not.toBe(original)
  })
})
