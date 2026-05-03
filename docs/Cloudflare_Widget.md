# Widget Cloudflare

Widget Cloudflare pozwala na wyświetlanie ilości zapytań do stron zabezpieczonych przez Cloudflare wraz z ilością zablokowanych użytkowników.

## Konfiguracja

Aby skonfigurować widget Cloudflare, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: cloudflare
  api-key: KLUCZ            # Jest to klucz pobrany z sekcji "Get your API token" po wejściu na domenę na Cloudflare (wystarczy template "Read analytics and Logs")
  zone-id: ID_STREFY       # Jest to ID Strefy pobrany z sekcji "Zone ID" nad sekcją pobierania klucza API
  time-range: "24h"      # Rama czasowa pobieranych danych: 24h, 7d lub 30d
  title-url: "https://dash.cloudflare.com/{account_id}/{domena}/security/analytics"  # Opcjonalnie - link po kliknięciu w tytuł
```

### title-url

Opcjonalne pole `title-url` pozwala ustawić link, który otworzy się po kliknięciu w tytuł widgetu. Przykładowe wartości:

- `https://dash.cloudflare.com/?to=/{account_id}/{domena}/security/analytics` - bezpośrednio do statystyk bezpieczeństwa
- `https://dash.cloudflare.com/?to=/{account_id}/{domena}/analytics/traffic` - do statystyk ruchu