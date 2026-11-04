// @vitest-environment jsdom

import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DataTablePagination } from './data-table-pagination'
import { I18nProvider } from '@/i18n/context'

describe('DataTablePagination', () => {
  it('does not claim that rows are selected when selection is not available', async () => {
    render(
      <I18nProvider>
        <DataTablePagination
          currentPage={1}
          lastPage={2}
          perPage={10}
          total={20}
          selectedCount={0}
          onPageChange={vi.fn()}
          onPerPageChange={vi.fn()}
        />
      </I18nProvider>,
    )

    expect(await screen.findByText('Página 1 de 2')).toBeTruthy()
    expect(screen.getByText('20 resultados')).toBeTruthy()
    expect(screen.queryByText(/filas seleccionadas/)).toBeNull()
    expect(screen.getByRole('button', { name: 'Ir a la página anterior' })).toBeTruthy()
  })
})
