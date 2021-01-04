package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spf13/viper"
)

func TestSimpleConfig(t *testing.T) {
	assert := assert.New(t)

	file, err := os.Open(filepath.FromSlash("test_data/simple_conf.json"))
	if err != nil {
		t.Fatal(err)
	}
	viper.SetConfigType("json")
	err = viper.ReadConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(2, len(config.Directories))

	dir1 := config.Directories[0]
	dir2 := config.Directories[1]

	assert.Equal("/home/deep/work/aoc", dir1.Local)
	assert.Equal("AOC", dir1.Remote)
	assert.True(dir1.Recursive)

	assert.Equal("/home/deep/.config", dir2.Local)
	assert.Equal("config", dir2.Remote)
	assert.False(dir2.Recursive)
}
