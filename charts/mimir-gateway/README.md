# mimir-gateway

![Version: 0.1.7](https://img.shields.io/badge/Version-0.1.7-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.1.0](https://img.shields.io/badge/AppVersion-v1.1.0-informational?style=flat-square)

A Helm chart for mimir-gateway

**Homepage:** <https://github.com/rewe-digital/mimir-gateway>

## How to install this chart

Add Delivery Hero public chart repo:

```console
helm repo add mimir-gateway https://celest-io.github.io/mimir-gateway
```

A simple install with default values:

```console
helm install mimir-gateway/mimir-gateway
```

To install the chart with the release name `my-release`:

```console
helm install my-release mimir-gateway/mimir-gateway
```

To install with some set values:

```console
helm install my-release mimir-gateway/mimir-gateway --set values_key1=value1 --set values_key2=value2
```

To install with custom values file:

```console
helm install my-release mimir-gateway/mimir-gateway -f values.yaml
```

## Values

| Key                                   | Type   | Default                                     | Description                                |
| ------------------------------------- | ------ | ------------------------------------------- | ------------------------------------------ |
| affinity                              | object | `{}`                                        |                                            |
| args.alertmanagerAddress              | string | `"http://your_alertmanager_address_here"`   |                                            |
| args.distributorAddress               | string | `"http://your_distributor_address_here"`    |                                            |
| args.jwtExtraHeaders                  | string | `""`                                        |                                            |
| args.jwtSecret                        | string | `"your_jwt_secret"`                         |                                            |
| args.jwksURL                          | string | `""`                                        |                                            |
| args.jwksRefreshEnabled               | bool   | `false`                                     |                                            |
| args.jwksRefreshInterval              | int    | `10`                                        |                                            |
| args.queryfrontendAddress             | string | `"http://your_query_frontend_address_here"` |                                            |
| args.rulerAddress                     | string | `"http://your_ruler_address_here"`          |                                            |
| args.tenantIdClaim                    | string | `""`                                        |                                            |
| args.tenantName                       | string | `""`                                        |                                            |
| extraLabels                           | object | `{}`                                        | Any extra labels to apply to all resources |
| fullnameOverride                      | string | `""`                                        |                                            |
| image.pullPolicy                      | string | `"IfNotPresent"`                            |                                            |
| image.repository                      | string | `"goelankit/mimir-gateway"`                 |                                            |
| image.tag                             | string | `""`                                        |                                            |
| imagePullSecrets                      | list   | `[]`                                        |                                            |
| ingress.enabled                       | bool   | `false`                                     |                                            |
| nameOverride                          | string | `""`                                        |                                            |
| nodeSelector                          | object | `{}`                                        |                                            |
| podAnnotations                        | object | `{}`                                        |                                            |
| podSecurityContext                    | object | `{}`                                        |                                            |
| replicaCount                          | int    | `1`                                         |                                            |
| resources                             | object | `{}`                                        |                                            |
| securityContext                       | object | `{}`                                        |                                            |
| service.annotations                   | object | `{}`                                        |                                            |
| service.port                          | int    | `80`                                        |                                            |
| service.type                          | string | `"ClusterIP"`                               |                                            |
| serviceAccount.annotations            | object | `{}`                                        |                                            |
| serviceAccount.create                 | bool   | `true`                                      |                                            |
| serviceAccount.name                   | string | `"mimir-gateway"`                           |                                            |
| strategy.rollingUpdate.maxSurge       | string | `"25%"`                                     |                                            |
| strategy.rollingUpdate.maxUnavailable | string | `"10%"`                                     |                                            |
| strategy.type                         | string | `"RollingUpdate"`                           |                                            |
| tolerations                           | list   | `[]`                                        |                                            |

## Maintainers

| Name    | Email              | Url |
| ------- | ------------------ | --- |
| yvigara | no-reply@celest.io |     |
