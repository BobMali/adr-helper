import { createRouter, createWebHistory } from 'vue-router'
import ADRListView from '../views/ADRListView.vue'
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
      path: '/adr/:number',
      name: 'detail',
      component: ADRDetailView,
      props: (route) => ({ number: Number(route.params.number) }),
    },
  ],
})

export default router
