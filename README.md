# LSTMServerOperator
A LSTM Time Series Prediction Server In Operator

## Description
本来是想将一个LSTM时间序列预测服务构建为Operator，但是后来发现，由于LSTM预测服务尚未实现模型热更新的功能，不需要添加LSTM相关的特殊属性；
即使LSTM时间序列预测服务后来更新了模型热更新的功能，也可以通过在CR的Spec中，添加volumeMounts字段就可以解决
因此发现，其实该Operator就是一个将Deployment和Service资源组合在一起并自动维护的Operator
与原生Deployment和Service组合不同的是，该Operator需要填入的字段仅有
1. AppImage：即要部署的服务的镜像
    不可为空，必须提供
2. ContainerPort：即容器镜像开放端口
    不可为空，必须提供
3. BackendAppReplicas：即后端服务的副本数量
    可以为空，默认为`replicas=1`，限制最大replicas为10
4. ResourceLimit：即限制容器的资源使用量
    可以为空，默认仅有`Request`
    `request.cpu=100m`,`request.memory=128Mi`
5. ServicePort：即构建的Service在集群内的端口
    可以为空，默认为`servicePort=8001`
    并且要求在范围[1~30000]之间
6. ServiceType：即构建的服务类型
    可以为空，默认为`ClusterIP`
    并且目前仅可处理`ClusterIP、NodePort`两种情况


## Getting Started

### Prerequisites
- go version v1.24.0+
- containerd version v2.1.0+
- nerdctl version v2.1.1+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**在本地构建镜像通过参数`IMG`:**

```sh
make docker-build IMG=<some-registry>/lsmtserver-operator:tag
```

**在本地集群上安装该CRD:**

```sh
make install
```

**通过参数IMG在本地集群上部署对应Operator:**

```sh
make deploy IMG=<some-registry>/lstmserver-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
kubeclt apply -f test-yaml/apps_v1_application.yaml
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
kubeclt delete -f test-yaml/apps_v1_application.yaml
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/application-operator-plus:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/application-operator-plus/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025 wuyong.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

