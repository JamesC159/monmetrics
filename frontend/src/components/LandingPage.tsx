// src/components/LandingPage.tsx
import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  TrendingUp,
  BarChart3,
  Zap,
  Shield,
  Search,
  ArrowRight,
  Play,
  Star,
  CheckCircle,
  Menu,
  X
} from 'lucide-react'
import { useAuth } from '@/context/AuthContext'

const LandingPage = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [isVisible, setIsVisible] = useState(false)
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    setIsVisible(true)
  }, [])

  const features = [
    {
      icon: <TrendingUp className="w-8 h-8" />,
      title: "Advanced Price Analytics",
      description: "Track price movements with 5 years of historical data from eBay and TCGPlayer"
    },
    {
      icon: <BarChart3 className="w-8 h-8" />,
      title: "Technical Indicators",
      description: "Apply professional trading indicators like Bollinger Bands, RSI, and moving averages"
    },
    {
      icon: <Zap className="w-8 h-8" />,
      title: "Real-time Updates",
      description: "Get instant market updates and price alerts for your favorite cards"
    },
    {
      icon: <Shield className="w-8 h-8" />,
      title: "Secure & Reliable",
      description: "Enterprise-grade security with 99.9% uptime guarantee"
    }
  ]

  const pricingPlans = [
    {
      name: "Free",
      price: "$0",
      period: "forever",
      features: [
        "Up to 3 technical indicators",
        "Basic price history",
        "Limited search results",
        "Community support"
      ],
      buttonText: "Get Started Free",
      popular: false
    },
    {
      name: "Pro",
      price: "$19",
      period: "month",
      features: [
        "Up to 10 technical indicators",
        "5 years price history",
        "Unlimited searches",
        "Advanced charting tools",
        "Price alerts",
        "Priority support",
        "Export data"
      ],
      buttonText: "Start Pro Trial",
      popular: true
    }
  ]

  const testimonials = [
    {
      name: "Alex Chen",
      role: "Pokemon Card Investor",
      content: "MonMetrics helped me identify market trends that increased my portfolio value by 40%",
      rating: 5
    },
    {
      name: "Sarah Williams",
      role: "MTG Collector",
      content: "The technical analysis tools are incredible. Finally, professional trading analysis for cards!",
      rating: 5
    },
    {
      name: "Mike Rodriguez",
      role: "Sports Card Dealer",
      content: "I wish I had found this platform sooner. The price predictions are remarkably accurate.",
      rating: 5
    }
  ]

  const handleGetStarted = () => {
    if (isAuthenticated) {
      navigate('/dashboard')
    } else {
      navigate('/register')
    }
  }

  const handleSignIn = () => {
    if (isAuthenticated) {
      navigate('/dashboard')
    } else {
      navigate('/login')
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
      {/* Navigation */}
      <nav className="fixed top-0 w-full z-50 bg-black/20 backdrop-blur-md border-b border-white/10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Link to="/" className="flex-shrink-0">
                <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                  MonMetrics
                </h1>
              </Link>
            </div>

            <div className="hidden md:block">
              <div className="ml-10 flex items-baseline space-x-8">
                <a href="#features" className="text-gray-300 hover:text-white transition-colors">Features</a>
                <a href="#pricing" className="text-gray-300 hover:text-white transition-colors">Pricing</a>
                <a href="#about" className="text-gray-300 hover:text-white transition-colors">About</a>
                <Link to="/search" className="text-gray-300 hover:text-white transition-colors">Search</Link>
                <button
                  onClick={handleSignIn}
                  className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors"
                >
                  {isAuthenticated ? 'Dashboard' : 'Sign In'}
                </button>
              </div>
            </div>

            <div className="md:hidden">
              <button
                onClick={() => setIsMenuOpen(!isMenuOpen)}
                className="text-gray-300 hover:text-white"
              >
                {isMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile menu */}
        {isMenuOpen && (
          <div className="md:hidden bg-black/50 backdrop-blur-md">
            <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3">
              <a href="#features" className="block text-gray-300 hover:text-white px-3 py-2">Features</a>
              <a href="#pricing" className="block text-gray-300 hover:text-white px-3 py-2">Pricing</a>
              <a href="#about" className="block text-gray-300 hover:text-white px-3 py-2">About</a>
              <Link to="/search" className="block text-gray-300 hover:text-white px-3 py-2">Search</Link>
              <button
                onClick={handleSignIn}
                className="w-full text-left bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-lg"
              >
                {isAuthenticated ? 'Dashboard' : 'Sign In'}
              </button>
            </div>
          </div>
        )}
      </nav>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className={`text-center transition-all duration-1000 ${isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-10'}`}>
            <h1 className="text-5xl md:text-7xl font-bold text-white mb-6">
              <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
                Professional Trading Analysis
              </span>
              <br />
              <span className="text-white">for Trading Cards</span>
            </h1>

            <p className="text-xl md:text-2xl text-gray-300 mb-12 max-w-4xl mx-auto leading-relaxed">
              Track, analyze, and predict trading card prices with advanced technical indicators.
              Make informed investment decisions with 5 years of market data.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-16">
              <button
                onClick={handleGetStarted}
                className="group bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-8 py-4 rounded-xl font-semibold text-lg transition-all duration-300 shadow-2xl hover:shadow-blue-500/25 animate-pulse-glow"
              >
                {isAuthenticated ? 'Go to Dashboard' : 'Start Free Analysis'}
                <ArrowRight className="inline ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </button>

              <Link
                to="/search"
                className="group flex items-center text-white px-8 py-4 rounded-xl font-semibold text-lg border border-white/20 hover:bg-white/10 transition-all duration-300"
              >
                <Search className="w-5 h-5 mr-2 group-hover:scale-110 transition-transform" />
                Search Cards
              </Link>
            </div>

            {/* Hero Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8 max-w-3xl mx-auto">
              <div className="text-center">
                <div className="text-3xl font-bold text-blue-400 mb-2">1M+</div>
                <div className="text-gray-400">Cards Tracked</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-purple-400 mb-2">5 Years</div>
                <div className="text-gray-400">Price History</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-pink-400 mb-2">10+</div>
                <div className="text-gray-400">Indicators</div>
              </div>
              <div className="text-center">
                <div className="text-3xl font-bold text-green-400 mb-2">50K+</div>
                <div className="text-gray-400">Users</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 px-4 sm:px-6 lg:px-8 bg-black/20">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
              Powerful Features for
              <span className="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent"> Smart Investing</span>
            </h2>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto">
              Everything you need to make informed trading card investment decisions
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {features.map((feature, index) => (
              <div
                key={index}
                className="group bg-white/5 backdrop-blur-sm rounded-2xl p-8 border border-white/10 hover:border-blue-500/50 transition-all duration-300 hover:transform hover:scale-105"
              >
                <div className="text-blue-400 mb-4 group-hover:scale-110 transition-transform duration-300">
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold text-white mb-3">{feature.title}</h3>
                <p className="text-gray-300 leading-relaxed">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Demo Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
              See MonMetrics in Action
            </h2>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto">
              Professional-grade analysis tools designed specifically for trading cards
            </p>
          </div>

          <div className="bg-gradient-to-br from-blue-900/30 to-purple-900/30 rounded-3xl p-8 border border-white/10">
            <div className="bg-black/50 rounded-2xl p-6 aspect-video flex items-center justify-center">
              <div className="text-center">
                <div className="w-24 h-24 mx-auto mb-6 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center">
                  <Play className="w-12 h-12 text-white ml-2" />
                </div>
                <h3 className="text-2xl font-semibold text-white mb-4">Interactive Chart Demo</h3>
                <p className="text-gray-300 mb-6">
                  See how Bollinger Bands and RSI can help identify optimal buy/sell opportunities
                </p>
                <Link
                  to="/search"
                  className="bg-white/10 hover:bg-white/20 text-white px-6 py-3 rounded-lg transition-colors inline-block"
                >
                  Try Live Demo
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="py-20 px-4 sm:px-6 lg:px-8 bg-black/20">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
              Choose Your Plan
            </h2>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto">
              Start free and upgrade when you're ready for professional features
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
            {pricingPlans.map((plan, index) => (
              <div
                key={index}
                className={`relative bg-white/5 backdrop-blur-sm rounded-2xl p-8 border transition-all duration-300 hover:transform hover:scale-105 ${
                  plan.popular
                    ? 'border-blue-500/50 ring-2 ring-blue-500/20'
                    : 'border-white/10 hover:border-blue-500/30'
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                    <span className="bg-gradient-to-r from-blue-500 to-purple-500 text-white px-4 py-1 rounded-full text-sm font-semibold">
                      Most Popular
                    </span>
                  </div>
                )}

                <div className="text-center mb-8">
                  <h3 className="text-2xl font-bold text-white mb-4">{plan.name}</h3>
                  <div className="flex items-baseline justify-center">
                    <span className="text-5xl font-bold text-white">{plan.price}</span>
                    <span className="text-gray-400 ml-2">/{plan.period}</span>
                  </div>
                </div>

                <ul className="space-y-4 mb-8">
                  {plan.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-center text-gray-300">
                      <CheckCircle className="w-5 h-5 text-green-400 mr-3 flex-shrink-0" />
                      {feature}
                    </li>
                  ))}
                </ul>

                <button
                  onClick={handleGetStarted}
                  className={`w-full py-4 rounded-xl font-semibold transition-all duration-300 ${
                    plan.popular
                      ? 'bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white shadow-xl hover:shadow-blue-500/25'
                      : 'bg-white/10 hover:bg-white/20 text-white border border-white/20'
                  }`}
                >
                  {plan.buttonText}
                </button>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
              Trusted by Collectors Worldwide
            </h2>
            <p className="text-xl text-gray-300 max-w-3xl mx-auto">
              See what our users are saying about their MonMetrics experience
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {testimonials.map((testimonial, index) => (
              <div
                key={index}
                className="bg-white/5 backdrop-blur-sm rounded-2xl p-8 border border-white/10 hover:border-blue-500/30 transition-all duration-300"
              >
                <div className="flex mb-4">
                  {[...Array(testimonial.rating)].map((_, i) => (
                    <Star key={i} className="w-5 h-5 text-yellow-400 fill-current" />
                  ))}
                </div>
                <p className="text-gray-300 mb-6 italic">"{testimonial.content}"</p>
                <div>
                  <div className="font-semibold text-white">{testimonial.name}</div>
                  <div className="text-gray-400 text-sm">{testimonial.role}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-r from-blue-900/50 to-purple-900/50">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
            Ready to Start Trading Smarter?
          </h2>
          <p className="text-xl text-gray-300 mb-8">
            Join thousands of collectors using professional analysis to maximize their investments
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <button
              onClick={handleGetStarted}
              className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-8 py-4 rounded-xl font-semibold text-lg transition-all duration-300 shadow-2xl hover:shadow-blue-500/25"
            >
              {isAuthenticated ? 'Go to Dashboard' : 'Start Free Today'}
            </button>
            <a
              href="mailto:contact@monmetrics.com"
              className="text-white px-8 py-4 rounded-xl font-semibold text-lg border border-white/20 hover:bg-white/10 transition-all duration-300 inline-block"
            >
              Contact Sales
            </a>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-black/50 py-12 px-4 sm:px-6 lg:px-8 border-t border-white/10">
        <div className="max-w-7xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8">
            <div>
              <Link to="/">
                <h3 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent mb-4">
                  MonMetrics
                </h3>
              </Link>
              <p className="text-gray-400">
                Professional trading analysis for the modern card collector.
              </p>
            </div>

            <div>
              <h4 className="text-white font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#features" className="hover:text-white transition-colors">Features</a></li>
                <li><a href="#pricing" className="hover:text-white transition-colors">Pricing</a></li>
                <li><Link to="/search" className="hover:text-white transition-colors">Search</Link></li>
                <li><a href="#" className="hover:text-white transition-colors">Mobile App</a></li>
              </ul>
            </div>

            <div>
              <h4 className="text-white font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#about" className="hover:text-white transition-colors">About</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Blog</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Careers</a></li>
                <li><a href="mailto:contact@monmetrics.com" className="hover:text-white transition-colors">Contact</a></li>
              </ul>
            </div>

            <div>
              <h4 className="text-white font-semibold mb-4">Support</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#" className="hover:text-white transition-colors">Help Center</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Documentation</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Status</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Privacy</a></li>
              </ul>
            </div>
          </div>

          <div className="border-t border-white/10 mt-12 pt-8 text-center text-gray-400">
            <p>&copy; 2025 MonMetrics. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default LandingPage