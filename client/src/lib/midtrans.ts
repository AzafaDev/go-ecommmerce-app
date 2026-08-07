let snapScriptPromise: Promise<void> | null = null

export function loadSnapScript(): Promise<void> {
  if (window.snap) return Promise.resolve()
  if (snapScriptPromise) return snapScriptPromise

  snapScriptPromise = new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src =
      import.meta.env.VITE_MIDTRANS_ENV === 'production'
        ? 'https://app.midtrans.com/snap/snap.js'
        : 'https://app.sandbox.midtrans.com/snap/snap.js'
    script.setAttribute('data-client-key', import.meta.env.VITE_MIDTRANS_CLIENT_KEY)
    script.onload = () => resolve()
    script.onerror = () => {
      snapScriptPromise = null
      reject(new Error('Failed to load payment provider'))
    }
    document.body.appendChild(script)
  })
  return snapScriptPromise
}

export type SnapResult =
  | { type: 'success' }
  | { type: 'pending' }
  | { type: 'error' }
  | { type: 'closed' }

export function payWithSnap(snapToken: string): Promise<SnapResult> {
  return new Promise((resolve) => {
    window.snap!.pay(snapToken, {
      onSuccess: () => resolve({ type: 'success' }),
      onPending: () => resolve({ type: 'pending' }),
      onError: () => resolve({ type: 'error' }),
      onClose: () => resolve({ type: 'closed' }),
    })
  })
}
