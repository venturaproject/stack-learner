import {
  IconBrowserCheck,
  IconHelp,
  IconLayoutDashboard,
  IconNotification,
  IconPalette,
  IconSettings,
  IconTool,
  IconUserCog,
  IconUsers,
  IconBook,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'
import TradivelLogo from '../tradivel-logo'

export const sidebarData: SidebarData = {
  teams: [
    {
      name: 'Learner',
      logo: TradivelLogo,
      plan: 'Plataforma de aprendizaje',
    },
  ],
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Aprendizaje',
          url: '/admin/learning',
          icon: IconLayoutDashboard,
        },
      ],
    },
    {
      title: 'Formación',
      items: [
        {
          title: 'Cursos',
          url: '/admin/learning',
          icon: IconBook,
        },
      ],
    },
    {
      title: 'Control de Acceso',
      items: [
        {
          title: 'Usuarios',
          url: '/admin/users',
          icon: IconUsers,
          permission: 'users.view',
        },
      ],
    },
    {
      title: 'Otros',
      items: [
        {
          title: 'Configuración',
          icon: IconSettings,
          items: [
            {
              title: 'Perfil',
              url: '/admin/settings',
              icon: IconUserCog,
            },
            {
              title: 'Cuenta',
              url: '/admin/settings/account',
              icon: IconTool,
            },
            {
              title: 'Apariencia',
              url: '/admin/settings/appearance',
              icon: IconPalette,
            },
            {
              title: 'Notificaciones',
              url: '/admin/settings/notifications',
              icon: IconNotification,
            },
            {
              title: 'Visualización',
              url: '/admin/settings/display',
              icon: IconBrowserCheck,
            },
          ],
        },
        {
          title: 'Centro de Ayuda',
          url: '/admin/help-center',
          icon: IconHelp,
        },
      ],
    },
  ],
}
