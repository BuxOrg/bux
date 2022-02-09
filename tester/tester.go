/*
Package tester is a generic testing package with helpful methods for all packages
*/
package tester

import (
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/require"
)

// RandomTablePrefix will make a random prefix (avoid same tables for parallel tests)
func RandomTablePrefix(t *testing.T) string {
	prefix, err := utils.RandomHex(8)
	require.NoError(t, err)
	// add an underscore just in case the table name starts with a number, this is not allowed in sqlite
	return "_" + prefix
}
