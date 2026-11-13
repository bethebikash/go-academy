import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import MarkdownViewer from './MarkdownViewer'

describe('MarkdownViewer', () => {
  it('renders markdown headers, paragraphs, and code blocks correctly', () => {
    const markdown = `
# Concurrency in Go

Go routines are lightweight threads managed by the Go runtime.

\`\`\`go
func main() {
    go worker()
}
\`\`\`

> Go makes concurrency easy.
`
    render(<MarkdownViewer content={markdown} />)

    expect(screen.getByText('Concurrency in Go')).toBeInTheDocument()
    expect(screen.getByText(/Go routines are lightweight threads/i)).toBeInTheDocument()
    expect(screen.getByText(/go worker\(\)/i)).toBeInTheDocument()
    expect(screen.getByText(/Go makes concurrency easy/i)).toBeInTheDocument()
    expect(screen.getByText('go')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /copy/i })).toBeInTheDocument()
  })
})
