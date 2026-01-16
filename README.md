<h1 align="center">Block3</h1>
<p align="center">B3 is your all-in-one home for Minecraft schematics, heightmaps, and everything your builds need.</p>
<br />

<h2>Table of Contents</h2>

- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Auth Method: Basic Authentication](#auth-method-basic-authentication)
    - [Format](#format)
    - [Permission Levels](#permission-levels)
    - [Password Hash](#password-hash)
  - [Auth Method: OpenID Connect](#auth-method-openid-connect)
    - [Required Scopes](#required-scopes)
    - [Scope Mapping](#scope-mapping)
    - [Supported b3\_group Values](#supported-b3_group-values)
- [License](#license)


## Configuration

### Environment Variables

All variables are read from the environment. Defaults are listed when present.

| Variable | Default | Required | Notes |
| --- | --- | --- | --- |
| `WEB_LISTEN_ADDR` | `:8080` | No | HTTP listen address. |
| `WEB_BASE_URL` | `http://localhost:8080` | No | Public base URL used for redirects and CORS. |
| `WEB_SESSION_KEY` | none (generated) | No | If unset, a random key is generated. Set in production to keep sessions stable. |
| `AUTH_TYPE` | none | Yes | `basic_auth` or `oidc`. Controls which auth variables are required. |
| `AUTH_BASIC_USERS` | none | Yes (basic_auth) | Comma-separated list: `name:permission-level:bcrypt-hash`. |
| `AUTH_OIDC_DISCOVERY_URL` | none | Yes (oidc) | OIDC discovery URL. |
| `AUTH_OIDC_CLIENT_ID` | none | Yes (oidc) | OIDC client ID. |
| `AUTH_OIDC_CLIENT_SECRET` | none | Yes (oidc) | OIDC client secret. |

### Auth Method: Basic Authentication
Basic Authentication expects a list of users to be provided via the `AUTH_BASIC_USERS` environment variable.

See the provided [Docker Compose](/examples/basic_auth/compose.yml) example for a complete setup.

#### Format
Each user must be defined using the following format:
```
<username>:<permission-level>:<bcrypt-hash>
```
Multiple users must be separated by commas:
```
user1:10:$2b$...,user2:5:$2b$...
```

#### Permission Levels
Permission levels are defined numerically as follows:
| Role | Level |
|------|-------|
|VIEWER| 5     |
|EDITOR| 10    |
|ADMIN | 20    |

#### Password Hash
Passwords must be provided as bcrypt hashes.
A bcrypt hash can be generated using the following command:
```shell
b3 hash <password>
```

### Auth Method: OpenID Connect
To enable OpenID Connect (OIDC), the client ID (`AUTH_OIDC_CLIENT_ID`), client secret (`AUTH_OIDC_CLIENT_SECRET`), and discovery URL (`AUTH_OIDC_DISCOVERY_URL`) must be configured.

See the provided [Docker Compose](/examples/oidc/compose.yml) example for a complete setup.

#### Required Scopes
B3 requests the following scopes during authentication:
- `openid`
- `profile`
- `b3`

The `b3` scope is **mandatory**. If it is not granted, the login attempt will be rejected.

#### Scope Mapping
In the identity provider, a scope-to-claim mapping must be configured:
- Scope: `b3`
- Claim: `b3_group`

If the `b3_group` claim is not present in the UserInfo, authentication will fail.

#### Supported b3_group Values
The following values are supported for the `b3_group` claim:
- `viewer`
- `editor`
- `admin`

## License

This project is licensed under the GNU Affero General Public License v3 (AGPLv3).

If you use this software to provide a service over a network, you must make the
complete corresponding source code available to the users of that service,
as required by the AGPL.