import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UsageView from '../UsageView.vue'

const { list, getStats, getSnapshotV2, getModelStats, getById, getDetail, listErrorLogs, route } = vi.hoisted(() => {
  vi.stubGlobal('localStorage', {
    getItem: vi.fn(() => null),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  })

  return {
    list: vi.fn(),
    getStats: vi.fn(),
    getSnapshotV2: vi.fn(),
    getModelStats: vi.fn(),
    getById: vi.fn(),
    getDetail: vi.fn(),
    listErrorLogs: vi.fn(),
    route: {
      query: {} as Record<string, unknown>,
    },
  }
})

const messages: Record<string, string> = {
  'admin.dashboard.timeRange': 'Time Range',
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
      getModelStats,
    },
    users: {
      getById,
    },
  },
}))

vi.mock('@/api/admin/usage', () => ({
  adminUsageAPI: {
    list: vi.fn(),
    getDetail,
  },
}))

vi.mock('@/api/admin/ops', () => ({
  listErrorLogs,
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
const UsageTableStub = {
  emits: ['userClick'],
  template: '<div data-test="usage-table"><button class="user-click" @click="$emit(\'userClick\', 2)">user</button></div>',
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

const baseStubs = {
  AppLayout: AppLayoutStub,
  UsageStatsCards: true,
  UsageFilters: UsageFiltersStub,
  UsageTable: true,
  UsageExportProgress: true,
  UsageCleanupDialog: true,
  UsageDetailModal: true,
  UserBalanceHistoryModal: true,
  Pagination: true,
  Select: true,
  DateRangePicker: true,
  Icon: true,
  TokenUsageTrend: true,
  EndpointDistributionChart: true,
  ModelDistributionChart: ModelDistributionChartStub,
  GroupDistributionChart: GroupDistributionChartStub,
  OpsErrorLogTable: true,
  OpsErrorDetailModal: true,
}

const mountUsageView = async (stubs: Record<string, unknown> = {}) => {
  const wrapper = mount(UsageView, {
    global: {
      stubs: {
        ...baseStubs,
        ...stubs,
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

const mockBaseResponses = () => {
  list.mockResolvedValue({ items: [], total: 0, pages: 0 })
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
  getModelStats.mockResolvedValue({ models: [] })
  listErrorLogs.mockResolvedValue({ items: [], total: 0, pages: 0 })
}

describe('admin UsageView distribution metric toggles', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(FIXED_NOW)
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getById.mockReset()
    getDetail.mockReset()
    listErrorLogs.mockReset()
    route.query = {}
    mockBaseResponses()
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

  it('keeps previous model stats visible during refresh until new data arrives', async () => {
    getModelStats.mockResolvedValueOnce({ models: [{ model: 'A', total_tokens: 10 }] })

    const wrapper = await mountUsageView()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'A', total_tokens: 10 }])

    let resolveSecond: (v: any) => void = () => {}
    getModelStats.mockReturnValueOnce(new Promise((res) => { resolveSecond = res }))
    ;(wrapper.vm as any).refreshData()
    await flushPromises()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'A', total_tokens: 10 }])

    resolveSecond({ models: [{ model: 'B', total_tokens: 20 }] })
    await flushPromises()
    expect((wrapper.vm as any).requestedModelStats).toEqual([{ model: 'B', total_tokens: 20 }])
  })

  it('keeps model and group metric toggles independent without refetching chart data', async () => {
    const wrapper = await mountUsageView()

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      ...DEFAULT_DATE_RANGE,
      granularity: 'day',
    }))

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

describe('admin UsageView handleUserClick', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getById.mockReset()
    listErrorLogs.mockReset()
    route.query = {}
    mockBaseResponses()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('opens user via include_deleted when clicking a usage row user', async () => {
    getById.mockResolvedValue({ id: 2, email: 'd@test.com', deleted_at: '2026-05-28T00:00:00Z' })

    const wrapper = await mountUsageView({ UsageTable: UsageTableStub })

    await wrapper.find('[data-test="usage-table"] .user-click').trigger('click')
    await flushPromises()

    expect(getById).toHaveBeenCalledWith(2, true)
  })
})

describe('admin UsageView errors tab filter forwarding', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    list.mockReset()
    getStats.mockReset()
    getSnapshotV2.mockReset()
    getModelStats.mockReset()
    getById.mockReset()
    listErrorLogs.mockReset()
    route.query = {}
    mockBaseResponses()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('forwards model/account_id/group_id to listErrorLogs on the errors tab', async () => {
    const wrapper = await mountUsageView()

    const vm = wrapper.vm as any
    vm.filters.model = 'gpt-5.3-codex'
    vm.filters.account_id = 7
    vm.filters.group_id = 3
    await flushPromises()

    const tabs = wrapper.findAll('button.tab')
    await tabs[1].trigger('click')
    await flushPromises()

    expect(listErrorLogs).toHaveBeenCalledWith(expect.objectContaining({
      view: 'all',
      model: 'gpt-5.3-codex',
      account_id: 7,
      group_id: 3,
    }))
  })
})
