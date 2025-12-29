# KodiScript Go SDK

Un interpréteur KodiScript v1.2 pour Go, conçu comme module à intégrer dans vos projets.

## 🎯 Pourquoi KodiScript ?

Vous avez déjà eu besoin d'exécuter du code dynamiquement dans votre application ? De laisser vos utilisateurs définir des règles métier sans recompiler tout le projet ? C'est exactement pour ça que KodiScript existe.

**KodiScript** est un langage de script minimaliste, pensé pour être :

- **Simple à apprendre** — Une syntaxe épurée, proche du JavaScript, que n'importe qui peut comprendre en quelques minutes
- **Léger** — Pas de dépendances lourdes, juste l'essentiel pour faire le travail
- **Sécurisé** — Exécution sandboxée, vos utilisateurs peuvent écrire des scripts sans risquer de casser votre système
- **Facile à intégrer** — Quelques lignes de code suffisent pour l'embarquer dans votre projet Go

Imaginez : un admin qui configure des règles de validation, un workflow qui s'adapte selon le contexte, ou des transformations de données à la volée. Tout ça devient possible sans toucher à votre code source.

## Installation

```bash
go get github.com/kodi-script/kodi-go
```

## Utilisation Rapide

```go
package main

import (
    "fmt"
    kodi "github.com/kodi-script/kodi-go"
)

func main() {
    // Exécution simple
    result := kodi.Run(`
        let name = "Kodi"
        let version = 1.2
        print("Hello " + name)
    `, nil)

    for _, line := range result.Output {
        fmt.Println(line)
    }
}
```

## Injection de Variables

```go
vars := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
        "role": "admin",
    },
    "config": map[string]interface{}{
        "debug": true,
    },
}

result := kodi.Run(`
    let greeting = "Hello " + user.name
    let status = user?.active ?: "offline"
    print(greeting)
`, vars)
```

## Fonctions Natives

| Fonction | Description |
|----------|-------------|
| `print(...)` | Affiche des valeurs |
| `toString(val)` | Convertit en string |
| `toNumber(val)` | Convertit en nombre |
| `base64Encode(str)` | Encode en Base64 |
| `base64Decode(str)` | Décode du Base64 |
| `urlEncode(str)` | Encode pour URL |
| `urlDecode(str)` | Décode une URL |
| `jsonParse(str)` | Parse du JSON |
| `jsonStringify(val)` | Sérialise en JSON |
| `typeOf(val)` | Retourne le type |
| `isNull(val)` | Vérifie si null |

## Fonctions Personnalisées

```go
script := kodi.New(`
    let result = myCustomFunc("hello")
`)

script.RegisterFunction("myCustomFunc", func(args ...interface{}) (interface{}, error) {
    return strings.ToUpper(args[0].(string)), nil
})

result := script.Execute()
```

## Syntaxe KodiScript v1.2

```javascript
// Variables
let name = "Kodi"
let version = 1.2

// Null-safety
let status = user?.active ?: "offline"

// Conditions
if (version > 1.0) {
    print("Modern version")
} else {
    print("Legacy version")
}

// Expressions multi-lignes (continuation automatique)
let total = 10 +
            20 +
            30

// Point-virgule optionnel
let x = 1
let y = 2;  // Les deux sont valides
```

## Tests

```bash
go test ./... -v
```
