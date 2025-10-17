## Astron-xmod-shim 配置文件结构说明

Astron-xmod-shim 采用分层配置管理机制，配置文件主要集中在 conf/ 目录下，按照功能和插件类型进行分类存放。以下是项目的配置文件结构及其各文件的含义和归属：

conf/
├── base/
│ └── conf.yaml # 核心配置文件
└── shimlets/
├── k8s-shimlet.yaml # Kubernetes shimlet 配置
└── kubeconfig # Kubernetes 集群连接配置

## 各配置文件详细说明

1. 核心配置文件 (conf/base/conf.yaml)
   归属：系统级核心配置，由主程序直接加载和解析。

   主要功能：定义整个应用的基础配置参数，包括Kubernetes客户端、HTTP服务器、日志系统、当前激活的shimlet、各shimlet插件的配置路径、模型管理和跟踪器等核心组件的配置。

2. Shimlet配置文件 (conf/shimlets/k8s-shimlet.yaml)
   归属：Kubernetes shimlet插件专用配置文件。

主要功能：定义内置的 Kubernetes shimlet插件的特定配置参数，用于对接Kubernetes集群。

## 配置扩展性

从配置结构可以看出，项目采用了模块化设计，支持多种shimlet插件的扩展。新的shimlet插件可以通过在conf/base/conf.yaml的shimlets部分注册，并提供对应的配置文件路径来实现集成。