import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import { TrendingDown, TrendingUp, Wallet } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { PortfolioSummary as Summary } from '@/types'
import { formatPercentage, formatPrice } from '@/utils/formatters'
import { gainClass, ITEM_TYPE_LABELS, signed } from '@/utils/portfolio'
import { INDICATOR_COLORS } from '@/utils/indicators'

export default function PortfolioSummary({ summary }: { summary: Summary }) {
  const allocation = summary.by_game.filter((s) => s.value > 0)
  return (
    <div className='grid grid-cols-1 lg:grid-cols-3 gap-6'>
      <div className='lg:col-span-2 grid grid-cols-2 md:grid-cols-4 gap-4'>
        <StatCard
          label='Total value'
          value={formatPrice(summary.total_value)}
          icon={<Wallet className='w-5 h-5 text-primary-400' />}
        />
        <StatCard label='Cost basis' value={formatPrice(summary.cost_basis)} />
        <StatCard
          label='Gain / loss'
          value={signed(formatPrice(summary.gain_loss), summary.gain_loss)}
          sub={signed(formatPercentage(summary.gain_loss_pct), summary.gain_loss_pct)}
          valueClass={gainClass(summary.gain_loss)}
        />
        <StatCard
          label='Items'
          value={String(summary.item_count)}
          sub={`${summary.unit_count} units${summary.unvalued_count ? ` · ${summary.unvalued_count} unvalued` : ''}`}
        />

        <MoverList
          title='Top gainers'
          icon={<TrendingUp className='w-4 h-4 text-success-400' />}
          movers={summary.top_gainers}
        />
        <MoverList
          title='Top losers'
          icon={<TrendingDown className='w-4 h-4 text-error-400' />}
          movers={summary.top_losers}
        />
      </div>

      <div className='glass-effect rounded-2xl p-4'>
        <h4 className='text-sm font-semibold text-white mb-2'>Allocation by game</h4>
        {allocation.length === 0 ? (
          <p className='text-sm text-gray-500 py-8 text-center'>No valued items yet</p>
        ) : (
          <>
            <div className='h-40'>
              <ResponsiveContainer width='100%' height='100%'>
                <PieChart>
                  <Pie
                    data={allocation}
                    dataKey='value'
                    nameKey='key'
                    innerRadius={40}
                    outerRadius={70}
                    paddingAngle={2}
                  >
                    {allocation.map((s, i) => (
                      <Cell key={s.key} fill={INDICATOR_COLORS[i % INDICATOR_COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(v: number) => formatPrice(v)}
                    contentStyle={{
                      backgroundColor: '#1F2937',
                      border: '1px solid #374151',
                      borderRadius: 8,
                      color: '#F9FAFB',
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <ul className='mt-2 space-y-1 text-xs'>
              {allocation.map((s, i) => (
                <li key={s.key} className='flex items-center justify-between text-gray-300'>
                  <span className='flex items-center gap-2'>
                    <span
                      className='w-2 h-2 rounded-full'
                      style={{ backgroundColor: INDICATOR_COLORS[i % INDICATOR_COLORS.length] }}
                    />
                    {s.key}
                  </span>
                  <span>
                    {formatPrice(s.value)} (
                    {summary.total_value > 0
                      ? Math.round((s.value / summary.total_value) * 100)
                      : 0}
                    %)
                  </span>
                </li>
              ))}
            </ul>
            <div className='mt-3 pt-3 border-t border-white/10 flex flex-wrap gap-2'>
              {summary.by_item_type.map((s) => (
                <span key={s.key} className='badge badge-neutral text-xs'>
                  {ITEM_TYPE_LABELS[s.key as keyof typeof ITEM_TYPE_LABELS] ?? s.key}: {s.count}
                </span>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}

function StatCard({
  label,
  value,
  sub,
  icon,
  valueClass = 'text-white',
}: {
  label: string
  value: string
  sub?: string
  icon?: React.ReactNode
  valueClass?: string
}) {
  return (
    <div className='glass-effect rounded-xl p-4'>
      <div className='flex items-center justify-between text-xs text-gray-400 mb-1'>
        {label}
        {icon}
      </div>
      <div className={`text-xl font-bold ${valueClass}`}>{value}</div>
      {sub && <div className='text-xs text-gray-400 mt-1'>{sub}</div>}
    </div>
  )
}

function MoverList({
  title,
  icon,
  movers,
}: {
  title: string
  icon: React.ReactNode
  movers: Summary['top_gainers']
}) {
  return (
    <div className='col-span-2 glass-effect rounded-xl p-4'>
      <div className='flex items-center gap-2 text-xs text-gray-400 mb-2'>
        {icon}
        {title}
      </div>
      {movers.length === 0 ? (
        <p className='text-xs text-gray-500'>None</p>
      ) : (
        <ul className='space-y-1'>
          {movers.map((m) => (
            <li key={m.id} className='flex items-center justify-between text-sm'>
              <Link
                to={`/portfolio/${m.id}`}
                className='text-gray-200 hover:text-primary-300 truncate mr-2'
              >
                {m.display_name}
              </Link>
              <span className={gainClass(m.gain_loss)}>
                {signed(formatPrice(m.gain_loss), m.gain_loss)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
