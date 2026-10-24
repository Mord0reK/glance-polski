<h1>To jest zmodyfikowana wersja dashboardu Glance, z elementami przetłumaczonymi na język polski</h1>
<p>Wersja ta zawiera także następujące zmiany:</p>
<ul>
  <li>Widget Radyjko</li>
  <li>Natywny widget Cloudflare</li>
  <li>Natywny widget Navidrome</li>
  <li>Natywny widget Vikunja</li>
  <li>Natywny widget Beszel</li>
  <li>Natywny widget Tailscale</li>
  <li>Natywny widget qBittorrent</li>
  <li>Natywny widget GitHub</li>
  <li>Widget OpenCode Go z prognozą zużycia</li>
  <li>Integracja z Affine</li>
  <li>Poprawiony widget pogody</li>
  <li>Możliwość przesuwania widgetów między lewą, środkową i prawą kolumną na urządzeniach mobilnych</li>
  <li>Przetłumaczona dokumentacja w katalogu <code>docs/</code></li>
</ul>

<h1>Poniżej znajduje się przetłumaczona reszta oryginalnego pliku README.md z repozytorium glanceapp/glance</h1>

<p align="center"><img src="docs/logo.png"></p>
<h1 align="center">Glance</h1>
<p align="center">
  <a href="#instalacja">Instalacja</a> •
  <a href="docs/configuration.md#configuring-glance">Konfiguracja</a> •
  <a href="https://discord.com/invite/7KQ7Xa9kJd">Discord</a> •
  <a href="https://github.com/sponsors/glanceapp">Wesprzyj projekt</a>
</p>
<p align="center">
  <a href="https://github.com/glanceapp/community-widgets">Widgety społeczności</a> •
  <a href="docs/preconfigured-pages.md">Gotowe strony</a> •
  <a href="docs/themes.md">Motywy</a>
</p>

<p align="center">Lekki, wysoce konfigurowalny dashboard, który wyświetla<br> Twoje kanały w pięknym, przejrzystym interfejsie</p>

![](docs/images/readme-main-image.png)

## Funkcje
### Różnorodne widgety
* Kanały RSS
* Wpisy z subredditów
* Wpisy z Hacker News
* Prognoza pogody
* Filmy z kanałów YouTube
* Kanały Twitch
* Ceny rynkowe
* Status kontenerów Docker
* Statystyki serwera
* Własne widgety
* [I wiele więcej...](docs/configuration.md#configuring-glance)

### Szybki i lekki
* Niskie zużycie pamięci
* Niewiele zależności
* Minimalna ilość czystego JavaScriptu
* Pojedyncza binarka mniejsza niż 20 MB dostępna dla wielu systemów i architektur oraz równie mały kontener Docker
* Niezcache'owane strony zwykle ładują się w około 1 sekundę (w zależności od prędkości internetu i liczby widgetów)

### Ogromne możliwości dostosowania
* Różne układy
* Dowolna liczba stron i kart
* Liczne opcje konfiguracji każdego widgetu
* Wiele stylów dla niektórych widgetów
* Własny CSS

### Zoptymalizowany pod urządzenia mobilne
Ponieważ z pewnością zechcesz zabrać go ze sobą w drogę.

![](docs/images/mobile-preview.png)

### Motywy
Łatwo stwórz własny motyw, zmieniając kilka liczb, albo wybierz jeden z [gotowych motywów](docs/themes.md).

![](docs/images/themes-example.png)

<br>

## Konfiguracja
Konfiguracja odbywa się za pomocą plików YAML. Aby dowiedzieć się więcej o układzie strony, dodawaniu kolejnych stron i konfiguracji widgetów, zajrzyj do [dokumentacji konfiguracji](docs/configuration.md#configuring-glance).
<details>
<summary><strong>Podgląd przykładowego pliku konfiguracyjnego</strong></summary>
<br>

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: calendar
            first-day-of-week: monday

          - type: rss
            limit: 10
            collapse-after: 3
            cache: 12h
            feeds:
              - url: https://selfh.st/rss/
                title: selfh.st
                limit: 4
              - url: https://ciechanow.ski/atom.xml
              - url: https://www.joshwcomeau.com/rss.xml
                title: Josh Comeau
              - url: https://samwho.dev/rss.xml
              - url: https://ishadeed.com/feed.xml
                title: Ahmad Shadeed

          - type: twitch-channels
            channels:
              - theprimeagen
              - j_blow
              - piratesoftware
              - cohhcarnage
              - christitustech
              - EJ_SA

      - size: full
        widgets:
          - type: group
            widgets:
              - type: hacker-news
              - type: lobsters

          - type: videos
            channels:
              - UCXuqSBlHAE6Xw-yeJA0Tunw # Linus Tech Tips
              - UCR-DXc1voovS8nhAvccRZhg # Jeff Geerling
              - UCsBjURrPoezykLs9EqgamOA # Fireship
              - UCBJycsmduvYEL83R_U4JriQ # Marques Brownlee
              - UCHnyfMqiRRG1u-2MsSQLbXA # Veritasium

          - type: group
            widgets:
              - type: reddit
                subreddit: technology
                show-thumbnails: true
              - type: reddit
                subreddit: selfhosted
                show-thumbnails: true

      - size: small
        widgets:
          - type: weather
            location: London, United Kingdom
            units: metric
            hour-format: 12h

          - type: markets
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

          - type: releases
            cache: 1d
            repositories:
              - glanceapp/glance
              - go-gitea/gitea
              - immich-app/immich
              - syncthing/syncthing
```
</details>

<br>

## Instalacja

Wybierz jedną z poniższych metod:

> [!NOTE]
>
> Ten fork jest publikowany jako obraz Docker `ghcr.io/mord0rek/glance-polski`. Poniższe przykłady korzystają właśnie z niego, a nie z obrazu upstreamowego `glanceapp/glance`.

<details>
<summary><strong>Docker Compose z gotową strukturą katalogów (zalecane)</strong></summary>
<br>

Utwórz nowy katalog o nazwie `glance` wraz z plikami szablonu, uruchamiając:

```bash
mkdir glance && cd glance && curl -sL https://github.com/glanceapp/docker-compose-template/archive/refs/heads/main.tar.gz | tar -xzf - --strip-components 2
```

*[Kliknij tutaj, aby zobaczyć pliki, które zostaną utworzone](https://github.com/glanceapp/docker-compose-template/tree/main/root)*

Następnie w pliku `docker-compose.yml` ustaw obraz na wersję z tego forka:

```yaml
services:
  glance:
    image: ghcr.io/mord0rek/glance-polski
```

Następnie edytuj pliki według potrzeb:
* `docker-compose.yml`, aby skonfigurować port, wolumeny i inne ustawienia kontenera
* `config/home.yml`, aby skonfigurować widgety lub układ strony głównej
* `config/glance.yml`, jeśli chcesz zmienić motyw lub dodać kolejne strony

<details>
<summary>Inne pliki, które możesz chcieć edytować</summary>

* `.env`, aby skonfigurować zmienne środowiskowe dostępne w plikach konfiguracyjnych
* `assets/user.css`, aby dodać własny CSS
</details>

Gdy wszystko będzie gotowe, uruchom:

```bash
docker compose up -d
```

Jeśli napotkasz problemy, możesz sprawdzić logi, uruchamiając:

```bash
docker compose logs
```

<hr>
</details>

<details>
<summary><strong>Docker Compose — konfiguracja ręczna</strong></summary>
<br>

Utwórz plik `docker-compose.yml` o następującej zawartości:

```yaml
services:
  glance:
    container_name: glance
    image: ghcr.io/mord0rek/glance-polski
    restart: unless-stopped
    volumes:
      - ./config:/app/config
    ports:
      - 8080:8080
```

Następnie utwórz katalog o nazwie `config` i pobierz do niego przykładowy plik [`glance.yml`](https://github.com/glanceapp/glance/blob/main/docs/glance.yml), uruchamiając:

```bash
mkdir config && wget -O config/glance.yml https://raw.githubusercontent.com/glanceapp/glance/refs/heads/main/docs/glance.yml
```

Możesz swobodnie edytować plik `glance.yml` według własnych potrzeb. Gdy wszystko będzie gotowe, uruchom:

```bash
docker compose up -d
```

Jeśli napotkasz problemy, możesz sprawdzić logi, uruchamiając:

```bash
docker logs glance
```

<hr>
</details>

<details>
<summary><strong>Ręczna instalacja binarki</strong></summary>
<br>

Ten fork nie publikuje gotowych binarek. Możesz zbudować go ze źródeł (patrz sekcja [Budowanie ze źródeł](#budowanie-ze-zrodel)) albo użyć obrazu Docker opisanego powyżej.

Jeśli budujesz binarkę samodzielnie, domyślnie będzie ona szukać pliku `glance.yml` w katalogu, w którym się znajduje. Aby wskazać inną ścieżkę do pliku konfiguracyjnego, użyj opcji `--config`:

```bash
/opt/glance/glance --config /etc/glance.yml
```

<hr>
</details>

<details>
<summary><strong>Inne</strong></summary>
<br>

Glance można również zainstalować za pośrednictwem następujących kanałów firm trzecich:

* [Proxmox VE Helper Script](https://community-scripts.github.io/ProxmoxVE/scripts?id=glance)
* [Pakiet NixOS](https://search.nixos.org/packages?channel=unstable&show=glance)
* [Coolify.io](https://coolify.io/docs/services/glance/)

> [!NOTE]
>
> Powyższe kanały instalują wersję upstreamową Glance, a nie ten fork. Aby korzystać z polskiej wersji, użyj obrazu Docker `ghcr.io/mord0rek/glance-polski`.

<hr>
</details>

<br>

## Częste problemy
<details>
<summary><strong>Przekroczenia czasu żądań</strong></summary>

Najczęstszą przyczyną jest korzystanie z Pi-hole, AdGuard Home lub innych usług DNS blokujących reklamy, które domyślnie mają dość niski limit zapytań. W zależności od liczby widgetów na jednej stronie limit ten można bardzo łatwo przekroczyć. Aby to naprawić, zwiększ limit zapytań w ustawieniach swojej usługi DNS.

Jeśli używasz Podmana, w rzadkich przypadkach przekroczenie czasu może być spowodowane nieznanym problemem — może je rozwiązać dodanie następującego fragmentu na końcu pliku `docker-compose.yml`:

```yaml
networks:
  podman:
    external: true
```
</details>

<details>
<summary><strong>Zepsuty układ widgetów markets, bookmarks lub innych</strong></summary>

Prawie zawsze jest to spowodowane rozszerzeniem przeglądarki Dark Reader. Aby to naprawić, wyłącz tryb ciemny dla domeny, na której hostowany jest Glance.
</details>

<details>
<summary><strong>cannot unmarshal !!map into []glance.page</strong></summary>

Najczęstszą przyczyną jest posiadanie klucza `pages` w pliku `glance.yml`, a następnie także klucza `pages` w jednym z dołączanych plików. Aby to naprawić, usuń klucz `pages` z góry dołączanych stron.

</details>

<br>

## FAQ
<details>
<summary><strong>Czy informacje na stronie aktualizują się automatycznie?</strong></summary>

Nie, aby zaktualizować informacje, trzeba odświeżyć stronę. Niektóre elementy aktualizują się dynamicznie, tam gdzie ma to sens — na przykład widget zegara oraz czas względny pokazujący, jak dawno coś się wydarzyło.
</details>

<details>
<summary><strong>Jak często aktualizują się widgety?</strong></summary>

W tle nie są wykonywane żadne cykliczne żądania — informacje są pobierane tylko przy ładowaniu strony, a następnie cache'owane. Domyślny czas życia cache jest inny dla każdego widgetu i można go skonfigurować.
</details>

<details>
<summary><strong>Czy mogę tworzyć własne widgety?</strong></summary>

Tak, istnieje kilka sposobów tworzenia własnych widgetów:

* widget `iframe` — umożliwia osadzanie treści z innych stron
* widget `html` — umożliwia wstawienie własnego statycznego HTML
* widget `extension` — pobiera HTML z adresu URL
* widget `custom-api` — pobiera JSON z adresu URL i renderuje go za pomocą własnego HTML
</details>

<details>
<summary><strong>Czy mogę zmienić tytuł widgetu?</strong></summary>

Tak, tytuł każdego widgetu można zmienić, podając właściwość `title` w konfiguracji widgetu:

```yaml
- type: rss
  title: My custom title

- type: markets
  title: My custom title

- type: videos
  title: My custom title

# i tak dalej dla wszystkich widgetów...
```
</details>

<br>

## Propozycje funkcji

Nowe propozycje funkcji są zawsze mile widziane i będą rozważane, pamiętaj jednak, że niektóre z nich mogą wykraczać poza zakres projektu (lub możliwości jego realizacji). Jeśli masz pomysł na nową funkcję i chcesz go zgłosić, możesz to zrobić [tutaj](https://github.com/Mord0rek/glance-polski/issues).

Propozycje funkcji są oznaczane jedną z następujących etykiet:

* [Roadmap](https://github.com/glanceapp/glance/labels/roadmap) — zostanie zrealizowana w przyszłym wydaniu
* [Backlog](https://github.com/glanceapp/glance/labels/backlog) — może zostać zrealizowana w przyszłości, ale wymaga dalszych informacji zwrotnych lub zainteresowania społeczności
* [Icebox](https://github.com/glanceapp/glance/labels/icebox) — brak planów realizacji, ponieważ nie jest obecnie zgodna z celami lub możliwościami projektu; może zostać ponownie rozważona w przyszłości

<br>

<a id="budowanie-ze-zrodel"></a>
## Budowanie ze źródeł

Wybierz jedną z poniższych metod:

<details>
<summary><strong>Budowanie binarki za pomocą Go</strong></summary>
<br>

Wymagania: [Go](https://go.dev/dl/) >= v1.23

Aby zbudować projekt dla bieżącego systemu i architektury, uruchom:

```bash
go build -o build/glance .
```

Aby zbudować dla konkretnego systemu i architektury, uruchom:

```bash
GOOS=linux GOARCH=amd64 go build -o build/glance .
```

[*Kliknij tutaj, aby zobaczyć pełną listę kombinacji GOOS i GOARCH*](https://go.dev/doc/install/source#:~:text=$GOOS%20and%20$GOARCH)

Alternatywnie, jeśli chcesz tylko uruchomić aplikację bez tworzenia binarki — na przykład podczas testowania zmian — możesz uruchomić:

```bash
go run .
```
<hr>
</details>

<details>
<summary><strong>Budowanie projektu i obrazu Docker za pomocą Docker</strong></summary>
<br>

Wymagania: [Docker](https://docs.docker.com/engine/install/)

Aby zbudować projekt i obraz wyłącznie za pomocą Docker, uruchom:

*(zamień `owner` na swoją nazwę lub nazwę organizacji)*

```bash
docker build -t owner/glance:latest .
```

Jeśli chcesz wypchnąć obraz do rejestru (domyślnie Docker Hub), uruchom:

```bash
docker push owner/glance:latest
```

<hr>
</details>

<br>

## Zasady współtworzenia

* Przed rozpoczęciem pracy nad nową funkcją najlepiej najpierw zgłosić propozycję funkcji i zaznaczyć, że chcesz ją zaimplementować samodzielnie
* Nie zgłaszaj PR-ów dla propozycji funkcji, które znajdują się w roadmap<sup>[1]</sup>, backlog<sup>[2]</sup> lub icebox<sup>[3]</sup>
* Jeśli dodajesz nowe funkcje lub poprawiasz błędy, używaj gałęzi `dev` jako bazowej; w przeciwnym razie używaj `main`
* Unikaj wprowadzania nowych zależności
* Unikaj wprowadzania niekompatybilnych wstecznie zmian konfiguracji
* Unikaj wprowadzania nowych kolorów lub hardkodowania kolorów — używaj standardowych `primary`, `positive` i `negative`
* W przypadku ikon staraj się używać [heroicons](https://heroicons.com/), gdzie to możliwe
* Jeśli zmiany dotyczą interfejsu, w miarę możliwości dołącz zrzut ekranu
* Nie dodawaj `package.json`

<details>
<summary><strong><sup>[1] [2] [3]</sup></strong></summary>

[1] Funkcja prawdopodobnie ma już rozpoczęte prace, które mogą kolidować z Twoją implementacją

[2] Popyt, implementacja lub funkcjonalność tej funkcji nie są jeszcze jasne

[3] Brak planów dodania tej funkcji w najbliższym czasie

</details>

<br>

## Podziękowania

Dla wszystkich osób, które były wystarczająco hojne, aby [wesprzeć](https://github.com/sponsors/glanceapp) projekt, oraz dla wszystkich, którzy w jakikolwiek sposób przyczynili się do jego rozwoju — poprzez PR-y, zgłaszanie problemów, pomoc innym w dyskusjach lub na serwerze Discord, tworzenie poradników i narzędzi albo po prostu wspominanie o Glance w mediach społecznościowych. Wasze wsparcie jest bardzo doceniane i pomaga utrzymać projekt przy życiu.
