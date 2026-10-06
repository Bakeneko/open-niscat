# Open Niscat

Consultez le catalogue de pièces NISCAT (véhicules utilitaires légers et camions Nissan Motor Ibérica, édition 01/2015) dans un navigateur : identification d'un véhicule par son VIN, navigation dans les vues éclatées, recherche de pièces, suivi des remplacements, et constitution d'une liste de pièces à copier, exporter ou partager.

Les données du catalogue ne sont **pas** fournies : elles sont sous licence Nissan. Il vous faut votre propre dossier `data/` (`data.db`, `manifest.json`, `img/`, `gindex/`, `cinfo/`).

*English version: [README.md](README.md).*

## Lancer

1. Placez le binaire et le dossier `data/` côte à côte :
   ```
   open-niscat-windows-amd64.exe   (ou open-niscat-linux-amd64)
   data/
   ```
2. Lancez le binaire (double-clic sous Windows). Le navigateur s'ouvre sur http://127.0.0.1:8080/.

Options (également réglables dans `open-niscat.toml` à côté du binaire — voir `open-niscat.example.toml`) :

| Option | Valeur par défaut | Rôle |
|---|---|---|
| `--data` | `./data` à côté du binaire | dossier des données |
| `--addr` | `127.0.0.1:8080` | adresse d'écoute ; `0.0.0.0:8080` pour autoriser les autres machines du réseau local |
| `--open-browser` | `true` | ouvrir le navigateur par défaut au démarrage |
| `--default-lang` | `en` | langue proposée sur `/` (`en` ou `fr`) |
| `--config` | `open-niscat.toml` à côté du binaire | fichier de configuration |

Priorité : options de la ligne de commande, puis fichier, puis valeurs par défaut. Une clé inconnue ou une valeur invalide arrête le programme avec un message.

## Compiler

Prérequis : Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # une fois, puis après chaque modification de web/package-lock.json
make lint test     # Go + frontend
make build         # frontend, puis binaires Linux et Windows dans dist/
```

Développement : `make run` (API sur :8080 avec `./data`) et `npm --prefix web run dev` (Vite avec rechargement à chaud, relaie `/api` et `/files`).

## Licence et données

Le code source est libre d'utilisation. Les données du catalogue appartiennent à Nissan et ne doivent jamais être commitées ni publiées.
