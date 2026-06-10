<template>
  <div class="space-y-3">
    <div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex flex-1 items-center gap-2">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('admin.usage.detail.searchText')"
          class="input-field"
        />
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ matchSummary }}
        </span>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="matchCount === 0" @click="moveToPrevious">
          {{ t('admin.usage.detail.prev') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="matchCount === 0" @click="moveToNext">
          {{ t('admin.usage.detail.next') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="copyRaw">
          {{ t('admin.usage.detail.copyRaw') }}
        </button>
      </div>
    </div>

    <textarea
      ref="textareaRef"
      :value="content"
      readonly
      spellcheck="false"
      class="h-[28rem] w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 font-mono text-xs leading-6 text-gray-900 outline-none dark:border-dark-600 dark:bg-dark-900 dark:text-gray-100"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  content: string
}>()

const { t } = useI18n()
const appStore = useAppStore()
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const searchQuery = ref('')
const currentMatchIndex = ref(0)

const matchIndexes = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const source = props.content.toLowerCase()
  if (!query || !source) {
    return []
  }
  const indexes: number[] = []
  let start = 0
  while (start <= source.length) {
    const idx = source.indexOf(query, start)
    if (idx < 0) {
      break
    }
    indexes.push(idx)
    start = idx + Math.max(query.length, 1)
  }
  return indexes
})

const matchCount = computed(() => matchIndexes.value.length)

const matchSummary = computed(() => {
  if (!searchQuery.value.trim()) {
    return t('admin.usage.detail.noSearch')
  }
  if (matchCount.value === 0) {
    return t('admin.usage.detail.noMatches')
  }
  return t('admin.usage.detail.matchSummary', {
    current: currentMatchIndex.value + 1,
    total: matchCount.value
  })
})

const focusMatch = (index: number) => {
  const textarea = textareaRef.value
  const query = searchQuery.value.trim()
  const indexes = matchIndexes.value
  if (!textarea || !query || indexes.length === 0) {
    return
  }
  const normalizedIndex = (index + indexes.length) % indexes.length
  currentMatchIndex.value = normalizedIndex
  const start = indexes[normalizedIndex]
  const end = start + query.length
  textarea.focus()
  textarea.setSelectionRange(start, end)
}

const moveToPrevious = () => focusMatch(currentMatchIndex.value - 1)

const moveToNext = () => focusMatch(currentMatchIndex.value + 1)

const copyRaw = async () => {
  try {
    await navigator.clipboard.writeText(props.content)
    appStore.showSuccess(t('admin.usage.detail.copySuccess'))
  } catch {
    appStore.showError(t('admin.usage.detail.copyFailed'))
  }
}

watch(searchQuery, () => {
  currentMatchIndex.value = 0
  if (matchIndexes.value.length > 0) {
    focusMatch(0)
  }
})

watch(() => props.content, () => {
  currentMatchIndex.value = 0
})
</script>
