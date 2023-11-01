package traces

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBenchmarkEnv(t *testing.T) {
	b := NewBenchmark("foo")
	os.Setenv(TRACING_DIR_ENV_VAR, "")
	assert.Equal(t, b.Env(), []string{})

	os.Setenv(TRACING_DIR_ENV_VAR, "tracing_dir")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_TAG_BENCHMARK_NAME=foo")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_MEMSTATS_POLL_INTERVAL=100ms")
}

func TestFullBenchmarkEnv(t *testing.T) {
	b := NewBenchmark("bar")
	b.Provider = "aws"
	b.Runtime = "dotnet"
	b.Language = "csharp"
	b.Repository = "pulumi/templates"

	os.Setenv(TRACING_DIR_ENV_VAR, "")
	assert.Equal(t, b.Env(), []string{})

	os.Setenv(TRACING_DIR_ENV_VAR, "tracing_dir")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_TAG_REPO=pulumi/templates")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_TAG_BENCHMARK_NAME=bar")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_TAG_BENCHMARK_RUNTIME=dotnet")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_TAG_BENCHMARK_LANGUAGE=csharp")
	assert.Contains(t, b.Env(), "PULUMI_TRACING_MEMSTATS_POLL_INTERVAL=100ms")
}

func TestOmitMemstats(t *testing.T) {
	b := NewBenchmark("bar")

	os.Setenv(TRACING_DIR_ENV_VAR, "tracing_dir")
	b.MemstatsPollInterval = 0
	assert.NotContains(t, b.Env(), "PULUMI_TRACING_MEMSTATS_POLL_INTERVAL=100ms")
}
