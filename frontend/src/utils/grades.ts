const GRADING_COMPANIES = new Set(['psa', 'bgs', 'cgc', 'sgc', 'tag', 'ace'])

/** Formats a stored series key for display: "psa-10" -> "PSA 10", "reverse-holofoil" -> "Reverse Holofoil". */
export function gradeLabel(grade: string): string {
  if (!grade || grade === 'raw') return 'Raw / Ungraded'
  const parts = grade.split('-')
  if (parts.length === 2 && GRADING_COMPANIES.has(parts[0])) {
    return `${parts[0].toUpperCase()} ${parts[1]}`
  }
  return parts.map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(' ')
}

/** True for grading-company series, which need a graded-sales source to exist. */
export function isGradedSeries(grade: string): boolean {
  const [company, value] = grade.split('-')
  return !!value && GRADING_COMPANIES.has(company)
}
