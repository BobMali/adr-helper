import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import App from './App.vue'

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div>Home stub</div>' } },
    ],
  })
}

describe('App', () => {
  it('renders routed component content inside router-view', async () => {
    const router = makeRouter()
    router.push('/')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [router] } })
    expect(wrapper.text()).toContain('Home stub')
  })

  it('wraps router-view in a <main> landmark element', async () => {
    const router = makeRouter()
    router.push('/')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [router] } })
    const main = wrapper.find('main')
    expect(main.exists()).toBe(true)
    expect(main.text()).toContain('Home stub')
  })

  it('has a viewport-height layout with a scrollable main and max-w-4xl content', async () => {
    const router = makeRouter()
    router.push('/')
    await router.isReady()

    const wrapper = mount(App, { global: { plugins: [router] } })
    // Root is a fixed-height flex column so only <main> scrolls, not the document.
    expect(wrapper.find('.h-dvh').exists()).toBe(true)
    expect(wrapper.find('main.overflow-y-auto').exists()).toBe(true)
    expect(wrapper.find('.max-w-4xl').exists()).toBe(true)
  })
})
