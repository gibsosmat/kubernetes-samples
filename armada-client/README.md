# Armada gRPC setup
This is a sample http service that interacts with Aramada cluster over gRPC to schedule jobs and get its status updates as a HTTP stream. For the sake of simplicity I ignored passing parameters to http service.

## Pre-requisites

build armada service and run it with following command in its git repo folder

```bash
mage localdev full
```


## Execution
clone this repo to a different location other than the armada repo folder.
```bash
cd ..
git clone https://github.com/gibsosmat/kubernetes-samples.git
cd kubernetes-samples/armada-client/
```

main.go has a small http server listneing on port 8090

```bash
go run main.go
```

### Submit Job
Run the following command OR open the link in browser too. This will submit job sample from job-queue-a.yaml to armada cluster on each request.

⚠️ do not submit too many requests if it is laptop

```bash
curl http://localhost:8090/submit-job/
```
> [!NOTE] 
> This will fail if `queue-a` is not already created. it can be created with armadactl
> ```bash
> go run cmd/armadactl/main.go create -f ./docs/quickstart/queue-a.yaml
> ```

### Stream Job Status
To check the status of the jobset, following URL sends Server Sent Events over http. chrome browser can display it (press Escape to stop)

```bash
curl http://localhost:8090/get-job-status
```

