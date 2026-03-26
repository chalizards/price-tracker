package gemini

import "fmt"

const systemInstruction = "You are a price extraction API. You MUST respond with ONLY a JSON object, no other text."

func buildPriceExtractionPrompt(productName string, pageText string) string {
	return fmt.Sprintf(`Find the current selling prices of the product "%s" in the page text below.
Brazilian e-commerce sites usually show two prices:
1. PIX/cash price (discounted) - labeled as "à vista", "no PIX", "no boleto", or similar
2. Credit card price (full price) - the base price or labeled as "no cartão", "parcelado", or the higher price

Brazilian prices use comma as decimal separator (e.g. "R$ 4.999,00" = 4999.00).
If only one price is visible, return it as "pix".
If no prices are found, return {"prices": []}.

Respond ONLY with a JSON object in this exact format, no other text:
{"prices": [{"price": 1234.56, "currency": "BRL", "payment_type": "pix"}, {"price": 1399.90, "currency": "BRL", "payment_type": "credit"}]}

Page text:
%s`, productName, pageText)
}
