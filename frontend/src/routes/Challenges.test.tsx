import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import Challenges from './Challenges'
import * as api from '../lib/api'

vi.mock('../lib/api', () => ({
  apiRequest: vi.fn(),
}))

const mockChallenges = [
  {
    id: 'c1',
    title: 'Two Sum in Go',
    slug: 'two-sum-in-go',
    description: 'Find two indices that sum to target.',
    difficulty: 'beginner',
    estimated_minutes: 15,
    starter_code: 'package main',
    status: 'published',
    sort_order: 1,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c2',
    title: 'Thread-Safe LRU Cache',
    slug: 'thread-safe-lru-cache',
    description: 'Implement an O(1) concurrent LRU cache.',
    difficulty: 'advanced',
    estimated_minutes: 35,
    starter_code: 'package main',
    status: 'published',
    sort_order: 2,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
]

describe('Challenges Catalog View', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders challenges after loading', async () => {
    vi.mocked(api.apiRequest).mockResolvedValueOnce(mockChallenges)

    render(
      <BrowserRouter>
        <Challenges />
      </BrowserRouter>,
    )

    expect(screen.getByText('Interactive Go Challenges')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByText('Two Sum in Go')).toBeInTheDocument()
      expect(screen.getByText('Thread-Safe LRU Cache')).toBeInTheDocument()
    })

    expect(screen.getByPlaceholderText('Search challenges...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^all$/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^beginner$/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^advanced$/i })).toBeInTheDocument()
  })

  it('renders empty state when no challenges exist', async () => {
    vi.mocked(api.apiRequest).mockResolvedValueOnce([])

    render(
      <BrowserRouter>
        <Challenges />
      </BrowserRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No challenges found')).toBeInTheDocument()
    })
  })
})
