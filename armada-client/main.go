package main

import (
	"fmt"
	"net/http"
	"os"

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

func submitJobHandler(w http.ResponseWriter, req *http.Request) {
	e := submitJobFile("job-queue-a.yaml", false)
	if e != nil {
		fmt.Println("Error Occured")
		fmt.Println(e)
		fmt.Fprintf(w, "Error occured: %v\n", e)

	} else {
		fmt.Fprintf(w, "Successfully Submitted Job")
	}

}

func main() {
	http.HandleFunc("/submit-job", submitJobHandler)
	http.ListenAndServe(":8090", nil)
}
