// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package languages

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPackageSuffix(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		suffix PackageSuffix
		want   string
	}{
		{
			name:   "contains suffix for tar.gz",
			input:  "test.tar.gz",
			suffix: Tar,
			want:   "test.tar.gz",
		},
		{
			name:   "matches suffix for tar.gz",
			input:  "test",
			suffix: Tar,
			want:   "test.tar.gz",
		},
		{
			name:   "contains suffix for zip",
			input:  "test.zip",
			suffix: Zip,
			want:   "test.zip",
		},
		{
			name:   "matches suffix for zip",
			input:  "test",
			suffix: Zip,
			want:   "test.zip",
		},
		{
			name:   "contains suffix for pkg",
			input:  "test.pkg",
			suffix: Pkg,
			want:   "test.pkg",
		},
		{
			name:   "matches suffix for pkg",
			input:  "test",
			suffix: Pkg,
			want:   "test.pkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.suffix.Composition(tt.input)
			require.Equal(t, tt.want, actual)
		})
	}
}

func TestAllSuffix(t *testing.T) {
	packageName := fmt.Sprintf("test%s", Tar.String())
	require.True(t, true, AllSuffix.Has(packageName))

	noSuffix := "test"
	require.False(t, false, AllSuffix.Has(noSuffix))

	require.Equal(t, "test", AllSuffix.Trim(packageName))
	require.Equal(t, "test", AllSuffix.Trim(noSuffix))

	//no suffix defined
	illegalSuffix := "test.xxxx"
	require.Equal(t, illegalSuffix, AllSuffix.Trim(illegalSuffix))

	zipPackage := fmt.Sprintf("test%s", Zip)
	require.Equal(t, Zip.String(), AllSuffix.GetSuffix(zipPackage).String())

	require.Equal(t, "", AllSuffix.GetSuffix(illegalSuffix).String())
}
