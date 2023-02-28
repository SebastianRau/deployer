package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	filetools "github.com/sebastianRau/deployer/pkg/fileTools"
	"github.com/sebastianRau/deployer/pkg/steps"
)

var (
	version = "1.0.0 abcdef"
)

func main() {
	var (
		configTemplateFile = flag.String("config.file", "./test/kw_bem_make.json.tpl", "Config File") //toDo remove default
		configDataFile     = flag.String("config.data", "./test/kw_bem_make.data.json", "Data file")
		configOutputFile   = flag.String("config.output", "", "Output json file")
		verbose            = flag.Bool("v", false, "verbose output")
	)
	flag.Parse()

	var (
		err error
	)

	if *configTemplateFile == "" {
		panic("No Config File given")
	}

	// read and unmarshal template
	st, err := steps.UnmarshalConfigTemplate(*configTemplateFile, *configDataFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *configOutputFile != "" {
		err := filetools.ParseTemplate(*configTemplateFile, *configDataFile, *configOutputFile, os.Stdout)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	fmt.Printf("Deployer v%s\n", version)

	start := time.Now()
	err = st.Exceute(os.Stdout, *verbose)
	if err != nil {
		fmt.Println(err.Error())
	}
	stop := time.Now()
	elapsed := stop.Sub(start)

	fmt.Printf("Time: %s\n", elapsed.Round(time.Millisecond))

}
