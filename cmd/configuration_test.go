package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIssuePatternsFromConfigFile(t *testing.T) {
	// Reset viper for clean state
	viper.Reset()
	viper.SetConfigFile("testdata/config.yaml")

	err := viper.ReadInConfig()
	require.NoError(t, err)

	// Get the patterns
	patterns := GetIssuePatterns()

	// Verify the patterns were loaded correctly
	require.NotNil(t, patterns, "patterns should not be nil")
	assert.Len(t, patterns, 3, "should have 3 patterns")

	if len(patterns) >= 3 {
		assert.Equal(t, "silk", patterns[0].IssueTracker)
		assert.Equal(t, `SILK-\d+|silk-\d+`, patterns[0].Pattern)

		assert.Equal(t, "jira", patterns[1].IssueTracker)
		assert.Equal(t, `[A-Z]+-\d+`, patterns[1].Pattern)

		assert.Equal(t, "git", patterns[2].IssueTracker)
		assert.Equal(t, `#(\d+)`, patterns[2].Pattern)
	}
}
