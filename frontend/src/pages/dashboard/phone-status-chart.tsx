import { Bar, BarChart, CartesianGrid, Cell, XAxis, YAxis } from 'recharts'
import { ChartConfig, ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart'
import { useI18n } from '@/i18n/context'

const COLORS = ['#10b981', '#60a5fa', '#f87171', '#94a3b8']

const chartConfig = {
  value: { label: 'Teléfonos' },
} satisfies ChartConfig

interface PhoneStatusChartProps {
  stats: {
    activos: number
    congelados: number
    bajas: number
    sinDatos: number
  }
}

export function PhoneStatusChart({ stats }: PhoneStatusChartProps) {
  const { t } = useI18n()

  const data = [
    { name: t('stat_telefonos_activos_title'), value: stats.activos, fill: COLORS[0] },
    { name: t('stat_telefonos_congelados_title'), value: stats.congelados, fill: COLORS[1] },
    { name: t('stat_telefonos_bajas_title'), value: stats.bajas, fill: COLORS[2] },
    { name: t('tab_sin_datos'), value: stats.sinDatos, fill: COLORS[3] },
  ]

  return (
    <ChartContainer config={chartConfig}>
      <BarChart accessibilityLayer data={data}>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey='name'
          tickLine={false}
          tickMargin={10}
          axisLine={false}
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          allowDecimals={false}
        />
        <ChartTooltip
          cursor={false}
          content={<ChartTooltipContent />}
        />
        <Bar dataKey='value' radius={8}>
          {data.map((entry) => (
            <Cell key={entry.name} fill={entry.fill} />
          ))}
        </Bar>
      </BarChart>
    </ChartContainer>
  )
}
