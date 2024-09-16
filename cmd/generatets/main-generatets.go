// Copyright 2024, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/wavetermdev/waveterm/pkg/service"
	"github.com/wavetermdev/waveterm/pkg/tsgen"
	"github.com/wavetermdev/waveterm/pkg/util/utilfn"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
)

func generateTypesFile(tsTypesMap map[reflect.Type]string) error {
	fd, err := os.Create("frontend/types/gotypes.d.ts")
	if err != nil {
		return err
	}
	defer fd.Close()
	fmt.Fprintf(os.Stderr, "generating types file to %s\n", fd.Name())
	tsgen.GenerateWaveObjTypes(tsTypesMap)
	err = tsgen.GenerateServiceTypes(tsTypesMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating service types: %v\n", err)
		os.Exit(1)
	}
	err = tsgen.GenerateWshServerTypes(tsTypesMap)
	if err != nil {
		return fmt.Errorf("error generating wsh server types: %w", err)
	}
	fmt.Fprintf(fd, "// Copyright 2024, Command Line Inc.\n")
	fmt.Fprintf(fd, "// SPDX-License-Identifier: Apache-2.0\n\n")
	fmt.Fprintf(fd, "// generated by cmd/generate/main-generatets.go\n\n")
	fmt.Fprintf(fd, "declare global {\n\n")
	var keys []reflect.Type
	for key := range tsTypesMap {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		iname, _ := tsgen.TypeToTSType(keys[i], tsTypesMap)
		jname, _ := tsgen.TypeToTSType(keys[j], tsTypesMap)
		return iname < jname
	})
	for _, key := range keys {
		// don't output generic types
		if strings.Index(key.Name(), "[") != -1 {
			continue
		}
		tsCode := tsTypesMap[key]
		istr := utilfn.IndentString("    ", tsCode)
		fmt.Fprint(fd, istr)
	}
	fmt.Fprintf(fd, "}\n\n")
	fmt.Fprintf(fd, "export {}\n")
	return nil
}

func generateServicesFile(tsTypesMap map[reflect.Type]string) error {
	fd, err := os.Create("frontend/app/store/services.ts")
	if err != nil {
		return err
	}
	defer fd.Close()
	fmt.Fprintf(os.Stderr, "generating services file to %s\n", fd.Name())
	fmt.Fprintf(fd, "// Copyright 2024, Command Line Inc.\n")
	fmt.Fprintf(fd, "// SPDX-License-Identifier: Apache-2.0\n\n")
	fmt.Fprintf(fd, "// generated by cmd/generate/main-generatets.go\n\n")
	fmt.Fprintf(fd, "import * as WOS from \"./wos\";\n\n")
	orderedKeys := utilfn.GetOrderedMapKeys(service.ServiceMap)
	for _, serviceName := range orderedKeys {
		serviceObj := service.ServiceMap[serviceName]
		svcStr := tsgen.GenerateServiceClass(serviceName, serviceObj, tsTypesMap)
		fmt.Fprint(fd, svcStr)
		fmt.Fprint(fd, "\n")
	}
	return nil
}

func generateWshClientApiFile(tsTypeMap map[reflect.Type]string) error {
	fd, err := os.Create("frontend/app/store/wshclientapi.ts")
	if err != nil {
		return err
	}
	defer fd.Close()
	declMap := wshrpc.GenerateWshCommandDeclMap()
	fmt.Fprintf(os.Stderr, "generating wshclientapi file to %s\n", fd.Name())
	fmt.Fprintf(fd, "// Copyright 2024, Command Line Inc.\n")
	fmt.Fprintf(fd, "// SPDX-License-Identifier: Apache-2.0\n\n")
	fmt.Fprintf(fd, "// generated by cmd/generate/main-generatets.go\n\n")
	fmt.Fprintf(fd, "import { WshClient } from \"./wshclient\";\n\n")
	orderedKeys := utilfn.GetOrderedMapKeys(declMap)
	fmt.Fprintf(fd, "// WshServerCommandToDeclMap\n")
	fmt.Fprintf(fd, "class RpcApiType {\n")
	for _, methodDecl := range orderedKeys {
		methodDecl := declMap[methodDecl]
		methodStr := tsgen.GenerateWshClientApiMethod(methodDecl, tsTypeMap)
		fmt.Fprint(fd, methodStr)
		fmt.Fprintf(fd, "\n")
	}
	fmt.Fprintf(fd, "}\n\n")
	fmt.Fprintf(fd, "export const RpcApi = new RpcApiType();\n")
	return nil
}

func main() {
	err := service.ValidateServiceMap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating service map: %v\n", err)
		os.Exit(1)
	}
	tsTypesMap := make(map[reflect.Type]string)
	err = generateTypesFile(tsTypesMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating types file: %v\n", err)
		os.Exit(1)
	}
	err = generateServicesFile(tsTypesMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating services file: %v\n", err)
		os.Exit(1)
	}
	err = generateWshClientApiFile(tsTypesMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating wshserver file: %v\n", err)
		os.Exit(1)
	}
}
