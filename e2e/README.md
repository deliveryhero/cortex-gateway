# Run local End-to-End dev environment

Adapt and export the following environment variables

```
export BEARER_TOKEN=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJkZXZlbG9wbWVudCIsInZlcnNpb24iOjEsImF1ZCI6InJld2UifQ.a7zdFoVYTk_CKR1Cj3h6S3SBGPseS0x3PNJjy4jdvPtWaraDC638QlXu0CeBMf3vLWPBJB0fopbSMQvm6IoPtw
export JWT_SECRET=<The secret used to sign the BEARER_TOKEN>

export AGENT_CLIENT_ID=<Auth0 M2M Client ID>
export AGENT_CLIENT_SECRET=<Auth0 M2M Client Secret>

export OAUTH_TOKEN_URL=https://<tenant>.<region>.auth0.com/oauth/token
export OAUTH_AUTH_URL=https://<tenant>.<region>.auth0.com/authorize
export OAUTH_API_URL=https://<tenant>.<region>.auth0.com/userinfo
export OAUTH_AUDIENCE=<Auth0 API Audience for M2M client>
export OAUTH_JWKS_URL=https://<tenant>.<region>.auth0.com/.well-known/jwks.json

export GRAFANA_ROLE_ATTRIBUTE_PATH="contains(\"https://example.com/roles\"[*], 'admin') && 'Admin' || contains(\"https://example.com/roles\"[*], 'editor') && 'Editor' || 'Viewer'"
export GRAFANA_CLIENT_ID=<Auth0 Grafana Application Client ID>
export GRAFANA_CLIENT_SECRET=<Auth0 Grafana Application Client Secret>

export TENANT_ID=<TenantID Claim> # example: https://example.com/tenant_id
```

Run locally the following command
`go run main.go -gateway.distributor.address http://localhost:9009 -gateway.query-frontend.address http://localhost:9009 -gateway.ruler.address http://localhost:9009 -gateway.alertmanager.address http://localhost:9009 -server.http-listen-port 8900 -server.grpc-listen-port 8995 -log.level debug -gateway.auth.jwt-secret $JWT_SECRET -gateway.auth.jwt-extra-headers "X-Id-Token" -gateway.auth.tenant-id-claim $TENANT_ID -gateway.auth.jwks-url $OAUTH_JWKS_URL`

Uncomment the relevant `remote_write` config in the `agent/config.yaml`.

Start Cortex, Grafana and Grafana Agent
`docker-compose up`
