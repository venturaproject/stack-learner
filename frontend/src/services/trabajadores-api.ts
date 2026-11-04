import { axios } from '@/lib/axios'
import { API_ENDPOINTS } from '@/config'
import { normalizeEntityPayload } from '@/lib/api-utils'
import type { TrabajadorDetail } from '@/pages/trabajadores/show'
import type { Trabajador } from '@/pages/trabajadores/types'

export interface TrabajadoresListResponse {
  results?: Trabajador[]
  data?: Trabajador[]
  count?: number
  page?: number
  per_page?: number
}

export const trabajadoresApi = {
  async list(params?: Record<string, string>): Promise<TrabajadoresListResponse> {
    const { data } = await axios.get(API_ENDPOINTS.trabajadores, { params })
    return data as TrabajadoresListResponse
  },

  async detail(id: string | number): Promise<TrabajadorDetail | null> {
    const { data } = await axios.get(`${API_ENDPOINTS.trabajadores}/${id}`)
    return normalizeEntityPayload<TrabajadorDetail>(data)
  },
}
