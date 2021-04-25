#!/bin/bash

# 1. setup sockshop and carts as described in keptn tutorials
# https://tutorials.keptn.sh/tutorials/keptn-full-tour-dynatrace-08/index.html

# 2. setup artillery service
kubectl apply -f ./deploy/service.yaml

# 3. Disable jmeter-service (in case you don't want to run it in parallel)
kubectl scale deployment/jmeter-service -n "keptn" --replicas=0

# 3. Add artillery test files
keptn add-resource --project=sockshop --service=carts --stage=dev --resource=./scenarios/basic.yaml --resourceUri=scenarios/basic.yaml
keptn add-resource --project=sockshop --service=carts --stage=staging --resource=./scenarios/load.yaml --resourceUri=scenarios/load.yaml
keptn add-resource --project=sockshop --service=carts --stage=production --resource=./scenarios/health.yaml --resourceUri=scenarios/health.yaml

# 4. Trigger a delivery
keptn trigger delivery --project=sockshop --service=carts --image=docker.io/keptnexamples/carts --tag=0.12.3
