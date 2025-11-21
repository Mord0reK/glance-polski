# Widget Cloudflare

Widget Cloudflare pozwala na wyświetlanie ilości zapytań do stron zabezpieczonych przez Cloudflare wraz z ilością uniklanych użytkowników.

## Konfiguracja

Aby skonfigurować widget Cloudflare, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: cloudflare
  api-key: KLUCZ            # Jest to klucz pobrany z sekcji "Get your API token" po wejsciu na domenę na Cloudflare (wystarczy template "Read analytics and Logs")
  zone-id:                  # Jest to ID Strefy pobrany z sekcji "Zone ID" nad sekcja pobierania klucza API
  time-range: "X"           # Domyślna rama czasowa pobieranych danych, gdzie X to 24h, 7d lub 30d
```