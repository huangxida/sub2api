<template>
  <div class="space-y-3">
    <div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
      <div class="flex flex-1 items-center gap-2">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('admin.usage.detail.searchJson')"
          class="input-field"
        />
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button type="button" class="btn btn-secondary btn-sm" @click="expandAll">
          {{ t('admin.usage.detail.expandAll') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="collapseAll">
          {{ t('admin.usage.detail.collapseAll') }}
        </button>
        <button type="button" class="btn btn-secondary btn-sm" @click="copyRaw">
          {{ t('admin.usage.detail.copyRaw') }}
        </button>
      </div>
    </div>

    <div
      v-if="quickCopyEntries.length || selectedPathLabel"
      class="space-y-2 rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-900"
    >
      <div v-if="quickCopyEntries.length" class="flex flex-wrap items-center gap-2">
        <span class="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
          {{ t('admin.usage.detail.quickCopy') }}
        </span>
        <button
          v-for="entry in quickCopyEntries"
          :key="entry.pathText"
          type="button"
          class="rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 font-mono text-xs text-gray-700 transition-colors hover:border-primary-300 hover:text-primary-700 dark:border-dark-500 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-500 dark:hover:text-primary-300"
          :title="entry.pathText"
          @click="copyQuickEntry(entry)"
        >
          {{ entry.label }}
        </button>
      </div>

      <div v-if="selectedPathLabel" class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex min-w-0 items-center gap-2">
          <span class="text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
            {{ t('admin.usage.detail.currentNode') }}
          </span>
          <span class="truncate rounded-md bg-white px-2 py-1 font-mono text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-200">
            {{ selectedPathLabel }}
          </span>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button type="button" class="btn btn-secondary btn-sm" @click="copySelectedNodeContent">
            {{ t('admin.usage.detail.copyValue') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" @click="copySelectedNodeJson">
            {{ t('admin.usage.detail.copyJson') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" @click="copySelectedNodePath">
            {{ t('admin.usage.detail.copyPath') }}
          </button>
        </div>
      </div>
    </div>

    <div ref="containerRef" class="json-tree-viewer"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import JSONEditor from 'jsoneditor'
import 'jsoneditor/dist/jsoneditor.css'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  value: unknown
  raw: string
}>()

interface QuickCopyEntry {
  label: string
  path: Array<string | number>
  pathText: string
}

interface JsonEditorMenuItem {
  text?: string
  title?: string
  className?: string
  type?: string
  submenu?: JsonEditorMenuItem[]
  click?: () => void
}

const QUICK_COPY_PRIORITY_KEYS = ['instructions', 'system', 'input', 'messages', 'tools', 'metadata']

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const containerRef = ref<HTMLElement | null>(null)
const searchQuery = ref('')
const selectedPath = ref<Array<string | number> | null>(null)
let editor: JSONEditor | null = null

const isRecord = (value: unknown): value is Record<string, unknown> => (
  typeof value === 'object' && value !== null && !Array.isArray(value)
)

const getValueAtPath = (value: unknown, path: Array<string | number>): unknown => {
  let current: unknown = value

  for (const segment of path) {
    if (Array.isArray(current)) {
      if (typeof segment !== 'number') return undefined
      current = current[segment]
      continue
    }

    if (!isRecord(current) || !Object.prototype.hasOwnProperty.call(current, segment)) {
      return undefined
    }

    current = current[String(segment)]
  }

  return current
}

const formatJsonPath = (path: Array<string | number>): string => {
  if (path.length === 0) return '$'

  return path.reduce<string>((acc, segment) => {
    if (typeof segment === 'number') {
      return `${acc}[${segment}]`
    }
    if (/^[A-Za-z_$][\w$]*$/.test(segment)) {
      return `${acc}.${segment}`
    }
    return `${acc}[${JSON.stringify(segment)}]`
  }, '$')
}

const serializeNodeValue = (value: unknown, asJson = false): string => {
  if (asJson) {
    return JSON.stringify(value, null, 2) ?? ''
  }

  if (typeof value === 'string') return value
  if (value == null) return String(value)
  if (typeof value === 'number' || typeof value === 'boolean') return String(value)

  return JSON.stringify(value, null, 2) ?? ''
}

const estimateValueSize = (value: unknown): number => {
  if (typeof value === 'string') return value.trim().length
  if (typeof value === 'number' || typeof value === 'boolean') return String(value).length
  if (value == null) return 0
  try {
    return JSON.stringify(value)?.length ?? 0
  } catch {
    return 0
  }
}

const shouldOfferQuickCopy = (value: unknown): boolean => {
  if (typeof value === 'string') return value.trim().length > 0
  if (Array.isArray(value)) return value.length > 0
  if (isRecord(value)) return Object.keys(value).length > 0
  return value != null
}

const quickCopyEntries = computed<QuickCopyEntry[]>(() => {
  if (!isRecord(props.value)) {
    return []
  }

  const root = props.value
  const seen = new Set<string>()
  const entries: QuickCopyEntry[] = []

  const pushEntry = (key: string) => {
    if (seen.has(key) || !Object.prototype.hasOwnProperty.call(root, key)) {
      return
    }

    const value = root[key]
    if (!shouldOfferQuickCopy(value)) {
      return
    }

    const path = [key]
    entries.push({
      label: key,
      path,
      pathText: formatJsonPath(path)
    })
    seen.add(key)
  }

  QUICK_COPY_PRIORITY_KEYS.forEach(pushEntry)

  Object.keys(root)
    .sort((left, right) => estimateValueSize(root[right]) - estimateValueSize(root[left]))
    .forEach(pushEntry)

  return entries.slice(0, 8)
})

const selectedPathLabel = computed(() => {
  if (!selectedPath.value) {
    return ''
  }
  return formatJsonPath(selectedPath.value)
})

const copyValueByPath = async (
  path: Array<string | number>,
  options?: {
    asJson?: boolean
    successMessage?: string
  }
) => {
  const value = getValueAtPath(props.value, path)
  if (typeof value === 'undefined') {
    appStore.showError(t('admin.usage.detail.copyFailed'))
    return
  }

  const text = serializeNodeValue(value, options?.asJson)
  await copyToClipboard(text, options?.successMessage)
}

const ensureEditor = () => {
  if (editor || !containerRef.value) {
    return
  }

  editor = new JSONEditor(containerRef.value, {
    mode: 'tree',
    mainMenuBar: false,
    navigationBar: false,
    statusBar: false,
    search: true,
    onEditable: () => false,
    onEvent: (node: { path?: Array<string | number> }, event: Event) => {
      if (!Array.isArray(node.path)) {
        return
      }
      if (event.type === 'click' || event.type === 'focus' || event.type === 'contextmenu') {
        selectedPath.value = [...node.path]
      }
    },
    onCreateMenu: (items: JsonEditorMenuItem[], node: { type?: string, path?: Array<string | number> }) => {
      if (!Array.isArray(node.path) || node.type === 'append') {
        return items
      }

      const path = [...node.path]
      const fieldName = path.length > 0 ? String(path[path.length - 1]) : '$'
      const extraItems: JsonEditorMenuItem[] = [
        {
          text: t('admin.usage.detail.copyValue'),
          title: t('admin.usage.detail.copyField', { field: fieldName }),
          className: 'jsoneditor-type-auto',
          click: () => {
            selectedPath.value = path
            void copyValueByPath(path, {
              successMessage: t('admin.usage.detail.copyFieldSuccess', { field: fieldName })
            })
          }
        },
        {
          text: t('admin.usage.detail.copyJson'),
          title: t('admin.usage.detail.copyFieldJson', { field: fieldName }),
          className: 'jsoneditor-type-auto',
          click: () => {
            selectedPath.value = path
            void copyValueByPath(path, {
              asJson: true,
              successMessage: t('admin.usage.detail.copyJsonSuccess')
            })
          }
        },
        {
          text: t('admin.usage.detail.copyPath'),
          title: t('admin.usage.detail.copyPath'),
          className: 'jsoneditor-type-auto',
          click: () => {
            selectedPath.value = path
            void copyToClipboard(formatJsonPath(path), t('admin.usage.detail.copyPathSuccess'))
          }
        }
      ]

      return [...items, { type: 'separator' }, ...extraItems]
    }
  }, props.value)
}

const setEditorValue = (value: unknown) => {
  ensureEditor()
  if (!editor) {
    return
  }

  editor.update(value)
  if (searchQuery.value.trim()) {
    editor.search(searchQuery.value.trim())
  }
}

const expandAll = () => editor?.expandAll()

const collapseAll = () => editor?.collapseAll()

const copyRaw = async () => {
  try {
    await navigator.clipboard.writeText(props.raw)
    appStore.showSuccess(t('admin.usage.detail.copySuccess'))
  } catch {
    appStore.showError(t('admin.usage.detail.copyFailed'))
  }
}

const copyQuickEntry = async (entry: QuickCopyEntry) => {
  await copyValueByPath(entry.path, {
    successMessage: t('admin.usage.detail.copyFieldSuccess', { field: entry.label })
  })
}

const copySelectedNodeContent = async () => {
  if (!selectedPath.value) return
  const fieldName = selectedPath.value.length > 0 ? String(selectedPath.value[selectedPath.value.length - 1]) : '$'
  await copyValueByPath(selectedPath.value, {
    successMessage: t('admin.usage.detail.copyFieldSuccess', { field: fieldName })
  })
}

const copySelectedNodeJson = async () => {
  if (!selectedPath.value) return
  await copyValueByPath(selectedPath.value, {
    asJson: true,
    successMessage: t('admin.usage.detail.copyJsonSuccess')
  })
}

const copySelectedNodePath = async () => {
  if (!selectedPath.value) return
  await copyToClipboard(selectedPathLabel.value, t('admin.usage.detail.copyPathSuccess'))
}

watch(searchQuery, (value) => {
  editor?.search(value.trim())
})

watch(() => props.value, (value) => {
  setEditorValue(value)
  if (selectedPath.value && typeof getValueAtPath(value, selectedPath.value) === 'undefined') {
    selectedPath.value = null
  }
}, { deep: true })

onMounted(() => {
  ensureEditor()
})

onBeforeUnmount(() => {
  editor?.destroy()
  editor = null
})
</script>

<style>
.json-tree-viewer .jsoneditor {
  border-radius: 0.75rem;
  border: 1px solid rgb(229 231 235);
}

.json-tree-viewer .jsoneditor-outer,
.json-tree-viewer .jsoneditor-tree {
  overflow: auto;
}

.json-tree-viewer .jsoneditor-menu {
  display: none;
}

.json-tree-viewer .jsoneditor-navigation-bar {
  display: none;
}

.json-tree-viewer .jsoneditor-statusbar {
  display: none;
}

.json-tree-viewer div.jsoneditor-field {
  min-width: 120px;
  white-space: nowrap;
  word-break: normal;
  overflow-wrap: normal;
}

.json-tree-viewer div.jsoneditor-value,
.json-tree-viewer a.jsoneditor-value {
  word-break: break-word;
  overflow-wrap: anywhere;
}

.dark .json-tree-viewer .jsoneditor,
.dark .json-tree-viewer .jsoneditor-tree,
.dark .json-tree-viewer .jsoneditor-outer {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39);
  color: rgb(243 244 246);
}
</style>
