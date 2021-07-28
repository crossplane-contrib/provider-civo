## provider-civo

### Overview

This provider-civo repository is the Crossplane infrastructure provider for [Civo](https://www.civo.com). 
The provider that is built from the source code in this repository can be installed into a Crossplane control plane and adds the following new functionality:
- CivoKubernetes

### Contributing
provider-civo is a community driven project and we welcome contributions. See the Crossplane Contributing guidelines to get started.

### Contact

Please use the following to reach members of the community:

Twitter: @crossplane_io

Email: info@crossplane.io

### Code of Conduct

provider-civo adheres to the same [Code of Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md) as the core Crossplane project.

## Usage

### Prerequisites
* [kubectl installed](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [KinD installed](https://kind.sigs.k8s.io/docs/user/quick-start/) (optionally another Kubernetes Cluster)
* [Helm installed9](https://helm.sh/) 
* [yq installed](https://mikefarah.gitbook.io/yq/) 

Set-up a Kubernetes cluster with Crossplane installed. The instructions can be found in the official [Crossplane documentation](https://crossplane.io/docs/v1.3/getting-started/install-configure.html#start-with-a-self-hosted-crossplane).

To add the Civo Provider Configuration Package, run:
```
kubectl crossplane install provider crossplane/provider-civo:main
```
In this case, we are going to follow the resources in the example repostory. 

Before creating a Provider resource, edit the API key in [provider.yaml](examples/civo/provider/provider.yaml). You can find the API key in your Civo account within Account>Settings>Security.

Next, we can apply the Provider:
```
kubectl apply -f examples/civo/provider/provider.yaml
```

Once the resource has been created, we can apply the cluster resource:
```
kubectl apply -f examples/civo/cluster/cluster.yaml
```

This will create a new Kubernetes cluster, according to the specifications provided in the cluster.

### Connection details

With the use of `yq` and `kubectl` it is possible to retrieve the `CivoKubernetes` kubeconfig directly.

_Example below_

```
kubectl get secrets cluster-details -o yaml | yq eval '.data.kubeconfig' - | base64 -d > kubeconfig
```

_Validating our new cluster_

```
❯ k get nodes --kubeconfig kubeconfig
NAME                                       STATUS   ROLES    AGE     VERSION
k3s-test-cluster-ec4e8ef1-node-pool-41cf   Ready    <none>   4m21s   v1.20.2+k3s1
k3s-test-cluster-ec4e8ef1-node-pool-23e0   Ready    <none>   4m13s   v1.20.2+k3s1
```