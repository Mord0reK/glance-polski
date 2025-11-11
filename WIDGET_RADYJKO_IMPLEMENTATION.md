# Widget Radyjko - Instrukcja Implementacji

## Opis
Widget **Radyjko** to inteligentny player stacji radiowych dla dashboardu Glance. Widget zapewnia interfejs podobny do YouTube Music lub Spotify, ale zamiast utworów, oferuje dostęp do polskich i międzynarodowych stacji radiowych.

## Co zostało zaimplementowane

### 1. Backend (Go)
**Plik:** `internal/glance/widget-radyjko.go`
- Struktura `radyjkoWidget` z metodami `initialize()`, `update()` i `Render()`
- Funkcja `fetchRadioStations()` pobierająca listę stacji z API
- Struktura `station` reprezentująca jedną stację radiową
- Obsługa cache'u (24 godziny)

### 2. Szablon HTML
**Plik:** `internal/glance/templates/radyjko.html`
- Responsywny layout playera
- Sekcja nagłówka z wyświetlaniem aktualnie odtwarzanej stacji
- Kontrolki (play/pause, kontrola głośności)
- Scrollowalny list stacji radiowych
- Element audio do odtwarzania

### 3. Style CSS
**Plik:** `internal/glance/static/css/widget-radyjko.css`
- Design wzorowany na nowoczesnych playerach muzycznych
- Gradient tło
- Animacje przejść i hover'ów
- Responsywny design dla mobilnych urządzeń
- Dostosowane kolory do palety Glance'a

### 4. Logika JavaScript
**Plik:** `internal/glance/static/js/radyjko.js`
- Klasa `RadyjkoPlayer` zarządzająca stanem playera
- Obsługa play/pause
- Kontrola głośności
- Zmiana stacji
- Obsługa strumieni HLS (m3u8)
- Visual feedback (aktywna stacja)

### 5. Integracja
- Dodanie typu widgetu "radyjko" do `widget.go`
- Import CSS do `static/css/widgets.css`
- Setup funkcji w `static/js/page.js`
- Dokumentacja w `docs/configuration.md`

## Konfiguracja

Aby dodać widget Radyjko do dashboardu, dodaj następujące linie do `glance.yml`:

```yaml
- type: radyjko
```

## Wymagania

Widget wymaga:
- API proxy dostępne pod `https://proxy.mordorek.dev/stations`
- Przeglądarki wspierającej HTML5 Audio API
- JavaScriptu włączonego w przeglądarce

## Cechy

✅ Play/Pause Control - kliknij przycisk do uruchomienia/zatrzymania odtwarzania
✅ Wybór Stacji - kliknij na stację, aby ją wybrać i odtworzyć
✅ Kontrola Głośności - suwak do regulacji głośności
✅ Responsywny Design - działa na desktop, tablet i mobile
✅ Obsługa HLS - wspiera zarówno strumienie HLS (m3u8) jak i standardowe audio
✅ Visual Feedback - aktywna stacja jest podświetlona

## Testy

Aby przetestować widget:

1. Zbuduj projekt: `go build -o glance .`
2. Uruchom: `./glance`
3. Dodaj widget Radyjko do konfiguracji `glance.yml`
4. Otwórz dashboard w przeglądarce
5. Sprawdź czy lista stacji się ładuje i player działa

## Pliki zmienione/utworzone

- ✅ `internal/glance/widget-radyjko.go` - NOWY
- ✅ `internal/glance/templates/radyjko.html` - NOWY
- ✅ `internal/glance/static/css/widget-radyjko.css` - NOWY
- ✅ `internal/glance/static/js/radyjko.js` - NOWY
- ✅ `internal/glance/widget.go` - ZMODYFIKOWANY
- ✅ `internal/glance/static/css/widgets.css` - ZMODYFIKOWANY
- ✅ `internal/glance/static/js/page.js` - ZMODYFIKOWANY
- ✅ `docs/configuration.md` - ZMODYFIKOWANY

## Notatki

- Widget pobiera listę stacji z tego samego API co strona Radyjko (`https://proxy.mordorek.dev/stations`)
- Elementy graficzne (SVG) są wbudowane bezpośrednio w HTML
- Style używają zmiennych CSS z palety Glance'a (`--color-primary`, `--color-text-*`, itp.)
- Player obsługuje dynamiczne przychodzące strumienie HLS dzięki bibliotece HLS.js
