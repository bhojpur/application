# Bhojpur Application - Client-side Identity & Access Management

It is a client-side [Bhojpur IAM](https://github.com/bhojpur/iam) software development kit, which would allow you to easily connect your application to the [Bhojpur IAM](https://github.com/bhojpur/iam) authentication system without having to implement it from scratch.

The [Bhojpur IAM](https://github.com/bhojpur/iam) SDK is very simple to use. We will show you the steps below.

> Notice that this SDK has been applied to several [Bhojpur.NET Platform](https://github.com/bhojpur/platform) applications and/or services, if you still don’t know how to use it after reading README.md, you can refer to it

## Step 1. Init Config

Initialization requires five parameters, which are all string type:

| Name (in order)  | Must | Description                                             |
| ---------------- | ---- | ------------------------------------------------------- |
| endpoint         | Yes  | Bhojpur IAM Server URL, such as `http://localhost:8000` |
| clientId         | Yes  | Application.client_id                                   |
| clientSecret     | Yes  | Application.client_secret                               |
| jwtSecret        | Yes  | Same as Bhojpur IAM JWT secret                          |
| organizationName | Yes  | Application.organization                                |

```go
func InitConfig(endpoint string, clientId string, clientSecret string, jwtSecret string, organizationName string)
```

## Step 2. Get token and parse

After the [Bhojpur IAM](https://github.com/bhojpur/iam) verification passed, it will be redirected to your application with code and state, like `http://erp.bhojpur.net?code=xxx&state=yyyy`.

Your web application can get the `code`,`state`, and call `GetOAuthToken(code, state)`, then parse out `JWT` token.

The general process is as follows:

```go
token, err := auth.GetOAuthToken(code, state)
if err != nil {
	panic(err)
}

claims, err := auth.ParseJwtToken(token.AccessToken)
if err != nil {
	panic(err)
}

claims.AccessToken = token.AccessToken
```

## Step 3. Set Session in your application

`auth.Claims` contains the basic information about the user provided by [Bhojpur IAM](https://github.com/bhojpur/iam), you can use it as a keyword to set the session in your application, like this:

```go
data, _ := json.Marshal(claims)
c.setSession("user", data)
```

## Step 4. Interact with the users

The [Bhojpur IAM](https://github.com/bhojpur/iam) SDK supports basic user operations, like:

- `GetUser(name string)`, get a user by user name.
- `GetUsers()`, get all users.
- `UpdateUser(auth.User)/AddUser(auth.User)/DeleteUser(auth.User)`, write user to the database.
