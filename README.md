# Guardian: HTTP Session Management for Go

## Introduction

Guardian is a Go package that provides HTTP session management functionality.  
It allows you to create and manage user sessions in your Go web applications. With Guardian, you can easily store and retrieve session data, set session timeouts, and handle session-related tasks.

## Features

-   **In-Memory Session Store**: Guardian comes with an in-memory session store, making it easy to get started. You can also implement custom session stores if needed.

-   **Session Lifecycle Management**: Guardian provides features like session expiration, renewal, and invalidation to ensure session data is secure and up-to-date.

-   **Middleware Integration**: You can easily integrate Guardian into your HTTP handlers using middleware, making session management a seamless part of your application.

-   **Customizable**: Guardian allows you to configure session timeouts, cookie settings, and more to fit your application's needs.

## Installation

Guardian can be installed using Go modules:

```sh
$ go get github.com/1jack80/guardian
```

## Basic Usage

Here's a basic example of how to use Guardian to manage sessions in your Go application:

```go
package main

import (
    "net/http"
    "github.com/1jack80/guardian"
)

func main() {
    // Create a new instance of the Guardian session manager with an in-memory store.
    store := guardian.NewInMemoryStore()
    sessionManager := guardian.NewSessionManager("myapp", store)

    // Define your HTTP handlers.
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Create a new session.
        session := sessionManager.CreateSession()
        session.Data["username"] = "user123"
        sessionManager.Store.Save(&session)

        // Use the session.
        // ...

        w.WriteHeader(http.StatusOK)
    })

    // Wrap your handlers with the Guardian middleware.
    http.Handle("/protected", sessionManager.SessionMiddleware(func(w http.ResponseWriter, r *http.Request) {
        session := sessionManager.GetFromContext(r.Context())
        // Access session data and perform actions.
        // ...

        w.WriteHeader(http.StatusOK)
    }))

    // Start your HTTP server.
    http.ListenAndServe(":8080", nil)
}
```

## Documentation

For detailed documentation and additional features, refer to the [Guardian documentation](https://github.com/1jack80/guardian).

## License

Guardian is licensed under the MIT License. See the [LICENSE](https://github.com/1jack80/guardian/blob/main/LICENSE) file for details.

## Contributing

We welcome contributions! Feel free to open issues or submit pull requests on the [GitHub repository](https://github.com/1jack80/guardian).

## Acknowledgments

-   Guardian was inspired by various Go session management packages most especially that of [Alex Edwards](https://github.com/alexedwards/scs/)
    and the need to advance my go coding skills

-   A special thanks goes to the OWASP and their [informative cheatsheet](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md)
    which helped immensely with providing me with a better understanding
    of how session managers work.

Documentation generated entirely by [chatGPT](chat.openai.com)
