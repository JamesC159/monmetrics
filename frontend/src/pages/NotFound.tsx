import { Link } from 'react-router-dom'
import { Home, Search, ArrowLeft } from 'lucide-react'

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center px-4">
      <div className="text-center">
        <div className="mb-8">
          <h1 className="text-9xl font-bold gradient-text mb-4">404</h1>
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">Page Not Found</h2>
          <p className="text-xl text-gray-300 mb-8 max-w-md mx-auto">
            Sorry, we couldn't find the page you're looking for. It might have been moved or deleted.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link to="/" className="button-primary inline-flex items-center">
            <Home className="w-5 h-5 mr-2" />
            Go Home
          </Link>

          <Link to="/search" className="button-secondary inline-flex items-center">
            <Search className="w-5 h-5 mr-2" />
            Search Cards
          </Link>
        </div>

        <div className="mt-8">
          <button
            onClick={() => window.history.back()}
            className="text-gray-400 hover:text-white transition-colors inline-flex items-center"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </button>
        </div>
      </div>
    </div>
  )
}