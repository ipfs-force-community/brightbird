# 部署

## 安装前准备

服务器至少3台，服务分配如下：

| 服务器   | 服务描述                                      |
|----------|-----------------------------------------------|
| **pc1**  | kuboard-spray，harbor，mongodb，nfs存储服务 |
| **pc2**  | k8s-master                       |
| **pc3**  | k8s-node                         |

___

在pc1上安装一下服务:

- 安装[kuboard-spray](https://github.com/eip-work/kuboard-spray)
- 安装harbor
- 安装mongodb

## k8s部署

### 部署k8s集群

使用[kuboard-spray](https://github.com/eip-work/kuboard-spray)安装Kubernetes集群，资源包:`spray-v2.18.0a-8_k8s-v1.23.9_v1.16-amd64`。

### 安装k8s多集群管理工具-Kuboard v3

安装步骤参考：[Kuboard v3](https://kuboard.cn/install/v3/install-in-k8s.html#%E6%96%B9%E6%B3%95%E4%B8%80-%E4%BD%BF%E7%94%A8-hostpath-%E6%8F%90%E4%BE%9B%E6%8C%81%E4%B9%85%E5%8C%96)

### 配置k8s集群containerd镜像服务

添加harbor仓库访问权限([参考文档](https://github.com/containerd/containerd/blob/main/docs/cri/registry.md#configure-registry-credentials)), 例如harbor仓库为192.168.200.60, 修改配置文件`/etc/containerd/config.toml`:

```toml
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."192.168.200.60"]
          endpoint = ["http://192.168.200.60"]
        [plugins."io.containerd.grpc.v1.cri".registry.configs."192.168.200.60".tls]
          insecure_skip_verify = true
        [plugins."io.containerd.grpc.v1.cri".registry.configs."192.168.200.60".auth] 
          username = "admin"
          password = "Harbor12345"
```

### 创建存储类

```bash
cd script/storageclass
kubectl apply -f rbac.yaml
kubectl apply -f nfs-storageclasses.yaml
kubectl apply -f provisioner.yaml
```

- 部署fluent-bit日志收集服务

```bash
# default 名称空间
cd scripts/fluent-bit
#
kubectl apply -f serviceaccount.yaml
kubectl apply -f configmap.yaml
kubectl apply -f clusterrole.yaml
kubectl apply -f clusterrolebindings.yaml
# 如果没有pvc(shared-dir)则创建
# create
kubectl apply -f daemonset.yaml

# 要注意configmap.yaml中的配置文件
```

## brightbird部署

### 创建brightbird名称空间

### 部署backend服务

```bash
# build and push image for backend , 根据实际情况设置PRIVATE_REGISTRY
PRIVATE_REGISTRY=192.168.200.60 make docker-push-backend
  # 根据 实际情况调整 backend-configmap.yaml，
  # config.toml : backend  启动所需的配置文件
  # kube_config : backend pod 运行时的kube config，建议直接使用kubernetes-admin的配置:(master : ~/.kube/config)
kubectl apply -f scripts/web/backend-configmap.yaml
kubectl apply -f scripts/web/backend-deployment.yaml
```

### 部署web-ui服务

```bash
# build and push image for web-ui , 根据实际情况设置PRIVATE_REGISTRY
PRIVATE_REGISTRY=192.168.200.60 make docker-push-front 
# deploy
kubectl apply -f scripts/web/ui-deployment.yaml
```

#### 创建存储卷(PVC)

```bash
# 根据情况调整namespace参数

sed 's/{namespace}/your-new-namespace/g' scripts/shared-dir/shared-dir.yaml.template > shared-dir-xxx.yaml
  # 根据实际情况调整 nfs server,path 等参数
  # nfs:
  #   path: /mnt/storage-nfs-4
  #   server: 192.168.200.35

# 创建
kubectl apply -f shared-dir-xxx.yaml
```
