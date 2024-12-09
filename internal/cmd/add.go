package cmd

import (
	"fmt"
	t "html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosharmot/clog"
	"github.com/gosharmot/pbtree/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const protoTemplate = `syntax = "proto3";

package {{ .Package }};

option go_package = "{{ .GoPackage }}";

import "google/api/annotations.proto";

service {{ .Service }} {
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
}`

var Add = &cobra.Command{
	Use:        "add",
	Args:       cobra.ExactArgs(1),
	Example:    "add test-service",
	ArgAliases: []string{"service-name"},
	Short:      "Add proto file template",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	RunE: addF,
}

func init() {
	Add.Flags().StringVar(&projectRepo, "project", "", "project name")
	Add.Flags().BoolVar(&force, "force", false, "force create")
	Add.Flags().StringVar(&configFile, "config", "pbtree.yaml", "pbtree config file")
	_ = Add.MarkFlagRequired("project")
}

func addF(_ *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get wd: %w", err)
	}

	service := args[0]

	protoPath := filepath.Join("api", service, strings.ReplaceAll(service, "-", "_")+".proto")
	filename := filepath.Join(wd, protoPath)

	_, err = os.Stat(filename)
	if err == nil && !force {
		clog.Warning("service already exists")
		return nil
	}

	if dir := filepath.Dir(filename); dir != "" {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir: %s", err)
		}
	}
	create, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	tmpl, err := t.New("proto").Parse(protoTemplate)
	if err != nil {
		return fmt.Errorf("parse proto template: %w", err)
	}

	err = tmpl.Execute(create, struct {
		Package   string
		GoPackage string
		Service   string
	}{
		Package:   strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%s/api/%s", projectRepo, service), "/", "."), "-", "_"),
		GoPackage: fmt.Sprintf("%s/pkg/%s", projectRepo, service),
		Service:   strings.ReplaceAll(cases.Title(language.English).String(service), "-", ""),
	})
	if err != nil {
		return fmt.Errorf("execute template: %s", err)
	}

	configFileName := filepath.Join(wd, configFile)
	cfg, err := config.Parse(configFileName)
	if err != nil {
		return fmt.Errorf("parse config: %s", err)
	}

	var exist bool
	for _, proto := range cfg.LocalProto {
		if proto == protoPath {
			exist = true
		}
	}
	if !exist {
		cfg.LocalProto = append(cfg.LocalProto, protoPath)
	}
	file, err := os.OpenFile(configFileName, os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open config file: %s", err)
	}
	defer func() { _ = file.Close() }()

	marshal, err := cfg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal config: %s", err)
	}

	if _, err = file.Write(marshal); err != nil {
		return fmt.Errorf("write config: %s", err)
	}

	return nil
}
