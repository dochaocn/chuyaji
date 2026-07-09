import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/auth'

const router = createRouter({
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/',
      component: () => import('../components/AdminLayout.vue'),
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '数据总览' } },
        { path: 'users', name: 'UserList', component: () => import('../views/UserList.vue'), meta: { title: '用户管理' } },
        { path: 'users/:id', name: 'UserDetail', component: () => import('../views/UserDetail.vue'), meta: { title: '用户详情' } },
        { path: 'records', name: 'RecordList', component: () => import('../views/RecordList.vue'), meta: { title: '宝宝记录' } },
        { path: 'mother-records', name: 'MotherRecordList', component: () => import('../views/MotherRecordList.vue'), meta: { title: '宝妈记录' } },
        { path: 'reminders', name: 'ReminderList', component: () => import('../views/ReminderList.vue'), meta: { title: '提醒管理' } },
        { path: 'analytics', name: 'Analytics', component: () => import('../views/Analytics.vue'), meta: { title: '数据分析' } },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.name !== 'Login' && !auth.token) {
    return { name: 'Login' }
  }
})

export default router
