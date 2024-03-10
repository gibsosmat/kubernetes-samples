# Armada gRPC setup
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
main.go has a small http server listneing to submit-job
```bash
go run main.go
```
then open http://localhost:8090/submit-job/ 
This will submit job sample from job-queue-a.yaml to armada cluster.

BUT This will fail the first time if queue-a is not already created.
new queue can be created using armadactl in armada git repo
```bash
go run cmd/armadactl/main.go create -f ./docs/quickstart/queue-a.yaml
```
