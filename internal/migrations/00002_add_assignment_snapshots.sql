-- +goose Up
-- +goose StatementBegin
CREATE TABLE assignment_snapshots (
	date TEXT NOT NULL,
	srs_stage INTEGER NOT NULL,
	subject_type TEXT NOT NULL,
	count INTEGER NOT NULL,
	PRIMARY KEY (date, srs_stage, subject_type)
);

CREATE INDEX idx_assignment_snapshots_date ON assignment_snapshots(date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_assignment_snapshots_date;
DROP TABLE IF EXISTS assignment_snapshots;
-- +goose StatementEnd
