package database

import (
	"github.com/rhinoman/couchdb-go"
	"github.com/satori/go.uuid"
	"gitlab.com/scalifyme/puppet-master/puppet-master/pkg/api"
)

type JobDB struct {
	db *couchdb.Database
}

func NewJobDB(db *couchdb.Database) *JobDB {
	return &JobDB{
		db: db,
	}
}

func (db *JobDB) Get(id string) (*api.Job, error) {
	job := &api.Job{}
	rev, err := db.db.Read(id, job, nil)
	if err != nil {
		return nil, db.checkKnownErrors(err)
	}

	job.Rev = rev
	return job, nil
}

type jobList struct {
	Docs []*api.Job `json:"docs"`
}

// GetByStatus returns
func (db *JobDB) GetByStatus(status string, limit int) ([]*api.Job, error) {
	result := &jobList{}
	query := &couchdb.FindQueryParams{
		Selector: map[string]interface{}{
			"status": map[string]interface{}{
				"$eq": status,
			},
		},
		Limit: limit,
		Skip:  0,
	}

	if err := db.db.Find(result, query); err != nil {
		return nil, err
	}

	return result.Docs, nil
}

func (db *JobDB) Save(job *api.Job) error {
	if job.ID == "" {
		job.ID = uuid.NewV4().String()
	}

	rev, err := db.db.Save(job, job.ID, job.Rev)
	if err != nil {
		return err
	}

	job.Rev = rev
	return nil
}

func (db *JobDB) Delete(job *api.Job) error {
	_, err := db.db.Delete(job.ID, job.Rev)
	return db.checkKnownErrors(err)
}

func (db *JobDB) checkKnownErrors(err error) error {
	if err == nil {
		return nil
	}

	if couchErr, ok := err.(*couchdb.Error); ok {
		if couchErr.StatusCode == 404 {
			return ErrNotFound
		}
	}

	return err
}