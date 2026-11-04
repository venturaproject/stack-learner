import { axios } from '@/lib/axios'
import { API_ENDPOINTS } from '@/config'

export interface DashboardSummary {
  worker_stats: {
    activos: number
    con_telefono: number
    incorporados_este_mes: number
    con_imei: number
  }
  phone_stats: {
    total: number
    activos: number
    congelados: number
    bajas: number
    sin_datos: number
  }
  recent_devices: Array<{
    id: number
    marca: string | null
    modelo: string | null
    numero: string | null
    grupo: string | null
    employee_nombre: string | null
    updated_at: string
  }>
  monthly_workers: Array<{ month: string; incorporations: number }>
  totals: {
    trabajadores: number
    telefonos: number
    dispositivos: number
    users: number
  }
}

export const dashboardApi = {
  async summary(): Promise<DashboardSummary> {
    const { data } = await axios.get<DashboardSummary>(API_ENDPOINTS.dashboard)
    return data
  },
}
