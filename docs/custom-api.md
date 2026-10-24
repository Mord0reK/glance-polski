[Przejdź do definicji funkcji](#functions)

## Przykłady

Najlepszym sposobem na zrozumienie działania szablonów jest zapoznanie się z wieloma przykładami. Oto najczęstsze przypadki użycia:

Odpowiedź JSON:

```json
{
  "title": "My Title",
  "content": "My Content",
}
```

Aby uzyskać dostęp do obu pól odpowiedzi JSON, użyj następującego kodu:

```html
<div>{{ .JSON.String "title" }}</div>
<div>{{ .JSON.String "content" }}</div>
```

Wynik:

```html
<div>My Title</div>
<div>My Content</div>
```

<hr>

Odpowiedź JSON:

```json
{
  "author": "John Doe",
  "posts": [
    {
      "title": "My Title",
      "content": "My Content"
    },
    {
      "title": "My Title 2",
      "content": "My Content 2"
    }
  ]
}
```

Aby przeiterować po tablicy postów, użyj następującego kodu:

```html
{{ range .JSON.Array "posts" }}
  <div>{{ .String "title" }}</div>
  <div>{{ .String "content" }}</div>
{{ end }}
```

Wynik:

```html
<div>My Title</div>
<div>My Content</div>
<div>My Title 2</div>
<div>My Content 2</div>
```

Zwróć uwagę na brak `.JSON` przy dostępie do tytułu i zawartości — wynika to z faktu, że funkcja `range` ustawia kontekst na bieżący element tablicy.

Jeśli chcesz uzyskać dostęp do kontekstu najwyższego poziomu wewnątrz `range`, możesz użyć następującego kodu:

```html
{{ range .JSON.Array "posts" }}
  <div>{{ .String "title" }}</div>
  <div>{{ .String "content" }}</div>
  <div>{{ $.JSON.String "author" }}</div>
{{ end }}
```

Wynik:

```html
<div>My Title</div>
<div>My Content</div>
<div>John Doe</div>
<div>My Title 2</div>
<div>My Content 2</div>
<div>John Doe</div>
```

<hr>

Odpowiedź JSON:

```json
[
    "Apple",
    "Banana",
    "Cherry",
    "Watermelon"
]
```

Nieco niezręcznie: gdy bieżący kontekst jest typem podstawowym, który nie jest obiektem, jego typ określa się, używając pustego ciągu znaków jako klucza. Aby przeiterować po tablicy ciągów znaków, użyj następującego kodu:

```html
{{ range .JSON.Array "" }}
  <div>{{ .String "" }}</div>
{{ end }}
```

Wynik:

```html
<div>Apple</div>
<div>Banana</div>
<div>Cherry</div>
<div>Watermelon</div>
```

Aby uzyskać dostęp do elementu o określonym indeksie, możesz użyć następującego kodu:

```html
<div>{{ .JSON.String "0" }}</div>
```

Wynik:

```html
<div>Apple</div>
```

<hr>

Odpowiedź JSON:

```json
{
    "user": {
        "address": {
            "city": "New York",
            "state": "NY"
        }
    }
}
```

Aby łatwo uzyskać dostęp do głęboko zagnieżdżonych obiektów, możesz użyć notacji kropkowej:

```html
<div>{{ .JSON.String "user.address.city" }}</div>
<div>{{ .JSON.String "user.address.state" }}</div>
```

Wynik:

```html
<div>New York</div>
<div>NY</div>
```

Obsługiwane jest również używanie indeksów w dowolnym miejscu ścieżki:

```json
{
    "users": [
        {
            "name": "John Doe"
        },
        {
            "name": "Jane Doe"
        }
    ]
}
```

```html
<div>{{ .JSON.String "users.0.name" }}</div>
<div>{{ .JSON.String "users.1.name" }}</div>
```

Wynik:

```html
<div>John Doe</div>
<div>Jane Doe</div>
```

<hr>

Odpowiedź JSON:

```json
{
    "user": {
        "name": "John Doe",
        "age": 30
    }
}
```

Aby sprawdzić, czy pole istnieje, możesz użyć następującego kodu:

```html
{{ if .JSON.Exists "user.age" }}
  <div>{{ .JSON.Int "user.age" }}</div>
{{ else }}
  <div>Age not provided</div>
{{ end }}
```

Wynik:

```html
<div>30</div>
```

<hr>

Odpowiedź JSON:

```json
{
    "price": 100,
    "discount": 10
}
```

Obliczenia można wykonywać zarówno na liczbach całkowitych, jak i zmiennoprzecinkowych. Jeśli obie liczby są całkowite, zwracana jest liczba całkowita; w przeciwnym razie zwracana jest liczba zmiennoprzecinkowa. Przy próbie dzielenia przez zero zwracane jest 0. Jeśli podasz wartości nienumeryczne, zwrócone zostanie `NaN`.

```html
<div>{{ sub (.JSON.Int "price") (.JSON.Int "discount") }}</div>
```

Wynik:

```html
<div>90</div>
```

Inne dostępne operacje to `add`, `mul`, `div` oraz `mod`.

<hr>

Odpowiedź JSON:

```json
{
  "posts": [
    {
      "title": "Exploring the Depths of Quantum Computing",
      "date": "2023-10-27T10:00:00Z"
    },
    {
      "title": "A Beginner's Guide to Sustainable Living",
      "date": "2023-11-15T14:30:00+01:00"
    },
    {
      "title": "The Art of Baking Sourdough Bread",
      "date": "2023-12-03T08:45:22-08:00"
    }
  ]
}
```

Aby sparsować datę i wyświetlić czas względny (np. 2h, 1d itp.), użyj następującego kodu:

```
{{ range .JSON.Array "posts" }}
  <div>{{ .String "title" }}</div>
  <div {{ .String "date" | parseTime "rfc3339" | toRelativeTime }}></div>
{{ end }}
```

Funkcja `parseTime` przyjmuje dwa argumenty: układ daty oraz sam ciąg znaków z datą. Układ może być jednym z następujących: "RFC3339", "RFC3339Nano", "DateTime", "DateOnly", "TimeOnly" lub własnym układem zgodnym z [formatem daty](https://pkg.go.dev/time#pkg-constants) Go.

Wynik:

```html
<div>Exploring the Depths of Quantum Computing</div>
<div data-dynamic-relative-time="1698400800"></div>

<div>A Beginner's Guide to Sustainable Living</div>
<div data-dynamic-relative-time="1700055000"></div>

<div>The Art of Baking Sourdough Bread</div>
<div data-dynamic-relative-time="1701621922"></div>
```

Nie musisz martwić się o wewnętrzną implementację — Glance dynamicznie uzupełni te wartości po stronie klienta, aby wyświetlić prawidłowy czas względny.

Warto zauważyć, że wartość zwracana przez `toRelativeTime` musi być użyta jako atrybut znacznika HTML, np. `div`, `li`, `span` itp.

<hr>

W niektórych przypadkach możesz chcieć poznać kod statusu odpowiedzi. Możesz to zrobić w następujący sposób:

```html
{{ if eq .Response.StatusCode 200 }}
  <p>Success!</p>
{{ else }}
  <p>Failed to fetch data</p>
{{ end }}
```

Możesz również uzyskać dostęp do nagłówków odpowiedzi:

```html
<div>{{ .Response.Header.Get "Content-Type" }}</div>
```

<hr>

Odpowiedź JSON:

```json
{"name": "Steve", "age": 30}
{"name": "Alex", "age": 25}
{"name": "John", "age": 35}
```

Powyższy format to "[ndjson](https://docs.mulesoft.com/dataweave/latest/dataweave-formats-ndjson)" lub "[JSON Lines](https://jsonlines.org/)", gdzie każda linia jest osobnym obiektem JSON. Aby sparsować ten format, musisz najpierw wyłączyć sprawdzanie poprawności JSON w konfiguracji, ponieważ domyślnie oczekiwana jest pojedyncza poprawna odpowiedź JSON:

```yaml
- type: custom-api
  skip-json-validation: true
```

Następnie, aby przeiterować po każdym obiekcie, możesz użyć `.JSONLines`:

```html
{{ range .JSONLines }}
  <p>{{ .String "name" }} is {{ .Int "age" }} years old</p>
{{ end }}
```

Wynik:

```html
<p>Steve is 30 years old</p>
<p>Alex is 25 years old</p>
<p>John is 35 years old</p>
```

Inne sposoby wybierania danych z odpowiedzi JSON Lines znajdziesz w dokumentacji [tidwall/gjson](https://github.com/tidwall/gjson/tree/master?tab=readme-ov-file#json-lines). Na przykład, aby uzyskać tablicę wszystkich imion, możesz użyć następującego kodu:

```html
{{ range .JSON.Array "..#.name" }}
  <p>{{ .String "" }}</p>
{{ end }}
```

Wynik:

```html
<p>Steve</p>
<p>Alex</p>
<p>John</p>
```

<hr>

W niektórych przypadkach możesz potrzebować wykonać dwa kolejne wywołania API, gdzie wynik pierwszego wywołania wykorzystasz w drugim. Aby to osiągnąć, możesz wykonać dodatkowe żądania HTTP bezpośrednio z poziomu szablonu, używając następującej składni:

```yaml
- type: custom-api
  url: https://api.example.com/get-id-of-something
  template: |
    {{ $theID := .JSON.String "id" }}

    {{
      $something := newRequest (concat "https://api.example.com/something/" $theID)
        | withParameter "key" "value"
        | withHeader "Authorization" "Bearer token"
        | getResponse
    }}

    {{ $something.JSON.String "title" }}
```

Tutaj `$theID` jest pobierany z wyniku pierwszego wywołania API i używany w drugim wywołaniu. Funkcja `newRequest` tworzy nowe żądanie, a `getResponse` je wykonuje. Możesz także użyć `withParameter` i `withHeader`, aby opcjonalnie dodać parametry i nagłówki do żądania.

Jeśli musisz wykonać żądanie do adresu URL wymagającego dynamicznych parametrów, możesz pominąć właściwość `url` w YAML i wykonać żądanie w całości z poziomu szablonu:

```yaml
- type: custom-api
  title: Events from the last 24h
  template: |
    {{
      $events := newRequest "https://api.example.com/events"
        | withParameter "after" (offsetNow "-24h" | formatTime "rfc3339")
        | getResponse
    }}

    {{ if eq $events.Response.StatusCode 200 }}
      {{ range $events.JSON.Array "events" }}
        <div>{{ .String "title" }}</div>
        <div {{ .String "date" | parseTime "rfc3339" | toRelativeTime }}></div>
      {{ end }}
    {{ else }}
      <p>Failed to fetch data: {{ $events.Response.Status }}</p>
    {{ end }}
```

*Pamiętaj, że musisz ręcznie sprawdzić poprawność kodu statusu.*

<a id="functions"></a>
## Funkcje

W obiekcie `JSON` dostępne są następujące funkcje:

- `String(key string) string`: Zwraca wartość klucza jako ciąg znaków.
- `Int(key string) int`: Zwraca wartość klucza jako liczbę całkowitą.
- `Float(key string) float`: Zwraca wartość klucza jako liczbę zmiennoprzecinkową.
- `Bool(key string) bool`: Zwraca wartość klucza jako wartość logiczną.
- `Array(key string) []JSON`: Zwraca wartość klucza jako tablicę obiektów `JSON`.
- `Exists(key string) bool`: Zwraca `true`, jeśli klucz istnieje w obiekcie JSON.

W obiekcie `Options` dostępne są następujące funkcje:

- `StringOr(key string, default string) string`: Zwraca wartość klucza jako ciąg znaków lub wartość domyślną, jeśli klucz nie istnieje.
- `IntOr(key string, default int) int`: Zwraca wartość klucza jako liczbę całkowitą lub wartość domyślną, jeśli klucz nie istnieje.
- `FloatOr(key string, default float) float`: Zwraca wartość klucza jako liczbę zmiennoprzecinkową lub wartość domyślną, jeśli klucz nie istnieje.
- `BoolOr(key string, default bool) bool`: Zwraca wartość klucza jako wartość logiczną lub wartość domyślną, jeśli klucz nie istnieje.
- `JSON(key string) JSON`: Zwraca wartość klucza jako zserializowany obiekt `JSON` lub zgłasza błąd, jeśli klucz nie istnieje.

Dostępne są następujące funkcje pomocnicze udostępniane przez Glance:

- `toFloat(i int) float`: Konwertuje liczbę całkowitą na liczbę zmiennoprzecinkową.
- `toInt(f float) int`: Konwertuje liczbę zmiennoprzecinkową na liczbę całkowitą.
- `toRelativeTime(t time.Time) template.HTMLAttr`: Konwertuje czas na czas względny, np. 2h, 1d itp., który dynamicznie się aktualizuje. **UWAGA:** wartość tej funkcji należy użyć jako atrybutu znacznika HTML, np. `<span {{ toRelativeTime .Time }}></span>`.
- `now() time.Time`: Zwraca bieżący czas.
- `offsetNow(offset string) time.Time`: Zwraca bieżący czas z przesunięciem. Przesunięcie może być dodatnie lub ujemne i musi mieć format "3h", "-1h" lub "2h30m10s".
- `duration(str string) time.Duration`: Parsuje ciąg znaków taki jak `1h`, `24h`, `5h30m` itp. do typu `time.Duration`.
- `parseTime(layout string, s string) time.Time`: Parsuje ciąg znaków do `time.Time`. Układ musi być zgodny z [formatem daty](https://pkg.go.dev/time#pkg-constants) Go. Zamiast dosłownego formatu możesz użyć następujących wartości: "unix", "RFC3339", "RFC3339Nano", "DateTime", "DateOnly".
- `formatTime(layout string, s string) time.Time`: Formatuje `time.Time` do ciągu znaków. Układ jest zgodny z formatem używanym przez `parseTime`.
- `parseLocalTime(layout string, s string) time.Time`: Działa jak powyższa funkcja, z tym że w razie braku strefy czasowej używa strefy lokalnej zamiast UTC.
- `parseRelativeTime(layout string, s string) time.Time`: Skrót dla `{{ .String "date" | parseTime "rfc3339" | toRelativeTime }}`.
- `add(a, b float) float`: Dodaje dwie liczby.
- `sub(a, b float) float`: Odejmuje dwie liczby.
- `mul(a, b float) float`: Mnoży dwie liczby.
- `div(a, b float) float`: Dzieli dwie liczby.
- `mod(a, b int) int`: Reszta z dzielenia `a` przez `b` (a % b).
- `formatApproxNumber(n int) string`: Formatuje liczbę w bardziej czytelny sposób, np. 1000 -> 1k.
- `formatNumber(n float|int) string`: Formatuje liczbę z separatorami tysięcy, np. 1000 -> 1,000.
- `trimPrefix(prefix string, str string) string`: Usuwa prefiks z ciągu znaków.
- `trimSuffix(suffix string, str string) string`: Usuwa sufiks z ciągu znaków.
- `trimSpace(str string) string`: Usuwa białe znaki z obu końców ciągu znaków.
- `replaceAll(old string, new string, str string) string`: Zastępuje wszystkie wystąpienia ciągu znaków w innym ciągu.
- `replaceMatches(pattern string, replacement string, str string) string`: Zastępuje wszystkie wystąpienia wyrażenia regularnego w ciągu znaków.
- `findMatch(pattern string, str string) string`: Znajduje pierwsze dopasowanie wyrażenia regularnego w ciągu znaków.
- `findSubmatch(pattern string, str string) string`: Znajduje pierwsze dopasowanie podgrup wyrażenia regularnego w ciągu znaków.
- `sortByString(key string, order string, arr []JSON): []JSON`: Sortuje tablicę obiektów JSON według klucza tekstowego w porządku rosnącym lub malejącym.
- `sortByInt(key string, order string, arr []JSON): []JSON`: Sortuje tablicę obiektów JSON według klucza całkowitego w porządku rosnącym lub malejącym.
- `sortByFloat(key string, order string, arr []JSON): []JSON`: Sortuje tablicę obiektów JSON według klucza zmiennoprzecinkowego w porządku rosnącym lub malejącym.
- `sortByTime(key string, layout string, order string, arr []JSON): []JSON`: Sortuje tablicę obiektów JSON według klucza czasu w porządku rosnącym lub malejącym. Format musi być zgodny z [formatem daty](https://pkg.go.dev/time#pkg-constants) Go.
- `concat(strings ...string) string`: Łączy wiele ciągów znaków w jeden.
- `unique(key string, arr []JSON) []JSON`: Zwraca tablicę unikalnych obiektów JSON na podstawie podanego klucza.
- `percentChange(current float, previous float) float`: Oblicza procentową zmianę między dwiema liczbami.
- `startOfDay(t time.Time) time.Time`: Zwraca początek dnia dla podanego czasu.
- `endOfDay(t time.Time) time.Time`: Zwraca koniec dnia dla podanego czasu.

Dostępne są następujące funkcje pomocnicze udostępniane przez `text/template` Go:

- `eq(a, b any) bool`: Porównuje dwie wartości pod kątem równości.
- `ne(a, b any) bool`: Porównuje dwie wartości pod kątem nierówności.
- `lt(a, b any) bool`: Sprawdza, czy pierwsza wartość jest mniejsza od drugiej.
- `le(a, b any) bool`: Sprawdza, czy pierwsza wartość jest mniejsza lub równa drugiej.
- `gt(a, b any) bool`: Sprawdza, czy pierwsza wartość jest większa od drugiej.
- `ge(a, b any) bool`: Sprawdza, czy pierwsza wartość jest większa lub równa drugiej.
- `and(args ...bool) bool`: Zwraca `true`, jeśli **wszystkie** argumenty są prawdziwe; przyjmuje dwie lub więcej wartości logicznych.
- `or(args ...bool) bool`: Zwraca `true`, jeśli **którykolwiek** argument jest prawdziwy; przyjmuje dwie lub więcej wartości logicznych.
- `not(a bool) bool`: Zwraca przeciwieństwo wartości.
- `index(a any, b int) any`: Zwraca wartość pod podanym indeksem tablicy.
- `len(a any) int`: Zwraca długość tablicy.
- `printf(format string, a ...any) string`: Zwraca sformatowany ciąg znaków.
