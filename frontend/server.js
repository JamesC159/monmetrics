import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import express from 'express'
import compression from 'compression'
import sirv from 'sirv'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const isProduction = process.env.NODE_ENV === 'production'
const port = process.env.PORT || 3000
const base = process.env.BASE || '/'

// Cached production assets
const templateHtml = isProduction
  ? fs.readFileSync(path.resolve(__dirname, 'dist/client/index.html'), 'utf-8')
  : ''
const ssrManifest = isProduction
  ? fs.readFileSync(path.resolve(__dirname, 'dist/client/ssr-manifest.json'), 'utf-8')
  : undefined

// Create http server
const app = express()

// Add compression middleware
app.use(compression())

// Add Vite or respective production middlewares
let vite
if (!isProduction) {
  const { createServer } = await import('vite')
  vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    base
  })
  app.use(vite.ssrLoadModule)
} else {
  app.use(base, sirv(path.resolve(__dirname, 'dist/client'), { extensions: [] }))
}

// Serve HTML
app.use('*', async (req, res) => {
  try {
    const url = req.originalUrl.replace(base, '')

    let template
    let render
    if (!isProduction) {
      // Always read fresh template in development
      template = fs.readFileSync(path.resolve(__dirname, 'index.html'), 'utf-8')
      template = await vite.transformIndexHtml(url, template)
      render = (await vite.ssrLoadModule('/src/entry-server.tsx')).render
    } else {
      template = templateHtml
      render = (await import(path.resolve(__dirname, 'dist/server/entry-server.js'))).render
    }

    const rendered = await render(url, ssrManifest)

    const html = template
      .replace(`<!--app-html-->`, rendered.html ?? '')
      .replace(
        `<head>`,
        `<head>\n${rendered.head ?? ''}`
      )

    res.status(200).set({ 'Content-Type': 'text/html' }).send(html)
  } catch (e) {
    vite?.ssrFixStacktrace(e)
    console.log(e.stack)
    res.status(500).end(e.stack)
  }
})

// Start http server
app.listen(port, () => {
  console.log(`Server started at http://localhost:${port}`)