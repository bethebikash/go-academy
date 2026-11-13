import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { BrowserRouter } from 'react-router-dom'
import Progress from './Progress'
import * as api from '../lib/api'

vi.mock('../lib/api', () => ({
  apiRequest: vi.fn(),
}))

const mockStats = {
  streak: {
    current_streak: 5,
    longest_streak: 12,
    last_activity_date: new Date().toISOString(),
  },
  completed_lessons: 14,
  solved_challenges: 3,
  total_notes: 7,
  total_certificates: 1,
  earned_badges_count: 2,
  total_badges_count: 5,
}

const mockActivity = [
  { date: new Date().toISOString().split('T')[0], count: 4 },
]

const mockBadges = [
  {
    key: 'first_step',
    name: 'First Step',
    description: 'Complete your first lesson.',
    category: 'lessons',
    icon: 'Footprints',
    awarded: true,
    awarded_at: new Date().toISOString(),
  },
  {
    key: 'streak_7',
    name: 'Week on Fire',
    description: 'Maintain a 7-day learning streak.',
    category: 'streaks',
    icon: 'Flame',
    awarded: false,
  },
]

const mockCertificates = [
  {
    id: 'cert-1',
    user_id: 'user-1',
    learning_path_id: 'path-1',
    certificate_code: 'GOACAD-TEST-1234',
    issued_at: new Date().toISOString(),
    learning_path_title: 'Go Foundations Track',
    recipient_name: 'Gopher User',
  },
]

const mockPaths = [
  {
    id: 'path-1',
    title: 'Go Foundations Track',
    slug: 'go-foundations-track',
    description: 'Master Go from zero to hero.',
    status: 'published',
    sort_order: 1,
  },
]

describe('Progress and Gamification Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.apiRequest).mockImplementation((endpoint: string) => {
      if (endpoint === '/analytics/summary') return Promise.resolve(mockStats)
      if (endpoint.startsWith('/analytics/activity')) return Promise.resolve(mockActivity)
      if (endpoint === '/analytics/badges') return Promise.resolve(mockBadges)
      if (endpoint === '/certificates/me') return Promise.resolve(mockCertificates)
      if (endpoint === '/learning-paths') return Promise.resolve(mockPaths)
      return Promise.reject(new Error('Unknown endpoint: ' + endpoint))
    })
  })

  it('renders stats overview, streak cards, and recent achievements', async () => {
    render(
      <BrowserRouter>
        <Progress />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Active Streak')).toBeInTheDocument()
      expect(screen.getByText('5')).toBeInTheDocument()
      expect(screen.getByText('14')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument()
    })

    expect(screen.getByText('90-Day Learning Activity')).toBeInTheDocument()
    expect(screen.getByText('Best record: 12 consecutive days')).toBeInTheDocument()
  })

  it('switches to badges tab and displays unlocked and locked badges', async () => {
    render(
      <BrowserRouter>
        <Progress />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Active Streak')).toBeInTheDocument()
    })

    const badgesTab = screen.getByRole('button', { name: /badges/i })
    fireEvent.click(badgesTab)

    expect(screen.getByText('First Step')).toBeInTheDocument()
    expect(screen.getByText('Week on Fire')).toBeInTheDocument()
    expect(screen.getByText('Complete your first lesson.')).toBeInTheDocument()
    expect(screen.getByText('Maintain a 7-day learning streak.')).toBeInTheDocument()
  })

  it('switches to certificates tab and shows earned certificates with verification links', async () => {
    render(
      <BrowserRouter>
        <Progress />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Active Streak')).toBeInTheDocument()
    })

    const certsTab = screen.getByRole('button', { name: /certificates/i })
    fireEvent.click(certsTab)

    expect(screen.getByText('GOACAD-TEST-1234')).toBeInTheDocument()
    expect(screen.getByText('Go Foundations Track')).toBeInTheDocument()
    expect(screen.getByText('Public Verification')).toBeInTheDocument()
    expect(screen.getByText('View Certificate Card')).toBeInTheDocument()
  })
})
