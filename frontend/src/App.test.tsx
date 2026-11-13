import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

describe('App Routing and Auth Screens', () => {
  it('renders login screen when unauthenticated', async () => {
    render(<App />)
    expect(await screen.findByText('Sign in to Go Academy')).toBeInTheDocument()
    expect(screen.getByText('Email Address')).toBeInTheDocument()
    expect(screen.getByText('Password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    expect(screen.getByText(/invite-only platform/i)).toBeInTheDocument()
  })
})
