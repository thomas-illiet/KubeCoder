<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, useTemplateRef } from 'vue'
import SwaggerUIBundle from 'swagger-ui-dist/swagger-ui-es-bundle.js'
import 'swagger-ui-dist/swagger-ui.css'
import { useAuth } from '../../composables/useAuth'
import { createSwaggerOptions } from '../../documentation/swagger'
import { useRuntimeConfig } from '../../runtime/config'

const config = useRuntimeConfig()
const { accessToken } = useAuth()
const swaggerRoot = useTemplateRef<HTMLDivElement>('swaggerRoot')
const initializationError = shallowRef('')
let destroySwagger: (() => void) | undefined
let addedDarkModeClass = false

onMounted(() => {
  if (!swaggerRoot.value) return
  try {
    if (!document.documentElement.classList.contains('dark-mode')) {
      document.documentElement.classList.add('dark-mode')
      addedDarkModeClass = true
    }
    const options = createSwaggerOptions(config.apiBaseUrl, () => accessToken.value)
    const instance = SwaggerUIBundle({
      ...options,
      domNode: swaggerRoot.value,
      deepLinking: true,
      displayRequestDuration: true,
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout',
    })
    destroySwagger = instance.destroy
  } catch (error) {
    initializationError.value = error instanceof Error ? error.message : 'Swagger UI could not be initialized.'
  }
})

onBeforeUnmount(() => {
  destroySwagger?.()
  if (addedDarkModeClass) document.documentElement.classList.remove('dark-mode')
})
</script>

<template>
  <v-alert v-if="initializationError" type="error" variant="tonal" title="Documentation unavailable">
    {{ initializationError }}
  </v-alert>
  <div v-else ref="swaggerRoot" class="swagger-documentation" />
</template>

<style scoped>
.swagger-documentation {
  min-height: 480px;
  padding: 8px 18px 24px;
  border: 1px solid var(--kc-border);
  border-radius: 12px;
  background: #131820;
  color-scheme: dark;
  overflow: hidden;
}

.swagger-documentation :deep(.swagger-ui) {
  font-family: 'DM Sans', sans-serif;
  color: #d5dae2;
  background: transparent !important;
}

.swagger-documentation :deep(.swagger-ui .info .title),
.swagger-documentation :deep(.swagger-ui .info p),
.swagger-documentation :deep(.swagger-ui .info li),
.swagger-documentation :deep(.swagger-ui .info table),
.swagger-documentation :deep(.swagger-ui .opblock-tag),
.swagger-documentation :deep(.swagger-ui .opblock-description-wrapper p),
.swagger-documentation :deep(.swagger-ui .opblock-external-docs-wrapper p),
.swagger-documentation :deep(.swagger-ui .opblock-title_normal p),
.swagger-documentation :deep(.swagger-ui .response-col_status),
.swagger-documentation :deep(.swagger-ui .response-col_description),
.swagger-documentation :deep(.swagger-ui table thead tr td),
.swagger-documentation :deep(.swagger-ui table thead tr th),
.swagger-documentation :deep(.swagger-ui .parameter__name),
.swagger-documentation :deep(.swagger-ui .parameter__type),
.swagger-documentation :deep(.swagger-ui .model-title),
.swagger-documentation :deep(.swagger-ui .model) {
  color: #d5dae2;
}

.swagger-documentation :deep(.swagger-ui .info .title small pre),
.swagger-documentation :deep(.swagger-ui .prop-format),
.swagger-documentation :deep(.swagger-ui .parameter__in) {
  color: #9aa6b6;
}

.swagger-documentation :deep(.swagger-ui .opblock-tag),
.swagger-documentation :deep(.swagger-ui .models),
.swagger-documentation :deep(.swagger-ui section.models .model-container) {
  border-color: #353d49;
}

.swagger-documentation :deep(.swagger-ui section.models),
.swagger-documentation :deep(.swagger-ui section.models .model-container),
.swagger-documentation :deep(.swagger-ui .scheme-container),
.swagger-documentation :deep(.swagger-ui .responses-inner) {
  background: #171c24;
}

.swagger-documentation :deep(.swagger-ui input),
.swagger-documentation :deep(.swagger-ui select),
.swagger-documentation :deep(.swagger-ui textarea) {
  border-color: #46505f;
  background: #0d1117;
  color: #e8ebf4;
}

.swagger-documentation :deep(.swagger-ui .btn) {
  border-color: #6f7d93;
  color: #dfe3e9;
}

.swagger-documentation :deep(.swagger-ui .btn.authorize) {
  border-color: #55c8bc;
  color: #55c8bc;
}

.swagger-documentation :deep(.swagger-ui .opblock .opblock-section-header) {
  background: rgba(13, 17, 23, .72);
  box-shadow: none;
}

.swagger-documentation :deep(.swagger-ui .opblock .opblock-section-header h4),
.swagger-documentation :deep(.swagger-ui .tab li) {
  color: #dfe3e9;
}

.swagger-documentation :deep(.swagger-ui .highlight-code),
.swagger-documentation :deep(.swagger-ui .microlight) {
  background: #0a0d12 !important;
}

.swagger-documentation :deep(.swagger-ui .info) {
  margin: 32px 0;
}

.swagger-documentation :deep(.swagger-ui .topbar) {
  display: none;
}

.swagger-documentation :deep(.swagger-ui .scheme-container) {
  border-radius: 10px;
  box-shadow: none;
}

@media (max-width: 600px) {
  .swagger-documentation {
    padding: 0 8px 16px;
  }
}
</style>
