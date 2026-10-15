# Widget OpenCode Go - Instrukcja użycia

Widget OpenCode Go wyświetla wykorzystanie subskrypcji [OpenCode Go](https://opencode.ai/) w trzech oknach: kroczącym (5 godzin), tygodniowym i miesięcznym. Dane pochodzą z oficjalnego API limitów OpenCode.

Dla każdego okna pokazywany jest procent wykorzystania oraz czas do resetu. Gdy wykorzystanie przekroczy próg ostrzeżenia (domyślnie 85%) lub okno ma status `rate-limited`, pasek zmienia kolor na ostrzegawczy.

Tytuł widgetu jest klikalny i domyślnie przenosi do [panelu OpenCode](https://opencode.ai/console). Możesz to zmienić standardową właściwością `title-url`.

## Konfiguracja

Aby skonfigurować widget, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: opencode-go
  api-key: ${OPENCODE_GO_KEY}   # Twój klucz API OpenCode Go
```

### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| api-key | string | tak | |
| url | string | nie | https://opencode.ai/zen/go/v1/usage |
| warning-threshold | number | nie | 85 |
| cache | duration | nie | 5m |

#### `api-key`
Twój klucz API OpenCode Go (ten sam, którego używasz do API). Wysyłany jest jako token `Bearer`. Zalecane jest podawanie go przez zmienną środowiskową lub sekret, np. `api-key: ${OPENCODE_GO_KEY}`.

#### `url`
Adres endpointu z danymi o wykorzystaniu. Zmieniaj tylko wtedy, gdy chcesz wskazać inny (np. własny lub za proxy) endpoint. Pełna ścieżka to `https://opencode.ai/zen/go/v1/usage`.

#### `warning-threshold`
Procent wykorzystania, od którego pasek danego okna zmienia kolor na negatywny (ostrzegawczy). Domyślnie `85`.

#### `cache`
Jak długo cache'owana jest odpowiedź przed ponownym pobraniem. Domyślnie `5m`.

## Klucz API

Widget korzysta z tego samego klucza, co API OpenCode Go. Klucz przekazywany jest w nagłówku:

```http
Authorization: Bearer <api-key>
```

Jeśli subskrypcja Go nie jest aktywna dla danego klucza, API zwróci błąd `403 EntitlementError`, który pojawi się w widgecie jako komunikat błędu.

## Przykładowa konfiguracja

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: opencode-go
            title: OpenCode Go
            api-key: ${OPENCODE_GO_KEY}
            warning-threshold: 85
            cache: 5m
```

## Rozwiązywanie problemów

### Widget nie wyświetla danych

1. Sprawdź, czy `api-key` jest ustawiony i poprawny.
2. Upewnij się, że serwer, na którym działa Glance, ma dostęp sieciowy do `https://opencode.ai/zen/go/v1/usage`.
3. Sprawdź, czy subskrypcja OpenCode Go jest aktywna dla użytego klucza.

### Widok „Unauthorized” / „EntitlementError”

- `Unauthorized` (HTTP 401) oznacza nieprawidłowy lub nieistniejący klucz API.
- `EntitlementError` (HTTP 403, „OpenCode Go subscription required”) oznacza, że klucz nie ma aktywnej subskrypcji Go.

### Wartości różnią się od panelu OpenCode

Endpoint API zaokrągla procenty w dół (`floor`), a panel webowy pokazuje jedną cyfrę po przecinku, więc wynik może różnić się o około 1 punkt procentowy. Jest to zgodne z działaniem samego API.
