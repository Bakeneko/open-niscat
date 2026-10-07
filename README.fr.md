# Open Niscat

Consultez le catalogue de pièces NISCAT (véhicules utilitaires légers et camions Nissan Motor Ibérica, édition 01/2015) dans un navigateur : identification d'un véhicule par son VIN, navigation dans les vues éclatées, recherche de pièces, suivi des remplacements, et constitution d'une liste de pièces à copier, exporter ou partager.

Un seul binaire sert le catalogue et son interface web : copiez-le à côté d'un dossier `data/` sur un poste d'atelier et ouvrez un navigateur, ou lancez-le avec Docker sur un serveur. Les téléphones et tablettes du réseau local peuvent aussi l'utiliser.

Les données du catalogue ne sont **pas** fournies : elles sont sous licence Nissan. Il vous faut votre propre dossier `data/` (voir [Données](#données)).

*English version: [README.md](README.md).*

## Fonctionnalités

- **Identification du véhicule** : par VIN complet ou par ses 6 derniers caractères (ou plus), ou par catalogue et code modèle (filtres sur le moteur, la carrosserie, la boîte…). La fiche véhicule présente ses caractéristiques, ses documents et l'index général.
- **Vues éclatées** : index de groupe et planches de pièces avec zoom (molette, pincement), déplacement, repères cliquables et infobulles rapides ; les numéros des repères sont réécrits sur les dessins comme dans NISCAT.
- **Applicabilité** : les sections et pièces qui ne concernent pas le véhicule sont masquées ou grisées ; les pièces hors de la date de production du véhicule sont signalées, jamais masquées.
- **Recherche** : VIN, sections et pièces (référence ou texte), limitée au véhicule actif ou sur tous les véhicules.
- **Remplacements** : références alternatives et dernière référence connue, et toutes les planches où une pièce est utilisée.
- **Liste de pièces** : conservée dans le navigateur, quantités, copie pour un tableur, export CSV, impression / PDF, et lien de partage qui ouvre la même liste ailleurs.
- **URL partageables** : chaque page (véhicule, planche, repère sélectionné, recherche, liste) a son propre lien.
- Interface en anglais, français, espagnol et allemand (les langues des données NISCAT) ; pensée pour l'ordinateur comme pour le mobile ; les planches s'impriment dessin puis tableau des pièces.

## Données

Le dossier `data/` contient le catalogue converti depuis une installation de NISCAT (édition 01/2015) :

```
data/
├── manifest.json   schéma, version des données, édition source, informations de construction
├── data.db         base SQLite (catalogues, VIN, codes modèle, sections, pièces, repères)
├── img/<série>/    planches de pièces (PNG)
├── gindex/<série>/ dessins des index de groupe (PNG)
└── cinfo/<série>/  documents des catalogues (PDF)
```

Le programme ouvre `data.db` en lecture seule et refuse un dossier dont le schéma indiqué dans `manifest.json` n'est pas pris en charge.

### Construire le dossier de données

Si vous possédez le programme d'installation de NISCAT 01/2015, construisez le dossier vous-même (Windows, Python 3.12+ avec Pillow) :

```
python tools/data/build_data.py --installer <dossier d'installation> --out data
```

Voir [tools/data/README.md](tools/data/README.md) (en anglais) pour les prérequis, les options et les autres systèmes.

## Lancer

Téléchargez le binaire de votre système sur la [page des releases](https://github.com/Bakeneko/open-niscat/releases) : `windows-amd64`, `linux-amd64`, `linux-arm64` ou `darwin-arm64` (Mac Apple Silicon). `SHA256SUMS` permet de vérifier le téléchargement (`sha256sum -c SHA256SUMS --ignore-missing`). Vous pouvez aussi le compiler vous-même (voir [Compiler](#compiler)).

1. Placez le binaire et le dossier `data/` côte à côte :
   ```
   open-niscat-windows-amd64.exe   (ou open-niscat-linux-amd64, …)
   data/
   ```
2. Lancez le binaire (double-clic sous Windows). Le navigateur s'ouvre sur http://127.0.0.1:8080/.
   - Linux et macOS : rendez-le d'abord exécutable (`chmod +x open-niscat-*`).
   - macOS bloque le binaire non signé au premier lancement : clic droit → Ouvrir, ou `xattr -d com.apple.quarantine open-niscat-darwin-arm64`.

Options (également réglables dans `open-niscat.toml` à côté du binaire — voir `open-niscat.example.toml` — ou par variables d'environnement) :

| Option | Variable d'environnement | Valeur par défaut | Rôle |
|---|---|---|---|
| `--data` | `OPEN_NISCAT_DATA` | `./data` à côté du binaire | dossier des données |
| `--addr` | `OPEN_NISCAT_ADDR` | `127.0.0.1:8080` | adresse d'écoute ; `0.0.0.0:8080` pour autoriser les autres machines du réseau local |
| `--open-browser` | `OPEN_NISCAT_OPEN_BROWSER` | `true` | ouvrir le navigateur par défaut au démarrage |
| `--default-lang` | `OPEN_NISCAT_DEFAULT_LANG` | `en` | langue proposée sur `/` (`en`, `fr`, `es` ou `de`) |
| `--config` | | `open-niscat.toml` à côté du binaire | fichier de configuration |

Priorité : options de la ligne de commande, puis variables d'environnement, puis fichier, puis valeurs par défaut. Un chemin relatif donné par une option ou une variable part du dossier courant, celui du fichier part du dossier du fichier. Une clé inconnue ou une valeur invalide arrête le programme avec un message.

### Docker

L'image `ghcr.io/bakeneko/open-niscat` (amd64 et arm64) ne contient que le programme : montez votre dossier `data/` en lecture seule sur `/data`. Copiez [compose.example.yaml](compose.example.yaml) en `compose.yaml` à côté du dossier `data/`, puis :

```
docker compose up -d
```

ou sans compose :

```
docker run -d --name open-niscat --restart unless-stopped -p 8080:8080 -v ./data:/data:ro ghcr.io/bakeneko/open-niscat
```

- Les réglages sont des variables d'environnement (`-e OPEN_NISCAT_DEFAULT_LANG=fr`, ou `environment:` dans `compose.yaml`) ; l'image fixe déjà `OPEN_NISCAT_DATA=/data`, `OPEN_NISCAT_ADDR=0.0.0.0:8080` et `OPEN_NISCAT_OPEN_BROWSER=false`. Gardez le port 8080 dans le conteneur et changez plutôt le port publié (`-p 80:8080`).
- Sous Linux, les fichiers de données doivent être lisibles par l'utilisateur du conteneur (uid 10001), par exemple `chmod -R a+rX data`.
- `GET /health` répond 200 tant que le catalogue est lisible ; le contrôle de santé de l'image l'utilise (`docker ps` affiche `healthy`).

## Compiler

Prérequis : Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # une fois, puis après chaque modification de web/package-lock.json
make tools-install # une fois : dépendances Python de tools/data (Pillow, ruff, mypy, pytest)
make lint test     # Go, frontend et outillage des données
make vuln          # vulnérabilités connues des dépendances Go (govulncheck)
make build         # frontend, puis les quatre binaires de release dans dist/
```

Développement : `make run` (API sur :8080 avec `./data`) et `npm --prefix web run dev` (Vite avec rechargement à chaud sur :5173, relaie `/api` et `/files`).

Les tests utilisent un petit catalogue synthétique ; les tests sur le vrai catalogue ne s'exécutent que si `data/` est présent.

`make docker` construit l'image pour votre machine (`open-niscat:<version>`).

### Publier une version

Créez un tag de version et poussez-le : `git tag v1.2.0 && git push origin v1.2.0`. Le workflow de release lance les vérifications, attache les quatre binaires et `SHA256SUMS` à une release GitHub et publie l'image en `1.2.0`, `1.2` et `latest`. Un suffixe (`v1.2.0-rc.1`, `v1.2.0-beta.1`) crée une pré-version, publiée sous son seul tag.

Une release peut aussi être publiée depuis l'interface web de GitHub avec un nouveau tag : le workflow y attache alors les fichiers et garde ses notes. Les tags de l'image suivent toujours le nom du tag, pas la case « pre-release ».

## Licence et données

Le code source est publié sous [licence MIT](LICENSE). Les données du catalogue appartiennent à Nissan : elles ne sont pas distribuées avec ce projet et ne doivent jamais être commitées ni publiées.
