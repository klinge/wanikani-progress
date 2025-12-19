import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import AboutPage from './AboutPage'

describe('AboutPage', () => {
  describe('Required Content Sections', () => {
    it('displays the application title and description', () => {
      render(<AboutPage />)
      
      // Test application title (Requirement 2.1)
      expect(screen.getByRole('heading', { name: /about wanikani dashboard/i })).toBeInTheDocument()
      
      // Test application description (Requirement 2.1)
      expect(screen.getByText(/comprehensive progress tracking dashboard for wanikani learners/i)).toBeInTheDocument()
    })

    it('explains what WaniKani is and how the dashboard helps users', () => {
      render(<AboutPage />)
      
      // Test WaniKani explanation (Requirement 2.2)
      expect(screen.getByRole('heading', { name: /what is wanikani/i })).toBeInTheDocument()
      expect(screen.getByText(/wanikani is a spaced repetition system/i)).toBeInTheDocument()
      expect(screen.getByText(/this dashboard helps you track your progress/i)).toBeInTheDocument()
      
      // Test external link to WaniKani
      const wanikaniLink = screen.getByRole('link', { name: /visit wanikani/i })
      expect(wanikaniLink).toBeInTheDocument()
      expect(wanikaniLink).toHaveAttribute('href', 'https://www.wanikani.com')
      expect(wanikaniLink).toHaveAttribute('target', '_blank')
    })

    it('describes the main dashboard features', () => {
      render(<AboutPage />)
      
      // Test dashboard features section (Requirement 2.3)
      expect(screen.getByRole('heading', { name: /dashboard features/i })).toBeInTheDocument()
      
      // Test individual feature descriptions
      expect(screen.getByText(/item spread overview/i)).toBeInTheDocument()
      expect(screen.getByText(/daily totals chart/i)).toBeInTheDocument()
      expect(screen.getByText(/daily proportions chart/i)).toBeInTheDocument()
      expect(screen.getByText(/progress visualization/i)).toBeInTheDocument()
      
      // Test feature descriptions contain helpful information
      expect(screen.getByText(/distribution of your learned items across different srs stages/i)).toBeInTheDocument()
      expect(screen.getByText(/track your daily review activity/i)).toBeInTheDocument()
      expect(screen.getByText(/percentage breakdown of your reviews by srs stage/i)).toBeInTheDocument()
    })

    it('provides information about chart interpretation', () => {
      render(<AboutPage />)
      
      // Test chart interpretation guide (Requirement 2.4)
      expect(screen.getByRole('heading', { name: /understanding your charts/i })).toBeInTheDocument()
      
      // Test SRS stages explanation
      expect(screen.getByText(/srs stages explained/i)).toBeInTheDocument()
      expect(screen.getByText(/apprentice:/i)).toBeInTheDocument()
      expect(screen.getByText(/guru:/i)).toBeInTheDocument()
      expect(screen.getByText(/master:/i)).toBeInTheDocument()
      expect(screen.getByText(/enlightened:/i)).toBeInTheDocument()
      expect(screen.getByText(/burned:/i)).toBeInTheDocument()
      
      // Test progress reading tips
      expect(screen.getByText(/reading your progress/i)).toBeInTheDocument()
      expect(screen.getByText(/high apprentice numbers/i)).toBeInTheDocument()
      expect(screen.getByText(/growing guru items/i)).toBeInTheDocument()
      
      // Test success tips
      expect(screen.getByText(/tips for success/i)).toBeInTheDocument()
      expect(screen.getByText(/aim for consistent daily reviews/i)).toBeInTheDocument()
    })

    it('maintains consistent styling with the application', () => {
      render(<AboutPage />)
      
      // Test main container has consistent styling (Requirement 2.5)
      const main = screen.getByRole('main')
      expect(main).toHaveClass('mx-auto', 'px-2', 'sm:px-4', 'py-8')
      
      // Test content container
      const contentContainer = main.querySelector('.max-w-4xl')
      expect(contentContainer).toBeInTheDocument()
      expect(contentContainer).toHaveClass('mx-auto', 'space-y-8')
      
      // Test section styling consistency
      const sections = main.querySelectorAll('section.bg-white')
      expect(sections.length).toBeGreaterThan(0)
      sections.forEach(section => {
        expect(section).toHaveClass('rounded-lg', 'shadow-md', 'p-6')
      })
    })
  })

  describe('Responsive Layout Behavior', () => {
    it('has responsive container classes', () => {
      render(<AboutPage />)
      
      const main = screen.getByRole('main')
      expect(main).toHaveClass('px-2', 'sm:px-4') // Responsive padding
      
      const contentContainer = main.querySelector('.max-w-4xl')
      expect(contentContainer).toBeInTheDocument() // Max width constraint
    })

    it('has responsive grid layout for features section', () => {
      render(<AboutPage />)
      
      // Test responsive grid in dashboard features section
      const featuresGrid = screen.getByRole('heading', { name: /dashboard features/i })
        .closest('section')
        .querySelector('.grid')
      
      expect(featuresGrid).toBeInTheDocument()
      expect(featuresGrid).toHaveClass('md:grid-cols-2', 'gap-6')
    })

    it('has responsive text sizing and spacing', () => {
      render(<AboutPage />)
      
      // Test responsive title
      const title = screen.getByRole('heading', { name: /about wanikani dashboard/i })
      expect(title).toHaveClass('text-3xl')
      
      // Test responsive description text
      const description = screen.getByText(/comprehensive progress tracking dashboard/i)
      expect(description).toHaveClass('text-lg', 'max-w-2xl', 'mx-auto')
    })

    it('has responsive SRS stages layout', () => {
      render(<AboutPage />)
      
      // Test responsive grid for SRS stages
      const srsStagesSection = screen.getByText(/srs stages explained/i).closest('.border-l-4')
      const srsGrid = srsStagesSection.querySelector('.grid')
      
      expect(srsGrid).toBeInTheDocument()
      expect(srsGrid).toHaveClass('sm:grid-cols-2', 'gap-3')
    })

    it('maintains readability across different screen sizes', () => {
      render(<AboutPage />)
      
      // Test text elements have appropriate sizing for readability
      const bodyText = screen.getByText(/wanikani is a spaced repetition system/i)
      expect(bodyText).toHaveClass('text-gray-700')
      
      // Test small text elements are appropriately sized
      const smallText = screen.getByText(/this dashboard is designed to complement/i)
      const footerSection = smallText.closest('section')
      expect(footerSection).toHaveClass('text-sm')
      
      // Test list items have proper spacing
      const listItems = screen.getByText(/high apprentice numbers/i).closest('ul')
      expect(listItems).toHaveClass('space-y-1')
    })
  })

  describe('Content Structure and Accessibility', () => {
    it('has proper heading hierarchy', () => {
      render(<AboutPage />)
      
      // Test main heading (h2)
      expect(screen.getByRole('heading', { level: 2, name: /about wanikani dashboard/i })).toBeInTheDocument()
      
      // Test section headings (h3)
      expect(screen.getByRole('heading', { level: 3, name: /what is wanikani/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { level: 3, name: /dashboard features/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { level: 3, name: /understanding your charts/i })).toBeInTheDocument()
      
      // Test subsection headings (h4)
      expect(screen.getByRole('heading', { level: 4, name: /item spread overview/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { level: 4, name: /srs stages explained/i })).toBeInTheDocument()
    })

    it('has proper semantic structure with main and sections', () => {
      render(<AboutPage />)
      
      // Test main landmark
      expect(screen.getByRole('main')).toBeInTheDocument()
      
      // Test sections are properly structured
      const main = screen.getByRole('main')
      const sections = main.querySelectorAll('section')
      expect(sections.length).toBeGreaterThan(3) // Multiple content sections
    })

    it('has accessible external links', () => {
      render(<AboutPage />)
      
      const wanikaniLink = screen.getByRole('link', { name: /visit wanikani/i })
      expect(wanikaniLink).toHaveAttribute('rel', 'noopener noreferrer')
      expect(wanikaniLink).toHaveAttribute('target', '_blank')
    })
  })
})