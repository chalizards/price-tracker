package gemini

import "fmt"

func buildPriceExtractionPrompt(productName string, html string) string {
	return fmt.Sprintf(`Analyze this HTML page and find the current total price of the product "%s".
Respond ONLY with a JSON object in this exact format, no other text:
{"price": 1234.56, "currency": "BRL"}

HTML:
%s`, productName, html)
}
