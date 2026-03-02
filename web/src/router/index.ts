import { createRouter, createWebHistory } from 'vue-router'
import ADRListView from '../views/ADRListView.vue'
import ADRCreateView from '../views/ADRCreateView.vue'
import ADRDetailView from '../views/ADRDetailView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'list',
      component: ADRListView,
    },
    {
      path: '/adr/new',
      name: 'create',
      component: ADRCreateView,
    },
    {
      path: '/adr/:number',
      name: 'detail',
      component: ADRDetailView,
      props: (route) => ({ number: Number(route.params.number) }),
    },
  ],
})

export default router
