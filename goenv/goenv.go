//Package godotenv is based on github.com/joho/godotenv
//adding envMap injection,
//and using github.com/a8m/envsubst modified to expandVars
package goenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/goenv/parse"
)

const doubleQuoteSpecialChars = "\\\n\r\"!$`"

// Load will read your env file(s) and load them into ENV for this process.
// It's important to note this WILL OVERRIDE an env variable that already exists - consider the .env file to forcefilly set all vars.
func Load(filenames ...string) (err error) {
	for _, filename := range filenames {
		err = loadFile(filename, true)
		if err != nil {
			return // return early on a spazout
		}
	}
	return
}

// LoadDefault will read your env file(s) and load them into ENV for this process.
// It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults
func LoadDefault(filenames ...string) (err error) {
	for _, filename := range filenames {
		err = loadFile(filename, false)
		if err != nil {
			return // return early on a spazout
		}
	}
	return
}

// Marshal outputs the given environment as a dotenv-formatted environment file.
// Each line is in the format: KEY="VALUE" where VALUE is backslash-escaped.
func Marshal(envMap map[string]string) (string, error) {
	lines := make([]string, 0, len(envMap))
	for k, v := range envMap {
		lines = append(lines, fmt.Sprintf(`%s="%s"`, k, doubleQuoteEscape(v)))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
}

//Unmarshal reads an env file from a string
func Unmarshal(str string, envMap map[string]string) (err error) {
	mapping := createMapping(envMap)
	return Parse(strings.NewReader(str), envMap, mapping)
}

//UnmarshalDefault reads an env file from a string
func UnmarshalDefault(str string, envMap map[string]string, envMapDefault map[string]string) (err error) {
	mapping := createMapping(envMap, envMapDefault)
	return Parse(strings.NewReader(str), envMap, mapping)
}

// Write serializes the given environment and writes it to a file
func Write(envMap map[string]string, filename string) error {
	content, error := Marshal(envMap)
	if error != nil {
		return error
	}
	file, error := os.Create(filename)
	if error != nil {
		return error
	}
	_, err := file.WriteString(content)
	if err == nil {
		err = file.Sync()
	}
	return err
}

// Read all env to a map
func Read(filename string, envMap map[string]string) error {
	mapping := createMapping(envMap)
	return ReadMapping(filename, envMap, mapping)
}

func ReadDefault(filename string, envMap map[string]string, envMapDefault map[string]string) error {
	mapping := createMapping(envMap, envMapDefault)
	return ReadMapping(filename, envMap, mapping)
}

// Parse reads an env file from io.Reader, returning a map of keys and values.
func Parse(r io.Reader, envMap map[string]string, mapping func(string) (string, bool)) (err error) {

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine, envMap, mapping)

			if err != nil {
				return
			}
			envMap[key] = value
		}
	}
	return
}

func ReadMapping(filename string, envMap map[string]string, mapping func(string) (string, bool)) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return Parse(file, envMap, mapping)
}

func EnvToPairs(envmap map[string]string) []string {
	environ := make([]string, len(envmap))
	for key, val := range envmap {
		environ = append(environ, key+"="+val)
	}
	return environ
}

func EnvToMap(data []string) map[string]string {
	items := make(map[string]string)
	for _, item := range data {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		items[key] = val
	}
	return items
}

func loadFile(filename string, overload bool) error {
	envMap := make(map[string]string)
	envMapDefault := EnvToMap(os.Environ())
	if err := ReadDefault(filename, envMap, envMapDefault); err != nil {
		return err
	}

	for key, value := range envMap {
		if _, hasKey := envMap[key]; hasKey || overload {
			os.Setenv(key, value)
		}
	}

	return nil
}

var exportRegex = regexp.MustCompile(`^\s*(?:export\s+)?(.*?)\s*$`)

func parseLine(line string, envMap map[string]string, mapping func(string) (string, bool)) (key string, value string, err error) {
	if len(line) == 0 {
		err = errors.New("zero length string")
		return
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(line, "=")
	firstColon := strings.Index(line, ":")
	splitString := strings.SplitN(line, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		//this is a yaml-style line
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		err = errors.New("Can't separate key from value")
		return
	}

	// Parse the key
	key = splitString[0]
	if strings.HasPrefix(key, "export") {
		key = strings.TrimPrefix(key, "export")
	}
	key = strings.TrimSpace(key)

	key = exportRegex.ReplaceAllString(splitString[0], "$1")

	// Parse the value
	value, err = parseValue(splitString[1], envMap, mapping)
	return
}

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

func parseValue(value string, envMap map[string]string, mapping func(string) (string, bool)) (string, error) {

	var err error

	// trim
	value = strings.Trim(value, " ")

	// check if we've got quoted values or possible escapes
	if len(value) > 1 {
		singleQuotes := singleQuotesRegex.FindStringSubmatch(value)

		doubleQuotes := doubleQuotesRegex.FindStringSubmatch(value)

		if singleQuotes != nil || doubleQuotes != nil {
			// pull the quotes off the edges
			value = value[1 : len(value)-1]
		}

		if doubleQuotes != nil {
			// expand newlines
			value = escapeRegex.ReplaceAllStringFunc(value, func(match string) string {
				c := strings.TrimPrefix(match, `\`)
				switch c {
				case "n":
					return "\n"
				case "r":
					return "\r"
				default:
					return match
				}
			})
			// unescape characters
			value = unescapeCharsRegex.ReplaceAllString(value, "$1")
		}

		if singleQuotes == nil {
			value, err = ExpandMapping(value, mapping)
		}
	}

	return value, err
}

func Expand(v string, envMap map[string]string) (string, error) {
	mapping := createMapping(envMap)
	return ExpandMapping(v, mapping)
}

func ExpandMapping(v string, mapping func(string) (string, bool)) (string, error) {
	env := parse.Env{
		Mapping: mapping,
	}
	parser := &parse.Parser{
		Name:     "bytes",
		Env:      env,
		Restrict: parse.Relaxed,
	}
	return parser.Parse(v)
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

func doubleQuoteEscape(line string) string {
	for _, c := range doubleQuoteSpecialChars {
		toReplace := "\\" + string(c)
		if c == '\n' {
			toReplace = `\n`
		}
		if c == '\r' {
			toReplace = `\r`
		}
		line = strings.Replace(line, string(c), toReplace, -1)
	}
	return line
}

func createMapping(envMaps ...map[string]string) func(key string) (string, bool) {
	mapping := func(key string) (string, bool) {
		for _, envMap := range envMaps {
			if val, ok := envMap[key]; ok {
				return val, true
			}
		}
		return "", false
	}
	return mapping
}
