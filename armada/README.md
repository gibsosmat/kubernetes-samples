# Armada local setup
## Pre-requisites

**System should have minimum 4 core cpu & 16GB ram**

install following software packages
- [Go](https://go.dev/doc/install) (version 1.20 or later)
- gcc (for Windows, see, e.g., [tdm-gcc](https://jmeubank.github.io/tdm-gcc/))
- [mage](https://magefile.org/)
- [docker](https://docs.docker.com/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [protoc](https://github.com/protocolbuffers/protobuf/releases)

for UI you need 
- [node.js & npm](https://nodejs.org/en/learn/getting-started/how-to-install-nodejs) preferably using [nvm](https://github.com/nvm-sh/nvm)
- [yarn](https://classic.yarnpkg.com/en/docs/install)

## Installation

Install the [Pre-requisites](#pre-requisites) and then run:

```bash
git clone https://github.com/armadaproject/armada.git
cd armada

mage localdev full
```

## Execution
following commands will
- create job queues
- send jobs/tasks for them to be executed in pods  
currently this executes an `alpine` linux image and execute `sleep` for 10+ seconds

```bash
go run cmd/armadactl/main.go create -f ./docs/quickstart/queue-a.yaml
go run cmd/armadactl/main.go submit ./docs/quickstart/job-queue-a.yaml

go run cmd/armadactl/main.go create -f ./docs/quickstart/queue-b.yaml
go run cmd/armadactl/main.go submit ./docs/quickstart/job-queue-b.yaml

#check logs of armada in another terminal
docker compose logs -f
```
input files 
```yaml
# cat ./docs/quickstart/queue-a.yaml 
apiVersion: armadaproject.io/v1beta1
kind: Queue
name: queue-a
permissions:
- subjects: 
  - name: group2
    kind: Group
  verbs:
  - cancel
  - reprioritize
  - watch
priorityFactor: 3.0
resourceLimits:
  cpu: 1.0
  memory: 1.0%

# cat ./docs/quickstart/job-queue-a.yaml
queue: queue-a
jobSetId: job-set-1
jobs:
  - priority: 0
    podSpec:
      terminationGracePeriodSeconds: 0
      restartPolicy: Never
      containers:
        - name: sleeper
          image: alpine:latest
          command:
            - sh
          args:
            - -c
            - sleep $(( (RANDOM % 60) + 10 ))
          resources:
            limits:
              memory: 128Mi
              cpu: 0.2
            requests:
              memory: 128Mi
              cpu: 0.2
```


## Sample output
you can also use `./armadactl` if it is built & with env variables to point to local

### watch queue status
```bash
#this command shows current state of queue and the jobs
bandy@Sandeeps-MacBook-Pro armada % ./armadactl watch queue-a job-set-1
Watching job set job-set-1
Mar  5 14:28:28 | Queued:   0, Leased:   0, Pending:   0, Running:   0, Succeeded:   0, Failed:   0, Cancelled:   0 | JobSubmittedEvent, job id: 01hr7g847v17qs20s15kwscckq
Mar  5 14:28:28 | Queued:   1, Leased:   0, Pending:   0, Running:   0, Succeeded:   0, Failed:   0, Cancelled:   0 | JobQueuedEvent, job id: 01hr7g847v17qs20s15kwscckq
Mar  5 14:28:30 | Queued:   0, Leased:   1, Pending:   0, Running:   0, Succeeded:   0, Failed:   0, Cancelled:   0 | JobLeasedEvent, job id: 01hr7g847v17qs20s15kwscckq
Mar  5 14:28:37 | Queued:   0, Leased:   0, Pending:   1, Running:   0, Succeeded:   0, Failed:   0, Cancelled:   0 | JobPendingEvent, job id: 01hr7g847v17qs20s15kwscckq pod: 0
Mar  5 14:28:46 | Queued:   0, Leased:   0, Pending:   0, Running:   1, Succeeded:   0, Failed:   0, Cancelled:   0 | JobRunningEvent, job id: 01hr7g847v17qs20s15kwscckq pod: 0
Mar  5 14:29:23 | Queued:   0, Leased:   0, Pending:   0, Running:   0, Succeeded:   1, Failed:   0, Cancelled:   0 | JobSucceededEvent, job id: 01hr7g847v17qs20s15kwscckq pod: 0
```
### pod status
```bash
bandy@Sandeeps-MacBook-Pro armada % kubectl get pods -A -o wide                            
NAMESPACE            NAME                                                READY   STATUS      RESTARTS      AGE     IP           NODE                        NOMINATED NODE   READINESS GATES
default              armada-01hr7g847v17qs20s15kwscckq-0                 0/1     Completed   0             2m59s   10.244.1.4   armada-test-worker          <none>           <none>
default              armada-01hr7gbxs10cv4qwpv1dpv314c-0                 0/1     Completed   0             49s     10.244.1.5   armada-test-worker          <none>           <none>
ingress-nginx        ingress-nginx-admission-create-w24x2                0/1     Completed   0             171m    10.244.1.3   armada-test-worker          <none>           <none>
ingress-nginx        ingress-nginx-admission-patch-vwsrs                 0/1     Completed   1             171m    10.244.1.2   armada-test-worker          <none>           <none>
ingress-nginx        ingress-nginx-controller-646df5f698-z77l4           1/1     Running     0             171m    10.244.0.5   armada-test-control-plane   <none>           <none>
kube-system          coredns-6d4b75cb6d-czqrt                            1/1     Running     0             172m    10.244.0.4   armada-test-control-plane   <none>           <none>
kube-system          coredns-6d4b75cb6d-h6jb4                            1/1     Running     0             172m    10.244.0.3   armada-test-control-plane   <none>           <none>
kube-system          etcd-armada-test-control-plane                      1/1     Running     0             173m    172.18.0.3   armada-test-control-plane   <none>           <none>
kube-system          kindnet-4fx8p                                       1/1     Running     0             172m    172.18.0.3   armada-test-control-plane   <none>           <none>
kube-system          kindnet-jnkt2                                       1/1     Running     0             172m    172.18.0.2   armada-test-worker          <none>           <none>
kube-system          kube-apiserver-armada-test-control-plane            1/1     Running     0             173m    172.18.0.3   armada-test-control-plane   <none>           <none>
kube-system          kube-controller-manager-armada-test-control-plane   1/1     Running     9 (39m ago)   173m    172.18.0.3   armada-test-control-plane   <none>           <none>
kube-system          kube-proxy-5cmmw                                    1/1     Running     0             172m    172.18.0.2   armada-test-worker          <none>           <none>
kube-system          kube-proxy-s8dhj                                    1/1     Running     0             172m    172.18.0.3   armada-test-control-plane   <none>           <none>
kube-system          kube-scheduler-armada-test-control-plane            1/1     Running     8 (39m ago)   173m    172.18.0.3   armada-test-control-plane   <none>           <none>
local-path-storage   local-path-provisioner-6b84c5c67f-lqjxz             1/1     Running     0             172m    10.244.0.2   armada-test-control-plane   <none>           <none>
```
### worker pods status
worker pods will be created in `default` namespace, these pods exist after running the command
```bash
bandy@Sandeeps-MacBook-Pro armada % kubectl get pods
NAME                                  READY   STATUS      RESTARTS   AGE
armada-01hr7g847v17qs20s15kwscckq-0   0/1     Completed   0          3m12s
armada-01hr7gbxs10cv4qwpv1dpv314c-0   0/1     Completed   0          62s
bandy@Sandeeps-MacBook-Pro armada %                            


```

### Testing if LocalDev is working

Running `mage testsuite` will run the full test suite against the localdev cluster. This is the recommended way to test changes to the core components of Armada.

You can also run the same commands yourself:

```bash
go run cmd/armadactl/main.go create queue e2e-test-queue

# To allow Ingress tests to pass
export ARMADA_EXECUTOR_INGRESS_URL="http://localhost"
export ARMADA_EXECUTOR_INGRESS_PORT=5001

go run cmd/testsuite/main.go test --tests "testsuite/testcases/basic/*" --junit junit.xml
```

UI is being changed to v2 there would be issues, run: 

```bash
mage ui
```
To access it, open http://localhost:8089 in your browser

## A note for Devs on Arm / Windows

There is limited information on issues that appear on Arm / Windows Machines when running this setup.

Feel free to create a ticket if you encounter any issues, and link them to the relavent issue:

* https://github.com/armadaproject/armada/issues/2493 (Arm)
* https://github.com/armadaproject/armada/issues/2492 (Windows)

For required enviromental variables, please see [The Enviromental Variables Guide](https://github.com/armadaproject/armada/tree/master/developer/env/README.md).
If you would like to run the individual mage targets yourself, you can do so. See the [Manually Running LocalDev](https://github.com/armadaproject/armada/blob/master/docs/developer/manual-localdev.md) guide for more information.

### full documentation at [armada](https://github.com/armadaproject/armada)

------------------------------------------------------------------------------------
