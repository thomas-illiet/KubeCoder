/// <reference types="vite/client" />

import type IconSelect from './components/IconSelect.vue'

declare module 'vue' {
  export interface GlobalComponents {
    IconSelect: typeof IconSelect
  }
}

export {}
