import router from './index'

describe('router', () => {
  const routes = router.getRoutes()

  it('has route "/" with name "list"', () => {
    const route = routes.find(r => r.path === '/')
    expect(route).toBeDefined()
    expect(route!.name).toBe('list')
  })

  it('has route "/adr/:number" with name "detail"', () => {
    const route = routes.find(r => r.path === '/adr/:number')
    expect(route).toBeDefined()
    expect(route!.name).toBe('detail')
  })

  it('detail route props function converts string param to number', () => {
    const route = routes.find(r => r.name === 'detail')!
    const propsFn = route.props.default as (route: { params: { number: string } }) => { number: number }
    const result = propsFn({ params: { number: '42' } } as never)
    expect(result).toEqual({ number: 42 })
  })
})
