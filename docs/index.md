---
title: Home
nav_icon: home
nav_level: 1
nav_order: 1
nav_collapsed: true
layout: home
---


{: .success }
> **SimpleAuth.link brings passwordless authentication to your users using just an email address.**

![Logo](assets/logo-brand.svg){: style="max-width: 400px; width: 100%; display: block; margin: 0 auto;" }

SimpleAuth.link empowers application developers to **eliminate the complexities of user authentication** and login systems. It streamlines the user experience by replacing traditional passwords with a secure, email-based login mechanism.

## Why Choose SimpleAuth.link?

- **Enhanced Security:** By removing passwords from the equation, you reduce the risks of password theft, reuse, and related vulnerabilities. Instead, authentication is handled via secure, cryptographically generated tokens signed with Ed25519.

- **Improved User Experience:** Users enjoy a simplified login process. With no need to remember or manage passwords, they can quickly and safely access your application using a magic link sent to their email.

- **Efficient Integration:** With only a couple of API requests, you can seamlessly integrate SimpleAuth.link into your application. A [Go client]({{ '/docs/tokens#go-client-examples' | relative_url }}) is also provided to get you started even faster.

- **Privacy & Uniqueness:** SimpleAuth.link stores nothing. Tokens are self-contained and stateless: the user's email is embedded (base64url-encoded) and the whole token is signed, so no personal data is ever stored on SimpleAuth.link servers.

## Getting Started

1. **Create Your App:** Begin by [creating your app]({{ '/dev/api/#/apps/post_apps' | absolute_url }}). This step involves setting up your application with a unique App ID, defining session parameters, and configuring the redirect URL and secret.

2. **Authenticate Your Users:** Use the authentication endpoint to [request a token for your users]({{ '/dev/api/#/tokens/post_tokens' | absolute_url }}). When a user provides their email, SimpleAuth.link will generate a secure token and send it via a magic link, allowing them to log in effortlessly.

3. **Verify the Token:** When the user returns through the magic link, [verify the token]({{ '/dev/api/#/tokens/put_tokens' | absolute_url }}) before opening the session.

## Try the API

The interactive [Swagger UI]({{ '/dev/api/' | absolute_url }}) is the fastest way to explore and test the endpoints. 

[Create a demo app]({{ '/dev/api/#/apps/post_apps' | absolute_url }}){: .btn .btn-success data-feather="plus-square" }

You can also run the API locally, read [Self-host section]({{ '/dev/self-host' | absolute_url }}) for more information.

[Running the Go Code]({{ '/dev/self-host#option-3-running-the-go-code-directly' | absolute_url }}){: .btn data-feather="terminal" }

---

SimpleAuth.link is designed to take the burden of user authentication off your shoulders, allowing you to focus on what matters most delivering an exceptional experience to your users.