# iotContainerSource

The Following Repo is the implementation of a custom [Knative Container Source]()https://knative.dev/docs/eventing/sources/ that delivers IOT data Enmasse to any Knative Service or Channel in the cloudevent format. The gernal orginazation of the stack in a real world use case could look like the following 

![stack Diagram](https://raw.githubusercontent.com/astoycos/iotContainerSource/master/docs/iotContainerSource.jpg)

## Prerequisites

This Project assumes the user has prior knowledge on [Knative](https://knative.dev/) and [Enmasse](https://enmasse.io/) along with the following components already setup.

1. A Running Kubernetes Cluster(Version > 1.14)  
    * I recommend Using [CodeReadyContainers](https://access.redhat.com/documentation/en-us/red_hat_codeready_containers/1.0/html/getting_started_guide/getting-started-with-codeready-containers_gsg#accessing-the-openshift-cluster_gsg) for a quick start
    * For the Rest of the Docs I will be using the openshift CLI command `oc`
2. Enmasse Downloaded and [IOT features](https://enmasse.io/documentation/master/openshift/#'iot-guide-messaging-iot) enabled 
3. [Knative Serving](https://knative.dev/docs/serving/) and [Knative Eventing and Sources](https://knative.dev/docs/eventing/) Setup on your cluster 

## Usage

### Prepare iotContainerSource image

Start by cloning the code source 
```
git clone https://github.com/astoycos/iotContainerSource.git 
```
Then make sure you have access to the iotContainerSource image with one of the following options 

1. Use the image already uploaded to my quay.io repo with tag `quay.io/astoycos/iotcontainersource` 
2. Generate your own image with provided `Dockerfile` and push to personal image repository
   * For more info about how to build and push an image see [Podman](https://docs.fedoraproject.org/en-US/iot/build-docker/)   

### Create Demo Knative Service on Kubernetes Instance

To confirm that the `iotContainerSource` is working we will use a provided Knative service that simply displays incoming messages into it's log

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

`oc apply -n knative-eventing -f demo-service.yaml`

Then to ensure the service is up use 

`oc get -n knative-eventing ksvc`

This should display something like the following, which lets the user know the service is up and ready to go

```
NAME            URL                                           LATESTCREATED         LATESTREADY           READY   REASON
event-display   http://event-display.default.1.2.3.4.xip.io   event-display-gqjbw   event-display-gqjbw   True    
```

### Deploy the iotContainerSource to a Kubernetes Instance

To deploy the iotContainerSource container to you cluster, a user specific instance must be created by populating the provided `manifest.yaml` file

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
              value : ${MESSAGING_HOST}
            - name: MESSAGE_PORT
              value : ${MESSAGING_PORT}
            - name: MESSAGE_TYPE
              value: <enmasse message type>
            - name: MESSAGE_TENANT
              value: <enmasse message tenant>
            - name: TLS_CONFIG
              value: <Enmasse tls config> # 0:no tls 1: tls insecue 2: tls secure
            - name: TLS_PATH
              value: "${TLS_CERT}"
            - name: CLIENT_USERNAME
              value: <enmasse client username>
            - name: CLIENT_PASSWORD
              value: <enmasse client password>
  sink:
    apiVersion: serving.knative.dev/v1
    kind: Service
    name: event-display
```
For initial setup purposes we will use `POD_NAME="mypod"` and `POD_NAMESPACE="event-test"` but be sure to change these if using a custom knative service 

The `MESSAGE_URI`, `MESSAGE_PORT` and messaging endpoint certificate -> `TLS_CERT` can be collected with the following commands (If the Enmasse ["Getting started using IOT" instructions](https://enmasse.io/documentation/master/openshift/#'iot-getting-started-messaging-iot-iot) were followed)

```
export TLS_CERT=$(oc -n myapp get addressspace iot -o jsonpath={.status.caCert} | base64 --decode)

export MESSAGING_HOST=$(oc -n myapp get addressspace iot -o jsonpath={.status.endpointStatuses[?\(@.name==\'messaging\'\)].externalHost})

export MESSAGING_PORT=443
```
Now the rest of the setup variables need to be addressed 
```
MESSAGE_TYPE : telemetry/event 
MESSAGE_TENANT: <IOTProject namespace>.<Enmasse Addressspace>
TLS_CONFIG: < 0:no tls 1: tls insecure 2: tls secure >
CLIENT_USERNAME: <Enmasse Messaging User Username>
CLIENT_PASSWORD: <Enmasse Messaging User Password>
```
Once all of the container variables are set the `iotContainerSource` is ready to be deployed using the command, which will populate it with preloaded environment variables  

```
cat manifest.yaml.in | envsubst | oc apply -n knative-eventing -f -
```
Make sure the pod was deployed correctly with, `oc get pods` which should look similar to the following 

```
[astoycos@localhost github.com]$ oc get pods 
NAME                                                              READY   STATUS    RESTARTS   AGE
containersource-iotcontain-07d8c895-4ed6-430c-bcee-1572bcan7wdk   1/1     Running   0          25s
eventing-controller-6f4bbb779b-8zznh                              1/1     Running   0          7d3h
eventing-webhook-9c697c59-4kj7q                                   1/1     Running   0          7d3h
imc-controller-675dd47677-7bpz8                                   1/1     Running   0          7d3h
imc-dispatcher-6c9875f557-tr4s6                                   1/1     Running   0          7d3h
sources-controller-6bf9f6d958-28gpt                               1/1     Running   0          7d3h
```

Now we are finally ready to test our iotContainerSource, go to a local terminal and run 

```
curl --insecure -X POST -i -u sensor1@myapp.iot:hono-secret -H 'Content-Type: application/json' --data-binary '{"temp": 5}' https://$(oc -n enmasse-infra get routes iot-http-adapter --template='{{ .spec.host }}')/telemetry
```

to simulate an iot device pushing data to the enmasse http adapter. 

now run `oc logs containersource-iotcontain-07d8c895-4ed6-430c-bcee-1572bcan7wdk` The output should resemble the following 

```
[astoycos@localhost github.com]$ oc logs containersource-iotcontain-07d8c895-4ed6-430c-bcee-1572bcan7wdk
2020/02/13 18:14:11 Sink set by Yaml
2020/02/13 18:14:11 Consuming telemetry data from enmasse endpoint: messaging-8lxzny44dx-enmasse-infra.apps.astoycos-ocp.shiftstack.com
2020/02/13 18:14:11 Consumer running, press Ctrl+C to stop...
2020/02/13 18:15:48 Device Data received
2020/02/13 18:15:48 Making cloudevent
2020/02/13 18:15:48 Sending cloudevent to http://event-display.knative-eventing.svc.cluster.local
2020/02/13 18:15:58 Consuming telemetry data from enmasse endpoint: messaging-8lxzny44dx-enmasse-infra.apps.astoycos-ocp.shiftstack.com
2020/02/13 18:15:58 Consumer running, press Ctrl+C to stop...
```

This Shows that our iot data was received by the iotContainerSource and that it sent a new CloudEvent to our temporary Knative Service which simply dumps the cloudevent into its log

If you quickly type `oc logs -l serving.knative.dev/service=event-display -c user-container --since=10m` you will see the cloud event received by the service

```
[astoycos@localhost github.com]$ kubectl logs -l serving.knative.dev/service=event-display -c user-container --since=10m
  datacontenttype: application/json
Extensions,
  deviceid: 4711
  messagetype: application/json
  resource: telemetry/myapp.iot/4711
  tenantid: myapp.iot
Data,
  {
    "temp": 5
  }
```

If you are having trouble setting up anything along the way feel free to raise an issue and I will do my best to address it :)

## Acknowledgments

* Much of the Enmasse consumer Code was borrowed from [@ctron's repo](https://github.com/ctron/hot) I could not have completed this project without his help
* etc
