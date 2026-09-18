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

func TestConfigureJira(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	t.Run("returns error when jiraURL is missing", func(t *testing.T) {
		viper.Reset()
		viper.Set(JiraUsername, "user")
		viper.Set(JiraPassword, "pass")

		_, err := ConfigureJira()
		assert.Error(t, err)
	})

	t.Run("returns error when neither bearer token nor username/password are set", func(t *testing.T) {
		viper.Reset()
		viper.Set(JiraURL, "https://jira.example.com")

		_, err := ConfigureJira()
		assert.Error(t, err)
	})

	t.Run("succeeds with username and password", func(t *testing.T) {
		viper.Reset()
		viper.Set(JiraURL, "https://jira.example.com")
		viper.Set(JiraUsername, "user")
		viper.Set(JiraPassword, "pass")

		_, err := ConfigureJira()
		assert.NoError(t, err)
	})

	t.Run("succeeds with bearer token only", func(t *testing.T) {
		viper.Reset()
		viper.Set(JiraURL, "https://jira.example.com")
		viper.Set(JiraBearerToken, "my-bearer-token")

		_, err := ConfigureJira()
		assert.NoError(t, err)
	})
}
