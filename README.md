# mxvalidator
Cloud Function to validate MX records

This function concurrently checks the MX records for the given domains to determine if it is valid or not.
It checks for the existence of a non-garbage (localhost, "", "127.0.0.1", "0.0.0.0") MX record.


Deployed at: `https://us-central1-mxvalidator.cloudfunctions.net/ValidateMX`

UI - `https://mxvalidator.tux-sudo.com`

Request:

```javascript
{
  "domains": [
    "domain1.com",
    "gmail.com",
    "example.com"
  ]
}
```

Response:

```javascript
{
  "results": [
    {
      "domain": "gmail.com",
      "valid": true,
      "status": "ValidMX",
      "any_mx": "gmail-smtp-in.l.google.com."
    },
    {
      "domain": "example.com",
      "valid": false,
      "status": "InvalidMX",
      "any_mx": "."
    },
    {
      "domain": "domain1.com",
      "valid": false,
      "status": "Unresolvable"
    }
  ]
}
```
