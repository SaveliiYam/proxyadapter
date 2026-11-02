package upstream

// Auth — исходящая авторизация к upstream API.
type Auth struct {
	Type          string `json:"type"` // basic | bearer | without
	BasicLogin    string `json:"basicLogin"`
	BasicPassword string `json:"basicPassword"`
	BearerToken   string `json:"bearerToken"`
}

// Config — HTTP-клиент для проброса запросов на upstream.
type Config struct {
	Host            string `json:"Host"`
	ResponseTimeout string `json:"ResponseTimeout"`
	Auth            Auth   `json:"Auth"`
}

func DefaultConfig() Config {
	return Config{
		Host:            "http://127.0.0.1:8080",
		ResponseTimeout: "10s",
		Auth: Auth{
			Type:          "basic",
			BasicLogin:    "",
			BasicPassword: "",
		},
	}
}
