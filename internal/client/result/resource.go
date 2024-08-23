/*
Copyright 2020 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package result

import (
	"cmp"
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/printer"
)

// ResourceAccess holds the access result for all resources.
type ResourceAccess map[string]map[string]Access

// Print implements MatrixPrinter.Print. It prints a tab-separated table with a header.
func (ra ResourceAccess) Table(verbs []string) *printer.Table {
	var names []string
	for name := range ra {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		x, y := names[i], names[j]
		resourcex, groupx, _ := strings.Cut(x, ".")
		resourcey, groupy, _ := strings.Cut(y, ".")
		// first sort by group, then resource
		if groupx == groupy {
			return cmp.Less(resourcex, resourcey)
		}
		return cmp.Less(groupx, groupy)
	})

	// table header
	headers := []printer.Renderable{printer.Text("NAME")}
	for _, v := range verbs {
		headers = append(headers, printer.Text(strings.ToUpper(v)))
	}

	p := printer.TableWithHeaders(headers)

	// table body
	lastGroup := ""
	for _, name := range names {
		// print an empty line if group changed
		resource, group, _ := strings.Cut(name, ".")
		if group != lastGroup {
			p.AddRow(printer.TextList(" "), printer.None) // at least one "none" outcome needed to get the tabprinter aligning all columns
			p.AddRow([]printer.Renderable{printer.BoldText(group + ":")}, printer.None)
			lastGroup = group
		}

		var outcomes []printer.Outcome

		res := ra[name]
		for _, v := range verbs {
			var o printer.Outcome
			switch res[v] {
			case Denied:
				o = printer.Down
			case Allowed:
				o = printer.Up
			case NotApplicable:
				o = printer.None
			case RequestErr:
				o = printer.Err
			}
			outcomes = append(outcomes, o)
		}
		p.AddRow(printer.TextList(resource), outcomes...)
	}
	return p
}
