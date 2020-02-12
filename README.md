# iotContainerSource

The Following Repo is the implementation of a custom Knative Container Source that delivers IOT data Enmasse to any Knative Service or Channel

## Prerequisites

This Project assumes the user has prior knowledge on [Knative]() and [Enmasse](https://enmasse.io/) along with the following components already setup.

1. A Running Kubernetes Cluster(Version > 1.14)  
    * I recommend Using [CodeReadyContainers](https://access.redhat.com/documentation/en-us/red_hat_codeready_containers/1.0/html/getting_started_guide/getting-started-with-codeready-containers_gsg#accessing-the-openshift-cluster_gsg) for a quick start
2. Enmasse Downloaded and [IOT features](https://enmasse.io/documentation/master/openshift/#'iot-guide-messaging-iot) enabled 
3. [Knative Serving](https://knative.dev/docs/serving/) and [Knative Eventing and Sources](https://knative.dev/docs/eventing/) Setup on your cluster 

## Usage

### Prepare iotContainerSource image

Start by cloning the code source 
```
git clone https://github.com/astoycos/iotContainerSource.git 
```
Then make sure you have access to the iotContainerSource image with one of the following options 

1. Use the image already uploaded to quay.io with tag `quay.io/astoycos/iotcontainersource` 
2. Generate your own image with provided `Dockerfile` and push to personal image repository

### Create Demo Knative Service

To confirm that the `iotContainerSource` is working we will use a provided Event Display Service that simply displays incoming messages into it's log

```yaml 
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: event-display
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-releases/github.com/knative/eventing-sources/cmd/event_display
```

To apply this service use 

`kubectl/oc apply -f demo-service.yaml`

Then to ensure the service is up use 

`kubectl/oc get ksvc`

This should display something like

```
NAME            URL                                           LATESTCREATED         LATESTREADY           READY   REASON
event-display   http://event-display.default.1.2.3.4.xip.io   event-display-gqjbw   event-display-gqjbw   True    
```

### Deploy the iotContainerSource

To deploy the iotContainerSource to you cluster, a user specific instance must be created by populating the provided `manifest.yaml` file that will be utilized by the custom containersource

```yaml
apiVersion: sources.eventing.knative.dev/v1alpha1
kind: ContainerSource
metadata:
  name: iotcontainersource
spec:
  template:
    spec:
      containers:
        - image: quay.io/astoycos/iotcontainersource:latest
          name: heartbeats
          env:
            - name: POD_NAME
              value: "mypod"
            - name: POD_NAMESPACE
              value: "event-test"
            - name: MESSAGE_URI
              value : <enmasse messaging endpoint>
            - name: MESSAGE_TYPE
              value: <enmasse message type>
            - name: MESSAGE_TENANT
              value: <enmasse message tenant>
            - name: TLS_CONFIG
              value: <Enmasse tls config> # 0:no tls 1: tls insecue 2: tls secure
            - name: TLS_PATH
              value: <absolute path to tls crt>
            - name: CLIENT_USERNAME
              value: <enmasse client username>
            - name: CLIENT_PASSWORD
              value: <enmasse client password>
  sink:
    apiVersion: serving.knative.dev/v1
    kind: Service
    name: event-display
```
For initial setup purposes we will use `POD_NAME=mypod` and `POD_NAMESPACE=event-test` but be sure to change these if using a custom knative service 

The `MESSAGE_URI` can be found with the following command (If the Enmasse "Getting started using IOT" instructions were followed)

```
oc -n myapp get addressspace iot -o jsonpath={.status.endpointStatuses[?\(@.name==\'messaging\'\)].externalHost}
```


STILL IN DEVELOPMENT 
