<a id="configuring-glance"></a>
# Konfiguracja Glance

- [Gotowa strona](#preconfigured-page)
- [Plik konfiguracyjny](#the-config-file)
  - [Automatyczne przeładowywanie](#auto-reload)
  - [Zmienne środowiskowe](#environment-variables)
    - [Inne sposoby przekazywania tokenów/haseł/sekretów](#other-ways-of-providing-tokenspasswordssecrets)
  - [Dołączanie innych plików konfiguracyjnych](#including-other-config-files)
  - [Ikony](#icons)
  - [Schemat konfiguracji](#config-schema)
- [Uwierzytelnianie](#authentication)
- [Serwer](#server)
- [Dokument](#document)
- [Branding](#branding)
- [Motyw](#theme)
  - [Dostępne motywy](#available-themes)
- [Strony i kolumny](#pages--columns)
- [Widgety](#widgets)
  - [RSS](#rss)
  - [Wideo](#videos)
  - [Hacker News](#hacker-news)
  - [Lobsters](#lobsters)
  - [Reddit](#reddit)
  - [Wyszukiwarka](#search-widget)
  - [Group](#group)
  - [Split Column](#split-column)
  - [Custom API](#custom-api)
  - [OpenCode Go](#opencode-go)
  - [Extension](#extension)
  - [Pogoda](#weather)
  - [Todo](#todo)
  - [Vikunja](#vikunja)
  - [Monitor](#monitor)
  - [Releases](#releases)
  - [Docker Containers](#docker-containers)
  - [qBittorrent](#qbittorrent)
  - [Google Compute Engine](#google-compute-engine)
  - [Tailscale](#tailscale)
  - [DNS Stats](#dns-stats)
  - [Cloudflare](#cloudflare)
  - [Server Stats](#server-stats)
  - [Beszel](#beszel)
  - [GitHub](#github)
  - [Repository](#repository)
  - [Radyjko](#radyjko)
  - [Navidrome](#navidrome)
  - [Zakładki](#bookmarks)
  - [Kalendarz](#calendar)
  - [Kalendarz (legacy)](#calendar-legacy)
  - [ChangeDetection.io](#changedetectionio)
  - [Zegar](#clock)
  - [Markets](#markets)
  - [Twitch Channels](#twitch-channels)
  - [Twitch Top Games](#twitch-top-games)
  - [iframe](#iframe)
  - [HTML](#html)
  - [Debugowanie](#debugowanie)


<a id="preconfigured-page"></a>
## Gotowa strona
Jeśli nie chcesz spędzać czasu na czytaniu wszystkich dostępnych opcji konfiguracji i chcesz szybko zacząć, możesz użyć [tego pliku `glance.yml`](glance.yml) i zmienić go według własnych potrzeb. Otrzymasz stronę wyglądającą następująco:

![](images/preconfigured-page-preview.png)

Skonfiguruj widgety, dodaj ich więcej, dodaj dodatkowe strony itp. Zrób ją swoją!

<a id="the-config-file"></a>
## Plik konfiguracyjny

<a id="auto-reload"></a>
### Automatyczne przeładowywanie
Obsługiwane jest automatyczne przeładowywanie konfiguracji, co oznacza, że możesz zmieniać plik konfiguracyjny, a zmiany zostaną zastosowane po zapisaniu, bez konieczności restartowania kontenera/usługi. Zmiany zmiennych środowiskowych nie powodują przeładowania i wymagają ręcznego restartu. Usunięcie pliku konfiguracyjnego spowoduje zaprzestanie obserwowania tego pliku, nawet jeśli zostanie on ponownie utworzony.

> [!NOTE]
>
> Jeśli spróbujesz uruchomić Glance z nieprawidłową konfiguracją, zakończy się ona błędem. Jeśli pomyślnie uruchomiłeś Glance z poprawną konfiguracją, a następnie wprowadziłeś zmiany powodujące błąd, zobaczysz ten błąd w konsoli, a Glance będzie dalej działać ze starą konfiguracją. Możesz wtedy kontynuować zmiany, a gdy nie będzie błędów, nowa konfiguracja zostanie załadowana.

> [!CAUTION]
>
> Przeładowanie pliku konfiguracyjnego czyści pamięć podręczną danych, co oznacza, że dane trzeba pobrać od nowa za każdym razem. Przy zbyt częstym przeładowywaniu może to prowadzić do limitów zapytań w niektórych API. Trwały cache między przeładowaniami zostanie dodany w przyszłości.

<a id="environment-variables"></a>
### Zmienne środowiskowe
Wstawianie zmiennych środowiskowych jest obsługiwane w dowolnym miejscu konfiguracji. Odbywa się to za pomocą składni `${ENV_VAR}`. Użycie nieistniejącej zmiennej środowiskowej spowoduje błąd i Glance albo nie uruchomi się, albo nie załaduje nowej konfiguracji po zapisaniu. Przykład:

```yaml
server:
  host: ${HOST}
  port: ${PORT}
```

Można jej również użyć w środku ciągu znaków:

```yaml
- type: rss
  title: ${RSS_TITLE}
  feeds:
    - url: http://domain.com/rss/${RSS_CATEGORY}.xml
```

Działa z każdym typem wartości, nie tylko z ciągami znaków:

```yaml
- type: rss
  limit: ${RSS_LIMIT}
```

Jeśli musisz użyć składni `${NAME}` w konfiguracji bez interpretowania jej jako zmiennej środowiskowej, możesz ją poprzedzić znakiem ucieczki `\`:

```yaml
something: \${NOT_AN_ENV_VAR}
```

<a id="other-ways-of-providing-tokenspasswordssecrets"></a>
#### Inne sposoby przekazywania tokenów/haseł/sekretów

Możesz używać [Docker secrets](https://docs.docker.com/compose/how-tos/use-secrets/) za pomocą następującej składni:

```yaml
# Zostanie to zastąpione zawartością pliku /run/secrets/github_token
# pod warunkiem, że sekret `github_token` jest przekazany do kontenera
token: ${secret:github_token}
```

Alternatywnie możesz załadować zawartość pliku, którego ścieżka jest przekazywana przez zmienną środowiskową:

`docker-compose.yml`
```yaml
services:
  glance:
    image: glanceapp/glance
    environment:
      - TOKEN_FILE=/home/user/token
    volumes:
      - /home/user/token:/home/user/token
```

`glance.yml`
```yaml
token: ${readFileFromEnv:TOKEN_FILE}
```

> [!NOTE]
>
> Białe znaki na początku i końcu zawartości pliku zostaną usunięte przed użyciem.

<a id="including-other-config-files"></a>
### Dołączanie innych plików konfiguracyjnych
Obsługiwane jest dołączanie plików konfiguracyjnych do głównego pliku konfiguracyjnego. Odbywa się to za pomocą dyrektywy `$include` wraz ze względną lub bezwzględną ścieżką do pliku, który chcesz dołączyć. Jeśli ścieżka jest względna, będzie liczona względem głównego pliku konfiguracyjnego. Dodatkowo w dołączanych plikach można używać zmiennych środowiskowych, a zmiany w dołączonych plikach powodują automatyczne przeładowanie. Przykład:

```yaml
pages:
  - $include: home.yml
  - $include: videos.yml
  - $include: homelab.yml
```

Dołączany plik nie powinien mieć dodatkowego wcięcia — jego wartości powinny znajdować się na najwyższym poziomie, a odpowiednie wcięcie zostanie dodane automatycznie w zależności od miejsca dołączenia pliku. Przykład:

`glance.yml`

```yaml
pages:
  - name: Home
    columns:
      - size: full
        widgets:
          $include: rss.yml
  - name: News
    columns:
      - size: full
        widgets:
          - type: group
            widgets:
              $include: rss.yml
              - type: reddit
                subreddit: news
```

`rss.yml`

```yaml
- type: rss
  title: News
  feeds:
    - url: ${RSS_URL}
```

Dyrektywy `$include` można użyć w dowolnym miejscu pliku konfiguracyjnego, nie tylko we właściwości `pages`, jednak musi ona znajdować się w osobnym wierszu i mieć odpowiednie wcięcie.

Jeśli podczas używania dyrektywy `$include` napotkasz błędy parsowania YAML, podane numery wierszy prawdopodobnie będą nieprawidłowe. Wynika to z faktu, że dołączanie plików odbywa się przed sparsowaniem YAML, ponieważ sam YAML nie obsługuje dołączania plików. Aby ułatwić debugowanie w takich przypadkach, możesz użyć polecenia `config:print` i przekazać jego wynik do `less -N`, aby zobaczyć pełny plik konfiguracyjny z rozwiązanymi dołączeniami i numerami wierszy:

```sh
glance --config /path/to/glance.yml config:print | less -N
```

Jest to nieco bardziej skomplikowane, gdy uruchamiasz Glance w kontenerze Docker:

```sh
docker run --rm -v ./glance.yml:/app/config/glance.yml glanceapp/glance config:print | less -N
```

Zakłada się, że konfiguracja, którą chcesz wyświetlić, znajduje się w bieżącym katalogu roboczym i nazywa się `glance.yml`.

<a id="icons"></a>
## Ikony

W widgetach, które umożliwiają określenie ikon, takich jak monitor, zakładki, kontenery Docker itp., możesz użyć właściwości `icon`, aby podać adres URL obrazu lub nazwę ikony z wielu bibliotek za pomocą prefiksów:

```yml
icon: si:immich # si dla Simple icons https://simpleicons.org/
icon: sh:immich # sh dla ikon selfh.st https://selfh.st/icons/
icon: di:immich # di dla Dashboard icons https://github.com/homarr-labs/dashboard-icons
icon: mdi:camera # mdi dla ikon Material Design https://pictogrammers.com/library/mdi/
```

> [!NOTE]
>
> Ikony są ładowane z zewnątrz i hostowane na `cdn.jsdelivr.net`. Jeśli nie chcesz zależeć od strony trzeciej, możesz pobrać ikony osobno i hostować je lokalnie.

Ikony z biblioteki Simple icons oraz ikony Material Design automatycznie odwracają kolor, aby dopasować się do jasnego lub ciemnego motywu. W przypadku innych ikon możesz włączyć to ręcznie, używając prefiksu `auto-invert`:

```yaml
icon: auto-invert https://example.com/path/to/icon.png # z adresem URL
icon: auto-invert sh:glance-dark # z ikoną selfh.st
```

Zakłada się, że ikona jest czarna, i zostanie automatycznie odwrócona na białą podczas korzystania z ciemnego motywu.

<a id="config-schema"></a>
## Schemat konfiguracji

Aby uzyskać opisy właściwości, walidację i autouzupełnianie konfiguracji w swoim IDE, @not-first uprzejmie stworzył [schemat](https://github.com/not-first/glance-schema). Ogromne podziękowania dla niego — sprawdź go i zostaw gwiazdkę!

<a id="authentication"></a>
## Uwierzytelnianie

Aby upewnić się, że tylko Ty i osoby, z którymi chcesz udostępnić pulpit, mają do niego dostęp, możesz skonfigurować uwierzytelnianie za pomocą nazwy użytkownika i hasła. Odbywa się to przez właściwość najwyższego poziomu `auth`. Przykład:

```yaml
auth:
  secret-key: # musi to być losowa wartość wygenerowana poleceniem CLI secret:make
  users:
    admin:
      password: 123456
    svilen:
      password: 123456
```

Aby wygenerować klucz sekretny, uruchom następujące polecenie:

```sh
./glance secret:make
```

Lub z Dockerem:

```sh
docker run --rm glanceapp/glance secret:make
```

### Korzystanie z hashy haseł

Jeśli nie chcesz przechowywać haseł w postaci jawnej w pliku konfiguracyjnym ani w zmiennych środowiskowych, możesz zahashować hasło i podać jego hash:

```sh
./glance password:hash mysecretpassword
```

Lub z Dockerem:

```sh
docker run --rm glanceapp/glance password:hash mysecretpassword
```

Następnie w pliku konfiguracyjnym użyj właściwości `password-hash` zamiast `password`:

```yaml
auth:
  secret-key: # musi to być losowa wartość wygenerowana poleceniem CLI secret:make
  users:
    admin:
      password-hash: $2a$10$o6SXqiccI3DDP2dN4ADumuOeIHET6Q4bUMYZD6rT2Aqt6XQ3DyO.6
```

### Zapobieganie atakom brute-force

Glance automatycznie blokuje adresy IP użytkowników, którzy 5 razy z rzędu nie zdołają się uwierzytelnić w ciągu 5 minut. Aby ta funkcja działała poprawnie, Glance musi znać rzeczywisty adres IP żądań. Jeśli używasz odwrotnego proxy, takiego jak nginx, Traefik, NPM itp., musisz ustawić właściwość `proxied` w konfiguracji `server` na `true`:

```yaml
server:
  proxied: true
```

Po ustawieniu na `true` Glance będzie używać nagłówka `X-Forwarded-For` do określenia oryginalnego adresu IP żądania, więc upewnij się, że Twoje odwrotne proxy jest poprawnie skonfigurowane do wysyłania tego nagłówka.

<a id="server"></a>
## Serwer
Konfiguracja serwera odbywa się przez właściwość najwyższego poziomu `server`. Przykład:

```yaml
server:
  port: 8080
  assets-path: /home/user/glance-assets
```

### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| host | string | nie |  |
| port | number | nie | 8080 |
| proxied | boolean | nie | false |
| base-url | string | nie | |
| assets-path | string | nie |  |

#### `host`
Adres, na którym będzie nasłuchiwał serwer. Ustawienie go na `localhost` oznacza, że tylko maszyna, na której działa serwer, będzie miała dostęp do pulpitu. Domyślnie nasłuchuje na wszystkich interfejsach.

#### `port`
Liczba z zakresu od 1 do 65 535, o ile port nie jest już używany przez coś innego.

#### `proxied`
Ustaw na `true`, jeśli używasz odwrotnego proxy przed Glance. Spowoduje to, że Glance będzie używać nagłówków `X-Forwarded-*` do określenia oryginalnych szczegółów żądania.

#### `base-url`
Bazowy adres URL, pod którym hostowany jest Glance. Nie musisz go podawać, chyba że używasz odwrotnego proxy i hostujesz Glance w podkatalogu. W takim przypadku możesz ustawić tę wartość na `/glance` lub inną nazwę katalogu. Zwróć uwagę, że początkowy ukośnik (`/`) jest wymagany, chyba że podajesz pełną domenę i ścieżkę.

> [!IMPORTANT]
> Musisz usunąć prefiks `base-url` przed przekazaniem żądania do serwera Glance.
> W Caddy możesz to zrobić za pomocą [`handle_path`](https://caddyserver.com/docs/caddyfile/directives/handle_path) lub [`uri strip_prefix`](https://caddyserver.com/docs/caddyfile/directives/uri).

#### `assets-path`
Ścieżka do katalogu, który będzie udostępniany przez serwer pod ścieżką `/assets/`. Jest to przydatne w widgetach takich jak Monitor, gdzie trzeba podać adres URL ikony, a chcesz hostować wszystkie ikony samodzielnie zamiast wskazywać zewnętrzne źródło.

> [!IMPORTANT]
>
> Podczas instalacji przez Docker ścieżka będzie wskazywać pliki wewnątrz kontenera. Nie zapomnij zamontować ścieżki zasobów pod tę samą ścieżkę wewnątrz kontenera.
> Przykład:
>
> Jeśli Twoje zasoby znajdują się w:
> ```
> /home/user/glance-assets
> ```
>
> Powinieneś zamontować:
> ```
> /home/user/glance-assets:/app/assets
> ```
>
> A Twoja konfiguracja powinna zawierać:
> ```
> assets-path: /app/assets
> ```

##### Przykłady

Załóżmy, że masz katalog `glance-assets` z plikiem `gitea-icon.png` i podajesz ścieżkę zasobów w następujący sposób:

```yaml
assets-path: /home/user/glance-assets
```

Aby wskazać zasób ze ścieżki zasobów, użyj ścieżki `/assets/` w następujący sposób:

```yaml
icon: /assets/gitea-icon.png
```

<a id="document"></a>
## Dokument
Jeśli chcesz wstawić własny kod HTML do sekcji `<head>` dokumentu na wszystkich stronach, możesz to zrobić za pomocą właściwości `document`. Przykład:

```yaml
document:
  head: |
    <script src="/assets/custom.js"></script>
```

<a id="branding"></a>
## Branding
Możesz dostosować różne elementy brandingu za pomocą właściwości najwyższego poziomu `branding`. Przykład:

```yaml
branding:
  custom-footer: |
    <p>Powered by <a href="https://github.com/glanceapp/glance">Glance</a></p>
  logo-url: /assets/logo.png
  favicon-url: /assets/logo.png
  app-name: "My Dashboard"
  app-icon-url: "/assets/app-icon.png"
  app-background-color: "#151519"
```

### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| hide-footer | bool | nie | false |
| custom-footer | string | nie |  |
| logo-text | string | nie | G |
| logo-url | string | nie | |
| favicon-url | string | nie | |
| app-name | string | nie | Glance |
| app-icon-url | string | nie | Domyślna ikona Glance |
| app-background-color | string | nie | Domyślny kolor tła Glance |

#### `hide-footer`
Ukrywa stopkę po ustawieniu na `true`.

#### `custom-footer`
Określa własny kod HTML używany w stopce.

#### `logo-text`
Określa własny tekst używany zamiast litery „G” widocznej w nawigacji.

#### `logo-url`
Określa adres URL własnego obrazu używanego zamiast litery „G” widocznej w nawigacji. Jeśli ustawione są zarówno `logo-text`, jak i `logo-url`, używany będzie tylko `logo-url`.

#### `favicon-url`
Określa adres URL własnego obrazu używanego jako favikona.

#### `app-name`
Określa nazwę aplikacji webowej wyświetlaną na karcie przeglądarki i w PWA.

#### `app-icon-url`
Określa adres URL ikony PWA i karty przeglądarki (PNG 512x512).

#### `app-background-color`
Określa kolor tła PWA. Musi to być prawidłowy kolor CSS.

<a id="theme"></a>
## Motyw
Motyw konfiguruje się przez właściwość najwyższego poziomu `theme`. Wartości kolorów są w formacie [HSL](https://giggster.com/guide/basics/hue-saturation-lightness/) (odcień, nasycenie, jasność). Możesz użyć próbnika kolorów [takiego jak ten](https://hslpicker.com/), aby przekonwertować kolory z innych formatów na HSL. Wartości oddziela się spacją, a znak `%` nie jest wymagany przy żadnej z liczb.

Przykład:

```yaml
theme:
  # To będzie motyw domyślny
  background-color: 100 20 10
  primary-color: 40 90 40
  contrast-multiplier: 1.1

  disable-picker: false
  presets:
    gruvbox-dark:
      background-color: 0 0 16
      primary-color: 43 59 81
      positive-color: 61 66 44
      negative-color: 6 96 59

    zebra:
      light: true
      background-color: 0 0 95
      primary-color: 0 0 10
      negative-color: 0 90 50
```

<a id="available-themes"></a>
### Dostępne motywy
Jeśli nie chcesz tracić czasu na konfigurowanie własnego motywu, dostępnych jest [kilka gotowych motywów](themes.md), z których możesz po prostu skopiować wartości.

### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| light | boolean | nie | false |
| background-color | HSL | nie | 240 8 9 |
| primary-color | HSL | nie | 43 50 70 |
| positive-color | HSL | nie | tak samo jak `primary-color` |
| negative-color | HSL | nie | 0 70 70 |
| contrast-multiplier | number | nie | 1 |
| text-saturation-multiplier | number | nie | 1 |
| custom-css-file | string | nie | |
| disable-picker | bool | false | |
| presets | object | nie | |

#### `light`
Określa, czy schemat jest jasny, czy ciemny. Nie zmienia koloru tła — odwraca kolory tekstu, aby wyglądały odpowiednio na jasnym tle.

#### `background-color`
Kolor strony i widgetów.

#### `primary-color`
Kolor używany w całej witrynie, głównie do oznaczania nieodwiedzonych linków.

#### `positive-color`
Używany do oznaczania pozytywnych sytuacji, takich jak wzrost ceny akcji, aktywny kanał Twitch lub działająca monitorowana witryna. Jeśli nie zostanie ustawiony, użyty zostanie kolor `primary-color`.

#### `negative-color`
Przeciwieństwo `positive-color`.

#### `contrast-multiplier`
Służy do zwiększania lub zmniejszania kontrastu (czyli widoczności) tekstu. Wartość `1.3` oznacza, że tekst będzie o 30% jaśniejszy/ciemniejszy w zależności od schematu. Użyj tego, jeśli uważasz, że niektóre teksty na stronie są zbyt ciemne i trudne do odczytania. Przykład:

![różnica między kontrastem 1 i 1.3](images/contrast-multiplier-example.png)

#### `text-saturation-multiplier`
Służy do zwiększania lub zmniejszania nasycenia tekstu. Przydatne przy niestandardowym kolorze tła o wysokim nasyceniu, gdy tekst powinien mieć bardziej neutralny kolor. Wartość `0.5` oznacza, że nasycenie będzie o 50% niższe, a `1.5` — o 50% wyższe.

#### `custom-css-file`
Ścieżka do niestandardowego pliku CSS — zewnętrznego lub znajdującego się w skonfigurowanej ścieżce zasobów serwera. Przykład:

```yaml
theme:
  custom-css-file: /assets/my-style.css
```

> [!TIP]
>
> Ponieważ Glance używa wielu klas użytkowych, stylowanie niektórych elementów może być trudne. Aby ułatwić stylowanie konkretnych widgetów, każdy widget ma klasę `widget-type-{name}`. Na przykład, jeśli chcesz powiększyć linki tylko w widgecie RSS, możesz użyć następującego selektora:
>
> ```css
> .widget-type-rss a {
>     font-size: 1.5rem;
> }
> ```
>
> Dodatkowo możesz użyć właściwości `css-class`, dostępnej w każdym widgecie, aby ustawić niestandardowe nazwy klas dla poszczególnych widgetów.

#### `disable-picker`
Po ustawieniu na `true` ukrywa próbnik motywów i wyłącza możliwość przełączania się między motywami. Wszyscy użytkownicy, którzy wcześniej wybrali motyw inny niż domyślny, zostaną przełączeni na motyw domyślny.

#### `presets`
Definiuje dodatkowe presety motywów, które można wybrać z próbnika motywów na stronie. Dla każdego presetu możesz określić te same właściwości co dla motywu domyślnego, takie jak `background-color`, `primary-color`, `positive-color`, `negative-color`, `contrast-multiplier` itp., z wyjątkiem właściwości `custom-css-file`.

Przykład:

```yaml
theme:
  presets:
    my-custom-dark-theme:
      background-color: 229 19 23
      contrast-multiplier: 1.2
      primary-color: 222 74 74
      positive-color: 96 44 68
      negative-color: 359 68 71
    my-custom-light-theme:
      light: true
      background-color: 220 23 95
      contrast-multiplier: 1.1
      primary-color: 220 91 54
      positive-color: 109 58 40
      negative-color: 347 87 44
```

Aby nadpisać domyślne motywy ciemny i jasny, użyj nazw kluczy `default-dark` i `default-light`.

<a id="pages--columns"></a>
## Strony i kolumny
![ilustracja stron i kolumn](images/pages-and-columns-illustration.png)

Strony i kolumny służą do organizacji widgetów. Każda strona zawiera maksymalnie 3 kolumny, a każda kolumna może mieć dowolną liczbę widgetów.

### Strony
Strony definiuje się przez właściwość najwyższego poziomu `pages`. Strona zdefiniowana jako pierwsza staje się stroną główną, a wszystkie strony są automatycznie dodawane do paska nawigacji w kolejności ich definiowania. Przykład:

```yaml
pages:
  - name: Home
    columns: ...

  - name: Videos
    columns: ...

  - name: Homelab
    columns: ...
```

### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| name | string | tak | |
| slug | string | nie | |
| width | string | nie | |
| desktop-navigation-width | string | nie | |
| center-vertically | boolean | nie | false |
| hide-desktop-navigation | boolean | nie | false |
| show-mobile-header | boolean | nie | false |
| head-widgets | array | nie | |
| columns | array | tak | |

#### `name`
Nazwa strony wyświetlana na pasku nawigacji.

#### `slug`
Przyjazna dla adresów URL wersja tytułu, używana do uzyskania dostępu do strony. Na przykład, jeśli tytuł strony to „RSS Feeds”, możesz udostępnić stronę pod adresem `localhost:8080/feeds`, ustawiając slug na `feeds`. Jeśli nie zostanie zdefiniowany, zostanie automatycznie wygenerowany z tytułu.

#### `width`
Maksymalna szerokość strony na pulpicie. Możliwe wartości: `default`, `slim` i `wide`.

#### `desktop-navigation-width`
Maksymalna szerokość nawigacji na pulpicie. Przydatne, gdy masz kilka stron o innej szerokości niż pozostałe i nie chcesz, aby nawigacja gwałtownie się przeskakiwała podczas przechodzenia do tych stron i z powrotem. Możliwe wartości: `default`, `slim` i `wide`.

Odpowiedniki w pikselach dla każdej wartości:

* default: `1600px`
* slim: `1100px`
* wide: `1920px`

> [!NOTE]
>
> Przy użyciu `slim` maksymalna liczba kolumn dozwolonych na stronie wynosi `2`.

#### `center-vertically`
Po ustawieniu na `true` zawartość strony jest centrowana w pionie. Nie ma efektu, jeśli zawartość jest wyższa niż wysokość okna widoku.

#### `hide-desktop-navigation`
Określa, czy na pulpicie mają być wyświetlane linki nawigacyjne u góry strony.

#### `show-mobile-header`
Określa, czy na urządzeniach mobilnych ma być wyświetlany nagłówek z nazwą strony. Nagłówek celowo ma dużo pionowego odstępu, aby odsunąć zawartość w dół i ułatwić jej dosięgnięcie na wysokich urządzeniach.

Podgląd:

![](images/mobile-header-preview.png)

#### `head-widgets`

Widgety nagłówkowe będą wyświetlane u góry strony, nad kolumnami, i zajmą łączną szerokość wszystkich kolumn. Możesz podać dowolny widget, choć niektóre będą wyglądać lepiej niż inne — na przykład markets, kanał RSS w stylu `horizontal-cards` oraz widgety videos. Przykład:

![](images/head-widgets-preview.png)

```yaml
pages:
  - name: Home
    head-widgets:
      - type: markets
        hide-header: true
        markets:
          - symbol: SPY
            name: S&P 500
          - symbol: BTC-USD
            name: Bitcoin
          - symbol: NVDA
            name: NVIDIA
          - symbol: AAPL
            name: Apple
          - symbol: MSFT
            name: Microsoft

    columns:
      - size: small
        widgets:
          - type: calendar
      - size: full
        widgets:
          - type: hacker-news
      - size: small
        widgets:
          - type: weather
            location: London, United Kingdom
```

### Kolumny
Kolumny definiuje się dla każdej strony za pomocą właściwości `columns`. Istnieją dwa typy kolumn — `full` i `small` — które określają ich szerokość. Kolumna mała zajmuje stałą szerokość (300px), a kolumna pełna zajmuje całą pozostałą szerokość. Na stronie możesz mieć maksymalnie 3 kolumny i musisz mieć 1 lub 2 kolumny pełne. Przykład:

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets: ...
      - size: full
        widgets: ...
      - size: small
        widgets: ...
```

### Właściwości
| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| size | string | tak |
| widgets | array | nie |

Oto niektóre z możliwych konfiguracji kolumn:

![konfiguracja kolumn small-full-small](images/column-configuration-1.png)

```yaml
columns:
  - size: small
    widgets: ...
  - size: full
    widgets: ...
  - size: small
    widgets: ...
```

![konfiguracja kolumn small-full-small](images/column-configuration-2.png)

```yaml
columns:
  - size: full
    widgets: ...
  - size: small
    widgets: ...
```

![konfiguracja kolumn small-full-small](images/column-configuration-3.png)

```yaml
columns:
  - size: full
    widgets: ...
  - size: full
    widgets: ...
```

<a id="widgets"></a>
## Widgety
Widgety definiuje się dla każdej kolumny za pomocą właściwości `widgets`. Przykład:

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: weather
            location: London, United Kingdom
```

> [!NOTE]
>
> Obecnie nie wszystkie widgety są zaprojektowane tak, aby pasować do każdego rozmiaru kolumny, jednak niektóre oferują różne „style”, które pomagają obejść to ograniczenie.

### Wspólne właściwości
| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| type | string | tak |
| title | string | nie |
| title-url | string | nie |
| hide-header | boolean | nie | false |
| cache | string | nie |
| css-class | string | nie |

#### `type`
Służy do określenia widgetu.

#### `title`
Tytuł widgetu. Jeśli pozostanie pusty, zostanie określony przez widget.

#### `title-url`
Adres URL otwierany po kliknięciu tytułu widgetu. Jeśli pozostanie pusty, zostanie określony przez widget (jeśli jest dostępny).

#### `hide-header`
Po ustawieniu na `true` nagłówek (tytuł) widgetu zostanie ukryty. Nie można ukryć nagłówka widgetu group.

> [!NOTE]
>
> Jeśli widget nie zdoła się zaktualizować, obok jego tytułu pojawi się czerwona kropka lub okrąg informujący o tym, że nie działa. Nie będzie tego widać, jeśli nagłówek jest ukryty.

#### `cache`
Jak długo przechowywać pobrane dane w pamięci. Wartość jest ciągiem znaków i musi być liczbą, po której następuje jedna z liter s, m, h, d. Przykłady:

```yaml
cache: 30s # 30 sekund
cache: 5m  # 5 minut
cache: 2h  # 2 godziny
cache: 1d  # 1 dzień
```

> [!NOTE]
>
> Nie we wszystkich widgetach można zmienić czas cache. Widgety kalendarza i pogody aktualizują się o pełnej godzinie i nie można tego zmienić.

#### `css-class`
Ustawia niestandardowe klasy CSS dla konkretnej instancji widgetu.

<a id="rss"></a>
### RSS
Wyświetla listę artykułów z wielu kanałów RSS.

Przykład:

```yaml
- type: rss
  title: News
  style: horizontal-cards
  feeds:
    - url: https://feeds.bloomberg.com/markets/news.rss
      title: Bloomberg
    - url: https://moxie.foxbusiness.com/google-publisher/markets.xml
      title: Fox Business
    - url: https://moxie.foxbusiness.com/google-publisher/technology.xml
      title: Fox Business
```

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| style | string | nie | vertical-list |
| feeds | array | tak |
| thumbnail-height | float | nie | 10 |
| card-height | float | nie | 27 |
| limit | integer | nie | 25 |
| preserve-order | bool | nie | false |
| single-line-titles | boolean | nie | false |
| collapse-after | integer | nie | 5 |

##### `limit`
Maksymalna liczba artykułów do wyświetlenia.

##### `collapse-after`
Liczba widocznych artykułów, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

##### `preserve-order`
Po ustawieniu na `true` kolejność artykułów zostanie zachowana tak, jak w kanałach. Przydatne, jeśli kanał używa własnej kolejności sortowania oznaczającej ważność artykułów. Jeśli używasz tej właściwości przy dużej liczbie kanałów, zaleca się ustawienie `limit` dla każdego kanału osobno, ponieważ jeśli pierwszy zdefiniowany kanał ma 15 artykułów, artykuły z drugiego kanału zaczną się po 15. artykule na liście.

##### `single-line-titles`
Po ustawieniu na `true` obcina tytuł każdego wpisu, jeśli przekracza jedną linię. Ma zastosowanie tylko wtedy, gdy styl jest ustawiony na `vertical-list`.

##### `style`
Służy do zmiany wyglądu widgetu. Możliwe wartości:

* `vertical-list` — odpowiedni dla kolumn `full` i `small`
* `detailed-list` — odpowiedni dla kolumn `full`
* `horizontal-cards` — odpowiedni dla kolumn `full`
* `horizontal-cards-2` — odpowiedni dla kolumn `full`

Poniżej znajduje się podgląd każdego stylu:

`vertical-list`

![podgląd stylu vertical-list dla widgetu RSS](images/rss-feed-vertical-list-preview.png)

`detailed-list`

![podgląd stylu detailed-list dla widgetu RSS](images/rss-widget-detailed-list-preview.png)

`horizontal-cards`

![podgląd stylu horizontal-cards dla widgetu RSS](images/rss-feed-horizontal-cards-preview.png)

`horizontal-cards-2`

![podgląd stylu horizontal-cards-2 dla widgetu RSS](images/rss-widget-horizontal-cards-2-preview.png)

##### `thumbnail-height`
Służy do zmiany wysokości miniatur. Działa tylko wtedy, gdy styl jest ustawiony na `horizontal-cards`. Wartość domyślna to `10`, a jednostką jest `rem`. Jeśli chcesz na przykład dwukrotnie zwiększyć wysokość miniatur, ustaw ją na `20`.

##### `card-height`
Służy do zmiany wysokości kart w stylu `horizontal-cards-2`. Wartość domyślna to `27`, a jednostką jest `rem`.

##### `feeds`
Tablica kanałów RSS/Atom. Tytuł można opcjonalnie zmienić.

###### Właściwości każdego kanału
| Nazwa | Typ | Wymagane | Domyślnie | Uwagi |
| ---- | ---- | -------- | ------- | ----- |
| url | string | tak | | |
| title | string | nie | tytuł podany przez kanał | |
| hide-categories | boolean | nie | false | Dotyczy tylko stylu `detailed-list` |
| hide-description | boolean | nie | false | Dotyczy tylko stylu `detailed-list` |
| limit | integer | nie | | |
| item-link-prefix | string | nie | | |
| headers | klucz (string) i wartość (string) | nie | | |

###### `limit`
Maksymalna liczba artykułów do wyświetlenia z tego konkretnego kanału. Przydatne, jeśli masz kanał publikujący często wiele artykułów i chcesz zapobiec nadmiernemu spychaniu artykułów z innych kanałów.

###### `item-link-prefix`
Jeśli kanał RSS nie zwraca linków do wpisów z domeną bazową, a Glance nie zdołał automatycznie wykryć prawidłowej domeny, możesz ręcznie dodać prefiks do każdego linku za pomocą tej właściwości.

###### `headers`
Opcjonalnie określa nagłówki wysyłane z żądaniem. Przykład:

```yaml
- type: rss
  feeds:
    - url: https://domain.com/rss
      headers:
        User-Agent: Custom User Agent
```

<a id="videos"></a>
### Wideo
Wyświetla listę najnowszych filmów z określonych kanałów YouTube.

Przykład:

```yaml
- type: videos
  channels:
    - UCXuqSBlHAE6Xw-yeJA0Tunw
    - UCBJycsmduvYEL83R_U4JriQ
    - UCHnyfMqiRRG1u-2MsSQLbXA
```

Podgląd:
![](images/videos-widget-preview.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| channels | array | tak | |
| playlists | array | nie | |
| limit | integer | nie | 25 |
| style | string | nie | horizontal-cards |
| collapse-after | integer | nie | 7 |
| collapse-after-rows | integer | nie | 4 |
| include-shorts | boolean | nie | false |
| video-url-template | string | nie | https://www.youtube.com/watch?v={VIDEO-ID} |

##### `channels`
Lista identyfikatorów kanałów.

Jednym ze sposobów uzyskania identyfikatora kanału jest przejście na stronę kanału i kliknięcie jego opisu:

![](images/videos-channel-description-example.png)

Następnie przewiń w dół, kliknij „Share channel”, a potem „Copy channel ID”:

![](images/videos-copy-channel-id-example.png)

##### `playlists`

Lista identyfikatorów playlist:

```yaml
- type: videos
  playlists:
    - PL8mG-RkN2uTyZZ00ObwZxxoG_nJbs3qec
    - PL8mG-RkN2uTxTK4m_Vl2dYR9yE41kRdBg
```

Identyfikator playlisty można znaleźć w jej linku, który ma postać:
```
https://www.youtube.com...&list={ID}&...
```

##### `limit`
Maksymalna liczba filmów do wyświetlenia.

##### `collapse-after`
Określa liczbę filmów wyświetlanych w stylu `vertical-list` przed pojawieniem się przycisku „SHOW MORE”.

##### `collapse-after-rows`
Określa liczbę wierszy wyświetlanych w stylu `grid-cards` przed pojawieniem się przycisku „SHOW MORE”.

##### `style`
Służy do zmiany wyglądu widgetu. Możliwe wartości: `horizontal-cards`, `vertical-list` i `grid-cards`.

Podgląd `vertical-list`:

![](images/videos-widget-vertical-list-preview.png)

Podgląd `grid-cards`:

![](images/videos-widget-grid-cards-preview.png)

##### `video-url-template`
Służy do zastąpienia domyślnego linku do filmów. Przydatne, gdy używasz własnego frontendu YouTube. Przykład:

```yaml
video-url-template: https://invidious.your-domain.com/watch?v={VIDEO-ID}
```

Placeholdery:

`{VIDEO-ID}` — identyfikator filmu

<a id="hacker-news"></a>
### Hacker News
Wyświetla listę wpisów z [Hacker News](https://news.ycombinator.com/).

Przykład:

```yaml
- type: hacker-news
  limit: 15
  collapse-after: 5
```

Podgląd:
![](images/hacker-news-widget-preview.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| limit | integer | nie | 15 |
| collapse-after | integer | nie | 5 |
| comments-url-template | string | nie | https://news.ycombinator.com/item?id={POST-ID} |
| sort-by | string | nie | top |
| extra-sort-by | string | nie | |

##### `comments-url-template`
Służy do zastąpienia domyślnego linku do komentarzy wpisu. Przydatne, jeśli chcesz użyć alternatywnego frontendu. Przykład:

```yaml
comments-url-template: https://www.hckrnws.com/stories/{POST-ID}
```

Placeholdery:

`{POST-ID}` — identyfikator wpisu

##### `sort-by`
Służy do określenia kolejności zwracanych wpisów. Możliwe wartości: `top`, `new` i `best`.

##### `extra-sort-by`
Może służyć do określenia dodatkowego sortowania, które zostanie zastosowane na już posortowanych wpisach. Domyślnie nie stosuje się żadnego dodatkowego sortowania; jedyną dostępną opcją jest `engagement`.

Sortowanie `engagement` stara się umieścić na górze wpisy z największą liczbą punktów i komentarzy, preferując również nowsze wpisy nad starszymi.

<a id="lobsters"></a>
### Lobsters
Wyświetla listę wpisów z [Lobsters](https://lobste.rs).

Przykład:

```yaml
- type: lobsters
  sort-by: hot
  tags:
    - go
    - security
    - linux
  limit: 15
  collapse-after: 5
```

Podgląd:
![](images/lobsters-widget-preview.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| instance-url | string | nie | https://lobste.rs/ |
| custom-url | string | nie | |
| limit | integer | nie | 15 |
| collapse-after | integer | nie | 5 |
| sort-by | string | nie | hot |
| tags | array | nie | |

##### `instance-url`
Bazowy adres URL instancji Lobsters hostowanej gdzie indziej niż na lobste.rs. Przykład:

```yaml
instance-url: https://www.journalduhacker.net/
```

##### `custom-url`
Niestandardowy adres URL do pobierania wpisów Lobsters. Jeśli zostanie podany, właściwości `instance-url`, `sort-by` i `tags` są ignorowane.

##### `limit`
Maksymalna liczba wpisów do wyświetlenia.

##### `collapse-after`
Liczba widocznych wpisów, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

##### `sort-by`
Kolejność sortowania zwracanych wpisów. Możliwe opcje: `hot` i `new`.

##### `tags`
Ogranicza wpisy do zawierających jeden z podanych tagów. **Nie można określić kolejności sortowania podczas filtrowania po tagach — domyślnie zostanie użyte `hot`.**

<a id="reddit"></a>
### Reddit
Wyświetla listę wpisów z określonego subreddita.

> [!WARNING]
>
> Reddit nie zezwala na nieuwierzytelniony dostęp do API z adresów IP VPS. Jeśli hostujesz Glance na VPS, otrzymasz odpowiedź 403.
> Jako obejście możesz [zarejestrować aplikację na Reddicie](https://ssl.reddit.com/prefs/apps/) i użyć wygenerowanego ID oraz sekretu w konfiguracji widgetu, aby uwierzytelnić żądania (patrz właściwość `app-auth`), użyć proxy
> (patrz właściwość `proxy`) albo skierować ruch z Glance przez VPN.

Przykład:

```yaml
- type: reddit
  subreddit: technology
```

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| subreddit | string | tak |  |
| style | string | nie | vertical-list |
| show-thumbnails | boolean | nie | false |
| show-flairs | boolean | nie | false |
| limit | integer | nie | 15 |
| collapse-after | integer | nie | 5 |
| comments-url-template | string | nie | https://www.reddit.com/{POST-PATH} |
| request-url-template | string | nie |  |
| proxy | string lub wiele parametrów | nie |  |
| sort-by | string | nie | hot |
| top-period | string | nie | day |
| search | string | nie | |
| extra-sort-by | string | nie | |
| app-auth | object | nie | |

##### `subreddit`
Subreddit, z którego mają być pobierane wpisy.

##### `style`
Służy do zmiany wyglądu widgetu. Możliwe wartości: `vertical-list`, `horizontal-cards` i `vertical-cards`. Pierwsze dwa są przeznaczone dla pełnych kolumn, a ostatni dla małych kolumn.

`vertical-list`

![](images/reddit-widget-preview.png)

`horizontal-cards`

![](images/reddit-widget-horizontal-cards-preview.png)

`vertical-cards`

![](images/reddit-widget-vertical-cards-preview.png)

##### `show-thumbnails`
Pokazuje lub ukrywa miniatury obok wpisu. Działa tylko wtedy, gdy `style` ma wartość `vertical-list`. Podgląd:

![](images/reddit-widget-vertical-list-thumbnails.png)

> [!NOTE]
>
> Miniatury nie działają w przypadku niektórych subredditów, ponieważ API Reddita nie zwraca adresu URL miniatury. Nie ma na to jeszcze obejścia.

##### `show-flairs`
Pokazuje oznaczenia wpisów (flairy) po ustawieniu na `true`.

##### `limit`
Maksymalna liczba wpisów do wyświetlenia.

##### `collapse-after`
Liczba widocznych wpisów, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać. Niedostępne w stylach `vertical-cards` i `horizontal-cards`.

##### `comments-url-template`
Służy do zastąpienia domyślnego linku do komentarzy wpisu. Przydatne, jeśli chcesz użyć starego wyglądu Reddita lub innego frontendu. Przykład:

```yaml
comments-url-template: https://old.reddit.com/{POST-PATH}
```

Placeholdery:

`{POST-PATH}` — pełna ścieżka do wpisu, na przykład:

```
r/selfhosted/comments/bsp01i/welcome_to_rselfhosted_please_read_this_first/
```

`{POST-ID}` — identyfikator znajdujący się po `/comments/`

`{SUBREDDIT}` — nazwa subreddita

##### `request-url-template`
Niestandardowy adres URL żądania używany do pobierania danych. Przydatne, gdy hostujesz Glance na VPS, gdzie Reddit blokuje żądania, i chcesz skierować je przez proxy akceptujące adres URL jako część ścieżki lub parametr zapytania.

Placeholdery:

`{REQUEST-URL}` — zostanie zastąpiony rozwiniętym adresem URL żądania (np. https://www.reddit.com/r/selfhosted/hot.json). Przykład:

```
https://proxy/{REQUEST-URL}
https://your.proxy/?url={REQUEST-URL}
```

##### `proxy`
Niestandardowy adres URL proxy HTTP/HTTPS używany do pobierania danych. Przydatne, gdy hostujesz Glance na VPS, gdzie Reddit blokuje żądania, i chcesz obejść ograniczenie, kierując żądania przez proxy. Przykład:

```yaml
proxy: http://user:pass@proxy.com:8080
proxy: https://user:pass@proxy.com:443
```

Alternatywnie możesz podać adres URL proxy oraz dodatkowe opcje za pomocą wielu parametrów:

```yaml
proxy:
  url: http://proxy.com:8080
  allow-insecure: true
  timeout: 10s
```

###### `allow-insecure`
Po ustawieniu na `true` zezwala na używanie niezabezpieczonych połączeń, na przykład gdy proxy ma certyfikat z podpisem własnym.

###### `timeout`
Maksymalny czas oczekiwania na odpowiedź proxy. Wartość jest ciągiem znaków i musi być liczbą, po której następuje jedna z liter s, m, h, d. Przykład: `10s` dla 10 sekund, `1m` dla 1 minuty itp.

##### `sort-by`
Może służyć do określenia kolejności zwracanych wpisów. Możliwe wartości: `hot`, `new`, `top` i `rising`.

##### `top-period`
Dostępne tylko wtedy, gdy `sort-by` jest ustawione na `top`. Możliwe wartości: `hour`, `day`, `week`, `month`, `year` i `all`.

##### `search`
Słowa kluczowe do wyszukania. Możliwe jest również wyszukiwanie w określonych polach, **pamiętaj jednak, że Reddit może w każdej chwili usunąć możliwość korzystania z nich**:

![](images/reddit-field-search.png)

##### `extra-sort-by`
Może służyć do określenia dodatkowego sortowania, które zostanie zastosowane na już posortowanych wpisach. Domyślnie nie stosuje się żadnego dodatkowego sortowania; jedyną dostępną opcją jest `engagement`.

Sortowanie `engagement` stara się umieścić na górze wpisy z największą liczbą punktów i komentarzy, preferując również nowsze wpisy nad starszymi.

##### `app-auth`
```yaml
widgets:
  - type: reddit
    subreddit: technology
    app-auth:
      name: ${REDDIT_APP_NAME}
      id: ${REDDIT_APP_CLIENT_ID}
      secret: ${REDDIT_APP_SECRET}
```

Aby zarejestrować aplikację na Reddicie, przejdź na [tę stronę](https://ssl.reddit.com/prefs/apps/).

<a id="search-widget"></a>
### Wyszukiwarka
Wyświetla pasek wyszukiwania, który można wykorzystać do wyszukiwania określonych fraz w różnych wyszukiwarkach.

Przykład:

```yaml
- type: search
  search-engine: duckduckgo
  recent-bangs-count: 5
  hide-dropdown: false
  bangs:
    - title: YouTube
      shortcut: "!yt"
      url: https://www.youtube.com/results?search_query={QUERY}
      icon: https://youtube.com/favicon.ico
      subtitle: Search on YouTube
```

Podgląd:

![](images/search-widget-preview.png)

#### Skróty klawiszowe
| Klawisze | Akcja | Warunek |
| ---- | ------ | --------- |
| <kbd>S</kbd> | Ustawia fokus na pasku wyszukiwania | Fokus nie jest już ustawiony na innym polu wprowadzania |
| <kbd>Enter</kbd> | Wykonuje wyszukiwanie w tej samej karcie | Pole wyszukiwania ma fokus i nie jest puste |
| <kbd>Ctrl</kbd> + <kbd>Enter</kbd> | Wykonuje wyszukiwanie w nowej karcie | Pole wyszukiwania ma fokus i nie jest puste |
| <kbd>Escape</kbd> | Opuszcza pole | Pole wyszukiwania ma fokus |
| <kbd>Up</kbd> | Wstawia do pola ostatnie zapytanie od otwarcia strony | Pole wyszukiwania ma fokus |

> [!TIP]
>
> Możesz użyć właściwości `new-tab` z wartością `true`, jeśli chcesz domyślnie wyświetlać wyniki wyszukiwania w nowej karcie. Wtedy <kbd>Ctrl</kbd> + <kbd>Enter</kbd> wyświetli wyniki w tej samej karcie.

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| search-engine | string | nie | duckduckgo |
| new-tab | boolean | nie | false |
| autofocus | boolean | nie | false |
| target | string | nie | _blank |
| placeholder | string | nie | Type here to search… |
| hide-dropdown | boolean | nie | false |
| recent-bangs-count | number | nie | 3 |
| bangs | array | nie | |

##### `search-engine`
Wartość z poniższej tabeli lub adres URL niestandardowej wyszukiwarki. Użyj `{QUERY}`, aby wskazać miejsce wstawienia zapytania.

| Nazwa | URL |
| ---- | --- |
| duckduckgo | `https://duckduckgo.com/?q={QUERY}` |
| google | `https://www.google.com/search?q={QUERY}` |
| bing | `https://www.bing.com/search?q={QUERY}` |
| perplexity | `https://www.perplexity.ai/search?q={QUERY}` |
| kagi | `https://kagi.com/search?q={QUERY}` |
| startpage | `https://www.startpage.com/search?q={QUERY}` |

##### `new-tab`
Po ustawieniu na `true` zamienia skróty odpowiedzialne za wyświetlanie wyników w tej samej lub nowej karcie, domyślnie otwierając wyniki w nowej karcie.

##### `autofocus`
Po ustawieniu na `true` automatycznie ustawia fokus na polu wyszukiwania po załadowaniu strony.

##### `target`
Wartość atrybutu `target` używana przy otwieraniu wyników wyszukiwania w nowej karcie. Możliwe wartości: `_blank`, `_self`, `_parent` i `_top`.

##### `placeholder`
Po ustawieniu zmienia tekst wyświetlany w polu przed wpisaniem zapytania.

##### `hide-dropdown`
Po ustawieniu na `true` ukrywa listę rozwijaną z dostępnymi bangami, gdy pole wyszukiwania ma fokus. Po ustawieniu na `false` lista jest widoczna. Domyślnie `false`.

##### `recent-bangs-count`
Określa, ile ostatnio używanych bangów pojawia się w sekcji „Recent” listy rozwijanej. Domyślnie `3`. Maksymalnie 10. Po ustawieniu na `0` sekcja „Recent” jest całkowicie ukryta.

##### `bangs`
Co teraz? [Bangi](https://duckduckgo.com/bangs). To skróty, które pozwalają używać tego samego pola wyszukiwania dla wielu różnych witryn. Zakładając, że masz je skonfigurowane, jeśli na przykład rozpoczniesz wpisywanie od `!yt`, będziesz mógł wyszukać coś na YouTube:

![](images/search-widget-bangs-preview.png)

##### Właściwości każdego bangu
| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| title | string | nie |
| subtitle | string | nie |
| shortcut | string | tak |
| url | string | tak |
| icon | string | nie |

###### `title`
Opcjonalny tytuł, który pojawi się po prawej stronie paska wyszukiwania, gdy zapytanie rozpoczyna się od powiązanego skrótu.

###### `subtitle`
Opcjonalny podtytuł wyświetlany pod tytułem na liście rozwijanej.

###### `title`
Opcjonalny tytuł, który pojawi się po prawej stronie paska wyszukiwania, gdy zapytanie rozpoczyna się od powiązanego skrótu.

###### `shortcut`
Dowolna wartość, której chcesz użyć jako skrótu wyszukiwarki. Nie musi zaczynać się od `!`.

> [!IMPORTANT]
>
> W YAML niektóre znaki mają specjalne znaczenie, gdy znajdują się na początku wartości. Jeśli Twój skrót zaczyna się od `!` (i potencjalnie innych znaków specjalnych), musisz ująć wartość w cudzysłów:
> ```yaml
> shortcut: "!yt"
>```

###### `url`
Adres URL wyszukiwarki. Użyj `{QUERY}`, aby wskazać miejsce wstawienia zapytania. Przykłady:

```yaml
url: https://www.reddit.com/search?q={QUERY}
url: https://store.steampowered.com/search/?term={QUERY}
url: https://www.amazon.com/s?k={QUERY}
```

<a id="group"></a>
### Group
Grupuje wiele widgetów w jeden za pomocą kart. Widgety definiuje się za pomocą właściwości `widgets`, dokładnie tak jak w kolumnie strony. Jedynym ograniczeniem jest to, że nie można umieścić widgetu group ani widgetu split-column wewnątrz widgetu group.

Przykład:

```yaml
- type: group
  widgets:
    - type: reddit
      subreddit: gamingnews
      show-thumbnails: true
      collapse-after: 6
    - type: reddit
      subreddit: games
    - type: reddit
      subreddit: pcgaming
      show-thumbnails: true
```

Podgląd:

![](images/group-widget-preview.png)

#### Współdzielenie właściwości

Aby uniknąć powtarzania, możesz użyć [kotwic YAML](https://support.atlassian.com/bitbucket-cloud/docs/yaml-anchors/) i współdzielić właściwości między widgetami.

Przykład:

```yaml
- type: group
  define: &shared-properties
      type: reddit
      show-thumbnails: true
      collapse-after: 6
  widgets:
    - subreddit: gamingnews
      <<: *shared-properties
    - subreddit: games
      <<: *shared-properties
    - subreddit: pcgaming
      <<: *shared-properties
```

<a id="split-column"></a>
### Split Column
Dzieli kolumnę pełnej szerokości na pół, umożliwiając umieszczenie widgetów obok siebie w poziomie. Na urządzeniach mobilnych lub przy niewystarczającej szerokości jest to zamieniane na jedną kolumnę. Widgety definiuje się za pomocą właściwości `widgets`, dokładnie tak jak w kolumnie strony.

Dwa widgety obok siebie w kolumnie `full`:

![](images/split-column-widget-preview.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
# ...
- size: full
  widgets:
    - type: split-column
      widgets:
        - type: hacker-news
          collapse-after: 3
        - type: lobsters
          collapse-after: 3

    - type: videos
# ...
```
</details>
<br>

Za pomocą samego tego widgetu możesz również uzyskać różne układy pełnej strony, na przykład:

Układ 3 kolumn o równej szerokości:

![](images/split-column-widget-3-columns.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
pages:
  - name: Home
    columns:
      - size: full
        widgets:
          - type: split-column
            max-columns: 3
            widgets:
              - type: reddit
                subreddit: selfhosted
                collapse-after: 15
              - type: reddit
                subreddit: homelab
                collapse-after: 15
              - type: reddit
                subreddit: sysadmin
                collapse-after: 15
```
</details>
<br>

Układ 4 kolumn o równej szerokości (a strona ustawiona na `width: wide`):

![](images/split-column-widget-4-columns.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
pages:
  - name: Home
    width: wide
    columns:
      - size: full
        widgets:
          - type: split-column
            max-columns: 4
            widgets:
              - type: reddit
                subreddit: selfhosted
                collapse-after: 15
              - type: reddit
                subreddit: homelab
                collapse-after: 15
              - type: reddit
                subreddit: linux
                collapse-after: 15
              - type: reddit
                subreddit: sysadmin
                collapse-after: 15
```
</details>
<br>

Układ kaskadowy (masonry) z maksymalnie 5 kolumnami o równej szerokości (a strona ustawiona na `width: wide`):

![](images/split-column-widget-masonry.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
define:
  - &subreddit-settings
    type: reddit
    collapse-after: 5

pages:
  - name: Home
    width: wide
    columns:
      - size: full
        widgets:
          - type: split-column
            max-columns: 5
            widgets:
              - subreddit: selfhosted
                <<: *subreddit-settings
              - subreddit: homelab
                <<: *subreddit-settings
              - subreddit: linux
                <<: *subreddit-settings
              - subreddit: sysadmin
                <<: *subreddit-settings
              - subreddit: DevOps
                <<: *subreddit-settings
              - subreddit: Networking
                <<: *subreddit-settings
              - subreddit: DataHoarding
                <<: *subreddit-settings
              - subreddit: OpenSource
                <<: *subreddit-settings
              - subreddit: Privacy
                <<: *subreddit-settings
              - subreddit: FreeSoftware
                <<: *subreddit-settings
```
</details>
<br>

Podobnie jak w widgecie `group`, możesz wstawić dowolny typ widgetu — możesz nawet wstawić widget `group` do widgetu `split-column`, ale nie możesz wstawić widgetu `split-column` do widgetu `group`.


<a id="custom-api"></a>
### Custom API

Wyświetla dane z API JSON za pomocą niestandardowego szablonu.

> [!NOTE]
>
> Konfiguracja tego widgetu wymaga podstawowej znajomości programowania, HTML, CSS, języka szablonów Go oraz koncepcji specyficznych dla Glance.

Przykłady:

![](images/custom-api-preview-1.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
- type: custom-api
  title: Random Fact
  cache: 6h
  url: https://uselessfacts.jsph.pl/api/v2/facts/random
  template: |
    <p class="size-h4 color-paragraph">{{ .JSON.String "text" }}</p>
```
</details>
<br>

![](images/custom-api-preview-2.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
- type: custom-api
  title: Immich stats
  cache: 1d
  url: https://${IMMICH_URL}/api/server/statistics
  headers:
    x-api-key: ${IMMICH_API_KEY}
    Accept: application/json
  template: |
    <div class="flex justify-between text-center">
      <div>
          <div class="color-highlight size-h3">{{ .JSON.Int "photos" | formatNumber }}</div>
          <div class="size-h6">PHOTOS</div>
      </div>
      <div>
          <div class="color-highlight size-h3">{{ .JSON.Int "videos" | formatNumber }}</div>
          <div class="size-h6">VIDEOS</div>
      </div>
      <div>
          <div class="color-highlight size-h3">{{ div (.JSON.Int "usage" | toFloat) 1073741824 | toInt | formatNumber }}GB</div>
          <div class="size-h6">USAGE</div>
      </div>
    </div>
```
</details>
<br>

![](images/custom-api-preview-3.png)

<details>
<summary>Pokaż <code>glance.yml</code></summary>
<br>

```yaml
- type: custom-api
  title: Steam Specials
  cache: 12h
  url: https://store.steampowered.com/api/featuredcategories?cc=us
  template: |
    <ul class="list list-gap-10 collapsible-container" data-collapse-after="5">
    {{ range .JSON.Array "specials.items" }}
      <li>
        <a class="size-h4 color-highlight block text-truncate" href="https://store.steampowered.com/app/{{ .Int "id" }}/">{{ .String "name" }}</a>
        <ul class="list-horizontal-text">
          <li>{{ div (.Int "final_price" | toFloat) 100 | printf "$%.2f" }}</li>
          {{ $discount := .Int "discount_percent" }}
          <li{{ if ge $discount 40 }} class="color-positive"{{ end }}>{{ $discount }}% off</li>
        </ul>
      </li>
    {{ end }}
    </ul>
```
</details>

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| url | string | nie | |
| headers | klucz (string) i wartość (string) | nie | |
| method | string | nie | GET |
| body-type | string | nie | json |
| body | any | nie | |
| frameless | boolean | nie | false |
| allow-insecure | boolean | nie | false |
| skip-json-validation | boolean | nie | false |
| template | string | tak | |
| options | map | nie | |
| parameters | klucz (string) i wartość (string|array) | nie | |
| subrequests | mapa żądań | nie | |

##### `url`
Adres URL, z którego mają być pobrane dane. Musi być dostępny z serwera, na którym działa Glance.

##### `headers`
Opcjonalnie określa nagłówki wysyłane z żądaniem. Przykład:

```yaml
headers:
  x-api-key: your-api-key
  Accept: application/json
```

##### `method`
Metoda HTTP używana podczas wykonywania żądania. Możliwe wartości: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS` i `HEAD`.

##### `body-type`
Typ treści wysyłanej z żądaniem. Możliwe wartości: `json` i `string`.

##### `body`
Treść wysyłana z żądaniem. Może to być ciąg znaków lub mapa. Przykład:

```yaml
body-type: json
body:
  key1: value1
  key2: value2
  multiple-items:
    - item1
    - item2
```

```yaml
body-type: string
body: |
  key1=value1&key2=value2
```

##### `frameless`
Po ustawieniu na `true` usuwa obramowanie i wewnętrzny odstęp wokół widgetu.

##### `allow-insecure`
Określa, czy ignorować nieprawidłowe certyfikaty lub certyfikaty z podpisem własnym.

##### `skip-json-validation`
Po ustawieniu na `true` pomija sprawdzanie poprawności JSON. Jest to przydatne, gdy API zwraca JSON Lines / newline-delimited JSON, czyli format składający się z kilku obiektów JSON oddzielonych znakami nowej linii.

##### `template`
Szablon używany do wyświetlania danych. Opiera się na pakiecie `html/template` Go, dlatego zaleca się zapoznanie z [jego dokumentacją](https://pkg.go.dev/text/template), aby zrozumieć podstawy, takie jak warunki i pętle. Ponadto używa pakietu [tidwall's gjson](https://github.com/tidwall/gjson) do parsowania danych JSON, więc warto przejrzeć jego dokumentację, jeśli chcesz używać bardziej zaawansowanych selektorów JSON. Więcej przykładów z wyjaśnieniami i definicjami funkcji znajdziesz [tutaj](custom-api.md).

##### `options`
Mapa opcji przekazywanych do szablonu, które można wykorzystać do zmiany zachowania widgetu.

<details>
<summary>Pokaż przykłady</summary>

<br>

Zamiast definiować opcje wewnątrz szablonu i modyfikować sam szablon w następujący sposób:

```yaml
- type: custom-api
  template: |
    {{ /* Opcje konfigurowane przez użytkownika */ }}
    {{ $collapseAfter := 5 }}
    {{ $showThumbnails := true }}
    {{ $showFlairs := false }}

     <ul class="list list-gap-10 collapsible-container" data-collapse-after="{{ $collapseAfter }}">
      {{ if $showThumbnails }}
        <li>
          <img src="{{ .JSON.String "thumbnail" }}" alt="thumbnail" />
        </li>
      {{ end }}
      {{ if $showFlairs }}
        <li>
          <span class="flair">{{ .JSON.String "flair" }}</span>
        </li>
      {{ end }}
     </ul>
```

Możesz użyć właściwości `options`, aby pobrać i zdefiniować wartości domyślne dla tych zmiennych:

```yaml
- type: custom-api
  template: |
    <ul class="list list-gap-10 collapsible-container" data-collapse-after="{{ .Options.IntOr "collapse-after" 5 }}">
      {{ if (.Options.BoolOr "show-thumbnails" true) }}
        <li>
          <img src="{{ .JSON.String "thumbnail" }}" alt="thumbnail" />
        </li>
      {{ end }}
      {{ if (.Options.BoolOr "show-flairs" false) }}
        <li>
          <span class="flair">{{ .JSON.String "flair" }}</span>
        </li>
      {{ end }}
    </ul>
```

Dzięki temu możesz opcjonalnie określić właściwości `collapse-after`, `show-thumbnails` i `show-flairs` w konfiguracji widgetu:

```yaml
- type: custom-api
  options:
    collapse-after: 5
    show-thumbnails: true
    show-flairs: false
```

Co oznacza, że możesz ponownie użyć tego samego szablonu w wielu widgetach z różnymi opcjami:

```yaml
# Pamiętaj, że `custom-widgets` nie jest specjalną właściwością — służy jedynie do zdefiniowania wielokrotnie używanego „anchor”, patrz https://support.atlassian.com/bitbucket-cloud/docs/yaml-anchors/
custom-widgets:
  - &example-widget
    type: custom-api
    template: |
      {{ .Options.StringOr "custom-option" "not defined" }}

pages:
  - name: Home
    columns:
      - size: full
        widgets:
          - <<: *example-widget
            options:
              custom-option: "Value 1"

          - <<: *example-widget
            options:
              custom-option: "Value 2"
```

Obecnie dostępne metody obiektu `.Options` to: `StringOr`, `IntOr`, `BoolOr` i `FloatOr`.

</details>

##### `parameters`
Lista kluczy i wartości wysyłanych do custom-api jako parametry zapytania.

##### `subrequests`
Mapa dodatkowych żądań, które zostaną wykonane równolegle, a następnie udostępnione w szablonie przez właściwość `.Subrequest`. Przykład:

```yaml
- type: custom-api
  cache: 2h
  subrequests:
    another-one:
      url: https://uselessfacts.jsph.pl/api/v2/facts/random
  title: Random Fact
  url: https://uselessfacts.jsph.pl/api/v2/facts/random
  template: |
    <p class="size-h4 color-paragraph">{{ .JSON.String "text" }}</p>
    <p class="size-h4 color-paragraph margin-top-15">{{ (.Subrequest "another-one").JSON.String "text" }}</p>
```

Podżądania obsługują te same właściwości co żądanie główne, z wyjątkiem samego `subrequests`, więc możesz używać `headers`, `parameters` itp.

Zapis `(.Subrequest "key")` może być nieco uciążliwy, dlatego możesz zdefiniować zmienną, aby go uprościć:

```yaml
  template: |
    {{ $anotherOne := .Subrequest "another-one" }}
    <p>{{ $anotherOne.JSON.String "text" }}</p>
```

Możesz również uzyskać dostęp do właściwości `.Response` podżądania, tak jak w przypadku żądania głównego:

```yaml
  template: |
    {{ $anotherOne := .Subrequest "another-one" }}
    <p>{{ $anotherOne.Response.StatusCode }}</p>
```

> [!NOTE]
>
> Ustawienie tej właściwości nadpisze wszelkie parametry zapytania znajdujące się już w adresie URL.

```yaml
parameters:
  param1: value1
  param2:
    - item1
    - item2
```

<a id="opencode-go"></a>
### OpenCode Go
Wyświetla wykorzystanie subskrypcji [OpenCode Go](https://opencode.ai/). Widget pokazuje okna zużycia kroczącego (5 godzin), tygodniowego i miesięcznego wraz z czasem resetu, korzystając z danych z oficjalnego API limitów `https://opencode.ai/zen/go/v1/usage`.

Dla okna miesięcznego może również pokazać prognozę wykorzystania na koniec cyklu rozliczeniowego wraz ze średnim dziennym zużyciem.

Przykład:

```yaml
- type: opencode-go
  api-key: ${OPENCODE_GO_KEY}
```

Podgląd:

(Trzy paski zużycia: 5 godzin, Tydzień i Miesiąc, każdy z procentem i czasem resetu; pasek miesięczny ma dodatkowo kreskowany odcinek prognozy oraz plakietkę `PROGNOZA 61% · 2.0%/dzień`)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| api-key | string | tak | |
| url | string | nie | https://opencode.ai/zen/go/v1/usage |
| warning-threshold | number | nie | 85 |
| show-forecast | boolean | nie | true |
| cache | duration | nie | 5m |

##### `title-url`
Domyślnie tytuł widgetu jest klikalnym linkiem do [konsoli OpenCode](https://opencode.ai/console). Użyj standardowej właściwości `title-url`, aby wskazać inne miejsce.

##### `api-key`
Twój klucz API OpenCode Go (ten sam, którego używasz do API). Jest wysyłany jako token `Bearer`. Zaleca się przekazywanie go przez [zmienną środowiskową](#environment-variables) lub [sekret](#other-ways-of-providing-tokenspasswordssecrets), na przykład `api-key: ${OPENCODE_GO_KEY}`.

##### `url`
Adres URL endpointu zużycia. Zmieniaj go tylko wtedy, gdy chcesz wskazać inny endpoint (np. własny lub za proxy).

##### `warning-threshold`
Procent wykorzystania, przy którym pasek danego okna zmienia kolor na negatywny (ostrzegawczy). Domyślnie `85`.

##### `show-forecast`
Określa, czy pokazywać szacowane zużycie na koniec cyklu dla okna miesięcznego. Domyślnie `true`.

> [!NOTE]
>
> API limitów nie udostępnia prognozy. Zwraca jedynie bieżący procent (ucinany do liczby całkowitej) oraz czas resetu, podczas gdy konsola OpenCode używa dokładnego zużycia w mikrogroszach i dokładnego początku okresu rozliczeniowego. Widget szacuje więc prognozę lokalnie: zakłada, że cykl rozliczeniowy trwa jeden miesiąc kalendarzowy, dodaje pół punktu procentowego, aby skompensować ucinanie, i liniowo ekstrapoluje średnie dzienne zużycie. Niewielka różnica względem konsoli jest oczekiwana — zwykle 1–2 punkty procentowe, a przy bardzo niskim zużyciu większa. Prognoza jest ukryta przez pierwszy dzień cyklu, gdy danych jest zbyt mało.

##### `cache`
Jak długo cache'ować odpowiedź przed ponownym pobraniem. Domyślnie `5m`.

##### Klucz API

Widget korzysta z tego samego klucza co API OpenCode Go. Klucz jest przekazywany w nagłówku:

```http
Authorization: Bearer <api-key>
```

Jeśli subskrypcja Go nie jest aktywna dla danego klucza, API zwróci błąd `403 EntitlementError`, który pojawi się w widgecie jako komunikat błędu.

#### Rozwiązywanie problemów

##### Widget nie wyświetla danych
Sprawdź, czy `api-key` jest ustawiony i poprawny oraz czy serwer Glance ma dostęp do `https://opencode.ai/zen/go/v1/usage`.

##### „Unauthorized” / „EntitlementError”
`Unauthorized` (HTTP 401) oznacza nieprawidłowy lub nieistniejący klucz API. `EntitlementError` (HTTP 403, „OpenCode Go subscription required”) oznacza, że klucz nie ma aktywnej subskrypcji Go.

##### Wartości różnią się od panelu OpenCode
Endpoint API zaokrągla procenty w dół, a panel webowy pokazuje jedną cyfrę po przecinku, więc wynik może różnić się o około 1 punkt procentowy.

<a id="extension"></a>
### Extension
Wyświetla widget dostarczany przez zewnętrzne źródło (stronę trzecią). Jeśli chcesz dowiedzieć się więcej o tworzeniu rozszerzeń, zapoznaj się z [dokumentacją rozszerzeń](extensions.md) (WIP).

```yaml
- type: extension
  url: https://domain.com/widget/display-a-message
  allow-potentially-dangerous-html: true
  parameters:
    message: Hello, world!
```

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| url | string | tak | |
| fallback-content-type | string | nie | |
| allow-potentially-dangerous-html | boolean | nie | false |
| headers | klucz i wartość | nie | |
| parameters | klucz i wartość | nie | |

##### `url`
Adres URL rozszerzenia. **Pamiętaj, że zapytanie jest usuwane z tego adresu URL, a zamiast niego używane są parametry zdefiniowane w `parameters`.**

##### `fallback-content-type`
Opcjonalnie określa zapasowy typ zawartości rozszerzenia, jeśli adres URL nie zwraca prawidłowego nagłówka `Widget-Content-Type`. Obecnie jedyną obsługiwaną wartością tej właściwości jest `html`.

##### `headers`
Opcjonalnie określa nagłówki wysyłane z żądaniem. Przykład:

```yaml
headers:
  x-api-key: ${SECRET_KEY}
```

##### `allow-potentially-dangerous-html`
Określa, czy zezwolić rozszerzeniu na wyświetlanie HTML.

> [!WARNING]
>
> Ta właściwość ma przerażająco brzmiącą nazwę nie bez powodu. Jest przeznaczona dla programistów, którzy czują się swobodnie w tworzeniu i używaniu własnych rozszerzeń. Nie włączaj jej, jeśli nie wiesz, co oznacza, albo jeśli nie masz **absolutnej pewności**, że adres URL używanego rozszerzenia jest bezpieczny.

##### `parameters`
Lista kluczy i wartości wysyłanych do rozszerzenia jako parametry zapytania.

<a id="weather"></a>
### Pogoda
Wyświetla informacje o pogodzie dla określonej lokalizacji. Dane dostarcza https://open-meteo.com/.

Przykład:

```yaml
- type: weather
  units: metric
  hour-format: 12h
  location: London, United Kingdom
```

> [!NOTE]
>
> W przypadku miast w USA o popularnych nazwach można podać stan jako drugi parametr:
>
> * Greenville, North Carolina, United States
> * Greenville, South Carolina, United States
> * Greenville, Mississippi, United States


Podgląd:

![](images/weather-widget-preview.png)

Każdy pasek reprezentuje 2-godzinny przedział. Żółte tło oznacza wschód i zachód słońca. Niebieskie kropki oznaczają pory dnia, w których istnieje duża szansa na opady. Najedź kursorem na paski, aby zobaczyć dokładną temperaturę dla danego czasu.

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| location | string | tak |  |
| units | string | nie | metric |
| hour-format | string | nie | 12h |
| hide-location | boolean | nie | false |
| show-area-name | boolean | nie | false |
| use-brightsky | boolean | nie | false |
| lat | string | nie | |
| lon | string | nie | |

##### `location`
Nazwa miasta i kraju, dla których mają być pobierane informacje o pogodzie. Próba uruchomienia aplikacji z nieprawidłową lokalizacją zakończy się błędem. Możesz użyć [strony API geokodowania](https://open-meteo.com/en/docs/geocoding-api), aby wyszukać swoją lokalizację. Jeśli wyników jest wiele, Glance użyje pierwszego z listy.

#### `use-brightsky`
Po ustawieniu na `true` nie musisz podawać właściwości `location`, ale musisz określić `lat` i `lon`.

#### `lat`
Twoja szerokość geograficzna.
```
lat: 21.37
```

#### `lon`
Twoja długość geograficzna.
```
lon: 21.37
```

##### `units`
Określa, czy wyświetlać temperaturę w stopniach Celsjusza, czy Fahrenheita. Możliwe wartości: `metric` lub `imperial`.

#### `hour-format`
Określa, czy wyświetlać godziny w formacie 12- czy 24-godzinnym. Możliwe wartości: `12h` i `24h`.

##### `hide-location`
Opcjonalnie ukrywa nazwę lokalizacji w widgecie.

##### `show-area-name`
Określa, czy wyświetlać stan/region administracyjny w nazwie lokalizacji. Po ustawieniu na `true` lokalizacja będzie wyświetlana jako:

```
Greenville, North Carolina, United States
```

W przeciwnym razie, po ustawieniu na `false` (wartość domyślna), będzie wyświetlana jako:

```
Greenville, United States
```

<a id="todo"></a>
### Todo

Prosta lista zadań umożliwiająca dodawanie, edytowanie i usuwanie zadań. Zadania są przechowywane w pamięci lokalnej przeglądarki.

Przykład:

```yaml
- type: to-do
```

Podgląd:

![](images/todo-widget-preview.png)

Aby zmienić kolejność zadań, przeciągnij je, chwytając za górną część zadania:

![](images/reorder-todo-tasks-prevew.gif)

Aby usunąć zadanie, najedź na nie kursorem i kliknij ikonę kosza.

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| id | string | nie | |

##### `id`

Identyfikator listy zadań. Jeśli chcesz mieć wiele list zadań, musisz podać inny identyfikator dla każdej z nich. Identyfikator służy do przechowywania zadań w pamięci lokalnej przeglądarki. Oznacza to, że jeśli masz wiele list zadań o tym samym identyfikatorze, będą one współdzielić te same zadania.

#### Skróty klawiszowe
| Klawisze | Akcja | Warunek |
| ---- | ------ | --------- |
| <kbd>Enter</kbd> | Dodaje zadanie na dół listy | Gdy pole „Add a task” ma fokus |
| <kbd>Ctrl</kbd> + <kbd>Enter</kbd> | Dodaje zadanie na górę listy | Gdy pole „Add a task” ma fokus |
| <kbd>Down Arrow</kbd> | Ustawia fokus na ostatnio dodanym zadaniu | Gdy pole „Add a task” ma fokus |
| <kbd>Escape</kbd> | Ustawia fokus na polu „Add a task” | Gdy zadanie ma fokus |

<a id="vikunja"></a>
### Vikunja
Wyświetla zadania z Vikunja i pozwala dodawać, edytować oraz oznaczać je jako wykonane bezpośrednio z dashboardu.

Przykład:

```yaml
- type: vikunja
  url: https://vikunja.example.com
  token: twoj-token-api
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| url | string | tak | |
| token | string | tak | |
| limit | integer | nie | 10 |
| project-id | integer | nie | 1 |
| ignored-projects | array | nie | |
| affine-url | string | nie | |
| affine-email | string | nie | |
| affine-password | string | nie | |

##### `url`
Adres URL Twojej instancji Vikunja.

##### `token`
Token API Vikunja z uprawnieniami do odczytu i modyfikacji zadań oraz etykiet.

##### `limit`
Maksymalna liczba wyświetlanych zadań.

##### `project-id`
ID projektu, w którym będą tworzone nowe zadania. Można je znaleźć w adresie URL projektu, np. `/projects/5` oznacza ID `5`.

##### `ignored-projects`
Lista ID projektów, których zadania mają być pomijane w widgecie.

##### `affine-url`
Adres URL instancji Affine. Włącza integrację z Affine.

##### `affine-email`
Adres e-mail używany do logowania w Affine.

##### `affine-password`
Hasło do konta Affine.

#### Funkcje

Widget wyświetla niewykonane zadania posortowane według terminu, wraz z czasem pozostałym do końca, tytułem i etykietami. Możesz oznaczać zadania jako wykonane, dodawać nowe zadania oraz edytować istniejące, w tym zmieniać tytuł, termin, etykiety i link do notatki Affine.

#### Integracja z Affine

Integracja z Affine jest opcjonalna i pozwala powiązać zadanie Vikunja z notatką w Affine. Aby ją włączyć, podaj `affine-url`, `affine-email` i `affine-password`.

```yaml
- type: vikunja
  url: https://vikunja.example.com
  token: twoj-token-api
  affine-url: https://affine.example.com
  affine-email: user@example.com
  affine-password: twoje-haslo
```

Link do notatki podaje się w polu „Link do notatki Affine” podczas tworzenia lub edycji zadania. URL notatki ma format:

```
https://affine.example.com/workspace/WORKSPACE_ID/PAGE_ID
```

Po dodaniu linku w tabeli zadań pojawi się ikona dokumentu z tytułem notatki; kliknięcie otworzy notatkę w nowej karcie. Tytuł jest pobierany z Affine przy każdym odświeżeniu widgetu.

> [!NOTE]
>
> Dane logowania do Affine są przechowywane w pliku konfiguracyjnym. Zabezpiecz plik odpowiednimi uprawnieniami i nie commituj go z hasłami do repozytorium. Hasło możesz przekazać przez zmienną środowiskową, np. `affine-password: ${AFFINE_PASSWORD}`.

#### Rozwiązywanie problemów

##### Widget nie wyświetla zadań
Sprawdź poprawność `url` i `token` oraz konsolę przeglądarki pod kątem błędów połączenia.

##### Nie można oznaczyć zadania jako wykonane lub utworzyć zadania
Upewnij się, że token API ma uprawnienia do modyfikacji zadań, a instancja Vikunja działa poprawnie.

##### Etykiety nie są wyświetlane
Sprawdź, czy w Vikunja istnieją etykiety i czy token ma uprawnienia do ich odczytu.

##### Tytuł notatki Affine się nie pojawia
Sprawdź dane logowania do Affine i format URL notatki. URL musi mieć postać `https://affine-url/workspace/WORKSPACE_ID/PAGE_ID`.

##### Błąd „failed to sign in to Affine”
Sprawdź e-mail i hasło oraz dostępność Affine pod podanym adresem. URL musi zaczynać się od `https://` lub `http://`.

##### Błąd „could not extract workspace ID and page ID from URL”
Sprawdź format URL notatki — musi zawierać segmenty `/workspace/WORKSPACE_ID/PAGE_ID`.

<a id="monitor"></a>
### Monitor
Wyświetla listę witryn i informację, czy są osiągalne (online), czy nie. Określa się to, wysyłając żądanie GET pod podany adres URL — jeśli odpowiedź ma kod 200, witryna działa. Wyświetlany jest również czas otrzymania odpowiedzi w milisekundach.

Przykład:

```yaml
- type: monitor
  cache: 1m
  title: Services
  sites:
    - title: Jellyfin
      url: https://jellyfin.yourdomain.com
      icon: /assets/jellyfin-logo.png
    - title: Gitea
      url: https://gitea.yourdomain.com
      icon: /assets/gitea-logo.png
    - title: Immich
      url: https://immich.yourdomain.com
      icon: /assets/immich-logo.png
    - title: AdGuard Home
      url: https://adguard.yourdomain.com
      icon: /assets/adguard-logo.png
    - title: Vaultwarden
      url: https://vault.yourdomain.com
      icon: /assets/vaultwarden-logo.png
```

Podgląd:

![](images/monitor-widget-preview.png)

Najedź kursorem na tekst „ERROR”, aby zobaczyć więcej informacji.

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| sites | array | tak | |
| style | string | nie | |
| show-failing-only | boolean | nie | false |

##### `show-failing-only`
Po ustawieniu na `true` pokazuje tylko listę niedziałających witryn.

##### `style`
Służy do zmiany wyglądu widgetu. Możliwa wartość: `compact`.

Podgląd `compact`:

![](images/monitor-widget-compact-preview.png)

##### `sites`

Właściwości każdej witryny:

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| title | string | tak | |
| url | string | tak | |
| check-url | string | nie | |
| error-url | string | nie | |
| icon | string | nie | |
| timeout | string | nie | 3s |
| allow-insecure | boolean | nie | false |
| same-tab | boolean | nie | false |
| alt-status-codes | array | nie | |
| basic-auth | object | nie | |

`title`

Tytuł używany do oznaczenia witryny.

`url`

Adres URL monitorowanej usługi, który musi być osiągalny dla Glance. Będzie on używany jako link otwierany po kliknięciu tytułu. Jeśli `check-url` nie jest określony, ten adres jest używany do sprawdzania statusu.

`check-url`

Adres URL, do którego zostanie wysłane żądanie, a jego odpowiedź określi status witryny. Jeśli nie zostanie podany, używana jest właściwość `url`.

`error-url`

Jeśli monitorowana usługa zwróci błąd, użytkownik zostanie przekierowany tutaj. Jeśli nie zostanie podany, używana jest właściwość `url`.

`icon`

Więcej informacji o określaniu ikon znajdziesz w sekcji [Ikony](#icons).

`timeout`

Jak długo czekać na odpowiedź serwera, zanim uzna się go za nieosiągalny. Wartość jest ciągiem znaków i musi być liczbą, po której następuje jedna z liter s, m, h, d. Przykład: `5s` dla 5 sekund, `1m` dla 1 minuty itp.

`allow-insecure`

Określa, czy ignorować nieprawidłowe certyfikaty lub certyfikaty z podpisem własnym.

`same-tab`

Określa, czy otwierać link w tej samej, czy w nowej karcie.

`alt-status-codes`

Kody statusu inne niż 200, które mają być traktowane jako „OK”.

```yaml
alt-status-codes:
  - 403
```

`basic-auth`

Dane uwierzytelniania HTTP Basic dla chronionych witryn.

```yaml
basic-auth:
  username: your-username
  password: your-password
```

<a id="releases"></a>
### Releases
Wyświetla listę najnowszych wydań dla określonych repozytoriów na GitHubie, GitLabie, Codebergu lub Docker Hubie.

Przykład:

```yaml
- type: releases
  show-source-icon: true
  repositories:
    - go-gitea/gitea
    - jellyfin/jellyfin
    - glanceapp/glance
    - codeberg:redict/redict
    - gitlab:fdroid/fdroidclient
    - dockerhub:gotify/server
```

Podgląd:

![](images/releases-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| repositories | array | tak |  |
| show-source-icon | boolean | nie | false |  |
| token | string | nie | |
| gitlab-token | string | nie | |
| limit | integer | nie | 10 |
| collapse-after | integer | nie | 5 |

##### `repositories`
Lista repozytoriów, dla których mają być pobierane najnowsze wydania. Wymagana jest tylko nazwa/repo, a nie pełny adres URL. Dla repozytoriów hostowanych gdzie indziej, takich jak GitLab, Codeberg i Docker Hub, można podać prefiks. Przykład:

```yaml
repositories:
  - gitlab:inkscape/inkscape
  - dockerhub:glanceapp/glance
  - codeberg:redict/redict
```

Oficjalne obrazy na Docker Hubie można określić, pomijając właściciela:

```yaml
repositories:
  - dockerhub:nginx
  - dockerhub:node
  - dockerhub:alpine
```

Możesz również podać dokładne tagi obrazów Docker Hub:

```yaml
repositories:
  - dockerhub:nginx:latest
  - dockerhub:nginx:stable-alpine
```

Aby uwzględnić wersje przedpremierowe, możesz podać repozytorium jako obiekt i użyć właściwości `include-prereleases`:

**Uwaga: ta funkcja jest obecnie dostępna tylko dla repozytoriów GitHub.**

```yaml
repositories:
  - gitlab:inkscape/inkscape
  - repository: glanceapp/glance
    include-prereleases: true
  - codeberg:redict/redict
```

##### `show-source-icon`
Po ustawieniu na `true` pokazuje ikonę źródła (GitHub/GitLab/Codeberg/Docker Hub) obok nazwy repozytorium.

##### `token`
Bez uwierzytelniania GitHub zezwala na maksymalnie 60 żądań na godzinę. Łatwo przekroczyć ten limit i zacząć widzieć błędy, jeśli śledzisz wiele repozytoriów lub masz krótki czas cache. Aby temu zaradzić, możesz [utworzyć token tylko do odczytu na swoim koncie GitHub](https://github.com/settings/personal-access-tokens/new) i podać go tutaj.

Możesz również podać wartość tego tokena przez zmienną środowiskową, używając składni `${GITHUB_TOKEN}`, gdzie `GITHUB_TOKEN` to nazwa zmiennej przechowującej token. Jeśli zainstalowałeś Glance przez Docker, możesz podać token w `docker-compose`:

```yaml
services:
  glance:
    image: glanceapp/glance
    environment:
      - GITHUB_TOKEN=<your token>
```

a następnie użyć go w `glance.yml` w następujący sposób:

```yaml
- type: releases
  token: ${GITHUB_TOKEN}
  repositories: ...
```

Dzięki temu możesz bezpiecznie przechowywać `glance.yml` w systemie kontroli wersji bez ujawniania tokena.

##### `gitlab-token`
Działa jak powyższy, ale jest używany podczas pobierania wydań z GitLaba.

##### `limit`
Maksymalna liczba wydań do wyświetlenia.

#### `collapse-after`
Liczba widocznych wydań, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

<a id="docker-containers"></a>
### Docker Containers

Wyświetla status kontenerów Docker wraz z ikoną i opcjonalnym krótkim opisem.

![](images/docker-containers-preview.png)

```yaml
- type: docker-containers
  hide-by-default: false
```

> [!NOTE]
>
> Widget wymaga dostępu do `docker.sock`. Jeśli uruchamiasz Glance w kontenerze, możesz to zrobić, montując gniazdo jako wolumen:
>
> ```yaml
> services:
>   glance:
>     image: glanceapp/glance
>     volumes:
>       - /var/run/docker.sock:/var/run/docker.sock
> ```

Konfiguracja kontenerów odbywa się za pomocą etykiet (labels) przypisanych do każdego kontenera:

```yaml
  jellyfin:
    image: jellyfin/jellyfin:latest
    labels:
      glance.name: Jellyfin
      glance.icon: si:jellyfin
      glance.url: https://jellyfin.domain.com
      glance.description: Movies & shows
```

Alternatywnie możesz zdefiniować wartości w `glance.yml` za pomocą właściwości `containers`, gdzie kluczem jest nazwa kontenera, a każda wartość odpowiada etykietom, ale bez prefiksu „glance.”:

```yaml
- type: docker-containers
  containers:
    container_name_1:
      name: Container Name
      description: Description of the container
      url: https://container.domain.com
      icon: si:container-icon
      hide: false
```

W przypadku usług składających się z wielu kontenerów możesz podać `glance.id` na kontenerze „głównym” i `glance.parent` na każdym kontenerze „podrzędnym”:

<details>
<summary>Pokaż <code>docker-compose.yml</code></summary>
<br>

```yaml
services:
  immich-server:
    image: ghcr.io/immich-app/immich-server
    labels:
      glance.name: Immich
      glance.icon: si:immich
      glance.url: https://immich.domain.com
      glance.description: Image & video management
      glance.id: immich

  redis:
    image: docker.io/redis:6.2-alpine
    labels:
      glance.parent: immich
      glance.name: Redis

  database:
    image: docker.io/tensorchord/pgvecto-rs:pg14-v0.2.0
    labels:
      glance.parent: immich
      glance.name: DB

  proxy:
    image: nginx:stable
    labels:
      glance.parent: immich
      glance.name: Proxy
```
</details>
<br>

Spowoduje to umieszczenie wszystkich kontenerów podrzędnych pod kontenerem `Immich` po najechaniu na jego ikonę:

![](images/docker-container-parent.png)

Jeśli którykolwiek z kontenerów podrzędnych nie działa, jego status zostanie przekazany do kontenera nadrzędnego:

![](images/docker-container-parent2.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| hide-by-default | boolean | nie | false |
| format-container-names | boolean | nie | false |
| sock-path | string | nie | /var/run/docker.sock |
| category | string | nie | |
| running-only | boolean | nie | false |

##### `hide-by-default`
Określa, czy domyślnie ukrywać kontenery. Po ustawieniu na `true` musisz ręcznie dodać etykietę `glance.hide: false` do każdego kontenera, który chcesz wyświetlić. Domyślnie wszystkie kontenery są pokazywane, a aby ukryć konkretny kontener, możesz dodać etykietę `glance.hide: true`.

##### `format-container-names`
Po ustawieniu na `true` automatycznie konwertuje nazwy kontenerów, takie jak `container_name_1`, na `Container Name 1`.

##### `sock-path`
Ścieżka do gniazda Docker. Może to być również [zdalne gniazdo](https://docs.docker.com/engine/daemon/remote-access/) lub gniazdo za proxy, np. za pomocą [docker-socket-proxy](https://github.com/Tecnativa/docker-socket-proxy).

###### `category`
Filtruje kontenery tylko do tych, które mają kategorię określoną etykietą `glance.category`. Przydatne, jeśli chcesz mieć wiele widgetów kontenerów, z których każdy pokazuje inny zestaw kontenerów.

<details>
<summary>Pokaż przykład</summary>
<br>


```yaml
services:
  jellyfin:
    image: jellyfin/jellyfin:latest
    labels:
      glance.name: Jellyfin
      glance.icon: si:jellyfin
      glance.url: https://jellyfin.domain.com
      glance.category: media

  gitea:
    image: gitea/gitea:latest
    labels:
      glance.name: Gitea
      glance.icon: si:gitea
      glance.url: https://gitea.domain.com
      glance.category: dev-tools

  vaultwarden:
    image: vaultwarden/server:latest
    labels:
      glance.name: Vaultwarden
      glance.icon: si:vaultwarden
      glance.url: https://vaultwarden.domain.com
      glance.category: dev-tools
```

Następnie możesz użyć właściwości `category`, aby filtrować kontenery:

```yaml
- type: docker-containers
  title: Dev tool containers
  category: dev-tools

- type: docker-containers
  title: Media containers
  category: media
```

</details>

##### `running-only`
Określa, czy pokazywać tylko działające kontenery. Po ustawieniu na `true` wyświetlane będą tylko kontenery aktualnie działające. Po ustawieniu na `false` wyświetlane będą wszystkie kontenery niezależnie od stanu.

#### Etykiety
| Nazwa | Opis |
| ---- | ----------- |
| glance.name | Nazwa wyświetlana w interfejsie. Jeśli nie zostanie podana, użyta zostanie nazwa kontenera. |
| glance.icon | Więcej informacji o określaniu ikon znajdziesz w sekcji [Ikony](#icons). |
| glance.url | Adres URL, na który użytkownik zostanie przekierowany po kliknięciu kontenera. |
| glance.same-tab | Określa, czy otwierać link w tej samej, czy w nowej karcie. Domyślnie `false`. |
| glance.description | Krótki opis wyświetlany w interfejsie. Domyślnie pusty. |
| glance.hide | Określa, czy ukryć kontener. Po ustawieniu na `true` kontener nie będzie wyświetlany. Domyślnie `false`. |
| glance.id | Niestandardowy identyfikator kontenera. Używany do grupowania kontenerów pod jednym elementem nadrzędnym. |
| glance.parent | Identyfikator kontenera nadrzędnego. Używany do grupowania kontenerów pod jednym elementem nadrzędnym. |
| glance.category | Kategoria kontenera. Używana do filtrowania kontenerów według kategorii. |

<a id="qbittorrent"></a>
### qBittorrent
Wyświetla status torrentów z qBittorrent wraz z podsumowaniem pobierania i listą torrentów.

Przykład:

```yaml
- type: qbittorrent
  url: http://localhost:8080
  username: admin
  password: adminadmin
```

Podgląd:

![](images/qbittorrent-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| url | string | tak | |
| username | string | tak | |
| password | string | tak | |
| hide-seeding | boolean | nie | false |
| hide-completed | boolean | nie | false |
| show-only-active | boolean | nie | false |
| limit | integer | nie | 10 |
| sort-by | string | nie | progress |

##### `url`
Adres URL interfejsu webowego qBittorrent.

##### `username`
Nazwa użytkownika używana do logowania.

##### `password`
Hasło używane do logowania.

##### `hide-seeding`
Po ustawieniu na `true` ukrywa torrenty, które seedują.

##### `hide-completed`
Po ustawieniu na `true` ukrywa ukończone torrenty.

##### `show-only-active`
Po ustawieniu na `true` pokazuje tylko aktywne torrenty (z prędkością większą od 0).

##### `limit`
Maksymalna liczba wyświetlanych torrentów.

##### `sort-by`
Sposób sortowania torrentów. Możliwe wartości: `name`, `progress`, `speed` i `eta`.

#### Wyświetlane informacje

Widget pokazuje podsumowanie z liczbą pobieranych, seedujących i wstrzymanych torrentów oraz łączną prędkością pobierania i wysyłania. Poniżej znajduje się lista torrentów z nazwą, paskiem postępu, procentem ukończenia, kategorią, statusem, prędkością pobierania, szacowanym czasem zakończenia (ETA) i prędkością wysyłania.

#### Stany torrentów

| Ikona | Stan |
| ----- | ---- |
| 🔵 (niebieski) | Pobieranie |
| 🟢 (zielony) | Seedowanie |
| ⏸️ (szary) | Wstrzymany |
| ⚠️ (szary) | Zablokowany (stalled) |
| 🕐 (szary) | W kolejce |
| 🔄 (szary) | Sprawdzanie |
| 🔴 (czerwony) | Błąd |

#### Wymagania

- qBittorrent z włączonym interfejsem webowym (Tools → Options → Web UI).
- Prawidłowe dane uwierzytelniające.

#### Uwagi dotyczące bezpieczeństwa

Widget automatycznie zarządza sesją i loguje się ponownie po jej wygaśnięciu. Po 3 nieudanych próbach logowania blokuje dalsze próby na 30 minut. Hasło jest przesyłane bezpośrednio do API qBittorrent, dlatego zaleca się używanie HTTPS, jeśli qBittorrent jest dostępny z zewnątrz.

<a id="google-compute-engine"></a>
### Google Compute Engine
Wyświetla instancje Google Compute Engine dla projektu i pozwala uruchamiać, restartować lub zatrzymywać je bezpośrednio z Glance.

```yaml
- type: google-compute
  project-id: my-project-id
  service-account-key: ${GCP_SERVICE_ACCOUNT_JSON}
  zones:
    - europe-central2-a
    - europe-central2-b
```

##### `project-id`
Identyfikator projektu Google Cloud, z którego mają być odczytywane instancje.

##### `service-account-key`
Klucz konta usługi z uprawnieniami do listowania instancji oraz wykonywania operacji `start`, `stop` i `reset`. Możesz podać ścieżkę do pliku, klucz JSON zakodowany w base64 lub surową wartość JSON.

##### `zones`
Opcjonalna lista stref. Po podaniu widget wyświetla i kontroluje tylko instancje z tych stref.

<a id="tailscale"></a>
### Tailscale
Wyświetla listę urządzeń w sieci Tailscale wraz ze statusem połączenia i opcjonalnymi znacznikami.

Przykład:

```yaml
- type: tailscale
  token: twoj_token
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| token | string | tak | |
| tailnet | string | nie | - |
| url | string | nie | https://api.tailscale.com/api/v2/tailnet/-/devices |
| collapse-after | integer | nie | 5 |
| show-online-indicator | boolean | nie | false |
| show-expiry-disabled | boolean | nie | false |
| hide-offline | boolean | nie | false |
| show-disconnected | boolean | nie | false |
| show-blocks-incoming | boolean | nie | false |
| show-joined-date | boolean | nie | false |

##### `token`
Token API Tailscale z uprawnieniami do odczytu urządzeń (`Devices: Read only`). Można go wygenerować na stronie https://login.tailscale.com/admin/settings/keys.

##### `tailnet`
Identyfikator tailnetu, z którego mają być pobierane urządzenia. Wartość `-` oznacza tailnet powiązany z tokenem.

##### `url`
Pełny adres URL API Tailscale. Domyślnie jest generowany na podstawie `tailnet`; podawaj go tylko w celu nadpisania domyślnego zachowania.

##### `collapse-after`
Po ilu urządzeniach lista ma zostać zwinięta. Ustaw `-1`, aby nigdy nie zwijać.

##### `show-online-indicator`
Określa, czy pokazywać zielony (online) lub czerwony (offline) punkt przy nazwie urządzenia.

##### `show-expiry-disabled`
Pokazuje znacznik „Expiry disabled”, gdy klucz urządzenia nie wygasa automatycznie.

##### `hide-offline`
Po ustawieniu na `true` ukrywa urządzenia, które są offline.

##### `show-disconnected`
Pokazuje znacznik „Disconnected”, gdy urządzenie nie jest połączone z panelem kontrolnym Tailscale.

##### `show-blocks-incoming`
Pokazuje znacznik „Blocks Incoming”, gdy urządzenie blokuje połączenia przychodzące.

##### `show-joined-date`
Pokazuje datę dodania urządzenia do sieci Tailscale.

> [!NOTE]
>
> Publiczne API Tailscale nie udostępnia informacji o exit node, subnetach, SSH, tagach ani urządzeniach współdzielonych. Dane te są widoczne tylko w panelu webowym Tailscale.

#### Rozwiązywanie problemów

##### Brak urządzeń
Sprawdź, czy token ma uprawnienia `Devices: Read only` i czy jest poprawny. Możesz go przetestować poleceniem:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://api.tailscale.com/api/v2/tailnet/-/devices
```

##### Znaczniki się nie pokazują
Upewnij się, że włączyłeś odpowiednie opcje `show-*: true` i że API zwraca dane pola dla urządzenia.

##### Widget działa wolno
Zwiększ wartość `cache`, na przykład do `30m` lub `1h`.

<a id="dns-stats"></a>
### DNS Stats
Wyświetla statystyki z samodzielnie hostowanego resolvera DNS blokującego reklamy, takiego jak AdGuard Home, Pi-hole lub Technitium.

Przykład:

```yaml
- type: dns-stats
  service: adguard
  url: https://adguard.domain.com/
  username: admin
  password: ${ADGUARD_PASSWORD}
```

Podgląd:

![](images/dns-stats-widget-preview.png)

> [!NOTE]
>
> W przypadku AdGuard Home trzecią statystyką u góry będzie średnie opóźnienie, a w przypadku Pi-hole lub Technitium — łączna liczba zablokowanych domen ze wszystkich list reklamowych.

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| service | string | nie | pihole |
| allow-insecure | bool | nie | false |
| url | string | tak |  |
| username | string | gdy `service` to `adguard` |  |
| password | string | gdy `service` to `adguard` lub `pihole-v6` |  |
| token | string | gdy `service` to `pihole` |  |
| hide-graph | bool | nie | false |
| hide-top-domains | bool | nie | false |
| hour-format | string | nie | 12h |

##### `service`
Jedna z wartości: `adguard`, `technitium`, `pihole` (wersja główna 5 i starsze) lub `pihole-v6` (wersja główna 6 i nowsze).

##### `allow-insecure`
Określa, czy zezwalać na nieprawidłowe certyfikaty lub certyfikaty z podpisem własnym podczas wysyłania żądania do usługi.

##### `url`
Bazowy adres URL usługi.

##### `username`
Wymagane tylko w przypadku AdGuard Home. Nazwa użytkownika używana do logowania się do panelu administracyjnego.

##### `password`
Wymagane w przypadku AdGuard Home — hasło używane do logowania się do panelu administracyjnego.

Wymagane również w przypadku Pi-hole w wersji głównej 6 i nowszej; jest to hasło używane do logowania się do panelu administracyjnego lub hasło aplikacji, które można znaleźć w `Settings -> Web Interface / API -> Configure app password`.

##### `token`
Wymagane w przypadku Pi-hole w wersji głównej 5 lub starszej. Token API, który można znaleźć w `Settings -> API -> Show API token`.

Wymagane również w przypadku Technitium — token API można wygenerować w `Administration -> Sessions -> Create Token`.

##### `hide-graph`
Określa, czy ukryć wykres liczby zapytań w czasie.

##### `hide-top-domains`
Określa, czy ukryć listę najczęściej blokowanych domen.

##### `hour-format`
Określa, czy wyświetlać czas względny na wykresie w formacie `12h`, czy `24h`.

<a id="cloudflare"></a>
### Cloudflare
Wyświetla liczbę zapytań do stron chronionych przez Cloudflare wraz z liczbą zablokowanych użytkowników.

Przykład:

```yaml
- type: cloudflare
  api-key: KLUCZ
  zone-id: ID_STREFY
  time-range: 24h
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| api-key | string | tak | |
| zone-id | string | tak | |
| time-range | string | nie | 24h |

##### `api-key`
Klucz API Cloudflare z uprawnieniami do odczytu analityki i logów. Można go wygenerować w sekcji „Get your API token” po wejściu na domenę w Cloudflare (wystarczy szablon „Read analytics and Logs”).

##### `zone-id`
Identyfikator strefy (Zone ID), widoczny w panelu Cloudflare nad sekcją pobierania klucza API.

##### `time-range`
Zakres czasu pobieranych danych. Możliwe wartości: `24h`, `7d` i `30d`.

#### `title-url`
Domyślnie tytuł prowadzi do statystyk bezpieczeństwa strefy w panelu Cloudflare. Możesz to zmienić standardową właściwością `title-url`, na przykład:

- `https://dash.cloudflare.com/?to=/{account_id}/{domena}/security/analytics`
- `https://dash.cloudflare.com/?to=/{account_id}/{domena}/analytics/traffic`

<a id="server-stats"></a>
### Server Stats
Wyświetla statystyki, takie jak użycie procesora, pamięci i dysku, serwera, na którym działa Glance, lub innych serwerów.

Przykład:

```yaml
- type: server-stats
  servers:
    - type: local
      name: Services
```

Podgląd:

![](images/server-stats-preview.gif)

> [!NOTE]
>
> Ten widget jest obecnie w fazie rozwoju — niektóre funkcje mogą nie działać zgodnie z oczekiwaniami lub mogą się zmienić.

Aby wyświetlać dane z zdalnego serwera, musisz mieć uruchomionego agenta Glance na tym serwerze. Agenta możesz pobrać [tutaj](https://github.com/glanceapp/agent), pamiętaj jednak, że jest on wciąż w fazie rozwoju i może nie działać zgodnie z oczekiwaniami. Wsparcie dla innych dostawców, takich jak Glances, zostanie dodane w przyszłości.

Jeśli temperatura procesora przekroczy 80°C, obok procesora pojawi się ikona płomienia. Wskaźniki postępu zmienią również kolor na czerwony (lub odpowiednik Twojego koloru negatywnego), aby zwrócić uwagę na nietypowo wysokie wartości:

![](images/server-stats-flame-icon.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| servers | array | nie |  |

##### `servers`
Jeśli nie zostanie podane, wyświetlane będą statystyki serwera, na którym działa Glance.

##### Właściwości serwerów `local` i `remote`
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| type | string | tak |  |
| name | string | nie |  |
| hide-swap | boolean | nie | false |

###### `type`
Określa, czy wyświetlać statystyki serwera lokalnego, czy zdalnego. Możliwe wartości: `local` i `remote`.

###### `name`
Nazwa serwera wyświetlana w widgecie. Jeśli nie zostanie podana, użyta zostanie nazwa hosta serwera.

###### `hide-swap`
Określa, czy ukryć użycie pamięci swap.

##### Właściwości serwera `local`
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| cpu-temp-sensor | string | nie |  |
| hide-mountpoints-by-default | boolean | nie | false |
| mountpoints | map\[string\]object | nie |  |

###### `cpu-temp-sensor`
Nazwa czujnika używanego do odczytu temperatury procesora. Jeśli nie zostanie podana, widget spróbuje znaleźć właściwy czujnik; jeśli mu się to nie uda, temperatura nie będzie wyświetlana. Dostępne czujniki możesz wyświetlić poleceniem `sensors`.

###### `hide-mountpoints-by-default`
Po ustawieniu na `true` musisz ręcznie uczynić każdy punkt montowania widocznym, dodając do niego właściwość `hide: false` w następujący sposób:

```yaml
- type: server-stats
  servers:
    - type: local
      hide-mountpoints-by-default: true
      mountpoints:
        "/":
          hide: false
        "/mnt/data":
          hide: false
```

Jest to przydatne, jeśli uruchamiasz Glance w kontenerze, który zwykle montuje wiele nieistotnych systemów plików.

###### `mountpoints`
Mapa punktów montowania, dla których ma być wyświetlane użycie dysku. Kluczem jest ścieżka punktu montowania, a wartością obiekt z opcjonalnymi właściwościami. Przykład:

```yaml
mountpoints:
  "/":
    name: Root
  "/mnt/data":
    name: Data
  "/boot/efi":
    hide: true
```

##### Właściwości każdego `mountpoint`
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| name | string | nie |  |
| hide | boolean | nie | false |

###### `name`
Nazwa punktu montowania wyświetlana w widgecie. Jeśli nie zostanie podana, użyta zostanie ścieżka punktu montowania.

###### `hide`
Określa, czy ukryć ten punkt montowania w widgecie.

##### Właściwości serwerów `remote`
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| url | string | tak |  |
| token | string | nie |  |
| timeout | string | nie | 3s |

###### `url`
Adres URL i port serwera, z którego mają być pobierane statystyki.

###### `token`
Token uwierzytelniający używany podczas pobierania statystyk.

###### `timeout`
Maksymalny czas oczekiwania na odpowiedź serwera. Wartość jest ciągiem znaków i musi być liczbą, po której następuje jedna z liter s, m, h, d. Przykład: `10s` dla 10 sekund, `1m` dla 1 minuty itp.

<a id="beszel"></a>
### Beszel
Wyświetla status serwerów monitorowanych przez Beszel wraz z użyciem procesora, pamięci i dysku.

Przykład:

```yaml
- type: beszel
  url: https://beszel.example.com
  token: twoj-token
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| url | string | tak | |
| token | string | nie | |
| identity | string | nie | |
| password | string | nie | |
| redirect-url | string | nie | |
| show-charts | boolean | nie | true |
| cache | duration | nie | 10s |

##### `url`
Adres URL API Twojej instancji Beszel.

##### `token`
Token JWT używany do autoryzacji. Jeśli go nie podasz, możesz użyć `identity` i `password`.

##### `identity`
Login lub e-mail użytkownika Beszel. Wymaga jednoczesnego podania `password`.

##### `password`
Hasło użytkownika Beszel. Wymaga jednoczesnego podania `identity`.

> [!NOTE]
>
> Jeśli podasz `identity` i `password`, widget automatycznie pobierze token i będzie go odświeżał co 3 dni. Token jest przechowywany tylko w pamięci procesu.

##### `redirect-url`
Adres URL interfejsu webowego Beszel. Po podaniu kliknięcie nazwy serwera otworzy jego szczegóły w panelu Beszel.

##### `show-charts`
Określa, czy pokazywać interaktywne wykresy historyczne. Domyślnie `true`.

#### Funkcje

Widget pokazuje listę serwerów ze statusem online/offline, czasem działania oraz paskami użycia CPU, RAM i dysku. Po najechaniu na serwer lub pasek wyświetlane są dodatkowe informacje, takie jak host, wersja jądra, model procesora i średnie obciążenie. Wykresy historyczne obsługują metryki CPU, RAM, dysk i sieć oraz zakresy 1h, 12h, 24h, 7d i 30d.

#### Rozwiązywanie problemów

##### Widget nie wyświetla danych
Sprawdź, czy adres URL instancji Beszel jest poprawny i dostępny z serwera Glance oraz czy endpoint `/api/collections/systems/records` odpowiada. Jeśli instancja wymaga autoryzacji, upewnij się, że token lub dane logowania są poprawne.

##### Brakujące metryki
Niektóre metryki, np. Load Average, mogą być niedostępne w zależności od wersji agenta Beszel.

##### Wykresy się nie ładują
Sprawdź, czy endpoint `/api/collections/system_stats/records` jest dostępny i czy Beszel zbiera dane historyczne.

<a id="github"></a>
### GitHub
Wyświetla listę repozytoriów GitHub, do których masz dostęp — publicznych i prywatnych.

Przykład:

```yaml
- type: github
  token: ghp_xxxxx
  collapse-after: 5
  sort: updated
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| token | string | nie | |
| collapse-after | integer | nie | 5 |
| sort | string | nie | updated |
| title-link | string | nie | |

##### `token`
Personal Access Token (PAT). Bez tokena widget wyświetla tylko publiczne repozytoria i podlega limitowi 60 żądań na godzinę. Token zwiększa limit do 5000 żądań na godzinę i umożliwia dostęp do prywatnych repozytoriów.

Aby utworzyć token, przejdź do **Settings → Developer settings → Personal access tokens → Tokens (classic)** i wygeneruj token z uprawnieniem `repo`.

##### `collapse-after`
Liczba repozytoriów widocznych przed pojawieniem się przycisku „Pokaż więcej”.

##### `sort`
Sposób sortowania repozytoriów. Możliwe wartości: `updated`, `created`, `pushed` i `full_name`.

##### `title-link`
Link otwierany po kliknięciu tytułu widgetu. Jeśli nie zostanie podany, tytuł nie jest klikalny.

#### Wyświetlane informacje

Dla każdego repozytorium widget pokazuje nazwę z linkiem do GitHub, opis, czas od ostatniej aktywności, liczbę gwiazdek oraz główny język programowania.

#### Rozwiązywanie problemów

##### Widget wyświetla błąd
Sprawdź, czy token PAT jest poprawny, nie wygasł i ma uprawnienie `repo`.

##### Widoczne są tylko publiczne repozytoria
Token musi mieć uprawnienie `repo`, aby uzyskać dostęp do prywatnych repozytoriów.

##### Przekroczono limit zapytań
Dodaj ważny token PAT, aby zwiększyć limit do 5000 żądań na godzinę, lub zwiększ wartość `cache`.

<a id="repository"></a>
### Repository
Wyświetla ogólne informacje o repozytorium oraz listę najnowszych otwartych pull requestów i zgłoszeń.

Przykład:

```yaml
- type: repository
  repository: glanceapp/glance
  pull-requests-limit: 5
  issues-limit: 3
  commits-limit: 3
```

Podgląd:

![](images/repository-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| repository | string | tak |  |
| token | string | nie | |
| pull-requests-limit | integer | nie | 3 |
| issues-limit | integer | nie | 3 |
| commits-limit | integer | nie | -1 |

##### `repository`
Nazwa właściciela i repozytorium, których informacje mają być wyświetlane.

##### `token`
Bez uwierzytelniania GitHub zezwala na maksymalnie 60 żądań na godzinę. Łatwo przekroczyć ten limit i zacząć widzieć błędy, jeśli masz krótki czas cache lub wiele instancji tego widgetu. Aby temu zaradzić, możesz [utworzyć token tylko do odczytu na swoim koncie GitHub](https://github.com/settings/personal-access-tokens/new) i podać go tutaj.

##### `pull-requests-limit`
Maksymalna liczba najnowszych otwartych pull requestów do wyświetlenia. Ustaw `-1`, aby nie wyświetlać żadnych.

##### `issues-limit`
Maksymalna liczba najnowszych otwartych zgłoszeń do wyświetlenia. Ustaw `-1`, aby nie wyświetlać żadnych.

##### `commits-limit`
Maksymalna liczba najnowszych commitów z gałęzi domyślnej do wyświetlenia. Ustaw `-1`, aby nie wyświetlać żadnych.

<a id="radyjko"></a>
### Radyjko
Wyświetla odtwarzacz radia z listą stacji radiowych streamingowych. Widget zapewnia interfejs przypominający Spotify/YouTube Music do słuchania polskich i zagranicznych stacji radiowych.

Przykład:

```yaml
- type: radyjko
```

Podgląd:

(Widget odtwarzacza z listą stacji, przyciskiem odtwarzania/pauzy i regulacją głośności)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |

Obecnie widget nie ma konfigurowalnych właściwości. Automatycznie pobiera listę dostępnych stacji radiowych ze skonfigurowanego proxy API i wyświetla je w intuicyjnym interfejsie odtwarzacza.

#### Funkcje

- **Odtwarzanie/pauza**: kliknij przycisk odtwarzania, aby rozpocząć lub zatrzymać odtwarzanie
- **Wybór stacji**: kliknij dowolną stację na liście, aby ją wybrać i odtworzyć
- **Regulacja głośności**: dostosuj głośność za pomocą suwaka
- **Responsywny design**: działa bezproblemowo na komputerach, tabletach i urządzeniach mobilnych
- **Obsługa strumieni HLS**: obsługuje zarówno strumienie HLS (m3u8), jak i standardowe strumienie audio
- **Informacja zwrotna**: aktywna stacja jest wyróżniona, aby ułatwić jej identyfikację

<a id="navidrome"></a>
### Navidrome
Odtwarza muzykę z serwera Navidrome (lub innego serwera zgodnego z API Subsonic) bezpośrednio z dashboardu Glance.

Przykład:

```yaml
- type: navidrome
  url: https://navidrome.example.com
  user: twoj_login
  token: twoje_haslo
```

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ----- | --- | -------- | --------- |
| url | string | tak | |
| user | string | tak | |
| token | string | tak | |
| salt | string | nie | generowany automatycznie |

##### `url`
Adres URL serwera Navidrome.

##### `user`
Nazwa użytkownika.

##### `token`
Hasło użytkownika lub token.

##### `salt`
Własny ciąg „salt” używany do autoryzacji. Domyślnie generowany automatycznie.

#### Funkcje

Widget oferuje odtwarzacz z wyborem playlisty, okładką albumu, informacjami o utworze oraz przyciskami Play/Pauza, Poprzedni i Następny. Dostępne są również losowe odtwarzanie (Shuffle), pętla (Loop) i regulacja głośności. Ostatnio wybrana playlista, utwór, poziom głośności oraz stan Shuffle i Loop są zapamiętywane w przeglądarce.

#### Wymagania

Serwer musi obsługiwać endpointy `getPlaylists`, `getPlaylist`, `stream` i `getCoverArt` API Subsonic.

#### Rozwiązywanie problemów

##### Brak playlist
Sprawdź poprawność `url`, `user` i `token`, uprawnienia użytkownika do API oraz dostępność serwera z sieci, w której działa Glance.

##### Utwory się nie odtwarzają
Sprawdź konsolę przeglądarki pod kątem błędów CORS lub Mixed Content (HTTPS/HTTP) oraz upewnij się, że format plików jest obsługiwany przez przeglądarkę.

<a id="bookmarks"></a>
### Zakładki
Wyświetla listę linków, które można grupować.

Przykład:

```yaml
- type: bookmarks
  groups:
    - links:
        - title: Gmail
          url: https://mail.google.com/mail/u/0/
        - title: Amazon
          url: https://www.amazon.com/
        - title: Github
          url: https://github.com/
        - title: Wikipedia
          url: https://en.wikipedia.org/
    - title: Entertainment
      color: 10 70 50
      links:
        - title: Netflix
          url: https://www.netflix.com/
        - title: Disney+
          url: https://www.disneyplus.com/
        - title: YouTube
          url: https://www.youtube.com/
        - title: Prime Video
          url: https://www.primevideo.com/
    - title: Social
      color: 200 50 50
      links:
        - title: Reddit
          url: https://www.reddit.com/
        - title: Twitter
          url: https://twitter.com/
        - title: Instagram
          url: https://www.instagram.com/
```

Podgląd:

![](images/bookmarks-widget-preview.png)


#### Właściwości

| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| groups | array | tak |

##### `groups`
Tablica grup, które mogą opcjonalnie mieć tytuł i własny kolor.

###### Właściwości każdej grupy
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| title | string | nie | |
| color | HSL | nie | kolor primary motywu |
| links | array | tak | |
| same-tab | boolean | nie | false |
| hide-arrow | boolean | nie | false |
| target | string | nie | |

> [!TIP]
>
> Właściwości `same-tab`, `hide-arrow` i `target` możesz ustawić albo na grupie — wtedy będą miały zastosowanie do wszystkich linków w tej grupie — albo na poszczególnych linkach, co nadpisze wartość ustawioną na grupie.

###### Właściwości każdego linku
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| title | string | tak | |
| url | string | tak | |
| description | string | nie | |
| icon | string | nie | |
| same-tab | boolean | nie | false |
| hide-arrow | boolean | nie | false |
| target | string | nie | |

`icon`

Więcej informacji o określaniu ikon znajdziesz w sekcji [Ikony](#icons).

`same-tab`

Określa, czy otwierać link w tej samej, czy w nowej karcie.

`hide-arrow`

Określa, czy ukryć kolorową strzałkę przy każdym linku.

`target`

Ustawia niestandardową wartość atrybutu `target` linku. Możliwe wartości: `_blank`, `_self`, `_parent` i `_top`. Więcej o ich działaniu przeczytasz [tutaj](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/a#target). Ta właściwość ma pierwszeństwo przed `same-tab`.

<a id="changedetectionio"></a>
### ChangeDetection.io
Wyświetla listę obserwowanych stron z changedetection.io.

Przykład

```yaml
- type: change-detection
  instance-url: https://changedetection.mydomain.com/
  token: ${CHANGE_DETECTION_TOKEN}
```

Podgląd:

![](images/change-detection-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| instance-url | string | nie | `https://www.changedetection.io` |
| token | string | nie |  |
| limit | integer | nie | 10 |
| collapse-after | integer | nie | 5 |
| watches | array of strings | nie |  |

##### `instance-url`
Adres URL wskazujący Twoją instancję `changedetection.io`.

##### `token`
Token dostępu do API, który można znaleźć w `SETTINGS > API`. Opcjonalnie możesz podać go za pomocą zmiennej środowiskowej, używając składni `${VARIABLE_NAME}`.

##### `limit`
Maksymalna liczba obserwowanych stron do wyświetlenia.

##### `collapse-after`
Liczba widocznych obserwowanych stron, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

##### `watches`
Domyślnie wyświetlane są wszystkie skonfigurowane obserwowane strony. Opcjonalnie możesz podać listę UUID konkretnych stron, które chcesz wyświetlić:

```yaml
  - type: change-detection
    watches:
      - 1abca041-6d4f-4554-aa19-809147f538d3
      - 705ed3e4-ea86-4d25-a064-822a6425be2c
```

<a id="clock"></a>
### Zegar
Wyświetla zegar pokazujący aktualny czas i datę. Opcjonalnie wyświetla również czas w innych strefach czasowych.

Przykład:

```yaml
- type: clock
  hour-format: 24h
  timezones:
    - timezone: Europe/Paris
      label: Paris
    - timezone: America/New_York
      label: New York
    - timezone: Asia/Tokyo
      label: Tokyo
```

Podgląd:

![](images/clock-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| hour-format | string | nie | 24h |
| timezones | array | nie |  |

##### `hour-format`
Określa, czy wyświetlać czas w formacie 12- czy 24-godzinnym. Możliwe wartości: `12h` i `24h`.

#### Właściwości każdej strefy czasowej

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| timezone | string | tak | |
| label | string | nie | |

##### `timezone`
Identyfikator strefy czasowej, taki jak `Europe/London`, `America/New_York` itp. Pełną listę dostępnych identyfikatorów znajdziesz [tutaj](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).

##### `label`
Opcjonalnie nadpisuje wartość wyświetlaną dla strefy czasowej na coś bardziej zrozumiałego, np. „Home”, „Work” lub cokolwiek innego.


<a id="calendar"></a>
### Kalendarz
Wyświetla kalendarz.

Przykład:

```yaml
- type: calendar
  first-day-of-week: monday
```

Podgląd:

![](images/calendar-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| first-day-of-week | string | nie | monday |

##### `first-day-of-week`
Dzień tygodnia, od którego zaczyna się kalendarz. Jako wartości można podać wszystkie dni tygodnia.

<a id="calendar-legacy"></a>
### Kalendarz (legacy)
Wyświetla kalendarz.

Przykład:

```yaml
- type: calendar-legacy
  start-sunday: false
```

Podgląd:

![](images/calendar-legacy-widget-preview.png)

> [!NOTE]
>
> Ten widget jest przestarzały i może zostać usunięty w przyszłej wersji.

#### Właściwości

| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| start-sunday | boolean | nie | false |

##### `start-sunday`
Określa, czy tygodnie kalendarza zaczynają się w niedzielę, czy w poniedziałek.

> [!NOTE]
>
> Obecnie kalendarz oferuje niewiele możliwości dostosowania. Dodatkowe funkcje zostaną dodane w przyszłości.

<a id="markets"></a>
### Markets
Wyświetla listę rynków, ich aktualną wartość, zmianę w ciągu dnia oraz mały wykres 21-dniowy. Dane pochodzą z Yahoo Finance.

Przykład:

```yaml
- type: markets
  markets:
    - symbol: SPY
      name: S&P 500
    - symbol: BTC-USD
      name: Bitcoin
      chart-link: https://www.tradingview.com/chart/?symbol=INDEX:BTCUSD
    - symbol: NVDA
      name: NVIDIA
    - symbol: AAPL
      symbol-link: https://www.google.com/search?tbm=nws&q=apple
      name: Apple
```

Podgląd:

![](images/markets-widget-preview.png)

#### Właściwości

| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| markets | array | tak |
| sort-by | string | nie |
| chart-link-template | string | nie |
| symbol-link-template | string | nie |

##### `markets`
Tablica rynków, dla których mają być wyświetlane informacje.

##### `sort-by`
Domyślnie rynki są wyświetlane w kolejności ich zdefiniowania. Możesz dostosować kolejność, ustawiając właściwość `sort-by` na `change` (sortowanie malejąco według procentowej zmiany ceny akcji, np. 1% będzie wyżej niż -1%) lub `absolute-change` (sortowanie malejąco według bezwzględnej zmiany ceny, np. -1% będzie wyżej niż +0.5%).

##### `chart-link-template`
Szablon linku otwieranego po kliknięciu wykresu, stosowany do wszystkich rynków. Wartość `{SYMBOL}` zostanie zastąpiona symbolem rynku. Możesz to nadpisać dla konkretnego rynku, podając właściwość `chart-link`. Przykład:

```yaml
chart-link-template: https://www.tradingview.com/chart/?symbol={SYMBOL}
```

##### `symbol-link-template`
Szablon linku otwieranego po kliknięciu symbolu, stosowany do wszystkich rynków. Wartość `{SYMBOL}` zostanie zastąpiona symbolem rynku. Możesz to nadpisać dla konkretnego rynku, podając właściwość `symbol-link`. Przykład:

```yaml
symbol-link-template: https://www.google.com/search?tbm=nws&q={SYMBOL}
```

###### Właściwości każdego rynku
| Nazwa | Typ | Wymagane |
| ---- | ---- | -------- |
| symbol | string | tak |
| name | string | nie |
| symbol-link | string | nie |
| chart-link | string | nie |

`symbol`

Symbol widoczny w Yahoo Finance.

`name`

Nazwa wyświetlana pod symbolem.

`symbol-link`

Link otwierany po kliknięciu symbolu.

`chart-link`

Link otwierany po kliknięciu wykresu.

<a id="twitch-channels"></a>
### Twitch Channels
Wyświetla listę kanałów z Twitcha.

Przykład:

```yaml
- type: twitch-channels
  channels:
    - jembawls
    - giantwaffle
    - asmongold
    - cohhcarnage
    - j_blow
    - xQc
```

Podgląd:

![](images/twitch-channels-widget-preview.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| channels | array | tak | |
| collapse-after | integer | nie | 5 |
| sort-by | string | nie | viewers |

##### `channels`
Lista kanałów do wyświetlenia.

##### `collapse-after`
Liczba widocznych kanałów, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

##### `sort-by`
Może służyć do określenia kolejności wyświetlania kanałów. Możliwe wartości: `viewers` i `live`.

<a id="twitch-top-games"></a>
### Twitch top games
Wyświetla listę gier z największą liczbą widzów na Twitchu.

Przykład:

```yaml
- type: twitch-top-games
  exclude:
    - just-chatting
    - pools-hot-tubs-and-beaches
    - music
    - art
    - asmr
```

Podgląd:

![](images/twitch-top-games-widget-preview.png)

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| exclude | array | nie | |
| limit | integer | nie | 10 |
| collapse-after | integer | nie | 5 |

##### `exclude`
Lista kategorii, które nigdy nie będą wyświetlane. Musisz podać slug znaleziony po kliknięciu kategorii i sprawdzeniu adresu URL:

```
https://www.twitch.tv/directory/category/grand-theft-auto-v
                                         ^^^^^^^^^^^^^^^^^^
```

##### `limit`
Maksymalna liczba gier do wyświetlenia.

##### `collapse-after`
Liczba widocznych gier, po której pojawia się przycisk „SHOW MORE”. Ustaw `-1`, aby nigdy nie zwijać.

<a id="iframe"></a>
### iframe
Osadza iframe jako widget.

Przykład:

```yaml
- type: iframe
  source: <url>
  height: 400
```

#### Właściwości
| Nazwa | Typ | Wymagane | Domyślnie |
| ---- | ---- | -------- | ------- |
| source | string | tak | |
| height | integer | nie | 300 |

##### `source`
Źródło iframe'a.

##### `height`
Wysokość iframe'a. Minimalna dozwolona wysokość to 50.

<a id="html"></a>
### HTML
Osadza dowolny kod HTML.

Przykład:

```yaml
- type: html
  source: |
    <p>Hello, <span class="color-primary">World</span>!</p>
```

<a id="debugowanie"></a>
## Debugowanie

### API Timing
Jeśli Glance ładuje się wolniej niż powinien, możesz włączyć debugowanie requestów API, ustawiając zmienną środowiskową `DEBUG_API_TIMING`:

```yaml
# docker-compose.yml
services:
  glance:
    environment:
      - DEBUG_API_TIMING=true
```

Lub jako komentarz w docker-compose (odkomentuj, aby włączyć):

```yaml
environment:
  - TZ=Europe/Warsaw
  # Debug API timing - odkomentuj aby włączyć logowanie requestów API
  # - DEBUG_API_TIMING=true
```

Każdy request API będzie wtedy logowany w formacie:

```
[API #1] GET https://api.cloudflare.com/client/v4/zones/xxx - 200 - 234ms
[API #2] POST https://api.cloudflare.com/client/v4/graphql - 200 - 156ms
```

Gdzie:
- `#1` — kolejność wykonania requestu
- `GET/POST` — metoda HTTP
- `200` — kod odpowiedzi
- `234ms` — czas wykonania requestu

To pozwala zidentyfikować, które widgety/API wydłużają czas ładowania panelu.

Zwróć uwagę na użycie `|` po `source:` — pozwala to wstawić ciąg znaków wielowierszowy.
