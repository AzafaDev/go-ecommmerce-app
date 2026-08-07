/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_MIDTRANS_CLIENT_KEY: string
  readonly VITE_MIDTRANS_ENV: 'sandbox' | 'production'
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
