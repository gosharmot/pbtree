package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var errTest = errors.New("test error")

func Test_Vendor(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T)
		check     func(t *testing.T)
		wantErr   bool
		stringErr string
	}{
		{
			name: "only local without MFlag",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}
				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.1.yaml")
				require.NoError(t, err)
			},
			check: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)
				_, err = os.Stat(filepath.Join(wd, "testdata/only-local/internal/pb/api/test/test.pb.go"))
				require.NoError(t, err)
				_, err = os.Stat(filepath.Join(wd, "testdata/only-local/.vendorpb/api/test/test.proto"))
				require.NoError(t, err)

				err = os.RemoveAll(filepath.Join(wd, "testdata/only-local/internal"))
				require.NoError(t, err)
				err = os.RemoveAll(filepath.Join(wd, "testdata/only-local/.vendorpb"))
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "only local with MFlag",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.2.yaml")
				require.NoError(t, err)
			},
			check: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)
				_, err = os.Stat(filepath.Join(wd, "testdata/only-local/internal/pb/api/test/2/test.pb.go"))
				require.NoError(t, err)
				_, err = os.Stat(filepath.Join(wd, "testdata/only-local/.vendorpb/api/test/test.proto"))
				require.NoError(t, err)

				err = os.RemoveAll(filepath.Join(wd, "testdata/only-local/internal"))
				require.NoError(t, err)
				err = os.RemoveAll(filepath.Join(wd, "testdata/only-local/.vendorpb"))
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "wd err",
			setup: func(t *testing.T) {
				wdFunc = func() (string, error) {
					return "", errTest
				}
				err := Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "get wd: test error",
		},
		{
			name: "parse buf template: buf not found",
			setup: func(t *testing.T) {
				wdFunc = func() (string, error) {
					return "", nil
				}
				err := Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "parse buf template: file 'buf.gen.yaml' not found",
		},
		{
			name: "parse buf template: invalid buf",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.5.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "parse buf template: yaml:",
		},
		{
			name: "invalid MFlag",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.3.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "parse mFlags: invalid option",
		},
		{
			name: "invalid MFlag dst",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.4.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "migrate: invalid destination",
		},
		{
			name: "pbtree not found",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.1.yaml")
				require.NoError(t, err)
				err = Vendor.Flags().Set("config", "pb.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "parse config: file 'pb.yaml' not found. run pbtree init",
		},
		{
			name: "empty pbtree",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.1.yaml")
				require.NoError(t, err)
				err = Vendor.Flags().Set("config", "pbtree.empty.yaml")
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "inappropriate fetcher",
			setup: func(t *testing.T) {
				wd, err := os.Getwd()
				require.NoError(t, err)

				wdFunc = func() (string, error) {
					return filepath.Join(wd, "testdata/only-local"), nil
				}

				err = Vendor.Flags().Set("buf", "buf")
				require.NoError(t, err)
				err = Vendor.Flags().Set("project", "only-local")
				require.NoError(t, err)
				err = Vendor.Flags().Set("template", "buf.gen.1.yaml")
				require.NoError(t, err)
				err = Vendor.Flags().Set("config", "pbtree.2.yaml")
				require.NoError(t, err)
			},
			wantErr:   true,
			stringErr: "vendor proto: google/protobuf/wrappers.proto: inappropriate fetcher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			if err := Vendor.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Vendor.Execute() error = %v", err)
			} else if tt.wantErr {
				require.ErrorContains(t, err, tt.stringErr)
			}
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}
