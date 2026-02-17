import { statusColorFamily, statusDotClass, statusTextClass, chipBgClass } from './statusColors'

describe('statusColorFamily', () => {
  it('returns "green" for Accepted', () => {
    expect(statusColorFamily('Accepted')).toBe('green')
  })

  it('returns "amber" for Proposed', () => {
    expect(statusColorFamily('Proposed')).toBe('amber')
  })

  it('returns "red" for Rejected', () => {
    expect(statusColorFamily('Rejected')).toBe('red')
  })

  it('returns "red" for Superseded', () => {
    expect(statusColorFamily('Superseded')).toBe('red')
  })

  it('returns "red" for Deprecated', () => {
    expect(statusColorFamily('Deprecated')).toBe('red')
  })

  it('is case-insensitive', () => {
    expect(statusColorFamily('accepted')).toBe('green')
    expect(statusColorFamily('PROPOSED')).toBe('amber')
  })
})

describe('statusDotClass', () => {
  it('returns green dot for Accepted', () => {
    expect(statusDotClass('Accepted')).toBe('bg-green-500')
  })

  it('returns amber dot for Proposed', () => {
    expect(statusDotClass('Proposed')).toBe('bg-amber-500')
  })

  it('returns red dot for Rejected', () => {
    expect(statusDotClass('Rejected')).toBe('bg-red-500')
  })
})

describe('statusTextClass', () => {
  it('returns green text for Accepted', () => {
    expect(statusTextClass('Accepted')).toBe('text-green-600 dark:text-green-400')
  })

  it('returns amber text for Proposed', () => {
    expect(statusTextClass('Proposed')).toBe('text-amber-600 dark:text-amber-400')
  })

  it('returns red text for other statuses', () => {
    expect(statusTextClass('Superseded')).toBe('text-red-600 dark:text-red-400')
  })
})

describe('chipBgClass', () => {
  it('returns green chip for Accepted', () => {
    expect(chipBgClass('Accepted')).toBe('bg-green-600 hover:bg-green-700')
  })

  it('returns amber chip for Proposed', () => {
    expect(chipBgClass('Proposed')).toBe('bg-amber-600 hover:bg-amber-700')
  })

  it('returns red chip for Rejected', () => {
    expect(chipBgClass('Rejected')).toBe('bg-red-700 hover:bg-red-800')
  })
})
