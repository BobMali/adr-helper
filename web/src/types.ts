export interface ADRSummary {
  number: number
  title: string
  status: string
  date: string
}

export interface ADRDetail extends ADRSummary {
  content: string
}

export interface CreateADRPayload {
  title: string
  sections?: Record<string, string>
}

export interface TemplateSectionDef {
  key: string
  heading: string
  kind: string
  optional: boolean
  placeholder: string
  vocabulary?: boolean
}

export type SortField = 'number' | 'title' | 'status'
export type SortDirection = 'asc' | 'desc'
