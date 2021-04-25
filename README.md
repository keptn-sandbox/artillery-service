# Keptn Artillery Service

Keptn service for [artillery.io](https://artillery.io/) load testing tool.

---

# artillery-service
![GitHub release (latest by date)](https://img.shields.io/github/v/release/keptn-sandbox/artillery-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/keptn-sandbox/artillery-service)](https://goreportcard.com/report/github.com/keptn-sandbox/artillery-service)

## Compatibility Matrix

| Keptn Version    | [artillery-service Docker Image](https://hub.docker.com/r/artilleryio/keptn-artillery-service/tags) | Comment |
:----------------:|:----------------------------------------:|:----------------:|
|       0.8.1      | artilleryio/keptn-artilleryio-service:0.1.0 | Initial release |

## Installation

The *artillery-service* can be installed as a part of [Keptn's uniform](https://keptn.sh).

### Deploy in your Kubernetes cluster

To deploy the current version of the *artillery-service* in your Keptn Kubernetes cluster, apply the [`deploy/service.yaml`](deploy/service.yaml) file:

```console
kubectl apply -f deploy/service.yaml
```

This should install the `artillery-service` together with a Keptn `distributor` into the `keptn` namespace, which you can verify using

```console
kubectl -n keptn get deployment artillery-service -o wide
kubectl -n keptn get pods -l run=artillery-service
```

### Up- or Downgrading

Adapt and use the following command in case you want to up- or downgrade your installed version (specified by the `$VERSION` placeholder):

```console
kubectl -n keptn set image deployment/artillery-service artillery-service=artilleryio/keptn-artillery-service:$VERSION --record
```

### Uninstall

To delete a deployed *artillery-service*, use the file `deploy/*.yaml` files from this repository and delete the Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml
```

## Usage

Running the script `run.sh` will deploy the service to the Keptn Kubernetes cluster and add the resources used by the following test strategies:

| Test Strategy    | Scenario |
:----------------:|:-----------:|
|       performance      | scenarios/load.yaml |
|       functional      | scenarios/basic.yaml |
|       healthcheck      | scenarios/health.yaml |

You can update the scenarios and add them to Keptn using:

```console
keptn add-resource --project=PROJECTNAME --service=SERVICENAME --stage=STAGENAME --resource=./scenarios/basic.yaml --resourceUri=scenarios/basic.yaml
```

## Development

### Where to start

If you don't care about the details, your first entrypoint is [eventhandlers.go](eventhandlers.go). Within this file 
 you can add implementation for pre-defined Keptn Cloud events.
 
To better understand Keptn CloudEvents, please look at the [Keptn Spec](https://github.com/keptn/spec).
 
If you want to get more insights, please look into [main.go](main.go), [deploy/service.yaml](deploy/service.yaml),
 consult the [Keptn docs](https://keptn.sh/docs/) as well as existing [Keptn Core](https://github.com/keptn/keptn) and
 [Keptn Contrib](https://github.com/keptn-contrib/) services.

### Build yourself

* Build the binary: `go build -ldflags '-linkmode=external' -v -o artillery-service`
* Run tests: `go test -race -v ./...`
* Deploy the service using `kubectl`: `kubectl apply -f deploy/`
* Delete/undeploy the service using `kubectl`: `kubectl delete -f deploy/`
* Watch the deployment using `kubectl`: `kubectl -n keptn get deployment artillery-service -o wide`
* Get logs using `kubectl`: `kubectl -n keptn logs deployment/artillery-service -f`
* Watch the deployed pods using `kubectl`: `kubectl -n keptn get pods -l run=artillery-service`
* Deploy the service using [Skaffold](https://skaffold.dev/): `skaffold run --default-repo=your-docker-registry --tail` (Note: Replace `your-docker-registry` with your DockerHub username; also make sure to adapt the image name in [skaffold.yaml](skaffold.yaml))


### Testing Cloud Events

We have dummy cloud-events in the form of [RFC 2616](https://ietf.org/rfc/rfc2616.txt) requests in the [test-events/](test-events/) directory. These can be easily executed using third party plugins such as the [Huachao Mao REST Client in VS Code](https://marketplace.visualstudio.com/items?itemName=humao.rest-client).


## License

Please find more information in the [LICENSE](LICENSE) file.
