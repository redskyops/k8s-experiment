/*
Copyright 2021 GramLabs, Inc.

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

package generation

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// AsScaledInt scales the quantity to the appropriate scale, honoring the base
// as determined by the format of the supplied quantity. For example,
// AsScaledInt(NewQuantity(1, BinarySI), Milli) returns 1024.
func AsScaledInt(q resource.Quantity, scale resource.Scale) int32 {
	// ScaledValue works fine for base 10
	if q.Format != resource.BinarySI {
		return int32(q.ScaledValue(scale))
	}

	v := q.Value()
	if scale > 0 {
		for e := int(scale) / 3; e > 0; e-- {
			v /= 1024
		}
	} else {
		for e := int(scale) / 3; e < 0; e++ {
			v *= 1024
		}
	}
	return int32(v)
}

var (
	binarySuffix  = []string{"", "", "", "", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}
	decimalSuffix = []string{"n", "u", "m", "", "k", "M", "G", "T", "P", "E"}
)

// QuantitySuffix returns the suffix for a quantity or an empty string if it is
// known. Note that although scale is just an int, you should only use the
// predefined constants (or variables populated from them) when calling.
func QuantitySuffix(scale resource.Scale, format resource.Format) string {
	index := func(suffixes []string, scale resource.Scale) string {
		// Adjust the Scale (Nano=-9 ... Exa=18) to match the index in the suffix arrays
		i := int(scale-resource.Nano) / 3
		// Suffixes only exist if the scale is divisible by 3 (i.e. powers of 1000)
		if int(scale)%3 != 0 || i < 0 || i >= len(suffixes) {
			return ""
		}
		return suffixes[i]
	}

	switch format {
	case resource.BinarySI:
		return index(binarySuffix, scale)
	case resource.DecimalSI:
		return index(decimalSuffix, scale)
	default:
		return ""
	}
}
