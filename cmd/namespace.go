// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"

	"github.com/qri-io/dataset"
	"github.com/qri-io/fs"
	"github.com/qri-io/namespace"
	lns "github.com/qri-io/namespace/local"
	"github.com/spf13/cobra"
)

// namespaceCmd represents the namespace command
var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "List & edit namespaces",
	Long:  `Namespaces are a domain connected with a base address.`,
}

var nsList = &cobra.Command{
	Use:   "list",
	Short: "List namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		adr := GetAddress(cmd, args)

		adrs, err := namespace.ReadAllAddresses(GetNamespaces(cmd, args).ChildAddresses(adr))
		if err != nil {
			ErrExit(err)
		}

		for _, a := range adrs {
			fmt.Println(a.String())
		}
	},
}

var nsAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a namespace",
	Run: func(cmd *cobra.Command, args []string) {
		PrintNotYetFinished(cmd)
	},
}

var nsRemove = &cobra.Command{
	Use:   "remove",
	Short: "Remove a namespace",
	Run: func(cmd *cobra.Command, args []string) {
		PrintNotYetFinished(cmd)
	},
}

func init() {
	namespaceCmd.AddCommand(nsList)
	namespaceCmd.AddCommand(nsAdd)
	namespaceCmd.AddCommand(nsRemove)
	RootCmd.AddCommand(namespaceCmd)
}

// Namespaces is a collection of namespaces that also satisfies the namespace interface
// by querying each namespace in order
type Namespaces []namespace.Namespace

func (n Namespaces) Url() string {
	str := ""
	for _, ns := range n {
		str += ns.Url() + "\n"
	}
	return str
}
func (n Namespaces) Base() dataset.Address {
	// str := ""
	// for _, ns := range n {
	// 	str += ns.Base().String() + "\n"
	// }
	return dataset.NewAddress("")
}
func (n Namespaces) String() string {
	str := ""
	for _, ns := range n {
		str += ns.String() + "\n"
	}
	return str
}

func (n Namespaces) ChildAddresses(adr dataset.Address) (namespace.Addresses, error) {
	for _, ns := range n {
		if ds, err := ns.ChildAddresses(adr); err == nil {
			return ds, nil
		}
	}
	return nil, namespace.ErrNotFound
}

func (n Namespaces) ChildDatasets(adr dataset.Address) (namespace.Datasets, error) {
	for _, ns := range n {
		if ds, err := ns.ChildDatasets(adr); err == nil {
			return ds, nil
		}
	}
	return nil, namespace.ErrNotFound
}

func (n Namespaces) Dataset(adr dataset.Address) (*dataset.Dataset, error) {
	for _, ns := range n {
		if ds, err := ns.Dataset(adr); err == nil {
			return ds, nil
		}
	}
	return nil, namespace.ErrNotFound
}

func (n Namespaces) Package(adr dataset.Address) (io.ReaderAt, int64, error) {
	for _, ns := range n {
		if ds, size, err := ns.Package(adr); err == nil {
			return ds, size, nil
		}
	}

	return nil, 0, namespace.ErrNotFound
}

func (n Namespaces) Store(adr dataset.Address) (fs.Store, error) {
	for _, ns := range n {
		if _, err := ns.Dataset(adr); err == nil {
			// if the base is local, we can just hand back the local store
			if lcl, ok := ns.(*lns.Namespace); ok {
				return lcl.Store(adr)
			}

			// otherwise we need to download the dataset to our local store
			store, err := downloadPackage(ns, adr, adr.String())
			if err != nil {
				return nil, err
			}
			return store, nil
		}
	}

	return nil, namespace.ErrNotFound
}

func (ns Namespaces) Search(query string) ([]*dataset.Dataset, error) {
	found := false
	results := make([]*dataset.Dataset, 0)

	if len(ns) == 0 {
		return nil, fmt.Errorf("no namespaces available for search!")
	}

	for _, n := range ns {
		if s, ok := n.(namespace.SearchableNamespace); ok {
			found = true
			ds, err := namespace.ReadAllDatasets(s.Search(query, -1, 0))
			if err != nil {
				return results, err
			}
			results = append(results, ds...)
		}
	}

	if !found {
		return nil, fmt.Errorf("none of your namespaces are searchable!")
	}

	return results, nil
}