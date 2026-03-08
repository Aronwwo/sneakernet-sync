# Plan Projektu: sneakernet-sync

**Projekt inżynierski** | Zespół: @Aronwwo, @PawelMierzwa | Rok akademicki 2024/2025

---

## 1. Opis problemu

W środowiskach bez dostępu do sieci (sieci korporacyjne o podwyższonym poziomie bezpieczeństwa, komputery bez połączenia z internetem, lokalizacje z ograniczoną infrastrukturą) standardowe narzędzia do synchronizacji plików oparte na chmurze są niedostępne. Użytkownicy przenoszą dane ręcznie przy użyciu zewnętrznych nośników (pendrive USB, dyski przenośne), co jest podatne na błędy i nie zapewnia śledzenia wersji ani wykrywania konfliktów.

---

## 2. Cel projektu

Celem projektu jest stworzenie narzędzia CLI umożliwiającego **dwukierunkową synchronizację plików** między komputerami przy użyciu zewnętrznego nośnika pamięci (USB) jako medium transportowego — bez udziału sieci i chmury. Narzędzie śledzi zmiany, przechowuje metadane lokalnie w bazie SQLite i wykrywa konflikty wynikające z równoczesnych modyfikacji na różnych maszynach.

---

## 3. Zakres obowiązkowy (MVP)

- Wykrywanie zmian plików na podstawie skrótu SHA-256 (nowe, zmodyfikowane, usunięte)
- Przechowywanie metadanych w lokalnej bazie SQLite (bez CGo, sterownik ncruces/go-sqlite3)
- Transport zmian przez zewnętrzny nośnik (operacja push/pull)
- Wykrywanie konfliktów metodą trójdrożnego porównania (local / remote / archive)
- Interfejs wiersza poleceń (CLI) z podkomendami: `init`, `scan`, `status`, `push`, `pull`, `sync`, `resolve`, `doctor`
- Testy jednostkowe i integracyjne z pokryciem ≥ 80%
- Potok CI/CD (GitHub Actions) na 3 systemach operacyjnych: Linux, Windows, macOS

---

## 4. Zakres rozszerzony

- Opcjonalny transport sieciowy (TCP/IP w sieci lokalnej) jako alternatywa dla USB
- Szyfrowanie zawartości nośnika (AES-256-GCM)
- Graficzny interfejs użytkownika (GUI) zbudowany przy użyciu [Wails](https://wails.io/) (Go + WebView)
- Obsługa wielu urządzeń i centralna historia synchronizacji
- Kompresja przesyłanych danych (zstd)

---

## 5. Stack technologiczny

| Komponent          | Technologia                                              |
|--------------------|----------------------------------------------------------|
| Język              | Go 1.22+                                                 |
| Framework CLI      | github.com/spf13/cobra                                   |
| Baza danych        | SQLite via github.com/ncruces/go-sqlite3 (bez CGo)       |
| Haszowanie         | crypto/sha256 (biblioteka standardowa)                   |
| Logowanie          | log/slog (biblioteka standardowa)                        |
| Konfiguracja       | TOML                                                     |
| Testowanie         | testing (stdlib) + github.com/stretchr/testify           |
| Linting            | golangci-lint                                            |
| CI/CD              | GitHub Actions (matrix: Linux / Windows / macOS)         |

---

## 6. Model realizacji

Projekt realizowany jest zgodnie z modelem **iteracyjno-przyrostowym** z wybranymi praktykami Scrum:

- Sprinty dwutygodniowe
- Backlog produktu zarządzany przez GitHub Issues
- Przegląd sprintu i retrospektywa na koniec każdej iteracji
- Definicja Ukończenia (DoD) obowiązuje dla każdego przyrostu
- Kontrola wersji oparta na modelu feature-branch + Pull Request + review

---

## 7. Organizacja pracy zespołu

Zespół dwuosobowy: @Aronwwo i @PawelMierzwa. Obaj członkowie zespołu ponoszą **wspólną odpowiedzialność** za całość kodu. Podział pracy jest elastyczny i dostosowywany do bieżących potrzeb sprintu.

- Komunikacja: bezpośrednia + komentarze w GitHub PR
- Przeglądy kodu: wzajemne (każdy PR wymaga zatwierdzenia przez drugiego członka)
- Zadania śledzone w GitHub Issues i GitHub Projects

---

## 8. Zasady pracy z repozytorium

- Gałąź `main` jest chroniona — zmiany tylko przez Pull Request
- Nazewnictwo gałęzi: `feature/`, `fix/`, `test/`, `docs/`, `refactor/`, `ci/`
- Konwencja commitów: Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `refactor:`, `ci:`)
- Każdy PR wymaga: zaliczenia CI, zatwierdzenia przez recenzenta, braku otwartych komentarzy
- Scalanie: squash merge do `main`
- Tagowanie sprintów: `sprint-N` po zakończeniu każdej iteracji

Szczegóły: [docs/GIT_WORKFLOW.md](GIT_WORKFLOW.md)

---

## 9. Testy i CI

- **Testy jednostkowe**: każdy pakiet posiada testy w pliku `*_test.go`
- **Testy integracyjne**: testy otwierające rzeczywistą bazę SQLite i skanujące tymczasowe katalogi
- **Pokrycie kodu**: cel ≥ 80% dla kodu niebędącego placeholderem
- **Linting**: golangci-lint z włączonymi linterami: govet, staticcheck, errcheck, gofmt, goimports, misspell, revive
- **CI**: GitHub Actions — macierz: {ubuntu-latest, windows-latest, macos-latest} × {Go 1.22, Go 1.23}
- **Artefakty**: raport pokrycia kodu publikowany jako artefakt CI przy każdym uruchomieniu na ubuntu/Go 1.23

---

## 10. Definition of Done

Każde zadanie i Pull Request muszą spełniać wszystkie kryteria DoD przed scaleniem z `main`. Pełna lista: [docs/DEFINITION_OF_DONE.md](DEFINITION_OF_DONE.md).

Kluczowe punkty:
- Kod kompiluje się i testy przechodzą
- Pokrycie ≥ 80% dla nowego kodu
- CI zielone na wszystkich platformach
- Przegląd kodu zatwierdzony przez drugiego członka zespołu
- Dokumentacja zaktualizowana

---

## 11. Harmonogram

| Sprint   | Czas trwania | Cele                                                                          |
|----------|--------------|-------------------------------------------------------------------------------|
| Sprint 0 | Tydzień 1–2  | Szkielet projektu, go.mod, struktura katalogów, CI, README, dokumentacja      |
| Sprint 1 | Tydzień 3–4  | Implementacja skanera (`scan`), haszowania, bazy SQLite, `init`               |
| Sprint 2 | Tydzień 5–6  | Transport push/pull przez nośnik USB, format archiwum                        |
| Sprint 3 | Tydzień 7–8  | Wykrywanie i rozwiązywanie konfliktów, trójdrożne porównanie                  |
| Sprint 4 | Tydzień 9–10 | Testy integracyjne end-to-end, `doctor`, stabilizacja                        |
| Sprint 5 | Tydzień 11–12| Opcjonalnie: GUI (Wails) lub szyfrowanie lub transport sieciowy               |
| Sprint 6 | Tydzień 13–14| Dokumentacja końcowa, przygotowanie do obrony, prezentacja                   |
| Sprint 7+| Według potrzeb| Poprawki po recenzji promotora, dodatkowe funkcje z zakresu rozszerzonego    |

---

## 12. Kryteria sukcesu

Projekt zostanie uznany za zakończony sukcesem, gdy:

1. Narzędzie umożliwia pełny cykl synchronizacji (`push` → przeniesienie nośnika → `pull`) między dwoma komputerami bez dostępu do sieci.
2. Konflikty (równoczesne modyfikacje tego samego pliku) są wykrywane i raportowane użytkownikowi.
3. Wszystkie testy przechodzą na systemach Linux, Windows i macOS.
4. Pokrycie kodu wynosi co najmniej 80% dla zaimplementowanych funkcji.
5. Kod jest utrzymany w standardzie jakości potwierdzonym przez linter (`golangci-lint`).
6. Projekt posiada kompletną dokumentację techniczną i instrukcję użytkownika.
