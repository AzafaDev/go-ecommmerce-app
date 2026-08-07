export function formatOrderDate(iso: string) {
  return new Date(iso).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function formatOrderId(id: string) {
  return `ORD-${id.split('-')[0].toUpperCase()}`
}
