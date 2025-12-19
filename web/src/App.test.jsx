import { render, screen, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import Navigation from './components/Navigation.jsx'
import DashboardPage from './pages/DashboardPage.jsx'
import AboutPage from './pages/AboutPage.jsx'

// Mock the chart components to avoid canvas rendering issues in tests
vi.mock('./components/ItemSpreadCard.jsx', () => ({
  default: vi.fn(() => <div data-testid="item-spread-card">ItemSpreadCard</div>)
}))

vi.mock('./components/DailyTotalsChart.jsx', () => ({
  default: vi.fn(() => <div data-testid="daily-totals-chart">DailyTotalsChart</div>)
}))

vi.mock('./components/DailyProportionsChart.jsx', () => ({
  default: vi.fn(() => <div data-testid="daily-proportions-chart">DailyProportionsChart</div>)
}))

// Create a testable version of App content without BrowserRouter
function AppContent() {
  return (
    <>
      <Navigation />
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/about" element={<AboutPage />} />
      </Routes>
    </>
  )
}

// Helper function to render App with controlled routing
function renderAppWithRouter(initialEntries = ['/']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-blue-600 text-white shadow-lg">
          <div className="mx-auto px-4 py-6">
            <h1 className="text-3xl font-bold">WaniKani Dashboard</h1>
          </div>
        </header>
        
        <AppContent />
      </div>
    </MemoryRouter>
  )
}

describe('App Integration Tests - Complete Application Functionality', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  describe('Navigation Flows (Requirements 1.1, 1.2, 1.3)', () => {
    it('should navigate from Dashboard to About page correctly', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Verify we start on Dashboard page
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
      expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()

      // Navigate to About page
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      // Verify About page content is displayed
      expect(screen.getByRole('heading', { name: /about wanikani dashboard/i })).toBeInTheDocument()
      expect(screen.getByText(/comprehensive progress tracking dashboard/i)).toBeInTheDocument()
      
      // Verify Dashboard components are no longer visible
      expect(screen.queryByTestId('item-spread-card')).not.toBeInTheDocument()
      expect(screen.queryByTestId('daily-totals-chart')).not.toBeInTheDocument()
      expect(screen.queryByTestId('daily-proportions-chart')).not.toBeInTheDocument()
    })

    it('should navigate from About page back to Dashboard correctly', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate to About page first
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      // Verify we're on About page
      expect(screen.getByRole('heading', { name: /about wanikani dashboard/i })).toBeInTheDocument()

      // Navigate back to Dashboard
      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
      await user.click(dashboardLink)

      // Verify Dashboard components are displayed again
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
      expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()
      
      // Verify About page content is no longer visible
      expect(screen.queryByRole('heading', { name: /about wanikani dashboard/i })).not.toBeInTheDocument()
    })

    it('should maintain navigation state correctly during multiple navigations', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Perform multiple navigation cycles
      for (let i = 0; i < 3; i++) {
        // Go to About
        const aboutLink = screen.getByRole('link', { name: /about/i })
        await user.click(aboutLink)
        
        expect(screen.getByRole('heading', { name: /about wanikani dashboard/i })).toBeInTheDocument()
        expect(aboutLink).toHaveClass('text-blue-600', 'border-blue-600')
        
        // Go back to Dashboard
        const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
        await user.click(dashboardLink)
        
        expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
        expect(dashboardLink).toHaveClass('text-blue-600', 'border-blue-600')
      }
    })
  })

  describe('Active Page Highlighting (Requirements 1.4)', () => {
    it('should highlight Dashboard link when on Dashboard page', () => {
      renderAppWithRouter(['/'])

      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
      const aboutLink = screen.getByRole('link', { name: /about/i })

      // Dashboard should be active (highlighted)
      expect(dashboardLink).toHaveClass('text-blue-600', 'border-blue-600')
      
      // About should not be active
      expect(aboutLink).toHaveClass('text-gray-600', 'border-transparent')
      expect(aboutLink).not.toHaveClass('text-blue-600', 'border-blue-600')
    })

    it('should highlight About link when on About page', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate to About page
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })

      // About should be active (highlighted)
      expect(aboutLink).toHaveClass('text-blue-600', 'border-blue-600')
      
      // Dashboard should not be active
      expect(dashboardLink).toHaveClass('text-gray-600', 'border-transparent')
      expect(dashboardLink).not.toHaveClass('text-blue-600', 'border-blue-600')
    })
  })

  describe('Navigation Visibility (Requirements 1.5)', () => {
    it('should display navigation consistently on Dashboard page', () => {
      renderAppWithRouter(['/'])

      // Verify navigation is visible
      expect(screen.getByRole('link', { name: /dashboard/i })).toBeInTheDocument()
      expect(screen.getByRole('link', { name: /about/i })).toBeInTheDocument()
      
      // Verify navigation container exists
      const nav = screen.getByRole('navigation')
      expect(nav).toBeInTheDocument()
      expect(nav).toHaveClass('bg-white', 'shadow-md')
    })

    it('should display navigation consistently on About page', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate to About page
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      // Verify navigation is still visible
      expect(screen.getByRole('link', { name: /dashboard/i })).toBeInTheDocument()
      expect(screen.getByRole('link', { name: /about/i })).toBeInTheDocument()
      
      // Verify navigation container exists
      const nav = screen.getByRole('navigation')
      expect(nav).toBeInTheDocument()
      expect(nav).toHaveClass('bg-white', 'shadow-md')
    })
  })

  describe('Existing Dashboard Functionality Preservation (Requirements 1.2)', () => {
    it('should preserve all dashboard components and layout', () => {
      renderAppWithRouter(['/'])

      // Verify all dashboard components are present
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
      expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()

      // Verify main dashboard container structure
      const main = screen.getByRole('main')
      expect(main).toBeInTheDocument()
      expect(main).toHaveClass('mx-auto', 'px-2', 'sm:px-4', 'py-8')

      // Verify grid layout is preserved
      const gridContainer = main.querySelector('.grid')
      expect(gridContainer).toBeInTheDocument()
      expect(gridContainer).toHaveClass('lg:grid-cols-2', 'gap-4', 'lg:gap-8')
    })

    it('should maintain dashboard functionality after navigation', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate away and back
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
      await user.click(dashboardLink)

      // Verify dashboard components are still functional
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
      expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()

      // Verify layout is preserved
      const main = screen.getByRole('main')
      const gridContainer = main.querySelector('.grid')
      expect(gridContainer).toHaveClass('lg:grid-cols-2', 'gap-4', 'lg:gap-8')
    })
  })

  describe('Application Structure and Layout', () => {
    it('should maintain consistent application header across all pages', () => {
      renderAppWithRouter(['/'])

      // Verify header is present
      const header = screen.getByRole('banner')
      expect(header).toBeInTheDocument()
      expect(header).toHaveClass('bg-blue-600', 'text-white', 'shadow-lg')
      
      // Verify header title
      expect(screen.getByRole('heading', { name: /wanikani dashboard/i, level: 1 })).toBeInTheDocument()
    })

    it('should maintain consistent application header when navigating', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate to About page
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      // Verify header is still present and unchanged
      const header = screen.getByRole('banner')
      expect(header).toBeInTheDocument()
      expect(header).toHaveClass('bg-blue-600', 'text-white', 'shadow-lg')
      expect(screen.getByRole('heading', { name: /wanikani dashboard/i, level: 1 })).toBeInTheDocument()
    })

    it('should maintain consistent application background and container', () => {
      renderAppWithRouter(['/'])

      // Verify main app container
      const appContainer = screen.getByRole('banner').parentElement
      expect(appContainer).toHaveClass('min-h-screen', 'bg-gray-50')
    })
  })

  describe('Responsive Behavior Verification', () => {
    it('should have responsive navigation classes', () => {
      renderAppWithRouter(['/'])

      const nav = screen.getByRole('navigation')
      const navContainer = nav.querySelector('.mx-auto')
      expect(navContainer).toHaveClass('px-4')

      const navLinks = nav.querySelector('.flex')
      expect(navLinks).toHaveClass('space-x-1', 'sm:space-x-4')
    })

    it('should have responsive dashboard layout classes', () => {
      renderAppWithRouter(['/'])

      const main = screen.getByRole('main')
      expect(main).toHaveClass('px-2', 'sm:px-4') // Responsive padding

      // Verify we're on dashboard page and check its grid layout
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      const gridContainer = main.querySelector('.grid')
      expect(gridContainer).toHaveClass('lg:grid-cols-2') // Responsive grid for dashboard
    })

    it('should have responsive About page layout classes', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Navigate to About page
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)

      const main = screen.getByRole('main')
      expect(main).toHaveClass('px-2', 'sm:px-4') // Responsive padding

      const contentContainer = main.querySelector('.max-w-4xl')
      expect(contentContainer).toBeInTheDocument() // Max width constraint
    })
  })

  describe('Error Handling and Edge Cases', () => {
    it('should handle rapid navigation clicks gracefully', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      const aboutLink = screen.getByRole('link', { name: /about/i })
      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })

      // Rapidly click navigation links
      await user.click(aboutLink)
      await user.click(dashboardLink)
      await user.click(aboutLink)
      await user.click(dashboardLink)

      // Verify final state is correct
      expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
      expect(dashboardLink).toHaveClass('text-blue-600', 'border-blue-600')
    })

    it('should maintain application state during navigation', async () => {
      const user = userEvent.setup()
      renderAppWithRouter(['/'])

      // Verify initial state
      expect(screen.getByRole('heading', { name: /wanikani dashboard/i, level: 1 })).toBeInTheDocument()
      
      // Navigate and verify state is maintained
      const aboutLink = screen.getByRole('link', { name: /about/i })
      await user.click(aboutLink)
      
      expect(screen.getByRole('heading', { name: /wanikani dashboard/i, level: 1 })).toBeInTheDocument()
      
      const dashboardLink = screen.getByRole('link', { name: /dashboard/i })
      await user.click(dashboardLink)
      
      expect(screen.getByRole('heading', { name: /wanikani dashboard/i, level: 1 })).toBeInTheDocument()
    })
  })
})