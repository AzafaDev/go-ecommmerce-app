export {}

type SnapCallbacks = {
  onSuccess?: (result: unknown) => void
  onPending?: (result: unknown) => void
  onError?: (result: unknown) => void
  onClose?: () => void
}

declare global {
  interface Window {
    snap?: {
      pay: (snapToken: string, callbacks?: SnapCallbacks) => void
    }
  }
}
