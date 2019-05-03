# MicroService Operator Demo

This operator deploys a Spring-Boot service from a very simple custom resource.

## Installation

### Create Operator ServiceAccount, Role, Rolebinding and CustomResourceDefinition
```
$ deploy/operator-create.sh $NAMESPACE
```

### Delete Operator and Dependencies
```
$ deploy/operator-remove.sh $NAMESPACE
```


## Testing

### 1. Create Example Spring-Boot Service

```
$ oc create -f deploy/crds/paas_v1alpha1_microservice_cr.yaml -n your_project
```

```yaml
kind: MicroService
apiVersion: paas.redhat.com/v1alpha1
metadata:
  name: example-microservice
spec:
  image: "quay.io/redhatit/spring-boot-helloworld"
  replicas: 2
```
