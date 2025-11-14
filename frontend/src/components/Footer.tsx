import { Link } from 'react-router-dom'
import { Sparkles, Github, Twitter, Mail } from 'lucide-react'

export const Footer = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className='bg-dark-950 border-t border-primary-500/20 mt-auto'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12'>
        <div className='grid grid-cols-1 md:grid-cols-4 gap-8'>
          {/* Brand */}
          <div className='space-y-4'>
            <Link to='/' className='flex items-center group'>
              <Sparkles className='w-6 h-6 text-primary-400 mr-2 group-hover:animate-pulse' />
              <h3 className='text-2xl font-display font-bold gradient-text-gold'>MonMetrics</h3>
            </Link>
            <p className='text-gray-400 text-sm'>
              Professional trading analysis for the modern card collector.
            </p>
            <div className='flex space-x-4'>
              <a
                href='https://github.com'
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-400 hover:text-primary-400 transition-colors'
              >
                <Github className='w-5 h-5' />
              </a>
              <a
                href='https://twitter.com'
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-400 hover:text-secondary-400 transition-colors'
              >
                <Twitter className='w-5 h-5' />
              </a>
              <a
                href='mailto:contact@monmetrics.com'
                className='text-gray-400 hover:text-accent-400 transition-colors'
              >
                <Mail className='w-5 h-5' />
              </a>
            </div>
          </div>

          {/* Product */}
          <div>
            <h4 className='text-white font-display font-semibold mb-4'>Product</h4>
            <ul className='space-y-2'>
              <li>
                <Link
                  to='/search'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Search Cards
                </Link>
              </li>
              <li>
                <a
                  href='#features'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Features
                </a>
              </li>
              <li>
                <a
                  href='#pricing'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Pricing
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Mobile App
                </a>
              </li>
            </ul>
          </div>

          {/* Company */}
          <div>
            <h4 className='text-white font-display font-semibold mb-4'>Company</h4>
            <ul className='space-y-2'>
              <li>
                <a
                  href='#about'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  About Us
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Blog
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Careers
                </a>
              </li>
              <li>
                <a
                  href='mailto:contact@monmetrics.com'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Contact
                </a>
              </li>
            </ul>
          </div>

          {/* Support */}
          <div>
            <h4 className='text-white font-display font-semibold mb-4'>Support</h4>
            <ul className='space-y-2'>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Help Center
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Documentation
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Privacy Policy
                </a>
              </li>
              <li>
                <a
                  href='#'
                  className='text-gray-400 hover:text-primary-400 transition-colors text-sm'
                >
                  Terms of Service
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className='divider my-8'></div>

        <div className='flex flex-col md:flex-row justify-between items-center text-sm text-gray-400'>
          <p>&copy; {currentYear} MonMetrics. All rights reserved.</p>
          <p className='mt-2 md:mt-0'>
            Built with <span className='text-error-400'>❤</span> for collectors
          </p>
        </div>
      </div>
    </footer>
  )
}
