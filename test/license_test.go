// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package test

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/stretchr/testify/assert"
)

type dependencyAlert struct {
	dependency string
	license    string
	confidence float32
}

var dependencyLicenseOnce sync.Once

func TestDependencyLicenses(t *testing.T) {
	// No need to re-run it on `-count=N` where N>1
	dependencyLicenseOnce.Do(func() {
		// Get dependencies
		file, err := ioutil.ReadFile("../vendor/modules.txt")
		assert.Nil(t, err)
		modules := string(file)
		scanner := bufio.NewScanner(strings.NewReader(modules))
		dependencies := make([]string, 0)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "# ") {
				// found a dependency
				dependency := strings.Split(line, " ")[1]
				dependencies = append(dependencies, dependency)

			}
		}

		// Start scanning them
		var wg sync.WaitGroup
		alerts := make(chan dependencyAlert, len(dependencies))
		for _, dependency := range dependencies {
			wg.Add(1)
			go func(dependency string) {
				defer wg.Done()
				f, err := filer.FromDirectory(fmt.Sprintf("../vendor/%s", dependency))
				assert.Nil(t, err)
				licenses, err := licensedb.Detect(f)
				assert.Nil(t, err)
				if len(licenses) == 0 {
					alerts <- dependencyAlert{
						dependency: dependency,
						license:    "none",
						confidence: 1,
					}
					return
				}
				for license := range licenses {
					if strings.Contains(license, "GPL") {
						alerts <- dependencyAlert{
							dependency: dependency,
							license:    license,
							confidence: licenses[license].Confidence,
						}
					}
					return
				}
			}(dependency)

		}
		// Wait until the scan is done
		wg.Wait()
	loop:
		for {
			select {
			case alert := <-alerts:
				t.Fatalf("%s indicates %s license with %v confidence", alert.dependency, alert.license, alert.confidence)
			default:
				break loop
			}
		}
	})
}
