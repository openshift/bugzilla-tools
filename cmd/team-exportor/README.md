If you ever get an error like:

```
Error: Get "https://sheets.googleapis.com/v4/spreadsheets/[XXXXXX]/values/%27OCP%20Team%20Structure%27%21B%3AC?alt=json&prettyPrint=false": oauth2: cannot fetch token: 400 Bad Request
Response: {
  "error": "invalid_grant",
  "error_description": "Token has been expired or revoked."
}
Usage:
```

You likely need to delete the token.json and run it again. This will give you a link to follow, follow the link, approve, get the new token, past it onto the cli, voila!

I don't know why, or how often, google expires them.
