package apiMaintenance

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

func setUpEnv(t *testing.T, key string, value string) {
	prev, ok := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal("os.Setenvに失敗")
	}
	if ok {
		t.Cleanup(func(){ os.Setenv(key, prev) })
	} else {
		t.Cleanup(func(){ os.Unsetenv(key) })
	}
}

func Test_handler(t *testing.T) {
	const ENV_KEY = "RETRY_AFTER"
	const SUCCESS_ENV_VALUE = "Mon, 02 Jan 2006 15:04:05 GMT"
	t.Run("環境変数RETRY_AFTERがMon, 02 Jan 2006 15:04:05 GMTのとき、503で終了しRetry-AfterヘッダがMon, 02 Jan 2006 15:04:05 GMT", func(t *testing.T) {
		const expected = SUCCESS_ENV_VALUE
		setUpEnv(t, ENV_KEY, expected)

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

	t.Run("メソッドがOPTIONSの場合、200で終了しRetry-Afterヘッダが存在しない", func(t *testing.T) {
		setUpEnv(t, ENV_KEY, SUCCESS_ENV_VALUE)
		testHandler := http.HandlerFunc(handler)
		testServer := httptest.NewServer(testHandler)
		t.Cleanup(testServer.Close)
		req, _ := http.NewRequest("OPTIONS", testServer.URL, nil)
		res := httptest.NewRecorder()
		testServer.Config.Handler.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Errorf("終了ステータスが200でない, actual: %d", res.Code)
		}

		if actual := res.Header().Get("Retry-After"); actual != "" {
			t.Errorf("Retry-Afterが存在する, actual: %s", actual)
		}
	})

	t.Run("CORS対応を確認", func(t *testing.T) {
		testHandler := http.HandlerFunc(handler)
		testServer := httptest.NewServer(testHandler)
		t.Cleanup(testServer.Close)
		
		res, err := http.Get(testServer.URL)
		if err != nil {
			t.Fatal(err)
		}

		if expected, actual := "*", res.Header.Get("Access-Control-Allow-Headers"); actual != expected {
			t.Errorf("Access-Control-Allow-Headersが%sでない, actual: %s", expected, actual)
		}
		if expected, actual := "*", res.Header.Get("Access-Control-Allow-Origin"); actual != expected {
			t.Errorf("Access-Control-Allow-Originが%sでない, actual: %s", expected, actual)
		}
		if expected, actual := "GET, POST, PUT, DELETE, OPTIONS", res.Header.Get("Access-Control-Allow-Methods"); actual != expected {
			t.Errorf("Access-Control-Allow-Methodsが%sでない, actual: %s", expected, actual)
		}
	})
}