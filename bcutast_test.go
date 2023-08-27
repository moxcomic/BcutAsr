package bcutasr

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/ysmood/gson"
)

func TestParse(t *testing.T) {
	res, err := New().Parse("./1.mp3")
	if err != nil {
		panic(err)
	}

	m := make(map[string]any)
	res.Unmarshal(&m)

	fmt.Printf("%v\n", gson.New(m).JSON("", "  "))
}

func TestResult(t *testing.T) {
	data, _ := os.ReadFile("./result.data")

	v := viper.New()
	v.SetConfigType("json")
	err := v.ReadConfig(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
}
