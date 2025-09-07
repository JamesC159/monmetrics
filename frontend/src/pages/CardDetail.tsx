import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, TrendingUp, TrendingDown } from 'lucide-react'

export default function CardDetail() {
  const { id } = useParams<{ id: string }>()

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link
            to="/search"
            className="inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Search
          </Link>

          <h1 className="text-4xl font-bold text-white mb-2">Card Detail</h1>
          <p className="text-gray-400">Detailed analysis for card ID: {id}</p>
        </div>

        <div className="glass-effect rounded-2xl p-8">
          <div className="text-center py-16">
            <TrendingUp className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-400 mb-2">Card Analysis Coming Soon</h3>
            <p className="text-gray-500">
              Advanced charting and technical analysis will be available here
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}