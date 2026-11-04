import { IconSearch } from '@tabler/icons-react'
import { cn } from '@/lib/utils'
import { useSearch } from '@/context/search-context'
import { Button } from './ui/button'

interface Props {
  className?: string
  type?: React.HTMLInputTypeAttribute
  placeholder?: string
}

export function Search({ className = '', placeholder = 'Buscar...' }: Props) {
  const { setOpen } = useSearch()
  return (
    <Button
      variant='outline'
      className={cn(
        'relative h-8 w-8 p-0 sm:h-8 sm:w-full sm:flex-1 md:flex-none justify-center sm:justify-start rounded-md bg-muted/25 hover:bg-muted/50 text-sm font-normal text-muted-foreground shadow-none sm:px-3 sm:pr-12 md:w-40 lg:w-56 xl:w-64',
        className
      )}
      onClick={() => setOpen(true)}
    >
      <IconSearch
        aria-hidden='true'
        className='h-4 w-4 sm:absolute sm:left-1.5 sm:top-1/2 sm:-translate-y-1/2'
      />
      <span className='hidden sm:inline-block sm:ml-3'>{placeholder}</span>
      <kbd className='pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex'>
        <span className='text-xs'>⌘</span>K
      </kbd>
    </Button>
  )
}
