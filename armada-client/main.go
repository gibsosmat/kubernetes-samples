package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/armadaproject/armada/pkg/api"
	"github.com/armadaproject/armada/pkg/client"
	"github.com/armadaproject/armada/pkg/client/domain"
	"github.com/armadaproject/armada/pkg/client/util"
	"github.com/armadaproject/armada/pkg/client/validation"
)

func submitJobFile(path string, dryRun bool) error {
	ok, err := validation.ValidateSubmitFile(path)
	if !ok {
		return err
	}

	submitFile := &domain.JobSubmitFile{}
	err = util.BindJsonOrYaml(path, submitFile)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}
	fmt.Println("submitFile:  ", submitFile)

	// TODO use client.GetQueue to validate and fail
	// or use client.CreateQueue() when it fails
	requests := client.CreateChunkedSubmitRequests(submitFile.Queue, submitFile.JobSetId, submitFile.Jobs)

	connectionDetails := &client.ApiConnectionDetails{
		ArmadaUrl: "localhost:50051",
	}
	return client.WithSubmitClient(connectionDetails, func(originalClient api.SubmitClient) error {
		c := api.CustomSubmitClient{Inner: originalClient}

		for _, request := range requests {
			response, err := client.CustomClientSubmitJobs(c, request)
			if err != nil {
				if response != nil {
					fmt.Fprintln(os.Stdout, "[JobSubmitResponse]")
					for _, jobResponseItem := range response.JobResponseItems {
						fmt.Fprintf(os.Stdout, "Error submitting job with id %s, details: %s\n", jobResponseItem.JobId, jobResponseItem.Error)
					}
				}
				fmt.Fprintln(os.Stdout, "[Error]")
				return errors.WithMessagef(err, "error submitting request %#v", request)
			}

			for _, jobResponseItem := range response.JobResponseItems {
				if jobResponseItem.Error != "" {
					fmt.Fprintf(os.Stdout, "Error submitting job: %s\n", jobResponseItem.Error)
				} else {
					fmt.Fprintf(os.Stdout, "Submitted job with id %s to job set %s\n", jobResponseItem.JobId, request.JobSetId)
				}
			}
		}
		return nil
	})
}

func getJobSetEvents(rw http.ResponseWriter, queue string, jobSetId string) error {
	raw := false
	exitOnInactive := false
	forceNewEvents := true
	forceLegacyEvents := false

	connectionDetails := &client.ApiConnectionDetails{
		ArmadaUrl: "localhost:50051",
	}

	return client.WithEventClient(connectionDetails, func(c api.EventClient) error {
		fmt.Fprintf(rw, "Watching job set %s\n", jobSetId)
		fmt.Println("Starting gRPC event streaming call for queue:", queue, "job-set-id:", jobSetId)

		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		rw.Header().Set("Access-Control-Allow-Origin", "*")

		client.WatchJobSet(c, queue, jobSetId, true, true, forceNewEvents, forceLegacyEvents, context.Background(), func(state *domain.WatchContext, event api.Event) bool {
			if raw {
				// prints full json for debugging
				data, err := json.Marshal(event)
				if err != nil {
					fmt.Fprintf(rw, "error parsing event %s: %s\n", event, err)
				} else {
					fmt.Fprintf(rw, "%s %s\n", reflect.TypeOf(event), string(data))
				}
			} else {
				switch event2 := event.(type) {
				case *api.JobUtilisationEvent:
					// no print
				case *api.JobFailedEvent:
					// TODO print summary using writer
					fmt.Fprintf(rw, "%s\n", getPrintableSummary(state, event))
					fmt.Fprintf(rw, "Job failed: %s\n", event2.Reason)

					jobInfo := state.GetJobInfo(event2.JobId)
					if jobInfo != nil && jobInfo.ClusterId != "" && jobInfo.Job != nil {
						fmt.Fprintf(
							rw, "Found no logs for job; try '%s --tail=50\n",
							client.GetKubectlCommand(jobInfo.ClusterId, jobInfo.Job.Namespace, event2.JobId, int(event2.PodNumber), "logs"),
						)
					}
				default:
					fmt.Fprintf(rw, "%s\n", getPrintableSummary(state, event))
				}
			}
			if exitOnInactive && state.GetNumberOfJobs() == state.GetNumberOfFinishedJobs() {
				return true
			}
			return false
		})
		fmt.Println("Stopped gRPC event streaming call for queue:", queue, "job-set-id:", jobSetId)
		return nil
	})
}

func getPrintableSummary(state *domain.WatchContext, e api.Event) string {
	summary := fmt.Sprintf("%s | ", e.GetCreated().Format(time.Stamp))
	summary += state.GetCurrentStateSummary()
	summary += fmt.Sprintf(" | %s, job id: %s", reflect.TypeOf(e).String()[5:], e.GetJobId())

	if kubernetesEvent, ok := e.(api.KubernetesEvent); ok {
		summary += fmt.Sprintf(" pod: %d", kubernetesEvent.GetPodNumber())
	}
	// fmt.Fprintf(wri, "%s\n", summary)
	return summary
}

func submitJobHandler(w http.ResponseWriter, req *http.Request) {
	e := submitJobFile("job-queue-a.yaml", false)
	if e != nil {
		fmt.Println("Error Occured")
		fmt.Println(e)
		fmt.Fprintf(w, "Error occured: %v\n", e)

	} else {
		fmt.Fprintf(w, "Successfully Submitted Job\n")
	}

}

func getJobStatusHandler(w http.ResponseWriter, req *http.Request) {
	// TODO use done channel to close gRPC streaming request
	// requestClosed := req.Context().Done()

	getJobSetEvents(w, "queue-a", "job-set-1")
	fmt.Fprintf(w, "Exiting gRPC call exit")
}

func main() {

	http.HandleFunc("/submit-job", submitJobHandler)
	http.HandleFunc("/get-job-status", getJobStatusHandler)

	fmt.Println("Starting server, Ctrl+c to exit")
	http.ListenAndServe(":8090", nil)
}
