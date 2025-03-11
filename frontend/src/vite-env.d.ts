/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_DOMAIN: string
  // その他必要な環境変数があれば追加
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
