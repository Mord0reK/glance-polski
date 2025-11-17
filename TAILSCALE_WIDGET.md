# Natywny Widget Tailscale

## Opis
Natywny widget Tailscale dla Glance oferujący większe możliwości i łatwiejszą konfigurację niż wersja custom-api.

## Funkcje

### Co oferuje natywny widget:
- ✅ Łatwiejsza konfiguracja (bez potrzeby pisania szablonów HTML)
- ✅ Automatyczne parsowanie danych z API Tailscale
- ✅ Zachowana kolorystyka z wersji custom-api
- ✅ Wskaźniki:
  - Aktualizacji dostępnych (niebieski punkt)
  - Status online/offline (zielony/czerwony punkt)
  - Informacje o ostatniej aktywności
- ✅ Efekty hover pokazujące adres IP urządzenia
- ✅ **Łatwe kopiowanie IP jednym kliknięciem** (kliknij bezpośrednio na IP)
- ✅ Wizualny feedback przy kopiowaniu (tło zmienia się na zielone z ✓)
- ✅ Działa w HTTP i HTTPS (fallback dla starszych przeglądarek)
- ✅ Możliwość kontrolowania liczby widocznych urządzeń
- ✅ Opcjonalne pokazywanie wskaźnika "online"

### Możliwości rozszerzenia w przyszłości:
- Zarządzanie urządzeniami
- Włączanie/wyłączanie tras
- Zarządzanie kluczami API
- Statystyki ruchu
- Powiadomienia o zmianach

## Konfiguracja

### Minimalna konfiguracja:
```yaml
- type: tailscale
  token: your-tailscale-api-token
```

### Pełna konfiguracja:
```yaml
- type: tailscale
  title: Tailscale                                         # Opcjonalny, domyślnie "Tailscale"
  title-url: https://login.tailscale.com/admin/machines    # Opcjonalny
  token: your-tailscale-api-token                          # Wymagany
  tailnet: "-"                                             # Opcjonalny, domyślnie "-" (current tailnet)
  url: https://api.tailscale.com/api/v2/tailnet/-/devices  # Opcjonalny, można nadpisać URL API
  cache: 10m                                               # Opcjonalny, domyślnie 10m
  collapse-after: 4                                        # Opcjonalny, domyślnie 4
  show-online-indicator: false                             # Opcjonalny, domyślnie false
```

## Parametry

### `token` (wymagany)
Token API Tailscale. Możesz wygenerować go w panelu administracyjnym Tailscale:
- Przejdź do https://login.tailscale.com/admin/settings/keys
- Kliknij "Generate API access token"
- Skopiuj token i użyj go w konfiguracji

### `tailnet` (opcjonalny)
Nazwa tailnet. Domyślnie "-" oznacza bieżący tailnet. Możesz podać konkretną nazwę, jeśli masz dostęp do wielu tailnetów.

### `url` (opcjonalny)
Niestandardowy URL API. Domyślnie widget używa oficjalnego API Tailscale.

### `cache` (opcjonalny)
Czas cache'owania danych. Domyślnie 10 minut. Przykłady: `5m`, `1h`, `30s`.

### `collapse-after` (opcjonalny)
Liczba urządzeń widocznych przed przyciskiem "SHOW MORE". Domyślnie 4. Ustaw na `-1`, aby nigdy nie zwijać listy.

### `show-online-indicator` (opcjonalny)
Czy pokazywać zielony wskaźnik dla urządzeń online. Domyślnie `false` (pokazywany jest tylko czerwony wskaźnik dla urządzeń offline).

## Wizualne elementy

Widget zachowuje całą kolorystykę z wersji custom-api:
- **Kolor podstawowy** (`--color-primary`) - nazwa urządzenia i tło IP po hover
- **Kolor pozytywny** (`--color-positive`) - wskaźnik online (jeśli włączony) i tło IP po skopiowaniu
- **Kolor negatywny** (`--color-negative`) - wskaźnik offline
- **Kolor podstawowy** (`--color-primary`) - wskaźnik dostępnej aktualizacji

### Kopiowanie adresu IP
Po najechaniu na wiersz urządzenia:
1. Zamiast informacji o systemie i użytkowniku pojawia się adres IP
2. Adres IP jest klikalny (hover zmienia tło na niebieski)
3. Kliknięcie w IP kopiuje je do schowka
4. Po skopiowaniu tło zmienia się na zielone i pojawia się ✓ na 2 sekundy
5. Działa w każdej przeglądarce dzięki mechanizmowi fallback

## Przykładowe zastosowania

### Podstawowy monitoring:
```yaml
- type: tailscale
  token: ${TAILSCALE_TOKEN}
```

### Monitoring ze wskaźnikami online:
```yaml
- type: tailscale
  token: ${TAILSCALE_TOKEN}
  show-online-indicator: true
  collapse-after: 10
```

### Monitoring z niestandardowym cache:
```yaml
- type: tailscale
  token: ${TAILSCALE_TOKEN}
  cache: 5m
  title: Moja Sieć Tailscale
```
