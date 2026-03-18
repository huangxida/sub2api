<template>
  <BaseDialog :show="show" :title="t('admin.usage.detail.title')" width="full" @close="handleClose">
    <div class="space-y-4">
      <div
        v-if="usage"
        class="grid grid-cols-1 gap-3 rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm dark:border-dark-600 dark:bg-dark-900 md:grid-cols-4"
      >
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.usage.requestId') }}</div>
          <div class="mt-1 break-all font-mono text-xs text-gray-900 dark:text-gray-100">{{ detail?.request_id || usage.request_id || '-' }}</div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.model') }}</div>
          <div class="mt-1 text-gray-900 dark:text-gray-100">{{ detail?.model || usage.model }}</div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.type') }}</div>
          <div class="mt-1 text-gray-900 dark:text-gray-100">{{ requestTypeLabel }}</div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.time') }}</div>
          <div class="mt-1 text-gray-900 dark:text-gray-100">{{ formatDateTime(detail?.created_at || usage.created_at) }}</div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
          :class="activeSection === 'request'
            ? 'bg-primary-600 text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
          @click="activeSection = 'request'"
        >
          {{ t('admin.usage.detail.requestTab') }}
        </button>
        <button
          type="button"
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
          :class="activeSection === 'response'
            ? 'bg-primary-600 text-white'
            : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
          @click="activeSection = 'response'"
        >
          {{ t('admin.usage.detail.responseTab') }}
        </button>
      </div>

      <div v-if="loading" class="rounded-xl border border-gray-200 bg-white p-8 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-900 dark:text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else-if="errorMessage" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700 dark:border-rose-900/60 dark:bg-rose-950/40 dark:text-rose-300">
        {{ errorMessage }}
      </div>

      <div v-else-if="!detail?.has_detail" class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-300">
        {{ t('admin.usage.detail.noDetail') }}
      </div>

      <div v-else-if="activeSection === 'request'" class="space-y-4">
        <div v-if="requestHeadersPayload" class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-900">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-2">
              <div class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ t('admin.usage.detail.requestHeaders') }}</div>
              <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
                <span>{{ t('admin.usage.detail.contentType') }}: {{ requestHeadersPayload.content_type || '-' }}</span>
                <span>{{ t('admin.usage.detail.sizeBytes') }}: {{ requestHeadersPayload.size_bytes }}</span>
                <span v-if="requestHeadersPayload.complete != null">
                  {{ t('admin.usage.detail.complete') }}:
                  {{ requestHeadersPayload.complete ? t('common.yes') : t('common.no') }}
                </span>
              </div>
            </div>

            <div v-if="requestHeadersCanShowJSON" class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                :class="requestHeadersViewMode === 'json'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                @click="requestHeadersViewMode = 'json'"
              >
                {{ t('admin.usage.detail.jsonView') }}
              </button>
              <button
                type="button"
                class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                :class="requestHeadersViewMode === 'raw'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                @click="requestHeadersViewMode = 'raw'"
              >
                {{ t('admin.usage.detail.rawView') }}
              </button>
            </div>
          </div>

          <div class="mt-4">
            <JsonTreeViewer
              v-if="requestHeadersViewMode === 'json' && requestHeadersCanShowJSON && requestHeadersJSON != null"
              :value="requestHeadersJSON"
              :raw="requestHeadersRaw"
            />
            <TextSearchViewer
              v-else
              :content="requestHeadersRaw"
            />
          </div>
        </div>

        <div v-else class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-300">
          {{ t('admin.usage.detail.noRequestHeaders') }}
        </div>

        <div v-if="requestPayload" class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-900">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-2">
              <div class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ t('admin.usage.detail.requestBody') }}</div>
              <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
                <span>{{ t('admin.usage.detail.contentType') }}: {{ requestPayload.content_type || '-' }}</span>
                <span>{{ t('admin.usage.detail.sizeBytes') }}: {{ requestPayload.size_bytes }}</span>
                <span v-if="requestPayload.complete != null">
                  {{ t('admin.usage.detail.complete') }}:
                  {{ requestPayload.complete ? t('common.yes') : t('common.no') }}
                </span>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-2">
              <div v-if="requestPayload.kind === 'frames'" class="flex items-center gap-2">
                <label class="text-sm text-gray-600 dark:text-gray-300" for="request-frame-select">
                  {{ t('admin.usage.detail.frameSelect') }}
                </label>
                <select id="request-frame-select" v-model.number="selectedRequestFrameIndex" class="input-field w-40">
                  <option v-for="(_, index) in requestPayload.frames || []" :key="index" :value="index">
                    {{ t('admin.usage.detail.frameLabel', { index: index + 1 }) }}
                  </option>
                </select>
              </div>

              <template v-if="requestPayloadCanShowJSON">
                <button
                  type="button"
                  class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="requestViewMode === 'json'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                  @click="requestViewMode = 'json'"
                >
                  {{ t('admin.usage.detail.jsonView') }}
                </button>
                <button
                  type="button"
                  class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="requestViewMode === 'raw'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                  @click="requestViewMode = 'raw'"
                >
                  {{ t('admin.usage.detail.rawView') }}
                </button>
              </template>
            </div>
          </div>

          <div class="mt-4">
            <JsonTreeViewer
              v-if="requestViewMode === 'json' && requestPayloadCanShowJSON && requestPayloadJSON != null"
              :value="requestPayloadJSON"
              :raw="requestPayloadRaw"
            />
            <TextSearchViewer
              v-else
              :content="requestPayloadRaw"
            />
          </div>
        </div>
      </div>

      <div v-else class="space-y-4">
        <div v-if="responseHeadersPayload" class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-900">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-2">
              <div class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ t('admin.usage.detail.responseHeaders') }}</div>
              <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
                <span>{{ t('admin.usage.detail.contentType') }}: {{ responseHeadersPayload.content_type || '-' }}</span>
                <span>{{ t('admin.usage.detail.sizeBytes') }}: {{ responseHeadersPayload.size_bytes }}</span>
                <span v-if="responseHeadersPayload.complete != null">
                  {{ t('admin.usage.detail.complete') }}:
                  {{ responseHeadersPayload.complete ? t('common.yes') : t('common.no') }}
                </span>
              </div>
            </div>

            <div v-if="responseHeadersCanShowJSON" class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                :class="responseHeadersViewMode === 'json'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                @click="responseHeadersViewMode = 'json'"
              >
                {{ t('admin.usage.detail.jsonView') }}
              </button>
              <button
                type="button"
                class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                :class="responseHeadersViewMode === 'raw'
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                @click="responseHeadersViewMode = 'raw'"
              >
                {{ t('admin.usage.detail.rawView') }}
              </button>
            </div>
          </div>

          <div class="mt-4">
            <JsonTreeViewer
              v-if="responseHeadersViewMode === 'json' && responseHeadersCanShowJSON && responseHeadersJSON != null"
              :value="responseHeadersJSON"
              :raw="responseHeadersRaw"
            />
            <TextSearchViewer
              v-else
              :content="responseHeadersRaw"
            />
          </div>
        </div>

        <div v-else class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-300">
          {{ t('admin.usage.detail.noResponseHeaders') }}
        </div>

        <div v-if="responsePayload" class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-900">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-2">
              <div class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ t('admin.usage.detail.responseBody') }}</div>
              <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
                <span>{{ t('admin.usage.detail.contentType') }}: {{ responsePayload.content_type || '-' }}</span>
                <span>{{ t('admin.usage.detail.sizeBytes') }}: {{ responsePayload.size_bytes }}</span>
                <span v-if="responsePayload.complete != null">
                  {{ t('admin.usage.detail.complete') }}:
                  {{ responsePayload.complete ? t('common.yes') : t('common.no') }}
                </span>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-2">
              <div v-if="responsePayload.kind === 'frames'" class="flex items-center gap-2">
                <label class="text-sm text-gray-600 dark:text-gray-300" for="response-frame-select">
                  {{ t('admin.usage.detail.frameSelect') }}
                </label>
                <select id="response-frame-select" v-model.number="selectedResponseFrameIndex" class="input-field w-40">
                  <option v-for="(_, index) in responsePayload.frames || []" :key="index" :value="index">
                    {{ t('admin.usage.detail.frameLabel', { index: index + 1 }) }}
                  </option>
                </select>
              </div>

              <template v-if="responsePayloadCanShowJSON">
                <button
                  type="button"
                  class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="responseViewMode === 'json'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                  @click="responseViewMode = 'json'"
                >
                  {{ t('admin.usage.detail.jsonView') }}
                </button>
                <button
                  type="button"
                  class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
                  :class="responseViewMode === 'raw'
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
                  @click="responseViewMode = 'raw'"
                >
                  {{ t('admin.usage.detail.rawView') }}
                </button>
              </template>
            </div>
          </div>

          <div class="mt-4">
            <JsonTreeViewer
              v-if="responseViewMode === 'json' && responsePayloadCanShowJSON && responsePayloadJSON != null"
              :value="responsePayloadJSON"
              :raw="responsePayloadRaw"
            />
            <TextSearchViewer
              v-else
              :content="responsePayloadRaw"
            />
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminUsageDetailResponse, AdminUsageLog, UsageDetailPayload } from '@/types'
import { formatDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'
import JsonTreeViewer from './JsonTreeViewer.vue'
import TextSearchViewer from './TextSearchViewer.vue'

type ViewMode = 'json' | 'raw'

const props = defineProps<{
  show: boolean
  loading: boolean
  usage: AdminUsageLog | null
  detail: AdminUsageDetailResponse | null
  errorMessage?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const activeSection = ref<'request' | 'response'>('request')
const requestHeadersViewMode = ref<ViewMode>('json')
const requestViewMode = ref<ViewMode>('json')
const responseHeadersViewMode = ref<ViewMode>('json')
const responseViewMode = ref<ViewMode>('json')
const selectedRequestFrameIndex = ref(0)
const selectedResponseFrameIndex = ref(0)

const requestTypeLabel = computed(() => {
  const requestType = props.detail?.request_type || props.usage?.request_type
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
})

const requestHeadersPayload = computed<UsageDetailPayload | null>(() => props.detail?.request_headers || null)
const requestPayload = computed<UsageDetailPayload | null>(() => props.detail?.request || null)
const responseHeadersPayload = computed<UsageDetailPayload | null>(() => props.detail?.response_headers || null)
const responsePayload = computed<UsageDetailPayload | null>(() => props.detail?.response || null)

const getPayloadRaw = (payload: UsageDetailPayload | null, frameIndex = 0) => {
  if (!payload) {
    return ''
  }
  if (payload.kind === 'frames') {
    return payload.frames?.[frameIndex] || ''
  }
  return payload.body || ''
}

const getPayloadJSON = (payload: UsageDetailPayload | null, frameIndex = 0) => {
  if (!payload) {
    return null
  }
  if (payload.kind === 'body' && !payload.is_json) {
    return null
  }
  const raw = getPayloadRaw(payload, frameIndex)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

const canShowJSONView = (payload: UsageDetailPayload | null, frameIndex = 0) => {
  if (!payload || !payload.is_json) {
    return false
  }
  return getPayloadRaw(payload, frameIndex).trim().length > 0
}

const requestHeadersRaw = computed(() => getPayloadRaw(requestHeadersPayload.value))
const requestHeadersCanShowJSON = computed(() => canShowJSONView(requestHeadersPayload.value))
const requestHeadersJSON = computed(() => getPayloadJSON(requestHeadersPayload.value))

const requestPayloadRaw = computed(() => getPayloadRaw(requestPayload.value, selectedRequestFrameIndex.value))
const requestPayloadCanShowJSON = computed(() => canShowJSONView(requestPayload.value, selectedRequestFrameIndex.value))
const requestPayloadJSON = computed(() => getPayloadJSON(requestPayload.value, selectedRequestFrameIndex.value))

const responseHeadersRaw = computed(() => getPayloadRaw(responseHeadersPayload.value))
const responseHeadersCanShowJSON = computed(() => canShowJSONView(responseHeadersPayload.value))
const responseHeadersJSON = computed(() => getPayloadJSON(responseHeadersPayload.value))

const responsePayloadRaw = computed(() => getPayloadRaw(responsePayload.value, selectedResponseFrameIndex.value))
const responsePayloadCanShowJSON = computed(() => canShowJSONView(responsePayload.value, selectedResponseFrameIndex.value))
const responsePayloadJSON = computed(() => getPayloadJSON(responsePayload.value, selectedResponseFrameIndex.value))

watch([() => props.show, () => props.detail, activeSection], () => {
  selectedRequestFrameIndex.value = 0
  selectedResponseFrameIndex.value = 0
  requestHeadersViewMode.value = requestHeadersCanShowJSON.value ? 'json' : 'raw'
  requestViewMode.value = requestPayloadCanShowJSON.value ? 'json' : 'raw'
  responseHeadersViewMode.value = responseHeadersCanShowJSON.value ? 'json' : 'raw'
  responseViewMode.value = responsePayloadCanShowJSON.value ? 'json' : 'raw'
}, { immediate: true })

const handleClose = () => {
  emit('close')
}
</script>
