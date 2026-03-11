<template>
  <BaseDialog :show="show" :title="t('admin.usage.detail.title')" width="full" @close="handleClose">
    <div class="space-y-4">
      <div v-if="usage" class="grid grid-cols-1 gap-3 rounded-xl border border-gray-200 bg-gray-50 p-4 text-sm dark:border-dark-600 dark:bg-dark-900 md:grid-cols-4">
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

      <div v-else-if="activePayload" class="space-y-4">
        <div class="flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-900 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex flex-wrap items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
            <span>{{ t('admin.usage.detail.contentType') }}: {{ activePayload.content_type || '-' }}</span>
            <span>{{ t('admin.usage.detail.sizeBytes') }}: {{ activePayload.size_bytes }}</span>
            <span v-if="activePayload.complete != null">
              {{ t('admin.usage.detail.complete') }}:
              {{ activePayload.complete ? t('common.yes') : t('common.no') }}
            </span>
          </div>

          <div v-if="activePayload.kind === 'frames'" class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-300" for="frame-select">
              {{ t('admin.usage.detail.frameSelect') }}
            </label>
            <select id="frame-select" v-model.number="selectedFrameIndex" class="input-field w-40">
              <option v-for="(_, index) in activePayload.frames || []" :key="index" :value="index">
                {{ t('admin.usage.detail.frameLabel', { index: index + 1 }) }}
              </option>
            </select>
          </div>
        </div>

        <div v-if="showJsonToggle" class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
            :class="activeViewMode === 'json'
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
            @click="activeViewMode = 'json'"
          >
            {{ t('admin.usage.detail.jsonView') }}
          </button>
          <button
            type="button"
            class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors"
            :class="activeViewMode === 'raw'
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'"
            @click="activeViewMode = 'raw'"
          >
            {{ t('admin.usage.detail.rawView') }}
          </button>
        </div>

        <JsonTreeViewer
          v-if="activeViewMode === 'json' && activeJSON != null"
          :value="activeJSON"
          :raw="activeRaw"
        />
        <TextSearchViewer
          v-else
          :content="activeRaw"
        />
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
const activeViewMode = ref<'json' | 'raw'>('json')
const selectedFrameIndex = ref(0)

const requestTypeLabel = computed(() => {
  const requestType = props.detail?.request_type || props.usage?.request_type
  if (requestType === 'ws_v2') return t('usage.ws')
  if (requestType === 'stream') return t('usage.stream')
  if (requestType === 'sync') return t('usage.sync')
  return t('usage.unknown')
})

const activePayload = computed<UsageDetailPayload | null>(() => {
  if (!props.detail) {
    return null
  }
  return activeSection.value === 'request'
    ? (props.detail.request || null)
    : (props.detail.response || null)
})

const activeRaw = computed(() => {
  const payload = activePayload.value
  if (!payload) {
    return ''
  }
  if (payload.kind === 'frames') {
    return payload.frames?.[selectedFrameIndex.value] || ''
  }
  return payload.body || ''
})

const activeJSON = computed(() => {
  const payload = activePayload.value
  if (!payload) {
    return null
  }
  if (payload.kind === 'body' && !payload.is_json) {
    return null
  }
  if (payload.kind === 'frames') {
    const frame = payload.frames?.[selectedFrameIndex.value] || ''
    if (!frame) {
      return null
    }
  }
  if (!activeRaw.value) {
    return null
  }
  try {
    return JSON.parse(activeRaw.value)
  } catch {
    return null
  }
})

const showJsonToggle = computed(() => activeJSON.value !== null)

watch([() => props.show, () => props.detail, activeSection], () => {
  selectedFrameIndex.value = 0
  activeViewMode.value = activeJSON.value !== null ? 'json' : 'raw'
}, { immediate: true })

const handleClose = () => {
  emit('close')
}
</script>
