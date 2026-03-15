import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const { list, getStats, getSnapshotV2, getById, route } = vi.hoisted(() => {
  vi.stubGlobal('localStorage', {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  })

  return {
    list: vi.fn(),
    getStats: vi.fn(),
    getSnapshotV2: vi.fn(),
    getById: vi.fn(),
    route: {
      query: {} as Record<string, unknown>,
    },
  }
})

const messages: Record<string, string> = {
  'admin.dashboard.day': 'Day',
  'admin.dashboard.hour': 'Hour',
  'admin.usage.failedToLoadUser': 'Failed to load user',
}

vi.mock('@/api/admin', () => ({
  adminAPI: {
    usage: {
      list,
      getStats,
    },
    dashboard: {
      getSnapshotV2,
    },
    users: {
      getById,
    },
  },
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    list: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => route,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showWarning: vi.fn(),
    showSuccess: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/utils/format', () => ({
  formatReasoningEffort: (value: string | null | undefined) => value ?? '-',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const UsageFiltersStub = {
  emits: ['reset'],
  template: `
    <div>
      <button data-test="reset-filters" @click="$emit('reset')">reset</button>
      <slot name="after-reset" />
    </div>
  `,
}
const ModelDistributionChartStub = {
  props: ['metric'],
  emits: ['update:metric'],
  template: `
    <div data-test="model-chart">
      <span class="metric">{{ metric }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}
const GroupDistributionChartStub = {
  props: ['metric'],
  emits: ['update:metric'],
  template: `
    <div data-test="group-chart">
      <span class="metric">{{ metric }}</span>
      <button class="switch-metric" @click="$emit('update:metric', 'actual_cost')">switch</button>
    </div>
  `,
}

const FIXED_NOW = new Date(2026, 2, 15, 12, 0, 0)
const DEFAULT_DATE_RANGE = {
  start_date: '2026-02-14',
  end_date: '2026-03-15',
}

const mountUsageView = async () => {
  const wrapper = mount(UsageView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        UsageStatsCards: true,
        UsageFilters: UsageFiltersStub,
        UsageTable: true,
        UsageExportProgress: true,
        UsageCleanupDialog: true,
        UserBalanceHistoryModal: true,
        Pagination: true,
        Select: true,
        Icon: true,
        TokenUsageTrend: true,
        EndpointDistributionChart: true,
        ModelDistributionChart: ModelDistributionChartStub,
        GroupDistributionChart: GroupDistributionChartStub,
      },
    },
  })

  vi.advanceTimersByTime(120)
  await flushPromises()
  return wrapper
}

const expectRequestsToUseDateRange = (expected: { start_date: string; end_date: string }) => {
  expect(list).toHaveBeenLastCalledWith(
    expect.objectContaining(expected),
    expect.objectContaining({ signal: expect.any(Object) })
  )
  expect(getStats).toHaveBeenLastCalledWith(expect.objectContaining(expected))
  expect(getSnapshotV2).toHaveBeenLastCalledWith(expect.objectContaining(expected))
}

describe('admin UsageView distribution metric toggles', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(FIXED_NOW)
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getById.mockReset()
    route.query = {}

    list.mockResolvedValue({
      items: [],
      total: 0,
      pages: 0,
    })
    getStats.mockResolvedValue({
      total_requests: 0,
      total_input_tokens: 0,
      total_output_tokens: 0,
      total_cache_tokens: 0,
      total_tokens: 0,
      total_cost: 0,
      total_actual_cost: 0,
      average_duration_ms: 0,
    })
    getSnapshotV2.mockResolvedValue({
      trend: [],
      models: [],
      groups: [],
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads the admin usage page with the last 30 days by default', async () => {
    await mountUsageView()

    expectRequestsToUseDateRange(DEFAULT_DATE_RANGE)
  })

  it('resets filters back to the last 30 days range', async () => {
    const wrapper = await mountUsageView()

    list.mockClear()
    getStats.mockClear()
    getSnapshotV2.mockClear()

    await wrapper.find('[data-test="reset-filters"]').trigger('click')
    await flushPromises()

    expectRequestsToUseDateRange(DEFAULT_DATE_RANGE)
  })

  it('keeps explicit route query dates instead of overriding them with the default range', async () => {
    route.query = {
      start_date: '2026-03-01',
      end_date: '2026-03-02',
      user_id: '42',
    }

    await mountUsageView()

    expect(list).toHaveBeenLastCalledWith(
      expect.objectContaining({
        start_date: '2026-03-01',
        end_date: '2026-03-02',
        user_id: 42,
      }),
      expect.objectContaining({ signal: expect.any(Object) })
    )
    expect(getStats).toHaveBeenLastCalledWith(expect.objectContaining({
      start_date: '2026-03-01',
      end_date: '2026-03-02',
      user_id: 42,
    }))
    expect(getSnapshotV2).toHaveBeenLastCalledWith(expect.objectContaining({
      start_date: '2026-03-01',
      end_date: '2026-03-02',
      user_id: 42,
    }))
  })

  it('keeps model and group metric toggles independent without refetching chart data', async () => {
    const wrapper = await mountUsageView()

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)

    const modelChart = wrapper.find('[data-test="model-chart"]')
    const groupChart = wrapper.find('[data-test="group-chart"]')

    expect(modelChart.find('.metric').text()).toBe('tokens')
    expect(groupChart.find('.metric').text()).toBe('tokens')

    await modelChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(groupChart.find('.metric').text()).toBe('tokens')
    expect(getSnapshotV2).toHaveBeenCalledTimes(1)

    await groupChart.find('.switch-metric').trigger('click')
    await flushPromises()

    expect(modelChart.find('.metric').text()).toBe('actual_cost')
    expect(groupChart.find('.metric').text()).toBe('actual_cost')
    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
  })
})
