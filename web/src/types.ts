export interface ADRSummary {
  number: number
  title: string
  status: string
  date: string
  meta?: Record<string, string[]>
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

export type SortField = 'number' | 'title' | 'status' | 'date'
export type SortDirection = 'asc' | 'desc'

// How multiple selected values within one metadata facet combine.
export type MetaMatchMode = 'any' | 'all'

// A filterable metadata field (facet). For vocabulary facets, `values` holds the
// project vocabulary (the chip options).
export interface MetaField {
  key: string
  heading: string
  vocabulary?: boolean
  values?: string[]
}
