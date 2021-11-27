# Golib

Common core for Golang project.

### ⚠️ **Notice** ⚠️
Our modules are in private repo, so you need to config something bellow before start your develop.
#### 1. Setup `GOPRIVATE`

Run the command `export GOPRIVATE="gitlab.id.vin"` to add `gitlab.id.vin` as private repo.

For future usage, you might add above command to `.bashrc` or `.zshrc`.

#### 2. Add credentials to private host
Run the following command line to load `https://gitlab.id.vin/` using SSH:
```shell
git config --global url."git@gitlab.id.vin:".insteadOf "https://gitlab.id.vin/"
```

Or with access token in URL:
```shell
git config --global url."https://oath2:{your_access_token}@gitlab.id.vin/".insteadOf https://gitlab.id.vin/
```

### Setup instruction

Both `go get` and `go mod` are supported.
```shell
go get gitlab.com/golibs-starter/golib
```

### Usage

Using `fx.Option` to include dependencies for injection.

```go
package main

import (
    "gitlab.com/golibs-starter/golib"
    "gitlab.com/golibs-starter/golib/web/client"
    "go.uber.org/fx"
)

func main() {
    options := []fx.Option{
        // Required
        golib.AppOpt(),

        // Required
        golib.PropertiesOpt(),

        // When you want to use default logging strategy.
        golib.LoggingOpt(),

        // When you want to enable event publisher
        golib.EventOpt(),

        // When you want to enable actuator endpoints.
        // By default, we provide HealthService and InfoService.
        golib.ActuatorEndpointOpt(),
        // When you want to provide build info to above InfoService.
        golib.BuildInfoOpt(Version, CommitHash, BuildTime),

        // When you want to enable http client auto config with contextual client by default
        golib.HttpClientOpt(),

        // When you want to tell GoLib to load your properties.
        golib.ProvideProps(properties.NewCustomProperties),

        // When you want to register your event listener.
        golib.ProvideEventListener(listener.NewCustomListener),

        // Example using http client
        fx.Provide(NewExampleService),
    }
}

type ExampleService struct {
    httpClient client.ContextualHttpClient
}

func NewExampleService(httpClient client.ContextualHttpClient) *ExampleService {
    return &ExampleService{httpClient: httpClient}
}
```

### Configuration

#### 1. Environment variables

| Var | Default | Description |
|---|---|---|
| `APP_PROFILES` | None | Defines the list of active profiles, separate by comma. By default, `default` profile is always load even this env configured. Example: when `APP_PROFILES=internal,uat` then both `default` `internal` and `uat` will be loaded by order.  |
| `APP_CONFIG_PATHS` | `./config` | Defines the location of config directory, when the application is started, it will scan profiles in this path. |
| `APP_CONFIG_FORMAT` | `yaml` | Defines the format of config file. Currently we only support Yaml format (both `yaml` `yml` are accepted). |

Besides, all our configs can be overridden by environment variables. For example:

```yaml
store:
    name: Fruit store # Equivalent to STORE_NAME
    items:
        - name: Apple # Equivalent to STORE_ITEMS_0_NAME
          price: 5 # Equivalent to STORE_ITEMS_0_PRICE
        - name: Lemon # Equivalent to STORE_ITEMS_1_NAME
          price: 0.5 # Equivalent to STORE_ITEMS_1_PRICE
```

#### 2. Available configurations

```yaml
# Configuration available for AppOpt()
app:
  name: Service Name # Specify application name. Default `unspecified`
  port: 8080 # Defines the running port. Default `8080`
  path: /service-base-path/ # Defines base path (context path). Default `/`

  # Configuration available for LoggingOpt()
  logging:
    development: false # Enable or disable development mode. Default `false`
    jsonOutputMode: true # Enable or disable json output. Default `true`

vinid:
  event: # Configuration available for EventOpt()
    notLogPayloadForEvents:
      - OrderCreatedEvent
      - OrderUpdatedEvent
  httpClient: # Configuration for HttpClientOpt()
    timeout: 60s # Request timeout, in duration format. Default 60s
    maxIdleConns: 100 # Default 100
    maxIdleConnsPerHost: 10 # Default 10
    maxConnsPerHost: 100 # Default 100
    proxy:
      url: http://localhost:8080 # Proxy url
      appliedUris: # List of URIs, which will be requested under above proxy
        - https://foo.com/path/
        - https://bar.com/path/
```
