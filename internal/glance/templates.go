package glance

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var intl = message.NewPrinter(language.English)

var globalTemplateFunctions = template.FuncMap{
	"json": func(v interface{}) (template.JS, error) {
		a, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return template.JS(a), nil
	},
	"formatApproxNumber": formatApproxNumber,
	"formatNumber":       intl.Sprint,
	"safeCSS": func(str string) template.CSS {
		return template.CSS(str)
	},
	"safeURL": func(str string) template.URL {
		return template.URL(str)
	},
	"safeHTML": func(str string) template.HTML {
		return template.HTML(str)
	},
	"absInt": func(i int) int {
		return int(math.Abs(float64(i)))
	},
	"formatPrice": func(price float64) string {
		return intl.Sprintf("%.2f", price)
	},
	"formatPriceWithPrecision": func(precision int, price float64) string {
		return intl.Sprintf("%."+strconv.Itoa(precision)+"f", price)
	},
	"divideFloat": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"dynamicRelativeTimeAttrs": dynamicRelativeTimeAttrs,
	"formatPolishDate":         formatPolishDate,
	"formatPolishRelativeTime": formatPolishRelativeTime,
	"languageIcon":             languageIcon,
	"formatServerMegabytes": func(mb uint64) template.HTML {
		var value string
		var label string

		if mb < 1_000 {
			value = strconv.FormatUint(mb, 10)
			label = "MB"
		} else if mb < 1_000_000 {
			if mb < 10_000 {
				value = fmt.Sprintf("%.1f", float64(mb)/1_000)
			} else {
				value = strconv.FormatUint(mb/1_000, 10)
			}

			label = "GB"
		} else {
			value = fmt.Sprintf("%.1f", float64(mb)/1_000_000)
			label = "TB"
		}

		return template.HTML(value + ` <span class="color-base size-h5">` + label + `</span>`)
	},
}

func mustParseTemplate(primary string, dependencies ...string) *template.Template {
	t, err := template.New(primary).
		Funcs(globalTemplateFunctions).
		ParseFS(templateFS, append([]string{primary}, dependencies...)...)

	if err != nil {
		panic(err)
	}

	return t
}

func formatApproxNumber(count int) string {
	if count < 1_000 {
		return strconv.Itoa(count)
	}

	if count < 10_000 {
		return strconv.FormatFloat(float64(count)/1_000, 'f', 1, 64) + "k"
	}

	if count < 1_000_000 {
		return strconv.Itoa(count/1_000) + "k"
	}

	return strconv.FormatFloat(float64(count)/1_000_000, 'f', 1, 64) + "m"
}

func dynamicRelativeTimeAttrs(t interface{ Unix() int64 }) template.HTMLAttr {
	return template.HTMLAttr(`data-dynamic-relative-time="` + strconv.FormatInt(t.Unix(), 10) + `"`)
}

func formatPolishDate(t interface{ Unix() int64 }) string {
	days := []string{"niedziela", "poniedziałek", "wtorek", "środa", "czwartek", "piątek", "sobota"}
	months := []string{"stycznia", "lutego", "marca", "kwietnia", "maja", "czerwca", "lipca", "sierpnia", "września", "października", "listopada", "grudnia"}

	unix := t.Unix()
	if unix == 0 {
		return ""
	}

	timeObj := time.Unix(unix, 0)
	weekday := days[timeObj.Weekday()]
	day := timeObj.Day()
	month := months[timeObj.Month()-1]
	year := timeObj.Year()

	return fmt.Sprintf("%s, %d %s %d", weekday, day, month, year)
}

func formatPolishRelativeTime(t interface{ Unix() int64 }) string {
	unix := t.Unix()
	if unix == 0 {
		return ""
	}

	now := time.Now().Unix()
	delta := now - unix

	if delta < 0 {
		delta = -delta
	}

	const minuteInSeconds = 60
	const hourInSeconds = minuteInSeconds * 60
	const dayInSeconds = hourInSeconds * 24
	const monthInSeconds = dayInSeconds * 30
	const yearInSeconds = dayInSeconds * 365

	if delta < minuteInSeconds {
		return "przed chwilą"
	}
	if delta < hourInSeconds {
		mins := int(delta / minuteInSeconds)
		return formatPolishPlural(mins, "minutę", "minuty", "minut") + " temu"
	}
	if delta < dayInSeconds {
		hours := int(delta / hourInSeconds)
		return formatPolishPlural(hours, "godzinę", "godziny", "godz.") + " temu"
	}
	if delta < monthInSeconds {
		days := int(delta / dayInSeconds)
		return formatPolishPlural(days, "dzień", "dni", "dni") + " temu"
	}
	if delta < yearInSeconds {
		months := int(delta / monthInSeconds)
		return formatPolishPlural(months, "miesiąc", "miesiące", "mies.") + " temu"
	}

	years := int(delta / yearInSeconds)
	return formatPolishPlural(years, "rok", "lata", "lat") + " temu"
}

func formatPolishPlural(n int, one, twoFour, fivePlus string) string {
	n = n % 100
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	if n >= 2 && n <= 4 {
		return fmt.Sprintf("%d %s", n, twoFour)
	}
	return fmt.Sprintf("%d %s", n, fivePlus)
}

func languageIcon(language string) string {
	if language == "" {
		return ""
	}

	languageSlugs := map[string]string{
		"JavaScript":   "javascript",
		"TypeScript":   "typescript",
		"Python":       "python",
		"Java":         "java",
		"C#":           "csharp",
		"C++":          "cpp",
		"C":            "c",
		"Go":           "go",
		"Rust":         "rust",
		"Ruby":         "ruby",
		"PHP":          "php",
		"Swift":        "swift",
		"Kotlin":       "kotlin",
		"Scala":        "scala",
		"Shell":        "gnubash",
		"HTML":         "html5",
		"CSS":          "css3",
		"SCSS":         "sass",
		"Vue":          "vuedotjs",
		"Svelte":       "svelte",
		"JSON":         "json",
		"YAML":         "yaml",
		"Markdown":     "markdown",
		"SQL":          "postgresql",
		"Dockerfile":   "docker",
		"R":            "r",
		"Matlab":       "matlab",
		"Haskell":      "haskell",
		"Elixir":       "elixir",
		"Clojure":      "clojure",
		"Perl":         "perl",
		"Lua":          "lua",
		"Dart":         "dart",
		"Groovy":       "gradle",
		"Objective-C":  "objectivec",
		"Visual Basic": "visualbasic",
		"F#":           "fsharp",
		"OCaml":        "ocaml",
		"Julia":        "julia",
		"COBOL":        "cobol",
		"Fortran":      "fortran",
		"Crystal":      "crystal",
		"Zig":          "zig",
		"Nim":          "nim",
		"V":            "v",
		"Nix":          "nix",
		"Ada":          "ada",
		"Prolog":       "prolog",
		"Erlang":       "erlang",
		"Pascal":       "pascal",
		"Delphi":       "delphi",
		"Assembly":     "assemblyscript",
	}

	slug, exists := languageSlugs[language]
	if !exists {
		return ""
	}

	return "https://cdn.jsdelivr.net/npm/simple-icons@latest/icons/" + slug + ".svg"
}
