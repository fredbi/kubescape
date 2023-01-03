package cautils

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestFredGit(t *testing.T) {
	repo, err := NewLocalGitRepository("/home/fred/src/github.com/oneconcern/geodude")
	require.NoError(t, err)
	require.NotNil(t, repo)

	commit, err := repo.GetFileLastCommit("Makefile")
	require.NoError(t, err)
	require.NotNil(t, repo)

	spew.Dump(commit)
}
