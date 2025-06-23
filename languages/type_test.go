package languages

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
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

	require.Equal(t, "test", AllSuffix.Trim(noSuffix))

	//no suffix defined
	illegalSuffix := "test.xxxx"
	require.Equal(t, illegalSuffix, AllSuffix.Trim(illegalSuffix))

	zipPackage := fmt.Sprintf("test%s", Zip)
	require.Equal(t, Zip.String(), AllSuffix.GetSuffix(zipPackage).String())

	require.Equal(t, "", AllSuffix.GetSuffix(illegalSuffix).String())
}
