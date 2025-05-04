package dotenv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

const DEFAULT_FILENAME = ".env"

func evalDollars(s string) string {
	evaled := ""
	tmp := ""
	inEnv := false
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' && inEnv {
			evaled += os.Getenv(tmp)
			tmp = ""
			inEnv = false
		}

		if i+1 < len(s) && s[i+1] == '}' {
			if !inEnv {
				continue
			}

			inEnv = false
			evaled += os.Getenv(tmp + string(s[i]))
			tmp = ""
			i++
			continue
		}

		if s[i] != '$' {
			tmp += string(s[i])
			continue
		}

		if i+1 >= len(s) {
			continue
		}

		inEnv = true
		evaled += tmp
		tmp = ""
		if s[i+1] == '{' {
			i++
		}
	}

	if tmp != "" {
		if inEnv {
			evaled += os.Getenv(tmp)
		} else {
			evaled += tmp
		}
	}

	return evaled
}

func ReadFile(fname string) ([]string, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)

	return strings.Split(string(data), "\n"), nil
}

func LoadFile(fname string) error {
	lines, err := ReadFile(fname)
	if err != nil {
		return err
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		subs := strings.SplitN(line, "=", 2)
		if len(subs) != 2 {
			return fmt.Errorf("invalid line %d: %s", i, line)
		}

		first := strings.TrimSpace(subs[0])
		second := strings.TrimSpace(subs[1])
		if len(first) == 0 || len(second) == 0 {
			return fmt.Errorf("invalid line %d: %s", i, line)
		}

		second_begin := second[0]
		second_end := second[len(second)-1]
		if (second_begin == '"' && second_end == '"') ||
			(second_begin == '\'' && second_end == '\'') {
			second = second[1 : len(second)-1]
		}

		second = evalDollars(second)

		err = os.Setenv(first, second)
		if err != nil {
			return err
		}
	}

	return nil
}

func Load(fnames ...string) error {
	if len(fnames) == 0 {
		return LoadFile(DEFAULT_FILENAME)
	}

	for _, fname := range fnames {
		err := LoadFile(fname)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteFile(fname string, data string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.WriteString(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("write error: %d != %d", n, len(data))
	}

	return nil
}

func SaveFile(fname string, env map[string]string) error {
	data := ""
	for key, value := range env {
		data += fmt.Sprintf("%s=\"%s\"\n", key, value)
	}

	return WriteFile(fname, data)
}

func Save(fname string) error {
	env := os.Environ()
	env_map := make(map[string]string)
	for _, e := range env {
		subs := strings.SplitN(e, "=", 2)
		if len(subs) != 2 {
			return fmt.Errorf("invalid env: %s", e)
		}

		env_map[subs[0]] = subs[1]
	}

	return SaveFile(fname, env_map)
}
