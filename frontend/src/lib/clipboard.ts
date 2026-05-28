export async function copyText(text: string): Promise<boolean> {
  if (!navigator.clipboard) {
    return false;
  }

  await navigator.clipboard.writeText(text);
  return true;
}