[![Generic badge](https://img.shields.io/badge/Stage-Early%20Alpha-red.svg)](https://shields.io/)

<img width="1012" alt="Kubernetes native load testing" src="../assets/artillery-operator-header.png">

# Known issues

## Duplicated test reports (rare)

Duplicate test reports happen when a restarted failed worker reports test run results the worker has previously reported
before it failed.

We mitigate for this inside the operator by creating Kubernetes Job resources configured with
`RestartPolicy=Never` when starting Pods and `BackoffLimit=0` for the actual Job. This attempts to run the Job once
without restarting failed Pods (workers) for almost all circumstances.

This should stop failed workers from restarting therefore stopping duplicate test reports.

**HOWEVER**, the communication between the job controller, apiserver, kubelet and containers is not atomic. So, it is
possible that under some very unlikely combination of crashes/restarts in the kubelet, containers and/or job controller,
that the program in the Job's pod template might be started twice.

Here's a basic example of how things can go wrong:

- A user manually deletes a Pod using `kubectl`.
- The Load Test's Job resource create another one Pod to restart the work, therefore duplicating test results. 
