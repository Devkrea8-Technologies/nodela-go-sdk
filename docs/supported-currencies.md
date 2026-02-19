# Supported Currencies

The Nodela Go SDK validates the `Currency` field in `CreateInvoiceParams` against a list of **63 supported ISO 4217 currency codes** before making any HTTP request. Passing an unsupported code returns an `SDKError` with code `ErrCodeValidation` immediately — no network call is made.

---

## Currency Validation Behaviour

- Validation is **case-insensitive**: `"usd"`, `"USD"`, and `"Usd"` are all accepted.
- The currency is **normalized to uppercase** before being sent to the API.
- If validation fails, you receive an `*SDKError` — no HTTP request is ever made.

```go
// Valid — all equivalent
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "ngn"})
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "NGN"})
client.Invoices.Create(ctx, models.CreateInvoiceParams{Amount: 10, Currency: "Ngn"})

// Invalid — returns SDKError before any network call
_, err := client.Invoices.Create(ctx, models.CreateInvoiceParams{
    Amount:   10,
    Currency: "XYZ",
})

var sdkErr *nodela_errors.SDKError
if errors.As(err, &sdkErr) {
    // sdkErr.Code    == "VALIDATION"
    // sdkErr.Message == "unsupported currency: XYZ"
}
```

---

## Checking if a Currency is Supported

You can check the `SupportedInvoiceCurrencies` map directly without making an API call:

```go
import "nodela.co/go-sdk/pkg/models"

currency := "GHS"
if _, ok := models.SupportedInvoiceCurrencies[strings.ToUpper(currency)]; ok {
    fmt.Println(currency, "is supported")
} else {
    fmt.Println(currency, "is not supported")
}
```

---

## Currency Reference

### Americas — 10 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `USD` | United States Dollar | United States |
| `CAD` | Canadian Dollar | Canada |
| `MXN` | Mexican Peso | Mexico |
| `BRL` | Brazilian Real | Brazil |
| `ARS` | Argentine Peso | Argentina |
| `CLP` | Chilean Peso | Chile |
| `COP` | Colombian Peso | Colombia |
| `PEN` | Peruvian Sol | Peru |
| `JMD` | Jamaican Dollar | Jamaica |
| `TTD` | Trinidad and Tobago Dollar | Trinidad and Tobago |

### Europe — 16 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `EUR` | Euro | Eurozone |
| `GBP` | British Pound Sterling | United Kingdom |
| `CHF` | Swiss Franc | Switzerland |
| `SEK` | Swedish Krona | Sweden |
| `NOK` | Norwegian Krone | Norway |
| `DKK` | Danish Krone | Denmark |
| `PLN` | Polish Złoty | Poland |
| `CZK` | Czech Koruna | Czech Republic |
| `HUF` | Hungarian Forint | Hungary |
| `RON` | Romanian Leu | Romania |
| `BGN` | Bulgarian Lev | Bulgaria |
| `HRK` | Croatian Kuna | Croatia |
| `ISK` | Icelandic Króna | Iceland |
| `TRY` | Turkish Lira | Turkey |
| `RUB` | Russian Ruble | Russia |
| `UAH` | Ukrainian Hryvnia | Ukraine |

### Africa — 11 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `NGN` | Nigerian Naira | Nigeria |
| `ZAR` | South African Rand | South Africa |
| `KES` | Kenyan Shilling | Kenya |
| `GHS` | Ghanaian Cedi | Ghana |
| `EGP` | Egyptian Pound | Egypt |
| `MAD` | Moroccan Dirham | Morocco |
| `TZS` | Tanzanian Shilling | Tanzania |
| `UGX` | Ugandan Shilling | Uganda |
| `XOF` | West African CFA Franc | West Africa (WAEMU) |
| `XAF` | Central African CFA Franc | Central Africa (CEMAC) |
| `ETB` | Ethiopian Birr | Ethiopia |

### Asia — 15 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `JPY` | Japanese Yen | Japan |
| `CNY` | Chinese Yuan Renminbi | China |
| `INR` | Indian Rupee | India |
| `KRW` | South Korean Won | South Korea |
| `IDR` | Indonesian Rupiah | Indonesia |
| `MYR` | Malaysian Ringgit | Malaysia |
| `THB` | Thai Baht | Thailand |
| `PHP` | Philippine Peso | Philippines |
| `VND` | Vietnamese Dong | Vietnam |
| `SGD` | Singapore Dollar | Singapore |
| `HKD` | Hong Kong Dollar | Hong Kong |
| `TWD` | New Taiwan Dollar | Taiwan |
| `BDT` | Bangladeshi Taka | Bangladesh |
| `PKR` | Pakistani Rupee | Pakistan |
| `LKR` | Sri Lankan Rupee | Sri Lanka |

### Middle East — 8 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `AED` | UAE Dirham | United Arab Emirates |
| `SAR` | Saudi Riyal | Saudi Arabia |
| `QAR` | Qatari Riyal | Qatar |
| `KWD` | Kuwaiti Dinar | Kuwait |
| `BHD` | Bahraini Dinar | Bahrain |
| `OMR` | Omani Rial | Oman |
| `ILS` | Israeli New Shekel | Israel |
| `JOD` | Jordanian Dinar | Jordan |

### Oceania — 3 currencies

| Code | Currency | Country / Region |
| --- | --- | --- |
| `AUD` | Australian Dollar | Australia |
| `NZD` | New Zealand Dollar | New Zealand |
| `FJD` | Fijian Dollar | Fiji |

---

## Full List (Alphabetical)

`AED`, `ARS`, `AUD`, `BDT`, `BGN`, `BHD`, `BRL`, `CAD`, `CHF`, `CLP`, `CNY`, `COP`, `CZK`, `DKK`, `EGP`, `ETB`, `EUR`, `FJD`, `GBP`, `GHS`, `HKD`, `HRK`, `HUF`, `IDR`, `ILS`, `INR`, `ISK`, `JMD`, `JOD`, `JPY`, `KES`, `KRW`, `KWD`, `LKR`, `MAD`, `MXN`, `MYR`, `NGN`, `NOK`, `NZD`, `OMR`, `PEN`, `PHP`, `PKR`, `PLN`, `QAR`, `RON`, `RUB`, `SAR`, `SEK`, `SGD`, `THB`, `TTD`, `TWD`, `TRY`, `TZS`, `UAH`, `UGX`, `USD`, `VND`, `XAF`, `XOF`, `ZAR`

**Total: 63 currencies**

---

## Notes

- Currency availability may be subject to regional restrictions. Contact [Nodela support](https://nodela.co) if you need to accept a currency not on this list.
- Exchange rates are applied automatically when the customer's payment currency differs from your invoice currency.
- The `ExchangeRate` field in `CreateInvoiceData` and `Transaction` records the rate that was applied.

---

## Next Steps

- [Invoices](./invoices.md) — use a currency code when creating invoices.
- [Error Handling](./error-handling.md) — understand the `SDKError` returned for invalid currencies.
