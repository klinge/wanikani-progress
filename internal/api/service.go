package api

import (
	"context"
	"fmt"

	"wanikani-api/internal/domain"
)

// Service contains the business logic for the API
type Service struct {
	store       domain.DataStore
	syncService domain.SyncService
}

// NewService creates a new API service
func NewService(store domain.DataStore, syncService domain.SyncService) *Service {
	return &Service{
		store:       store,
		syncService: syncService,
	}
}

// GetSubjects retrieves subjects with optional filters
func (s *Service) GetSubjects(ctx context.Context, filters domain.SubjectFilters) ([]domain.Subject, error) {
	return s.store.GetSubjects(ctx, filters)
}

// AssignmentWithSubject represents an assignment with its associated subject
type AssignmentWithSubject struct {
	domain.Assignment
	Subject *domain.Subject `json:"subject"`
}

// GetAssignmentsWithSubjects retrieves assignments and joins them with their subjects
func (s *Service) GetAssignmentsWithSubjects(ctx context.Context, filters domain.AssignmentFilters) ([]AssignmentWithSubject, error) {
	// Fetch assignments
	assignments, err := s.store.GetAssignments(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments: %w", err)
	}

	// Fetch all subjects once
	subjects, err := s.store.GetSubjects(ctx, domain.SubjectFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve subjects: %w", err)
	}

	// Create a map for quick lookup
	subjectMap := make(map[int]*domain.Subject)
	for i := range subjects {
		subjectMap[subjects[i].ID] = &subjects[i]
	}

	// Join with subjects
	result := make([]AssignmentWithSubject, 0, len(assignments))
	for _, assignment := range assignments {
		result = append(result, AssignmentWithSubject{
			Assignment: assignment,
			Subject:    subjectMap[assignment.Data.SubjectID],
		})
	}

	return result, nil
}

// ReviewWithDetails represents a review with its associated assignment and subject
type ReviewWithDetails struct {
	domain.Review
	Assignment *domain.Assignment `json:"assignment"`
	Subject    *domain.Subject    `json:"subject"`
}

// GetReviewsWithDetails retrieves reviews and joins them with assignments and subjects
func (s *Service) GetReviewsWithDetails(ctx context.Context, filters domain.ReviewFilters) ([]ReviewWithDetails, error) {
	// Fetch reviews
	reviews, err := s.store.GetReviews(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve reviews: %w", err)
	}

	// Fetch all assignments and subjects once
	assignments, err := s.store.GetAssignments(ctx, domain.AssignmentFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments: %w", err)
	}

	subjects, err := s.store.GetSubjects(ctx, domain.SubjectFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve subjects: %w", err)
	}

	// Create maps for quick lookup
	assignmentMap := make(map[int]*domain.Assignment)
	for i := range assignments {
		assignmentMap[assignments[i].ID] = &assignments[i]
	}

	subjectMap := make(map[int]*domain.Subject)
	for i := range subjects {
		subjectMap[subjects[i].ID] = &subjects[i]
	}

	// Join with assignments and subjects
	result := make([]ReviewWithDetails, 0, len(reviews))
	for _, review := range reviews {
		result = append(result, ReviewWithDetails{
			Review:     review,
			Assignment: assignmentMap[review.Data.AssignmentID],
			Subject:    subjectMap[review.Data.SubjectID],
		})
	}

	return result, nil
}

// GetLatestStatistics retrieves the most recent statistics snapshot
func (s *Service) GetLatestStatistics(ctx context.Context) (*domain.StatisticsSnapshot, error) {
	return s.store.GetLatestStatistics(ctx)
}

// GetStatistics retrieves statistics snapshots within a date range
func (s *Service) GetStatistics(ctx context.Context, dateRange *domain.DateRange) ([]domain.StatisticsSnapshot, error) {
	return s.store.GetStatistics(ctx, dateRange)
}

// TriggerSync triggers a manual sync operation
func (s *Service) TriggerSync(ctx context.Context) ([]domain.SyncResult, error) {
	// Check if sync is already in progress
	if s.syncService.IsSyncing() {
		return nil, fmt.Errorf("sync already in progress")
	}

	return s.syncService.SyncAll(ctx)
}

// GetSyncStatus returns whether a sync is currently in progress
func (s *Service) GetSyncStatus() bool {
	return s.syncService.IsSyncing()
}

// GetAssignmentSnapshots retrieves assignment snapshots and transforms them into nested structure
func (s *Service) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) (map[string]map[string]map[string]int, error) {
	// Fetch snapshots from store
	snapshots, err := s.store.GetAssignmentSnapshots(ctx, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignment snapshots: %w", err)
	}

	// Transform flat snapshot records into nested structure grouped by date and SRS stage name
	// Structure: date -> SRS stage name -> subject type -> count
	result := make(map[string]map[string]map[string]int)

	for _, snapshot := range snapshots {
		dateStr := snapshot.Date.Format("2006-01-02")
		stageName := domain.GetSRSStageName(snapshot.SRSStage)

		// Initialize nested maps if they don't exist
		if result[dateStr] == nil {
			result[dateStr] = make(map[string]map[string]int)
		}
		if result[dateStr][stageName] == nil {
			result[dateStr][stageName] = make(map[string]int)
		}

		// Add count for this subject type (sum across multiple SRS stages that map to same name)
		result[dateStr][stageName][snapshot.SubjectType] += snapshot.Count
	}

	// Calculate and include totals for each SRS stage
	for date := range result {
		for stageName := range result[date] {
			total := 0
			for _, count := range result[date][stageName] {
				total += count
			}
			result[date][stageName]["total"] = total
		}
	}

	return result, nil
}
