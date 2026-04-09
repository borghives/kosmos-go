package ether

import (
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var knownEthers []Ether

type Ether interface {
	MergeFromFile(filenames ...string)
	MergeFromCmd(cmd *cobra.Command)
	Coalesce() error
}

type EtherStructure[T any] struct {
	Constants    T
	v            *viper.Viper
	vOnce        sync.Once
	collapseOnce sync.Once
}

func (e *EtherStructure[T]) getViper() *viper.Viper {
	e.vOnce.Do(func() {
		e.v = viper.New()
		LoadUniversalEnv(e.v)
		SetupEnvironmentStructures(e.v, GetConstantsKeyMap(e.Constants))
	})
	return e.v
}

func (e *EtherStructure[T]) Coalesce() error {
	var err error
	e.collapseOnce.Do(func() {
		v := e.getViper()
		err = v.Unmarshal(&e.Constants)
	})
	return err
}

func (e *EtherStructure[T]) Collapse() T {
	if err := e.Coalesce(); err != nil {
		log.Fatalf("Fatal: Failed to Collapse ether: %v", err)
	}
	return e.Constants
}

func (e *EtherStructure[T]) MergeFromFile(filenames ...string) {
	v := e.getViper()
	// 1. look for addition environment file
	for _, filename := range filenames {
		v.SetConfigFile(filename)
		_ = v.MergeInConfig()
	}
}

func (e *EtherStructure[T]) MergeFromCmd(cmd *cobra.Command) {
	v := e.getViper()
	keyMap := GetConstantsKeyMap(e.Constants)
	for key, constantTag := range keyMap {
		if constantTag.CmdFlag != "" {
			if cmdFlag := cmd.Flags().Lookup(constantTag.CmdFlag); cmdFlag != nil {
				v.BindPFlag(key, cmdFlag)
			}
		}
	}
}

type ConstantTag struct {
	CmdFlag  string
	Permeate bool
}

func ParseConstantTag(tag string) ConstantTag {
	if tag == "" {
		return ConstantTag{}
	}

	parts := strings.Split(tag, ",")
	return ConstantTag{
		CmdFlag:  parts[0],
		Permeate: len(parts) > 1 && parts[1] == "permeate",
	}
}

// GetConstantsKeyMap extracts the tags and returns a map,
// for example mapping the cmdflag to the mapstructure key.
func GetConstantsKeyMap[T any](constants T) map[string]ConstantTag {
	// 1. Get the type of the generic T
	targetType := reflect.TypeOf(constants)
	// 2. If it's a pointer, dereference it to get the underlying struct type
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	// 3. Ensure we are actually dealing with a struct
	if targetType.Kind() != reflect.Struct {
		return nil
	}
	keyMap := make(map[string]ConstantTag)
	// 4. Iterate over the struct fields
	for field := range targetType.Fields() {
		// 5. Get the specific tag values
		mapstructureTag := field.Tag.Get("mapstructure")
		cmdflagTag := field.Tag.Get("cmdflag")
		// Do whatever you want with them! For example: Create a mapping
		if mapstructureTag != "" {
			keyMap[mapstructureTag] = ParseConstantTag(cmdflagTag)
		}
	}
	return keyMap
}

func LoadUniversalEnv(v *viper.Viper) {
	// 1. Tell Viper where to look for the universal environment file
	v.AddConfigPath(".")    // Search in the current working directory
	v.SetConfigName(".env") // Look for a file named ".env"
	v.SetConfigType("env")  // Treat the file as a .env format

	// 2. Read the file
	_ = v.ReadInConfig()

}

func SetupEnvironmentStructures(v *viper.Viper, keyMap map[string]ConstantTag) {
	for key := range keyMap {
		v.BindEnv(key)
	}
	v.AutomaticEnv()
}

func RegisterEther(e Ether) {
	knownEthers = append(knownEthers, e)
}

func CollapseKnownEthers(source ...string) error {
	for _, e := range knownEthers {
		e.MergeFromFile(source...)
		if err := e.Coalesce(); err != nil {
			return err
		}
	}
	return nil
}
