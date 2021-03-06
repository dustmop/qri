package cmd

import (
	"fmt"
	"github.com/qri-io/jsonschema"
	"github.com/qri-io/qri/repo"
	"os"
	"path/filepath"

	// "github.com/ipfs/go-datastore"
	// "github.com/qri-io/dataset"
	// "github.com/qri-io/dataset/dsfs"
	"github.com/qri-io/qri/core"
	"github.com/spf13/cobra"
)

var (
	validateDsFilepath       string
	validateDsSchemaFilepath string
	validateDsURL            string
	validateDsPassive        bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "show schema validation errors",
	Long: `
validate checks data for errors using a structure, printing a list of issues.
By default validate checks dataset data against it’s own structure. validate is 
a flexible command that works with data and structures either inside our outside 
of qri by providing one or both of --data and --structure arguments. 

Providing --structure and --data is an “external validation that uses nothing 
stored in qri. When only one of structure or data args are provided, the other 
comes from a dataset reference. For example, to check how a file “data.csv” 
validates against a dataset "foo”, we would run:
	qri validate —data data.csv foo
In this case, qri will will print any validation as if data.csv was foo’s data.

To see how changes to a structure “structure.json” will validate against a 
dataset in qri, we would run:
	qri validate —structure structure.json foo
In this case, qri will print and validation errors as if stucture.json was the
structure for dataset foo

Using validate this way is a great way to see how changes to data or structure
will affect a dataset before saving changes to a dataset.`,
	Example: `  show errors in an existing dataset:
  $ qri validate b5/comics`,
	PreRun: func(cmd *cobra.Command, args []string) {
		loadConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var (
			dataFile, schemaFile *os.File
			err                  error
			ref                  repo.DatasetRef
		)

		if len(args) == 1 {
			ref, err = repo.ParseDatasetRef(args[0])
			ExitIfErr(err)
		}

		if ref.IsEmpty() && !(validateDsFilepath != "" && validateDsSchemaFilepath != "") {
			ErrExit(fmt.Errorf("please provide a dataset name to validate, or both  --file and --schema arguments"))
		}

		dataFile, err = loadFileIfPath(validateDsFilepath)
		ExitIfErr(err)
		schemaFile, err = loadFileIfPath(validateDsSchemaFilepath)
		ExitIfErr(err)

		req, err := datasetRequests(false)
		ExitIfErr(err)

		p := &core.ValidateDatasetParams{
			Ref: ref,
			// URL:          addDsURL,
			DataFilename: filepath.Base(validateDsSchemaFilepath),
		}

		// this is because passing nil to interfaces is bad
		// see: https://golang.org/doc/faq#nil_error
		if dataFile != nil {
			p.Data = dataFile
		}
		if schemaFile != nil {
			p.Schema = schemaFile
		}

		res := []jsonschema.ValError{}
		err = req.Validate(p, &res)
		ExitIfErr(err)
		if len(res) == 0 {
			printSuccess("✔ All good!")
			return
		}

		for i, err := range res {
			fmt.Printf("%d: %s\n", i, err.Error())
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateDsURL, "url", "u", "", "url to file to initialize from")
	validateCmd.Flags().StringVarP(&validateDsFilepath, "file", "f", "", "data file to initialize from")
	validateCmd.Flags().StringVarP(&validateDsSchemaFilepath, "schema", "", "", "json schema file to use for validation")
	validateCmd.Flags().BoolVarP(&validateDsPassive, "passive", "p", false, "disable interactive init")
	RootCmd.AddCommand(validateCmd)
}
