package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"wanikani-api/internal/domain"
)

// Store implements the DataStore interface using SQLite
type Store struct {
	db *sql.DB
}

// New creates a new SQLite store
// Note: Migrations should be run separately before creating the store
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &Store{db: db}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// BeginTx starts a new database transaction
func (s *Store) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, nil)
}

// UpsertSubjects inserts or updates subjects
func (s *Store) UpsertSubjects(ctx context.Context, subjects []domain.Subject) error {
	if len(subjects) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO subjects (id, object, url, data_updated_at, data)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			object = excluded.object,
			url = excluded.url,
			data_updated_at = excluded.data_updated_at,
			data = excluded.data
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, subject := range subjects {
		dataJSON, err := json.Marshal(subject.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal subject data: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			subject.ID,
			subject.Object,
			subject.URL,
			subject.DataUpdatedAt.Format(time.RFC3339),
			string(dataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert subject: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSubjects retrieves subjects matching the provided filters
func (s *Store) GetSubjects(ctx context.Context, filters domain.SubjectFilters) ([]domain.Subject, error) {
	query := `SELECT id, object, url, data_updated_at, data FROM subjects WHERE 1=1`
	args := []interface{}{}

	if filters.Type != "" {
		query += ` AND object = ?`
		args = append(args, filters.Type)
	}

	if filters.Level != nil {
		query += ` AND json_extract(data, '$.level') = ?`
		args = append(args, *filters.Level)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query subjects: %w", err)
	}
	defer rows.Close()

	var subjects []domain.Subject
	for rows.Next() {
		var subject domain.Subject
		var dataUpdatedAtStr string
		var dataJSON string

		err := rows.Scan(
			&subject.ID,
			&subject.Object,
			&subject.URL,
			&dataUpdatedAtStr,
			&dataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subject: %w", err)
		}

		subject.DataUpdatedAt, err = time.Parse(time.RFC3339, dataUpdatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data_updated_at: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &subject.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal subject data: %w", err)
		}

		subjects = append(subjects, subject)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subjects: %w", err)
	}

	return subjects, nil
}

// UpsertAssignments inserts or updates assignments
func (s *Store) UpsertAssignments(ctx context.Context, assignments []domain.Assignment) error {
	if len(assignments) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Validate that all referenced subjects exist
	for _, assignment := range assignments {
		if err := s.validateSubjectExists(ctx, tx, assignment.Data.SubjectID); err != nil {
			return fmt.Errorf("assignment %d references invalid subject %d: %w", assignment.ID, assignment.Data.SubjectID, err)
		}
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO assignments (id, object, url, data_updated_at, subject_id, data)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			object = excluded.object,
			url = excluded.url,
			data_updated_at = excluded.data_updated_at,
			subject_id = excluded.subject_id,
			data = excluded.data
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, assignment := range assignments {
		dataJSON, err := json.Marshal(assignment.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal assignment data: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			assignment.ID,
			assignment.Object,
			assignment.URL,
			assignment.DataUpdatedAt.Format(time.RFC3339),
			assignment.Data.SubjectID,
			string(dataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert assignment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAssignments retrieves assignments matching the provided filters
func (s *Store) GetAssignments(ctx context.Context, filters domain.AssignmentFilters) ([]domain.Assignment, error) {
	query := `SELECT id, object, url, data_updated_at, subject_id, data FROM assignments WHERE 1=1`
	args := []interface{}{}

	if filters.SRSStage != nil {
		query += ` AND json_extract(data, '$.srs_stage') = ?`
		args = append(args, *filters.SRSStage)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []domain.Assignment
	for rows.Next() {
		var assignment domain.Assignment
		var dataUpdatedAtStr string
		var dataJSON string
		var subjectID int

		err := rows.Scan(
			&assignment.ID,
			&assignment.Object,
			&assignment.URL,
			&dataUpdatedAtStr,
			&subjectID,
			&dataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignment.DataUpdatedAt, err = time.Parse(time.RFC3339, dataUpdatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data_updated_at: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &assignment.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment data: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}

// UpsertReviews inserts or updates reviews
func (s *Store) UpsertReviews(ctx context.Context, reviews []domain.Review) error {
	if len(reviews) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Validate that all referenced assignments and subjects exist
	for _, review := range reviews {
		if err := s.validateAssignmentExists(ctx, tx, review.Data.AssignmentID); err != nil {
			return fmt.Errorf("review %d references invalid assignment %d: %w", review.ID, review.Data.AssignmentID, err)
		}
		if err := s.validateSubjectExists(ctx, tx, review.Data.SubjectID); err != nil {
			return fmt.Errorf("review %d references invalid subject %d: %w", review.ID, review.Data.SubjectID, err)
		}
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO reviews (id, object, url, data_updated_at, assignment_id, subject_id, data)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			object = excluded.object,
			url = excluded.url,
			data_updated_at = excluded.data_updated_at,
			assignment_id = excluded.assignment_id,
			subject_id = excluded.subject_id,
			data = excluded.data
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, review := range reviews {
		dataJSON, err := json.Marshal(review.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal review data: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			review.ID,
			review.Object,
			review.URL,
			review.DataUpdatedAt.Format(time.RFC3339),
			review.Data.AssignmentID,
			review.Data.SubjectID,
			string(dataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert review: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetReviews retrieves reviews matching the provided filters
func (s *Store) GetReviews(ctx context.Context, filters domain.ReviewFilters) ([]domain.Review, error) {
	query := `SELECT id, object, url, data_updated_at, assignment_id, subject_id, data FROM reviews WHERE 1=1`
	args := []interface{}{}

	if filters.From != nil {
		query += ` AND json_extract(data, '$.created_at') >= ?`
		args = append(args, filters.From.Format(time.RFC3339))
	}

	if filters.To != nil {
		query += ` AND json_extract(data, '$.created_at') <= ?`
		args = append(args, filters.To.Format(time.RFC3339))
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var review domain.Review
		var dataUpdatedAtStr string
		var dataJSON string
		var assignmentID, subjectID int

		err := rows.Scan(
			&review.ID,
			&review.Object,
			&review.URL,
			&dataUpdatedAtStr,
			&assignmentID,
			&subjectID,
			&dataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}

		review.DataUpdatedAt, err = time.Parse(time.RFC3339, dataUpdatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data_updated_at: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &review.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal review data: %w", err)
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviews: %w", err)
	}

	return reviews, nil
}

// InsertStatistics inserts a new statistics snapshot
func (s *Store) InsertStatistics(ctx context.Context, stats domain.Statistics, timestamp time.Time) error {
	dataJSON, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO statistics_snapshots (timestamp, data)
		VALUES (?, ?)
	`, timestamp.Format(time.RFC3339), string(dataJSON))

	if err != nil {
		return fmt.Errorf("failed to insert statistics: %w", err)
	}

	return nil
}

// GetStatistics retrieves statistics snapshots within the provided date range
func (s *Store) GetStatistics(ctx context.Context, dateRange *domain.DateRange) ([]domain.StatisticsSnapshot, error) {
	query := `SELECT id, timestamp, data FROM statistics_snapshots WHERE 1=1`
	args := []interface{}{}

	if dateRange != nil {
		query += ` AND timestamp >= ? AND timestamp <= ?`
		args = append(args, dateRange.From.Format(time.RFC3339), dateRange.To.Format(time.RFC3339))
	}

	query += ` ORDER BY timestamp DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.StatisticsSnapshot
	for rows.Next() {
		var snapshot domain.StatisticsSnapshot
		var timestampStr string
		var dataJSON string

		err := rows.Scan(&snapshot.ID, &timestampStr, &dataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan statistics snapshot: %w", err)
		}

		snapshot.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		if err := json.Unmarshal([]byte(dataJSON), &snapshot.Statistics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating statistics: %w", err)
	}

	return snapshots, nil
}

// GetLatestStatistics retrieves the most recent statistics snapshot
func (s *Store) GetLatestStatistics(ctx context.Context) (*domain.StatisticsSnapshot, error) {
	var snapshot domain.StatisticsSnapshot
	var timestampStr string
	var dataJSON string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, timestamp, data FROM statistics_snapshots
		ORDER BY timestamp DESC
		LIMIT 1
	`).Scan(&snapshot.ID, &timestampStr, &dataJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest statistics: %w", err)
	}

	snapshot.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	if err := json.Unmarshal([]byte(dataJSON), &snapshot.Statistics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statistics: %w", err)
	}

	return &snapshot, nil
}

// UpsertAssignmentSnapshot inserts or updates an assignment snapshot
func (s *Store) UpsertAssignmentSnapshot(ctx context.Context, snapshot domain.AssignmentSnapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assignment_snapshots (date, srs_stage, subject_type, count)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date, srs_stage, subject_type) DO UPDATE SET
			count = excluded.count
	`, snapshot.Date.Format("2006-01-02"), snapshot.SRSStage, snapshot.SubjectType, snapshot.Count)

	if err != nil {
		return fmt.Errorf("failed to upsert assignment snapshot: %w", err)
	}

	return nil
}

// GetAssignmentSnapshots retrieves assignment snapshots within the provided date range
func (s *Store) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) ([]domain.AssignmentSnapshot, error) {
	query := `SELECT date, srs_stage, subject_type, count FROM assignment_snapshots WHERE 1=1`
	args := []interface{}{}

	if dateRange != nil {
		query += ` AND date >= ? AND date <= ?`
		args = append(args, dateRange.From.Format("2006-01-02"), dateRange.To.Format("2006-01-02"))
	}

	query += ` ORDER BY date ASC, srs_stage ASC, subject_type ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.AssignmentSnapshot
	for rows.Next() {
		var snapshot domain.AssignmentSnapshot
		var dateStr string

		err := rows.Scan(&dateStr, &snapshot.SRSStage, &snapshot.SubjectType, &snapshot.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment snapshot: %w", err)
		}

		snapshot.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment snapshots: %w", err)
	}

	return snapshots, nil
}

// CalculateAssignmentSnapshot computes a snapshot from current assignments for a given date
func (s *Store) CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]domain.AssignmentSnapshot, error) {
	// Query to count assignments by SRS stage and subject type
	// Exclude SRS stage 0 (unstarted assignments) as per requirement 12.2
	query := `
		SELECT 
			json_extract(data, '$.srs_stage') as srs_stage,
			json_extract(data, '$.subject_type') as subject_type,
			COUNT(*) as count
		FROM assignments
		WHERE json_extract(data, '$.srs_stage') > 0
		GROUP BY srs_stage, subject_type
		ORDER BY srs_stage, subject_type
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment counts: %w", err)
	}
	defer rows.Close()

	var snapshots []domain.AssignmentSnapshot
	for rows.Next() {
		var snapshot domain.AssignmentSnapshot
		var srsStage int
		var subjectType string
		var count int

		err := rows.Scan(&srsStage, &subjectType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment count: %w", err)
		}

		snapshot.Date = date
		snapshot.SRSStage = srsStage
		snapshot.SubjectType = subjectType
		snapshot.Count = count

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignment counts: %w", err)
	}

	return snapshots, nil
}

// GetLastSyncTime retrieves the last successful sync timestamp for a data type
func (s *Store) GetLastSyncTime(ctx context.Context, dataType domain.DataType) (*time.Time, error) {
	var lastSyncTimeStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT last_sync_time FROM sync_metadata WHERE data_type = ?
	`, string(dataType)).Scan(&lastSyncTimeStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query last sync time: %w", err)
	}

	lastSyncTime, err := time.Parse(time.RFC3339, lastSyncTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last sync time: %w", err)
	}

	return &lastSyncTime, nil
}

// SetLastSyncTime updates the last successful sync timestamp for a data type
func (s *Store) SetLastSyncTime(ctx context.Context, dataType domain.DataType, timestamp time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sync_metadata (data_type, last_sync_time)
		VALUES (?, ?)
		ON CONFLICT(data_type) DO UPDATE SET
			last_sync_time = excluded.last_sync_time
	`, string(dataType), timestamp.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to set last sync time: %w", err)
	}

	return nil
}

// validateSubjectExists checks if a subject with the given ID exists in the database
func (s *Store) validateSubjectExists(ctx context.Context, tx *sql.Tx, subjectID int) error {
	var exists bool
	var query string
	var err error

	if tx != nil {
		query = `SELECT EXISTS(SELECT 1 FROM subjects WHERE id = ?)`
		err = tx.QueryRowContext(ctx, query, subjectID).Scan(&exists)
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM subjects WHERE id = ?)`
		err = s.db.QueryRowContext(ctx, query, subjectID).Scan(&exists)
	}

	if err != nil {
		return fmt.Errorf("failed to check subject existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("subject with ID %d does not exist", subjectID)
	}

	return nil
}

// validateAssignmentExists checks if an assignment with the given ID exists in the database
func (s *Store) validateAssignmentExists(ctx context.Context, tx *sql.Tx, assignmentID int) error {
	var exists bool
	var query string
	var err error

	if tx != nil {
		query = `SELECT EXISTS(SELECT 1 FROM assignments WHERE id = ?)`
		err = tx.QueryRowContext(ctx, query, assignmentID).Scan(&exists)
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM assignments WHERE id = ?)`
		err = s.db.QueryRowContext(ctx, query, assignmentID).Scan(&exists)
	}

	if err != nil {
		return fmt.Errorf("failed to check assignment existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("assignment with ID %d does not exist", assignmentID)
	}

	return nil
}
