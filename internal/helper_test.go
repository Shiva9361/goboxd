package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAllowlist(t *testing.T) {
	allowlist := []string{"-O*", "-Wall", "-std=c++11"}
	lookup := buildAllowlist(allowlist)

	assert.Contains(t, lookup.ExactMatches, "-Wall")
	assert.Contains(t, lookup.ExactMatches, "-std=c++11")
	assert.NotContains(t, lookup.ExactMatches, "-O2")
	assert.Equal(t, []string{"-O*"}, lookup.patterns)
}

func TestSanitizeFlags(t *testing.T) {
	allowlist := []string{"-O*", "-Wall", "-std=c++11"}
	lookup := buildAllowlist(allowlist)

	flags := []string{"-O2", "-Wall", "-g", "-std=c++11", "-O3"}
	sanitized := sanitizeFlags(flags, lookup)

	expected := []string{"-O2", "-Wall", "-std=c++11", "-O3"}
	assert.Equal(t, expected, sanitized)
}

func TestPrepArgs(t *testing.T) {
	template := []string{"{{flags}}", "-o", "{{artifact}}", "{{source}}"}
	flags := []string{"-O2", "-Wall"}
	source := "main.cpp"
	artifact := "main"

	args := prepArgs(template, flags, source, artifact)

	expected := []string{"-O2", "-Wall", "-o", "main", "main.cpp"}
	assert.Equal(t, expected, args)
}

func TestClampResource(t *testing.T) {
	assert.Equal(t, 10, clampResource(100, 10))
	assert.Equal(t, 100, clampResource(100, 200))
	assert.Equal(t, 100, clampResource(100, 0))
	assert.Equal(t, 100, clampResource(100, -10))
}

func TestSanitizeFilename(t *testing.T) {
	assert.Equal(t, "main.cpp", sanitizeFilename("main.cpp"))
	assert.Equal(t, "passwd", sanitizeFilename("../../../etc/passwd"))
	assert.Equal(t, "test.sh", sanitizeFilename("/usr/bin/test.sh"))
}

func TestSanitizeLimits(t *testing.T) {
	defaultLimits := Resource{
		WallTime:   5,
		Memory:     1024,
		MaxProcess: 10,
	}
	requestedLimits := Resource{
		WallTime:   2,
		Memory:     2048,
		MaxProcess: 0,
	}

	sanitized := sanitizeLimits(defaultLimits, requestedLimits)

	assert.Equal(t, 2, sanitized.WallTime)
	assert.Equal(t, 1024, sanitized.Memory)
	assert.Equal(t, 10, sanitized.MaxProcess)
}

func TestCappedWriter(t *testing.T) {
	writer := NewCappedWriter(10)
	n, err := writer.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", writer.String())

	n, err = writer.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Contains(t, writer.String(), "hello worl")
	assert.Contains(t, writer.String(), "[Output Truncated: Exceeded Buffer Limit]")
	assert.True(t, writer.truncated)

	n, err = writer.Write([]byte("more"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.NotContains(t, writer.String(), "more")
}
