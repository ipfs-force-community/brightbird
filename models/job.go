package models

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/ipfs-force-community/brightbird/env"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JobType  task state
// swagger:alias
type JobType string

const (
	// CronJobType cron work
	CronJobType JobType = "cron_job"
	//PRMergedJobType run when pr merged
	PRMergedJobType JobType = "pr_merged_hook"
	//TagCreatedJobTyperun when tag create
	TagCreatedJobType JobType = "tag_created_hook"
)

// Job
// swagger:model job
type Job struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	TestFlowId   primitive.ObjectID `json:"testFlowId"`
	Name         string             `json:"name"`
	JobType      JobType            `json:"jobType"`
	ExecCount    int                `json:"execCount"`
	Description  string             `json:"description"`
	GlobalParams env.GlobalParams   `json:"globalParams"` //override value for params

	Versions map[string]string `json:"versions"` // save a version setting for user job specific
	//cron job params
	CronJobParams
	PRMergedJobParams
	TagCreateJobParams

	BaseTime `bson:",inline"`
}

func (job Job) CheckParams() error {
	switch job.JobType {
	case CronJobType:
		_, err := cron.ParseStandard(job.CronExpression)
		return err
	case PRMergedJobType:
		for _, match := range job.PRMergedJobParams.PRMergedEventMatches {
			if len(match.BasePattern) == 0 || len(match.SourcePattern) == 0 {
				return errors.New("pr merged job must have dest and source branch")
			}
			_, err := regexp.Compile(match.BasePattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.BasePattern, err)
			}

			_, err = regexp.Compile(match.SourcePattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.SourcePattern, err)
			}
		}

	case TagCreatedJobType:
		for _, match := range job.TagCreateEventMatches {
			if len(match.TagPattern) == 0 {
				return errors.New("tag create event must have a name")
			}

			_, err := regexp.Compile(match.TagPattern)
			if err != nil {
				return fmt.Errorf("%s not a correct regex pattern %v", match.TagPattern, err)
			}
		}
	default:
		return fmt.Errorf("unsupport job type")
	}
	return nil
}

type CronJobParams struct {
	CronExpression string `json:"cronExpression"`
}

type PRMergedJobParams struct {
	PRMergedEventMatches []PRMergedEventMatch `json:"prMergedEventMatches"`
}

type PRMergedEventMatch struct {
	Repo          string `json:"repo"`
	BasePattern   string `json:"basePattern"`
	SourcePattern string `json:"sourcePattern"`
}

type TagCreateJobParams struct {
	TagCreateEventMatches []TagCreateEventMatch `json:"tagCreateEventMatches"`
}

type TagCreateEventMatch struct {
	Repo       string `json:"repo"`
	TagPattern string `json:"tagPattern"`
}
