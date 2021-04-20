package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_isValidRetryAfter(t *testing.T) {
	tt := []struct {
		in string
		expected bool
	}{
		{ in: "Mon, 02 Jan 2006 15:04:05 GMT", expected: true },
		{ in: "Mon, 02 Jan 2006 15:04:05 JST", expected: false },
		{ in: "Mon, 02 Jan 2006 15:04:05", expected: false },
		{ in: "2006-01-02T15:04:05Z", expected: false },
	}

	for _, test := range tt {
		result := isValidRetryAfter(test.in)
		if result != test.expected {
			t.Errorf("in: %s, expected: %t", test.in, test.expected)
		}
	}
}

func Test_handler(t *testing.T) {
	const ENV_KEY = "RETRY_AFTER"
	t.Run("環境変数RETRY_AFTERがMon, 02 Jan 2006 15:04:05 GMTのとき、503で終了しRetry-AfterヘッダがMon, 02 Jan 2006 15:04:05 GMT", func(t *testing.T) {
		prev, ok := os.LookupEnv(ENV_KEY)
		const expected = "Mon, 02 Jan 2006 15:04:05 GMT"
		if err := os.Setenv(ENV_KEY, expected); err != nil {
			t.Fatal("os.Setenvに失敗")
		}
		if ok {
			t.Cleanup(func(){ os.Setenv(ENV_KEY, prev) })
		} else {
			t.Cleanup(func(){ os.Unsetenv(ENV_KEY) })
		}

		testHandler := http.HandlerFunc(handler)
		testServer := httptest.NewServer(testHandler)
		t.Cleanup(testServer.Close)

		res, err := http.Get(testServer.URL)
		if err != nil {
			 t.Fatal(err)
		}

		if res.StatusCode != 503 {
			t.Errorf("終了ステータスが503でない, actual: %d", res.StatusCode)
		}

		if actual := res.Header.Get("Retry-After"); actual != expected {
			t.Errorf("Retry-Afterの値が%sでない, actual: %s", expected, actual)
		}
	})
	
	t.Run("環境変数RETRY_AFTERが存在しないとき、503で終了しRetry-Afterヘッダが存在しない", func(t *testing.T) {
		prev, ok := os.LookupEnv(ENV_KEY)
		if ok {
			if err := os.Unsetenv(ENV_KEY); err != nil {
				t.Fatal("os.Unsetenvに失敗")
			}
			t.Cleanup(func(){ os.Setenv(ENV_KEY, prev) })
		}

		testHandler := http.HandlerFunc(handler)
		testServer := httptest.NewServer(testHandler)
		t.Cleanup(testServer.Close)

		res, err := http.Get(testServer.URL)
		if err != nil {
			 t.Fatal(err)
		}

		if res.StatusCode != 503 {
			t.Errorf("終了ステータスが503でない, actual: %d", res.StatusCode)
		}

		if actual := res.Header.Get("Retry-After"); actual != "" {
			t.Errorf("Retry-Afterが存在する, actual: %s", actual)
		}
	})
}