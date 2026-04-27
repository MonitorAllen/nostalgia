import axios, {
  AxiosInstance,
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
  AxiosResponse,
} from 'axios'
import { showFullScreenLoading, tryHideFullScreenLoading } from '@/components/Loading/fullScreen'
import { LOGIN_URL } from '@/config'
import { ElMessage } from 'element-plus'
import { ResultData } from '@/api/interface'
import { ResultEnum } from '@/enums/httpEnum'
import { checkStatus } from './helper/checkStatus'
import { AxiosCanceler } from './helper/axiosCancel'
import { useUserStore } from '@/stores/modules/user'
import router from '@/routers'

export interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
  loading?: boolean
  cancel?: boolean
  _isRetry?: boolean
}

const config = {
  // 默认地址请求地址，可在 .env.** 文件中修改
  baseURL: import.meta.env.VITE_API_URL as string,
  // 设置超时时间
  timeout: ResultEnum.TIMEOUT as number,
  // 跨域时候允许携带凭证
  withCredentials: true,
}

const axiosCanceler = new AxiosCanceler()

class RequestHttp {
  service: AxiosInstance
  private isRefreshing = false
  private requests: ((token: string) => void)[] = []

  public constructor(config: AxiosRequestConfig) {
    // instantiation
    this.service = axios.create(config)

    /**
     * @description 请求拦截器
     */
    this.service.interceptors.request.use(
      (config: CustomAxiosRequestConfig) => {
        const userStore = useUserStore()
        // 重复请求不需要取消，在 api 服务中通过指定的第三个参数: { cancel: false } 来控制
        config.cancel ??= true
        config.cancel && axiosCanceler.addPending(config)
        // 当前请求不需要显示 loading，在 api 服务中通过指定的第三个参数: { loading: false } 来控制
        config.loading ??= true
        config.loading && showFullScreenLoading()
        if (config.headers && typeof config.headers.set === 'function') {
          config.headers.set('Authorization', userStore.token ? `Bearer ${userStore.token}` : '')
        }
        return config
      },
      (error: AxiosError) => {
        return Promise.reject(error)
      },
    )

    /**
     * @description 响应拦截器
     */
    this.service.interceptors.response.use(
      (response: AxiosResponse & { config: CustomAxiosRequestConfig }) => {
        const { data, config } = response
        axiosCanceler.removePending(config)
        config.loading && tryHideFullScreenLoading()

        // Nostalgia: if status is 2xx, wrap data for Geeker compatibility
        return {
          code: ResultEnum.SUCCESS,
          msg: '成功',
          data: data,
        } as any
      },
      async (error: AxiosError & { config: CustomAxiosRequestConfig }) => {
        const { response, config } = error
        tryHideFullScreenLoading()

        // Handle 401 and Token Refresh
        if (response?.status === ResultEnum.OVERDUE && config && !config._isRetry) {
          const userStore = useUserStore()

          if (!this.isRefreshing) {
            this.isRefreshing = true
            try {
              if (!userStore.refreshToken) throw new Error('No refresh token')

              // Call refresh API directly with basic axios to avoid interceptor loop if possible
              const res = await axios.post(`${config.baseURL}/admin/renew_access`, {
                refresh_token: userStore.refreshToken,
              })

              const newToken = res.data.access_token
              userStore.setToken(newToken)
              if (res.data.access_token_expires_at) {
                userStore.setExpiresAt(res.data.access_token_expires_at)
              }

              this.isRefreshing = false
              this.onRefreshed(newToken)

              // Retry current request
              config._isRetry = true
              config.headers.Authorization = `Bearer ${newToken}`
              return this.service(config)
            } catch (refreshError) {
              this.isRefreshing = false
              userStore.setToken('')
              router.replace(LOGIN_URL)
              return Promise.reject(refreshError)
            }
          } else {
            // Wait for refresh to complete
            return new Promise((resolve) => {
              this.subscribeTokenRefresh((token) => {
                config._isRetry = true
                config.headers.Authorization = `Bearer ${token}`
                resolve(this.service(config))
              })
            })
          }
        }

        // Handle error messages
        if (error.message.indexOf('timeout') !== -1) ElMessage.error('请求超时！请您稍后重试')
        if (error.message.indexOf('Network Error') !== -1) ElMessage.error('网络错误！请您稍后重试')

        if (response) {
          const msg = (response.data as any)?.error || ''
          if (msg) ElMessage.error(msg)
          else checkStatus(response.status)
        }

        if (!window.navigator.onLine) router.replace('/500')
        return Promise.reject(error)
      },
    )
  }

  private subscribeTokenRefresh(cb: (token: string) => void) {
    this.requests.push(cb)
  }

  private onRefreshed(token: string) {
    this.requests.map((cb) => cb(token))
    this.requests = []
  }

  /**
   * @description 常用请求方法封装
   */
  get<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
    return this.service.get(url, { params, ..._object })
  }
  post<T>(url: string, params?: object | string, _object = {}): Promise<ResultData<T>> {
    return this.service.post(url, params, _object)
  }
  put<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
    return this.service.put(url, params, _object)
  }
  patch<T>(url: string, params?: object, _object = {}): Promise<ResultData<T>> {
    return this.service.patch(url, params, _object)
  }
  delete<T>(url: string, params?: any, _object = {}): Promise<ResultData<T>> {
    return this.service.delete(url, { params, ..._object })
  }
  download(url: string, params?: object, _object = {}): Promise<BlobPart> {
    return this.service.post(url, params, { ..._object, responseType: 'blob' })
  }
}

export default new RequestHttp(config)
