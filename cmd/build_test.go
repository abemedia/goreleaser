package cmd

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	setup(t)
	var cmd = newBuildCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2", "--deprecated"})
	require.NoError(t, cmd.cmd.Execute())
}

func TestBuildInvalidConfig(t *testing.T) {
	setup(t)
	createFile(t, "goreleaser.yml", "foo: bar")
	var cmd = newBuildCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2", "--deprecated"})
	require.EqualError(t, cmd.cmd.Execute(), "yaml: unmarshal errors:\n  line 1: field foo not found in type config.Project")
}

func TestBuildBrokenProject(t *testing.T) {
	setup(t)
	createFile(t, "main.go", "not a valid go file")
	var cmd = newBuildCmd()
	cmd.cmd.SetArgs([]string{"--snapshot", "--timeout=1m", "--parallelism=2"})
	require.EqualError(t, cmd.cmd.Execute(), "failed to parse dir: .: main.go:1:1: expected 'package', found not")
}

func TestBuildFlags(t *testing.T) {
	var setup = func(opts buildOpts) *context.Context {
		return setupBuildContext(context.New(config.Project{}), opts)
	}

	t.Run("snapshot", func(t *testing.T) {
		var ctx = setup(buildOpts{
			snapshot: true,
		})
		require.True(t, ctx.Snapshot)
		require.True(t, ctx.SkipValidate)
		require.True(t, ctx.SkipTokenCheck)
	})

	t.Run("snapshot auto clean", func(t *testing.T) {
		testlib.Mktmp(t)
		testlib.GitInit(t)
		testlib.GitRemoteAdd(t, "git@github.com:goreleaser/goreleaser.git")

		var ctx = setup(buildOpts{
			snapshotAuto: true,
		})

		require.False(t, ctx.Snapshot)
		require.False(t, ctx.SkipValidate)
	})

	t.Run("snapshot auto dirty", func(t *testing.T) {
		var folder = testlib.Mktmp(t)
		testlib.GitInit(t)
		testlib.GitRemoteAdd(t, "git@github.com:foo/bar.git")
		testlib.GitAdd(t)
		testlib.GitCommit(t, "whatever")
		testlib.GitTag(t, "v0.0.1")
		require.NoError(t, ioutil.WriteFile(filepath.Join(folder, "foo"), []byte("foobar"), 0644))

		var ctx = setup(buildOpts{
			snapshotAuto: true,
		})

		require.True(t, ctx.Snapshot)
		require.True(t, ctx.SkipValidate)
		require.True(t, ctx.SkipTokenCheck)
	})

	t.Run("skips", func(t *testing.T) {
		var ctx = setup(buildOpts{
			skipValidate:  true,
			skipPostHooks: true,
		})
		require.True(t, ctx.SkipValidate)
		require.True(t, ctx.SkipPostBuildHooks)
		require.True(t, ctx.SkipTokenCheck)
	})

	t.Run("parallelism", func(t *testing.T) {
		require.Equal(t, 1, setup(buildOpts{
			parallelism: 1,
		}).Parallelism)
	})

	t.Run("rm dist", func(t *testing.T) {
		require.True(t, setup(buildOpts{
			rmDist: true,
		}).RmDist)
	})
}
