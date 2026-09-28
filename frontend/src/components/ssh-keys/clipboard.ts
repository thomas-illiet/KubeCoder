export function copyPublicKey(value: string, clipboard: Pick<Clipboard, 'writeText'> = navigator.clipboard): Promise<void> {
  return clipboard.writeText(value)
}
