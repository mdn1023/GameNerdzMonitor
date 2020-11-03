package models

type Proxy struct {
	Status bool
	Host   string
	Un     string
	Pw     string
}

type Product struct {
	Data struct {
		AvailableModifierValues []interface{} `json:"available_modifier_values"`
		Base                    bool          `json:"base"`
		BulkDiscountRates       []interface{} `json:"bulk_discount_rates"`
		Image                   interface{}   `json:"image"`
		InStockAttributes       []interface{} `json:"in_stock_attributes"`
		Instock                 bool          `json:"instock"`
		OutOfStockBehavior      string        `json:"out_of_stock_behavior"`
		OutOfStockMessage       string        `json:"out_of_stock_message"`
		Price                   struct {
			RrpWithoutTax struct {
				Currency  string  `json:"currency"`
				Formatted string  `json:"formatted"`
				Value     float64 `json:"value"`
			} `json:"rrp_without_tax"`
			Saved struct {
				Currency  string  `json:"currency"`
				Formatted string  `json:"formatted"`
				Value     float64 `json:"value"`
			} `json:"saved"`
			TaxLabel   string `json:"tax_label"`
			WithoutTax struct {
				Currency  string  `json:"currency"`
				Formatted string  `json:"formatted"`
				Value     float64 `json:"value"`
			} `json:"without_tax"`
		} `json:"price"`
		Purchasable       bool        `json:"purchasable"`
		PurchasingMessage interface{} `json:"purchasing_message"`
		Sku               string      `json:"sku"`
		Stock             int         `json:"stock"`
		StockMessage      interface{} `json:"stock_message"`
		Upc               interface{} `json:"upc"`
		V3VariantID       int         `json:"v3_variant_id"`
		Weight            interface{} `json:"weight"`
	} `json:"data"`
}

// Message is a json representation of the request body sent to a discord webhook URL
type Message struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

type Embed struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color"`
	Thumbnail   URL    `json:"thumbnail"`
	Fields      []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"fields,omitempty"`
}

type URL struct {
	URL string `json:"url"`
}
