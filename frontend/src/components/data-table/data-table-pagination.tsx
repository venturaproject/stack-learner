import {
  ChevronsLeft,
  ChevronsRight,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { CardFooter } from '@/components/ui/card'
import { useI18n } from '@/i18n/context'
import { PAGINATION_OPTIONS } from '@/lib/constants'

interface DataTablePaginationProps {
  currentPage: number
  lastPage: number
  perPage: number
  total: number
  /** Kept optional for legacy table callers; no selection UI is shown unless implemented. */
  selectedCount?: number
  onPageChange: (page: number) => void
  onPerPageChange: (value: string) => void
}

export function DataTablePagination({
  currentPage,
  lastPage,
  perPage,
  total,
  onPageChange,
  onPerPageChange,
}: DataTablePaginationProps) {
  const { t } = useI18n()

  return (
    <CardFooter className="flex flex-col items-start justify-between gap-4 py-4 sm:flex-row sm:items-center">
      <span className="hidden flex-1 text-sm text-muted-foreground sm:block">
        {t('pagination_items_count', { count: total })}
      </span>
      <div className="flex w-full items-center justify-between gap-3 sm:w-auto sm:justify-end sm:gap-6 lg:gap-8">
        <div className="flex items-center gap-2">
          <span className="hidden whitespace-nowrap text-sm text-muted-foreground sm:block">
            {t('rows_per_page')}
          </span>
          <Select value={String(perPage)} onValueChange={onPerPageChange}>
            <SelectTrigger className="h-8 w-[70px]">
              <SelectValue />
            </SelectTrigger>
              <SelectContent side="top">
              {PAGINATION_OPTIONS.map((n) => (
                <SelectItem key={n} value={String(n)}>
                  {n}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <span className="min-w-[92px] whitespace-nowrap text-center text-sm font-medium text-muted-foreground">
          {t('page_of', { current: currentPage, total: lastPage })}
        </span>
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(1)}
            disabled={currentPage === 1}
            aria-label={t('go_to_first_page')}
          >
            <ChevronsLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage === 1}
            aria-label={t('go_to_previous_page')}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage === lastPage}
            aria-label={t('go_to_next_page')}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => onPageChange(lastPage)}
            disabled={currentPage === lastPage}
            aria-label={t('go_to_last_page')}
          >
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </CardFooter>
  )
}
