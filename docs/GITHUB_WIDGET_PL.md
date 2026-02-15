# Widget GitHub - Instrukcja użycia

Widget GitHub pozwala na wyświetlanie listy repozytoriów GitHub, do których masz dostęp - zarówno publicznych, jak i prywatnych.

## Konfiguracja

Aby skonfigurować widget GitHub, dodaj następującą konfigurację do swojego pliku `glance.yml`:

```yaml
- type: github
  token: ghp_xxxxx  # Opcjonalny token PAT dla repozytoriów prywatnych
  collapse-after: 5 # Liczba repozytoriów wyświetlanych domyślnie (domyślnie 5)
  sort: updated    # Sortowanie: updated, created, pushed, full_name (domyślnie updated)
  title-link: https://github.com/twoj-profil # Opcjonalny link przekierowania po kliknięciu w tytuł
```

### Uzyskiwanie tokenu PAT (Personal Access Token)

Jeśli chcesz wyświetlać również prywatne repozytoria, musisz utworzyć token PAT:

1. Zaloguj się na swoje konto GitHub
2. Przejdź do **Settings** -> **Developer settings** -> **Personal access tokens** -> **Tokens (classic)**
3. Kliknij **Generate new token (classic)**
4. Nadaj nazwę tokenowi i zaznacz uprawnienia:
   - `repo` - dla pełnego dostępu do repozytoriów (wymagane dla prywatnych)
5. Wygeneruj token i skopiuj go do konfiguracji widgetu

**Uwaga:** Bez tokena widget wyświetli tylko publiczne repozytoria. Token możesz również wykorzystać do zwiększenia limitu zapytań do API GitHub (60/h bez tokena, 5000/h z tokenem).

## Wyświetlane informacje

Widget wyświetla dla każdego repozytorium:

- **Nazwa repozytorium** - pełna nazwa (np. `username/nazwa-repozytorium`) z linkiem do GitHub
- **Opis** - opis repozytorium (jeśli dostępny)
- **Data ostatniej aktywności** - czas od ostatniej aktualizacji (format względny: "2 dni temu")
- **Gwiazdki** - liczba gwiazdek (★)
- **Język** - główny język programowania (jeśli wykryty)

## Parametry konfiguracji

| Parametr | Typ | Opis | Domyślnie |
|----------|-----|------|-----------|
| `token` | string | Personal Access Token (opcjonalny) | - |
| `collapse-after` | int | Liczba repozytoriów wyświetlanych domyślnie (reszta ukryta pod "Pokaż więcej") | 5 |
| `sort` | string | Sortowanie: `updated`, `created`, `pushed`, `full_name` | `updated` |
| `title-link` | string | Link przekierowania po kliknięciu w tytuł widgetu | - |

## Przykładowa konfiguracja

### Podstawowa konfiguracja (tylko publiczne repozytoria)

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: github
            collapse-after: 10
```

### Konfiguracja z tokenem (publiczne i prywatne)

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: github
            token: ghp_xxxxxxxxxxxxxxxxxxxx
            collapse-after: 5
            sort: updated
```

### Sortowanie według daty utworzenia

```yaml
- type: github
  token: ghp_xxxxx
  collapse-after: 10
  sort: created
```

### Przekierowanie tytułu na stronę profilu GitHub

```yaml
- type: github
  token: ghp_xxxxx
  collapse-after: 5
  sort: updated
  title-link: https://github.com/twoj-profil
```

## Rozwiązywanie problemów

### Widget wyświetla błąd

1. Sprawdź czy token PAT jest poprawny i nie wygasł
2. Sprawdź czy token ma uprawnienia `repo`
3. Upewnij się, że masz dostęp do co najmniej jednego repozytorium

### Wyświetlane są tylko publiczne repozytoria mimo podania tokena

1. Sprawdź czy token ma uprawnienia `repo`
2. Token musi mieć uprawnienie `repo` dla dostępu do prywatnych repozytoriów

### Rate limit (zbyt wiele zapytań)

Jeśli widget nie może pobrać danych z powodu limitu API:
1. Dodaj ważny token PAT - zwiększy to limit do 5000 zapytań na godzinę
2. Zwiększ czas cache'owania (używając parametru `cache` w konfiguracji widgetu)
