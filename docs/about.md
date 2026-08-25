---
title: ℹ️ About
layout: default
nav_order: 2
permalink: /about/
---

# About SimpleAuth.link 🛡️

SimpleAuth.link is a secure authentication service that enables your users to log in to your platform without the need for traditional passwords. By using SimpleAuth.link, your platform can verify user identities and manage active sessions through cryptographically generated tokens.

## How It Works ⚙️

### 1. App Registration 📝
Your platform first registers with SimpleAuth.link by creating an **App ID** along with a corresponding **secret**. The App ID is a *self-contained* representation of your application configuration: it deterministically embeds the app name, redirect URL, session duration and a hash of the secret. Nothing is stored on the server.

### 2. Token Request 🔏
When a user enters their email address to log in, your platform sends a request to SimpleAuth.link. The service uses your App ID and secret to generate a secure token using the Ed25519 signature algorithm. This process is deterministic — meaning the private key for signing is generated on the fly from your credentials without being stored.

### 3. Magic Link Delivery ✉️
The generated token is embedded in a secure URL, known as a **magic link**, which is then sent to the user's email. This link includes:

- The token itself, which authenticates the user and embeds its own expiration time.
- The user's email address, as a `user` query parameter.

### 4. User Authentication 👤
When the user clicks the magic link:

- The token is submitted back to your platform.
- Your platform verifies the token's authenticity and validity.
- Upon successful verification, the user is securely logged in.

## Technical Details 💻

- **Stateless Architecture:** SimpleAuth.link operates without a traditional database. It does not store any user data, including email addresses, on its servers. Both App IDs and tokens are self-contained and can be verified with nothing more than the app configuration and secret. This stateless design enhances security and reduces the risk of data breaches.
- **Token Generation Process:** By leveraging the Ed25519 signature algorithm, the service deterministically generates a private key using your App ID and secret. This ensures that each token is cryptographically secure and uniquely tied to your application, eliminating the need to store sensitive keys.
- **Token Structure:** A token is composed of three parts — the expiration time, the signature and the user's email — base64url-encoded and separated by dots. See [About Tokens](/about/tokens) for details.

## Benefits of Using SimpleAuth.link 🎯

- **Enhanced Security:** Eliminates the risks associated with password-based authentication, such as password fatigue, weak passwords, and password reuse. The use of robust cryptographic methods further secures the authentication process.
- **Improved User Experience:** Users can authenticate with a single click using the magic link sent to their email, streamlining the login process and reducing friction.
- **Simplified Backend Management:** As a stateless service, SimpleAuth.link removes the need for complex database management and lowers the potential for data storage vulnerabilities.

## API Reference

The full API is described in the interactive [Swagger UI](https://simpleauthlink.github.io/authapi/), with details about [Apps](/about/apps) and [Tokens](/about/tokens) in the rest of the documentation.

## Magic Links 🔗

Magic links are secure URLs that allow users to perform actions such as logging in without a password. They offer several advantages:

- **Security:** They reduce the risk of unauthorized access by eliminating password vulnerabilities.
- **Convenience:** They simplify the authentication process, requiring just a single click from the user.
- **User Verification:** Sending the link directly to the user's email confirms the validity of the email address and enhances the overall security of the login process.

SimpleAuth.link combines these benefits to deliver a reliable, secure, and user-friendly authentication experience for your platform.