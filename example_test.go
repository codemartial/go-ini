package ini_test

import (
	"os"
	"time"

	ini "github.com/pierrec/go-ini"
)

// Password holds a password in clear.
// It implements encoding.TextMarshaler and encoding.TextUnmarshaler.
type Password string

// MarshalText obfuscates the password.
func (p *Password) MarshalText() ([]byte, error) {
	buf := []byte(*p)
	rot13(buf)
	return buf, nil
}

// UnmarshalText unobfuscates the password.
func (p *Password) UnmarshalText(buf []byte) error {
	rot13(buf)
	*p = Password(buf)
	return nil
}

func rot13(buf []byte) {
	for i, c := range buf {
		if (c >= 'A' && c < 'N') || (c >= 'a' && c < 'n') {
			buf[i] += 13
		} else if (c > 'M' && c <= 'Z') || (c > 'm' && c <= 'z') {
			buf[i] -= 13
		}
	}
}

type User struct {
	Username string   `ini:"usr,client"`
	Password Password `ini:"pwd,client"`
}

// Config is the structure to hold the data found in the ini source.
type Config struct {
	Host     string        `ini:"host,server"`
	Port     int           `ini:"port,server"`
	Enabled  bool          `ini:"enabled,server"`
	Timeout  time.Duration `ini:"timeout,server"`
	Deadline time.Time     `ini:"deadline,"`
	// Embedded types are supported.
	User
	// As well as slices.
	Children []string `ini:"children,family"`
	Ages     []int    `ini:"ages,family"`
}

func Example() {
	date, _ := time.Parse("15:04:05Z", "05:01:01Z")
	conf := &Config{
		"localhost",
		8080,
		true,
		3 * time.Second,
		date,
		// Although the password is in clear,
		// it will be obfuscated when encoded.
		User{"bob the cat", "password"},
		[]string{"Brian", "Kelly"},
		[]int{3, 7},
	}

	// Encode the configuration.
	ini.Encode(os.Stdout, conf)

	// Output: deadline = 0000-01-01T05:01:01Z
	//
	// [server]
	// host    = localhost
	// port    = 8080
	// enabled = true
	// timeout = 3s
	//
	// [client]
	// usr = bob the cat
	// pwd = cnffjbeq
	//
	// [family]
	// children = Brian,Kelly
	// ages     = 3,7
}
