package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Scalify/puppet-master-gateway/pkg/api"
	"github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
)

func (s *Server) ensureQueues() error {
	var err error
	var queues = []string{api.QueueNameJobs, api.QueueNameJobResults}

	for _, queueName := range queues {
		_, err = s.queue.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("unable to create queue %s: %v", queueName, err)
		}

		s.logger.Debugf("Checked queue %s for existence.", queueName)
	}

	return nil
}

func (s *Server) publishNewJob(job *api.Job) error {
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.queue.Publish("", api.QueueNameJobs, false, false, amqp.Publishing{
		ContentType: api.ContentTypeJSON,
		Body:        b,
	})
}

func (s *Server) consumeJobResults(ctx context.Context) {
	logger := s.logger.WithField(task, taskConsumeJobResults)

	if err := s.queue.Qos(1, 0, false); err != nil {
		logger.Fatalf("Failed to set queue QOS: %v", err)
	}

	consumer, err := s.queue.Consume(api.QueueNameJobResults, "coordinator", false, false, false, false, nil)
	if err != nil {
		logger.Fatalf("Failed to create queue consumer: %v", err)
		return
	}

	var msg amqp.Delivery

	for {
		select {
		case <-ctx.Done():
			return
		case msg = <-consumer:
			if string(msg.Body) == "" {
				continue
			}

			s.handleJobResult(logger, msg)
		}
	}
}

func (s *Server) handleJobResult(logger *logrus.Entry, msg amqp.Delivery) {
	logger.Debugf("Consuming message from queue: %v", string(msg.Body))

	var result api.JobResult
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		logger.Errorf("Failed to unmarshal json body: %v", err)
		return
	}

	if result.UUID == "" {
		logger.Errorf("Failed to process job result: object has no UUID")
		if err := msg.Ack(false); err != nil {
			logger.Errorf("Failed to ack message: %v", err)
		}
		return
	}

	l := logger.WithField(api.LogFieldJobID, result.UUID)
	l.Debugf("Loading job from db")

	job, err := s.db.Get(result.UUID)
	if err != nil {
		l.Errorf("Failed to load job from db: %v", err)
		return
	}

	job.Status = api.JobStatusDone
	job.Logs = result.Logs
	job.Error = result.Error
	job.Results = result.Results
	job.FinishedAt = result.FinishedAt
	job.Duration = result.Duration

	if err := s.db.Save(job); err != nil {
		l.Errorf("Failed to save job back to db: %v", err)
	}

	if err := msg.Ack(false); err != nil {
		l.Errorf("Failed to ack message: %v", err)
		return
	}

	l.Debugf("Done processing job result")
}

func (s *Server) produceJobs(ctx context.Context) {
	logger := s.logger.WithField(task, taskProduceJobs)
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		jobs, err := s.db.GetByStatus(api.JobStatusCreated, 10)
		if err != nil {
			logger.Fatalf("Failed to get Jobs in status new: %v", err)
		}

		logger.Debugf("Got %d Jobs in state new from db.", len(jobs))

		for _, job := range jobs {
			l := logger.WithField(api.LogFieldJobID, job.UUID)
			if err := s.publishNewJob(job); err != nil {
				l.Errorf("Failed to queue job: %v", err)
				continue
			}

			l.Debugf("Queued job.")
		}
	}
}