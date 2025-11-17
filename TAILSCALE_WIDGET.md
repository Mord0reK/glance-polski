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
- ✅ **Znaczniki funkcji urządzenia (dane z API):**
  - **Expiry disabled** - czy klucz nie wygasa (cyjanowy #17a2b8)
  - **Disconnected** - urządzenie nie połączone z panelem kontrolnym (czerwony #dc3545)
  - **Blocks Incoming** - blokuje przychodzące połączenia (żółty #ffc107)
  - **Joined [data]** - kiedy urządzenie dołączyło do sieci (szary #6c757d)
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
  token: twoj_token
```

### Pełna konfiguracja:
```yaml
- type: tailscale
  title: Tailscale                                         # Opcjonalny, domyślnie "Tailscale"
  title-url: https://login.tailscale.com/admin/machines    # Opcjonalny
  token: twoj_token                         # Wymagany
  tailnet: "-"                                             # Opcjonalny, domyślnie "-" (current tailnet)
  url: https://api.tailscale.com/api/v2/tailnet/-/devices  # Opcjonalny, można nadpisać URL API
  cache: 10m                                               # Opcjonalny, domyślnie 10m
  collapse-after: 4                                        # Opcjonalny, zwija listę po N urządzeniach
  show-online-indicator: true                              # Opcjonalny, domyślnie false
  
  # Kontrola wyświetlania znaczników (domyślnie wszystkie false)
  show-expiry-disabled: true   # 🔵 Pokaż "Expiry disabled"
  show-disconnected: true      # 🔴 Pokaż "Disconnected"  
  show-blocks-incoming: true   # 🟡 Pokaż "Blocks Incoming"
  show-joined-date: true       # ⚫ Pokaż datę dołączenia
```

---

## Szczegółowy opis opcji konfiguracji

### 🔐 `token` (WYMAGANE)
```yaml
token: twoj_token
```
- **Typ:** `string`
- **Wymagane:** ✅ TAK
- **Opis:** Token API z Tailscale z uprawnieniami do odczytu urządzeń
- **Jak uzyskać:**
  1. Przejdź do https://login.tailscale.com/admin/settings/keys
  2. Kliknij "Generate API key"
  3. Wybierz uprawnienia: **Devices: Read only**
  4. Skopiuj wygenerowany token

### 📝 `title`
```yaml
title: "Moje urządzenia Tailscale"
```
- **Typ:** `string`
- **Wymagane:** ❌ Nie
- **Domyślnie:** `"Tailscale"`
- **Opis:** Tytuł widgetu wyświetlany u góry

### 🔗 `title-url`
```yaml
title-url: https://login.tailscale.com/admin/machines
```
- **Typ:** `string`
- **Wymagane:** ❌ Nie
- **Domyślnie:** brak (tytuł nie jest klikalny)
- **Opis:** Link pod tytułem widgetu - przydatny do szybkiego przejścia do panelu Tailscale

### 🌐 `tailnet`
```yaml
tailnet: "example-tailnet.ts.net"
```
- **Typ:** `string`
- **Wymagane:** ❌ Nie
- **Domyślnie:** `"-"` (current tailnet)
- **Opis:** ID tailnet z którego pobierać urządzenia. Wartość `-` oznacza current tailnet powiązany z tokenem.

### 🔌 `url`
```yaml
url: https://api.tailscale.com/api/v2/tailnet/-/devices
```
- **Typ:** `string`
- **Wymagane:** ❌ Nie
- **Domyślnie:** automatycznie generowane na podstawie `tailnet`
- **Opis:** Pełny URL API Tailscale. Użyj tylko jeśli chcesz nadpisać domyślne zachowanie.

### ⏱️ `cache`
```yaml
cache: 10m
```
- **Typ:** `duration`
- **Wymagane:** ❌ Nie
- **Domyślnie:** `10m`
- **Opis:** Jak długo cache'ować dane z API przed ponownym pobraniem
- **Przykłady:**
  - `30s` - 30 sekund
  - `5m` - 5 minut
  - `1h` - 1 godzina
  - `1d` - 1 dzień

### 📦 `collapse-after`
```yaml
collapse-after: 4
```
- **Typ:** `int`
- **Wymagane:** ❌ Nie
- **Domyślnie:** `4`
- **Opis:** Po ilu urządzeniach lista ma być zwinięta (z przyciskiem "Rozwiń")
- **Wartości:**
  - `0` - wyłączone (zawsze pokazuj wszystkie)
  - `> 0` - zwiń po N urządzeniach

### 🟢 `show-online-indicator`
```yaml
show-online-indicator: true
```
- **Typ:** `bool`
- **Wymagane:** ❌ Nie
- **Domyślnie:** `false`
- **Opis:** Czy pokazywać zielony (online) / czerwony (offline) punkt przy nazwie urządzenia
- **Uwaga:** Urządzenie jest uznawane za online jeśli `lastSeen` < 10 sekund temu

---

## 🏷️ Kontrola znaczników (Badges)

Wszystkie znaczniki są **domyślnie wyłączone**. Musisz je włączyć jawnie w konfiguracji.

### 🔵 `show-expiry-disabled`
```yaml
show-expiry-disabled: true
```
- **Typ:** `bool`
- **Domyślnie:** `false`
- **Pokazuje:** Cyjanowy znacznik "Expiry disabled"
- **Kiedy:** Gdy `keyExpiryDisabled: true` w API
- **Znaczenie:** Klucz autoryzacyjny urządzenia nie wygasa automatycznie (nie wymaga re-autoryzacji co 180 dni)

### 🔴 `show-disconnected`
```yaml
show-disconnected: true
```
- **Typ:** `bool`
- **Domyślnie:** `false`
- **Pokazuje:** Czerwony znacznik "Disconnected"
- **Kiedy:** Gdy `connectedToControl: false` w API
- **Znaczenie:** Urządzenie nie jest połączone z panelem kontrolnym Tailscale (wyłączone, brak internetu, lub problem z połączeniem)

### 🟡 `show-blocks-incoming`
```yaml
show-blocks-incoming: true
```
- **Typ:** `bool`
- **Domyślnie:** `false`
- **Pokazuje:** Żółty znacznik "Blocks Incoming"
- **Kiedy:** Gdy `blocksIncomingConnections: true` w API
- **Znaczenie:** Urządzenie blokuje wszystkie przychodzące połączenia (shields-up mode)
- **Jak włączyć:** `tailscale up --shields-up`

### ⚫ `show-joined-date`
```yaml
show-joined-date: true
```
- **Typ:** `bool`
- **Domyślnie:** `false`
- **Pokazuje:** Szary znacznik "Joined [date]"
- **Kiedy:** Zawsze (jeśli API zwraca `created`)
- **Znaczenie:** Data kiedy urządzenie zostało dodane do sieci Tailscale
- **Format:** "Joined Jan 2006" (np. "Joined May 2025")

---

## 📋 Przykładowe konfiguracje

### Minimalna (tylko lista urządzeń)
```yaml
- type: tailscale
  token: twoj_token
```
**Wyświetli:** Tylko podstawowe informacje o urządzeniach bez znaczników.

### Kompaktowa (z online indicator)
```yaml
- type: tailscale
  token: twoj_token
  show-online-indicator: true
```
**Wyświetli:** Podstawowe info + zielony/czerwony punkt przy każdym urządzeniu.

### Podstawowe znaczniki
```yaml
- type: tailscale
  token: twoj_token
  show-expiry-disabled: true
  show-disconnected: true
```
**Wyświetli:** Info o wygasaniu kluczy i statusie połączenia.

### Pełna widoczność (wszystko włączone)
```yaml
- type: tailscale
  title: Tailscale Network
  title-url: https://login.tailscale.com/admin/machines
  token: twoj_token
  cache: 5m
  collapse-after: 6
  show-online-indicator: true
  show-expiry-disabled: true
  show-disconnected: true
  show-blocks-incoming: true
  show-joined-date: true
```
**Wyświetli:** Wszystkie dostępne informacje i znaczniki.

### Monitoring produkcyjny
```yaml
- type: tailscale
  title: Production Devices
  token: twoj_token
  cache: 2m                    # Częstsze odświeżanie
  collapse-after: 10           # Więcej urządzeń przed zwinięciem
  show-online-indicator: true  # Ważny status online
  show-disconnected: true      # Alerty o disconnects
```
**Cel:** Szybkie wykrywanie problemów z połączeniem.

### Audyt bezpieczeństwa
```yaml
- type: tailscale
  title: Security Audit
  token: twoj_token
  show-expiry-disabled: true   # Które klucze nigdy nie wygasają
  show-blocks-incoming: true   # Które mają shields-up
  show-joined-date: true       # Kiedy dodano urządzenia
```
**Cel:** Przegląd ustawień bezpieczeństwa.

---

## Wizualne elementy

Widget zachowuje całą kolorystykę z wersji custom-api:
- **Kolor podstawowy** (`--color-primary`) - nazwa urządzenia, tło IP po hover
- **Kolor pozytywny** (`--color-positive`) - wskaźnik online (jeśli włączony), tło IP po skopiowaniu
- **Kolor negatywny** (`--color-negative`) - wskaźnik offline

### Kolory znaczników (badges) - dane dostępne z API:
- **🔵 Expiry disabled** - Cyjanowy (#17a2b8) - Klucz nie wygasa
- **� Disconnected** - Czerwony (#dc3545) - Nie połączony z kontrolą
- **� Blocks Incoming** - Żółty (#ffc107) - Blokuje połączenia przychodzące
- **⚫ Joined [data]** - Szary (#6c757d) - Data dołączenia do sieci

> **⚠️ Ograniczenia API Tailscale:**  
> Publiczne API Tailscale **NIE udostępnia** informacji o:
> - Exit Node / Advertised Exit Node
> - Subnets / Advertised Routes
> - SSH (enablesSSH)
> - Tags
> - Shared devices
> 
> Te informacje są widoczne tylko w panelu webowym Tailscale, ale nie są eksportowane przez API v2.

### Znaczniki (Badges) - dostępne dane
Pod każdym urządzeniem mogą pojawić się znaczniki oparte na rzeczywistych danych z API:

1. **🔵 Expiry disabled** - Klucz autoryzacyjny urządzenia nie wygasa automatycznie  
   *(Wszystkie Twoje urządzenia mają tę flagę włączoną)*

2. **🔴 Disconnected** - Urządzenie nie jest aktualnie połączone z panelem kontrolnym Tailscale  
   *(Występuje gdy `connectedToControl: false`)*

3. **🟡 Blocks Incoming** - Urządzenie blokuje przychodzące połączenia  
   *(Ustawienie bezpieczeństwa w konfiguracji urządzenia)*

4. **⚫ Joined [data]** - Data dołączenia urządzenia do sieci Tailscale  
   *(Wyświetla czytelny format daty utworzenia urządzenia)*

### Kopiowanie adresu IP
Po najechaniu na wiersz urządzenia:
1. Zamiast informacji o systemie i użytkowniku pojawia się adres IP
2. Adres IP jest klikalny (hover zmienia tło na niebieski)
3. Kliknięcie w IP kopiuje je do schowka
4. Po skopiowaniu tło zmienia się na zielone i pojawia się ✓ na 2 sekundy
5. Działa w każdej przeglądarce dzięki mechanizmowi fallback

---

## ⚠️ Ograniczenia API Tailscale

Publiczne API Tailscale **NIE udostępnia** informacji o:
- Exit Node / Advertised Exit Node
- Subnets / Advertised Routes
- SSH (enablesSSH)
- Tags
- Shared devices

Te informacje są widoczne tylko w panelu webowym Tailscale, ale nie są eksportowane przez API v2.

### ✅ Dostępne z API:
- Lista urządzeń
- Adresy IP
- Status połączenia (`connectedToControl`)
- Expiry status (`keyExpiryDisabled`)
- Blokada połączeń (`blocksIncomingConnections`)
- Data utworzenia (`created`)
- Ostatnia aktywność (`lastSeen`)
- Dostępne aktualizacje (`updateAvailable`)

---

## 🔧 Rozwiązywanie problemów

### Nie widać żadnych urządzeń
1. Sprawdź czy token ma uprawnienia: **Devices: Read only**
2. Sprawdź logi w terminalu: `./glance --config config/glance.yml`
3. Przetestuj token ręcznie:
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://api.tailscale.com/api/v2/tailnet/-/devices
   ```

### Znaczniki się nie pokazują
1. Upewnij się że włączyłeś odpowiednie opcje `show-*: true`
2. Sprawdź czy dane urządzenia faktycznie mają te właściwości (np. `keyExpiryDisabled: true`)
3. Sprawdź cache - może trzeba poczekać na odświeżenie

### Token się przedawnia
Token API Tailscale **nigdy nie wygasa** (w przeciwieństwie do device keys).
Jeśli przestaje działać:
1. Sprawdź czy token został usunięty z panelu Tailscale
2. Wygeneruj nowy token i zaktualizuj config

### Widget jest wolny
1. Zwiększ `cache:` do np. `30m` lub `1h`
2. Zmniejsz częstotliwość odświeżania całej strony Glance

---
