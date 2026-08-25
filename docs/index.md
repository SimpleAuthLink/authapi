---
title: Home
nav_title: "🏠 Home"
layout: home
nav_order: 1
---

# SimpleAuth.link

**Passwordless authentication for your users using just an email address.**

SimpleAuth.link empowers application developers to **eliminate the complexities of user authentication** and login systems. It streamlines the user experience by replacing traditional passwords with a secure, email-based login mechanism.

## Why Choose SimpleAuth.link?

- **🔐 Enhanced Security:** By removing passwords from the equation, you reduce the risks of password theft, reuse, and related vulnerabilities. Instead, authentication is handled via secure, cryptographically generated tokens signed with Ed25519.

- **🧑‍💻 Improved User Experience:** Users enjoy a simplified login process. With no need to remember or manage passwords, they can quickly and safely access your application using a magic link sent to their email.

- **🔌 Efficient Integration:** With only a couple of API requests, you can seamlessly integrate SimpleAuth.link into your application. A [Go client](/about/tokens#go-client-examples) is also provided to get you started even faster.

- **🛡️ Privacy & Uniqueness:** SimpleAuth.link stores nothing. Tokens are self-contained and stateless: the user's email is embedded (base64url-encoded) and the whole token is signed, so no personal data is ever stored on SimpleAuth.link servers.

## Getting Started

1. **Create Your App:** Begin by [creating your app](https://simpleauthlink.github.io/authapi/#/apps/post). This step involves setting up your application with a unique App ID, defining session parameters, and configuring the redirect URL and secret.

2. **Authenticate Your Users:** Use the authentication endpoint to [request a token for your users](https://simpleauthlink.github.io/authapi/#/tokens/post). When a user provides their email, SimpleAuth.link will generate a secure token and send it via a magic link, allowing them to log in effortlessly.

3. **Verify the Token:** When the user returns through the magic link, [verify the token](https://simpleauthlink.github.io/authapi/#/tokens/put) before opening the session.

## Try the API

The interactive [Swagger UI](https://simpleauthlink.github.io/authapi/) is the fastest way to explore and test the endpoints. You can also run the API locally:

```bash
go run ./cmd/authapi
```

SimpleAuth.link is designed to take the burden of user authentication off your shoulders, allowing you to focus on what matters most—delivering an exceptional experience to your users.