# Piano: Integrazione GitHub con token

## Context

Jig supporta attualmente solo GitLab. L'obiettivo è aggiungere il supporto a **GitHub** come provider alternativo. Il token di autenticazione (generato da una GitHub App lato CI/CD GitHub Actions) è già disponibile come variabile d'ambiente `GITHUB_TOKEN` — Jig lo legge semplicemente come Bearer token, esattamente come già fa con `gitToken` per GitLab.

Nessuna libreria di autenticazione GitHub App è necessaria lato Jig: la generazione del token è responsabilità del workflow GitHub Actions.

---

## Step 0: Aggiungere dipendenza SDK GitHub

```bash
go get github.com/google/go-github/v62
```

---

## Step 1: Nuovo client `internal/repo/clients/github.go`

Struct `GitHub` che implementa `entities.RepoClient` e `entities.IssuesTracker`:

```go
type GitHub struct {
    c *github.Client
}

func NewGitHub(token string) (GitHub, error) {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    httpClient := oauth2.NewClient(context.Background(), ts)
    return GitHub{c: github.NewClient(httpClient)}, nil
}
```

`oauth2` è già presente in `go.mod` come dipendenza indiretta.

Il `gitRepoID` nel file di modello sarà nel formato `"owner/repo"` (es. `"happyagosmith/jig"`), splittato su `/` internamente.

Metodi da implementare:

| Metodo | API GitHub |
|---|---|
| `GetCommits` | `Repositories.CompareCommits(owner, repo, base, head)` |
| `GetMergeRequests` | `PullRequests.List(state=closed)` + filtro per SHA commit |
| `GetRepoURL` | `Repositories.Get` → `.GetHTMLURL()` |
| `GetReleaseURL` | `Repositories.GetReleaseByTag` → `.GetHTMLURL()` |
| `GetIssues` | `Issues.Get` per ogni numero |
| `GetKnownIssues` | stub (non implementato, come GitLab) |

---

## Step 2: Nuovi campi di configurazione in `cmd/configuration.go`

Aggiungere solo una costante per il provider:

```go
const (
    GitProvider = "gitProvider" // "gitlab" | "github" (default: "gitlab")
)
```

Il token viene letto dal campo `gitToken` già esistente (valorizzabile nel `.jig.yaml` per test locali o tramite variabile d'ambiente in CI).

---

## Step 3: Factory in `cmd/configuration.go`

Modificare `ConfigureRepoTracker()`:

```go
func ConfigureRepoTracker() (entities.RepoTracker, error) {
    switch viper.GetString(GitProvider) {
    case "github":
        token := viper.GetString(GitToken)
        if token == "" {
            token = os.Getenv("GITHUB_TOKEN")
        }
        return clients.NewGitHub(token)
    default: // "gitlab" — retrocompatibile
        return clients.NewGitLab(
            viper.GetString(GitURL),
            viper.GetString(GitToken),
        )
    }
}
```

---

## Step 4: Aggiornare `examples/.jig.yaml`

```yaml
# Per GitHub:
# gitProvider: "github"
# gitToken: "ghs_xxxxxxxxxxxx"   # token GitHub App (per test locali)
#                                 # in CI/CD usa la variabile d'ambiente GITHUB_TOKEN
#
# Nel model.yaml, usare gitRepoID nel formato "owner/repo":
# gitRepoID: "happyagosmith/jig"
```

---

## File da modificare

| File | Tipo modifica |
|---|---|
| `go.mod` / `go.sum` | Aggiornati da `go get` |
| `internal/repo/clients/github.go` | Nuovo file |
| `cmd/configuration.go` | Nuova costante `GitProvider` + switch nel factory |
| `examples/.jig.yaml` | Esempio config GitHub |

---

## Verifica

1. `go build ./...` — nessun errore di compilazione
2. Test unitario `internal/repo/clients/github_test.go` (pattern uguale a `gitlab_test.go`)
3. Test manuale con `gitProvider: github` e `gitToken` settato nel `.jig.yaml` o via `GITHUB_TOKEN`
