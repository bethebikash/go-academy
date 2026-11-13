import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import VerifyCertificate from './VerifyCertificate'
import * as api from '../lib/api'

vi.mock('../lib/api', () => ({
  apiRequest: vi.fn(),
}))

const mockCert = {
  id: 'cert-99',
  user_id: 'user-42',
  learning_path_id: 'path-42',
  certificate_code: 'GOACAD-VERIFY-1234',
  issued_at: '2026-11-13T10:00:00Z',
  learning_path_title: 'Production Go Concurrency Track',
  recipient_name: 'Bikash Bhattarai',
}

describe('VerifyCertificate Public Route', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders verified certificate details when code matches', async () => {
    vi.mocked(api.apiRequest).mockResolvedValueOnce(mockCert)

    render(
      <MemoryRouter initialEntries={['/certificates/verify/GOACAD-VERIFY-1234']}>
        <Routes>
          <Route path="/certificates/verify/:code" element={<VerifyCertificate />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Certificate of Mastery')).toBeInTheDocument()
      expect(screen.getByText('Verified Official Credential')).toBeInTheDocument()
      expect(screen.getByText('Bikash Bhattarai')).toBeInTheDocument()
      expect(screen.getByText('Production Go Concurrency Track')).toBeInTheDocument()
      expect(screen.getByText('GOACAD-VERIFY-1234')).toBeInTheDocument()
    })
  })

  it('renders invalid certificate message when lookup fails', async () => {
    vi.mocked(api.apiRequest).mockRejectedValueOnce(new Error('Certificate not found'))

    render(
      <MemoryRouter initialEntries={['/certificates/verify/INVALID-CODE']}>
        <Routes>
          <Route path="/certificates/verify/:code" element={<VerifyCertificate />} />
        </Routes>
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('Invalid or Unrecognized Certificate')).toBeInTheDocument()
      expect(screen.getByText('INVALID-CODE')).toBeInTheDocument()
    })
  })
})
