import { renderToString } from 'react-dom/server'
import { StaticRouter } from 'react-router-dom/server'
import App from './App'

export function render(url: string) {
  const html = renderToString(
    <StaticRouter location={url}>
      <App />
    </StaticRouter>
  )

  return {
    html,
    head: `
      <meta property="og:title" content="MonMetrics - Professional Trading Card Analysis" />
      <meta property="og:description" content="Track, analyze, and predict trading card prices with advanced technical indicators" />
      <meta property="og:type" content="website" />
      <meta name="twitter:card" content="summary_large_image" />
    `
  }
}