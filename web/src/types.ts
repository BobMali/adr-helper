export interface ADRSummary {
  number: number
  title: string
  status: string
  date: string
}

export interface ADRDetail extends ADRSummary {
  content: string
}

export type SortField = 'number' | 'title' | 'status'
export type SortDirection = 'asc' | 'desc'
