package retryable_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/raniellyferreira/go-retryable"
)

func TestMustRetrySuccess(t *testing.T) {
	var attempt int
	fn := func() (bool, error) {
		attempt++
		if attempt < 2 {
			return false, errors.New("temporary error")
		}
		return true, nil
	}

	expected := true
	result, err := retryable.MustRetry(fn)
	if err != nil || result != expected {
		t.Errorf("Expected %v, got %v with error %v", expected, result, err)
	}
}

func TestMustRetryFail(t *testing.T) {
	fn := func() (bool, error) {
		return false, errors.New("permanent error")
	}

	expectedAttempts := retryable.DefaultMaxAttempts
	for i := 0; i < expectedAttempts; i++ {
		_, err := retryable.MustRetry(fn)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	}
}

func TestRetryWithCustomCheck(t *testing.T) {
	var attempt int
	fn := func() (int, error) {
		attempt++
		if attempt < 3 {
			return 0, errors.New("temporary error")
		}
		return attempt, nil
	}

	isRetryable := func(err error) bool {
		return err.Error() == "temporary error"
	}

	expected := 3
	result, err := retryable.RetryWithCustomCheck(fn, 5, 1*time.Millisecond, isRetryable)
	if err != nil || result != expected {
		t.Errorf("Expected %v, got %v with error %v", expected, result, err)
	}
}

func TestRetryWithNonRetryableErrors(t *testing.T) {
	nonRetryableErrors := []string{"fatal", "permanent"}
	fn := func() (bool, error) {
		return false, errors.New("fatal error")
	}

	result, err := retryable.RetryWithNonRetryableErrors(fn, 3, 1*time.Millisecond, nonRetryableErrors)
	if err == nil || err.Error() != "fatal error" {
		t.Errorf("Expected non-retryable error, got %v", err)
	}
	if result {
		t.Errorf("Expected result to be false")
	}
}

func TestRetryWithRetryableErrors(t *testing.T) {
	retryableErrors := []string{"timeout", "temporary"}
	var attempt int
	fn := func() (bool, error) {
		attempt++
		if attempt < 3 {
			return false, errors.New("timeout error")
		}
		return true, nil
	}

	result, err := retryable.RetryWithRetryableErrors(fn, 3, 1*time.Millisecond, retryableErrors)
	if err != nil || !result {
		t.Errorf("Expected successful retry on retryable error, got %v with error %v", result, err)
	}
}

func TestRetryAlways(t *testing.T) {
	var attempt int
	fn := func() (int, error) {
		attempt++
		if attempt < 4 {
			return 0, errors.New("error")
		}
		return attempt, nil
	}

	expected := 4
	result, err := retryable.RetryAlways(fn, 5, 1*time.Millisecond)
	if err != nil || result != expected {
		t.Errorf("Expected result %d after retries, got %d with error %v", expected, result, err)
	}
}

func TestContainsError(t *testing.T) {
	retryableErrors := []string{"temporary", "intermittent"}
	err := errors.New("this is a temporary issue")

	if !retryable.ContainsError(err, retryableErrors) {
		t.Errorf("Expected error to be considered retryable")
	}
}

// Teste para a função ContainsError com erro não retentável
func TestContainsErrorNonRetryable(t *testing.T) {
	retryableErrors := []string{"temporary", "intermittent"}
	err := errors.New("this is a fatal issue")

	if retryable.ContainsError(err, retryableErrors) {
		t.Errorf("Expected error to not be considered retryable")
	}
}

// TestRetryWithCustomCheckNonRetryableError tests the scenario where the error should not be retried.
func TestRetryWithCustomCheckNonRetryableError(t *testing.T) {
	fn := func() (bool, error) {
		return false, errors.New("non-retryable error")
	}

	isRetryable := func(err error) bool {
		// A função customizada deve retornar false para o erro "non-retryable error",
		// indicando que o erro não é retentável.
		return err.Error() != "non-retryable error"
	}

	// Chamamos RetryWithCustomCheck com maxAttempts e delay específicos,
	// juntamente com a função isRetryable customizada.
	result, err := retryable.RetryWithCustomCheck(fn, 5, 1*time.Millisecond, isRetryable)

	// O teste espera que o erro seja retornado imediatamente sem múltiplas tentativas,
	// já que o erro não é retentável segundo a função isRetryable.
	if err == nil || err.Error() != "non-retryable error" {
		t.Errorf("Expected non-retryable error, got %v", err)
	}
	if result {
		t.Errorf("Expected result to be false due to non-retryable error")
	}
}

// TestRetryAlwaysFailure tests the RetryAlways function where it always fails.
func TestRetryAlwaysFailure(t *testing.T) {
	fn := func() (bool, error) {
		return false, errors.New("error")
	}

	_, err := retryable.RetryAlways(fn, 3, 1*time.Millisecond)
	if err == nil {
		t.Errorf("Expected an error after maximum attempts, got nil")
	}
}

// TestSetLoggerWriter tests setting a custom logger.
func TestSetLoggerWriter(t *testing.T) {
	var logOutput string
	testLogger := func(format string, args ...interface{}) {
		logOutput = fmt.Sprintf(format, args...)
	}

	retryable.SetLoggerWriter(testLogger)
	fn := func() (bool, error) {
		return false, errors.New("error")
	}

	retryable.MustRetry(fn)

	if !strings.Contains(logOutput, "Retrying in") {
		t.Errorf("Expected log output to contain 'Retrying in', got %s", logOutput)
	}
}

// TestMustRetryCustomDelayAndAttempts tests MustRetry with custom delay and attempts.
func TestMustRetryCustomDelayAndAttempts(t *testing.T) {
	oldMaxAttempts := retryable.DefaultMaxAttempts
	oldDelay := retryable.DefaultDelay
	defer func() {
		retryable.DefaultMaxAttempts = oldMaxAttempts
		retryable.DefaultDelay = oldDelay
	}()

	retryable.DefaultMaxAttempts = 5
	retryable.DefaultDelay = 10 * time.Millisecond

	var attempt int
	fn := func() (bool, error) {
		attempt++
		if attempt < retryable.DefaultMaxAttempts {
			return false, errors.New("temporary error")
		}
		return true, nil
	}

	result, err := retryable.MustRetry(fn)
	if err != nil || !result {
		t.Errorf("Expected true result with no error, got %v with error %v", result, err)
	}
}

// TestMustRetryWhenFirstAttemptSucceeds tests the MustRetry function when the operation succeeds on the first attempt.
func TestMustRetryWhenFirstAttemptSucceeds(t *testing.T) {
	fn := func() (bool, error) {
		return true, nil // Simula uma operação bem-sucedida na primeira tentativa
	}

	result, err := retryable.MustRetry(fn)
	if err != nil || result != true {
		t.Errorf("Operation should succeed on first attempt, got result: %v, error: %v", result, err)
	}
}

// TestMustRetryWithCustomCheckWhenFirstAttemptSucceeds tests the MustRetryWithCustomCheck function when the operation succeeds on the first attempt.
func TestMustRetryWithCustomCheckWhenFirstAttemptSucceeds(t *testing.T) {
	fn := func() (bool, error) {
		return true, nil // Simula uma operação bem-sucedida na primeira tentativa
	}
	isRetryable := func(err error) bool {
		return true // Todos os erros são considerados retentáveis para este teste
	}

	result, err := retryable.MustRetryWithCustomCheck(fn, isRetryable)
	if err != nil || result != true {
		t.Errorf("Operation should succeed on first attempt with custom check, got result: %v, error: %v", result, err)
	}
}

// TestRetryWithCustomCheckMaxAttemptsReached tests the RetryWithCustomCheck function when it reaches the maximum number of attempts without succeeding.
func TestRetryWithCustomCheckMaxAttemptsReached(t *testing.T) {
	attempts := 0
	fn := func() (int, error) {
		attempts++
		return 0, errors.New("error") // Simula uma operação que sempre falha
	}
	isRetryable := func(err error) bool {
		return true // Todos os erros são considerados retentáveis para este teste
	}

	retryable.DefaultMaxAttempts = 3 // Define um número baixo de tentativas máximas para este teste
	_, err := retryable.RetryWithCustomCheck(fn, 3, 1*time.Millisecond, isRetryable)
	if err == nil || attempts != retryable.DefaultMaxAttempts {
		t.Errorf("Expected to reach max attempts without success, got attempts: %d, error: %v", attempts, err)
	}
}

// TestRetryWithNonRetryableErrorsWhenErrorIsRetryable tests the RetryWithNonRetryableErrors function when the error is considered retryable.
func TestRetryWithNonRetryableErrorsWhenErrorIsRetryable(t *testing.T) {
	nonRetryableErrors := []string{"fatal"} // Define uma lista de erros não retentáveis
	attempts := 0
	fn := func() (int, error) {
		attempts++
		if attempts < 3 {
			return 0, errors.New("temporary error") // Simula um erro retentável
		}
		return attempts, nil
	}

	_, err := retryable.RetryWithNonRetryableErrors(fn, 3, 1*time.Millisecond, nonRetryableErrors)
	if err != nil || attempts != 3 {
		t.Errorf("Expected to succeed after retries with retryable error, got attempts: %d, error: %v", attempts, err)
	}
}

// TestMustRetryWithCustomCheckNonRetryableError tests MustRetryWithCustomCheck with a non-retryable error.
func TestMustRetryWithCustomCheckNonRetryableError(t *testing.T) {
	var attempts int
	fn := func() (bool, error) {
		attempts++
		return false, errors.New("non-retryable error")
	}

	isRetryable := func(err error) bool {
		// Define o erro como não retentável
		return err.Error() != "non-retryable error"
	}

	_, err := retryable.MustRetryWithCustomCheck(fn, isRetryable)
	if err == nil || err.Error() != "non-retryable error" || attempts > 1 {
		t.Errorf("Expected non-retryable error without retry, got %v and attempts %d", err, attempts)
	}
}

// TestMustRetryWithCustomCheckMaxAttemptsExceeded tests MustRetryWithCustomCheck when max attempts are exceeded.
func TestMustRetryWithCustomCheckMaxAttemptsExceeded(t *testing.T) {
	attempts := 0
	fn := func() (bool, error) {
		attempts++
		return false, errors.New("retryable error")
	}

	isRetryable := func(err error) bool {
		// Define todos os erros como retentáveis
		return true
	}

	retryable.DefaultMaxAttempts = 3 // Configura um número específico de tentativas máximas

	_, err := retryable.MustRetryWithCustomCheck(fn, isRetryable)
	if err == nil || attempts != retryable.DefaultMaxAttempts {
		t.Errorf("Expected max attempts to be reached, got %v and attempts %d", err, attempts)
	}
}

// TestRetryWithNonRetryableErrorsMaxAttemptsExceeded tests RetryWithNonRetryableErrors when max attempts are exceeded.
func TestRetryWithNonRetryableErrorsMaxAttemptsExceeded(t *testing.T) {
	nonRetryableErrors := []string{"fatal"}
	attempts := 0
	fn := func() (bool, error) {
		attempts++
		return false, errors.New("temporary error")
	}

	_, err := retryable.RetryWithNonRetryableErrors(fn, 3, 1*time.Millisecond, nonRetryableErrors)
	if err == nil || attempts != 3 {
		t.Errorf("Expected to reach max attempts with temporary error, got attempts: %d, error: %v", attempts, err)
	}
}

// TestRetryWithRetryableErrorsRetryableError tests RetryWithRetryableErrors with a retryable error.
func TestRetryWithRetryableErrorsRetryableError(t *testing.T) {
	retryableErrors := []string{"retryable"}
	attempts := 0
	fn := func() (bool, error) {
		attempts++
		if attempts == 3 {
			return true, nil // Succeed on the third attempt
		}
		return false, errors.New("retryable")
	}

	_, err := retryable.RetryWithRetryableErrors(fn, 5, 1*time.Millisecond, retryableErrors)
	if err != nil || attempts != 3 {
		t.Errorf("Expected to succeed after retrying with a retryable error, got attempts: %d, error: %v", attempts, err)
	}
}

// TestRetryWithRetryableErrorsNonRetryableError tests RetryWithRetryableErrors with a non-retryable error.
func TestRetryWithRetryableErrorsNonRetryableError(t *testing.T) {
	retryableErrors := []string{"is-retryable"}
	attempts := 0
	fn := func() (bool, error) {
		attempts++
		return false, errors.New("non-retryable")
	}

	_, err := retryable.RetryWithRetryableErrors(fn, 3, 1*time.Millisecond, retryableErrors)
	if attempts != 1 {
		t.Errorf("Expected to stop after one attempt on a non-retryable error, got %d attempts", attempts)
	}
	if err == nil || !strings.Contains(err.Error(), "non-retryable") {
		t.Errorf("Expected a non-retryable error, got %v", err)
	}
}

// TestRetryWithRetryableErrorsMaxAttemptsReached tests RetryWithRetryableErrors when the maximum number of attempts is reached.
func TestRetryWithRetryableErrorsMaxAttemptsReached(t *testing.T) {
	retryableErrors := []string{"retryable"}
	attempts := 0
	maxAttempts := 3
	fn := func() (bool, error) {
		attempts++
		return false, errors.New("retryable") // Um erro diferente no final para simular a falha final
	}

	_, err := retryable.RetryWithRetryableErrors(fn, maxAttempts, 1*time.Millisecond, retryableErrors)
	if err == nil || attempts != maxAttempts {
		t.Errorf("Expected to reach max attempts with retryable errors, got %v, attempts: %d", err, attempts)
	}
}
