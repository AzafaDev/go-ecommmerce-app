import { useState } from 'react'
import { loadSnapScript, payWithSnap } from '../../lib/midtrans'
import { useSnapResultHandler } from './useSnapResultHandler'
import { useToast } from '../../components/toast/ToastContext'

export function useResumePayment() {
  const handleResult = useSnapResultHandler()
  const { push } = useToast()
  const [pendingOrderId, setPendingOrderId] = useState<string | null>(null)

  async function resume(orderId: string, snapToken: string) {
    setPendingOrderId(orderId)
    try {
      await loadSnapScript()
      const result = await payWithSnap(snapToken)
      handleResult(result, orderId)
    } catch {
      push({ variant: 'error', message: 'Could not open payment window. Please try again.' })
    } finally {
      setPendingOrderId(null)
    }
  }

  return { resume, pendingOrderId }
}
