import { createApp, defineComponent } from 'vue'

/**
 * Runs a composable in a temporary Vue component context.
 * Returns [result, app] so the caller can unmount if needed.
 */
export function withSetup<T>(composable: () => T): [T, ReturnType<typeof createApp>] {
  let result!: T
  const app = createApp(
    defineComponent({
      setup() {
        result = composable()
        return () => null
      },
    }),
  )
  app.mount(document.createElement('div'))
  return [result, app]
}
