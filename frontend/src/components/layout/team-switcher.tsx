import * as React from 'react'
import {
  SidebarMenu,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'

export function TeamSwitcher({
  teams,
}: {
  teams: {
    name: string
    logo: React.ElementType
    plan: string
  }[]
}) {
  const { isMobile, state } = useSidebar();
  const activeTeam = teams[0]

  return (
    <SidebarMenu>
      <SidebarMenuItem className={state === 'collapsed' ? 'flex justify-center px-0 py-3' : 'px-2 py-3'}>
        {React.createElement(activeTeam.logo, {
          className: state === 'collapsed' || isMobile ? 'h-10 w-10' : 'h-10 w-full max-w-full',
          isCollapsed: state === 'collapsed' || isMobile
        })}
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
