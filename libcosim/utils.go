// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package libcosim

import (
	"log"
	"strconv"
	"strings"
)

func parseFloat(argument string) float64 {
	f, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		log.Fatal(err)
		return 0.0
	}
	return f
}

func strCat(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}
