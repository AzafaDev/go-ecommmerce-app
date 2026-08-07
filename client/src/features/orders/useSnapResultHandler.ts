import { useNavigate } from 'react-router-dom'
import { useToast } from '../../components/toast/ToastContext'
import type { SnapResult } from '../../lib/midtrans'

// Shared outcome handling for both a fresh checkout and a "Pay Now" resume,
// so the four Snap.js callback branches are defined in exactly one place.
export function useSnapResultHandler() {
  const navigate = useNavigate()
  const { push } = useToast()

  return (result: SnapResult, orderId: string) => {
    switch (result.type) {
      case 'success':
        push({ variant: 'success', message: 'Payment successful!' })
        navigate(`/orders/${orderId}`)
        break
      case 'pending':
        push({ variant: 'info', message: 'Payment is being processed.' })
        navigate(`/orders/${orderId}`)
        break
      case 'error':
        push({
          variant: 'error',
          message: 'Payment failed. You can try again from Order History.',
        })
        navigate('/orders')
        break
      case 'closed':
        push({
          variant: 'info',
          message: 'Payment window closed. Your order is saved as pending — resume anytime from Order History.',
        })
        navigate('/orders')
        break
    }
  }
}
