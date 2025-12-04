-- +goose Up
-- +goose StatementBegin
CREATE TABLE subjects (
	id INTEGER PRIMARY KEY,
	object TEXT NOT NULL,
	url TEXT NOT NULL,
	data_updated_at TEXT NOT NULL,
	data TEXT NOT NULL
);

CREATE TABLE assignments (
	id INTEGER PRIMARY KEY,
	object TEXT NOT NULL,
	url TEXT NOT NULL,
	data_updated_at TEXT NOT NULL,
	subject_id INTEGER NOT NULL,
	data TEXT NOT NULL,
	FOREIGN KEY (subject_id) REFERENCES subjects(id)
);

CREATE TABLE reviews (
	id INTEGER PRIMARY KEY,
	object TEXT NOT NULL,
	url TEXT NOT NULL,
	data_updated_at TEXT NOT NULL,
	assignment_id INTEGER NOT NULL,
	subject_id INTEGER NOT NULL,
	data TEXT NOT NULL,
	FOREIGN KEY (assignment_id) REFERENCES assignments(id),
	FOREIGN KEY (subject_id) REFERENCES subjects(id)
);

CREATE TABLE statistics_snapshots (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp TEXT NOT NULL,
	data TEXT NOT NULL
);

CREATE TABLE sync_metadata (
	data_type TEXT PRIMARY KEY,
	last_sync_time TEXT NOT NULL
);

CREATE INDEX idx_subjects_data_updated_at ON subjects(data_updated_at);
CREATE INDEX idx_assignments_subject_id ON assignments(subject_id);
CREATE INDEX idx_assignments_data_updated_at ON assignments(data_updated_at);
CREATE INDEX idx_reviews_assignment_id ON reviews(assignment_id);
CREATE INDEX idx_reviews_subject_id ON reviews(subject_id);
CREATE INDEX idx_reviews_data_updated_at ON reviews(data_updated_at);
CREATE INDEX idx_statistics_snapshots_timestamp ON statistics_snapshots(timestamp);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_statistics_snapshots_timestamp;
DROP INDEX IF EXISTS idx_reviews_data_updated_at;
DROP INDEX IF EXISTS idx_reviews_subject_id;
DROP INDEX IF EXISTS idx_reviews_assignment_id;
DROP INDEX IF EXISTS idx_assignments_data_updated_at;
DROP INDEX IF EXISTS idx_assignments_subject_id;
DROP INDEX IF EXISTS idx_subjects_data_updated_at;

DROP TABLE IF EXISTS sync_metadata;
DROP TABLE IF EXISTS statistics_snapshots;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS subjects;
-- +goose StatementEnd
