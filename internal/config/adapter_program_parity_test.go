package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"goahk/internal/config"
	"goahk/internal/program"
)

func TestAdapterProgramParityFixtures(t *testing.T) {
	t.Parallel()

	cases := []string{"linear_steps", "flow_reference", "uia_selector"}
	for _, name := range cases {
		name := name
		t.Run(name, func(t *testing.T) {
			cfg, err := config.LoadFile(filepath.Join("testdata", name+".config.json"))
			if err != nil {
				t.Fatalf("LoadFile() error = %v", err)
			}

			got, err := config.ToProgram(cfg)
			if err != nil {
				t.Fatalf("ToProgram() error = %v", err)
			}
			want, err := loadProgramFixture(filepath.Join("testdata", name+".program.json"))
			if err != nil {
				t.Fatalf("loadProgramFixture() error = %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("program mismatch\n got=%#v\nwant=%#v", got, want)
			}
		})
	}
}

func loadProgramFixture(path string) (program.Program, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return program.Program{}, err
	}
	var p program.Program
	if err := json.Unmarshal(data, &p); err != nil {
		return program.Program{}, err
	}
	return program.Normalize(p), nil
}
