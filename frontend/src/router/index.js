import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/Home.vue')
    },
    {
      path: '/grpc',
      name: 'grpc',
      component: () => import('@/views/GRPC.vue')
    },
    {
      path: '/thrift',
      name: 'thrift',
      component: () => import('@/views/Thrift.vue')
    }
  ]
})

export default router
