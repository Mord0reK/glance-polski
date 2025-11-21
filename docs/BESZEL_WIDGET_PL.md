# Widget Beszel - Instrukcja użycia

Widget Beszel pozwala na monitorowanie statusu serwerów w czasie rzeczywistym, wykorzystując lekkie i nowoczesne narzędzie monitoringu Beszel.

## Konfiguracja

Aby skonfigurować widget Beszel, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: beszel
  url: https://twoja-instancja-beszel.pl    # URL do Twojej instancji Beszel (API)
  redirect-url: https://twoja-instancja-beszel.pl # URL do interfejsu webowego Beszel 
  token: twoj-token-jwt                     # Token JWT (
  cache: 10s                                # Częstotliwość odświeżania (domyślnie 10s)
```

### Uzyskiwanie tokenu

Jeśli Twoja instancja Beszel jest zabezpieczona i nie udostępnia danych publicznie, będziesz potrzebować tokenu.
Obecnie widget obsługuje tokeny Bearer. Możesz uzyskać token logując się do Beszel i sprawdzając żądania sieciowe w przeglądarce lub generując go w panelu administracyjnym (jeśli dostępne).

## Funkcje widgetu

### 1. Monitorowanie statusu serwerów

Widget wyświetla listę serwerów z następującymi informacjami:
- **Nazwa serwera**: Nazwa zdefiniowana w Beszel.
- **Status**: Ikona serwera zmienia kolor w zależności od dostępności (zielony - online, czerwony - offline).
- **Uptime**: Czas nieprzerwanej pracy serwera (np. "5 days uptime").

### 2. Szczegółowe metryki

Dla każdego serwera wyświetlane są paski postępu z aktualnym użyciem zasobów:
- **CPU**: Aktualne użycie procesora w procentach.
- **RAM**: Aktualne użycie pamięci RAM w procentach.
- **DISK**: Zajętość głównego dysku w procentach.

### 3. Dodatkowe informacje (Popover)

Po najechaniu kursorem na ikonę serwera lub paski postępu, wyświetlane są dodatkowe informacje w dymku:
- **Host/IP**: Adres IP lub nazwa hosta serwera.
- **Kernel**: Wersja jądra systemu.
- **CPU Model**: Model procesora.
- **Load Average**: Średnie obciążenie systemu (1m, 5m, 15m) - dostępne po najechaniu na pasek CPU.

### 4. Linkowanie do systemu

Jeśli skonfigurowano parametr `redirect-url`, kliknięcie w nazwę serwera otworzy nową kartę z szczegółowymi statystykami tego konkretnego systemu w panelu Beszel.

## Przykładowa konfiguracja

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: beszel
            title: Serwery
            url: https://beszel.example.com
            redirect-url: https://beszel.example.com
            token: TWóJ_TOKEN
            cache: 5s
```

## Rozwiązywanie problemów

### Widget nie wyświetla danych

1. Sprawdź czy URL do instancji Beszel jest poprawny i dostępny z serwera, na którym działa Glance.
2. Upewnij się, że endpoint `/api/collections/systems/records` jest dostępny.
3. Jeśli Twoja instancja wymaga autoryzacji, upewnij się, że podałeś poprawny token.

### Brakujące metryki

Niektóre metryki (np. Load Average) mogą nie być dostępne w zależności od wersji agenta Beszel zainstalowanego na monitorowanym serwerze.
