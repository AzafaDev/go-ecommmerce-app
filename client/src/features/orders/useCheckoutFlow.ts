import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCheckout } from './hooks'
import { useSnapResultHandler } from './useSnapResultHandler'
import { loadSnapScript, payWithSnap } from '../../lib/midtrans'
import { useToast } from '../../components/toast/ToastContext'
import { ApiError } from '../../lib/http'

export type CheckoutPhase = 'idle' | 'creating-order' | 'awaiting-payment'

export function useCheckoutFlow() {
  const checkout = useCheckout()
  const navigate = useNavigate()
  const { push } = useToast()
  const handleResult = useSnapResultHandler()
  const [phase, setPhase] = useState<CheckoutPhase>('idle')

  async function start() {
    setPhase('creating-order')
    try {
      const { order } = await checkout.mutateAsync()

      if (!order.snap_token) {
        // Order was created, but the Midtrans call failed server-side —
        // no regenerate-token endpoint exists, so this is a dead end for now.
        push({
          variant: 'error',
          message: "We couldn't start payment for this order. Check Order History to try again.",
        })
        navigate('/orders')
        return
      }

      setPhase('awaiting-payment')
      await loadSnapScript()
      const result = await payWithSnap(order.snap_token)
      handleResult(result, order.id)
    } catch (err) {
      push({
        variant: 'error',
        message: err instanceof ApiError ? err.message : 'Checkout failed. Please try again.',
      })
    } finally {
      setPhase('idle')
    }
  }

  return { start, phase }
}
