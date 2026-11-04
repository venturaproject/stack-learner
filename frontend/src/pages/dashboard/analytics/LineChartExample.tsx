"use client"

import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart"
import { useI18n } from "@/i18n/context"

interface MonthlyData {
  month: string
  incorporations: number
}

interface LineChartExampleProps {
  monthlyStats: MonthlyData[]
}

const chartConfig = {
  incorporations: {
    label: "Incorporaciones",
    color: "hsl(var(--chart-1))",
  },
} satisfies ChartConfig

const DEMO_MONTHS = (() => {
  const months = ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio',
                  'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre']
  const now = new Date()
  return Array.from({ length: 6 }, (_, i) => {
    const d = new Date(now.getFullYear(), now.getMonth() - 5 + i, 1)
    return { month: months[d.getMonth()], incorporations: [3, 5, 2, 7, 4, 6][i] }
  })
})()

export function LineChartExample({ monthlyStats }: LineChartExampleProps) {
  const { t } = useI18n()

  const hasData = monthlyStats.length > 0 && monthlyStats.some((m) => m.incorporations > 0)
  const source = hasData ? monthlyStats : DEMO_MONTHS

  const chartData = source.map((item) => ({
    month: item.month,
    incorporations: item.incorporations,
  }))

  const total = source.reduce((sum, item) => sum + item.incorporations, 0)

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('worker_incorporations')}</CardTitle>
        <CardDescription>
          {t('monthly_incorporations_last_6_months', { total })}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig}>
          <LineChart accessibilityLayer data={chartData}>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="month"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => value.slice(0, 3)}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={8}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent />}
            />
            <Line
              dataKey="incorporations"
              type="monotone"
              stroke="var(--color-incorporations)"
              strokeWidth={2}
              dot={{ fill: "var(--color-incorporations)", strokeWidth: 2, r: 4 }}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
