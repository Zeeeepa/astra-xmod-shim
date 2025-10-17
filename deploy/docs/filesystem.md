# 模型目录挂载说明


Astron-xmod-shim 采用共享存储模式挂载模型目录，具体特点如下：

- 主机上的模型目录直接挂载至推理服务 Pod。
- 控制平面与推理服务使用相同的模型目录路径。
- 无需额外的文件复制或同步操作。

该设计避免了模型文件的冗余存储，提升了访问效率。

## 实现方式

### 系统级挂载（控制平面）

Astron-xmod-shim 作为控制平面，通过 Helm Chart 配置将主机目录挂载到容器中。示例如下：

```yaml
# values.yaml
volumes:
  - name: maasmodels-volume
    hostPath:
      path: /mnt/maasmodels/
      type: Directory

volumeMounts:
  - name: maasmodels-volume
    mountPath: /mnt/maasmodels/
    readOnly: true
```

### 推理服务挂载（推理 Pod）

创建推理服务时，系统自动为每个 Pod 配置相同的 HostPath 挂载：

```go
spec.WithVolumes(
    corev1apply.Volume().
        WithName("models").
        WithHostPath(
            corev1apply.HostPathVolumeSource().
                WithPath(modelDirPath),
                WithType(corev1.HostPathDirectory),
        ),
)

container.WithVolumeMounts(
    corev1apply.VolumeMount().
        WithName("models").
        WithMountPath(modelDirPath),
)

container.WithArgs(
    "--host=0.0.0.0",
    "--port="+portStr,
    "--model="+modelDirPath,
)
```

## 关键要求

- 主机路径与容器内挂载路径必须完全一致。
- 模型文件直接从主机读取，无需同步。
- 使用 HostPath 实现低开销、高性能的文件访问。
- 模型路径只需在 Helm 配置中定义一次，系统自动应用于推理服务。

## 修改默认路径的注意事项

默认模型目录为 `/mnt/maasmodels/`。如需修改，必须同时满足以下条件：

1. 更新 Helm `values.yaml` 中的 `volumes.hostPath.path` 和 `volumeMounts.mountPath`。
2. 确保集群所有节点在指定路径下存在相同的模型文件。
3. 推理服务启动参数 `--model` 的值必须与挂载路径一致。

## 常见问题

- **模型加载失败**：检查主机路径是否存在，目录结构是否正确。
- **权限错误**：确认容器运行用户对模型目录具有读取权限。
- **多节点不一致**：HostPath 仅挂载本地节点目录，多节点部署时需保证各节点模型文件同步。
