export interface RepositoryFormValues {
  clone_url: string
  name: string
  default_branch: string
}

export function isValidCloneURL(cloneURL: string): boolean {
  try {
    const value = new URL(cloneURL)
    return value.protocol === 'https:' || value.protocol === 'ssh:'
  } catch {
    return false
  }
}

export function canContinueRepositoryStep(step: number, form: RepositoryFormValues): boolean {
  if (step === 1) return true
  if (step === 2) return isValidCloneURL(form.clone_url)
  if (step === 3) return Boolean(form.name.trim() && form.default_branch.trim())
  return true
}
