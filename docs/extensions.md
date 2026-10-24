# Rozszerzenia

> [!IMPORTANT]
>
> **Ten dokument oraz funkcja rozszerzeń są w trakcie opracowywania. API może się w przyszłości zmienić. Odpowiadasz za utrzymanie własnych rozszerzeń.**

## Przegląd

Aby tworzenie rozszerzeń wymagało minimalnej wiedzy, zamiast skomplikowanego protokołu są one niczym więcej jak żądaniem HTTP do serwera, który zwraca kilka specjalnych nagłówków. Wymianę danych między Glance a rozszerzeniami przedstawia poniższy diagram:

![](images/extension-overview.png)

Jeśli wiesz, jak skonfigurować serwer HTTP i znasz trochę HTML oraz CSS, możesz zacząć tworzyć własne rozszerzenia.

> [!TIP]
>
> Domyślnie widget rozszerzenia ma czas cache wynoszący 30 minut. Aby uniknąć restartowania Glance po każdej zmianie rozszerzenia, możesz ustawić czas cache widgetu na 1 sekundę:
> ```yaml
> - type: extension
>   url: http://localhost:8081
>   cache: 1s
> ```

## Nagłówki

### `Widget-Title`
Służy do określenia tytułu widgetu. Jeśli nie zostanie podany, tytułem widgetu będzie „Extension”.

### `Widget-Title-URL`
Służy do określenia adresu URL otwieranego po kliknięciu tytułu widgetu. Jeśli użytkownik podał `title-url` w konfiguracji, ma on pierwszeństwo przed tym nagłówkiem.

### `Widget-Content-Type`
Służy do określenia typu zawartości zwracanej przez rozszerzenie. Jeśli nie zostanie podany, zawartość zostanie wyświetlona jako zwykły tekst.

### `Widget-Content-Frameless`
Po ustawieniu na `true` zawartość widgetu będzie wyświetlana bez domyślnego tła ani „ramki”.

## Typy zawartości

> [!NOTE]
>
> Obecnie jedynym obsługiwanym typem zawartości jest `html`. Celem długoterminowym jest obsługa ogólnych typów zawartości, takich jak `videos`, `forum-posts`, `markets`, `streams` itp., które będą zwracane w formacie JSON i wyświetlane przez Glance z użyciem istniejących stylów i funkcji. Dzięki temu autorzy rozszerzeń uzyskają natywny wygląd, skupiając się wyłącznie na dostarczaniu danych z wybranego źródła.

### `html`
Wyświetla zawartość jako HTML. Wymaga to ustawienia przez użytkownika właściwości `allow-potentially-dangerous-html` na `true`; w przeciwnym razie zawartość zostanie wyświetlona jako zwykły tekst.


#### Korzystanie z istniejących klas i funkcji
Większość funkcji dostępnych w Glance można łatwo wykorzystać we własnych rozszerzeniach HTML. Poniżej znajduje się przykład niektórych z nich:

```html
<p class="color-subdue">Tekst w wyciszonym kolorze</p>
<p>Tekst w kolorze podstawowym</p>
<p class="color-highlight">Tekst w wyróżnionym kolorze</p>
<p class="color-primary">Tekst w kolorze primary</p>
<p class="color-positive">Tekst w kolorze pozytywnym</p>
<p class="color-negative">Tekst w kolorze negatywnym</p>

<hr class="margin-block-15">

<p class="size-h1">Rozmiar czcionki 1</p>
<p class="size-h2">Rozmiar czcionki 2</p>
<p class="size-h3">Rozmiar czcionki 3</p>
<p class="size-h4">Rozmiar czcionki 4</p>
<p class="size-base">Rozmiar czcionki podstawowy</p>
<p class="size-h5">Rozmiar czcionki 5</p>
<p class="size-h6">Rozmiar czcionki 6</p>

<hr class="margin-block-15">

<a class="visited-indicator" href="#notvisitedprobably">Link ze wskaźnikiem odwiedzenia</a>

<hr class="margin-block-15">

<a class="color-primary-if-not-visited" href="#notvisitedprobably">Link w kolorze primary, jeśli nie został odwiedzony</a>

<hr class="margin-block-15">

<p>Wydarzenie miało miejsce <span data-dynamic-relative-time="<unix timestamp>"></span> temu</p>

<hr class="margin-block-15">

<ul class="list-horizontal-text">
    <li>pozioma</li>
    <li>lista</li>
    <li>z</li>
    <li>wieloma</li>
    <li>elementami</li>
    <li>tekstu</li>
</ul>

<hr class="margin-block-15">

<ul class="list list-gap-10 list-with-separator">
    <li>lista</li>
    <li>z</li>
    <li>odstępami</li>
    <li>i</li>
    <li>poziomymi</li>
    <li>liniami</li>
</ul>

<hr class="margin-block-15">

<ul class="list collapsible-container" data-collapse-after="3">
    <li>zwijana</li>
    <li>lista</li>
    <li>z</li>
    <li>wieloma</li>
    <li>elementami</li>
    <li>które</li>
    <li>pojawią</li>
    <li>się</li>
    <li>po</li>
    <li>kliknięciu</li>
    <li>przycisku</li>
    <li>poniżej</li>
    <li>na</li>
    <li>liście</li>
</ul>

<hr class="margin-bottom-15">

<p class="margin-bottom-10">Obraz ładowany leniwie:</p>

<img src="https://picsum.photos/200" alt="" loading="lazy">

<hr class="margin-block-15">

<p class="margin-bottom-10">Lista postów:</p>

<ul class="list list-gap-14 collapsible-container" data-collapse-after="5">
    <li>
        <a class="size-h3 color-primary-if-not-visited" href="#link">Lorem ipsum dolor, sit amet consectetur adipisicing elit. Voluptatum, ipsa?</a>
        <ul class="list-horizontal-text">
            <li data-dynamic-relative-time="<unix timestamp>"></li>
            <li>3 321 punktów</li>
            <li>139 komentarzy</li>
        </ul>
    </li>
    <li>
        <a class="size-h3 color-primary-if-not-visited" href="#link">Lorem ipsum dolor, sit amet consectetur adipisicing elit. Voluptatum, ipsa?</a>
        <ul class="list-horizontal-text">
            <li data-dynamic-relative-time="<unix timestamp>"></li>
            <li>3 321 punktów</li>
            <li>139 komentarzy</li>
        </ul>
    </li>
    <li>
        <a class="size-h3 color-primary-if-not-visited" href="#link">Lorem ipsum dolor, sit amet consectetur adipisicing elit. Voluptatum, ipsa?</a>
        <ul class="list-horizontal-text">
            <li data-dynamic-relative-time="<unix timestamp>"></li>
            <li>3 321 punktów</li>
            <li>139 komentarzy</li>
        </ul>
    </li>
</ul>
```

Wszystko to da następujący efekt:

![](images/extension-html-reusing-existing-features-preview.png)

**Nazwy klas i funkcje mogą się zmienić — ponownie: odpowiadasz za utrzymanie własnych rozszerzeń.**
