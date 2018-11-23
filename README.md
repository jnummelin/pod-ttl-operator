[![Build Status](https://cloud.drone.io/api/badges/jnummelin/pod-ttl-operator/status.svg)](https://cloud.drone.io/jnummelin/pod-ttl-operator)

## Pod TTL Controller

Pod TTL Controller is a controller to manage pod TTLs. TTL == Time-To-Live.

There are some use cases where a Pod should not live longer than a given threshold.

One such use case is Docker registry "corrupting" it's cache every 7 days. :)

## Running the controller

Check out the `deploy` folder. It has all the needed components to run the controller.

## Enabling TTL on Pods

The Pod TTL Controller uses Pod annotations to trigger the TTL functionality. So the example Pod that would live for 5 minutes would look like:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: ttl-pod
  labels:
    name: ttl-pod
  annotations:
    nummel.in/pod-ttl: "300"
spec:
  containers:
  - name: ttl-pod
    image: nginx
    ports:
      - containerPort: 80

```

When the operator sees the pod ready, i.e. all the containers are up-and-running, it'll start a timer with the given ttl time. When the timer expires, the operator will go and delete the pod.

## Contributing

Use issues to report bugs and feature requests.