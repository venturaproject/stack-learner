import { useQuery } from '@tanstack/react-query'
import Dashboard from '@/pages/dashboard/index'
import { dashboardApi } from '@/services/dashboard-api'

const DEFAULT_STATS = { activos: 0, conTelefono: 0, incorporadosEsteMes: 0, conImei: 0 }
const DEFAULT_PHONE = { total: 0, activos: 0, congelados: 0, bajas: 0, sinDatos: 0 }

export default function DashboardRoute() {
  const { data } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const s = await dashboardApi.summary()
      return {
        stats: {
          activos: s.worker_stats.activos,
          conTelefono: s.worker_stats.con_telefono,
          incorporadosEsteMes: s.worker_stats.incorporados_este_mes,
          conImei: s.worker_stats.con_imei,
        },
        phoneStats: {
          total: s.phone_stats.total,
          activos: s.phone_stats.activos,
          congelados: s.phone_stats.congelados,
          bajas: s.phone_stats.bajas,
          sinDatos: s.phone_stats.sin_datos,
        },
        recentDevices: s.recent_devices,
        monthlyWorkerStats: s.monthly_workers,
      }
    },
  })

  return (
    <Dashboard
      stats={data?.stats ?? DEFAULT_STATS}
      phoneStats={data?.phoneStats ?? DEFAULT_PHONE}
      recentDevices={data?.recentDevices ?? []}
      monthlyWorkerStats={data?.monthlyWorkerStats ?? []}
    />
  )
}
