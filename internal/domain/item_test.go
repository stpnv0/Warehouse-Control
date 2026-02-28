package domain

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestUpdateItemInput_HasChanges(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		input := &UpdateItemInput{}
		assert.False(t, input.HasChanges())
	})

	t.Run("name set", func(t *testing.T) {
		name := "New Name"
		input := &UpdateItemInput{Name: &name}
		assert.True(t, input.HasChanges())
	})

	t.Run("sku set", func(t *testing.T) {
		sku := "NEW-SKU"
		input := &UpdateItemInput{SKU: &sku}
		assert.True(t, input.HasChanges())
	})

	t.Run("quantity set", func(t *testing.T) {
		qty := 10
		input := &UpdateItemInput{Quantity: &qty}
		assert.True(t, input.HasChanges())
	})

	t.Run("price set", func(t *testing.T) {
		price := decimal.NewFromFloat(9.99)
		input := &UpdateItemInput{Price: &price}
		assert.True(t, input.HasChanges())
	})

	t.Run("location set", func(t *testing.T) {
		loc := "Shelf A"
		input := &UpdateItemInput{Location: &loc}
		assert.True(t, input.HasChanges())
	})
}
