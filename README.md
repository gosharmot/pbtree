# pbtree

A tool for downloading dependencies and remote `.proto` files for later building based on `buf`.
<a name="твоё_название"></a>
и там где это необходимо, ссылку на этот якорь:

# Table of Contents
1. [Installation](#installation)
2. [Quick start](#quick-start)
3. [Full documentation](#full-documentation)
   1. [Init pbtree](#init-pbtree)
      1. [pbtree.yaml](#pbtreeyaml) 
      2. [buf.gen.yaml](#bufgenyaml)
   2. [Add new proto service to project](#add-new-proto-service-to-project)
   3. [Vendoring and generate](#vendoring-and-generate)
      1. [Remote files](#remote-files) 
      2. [Buf M flags](#buf-m-flags)

# Installation
<a id="installation"></a>
```shell
go install github.com/gosharmot/pbtree@latest
```

# Quick start

Run with script with replaced `github api token`.

```shell
mkdir -p awesome-project/bin && cd awesome-project &&  \
export GOBIN=$(pwd)/bin && \
go install github.com/bufbuild/buf/cmd/buf@v1.13.1 && \
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.18.1 && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.18.1 && \
go install github.com/gosharmot/pbtree/cmd/pbtree@latest && \
pbtree init github.com/me/awesome-project && \
pbtree add awesome-service --project=github.com/me/awesome-project && \
pbtree vendor --project=github.com/me/awesome-project --token=<your github api token> && \
go mod init && go mod tidy
```

After this your project will have a similar structure:

```
.
├── .vendorpb
│   ├── api
│   │   └── awesome-service
│   │       └── awesome_service.proto
│   └── google
│       ├── api
│       │   ├── annotations.proto
│       │   └── http.proto
│       └── protobuf
│           └── descriptor.proto
├── api
│   └── awesome-service
│       └── awesome_service.proto
├── bin
│   ├── buf
│   ├── protoc-gen-go
│   ├── protoc-gen-go-grpc
│   ├── protoc-gen-grpc-gateway
│   └── protoc-gen-openapiv2
├── pkg
│   └── awesome-service
│    ├── awesome_service.pb.go
│    ├── awesome_service.pb.gw.go
│    ├── awesome_service.swagger.json
│    └── awesome_service_grpc.pb.go
├── buf.gen.yaml
└── pbtree.yaml
```

# Full documentation

```
Usage:
  pbtree [command]

Available Commands:
  add         Add proto file template
  init        Init pbtree config
  vendor      Vendor and generate proto files

Flags:
  -h, --help   help for pbtree
```

## Init pbtree

`pbtree add new-service` add `buf.gen.yaml`, `pbtree.yaml` and `.gitignore` if not exist.

Flags:
- `--force` add files with override
- `--config` pbtree config filename (default "pbtree.yaml")
- `--vendor-dir` folder for vendoring files (default ".vendorpb")

### pbtree.yaml
Has 2 sections with string array format:
- `local_proto` For project proto files
- `external_proto` For remote proto files

```yaml
local_proto:
    - api/awesome-service/awesome_service.proto
external_proto:
    - google/api/annotations.proto
```

### buf.gen.yaml
File for configuration [buf](https://github.com/bufbuild/buf). [Learn more](https://buf.build/).

## Add new proto service to project

`pbtree add new-service` add new proto template to `/api/new-service/new_service.proto` and to pbtree.yaml config **if
not exist**.

Flags:
- `--force` add file with override
- `--config` pbtree config filename (default "pbtree.yaml")
- `--project` project name


Service template:
```protobuf
syntax = "proto3";

package github.com.me.awesome_project.api.awesome_service.awesome_service;

option go_package = "github.com/me/awesome-project/pkg/awesome-service";

import "google/api/annotations.proto";

service AwesomeService {
  rpc Call(CallRequest) returns (CallResponse) {
    option (google.api.http) = {
      post: "/v1/call"
      body: "*"
    };
  }
  rpc ClientStream(stream ClientStreamRequest) returns (ClientStreamResponse) {}
  rpc ServerStream(ServerStreamRequest) returns (stream ServerStreamResponse) {}
  rpc BidiStream(stream BidiStreamRequest) returns (stream BidiStreamResponse) {}
}

message CallRequest {
  string name = 1;
}

message CallResponse {
  string msg = 1;
}

message ClientStreamRequest {
  int64 stroke = 1;
}

message ClientStreamResponse {
  int64 count = 1;
}

message ServerStreamRequest {
  int64 count = 1;
}

message ServerStreamResponse {
  int64 count = 1;
}

message BidiStreamRequest {
  int64 stroke = 1;
}

message BidiStreamResponse {
  int64 stroke = 1;
}
```

## Vendoring and generate
`pbtree vendor --project=github.com/me/awesome-project --token=<github api token>` vendor and generate proto from pbtree.yaml.

Flags:
- `--buf string `         buf binary directory (default "./bin/buf")
- `--config` pbtree config file (default "pbtree.yaml")
- `--output` folder for generated files (default "internal/pb")
- `--project` project name (require)
- `--template` buf template (default "buf.gen.yaml")
- `--token` github token (without token-key && token available only local provider)
- `--token-key` env key for github token (without token-key && token available only local provider) (default "GITHUB_TOKEN")
- `--vendor-dir` folder for vendoring files (default ".vendorpb")

### Remote files
For vendoring remote files need use `--token` flag with [token](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api?apiVersion=2022-11-28),
or set him at env as GITHUB_TOKEN.
If you need to use a different key, pass it using the `--token-key` flag.

### Buf M flags
<a id="ins"></a>

Use `MFlags` in buf.gen.yaml for configure go imports and destination generate folder

For example, you have:

**buf.gen.yaml**
```yaml
version: v1
plugins:
  - name: go
    path: bin/protoc-gen-go
    out: .
    opt:
      - paths=source_relative
```

**api/awesome-service/awesome_service.proto**
```protobuf
syntax = "proto3";

package github.com.me.awesome_project.api.awesome_service.awesome_service;

option go_package = "github.com/me/awesome-project/pkg/awesome-service";

import "github.com/gosharmot/proto-example/api/example/example.proto";

message Message {
  github.com.gosharmot.example.api.example.Empty empty = 1;
}
```

**pbtree.yaml**
```yaml
local_proto:
    - api/awesome-service/awesome_service.proto
external_proto: []
```

After generation, you will have:
```
.
├── .vendorpb
│   ├── api
│   │   └── awesome-service
│   │       └── awesome_service.proto
│   └── github.com
│       └── gosharmot
│           └── proto-example
│               └── api
│                   └── example
│                       └── example.proto
├── api
│   └── awesome-service
│       └── awesome_service.proto
├── pkg
│   └── awesome-service
│     └── awesome_service.pb.go
├── buf.gen.yaml
└── pbtree.yaml
```

And in `internal/pb/api/awesome-service/awesome_service.pb.go` imports remote package:
```go
package awesome_service

import (
	example "github.com/me/proto-example/pkg/api/example"
	...
)
```

If add `MFlag` option in buf.gen.yaml:
```yaml
version: v1
plugins:
  - name: go
    path: bin/protoc-gen-go
    out: .
    opt:
      - Mgithub.com/gosharmot/proto-example/api/example/example.proto=github.com/me/awesome-project/internal/pb/example
      - paths=source_relative
```

You will have:
```
.
├── .vendorpb
│   ├── api
│   │   └── awesome-service
│   │       └── awesome_service.proto
│   └── github.com
│       └── gosharmot
│           └── proto-example
│               └── api
│                   └── example
│                       └── example.proto
├── api
│   └── awesome-service
│       └── awesome_service.proto
├── internal
│   └── pb
│       ├── example
│       │   └── example.pb.go
├── pkg
│   └── awesome-service
│     └── awesome_service.pb.go
├── buf.gen.yaml
└── pbtree.yaml
```

And in `internal/pb/api/awesome-service/awesome_service.pb.go` imports local package:
```go
package awesome_service

import (
	example "github.com/me/awesome-project/internal/pb/example"
)
```