import { useEffect, useState, type ReactNode } from 'react'
import {
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
} from '@tanstack/react-table'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { DataTable, DataTablePagination, DataTableViewOptions } from '@/components/data-table'
import { useColumnReorder } from '@/hooks/use-column-reorder'
import { useI18n } from '@/i18n/context'

interface DataTablePanelProps<TData> {
  data: TData[]
  columns: ColumnDef<TData, unknown>[]
  title: string
  description: string
  emptyMessage: string
  searchPlaceholder: string
  columnLabels: Record<string, string>
  createAction?: ReactNode
  serverPage?: {
    page: number
    perPage: number
    total: number
    totalPages: number
    onPageChange: (page: number) => void
    onPerPageChange: (perPage: number) => void
    onSearchChange: (search: string) => void
  }
}

export function DataTablePanel<TData>({
  data,
  columns,
  title,
  description,
  emptyMessage,
  searchPlaceholder,
  columnLabels,
  createAction,
  serverPage,
}: DataTablePanelProps<TData>) {
  const { t } = useI18n()
  const [search, setSearch] = useState('')
  const [localPagination, setLocalPagination] = useState({ pageIndex: 0, pageSize: 10 })
  const pagination = serverPage ? { pageIndex: serverPage.page - 1, pageSize: serverPage.perPage } : localPagination
  const columnIds = columns.map((column) => column.id ?? ('accessorKey' in column ? String(column.accessorKey) : '')).filter(Boolean)
  const { columnOrder, columnVisibility, setColumnVisibility, handleDragStart, handleDrop } = useColumnReorder(columnIds)
  const table = useReactTable({
    data,
    columns,
    state: { globalFilter: search, pagination, columnOrder, columnVisibility },
    onGlobalFilterChange: setSearch,
    onPaginationChange: (updater) => {
      const next = typeof updater === 'function' ? updater(pagination) : updater
      if (serverPage) {
        if (next.pageSize !== pagination.pageSize) serverPage.onPerPageChange(next.pageSize)
        if (next.pageIndex !== pagination.pageIndex) serverPage.onPageChange(next.pageIndex + 1)
      } else setLocalPagination(next)
    },
    onColumnOrderChange: () => {},
    onColumnVisibilityChange: setColumnVisibility,
    globalFilterFn: 'includesString',
    getCoreRowModel: getCoreRowModel(),
    ...(!serverPage && { getFilteredRowModel: getFilteredRowModel() }),
    getSortedRowModel: getSortedRowModel(),
    ...(!serverPage && { getPaginationRowModel: getPaginationRowModel() }),
    manualFiltering: !!serverPage,
    manualPagination: !!serverPage,
    ...(serverPage && { pageCount: serverPage.totalPages }),
  })
  const total = serverPage?.total ?? table.getFilteredRowModel().rows.length
  const lastPage = Math.max(1, serverPage?.totalPages ?? Math.ceil(total / pagination.pageSize))

  useEffect(() => {
    if (serverPage && serverPage.page > lastPage) serverPage.onPageChange(lastPage)
  }, [lastPage, serverPage])

  const updateSearch = (value: string) => {
    setSearch(value)
    if (serverPage) serverPage.onSearchChange(value)
    else setLocalPagination((current) => ({ ...current, pageIndex: 0 }))
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-end">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="relative w-full sm:w-72">
            <Input placeholder={searchPlaceholder} value={search} onChange={(event) => updateSearch(event.target.value)} className="w-full pr-9" />
            {search && <Button type="button" variant="ghost" size="icon" aria-label={t('clear_search')} className="absolute right-1 top-1/2 h-7 w-7 -translate-y-1/2 text-muted-foreground" onClick={() => updateSearch('')}><X className="h-4 w-4" /></Button>}
          </div>
          <div className="flex items-center gap-2"><DataTableViewOptions table={table} columnLabels={columnLabels} />{createAction}</div>
        </div>
      </div>
      <Card>
        <CardHeader><CardTitle>{title}</CardTitle><CardDescription>{description}</CardDescription></CardHeader>
        <CardContent><DataTable table={table} colCount={columns.length} emptyMessage={emptyMessage} onDragStart={handleDragStart} onDrop={handleDrop} /></CardContent>
        <DataTablePagination currentPage={pagination.pageIndex + 1} lastPage={lastPage} perPage={pagination.pageSize} total={total} onPageChange={(page) => table.setPageIndex(page - 1)} onPerPageChange={(value) => table.setPageSize(Number(value))} />
      </Card>
    </div>
  )
}
