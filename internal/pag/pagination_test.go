package pag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		p := New()
		assert.Equal(t, int32(0), p.Offset())
		assert.Equal(t, DefaultPerPage, int(p.Limit()))
		assert.Equal(t, DefaultCurrentPage, int(p.Page()))
	})

	t.Run("custom options", func(t *testing.T) {
		p := New(func(o *Options) {
			o.PerPage = 20
			o.CurrentPage = 2
		})
		assert.Equal(t, int32(20), p.Offset())
		assert.Equal(t, int32(20), p.Limit())
		assert.Equal(t, int32(2), p.Page())
	})

	t.Run("incorrect options", func(t *testing.T) {
		p := New(func(o *Options) {
			o.PerPage = -1
			o.CurrentPage = -1
		})
		assert.Equal(t, int32(0), p.Offset())
		assert.Equal(t, DefaultPerPage, int(p.Limit()))
		assert.Equal(t, DefaultCurrentPage, int(p.Page()))
	})
}

func TestPagination_Getters(t *testing.T) {
	p := New(func(o *Options) {
		o.PerPage = 15
		o.CurrentPage = 3
	})

	t.Run("Offset()", func(t *testing.T) {
		assert.Equal(t, int32(30), p.Offset())
	})

	t.Run("Limit()", func(t *testing.T) {
		assert.Equal(t, int32(15), p.Limit())
	})

	t.Run("Page()", func(t *testing.T) {
		assert.Equal(t, int32(3), p.Page())
	})

	t.Run("PerPage()", func(t *testing.T) {
		assert.Equal(t, int32(15), p.PerPage())
	})
}
