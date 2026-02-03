/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_DB_HOST: string
  readonly VITE_DB_PORT: string
  readonly VITE_DB_NAME: string
  readonly VITE_DB_USER: string
  readonly VITE_DB_PASSWORD: string
  readonly VITE_REDIS_HOST: string
  readonly VITE_REDIS_PORT: string
  readonly VITE_KAFKA_BROKER: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
