// src/components/LandingPage.tsx
import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  TrendingUp,
  BarChart3,
  Zap,
  Shield,
  ArrowRight,
  Play,
  Star,
  CheckCircle,
  Sparkles,
} from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { Navigation } from './Navigation'
import { Footer } from './Footer'

const LandingPage = () => {
  const [isVisible, setIsVisible] = useState(false)
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    setIsVisible(true)
  }, [])

  const features = [
    {
      icon: <TrendingUp className='w-8 h-8' />,
      title: 'Advanced Price Analytics',
      description: 'Track price movements with 5 years of historical data from eBay and TCGPlayer',
    },
    {
      icon: <BarChart3 className='w-8 h-8' />,
      title: 'Technical Indicators',
      description:
        'Apply professional trading indicators like Bollinger Bands, RSI, and moving averages',
    },
    {
      icon: <Zap className='w-8 h-8' />,
      title: 'Real-time Updates',
      description: 'Get instant market updates and price alerts for your favorite cards',
    },
    {
      icon: <Shield className='w-8 h-8' />,
      title: 'Secure & Reliable',
      description: 'Enterprise-grade security with 99.9% uptime guarantee',
    },
  ]

  const pricingPlans = [
    {
      name: 'Free',
      price: '$0',
      period: 'forever',
      features: [
        'Up to 3 technical indicators',
        'Basic price history',
        'Limited search results',
        'Community support',
      ],
      buttonText: 'Get Started Free',
      popular: false,
    },
    {
      name: 'Pro',
      price: '$19',
      period: 'month',
      features: [
        'Up to 10 technical indicators',
        '5 years price history',
        'Unlimited searches',
        'Advanced charting tools',
        'Price alerts',
        'Priority support',
        'Export data',
      ],
      buttonText: 'Start Pro Trial',
      popular: true,
    },
  ]

  const testimonials = [
    {
      name: 'Alex Chen',
      role: 'Pokemon Card Investor',
      content:
        'MonMetrics helped me identify market trends that increased my portfolio value by 40%',
      rating: 5,
    },
    {
      name: 'Sarah Williams',
      role: 'MTG Collector',
      content:
        'The technical analysis tools are incredible. Finally, professional trading analysis for cards!',
      rating: 5,
    },
    {
      name: 'Mike Rodriguez',
      role: 'Sports Card Dealer',
      content:
        'I wish I had found this platform sooner. The price predictions are remarkably accurate.',
      rating: 5,
    },
  ]

  const handleGetStarted = () => {
    if (isAuthenticated) {
      navigate('/dashboard')
    } else {
      navigate('/register')
    }
  }

  return (
    <div className='min-h-screen bg-dark-950'>
      <Navigation />

      {/* Hero Section - More Subtle */}
      <section className='pt-32 pb-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-dark-900 via-dark-950 to-black relative overflow-hidden'>
        {/* Decorative background elements */}
        <div className='absolute top-0 left-0 w-96 h-96 bg-primary-500/5 rounded-full blur-3xl'></div>
        <div className='absolute bottom-0 right-0 w-96 h-96 bg-secondary-500/5 rounded-full blur-3xl'></div>

        <div className='max-w-7xl mx-auto relative z-10'>
          <div
            className={`text-center transition-all duration-1000 ${
              isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-10'
            }`}
          >
            <div className='inline-flex items-center px-4 py-2 bg-white/5 rounded-full border border-primary-500/20 mb-6 animate-fade-in'>
              <Sparkles className='w-4 h-4 text-primary-400 mr-2' />
              <span className='text-sm text-gray-300'>Track, Analyze, Invest</span>
            </div>

            <h1 className='text-4xl md:text-6xl font-display font-bold text-white mb-6 leading-tight'>
              Smart Analytics for
              <br />
              <span className='gradient-text-gold'>Trading Card Collectors</span>
            </h1>

            <p className='text-lg md:text-xl text-gray-400 mb-12 max-w-3xl mx-auto'>
              Professional-grade price tracking and technical analysis tools designed for the modern
              collector.
            </p>

            <div className='flex flex-col sm:flex-row gap-4 justify-center items-center mb-16'>
              <button onClick={handleGetStarted} className='btn-primary group'>
                {isAuthenticated ? 'Go to Dashboard' : 'Start Free'}
                <ArrowRight className='inline ml-2 w-5 h-5 group-hover:translate-x-1 transition-transform' />
              </button>
            </div>

            {/* Hero Stats */}
            <div className='grid grid-cols-2 md:grid-cols-4 gap-8 max-w-4xl mx-auto'>
              <div className='text-center'>
                <div className='text-3xl font-bold gradient-text-gold mb-2'>1M+</div>
                <div className='text-gray-500 text-sm'>Cards Tracked</div>
              </div>
              <div className='text-center'>
                <div className='text-3xl font-bold gradient-text-cyan mb-2'>5 Years</div>
                <div className='text-gray-500 text-sm'>Price History</div>
              </div>
              <div className='text-center'>
                <div className='text-3xl font-bold gradient-text-purple mb-2'>10+</div>
                <div className='text-gray-500 text-sm'>Indicators</div>
              </div>
              <div className='text-center'>
                <div className='text-3xl font-bold text-success-400 mb-2'>50K+</div>
                <div className='text-gray-500 text-sm'>Active Users</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id='features' className='py-20 px-4 sm:px-6 lg:px-8 section-darker'>
        <div className='max-w-7xl mx-auto'>
          <div className='text-center mb-16 animate-fade-in-up'>
            <h2 className='text-4xl md:text-5xl font-display font-bold text-white mb-6'>
              Powerful Features for
              <span className='gradient-text-gold'> Smart Collecting</span>
            </h2>
            <p className='text-xl text-gray-400 max-w-3xl mx-auto'>
              Everything you need to make informed trading card investment decisions
            </p>
          </div>

          <div className='grid md:grid-cols-2 lg:grid-cols-4 gap-8'>
            {features.map((feature, index) => (
              <div
                key={index}
                className='card card-hover group'
                style={{ animationDelay: `${index * 100}ms` }}
              >
                <div className='text-primary-400 mb-4 group-hover:scale-110 group-hover:text-primary-300 transition-all duration-300'>
                  {feature.icon}
                </div>
                <h3 className='text-xl font-display font-semibold text-white mb-3'>
                  {feature.title}
                </h3>
                <p className='text-gray-400 leading-relaxed'>{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Demo Section */}
      <section className='py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-dark-900 via-dark-950 to-black'>
        <div className='max-w-7xl mx-auto'>
          <div className='text-center mb-16'>
            <h2 className='text-4xl md:text-5xl font-display font-bold text-white mb-6'>
              See MonMetrics in Action
            </h2>
            <p className='text-xl text-gray-400 max-w-3xl mx-auto'>
              Professional-grade analysis tools designed specifically for trading cards
            </p>
          </div>

          <div className='treasure-card'>
            <div className='bg-dark-800/50 rounded-2xl p-6 aspect-video flex items-center justify-center'>
              <div className='text-center'>
                <div className='w-24 h-24 mx-auto mb-6 bg-gradient-gold rounded-full flex items-center justify-center shadow-glow-gold'>
                  <Play className='w-12 h-12 text-dark-950 ml-2' />
                </div>
                <h3 className='text-2xl font-display font-semibold text-white mb-4'>
                  Interactive Chart Demo
                </h3>
                <p className='text-gray-400 mb-6'>
                  See how technical indicators can help identify optimal buy/sell opportunities
                </p>
                <Link to='/search' className='btn-ghost inline-block'>
                  Try Live Demo
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Pricing Section */}
      <section id='pricing' className='py-20 px-4 sm:px-6 lg:px-8 section-darker'>
        <div className='max-w-7xl mx-auto'>
          <div className='text-center mb-16'>
            <h2 className='text-4xl md:text-5xl font-display font-bold text-white mb-6'>
              Choose Your Plan
            </h2>
            <p className='text-xl text-gray-400 max-w-3xl mx-auto'>
              Start free and upgrade when you're ready for professional features
            </p>
          </div>

          <div className='grid md:grid-cols-2 gap-8 max-w-4xl mx-auto'>
            {pricingPlans.map((plan, index) => (
              <div
                key={index}
                className={`relative card card-hover ${
                  plan.popular
                    ? 'border-primary-500/50 ring-2 ring-primary-500/20 shadow-glow-gold'
                    : ''
                }`}
              >
                {plan.popular && (
                  <div className='absolute -top-3 left-1/2 transform -translate-x-1/2'>
                    <span className='badge badge-gold px-4 py-1'>Most Popular</span>
                  </div>
                )}

                <div className='text-center mb-8'>
                  <h3 className='text-2xl font-display font-bold text-white mb-4'>{plan.name}</h3>
                  <div className='flex items-baseline justify-center'>
                    <span className='text-5xl font-bold gradient-text-gold'>{plan.price}</span>
                    <span className='text-gray-400 ml-2'>/{plan.period}</span>
                  </div>
                </div>

                <ul className='space-y-4 mb-8'>
                  {plan.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className='flex items-center text-gray-300'>
                      <CheckCircle className='w-5 h-5 text-success-400 mr-3 flex-shrink-0' />
                      {feature}
                    </li>
                  ))}
                </ul>

                <button
                  onClick={handleGetStarted}
                  className={`w-full py-4 rounded-xl font-semibold transition-all duration-300 ${
                    plan.popular ? 'btn-primary' : 'btn-ghost'
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
      <section className='py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-dark-900 via-dark-950 to-black'>
        <div className='max-w-7xl mx-auto'>
          <div className='text-center mb-16'>
            <h2 className='text-4xl md:text-5xl font-display font-bold text-white mb-6'>
              Trusted by Collectors Worldwide
            </h2>
            <p className='text-xl text-gray-400 max-w-3xl mx-auto'>
              See what our users are saying about their MonMetrics experience
            </p>
          </div>

          <div className='grid md:grid-cols-3 gap-8'>
            {testimonials.map((testimonial, index) => (
              <div key={index} className='card card-hover'>
                <div className='flex mb-4'>
                  {[...Array(testimonial.rating)].map((_, i) => (
                    <Star key={i} className='w-5 h-5 text-primary-400 fill-current' />
                  ))}
                </div>
                <p className='text-gray-300 mb-6 italic'>"{testimonial.content}"</p>
                <div>
                  <div className='font-semibold text-white'>{testimonial.name}</div>
                  <div className='text-gray-400 text-sm'>{testimonial.role}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <Footer />
    </div>
  )
}

export default LandingPage
