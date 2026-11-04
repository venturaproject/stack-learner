import { useTheme } from '@/context/theme-context'

type LearnerLogoProps = {
  className?: string
  isCollapsed?: boolean
}

export default function LearnerLogo({ className = '', isCollapsed = false }: LearnerLogoProps) {
  const { theme } = useTheme()
  const isDark = theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  const variant = isDark ? 'dark' : 'light'
  const src = isCollapsed
    ? `/learner-logo-compact-${variant}.svg`
    : `/learner-logo-${variant}.svg`

  return (
    <img
      src={src}
      alt="Learner"
      className={`block ${isCollapsed ? 'h-10 w-10' : 'h-10 w-auto max-w-full'} ${className}`}
    />
  )
}
