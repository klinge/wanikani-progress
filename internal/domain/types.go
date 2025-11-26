package domain

import "time"

// DataType represents the type of WaniKani data being synced
type DataType string

const (
	DataTypeSubjects    DataType = "subjects"
	DataTypeAssignments DataType = "assignments"
	DataTypeReviews     DataType = "reviews"
	DataTypeStatistics  DataType = "statistics"
)

// Subject represents a WaniKani learning item
type Subject struct {
	ID            int       `json:"id"`
	Object        string    `json:"object"`
	URL           string    `json:"url"`
	DataUpdatedAt time.Time `json:"data_updated_at"`
	Data          SubjectData `json:"data"`
}

type SubjectData struct {
	Level      int       `json:"level"`
	Characters string    `json:"characters"`
	Meanings   []Meaning `json:"meanings"`
	Readings   []Reading `json:"readings,omitempty"`
}

type Meaning struct {
	Meaning string `json:"meaning"`
	Primary bool   `json:"primary"`
}

type Reading struct {
	Reading string `json:"reading"`
	Primary bool   `json:"primary"`
	Type    string `json:"type"`
}

// Assignment represents a user's progress on a subject
type Assignment struct {
	ID            int       `json:"id"`
	Object        string    `json:"object"`
	URL           string    `json:"url"`
	DataUpdatedAt time.Time `json:"data_updated_at"`
	Data          AssignmentData `json:"data"`
}

type AssignmentData struct {
	SubjectID   int        `json:"subject_id"`
	SubjectType string     `json:"subject_type"`
	SRSStage    int        `json:"srs_stage"`
	UnlockedAt  *time.Time `json:"unlocked_at"`
	StartedAt   *time.Time `json:"started_at"`
	PassedAt    *time.Time `json:"passed_at"`
}

// Review represents a user's answer to a quiz question
type Review struct {
	ID            int       `json:"id"`
	Object        string    `json:"object"`
	URL           string    `json:"url"`
	DataUpdatedAt time.Time `json:"data_updated_at"`
	Data          ReviewData `json:"data"`
}

type ReviewData struct {
	AssignmentID            int       `json:"assignment_id"`
	SubjectID               int       `json:"subject_id"`
	CreatedAt               time.Time `json:"created_at"`
	IncorrectMeaningAnswers int       `json:"incorrect_meaning_answers"`
	IncorrectReadingAnswers int       `json:"incorrect_reading_answers"`
}

// Statistics represents summary statistics
type Statistics struct {
	Object        string    `json:"object"`
	URL           string    `json:"url"`
	DataUpdatedAt time.Time `json:"data_updated_at"`
	Data          StatisticsData `json:"data"`
}

type StatisticsData struct {
	Lessons []LessonStatistics `json:"lessons"`
	Reviews []ReviewStatistics `json:"reviews"`
}

type LessonStatistics struct {
	AvailableAt time.Time `json:"available_at"`
	SubjectIDs  []int     `json:"subject_ids"`
}

type ReviewStatistics struct {
	AvailableAt time.Time `json:"available_at"`
	SubjectIDs  []int     `json:"subject_ids"`
}

type StatisticsSnapshot struct {
	ID         int        `json:"id"`
	Timestamp  time.Time  `json:"timestamp"`
	Statistics Statistics `json:"statistics"`
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Remaining int
	ResetAt   time.Time
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	DataType       DataType
	RecordsUpdated int
	Success        bool
	Error          string
	Timestamp      time.Time
}

// Filter types for querying
type SubjectFilters struct {
	Type  string
	Level *int
}

type AssignmentFilters struct {
	SRSStage *int
}

type ReviewFilters struct {
	From *time.Time
	To   *time.Time
}

type DateRange struct {
	From time.Time
	To   time.Time
}
