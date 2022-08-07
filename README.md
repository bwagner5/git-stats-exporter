# git-stats-exporter
Fetch and emit github stats as prometheus metrics.


## Description
The Git stats exporter is a Kubernetes controller that fetches Github statistics like open pull requests, issues, stars, etc. and emits them as prometheus metrics to build nice dashboards.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

## Example:


```
$ kubectl get repo aws-ec2-instance-selector -o yaml
apiVersion: src.bwag.me/v1
kind: Repo
metadata:
  name: aws-ec2-instance-selector
spec:
  owner: aws
  name: amazon-ec2-instance-selector
  ghTokenSecretRef: gh-token
status:
  lastQuery: "2022-08-07T17:34:45Z"
  state: Synchronized
```

```
$ curl git-stats-exporter:8080/metrics
...
# HELP gh_repo_forks Number of forks
# TYPE gh_repo_forks gauge
gh_repo_forks{owner="aws",repo="amazon-ec2-instance-selector"} 67
# HELP gh_repo_open_issues Number of open issues
# TYPE gh_repo_open_issues gauge
gh_repo_open_issues{owner="aws",repo="amazon-ec2-instance-selector"} 11
# HELP gh_repo_open_pull_requests Number of open pull requests
# TYPE gh_repo_open_pull_requests gauge
gh_repo_open_pull_requests{owner="aws",repo="amazon-ec2-instance-selector"} 1
# HELP gh_repo_stars Number of stars
# TYPE gh_repo_stars gauge
gh_repo_stars{owner="aws",repo="amazon-ec2-instance-selector"} 406
# HELP gh_repo_subscribers Number of subscribers
# TYPE gh_repo_subscribers gauge
gh_repo_subscribers{owner="aws",repo="amazon-ec2-instance-selector"} 11
...
```

### Running on the cluster
1. Setup a Github Token as a K8s Secret (optional)

The github API has a fairly low API request limit, so setting up a github token can be useful if you are monitoring more than on repository.
https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token

```sh
echo -n "<gh-token>" > /tmp/gh-token
kubectl create secret generic gh-token --from-file /tmp/gh-token
```

2. Install Sample Repos (or create your own):

```sh
kubectl apply -f config/samples/
```

3. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/git-stats-exporter:tag
```
	
4. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/git-stats-exporter:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```


### How it works
Create `Repo` resources that point to your Github repositories and collect the metrics with prometheus.


### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

