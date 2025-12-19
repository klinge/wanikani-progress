import { render, screen, cleanup } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, afterEach } from 'vitest'
import fc from 'fast-check'
import Navigation from './Navigation.jsx'

// Helper function to render Navigation with router context
function renderNavigationWithRouter(initialEntries = ['/']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Navigation />
    </MemoryRouter>
  )
}

// Ensure cleanup after each test to prevent multiple components
afterEach(() => {
  cleanup()
})

describe('Navigation Property Tests', () => {
  /**
   * Feature: about-page, Property 1: Navigation link functionality
   * Validates: Requirements 1.2, 1.3, 3.5
   * 
   * Property: For any navigation link in the application, clicking it should 
   * navigate to the correct page and update the URL accordingly
   */
  it('should navigate correctly for any valid navigation path', async () => {
    await fc.assert(
      fc.asyncProperty(
        // Generate navigation paths from the available routes
        fc.constantFrom('/', '/about'),
        async (targetPath) => {
          // Clean up any previous renders
          cleanup()
          
          const user = userEvent.setup()
          
          // Start from a different path to ensure navigation actually occurs
          const startPath = targetPath === '/' ? '/about' : '/'
          
          renderNavigationWithRouter([startPath])
          
          // Find the navigation link for the target path
          const targetLabel = targetPath === '/' ? 'Dashboard' : 'About'
          const navigationLink = screen.getByRole('link', { name: new RegExp(targetLabel, 'i') })
          
          // Verify the link exists and has correct href
          expect(navigationLink).toBeInTheDocument()
          expect(navigationLink).toHaveAttribute('href', targetPath)
          
          // Click the navigation link
          await user.click(navigationLink)
          
          // Verify the link becomes active (has active styling)
          expect(navigationLink).toHaveClass('text-blue-600', 'border-blue-600')
          
          // Verify other links are not active
          const allLinks = screen.getAllByRole('link')
          const otherLinks = allLinks.filter(link => link !== navigationLink)
          
          otherLinks.forEach(link => {
            expect(link).toHaveClass('text-gray-600', 'border-transparent')
            expect(link).not.toHaveClass('text-blue-600', 'border-blue-600')
          })
        }
      ),
      { numRuns: 50 } // Reduced runs for performance
    )
  }, 10000) // Increased timeout for property-based test

  /**
   * Additional property test to verify navigation state consistency
   * Tests that navigation links maintain correct active states
   */
  it('should maintain correct active state for any navigation sequence', async () => {
    await fc.assert(
      fc.asyncProperty(
        // Generate a single navigation target to test active state
        fc.constantFrom('/', '/about'),
        async (targetPath) => {
          // Clean up any previous renders
          cleanup()
          
          const user = userEvent.setup()
          
          // Start from the opposite path to ensure we can see the state change
          const startPath = targetPath === '/' ? '/about' : '/'
          renderNavigationWithRouter([startPath])
          
          // Navigate to the target path
          const targetLabel = targetPath === '/' ? 'Dashboard' : 'About'
          const navigationLink = screen.getByRole('link', { name: new RegExp(targetLabel, 'i') })
          
          await user.click(navigationLink)
          
          // Verify only the clicked link is active
          const allLinks = screen.getAllByRole('link')
          
          allLinks.forEach(link => {
            const linkHref = link.getAttribute('href')
            if (linkHref === targetPath) {
              // This should be the active link
              expect(link).toHaveClass('text-blue-600', 'border-blue-600')
            } else {
              // This should not be active
              expect(link).toHaveClass('text-gray-600', 'border-transparent')
              expect(link).not.toHaveClass('text-blue-600', 'border-blue-600')
            }
          })
        }
      ),
      { numRuns: 50 } // Reduced runs for performance
    )
  }, 10000) // Increased timeout for property-based test
})