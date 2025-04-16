import request from '@/util/request'
import type {User} from '@/types/user.js'
import type {LoginRequest, RegisterRequest} from '@/types/request/user'

// 用户注册
const register = ({username, email, full_name, password}: RegisterRequest) =>
    request.post('/users', {
        username,
        email,
        full_name,
        password,
    })


interface LoginResponse {
    user: User
    access_token: string
    access_token_expires_at: string
    refresh_token: string
    refresh_token_expires_at: string
}

// 用户登录
const login = (req: LoginRequest) =>
    request.post<LoginResponse>('/users/login', req
        , {
            // 设置不使用拦截器
            headers: {},
            skipAuth: true, // 自定义配置，用于标记该请求不需要经过拦截器
        })

const info = () => {
    return request.get('/auth/info')
}

const verifyEmail = (email_id: number, secreet_code: string) => {
    return request.get(`/users/verify_email?email_id=${email_id}&secret_code=${secreet_code}`)
}

const contributions = () => {
    return request.get('users/contributions')
}

export default {
    register,
    login,
    info,
    verifyEmail,
    contributions,
}
