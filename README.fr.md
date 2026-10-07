# Open Niscat

Consultez le catalogue de pièces NISCAT (véhicules utilitaires légers et camions Nissan Motor Ibérica, édition 01/2015) dans un navigateur : identification d'un véhicule par son VIN, navigation dans les vues éclatées, recherche de pièces, suivi des remplacements, et constitution d'une liste de pièces à copier, exporter ou partager.

Un seul binaire sert le catalogue et son interface web : copiez-le à côté d'un dossier `data/` sur un poste d'atelier et ouvrez un navigateur. Les téléphones et tablettes du réseau local peuvent aussi l'utiliser.

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
| `--default-lang` | `en` | langue proposée sur `/` (`en`, `fr`, `es` ou `de`) |
| `--config` | `open-niscat.toml` à côté du binaire | fichier de configuration |

Priorité : options de la ligne de commande, puis fichier, puis valeurs par défaut. Une clé inconnue ou une valeur invalide arrête le programme avec un message.

## Compiler

Prérequis : Go 1.26+, Node 24+, GNU Make, golangci-lint 2.x.

```
make web-install   # une fois, puis après chaque modification de web/package-lock.json
make tools-install # une fois : dépendances Python de tools/data (Pillow, ruff, mypy, pytest)
make lint test     # Go, frontend et outillage des données
make build         # frontend, puis binaires Linux et Windows dans dist/
```

Développement : `make run` (API sur :8080 avec `./data`) et `npm --prefix web run dev` (Vite avec rechargement à chaud sur :5173, relaie `/api` et `/files`).

Les tests utilisent un petit catalogue synthétique ; les tests sur le vrai catalogue ne s'exécutent que si `data/` est présent.

## Licence et données

Le code source est publié sous [licence MIT](LICENSE). Les données du catalogue appartiennent à Nissan : elles ne sont pas distribuées avec ce projet et ne doivent jamais être commitées ni publiées.
