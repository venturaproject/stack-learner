import {
  IconBrowserCheck,
  IconHelp,
  IconLayoutDashboard,
  IconNotification,
  IconPalette,
  IconSettings,
  IconUserCog,
  IconUsers,
  IconBook,
  IconBuildingCommunity,
  IconSchool,
  IconUsersGroup,
} from '@tabler/icons-react'
import { type SidebarData } from '../types'
import TradivelLogo from '../tradivel-logo'
import { useI18n } from '@/i18n/context'

export const DynamicSidebarData = () => {
  const { t } = useI18n()

  const sidebarData: SidebarData = {
    teams: [
      {
        name: 'Learner',
        logo: TradivelLogo,
        plan: t('learning_platform'),
      },
    ],
    navGroups: [
      {
        title: t('general'),
        items: [
          {
            title: t('dashboard'),
            url: '/admin',
            icon: IconLayoutDashboard,
          },
        ],
      },
      {
        title: t('learning_group'),
        items: [
          {
            title: t('learning_courses_navigation'),
            url: '/admin/learning',
            icon: IconBook,
          },
          {
            title: t('cohorts_navigation'),
            url: '/admin/cohorts',
            icon: IconUsersGroup,
          },
          {
            title: t('learning_organizations_navigation'),
            url: '/admin/organizations',
            icon: IconBuildingCommunity,
          },
          {
            title: t('learning_students_navigation'),
            url: '/admin/students',
            icon: IconSchool,
          },
        ],
      },
      {
        title: t('access_control'),
        items: [
          {
            title: t('users'),
            url: '/admin/users',
            icon: IconUsers,
            permission: 'users.view',
          },
        ],
      },
      {
        title: t('others'),
        items: [
          {
            title: t('configuration'),
            icon: IconSettings,
            items: [
              {
                title: t('profile'),
                url: '/admin/settings',
                icon: IconUserCog,
              },
              {
                title: t('appearance'),
                url: '/admin/settings/appearance',
                icon: IconPalette,
              },
              {
                title: t('notifications'),
                url: '/admin/settings/notifications',
                icon: IconNotification,
              },
              {
                title: t('display'),
                url: '/admin/settings/display',
                icon: IconBrowserCheck,
              },
            ],
          },
          {
            title: t('help_center'),
            url: '/admin/help-center',
            icon: IconHelp,
          },
        ],
      },
    ],
  };

  return sidebarData;
};
