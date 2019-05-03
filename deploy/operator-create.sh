#!/bin/bash
#!/bin/bash
NAMESPACE=${1:-paas-operators}

oc create -f service_account.yaml -n $NAMESPACE
oc create -f role.yaml -n $NAMESPACE
oc create -f role_binding.yaml -n $NAMESPACE
oc create -f crds/paas_v1alpha1_microservice_crd.yaml -n $NAMESPACE
oc create -f operator.yaml -n $NAMESPACE
