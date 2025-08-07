package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./bin <emotes-path> <output-file>")
		return
	}

	path := os.Args[1]
	output := os.Args[2]

	var out io.Writer
	if output == "" {
		out = os.Stdout
	} else {
		file, err := os.Create(output)
		if err != nil {
			fmt.Printf("cannot create/open file at %s: %v\n", output, err)
			return
		}
		defer file.Close() //nolint:errcheck
		out = file
	}

	if out == nil {
		panic("out is nil")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Cannot read file '%s': %v\n", path, err)
		return
	}

	var emotes JSONEmotes
	if err = json.Unmarshal(data, &emotes); err != nil {
		fmt.Printf("Cannot unmarshal json: %v\n", err)
		return
	}

	content, err := generateContent(emotes)
	if err != nil {
		fmt.Printf("Cannot generate content: %v\n", err)
		return
	}

	if _, err := out.Write(content); err != nil {
		fmt.Printf("Cannot write to output: %v\n", err)
	}
}

func generateContent(emotes JSONEmotes) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("package emote\n\n")
	buf.WriteString("import \"github.com/google/uuid\"\n\n")
	buf.WriteString("var (\n")

	nameCount := make(map[string]int)

	for _, emote := range emotes.Emotes {
		baseVarName := toCamelCase(emote.Name)
		if baseVarName == "" {
			baseVarName = "Emote_" + strings.ReplaceAll(emote.UUID, "-", "")
		} else {
			runes := []rune(baseVarName)
			if len(runes) > 0 && !unicode.IsLetter(runes[0]) {
				baseVarName = "Emote" + baseVarName
			}
		}

		count := nameCount[baseVarName]
		nameCount[baseVarName]++
		varName := baseVarName
		if count > 0 {
			varName = fmt.Sprintf("%s_%d", baseVarName, count)
		}

		escapedName := escapeString(emote.Name)
		escapedRarity := escapeString(emote.Rarity)

		keywords := make([]string, len(emote.Keywords))
		for i, kw := range emote.Keywords {
			keywords[i] = strconv.Quote(kw)
		}
		keywordsStr := "[]string{" + strings.Join(keywords, ", ") + "}"

		buf.WriteString("\t" + varName + " = Emote{\n")
		buf.WriteString("\t\tUUID:     uuid.MustParse(\"" + emote.UUID + "\"),\n")
		buf.WriteString("\t\tName:     \"" + escapedName + "\",\n")
		buf.WriteString("\t\tRarity:   \"" + escapedRarity + "\",\n")
		buf.WriteString("\t\tKeywords: " + keywordsStr + ",\n")
		buf.WriteString("\t}\n\n")
	}

	buf.WriteString(")\n")
	return buf.Bytes(), nil
}

func toCamelCase(s string) string {
	var result strings.Builder
	var word []rune
	inWord := false

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !inWord {
				inWord = true
				word = []rune{r}
			} else {
				word = append(word, r)
			}
		} else {
			if inWord {
				if len(word) > 0 {
					word[0] = unicode.ToUpper(word[0])
					result.WriteString(string(word))
				}
				inWord = false
			}
		}
	}

	if len(word) < 1 {
		panic("no words")
	}

	if inWord {
		word[0] = unicode.ToUpper(word[0])
		result.WriteString(string(word))
	}

	return result.String()
}

func escapeString(s string) string {
	quoted := strconv.Quote(s)
	return quoted[1 : len(quoted)-1]
}

type JSONEmotes struct {
	Emotes []JSONEmote `json:"emotes"`
}

type JSONEmote struct {
	UUID     string   `json:"uuid"`
	Name     string   `json:"title"`
	Rarity   string   `json:"rarity"`
	Keywords []string `json:"keywords"`
}
