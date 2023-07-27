package parsers_test

import (
	"testing"

	"github.com/happyagosmith/jig/internal/parsers"
	"github.com/stretchr/testify/assert"
)

func TestConventionalCommit(t *testing.T) {
	t.Run("parse FEATURE type", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(j_AAA-123)!: send an email to the customer when a product is shipped")

		assert.Equal(t, parsers.FEATURE, cc.Type)
		assert.Equal(t, "j_AAA-123", cc.Scope)
	})

	t.Run("parse BUG_FIX type", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("fix(j_AAA-123)!: send an email to the customer when a product is shipped")

		assert.Equal(t, parsers.BUG_FIX, cc.Type)
	})

	t.Run("parse UNKOWN type", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("doc(j_AAA-123)!: send an email to the customer when a product is shipped")

		assert.Equal(t, parsers.UNKNOWN, cc.Type)
	})

	t.Run("parse FEATURE type without scope", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat: send an email to the customer when a product is shipped")

		assert.Equal(t, parsers.FEATURE, cc.Type)
	})

	t.Run("parse BUG_FIX type without scope", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("fix: send an email to the customer when a product is shipped")

		assert.Equal(t, parsers.BUG_FIX, cc.Type)
	})

	t.Run("parse scope", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(api): send an email to the customer when a product is shipped")

		assert.Equal(t, "api", cc.Scope)
	})

	t.Run("parse empty scope", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat: send an email to the customer when a product is shipped")

		assert.Equal(t, "", cc.Scope)
	})

	t.Run("parse subject 1", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat: send an email to the customer when a product is shipped")

		assert.Equal(t, "send an email to the customer when a product is shipped", cc.Subject)
	})

	t.Run("parse subject 2", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(123): send an email to the customer when a product is shipped")

		assert.Equal(t, "send an email to the customer when a product is shipped", cc.Subject)
	})

	t.Run("parse is breaking change 1", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(123)!: send an email to the customer when a product is shipped")

		assert.Equal(t, true, cc.IsBreaking)
	})

	t.Run("parse is breaking change 2", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(123): send an email to the customer when a product is shipped\n BREAKING CHANGE: the details")

		assert.Equal(t, true, cc.IsBreaking)
	})

	t.Run("parse is not breaking change", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("feat(123): send an email to the customer when a product is shipped")

		assert.Equal(t, false, cc.IsBreaking)
	})

	t.Run("parse nil conventional commit", func(t *testing.T) {
		parser := parsers.NewCC()
		cc := parser.Parse("[12345] send an email to the customer when a product is shipped")

		assert.Nil(t, cc)
	})
}
