package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"wanikani-api/internal/domain"
)

// Service implements the SyncService interface
type Service struct {
	client  domain.WaniKaniClient
	store   domain.DataStore
	logger  *logrus.Logger
	mu      sync.Mutex
	syncing bool
}

// NewService creates a new sync service
func NewService(client domain.WaniKaniClient, store domain.DataStore, logger *logrus.Logger) *Service {
	return &Service{
		client:  client,
		store:   store,
		logger:  logger,
		syncing: false,
	}
}

// IsSyncing returns true if a sync operation is currently in progress
func (s *Service) IsSyncing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncing
}

// setSyncing sets the syncing flag
func (s *Service) setSyncing(syncing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncing = syncing
}

// SyncAll performs a full sync of all data types in the correct order
func (s *Service) SyncAll(ctx context.Context) ([]domain.SyncResult, error) {
	// Prevent concurrent syncs
	if s.IsSyncing() {
		s.logger.Warn("Sync already in progress, rejecting concurrent sync request")
		return nil, fmt.Errorf("sync already in progress")
	}

	s.logger.Info("Starting full sync operation")
	s.setSyncing(true)
	defer s.setSyncing(false)

	var results []domain.SyncResult

	// Sync in order: subjects → assignments → reviews → statistics
	// This maintains referential integrity

	// 1. Sync subjects
	s.logger.Info("Syncing subjects...")
	subjectsResult := s.SyncSubjects(ctx)
	results = append(results, subjectsResult)
	if !subjectsResult.Success {
		s.logger.WithFields(logrus.Fields{
			"data_type": subjectsResult.DataType,
			"error":     subjectsResult.Error,
		}).Error("Subjects sync failed")
		return results, fmt.Errorf("subjects sync failed: %s", subjectsResult.Error)
	}
	s.logger.WithField("records_updated", subjectsResult.RecordsUpdated).Info("Subjects sync completed successfully")

	// 2. Sync assignments
	s.logger.Info("Syncing assignments...")
	assignmentsResult := s.SyncAssignments(ctx)
	results = append(results, assignmentsResult)
	if !assignmentsResult.Success {
		s.logger.WithFields(logrus.Fields{
			"data_type": assignmentsResult.DataType,
			"error":     assignmentsResult.Error,
		}).Error("Assignments sync failed")
		return results, fmt.Errorf("assignments sync failed: %s", assignmentsResult.Error)
	}
	s.logger.WithField("records_updated", assignmentsResult.RecordsUpdated).Info("Assignments sync completed successfully")

	// 3. Sync reviews
	s.logger.Info("Syncing reviews...")
	reviewsResult := s.SyncReviews(ctx)
	results = append(results, reviewsResult)
	if !reviewsResult.Success {
		s.logger.WithFields(logrus.Fields{
			"data_type": reviewsResult.DataType,
			"error":     reviewsResult.Error,
		}).Error("Reviews sync failed")
		return results, fmt.Errorf("reviews sync failed: %s", reviewsResult.Error)
	}
	s.logger.WithField("records_updated", reviewsResult.RecordsUpdated).Info("Reviews sync completed successfully")

	// 4. Sync statistics
	s.logger.Info("Syncing statistics...")
	statisticsResult := s.SyncStatistics(ctx)
	results = append(results, statisticsResult)
	if !statisticsResult.Success {
		s.logger.WithFields(logrus.Fields{
			"data_type": statisticsResult.DataType,
			"error":     statisticsResult.Error,
		}).Error("Statistics sync failed")
		return results, fmt.Errorf("statistics sync failed: %s", statisticsResult.Error)
	}
	s.logger.WithField("records_updated", statisticsResult.RecordsUpdated).Info("Statistics sync completed successfully")

	s.logger.WithField("total_results", len(results)).Info("Full sync operation completed successfully")
	return results, nil
}

// SyncSubjects syncs only subjects
func (s *Service) SyncSubjects(ctx context.Context) domain.SyncResult {
	result := domain.SyncResult{
		DataType:  domain.DataTypeSubjects,
		Timestamp: time.Now(),
		Success:   false,
	}

	// Get last sync time for incremental updates
	lastSyncTime, err := s.store.GetLastSyncTime(ctx, domain.DataTypeSubjects)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get last sync time: %v", err)
		s.logger.WithError(err).Error("Failed to get last sync time for subjects")
		return result
	}

	if lastSyncTime != nil {
		s.logger.WithField("updated_after", lastSyncTime.Format(time.RFC3339)).Debug("Performing incremental sync for subjects")
	} else {
		s.logger.Debug("Performing full sync for subjects (no previous sync time)")
	}

	// Fetch subjects from API
	subjects, err := s.client.FetchSubjects(ctx, lastSyncTime)
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch subjects: %v", err)
		s.logger.WithError(err).Error("Failed to fetch subjects from API")
		return result
	}

	s.logger.WithField("count", len(subjects)).Debug("Fetched subjects from API")

	// Store subjects
	if len(subjects) > 0 {
		if err := s.store.UpsertSubjects(ctx, subjects); err != nil {
			result.Error = fmt.Sprintf("failed to store subjects: %v", err)
			s.logger.WithError(err).Error("Failed to store subjects in database")
			return result
		}
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(ctx, domain.DataTypeSubjects, result.Timestamp); err != nil {
		result.Error = fmt.Sprintf("failed to update sync time: %v", err)
		s.logger.WithError(err).Error("Failed to update last sync time for subjects")
		return result
	}

	result.RecordsUpdated = len(subjects)
	result.Success = true
	return result
}

// SyncAssignments syncs only assignments
func (s *Service) SyncAssignments(ctx context.Context) domain.SyncResult {
	result := domain.SyncResult{
		DataType:  domain.DataTypeAssignments,
		Timestamp: time.Now(),
		Success:   false,
	}

	// Get last sync time for incremental updates
	lastSyncTime, err := s.store.GetLastSyncTime(ctx, domain.DataTypeAssignments)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get last sync time: %v", err)
		s.logger.WithError(err).Error("Failed to get last sync time for assignments")
		return result
	}

	if lastSyncTime != nil {
		s.logger.WithField("updated_after", lastSyncTime.Format(time.RFC3339)).Debug("Performing incremental sync for assignments")
	} else {
		s.logger.Debug("Performing full sync for assignments (no previous sync time)")
	}

	// Fetch assignments from API
	assignments, err := s.client.FetchAssignments(ctx, lastSyncTime)
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch assignments: %v", err)
		s.logger.WithError(err).Error("Failed to fetch assignments from API")
		return result
	}

	s.logger.WithField("count", len(assignments)).Debug("Fetched assignments from API")

	// Store assignments
	if len(assignments) > 0 {
		if err := s.store.UpsertAssignments(ctx, assignments); err != nil {
			result.Error = fmt.Sprintf("failed to store assignments: %v", err)
			s.logger.WithError(err).Error("Failed to store assignments in database")
			return result
		}
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(ctx, domain.DataTypeAssignments, result.Timestamp); err != nil {
		result.Error = fmt.Sprintf("failed to update sync time: %v", err)
		s.logger.WithError(err).Error("Failed to update last sync time for assignments")
		return result
	}

	result.RecordsUpdated = len(assignments)
	result.Success = true
	return result
}

// SyncReviews syncs only reviews
func (s *Service) SyncReviews(ctx context.Context) domain.SyncResult {
	result := domain.SyncResult{
		DataType:  domain.DataTypeReviews,
		Timestamp: time.Now(),
		Success:   false,
	}

	// Get last sync time for incremental updates
	lastSyncTime, err := s.store.GetLastSyncTime(ctx, domain.DataTypeReviews)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get last sync time: %v", err)
		s.logger.WithError(err).Error("Failed to get last sync time for reviews")
		return result
	}

	if lastSyncTime != nil {
		s.logger.WithField("updated_after", lastSyncTime.Format(time.RFC3339)).Debug("Performing incremental sync for reviews")
	} else {
		s.logger.Debug("Performing full sync for reviews (no previous sync time)")
	}

	// Fetch reviews from API
	reviews, err := s.client.FetchReviews(ctx, lastSyncTime)
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch reviews: %v", err)
		s.logger.WithError(err).Error("Failed to fetch reviews from API")
		return result
	}

	s.logger.WithField("count", len(reviews)).Debug("Fetched reviews from API")

	// Store reviews
	if len(reviews) > 0 {
		if err := s.store.UpsertReviews(ctx, reviews); err != nil {
			result.Error = fmt.Sprintf("failed to store reviews: %v", err)
			s.logger.WithError(err).Error("Failed to store reviews in database")
			return result
		}
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(ctx, domain.DataTypeReviews, result.Timestamp); err != nil {
		result.Error = fmt.Sprintf("failed to update sync time: %v", err)
		s.logger.WithError(err).Error("Failed to update last sync time for reviews")
		return result
	}

	result.RecordsUpdated = len(reviews)
	result.Success = true
	return result
}

// SyncStatistics syncs only statistics
func (s *Service) SyncStatistics(ctx context.Context) domain.SyncResult {
	result := domain.SyncResult{
		DataType:  domain.DataTypeStatistics,
		Timestamp: time.Now(),
		Success:   false,
	}

	s.logger.Debug("Fetching statistics snapshot from API")

	// Fetch statistics from API
	statistics, err := s.client.FetchStatistics(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to fetch statistics: %v", err)
		s.logger.WithError(err).Error("Failed to fetch statistics from API")
		return result
	}

	// Store statistics snapshot
	if statistics != nil {
		if err := s.store.InsertStatistics(ctx, *statistics, result.Timestamp); err != nil {
			result.Error = fmt.Sprintf("failed to store statistics: %v", err)
			s.logger.WithError(err).Error("Failed to store statistics in database")
			return result
		}
		s.logger.Debug("Statistics snapshot stored successfully")
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(ctx, domain.DataTypeStatistics, result.Timestamp); err != nil {
		result.Error = fmt.Sprintf("failed to update sync time: %v", err)
		s.logger.WithError(err).Error("Failed to update last sync time for statistics")
		return result
	}

	result.RecordsUpdated = 1
	result.Success = true
	return result
}
