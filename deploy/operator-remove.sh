#!/bin/bash

NAMESPACE=${1:-paas-operators}

oc delete ServiceAccount microservice-operator -n $NAMESPACE
oc delete ClusterRole microservice-operator -n $NAMESPACE
oc delete ClusterRoleBinding microservice-operator -n $NAMESPACE
oc delete CustomResourceDefinition microservices.paas.redhat.com -n $NAMESPACE
oc delete Deployment microservice-operator -n $NAMESPACE
