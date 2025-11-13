package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfItGetAnErrorIfIDIsBlank(t *testing.T){
	order := Order{}
	assert.Error(t, order.Validate(), "id is required")
	// if order.Validate() == nil{
	// 	t.Error("ID is required")
	// }
}
func TestIfItGetAnErrorIfPriceIsBlank(t *testing.T){
	order := Order{ID: "123"}
	assert.Error(t, order.Validate(), "price must be greater than zero")
}
func TestIfItGetAnErrorIfTaxIsBlank(t *testing.T){
	order := Order{ID: "123", Price: 10.0}
	assert.Error(t, order.Validate(), "invalid tax")
}

func TestFinalPrice(t *testing.T){
	order := Order{ID: "123", Price: 10.0, Tax: 1.0}
	assert.NoError(t, order.Validate())
	assert.Equal(t, "123", order.ID) 
	assert.Equal(t, 10.0, order.Price) 
	assert.Equal(t, 1.0, order.Tax)
	order.CalculateFinalPrice()
	assert.Equal(t, 11.0, order.FinalPrice) 
}