package retryable

import (
	"log"
	"strings"
	"time"
)

var (
	// DefaultMaxAttempts is the default maximum number of attempts for the retry operations.
	DefaultMaxAttempts int = 3

	// DefaultDelay is the default time to wait before retrying an operation.
	DefaultDelay time.Duration = 1 * time.Second

	// logPrintf is the default log output function to handle formatted log messages.
	logPrintf func(string, ...interface{}) = log.Printf
)

// SetLoggerWriter sets a custom log output function to handle formatted log messages.
// writer: Function with signature matching log.Printf to output log messages.
func SetLoggerWriter(writer func(string, ...interface{})) {
	logPrintf = writer
}

// MustRetry executes a function until it succeeds or the maximum number of attempts is reached.
// It uses the global variables DefaultMaxAttempts and DefaultDelay for the retry configuration.
func MustRetry[T any](fn func() (T, error)) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= DefaultMaxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		logPrintf("Attempt %d/%d failed: %v. Retrying in %v...\n", attempt, DefaultMaxAttempts, err, DefaultDelay)
		time.Sleep(DefaultDelay)
	}
	return result, err // Return the last error encountered
}

// MustRetryWithCustomCheck executes a function until it succeeds, the maximum number of attempts is reached,
// or the provided custom check function returns false indicating that the error is not retryable.
func MustRetryWithCustomCheck[T any](fn func() (T, error), isRetryable func(error) bool) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= DefaultMaxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// Use the provided function to decide if we should retry.
		if !isRetryable(err) {
			return result, err // Do not retry if the error is not retryable.
		}

		log.Printf("Attempt %d/%d failed with an error: %v. Retrying in %v...\n", attempt, DefaultMaxAttempts, err, DefaultDelay)
		time.Sleep(DefaultDelay)
	}
	return result, err // Return the last error encountered.
}

// RetryWithCustomCheck provides a flexible retry mechanism, allowing custom logic to determine retryable errors.
// It retries a specified function with controlled delays and a user-defined check for whether to continue.
func RetryWithCustomCheck[T any](fn func() (T, error), maxAttempts int, delay time.Duration, isRetryable func(error) bool) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// Use the provided function to decide if we should retry.
		if !isRetryable(err) {
			return result, err // Return immediately if the error is not retryable.
		}

		log.Printf("Attempt %d/%d failed with an error: %v. Retrying in %v...\n", attempt, maxAttempts, err, delay)
		time.Sleep(delay)
	}
	return result, err // Last error encountered.
}

// RetryWithNonRetryableErrors gracefully handles retry logic for functions that may fail with retryable errors.
// It supports custom delays and distinguishes between errors that should halt retries.
func RetryWithNonRetryableErrors[T any](fn func() (T, error), maxAttempts int, delay time.Duration, nonRetryableErrors []string) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// Check if the error is non-retryable.
		if ContainsError(err, nonRetryableErrors) {
			return result, err // Return immediately on a non-retryable error.
		}

		log.Printf("Attempt %d/%d failed with an error: %v. Retrying in %v...\n", attempt, maxAttempts, err, delay)
		time.Sleep(delay)
	}
	return result, err // Last error encountered.
}

// RetryWithRetryableErrors executes a function until it succeeds, the maximum number of attempts is reached,
// or a non-retryable error is encountered.
func RetryWithRetryableErrors[T any](fn func() (T, error), maxAttempts int, delay time.Duration, retryableErrors []string) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// Checks if the error is retryable.
		if !ContainsError(err, retryableErrors) {
			return result, err
		}

		logPrintf("Attempt %d/%d failed with a retryable error: %v. Retrying in %v...\n", attempt, maxAttempts, err, delay)
		time.Sleep(delay)
	}
	return result, err // Return the last error encountered.
}

// RetryAlways attempts to execute the provided function up to a maximum number of times, pausing with a delay between each try, regardless of the error type.
// It's a relentless retry strategy that stops only when a success is achieved or the maxAttempts are exhausted.
func RetryAlways[T any](fn func() (T, error), maxAttempts int, delay time.Duration) (T, error) {
	var result T
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		logPrintf("Attempt %d/%d failed: %v. Retrying in %v...\n", attempt, maxAttempts, err, delay)
		time.Sleep(delay)
	}
	return result, err // Return the last error encountered
}

// ContainsError checks if the error message contains any of the substrings
// in the list of errors allowed for retrying.
func ContainsError(err error, retryableErrors []string) bool {
	for _, retryableError := range retryableErrors {
		if strings.Contains(err.Error(), retryableError) {
			return true
		}
	}
	return false
}
