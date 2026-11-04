import axios from 'axios'

export type ExportFormat = 'excel' | 'pdf' | 'csv'

export async function downloadExport(
  url: string,
  format: ExportFormat,
  selectedIds: Set<string | number>,
  selectAll: boolean,
  filters: Record<string, string | undefined>,
  filename: string,
): Promise<void> {
  const body = selectAll
    ? {
        format,
        select_all: true,
        filters: Object.fromEntries(
          Object.entries(filters).filter(([, v]) => v != null && v !== ''),
        ),
      }
    : { format, ids: Array.from(selectedIds).map(Number) }

  const response = await axios.post(url, body, { responseType: 'blob' })
  const blobUrl = window.URL.createObjectURL(new Blob([response.data]))
  const a = document.createElement('a')
  a.href = blobUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(blobUrl)
}
