import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

describe('App Component', () => {
  it('renders Go Academy title and navigation sections', () => {
    render(<App />)
    expect(screen.getByText('Go Academy')).toBeInTheDocument()
    expect(screen.getByText('Interactive Learning')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Learning Paths')).toBeInTheDocument()
    expect(screen.getAllByText('Challenges').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Platform Overview')).toBeInTheDocument()
  })
})
