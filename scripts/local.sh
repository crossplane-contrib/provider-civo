#! /bin/bash

RESET=`tput sgr0`
RED=`tput setaf 1`
GREEN=`tput setaf 2`
CYAN=`tput setaf 6`
YELLOW=`tput setaf 3`
BOLD=`tput bold`

# verify that golang is installed
if [ -z `which go` ]; then
   printf "${RED}go is not installed! Please install\n\n${RESET}"
   exit
fi
# verify that civo cli is installed
if [ -z `which civo` ]; then
   printf "${RED}Civo CLI is not installed! Please install\n\n${RESET}"
   return
fi

if [ "$1" == "" ] || [ "$1" == "--help" ]; then
   echo "Usage:"
   echo "$ ${GREEN}./crossplane-civo-local.sh ${YELLOW}[argument]${RESET}"
   echo "Allowed arguments:"
   echo "  ${YELLOW}init${RESET} to initialise crossplane-master"
   echo "  ${YELLOW}update${RESET} to install crossplane CRDs and start the civo-provider locally from your project folder"
   echo "  ${YELLOW}example${RESET} to have a CR example to install through crossplane"
   exit
fi  

civo_api_key=`civo apikey show | tail -n +4 | head -1 | awk '{print $4}'`

if [ "$1" == "init" ]; then
   read -p "Insert the ${GREEN}civo region${RESET} in which you want to spin up the ${GREEN}crossplane-master${RESET} cluster ${YELLOW}[lon1/fra1/nyc1/phx1]${RESET}: " source_region

   if [ -z "`civo kube ls --region=${source_region} | grep crossplane-master`" ]; then
       echo "Creating ${GREEN}crossplane-master${RESET} in $source_region region..."
       civo kube create crossplane-master --region=${source_region}
   else
       echo "Cluster ${GREEN}crossplane-master${RESET} in $source_region region found"
   fi
   
   # wait the master cluster to be ready
   while [ -z "`civo kube ls --region=${source_region} | grep crossplane-master | grep ACTIVE`" ]
   do
      printf "."
      sleep 5
   done
   printf "\nDownloading the ${GREEN}kubeconfig${RESET}: in ${YELLOW}~/.kube/config_crossplane-master${RESET}\n"
   sleep 5
   civo kube config crossplane-master --region=${source_region} > ~/.kube/config_crossplane-master
   echo "Exporting the ${GREEN}KUBECONFIG${RESET} env var to ${YELLOW}~/.kube/config_crossplane-master${RESET}" 
   export KUBECONFIG=~/.kube/config_crossplane-master
   echo "Creating namespace ${GREEN}crossplane-system${RESET} in ${GREEN}crossplane-master${RESET} cluster"
   kubectl create ns crossplane-system 2>/dev/null
   
elif [ "$1" == "update" ]; then
  read -p "Insert the ${GREEN}civo region${RESET} in which you want the civo resources to be created through crossplane ${YELLOW}[lon1/fra1/nyc1/phx1]${RESET}: " destination_region
  export KUBECONFIG=~/.kube/config_crossplane-master
  echo "Regenerating CRDs"
  make generate
  echo "Applying CRDs into ${GREEN}crossplane-master${RESET} cluster"
  kubectl apply -f package/crds

  echo "Applying Secret civo-provider-secret to the ${GREEN}crossplane-master${RESET} cluster and civo-provider providerconfig pointing to ${destination_region} region..."    
  civo_api_key_base64=$(echo $civo_api_key | base64)
    
read -r -d '' YAML_MANIFEST <<EOF
apiVersion: v1
kind: Secret
metadata:
  namespace: crossplane-system
  name: civo-provider-secret
type: Opaque
data:
  credentials: ${civo_api_key_base64}
---
apiVersion: civo.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: civo-provider
spec:
  region: ${destination_region}
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: civo-provider-secret
      key: credentials
EOF
  echo "${YAML_MANIFEST}" | kubectl apply -f -

  echo "Freeing tcp port 8080..."
  lsof -i tcp:8080 | tail -n +2 | awk '{print $2}' | xargs kill

  echo "Running the civo provider locally..."
  go run cmd/provider/main.go &
  sleep 10
  pid=`lsof -i tcp:8080 | tail -n +2 | awk '{print $2}'`
  echo "Process PID for the local civo provider: ${GREEN}${BOLD}$pid${RESET}"
  

elif [ "$1" == "example" ]; then
  printf "\nExport the ${GREEN}KUBECONFIG${RESET}:\n\n"
  printf "${YELLOW}export KUBECONFIG=~/.kube/config_crossplane-master${RESET}\n\n" 

  EXAMPLE_CLUSTER='
kubectl apply -f - << EOF    
kind: CivoKubernetes
apiVersion: cluster.civo.crossplane.io/v1alpha1
metadata: 
  name: hello-crossplane
spec:
  name: hello-crossplane
  pools:
  - id: "8382e422-dcdd-461f-afb4-2ab67f171c3e"
    count: 2
    size: g3.k3s.small
  - id: "8482f422-dcdd-461g-afb4-2ab67f171c3e"
    count: 1
    size: g3.k3s.small
  connectionDetails:
    connectionSecretNamePrefix: "cluster-details"
    connectionSecretNamespace: "default"
  providerConfigRef:
    name: civo-provider
EOF
'
  echo "Create a ${GREEN}CivoCluster${RESET} with the following:"
  echo "${YELLOW}$EXAMPLE_CLUSTER${RESET}"
  printf "To verify the ${GREEN}state${RESET} of your cluster: \n\n${YELLOW}civo kube ls --region=<destination_region>${RESET}\n\n"
fi