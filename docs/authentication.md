# Authentication

Every request to the Nodela API must include a valid API key. The SDK handles authentication automatically once you provide the key at client initialization — you never need to set the `Authorization` header yourself.

---

## How It Works

The SDK attaches the following header to every outbound HTTP request:

```
Authorization: Bearer <your-api-key>
```

This happens inside the core `doRequest` function in the transport layer. No request can bypass this mechanism.

---

## Obtaining an API Key

1. Log in to your Nodela dashboard at [nodela.co](https://nodela.co).
2. Navigate to **Settings → API Keys**.
3. Click **Create API Key** and give it a descriptive label (e.g., `production-backend`).
4. Copy the key immediately — it is shown only once.

---

## Providing the API Key to the SDK

Pass the key as the first argument to `NewClient`:

```go
import nodela "nodela.co/go-sdk/pkg/client"

client := nodela.NewClient("ndk_live_xxxxxxxxxxxxxxxxxxxx")
```

### Recommended: environment variable

Hard-coding credentials in source code is a security risk. Load the key from an environment variable instead:

```go
import (
    "os"
    nodela "nodela.co/go-sdk/pkg/client"
)

client := nodela.NewClient(os.Getenv("NODELA_API_KEY"))
```

Set the environment variable before running your application:

```bash
# local development
export NODELA_API_KEY="ndk_live_xxxxxxxxxxxxxxxxxxxx"

# or using a .env file with a tool like godotenv
```

### Using a secrets manager

For production workloads, retrieve the key from a secrets manager at startup:

```go
// Example using AWS Secrets Manager
import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
    nodela "nodela.co/go-sdk/pkg/client"
)

func buildNodelaClient(ctx context.Context, sm *secretsmanager.Client) (*nodela.Client, error) {
    out, err := sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String("prod/nodela/api-key"),
    })
    if err != nil {
        return nil, err
    }

    var secret struct {
        APIKey string `json:"apiKey"`
    }
    if err := json.Unmarshal([]byte(*out.SecretString), &secret); err != nil {
        return nil, err
    }

    return nodela.NewClient(secret.APIKey), nil
}
```

---

## What Happens with an Invalid Key

If the API key is missing, malformed, or revoked, the Nodela API returns HTTP **401 Unauthorized**. The SDK wraps this as an `*APIError` with status code `401` and error code `ErrCodeAuthentication`.

```go
import (
    "errors"
    nodela_errors "nodela.co/go-sdk/pkg/errors"
)

_, err := client.Invoices.Create(ctx, params)
if err != nil {
    if nodela_errors.IsAuthenticatedError(err) {
        // API key is invalid or missing
        log.Fatal("Check your NODELA_API_KEY environment variable")
    }
}
```

---

## Key Security Best Practices

| Practice | Reason |
| --- | --- |
| Load from environment variables or secrets managers | Prevents credentials appearing in source code or version history |
| Never commit `.env` files containing real keys | Even private repos can be compromised |
| Use separate keys for development and production | Limits blast radius of a leaked development key |
| Rotate keys immediately if compromised | Revoke the old key from the Nodela dashboard first, then update your secret |
| Grant least privilege | If Nodela supports scoped keys, use the minimum scope needed |

---

## User-Agent Header

In addition to the `Authorization` header, the SDK sends a `User-Agent` header on every request to help Nodela support identify the client version:

```
User-Agent: nodela-go-sdk/1.0.1
```

This is set automatically and cannot be overridden.

---

## Next Steps

- [Configuration](./configuration.md) — customize timeouts and provide a custom HTTP client.
- [Error Handling](./error-handling.md) — handle `401` and other error types correctly.
