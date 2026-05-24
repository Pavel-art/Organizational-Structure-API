package configs

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	loadedEnvFiles []string
	loadedEnvOnce  sync.Once
)

func Init() {
	loadedEnvOnce.Do(func() {
		loadedEnvFiles = loadEnvFiles(".env", ".env.local")
	})
}

func loadEnvFiles(files ...string) []string {
	var loaded []string
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		if err := godotenv.Overload(f); err != nil {
			log.Printf("Failed to load %s: %v", f, err)
			continue
		}
		loaded = append(loaded, f)
	}
	if len(loaded) == 0 {
		log.Println("No .env files found")
	} else {
		log.Printf("Loaded env files: %s", strings.Join(loaded, ", "))
	}
	return loaded
}

func isRunningInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// Fallback: check cgroup hints (best-effort).
	if b, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		s := string(b)
		return strings.Contains(s, "docker") || strings.Contains(s, "containerd") || strings.Contains(s, "kubepods")
	}
	return false
}

func getString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

type DatabaseConfig struct {
	Url string
}

func NewDataBaseConfig() *DatabaseConfig {
	dbURL := getString("DB_URL", "")
	if dbURL == "" {
		dbURL = buildPostgresURLFromParts()
	}
	dbURL = maybeRewriteDockerComposeHost(dbURL)
	return &DatabaseConfig{
		Url: dbURL,
	}
}

func buildPostgresURLFromParts() string {
	user := getString("POSTGRES_USER", "")
	pass := getString("POSTGRES_PASSWORD", "")
	db := getString("POSTGRES_DB", "")
	if user == "" || db == "" {
		return ""
	}

	host := getString("DB_HOST", "")
	port := getString("DB_PORT", "")
	if host == "" {
		if isRunningInDocker() {
			host = "postgres"
		} else {
			host = "localhost"
		}
	}
	if port == "" {
		if isRunningInDocker() {
			port = getString("POSTGRES_PORT", "5432")
		} else {
			// docker-compose maps container 5432 -> host 5435 in this repo.
			port = "5435"
		}
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   host + ":" + port,
		Path:   "/" + db,
	}
	q := url.Values{}
	q.Set("sslmode", getString("DB_SSLMODE", "disable"))
	u.RawQuery = q.Encode()
	return u.String()
}

func maybeRewriteDockerComposeHost(dbURL string) string {
	if dbURL == "" || isRunningInDocker() {
		return dbURL
	}
	// Only rewrite when DB_URL uses docker-compose service hostname ("postgres")
	// and the app is running on the host (common local-dev pitfall).
	u, err := url.Parse(dbURL)
	if err != nil {
		return dbURL
	}
	if strings.ToLower(u.Scheme) != "postgres" && strings.ToLower(u.Scheme) != "postgresql" {
		return dbURL
	}
	if u.Hostname() != "postgres" {
		return dbURL
	}
	// Respect explicit overrides.
	if os.Getenv("DB_HOST") != "" || os.Getenv("DB_PORT") != "" {
		return dbURL
	}

	// Prefer host-mapped port from docker-compose in this repo.
	port := u.Port()
	if port == "" || port == "5432" {
		port = "5435"
	}
	u.Host = "localhost:" + port

	// If run from outside repo root, keep message stable by using absolute path.
	cwd, _ := os.Getwd()
	envLocalHint := filepath.Join(cwd, ".env.local")
	log.Printf("DB_URL points to docker-compose host \"postgres\" but the app is not running in a container; using %q instead. To make it explicit, set DB_URL in %s.", u.String(), envLocalHint)

	return u.String()
}

type LogConfig struct {
	Level  int
	Format string
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  getInt("LOG_LEVEL", 0),
		Format: getString("LOG_FORMAT", "json"),
	}
}

type ServerConfig struct {
	Port string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: getString("HTTP_PORT", "8081"),
	}
}
