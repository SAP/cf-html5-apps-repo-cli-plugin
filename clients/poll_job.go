package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"fmt"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

const (
	MAX_ATTEMPTS = 10
)

func PollJob(cliConnection plugin.CliConnection, url string) (models.CFJob, error) {
	var job models.CFJob
	var err error

	for i := 1; i <= MAX_ATTEMPTS; i++ {
		<-time.After(time.Duration(i/2) * time.Second)
		log.Tracef("Getting job by URL: %s (try %d/%d)\n", url, i, MAX_ATTEMPTS)
		job, err = GetJobByUrl(cliConnection, url)
		if err != nil {
			return job, err
		}
		if job.State == "FAILED" {
			if len(job.Errors) > 0 {
				return job, fmt.Errorf("%d %s %s", job.Errors[0].Code, job.Errors[0].Title, job.Errors[0].Detail)
			}
			return job, fmt.Errorf("Job failed. Job GUID: %s", job.GUID)
		}
		if job.State == "COMPLETE" {
			break
		}
	}

	if job.State != "COMPLETE" {
		return job, fmt.Errorf("Job polling failed. After %d attempts job did't reach final state: %+v", MAX_ATTEMPTS, job)
	}

	return job, nil
}
