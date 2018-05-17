package api

import "encoding/json"

// A Job is executed by the executor and stored in the database and holds all information
// required to let the puppets dance in the browser
type Job struct {
	ID        string            `json:"id"`
	Rev       string            `json:"_rev,omitempty"`
	Code      string            `json:"code"`
	Status    string            `json:"status"`
	Variables map[string]string `json:"variables"`
	ExitCode  int               `json:"exit_code,omitempty"`
	Logs      json.RawMessage   `json:"logs,omitempty"`
}

// A JobResult is emitted after a worker did the job and synced to database
type JobResult struct {
	JobID    string          `json:"job_id"`
	ExitCode int             `json:"exit_code"`
	Logs     json.RawMessage `json:"logs"`
}