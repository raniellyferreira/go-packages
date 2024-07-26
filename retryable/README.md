# retryable

## Introduction

The retryable package provides a robust set of functions designed to simplify the retry logic in your Go applications. Whether you're dealing with transient network issues, intermittent service failures, or any situation where an operation might need to be attempted multiple times before it succeeds, retryable has got you covered.

## Installation

To use the retryable package in your project, execute the following command:

```bash
go get github.com/raniellyferreira/go-packages/retryable
```

## Usage Examples

Here's how you can use retryable to retry a function that might fail due to a transient error:

```go
import "github.com/raniellyferreira/go-packages/retryable"

func mightFailOperation() (int, error) {
    // Your code here that might fail
}

result, err := retryable.MustRetry(mightFailOperation)
if err != nil {
    log.Fatalf("Operation failed after retries: %v", err)
}
```

For more advanced usage with custom retry logic:

```go
import "github.com/raniellyferreira/go-packages/retryable"

func mightFailOperation() (string, error) {
    // Your code here that might fail
}

isRetryable := func(err error) bool {
    // Define your custom logic to determine if an error is retryable
}

result, err := retryable.RetryWithCustomCheck(mightFailOperation, 5, 2*time.Second, isRetryable)
if err != nil {
    log.Fatalf("Operation failed after retries: %v", err)
}
```

## Configuration Options

You can configure the retryable package to suit your needs. Here's an example:

```go
retryable.DefaultMaxAttempts = 5
retryable.DefaultDelay = 2 * time.Second
```

## Logger Configuration

The `retryable` package provides a custom logging feature that allows you to specify how log messages are output. By default, the package uses Go's standard logger, but you can easily customize this to integrate with your own logging infrastructure.

### Using the Default Logger

By default, `retryable` will output log messages to the standard logger provided by Go's `log` package. You don't need to perform any additional configuration to use this default behavior.

### Customizing Log Output

To redirect log messages from `retryable` to Logrus, provide a custom logging function that matches the signature of `log.Printf`. Below is an example using Logrus' `Warnf` method for warning level messages:

```go
package main

import (
 "github.com/sirupsen/logrus"
 "github.com/raniellyferreira/go-packages/retryable"
)

func main() {
 // Initialize a new Logrus logger.
 logger := logrus.New()

 // Configure retryable to use the custom logger.
 retryable.SetLoggerWriter(logger.Warnf)

 // Now, all retry logs will be handled by Logrus with warning level.
}
```

This configuration routes all retry-related log messages through Logrus, allowing you to take advantage of Logrus' structured logging and level-specific logging methods.

Remember to import the Logrus package and configure it according to your application's requirements before setting it as the logger for `retryable`.

## Contributing

Contributions to retryable are welcome! Feel free to fork the repository, make your changes, and submit a pull request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
