import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search as SearchIcon, Filter, ArrowLeft } from 'lucide-react'

export default function Search() {
  const [query, setQuery] = useState('')

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link
            to="/"
            className="inline-flex items-center text-gray-400 hover:text-white transition-colors mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Home
          </Link>

          <h1 className="text-4xl font-bold text-white mb-2">Search Trading Cards</h1>
          <p className="text-gray-400">Find and analyze your favorite trading cards</p>
        </div>

        <div className="glass-effect rounded-2xl p-8">
          <div className="flex gap-4 mb-6">
            <div className="flex-1 relative">
              <SearchIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                className="w-full pl-12 pr-4 py-4 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-lg"
                placeholder="Search for Pokemon, Magic, Yu-Gi-Oh cards..."
              />
            </div>
            <button className="button-secondary flex items-center">
              <Filter className="w-5 h-5 mr-2" />
              Filters
            </button>
          </div>

          <div className="text-center py-16">
            <SearchIcon className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-400 mb-2">Start Your Search</h3>
            <p className="text-gray-500">
              Enter a card name, set, or game to find detailed price analysis and charts
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}