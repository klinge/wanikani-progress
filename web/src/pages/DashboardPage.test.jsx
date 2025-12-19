import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DashboardPage from './DashboardPage'

// Mock the child components since we're testing DashboardPage in isolation
vi.mock('../components/ItemSpreadCard.jsx', () => ({
  default: vi.fn(() => <div data-testid="item-spread-card">ItemSpreadCard</div>)
}))

vi.mock('../components/DailyTotalsChart.jsx', () => ({
  default: vi.fn(() => <div data-testid="daily-totals-chart">DailyTotalsChart</div>)
}))

vi.mock('../components/DailyProportionsChart.jsx', () => ({
  default: vi.fn(() => <div data-testid="daily-proportions-chart">DailyProportionsChart</div>)
}))

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders correctly with all dashboard components', () => {
    render(<DashboardPage />)
    
    // Check that all three main components are rendered
    expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
    expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
    expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()
  })

  it('has correct layout structure', () => {
    render(<DashboardPage />)
    
    // Check that the main container exists
    const main = screen.getByRole('main')
    expect(main).toBeInTheDocument()
    expect(main).toHaveClass('mx-auto', 'px-2', 'sm:px-4', 'py-8')
    
    // Check that the grid container exists
    const gridContainer = main.querySelector('.grid')
    expect(gridContainer).toBeInTheDocument()
    expect(gridContainer).toHaveClass('lg:grid-cols-2', 'gap-4', 'lg:gap-8')
  })

  it('renders components in correct grid layout', () => {
    render(<DashboardPage />)
    
    const itemSpread = screen.getByTestId('item-spread-card')
    const dailyTotals = screen.getByTestId('daily-totals-chart')
    const dailyProportions = screen.getByTestId('daily-proportions-chart')
    
    // Check that components are wrapped in divs with correct classes
    expect(itemSpread.parentElement).toHaveClass('w-full', 'h-80', 'lg:h-96')
    expect(dailyTotals.parentElement).toHaveClass('w-full', 'h-80', 'lg:h-96')
    expect(dailyProportions.parentElement).toHaveClass('w-full', 'h-80', 'lg:h-120', 'lg:col-span-2')
  })

  it('renders dashboard with data when child components load successfully', () => {
    render(<DashboardPage />)
    
    // Verify all components are rendered (they handle their own data loading)
    expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
    expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
    expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()
    
    // Verify the dashboard layout is maintained
    expect(screen.getByRole('main')).toBeInTheDocument()
  })

  it('maintains layout structure regardless of child component states', () => {
    render(<DashboardPage />)
    
    // Verify dashboard maintains its structure even when child components 
    // might be in loading or error states (handled internally by each component)
    expect(screen.getByRole('main')).toBeInTheDocument()
    expect(screen.getByTestId('item-spread-card')).toBeInTheDocument()
    expect(screen.getByTestId('daily-totals-chart')).toBeInTheDocument()
    expect(screen.getByTestId('daily-proportions-chart')).toBeInTheDocument()
    
    // Verify grid layout is preserved
    const gridContainer = screen.getByRole('main').querySelector('.grid')
    expect(gridContainer).toBeInTheDocument()
  })
})