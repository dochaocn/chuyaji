import axios from 'axios'
import { useAuthStore } from '../store/auth'
import router from '../router'

const http = axios.create({ baseURL: '/api/v1' })

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('chuyaji_admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      router.push('/login')
    }
    return Promise.reject(err)
  },
)

export default http
