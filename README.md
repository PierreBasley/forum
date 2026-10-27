# Forum

Forum web écrit en Go, réalisé dans le cadre de la formation full-stack developer à Zone01 Normandie. Les utilisateurs peuvent échanger autour de questions de programmation : créer des posts, les commenter et réagir avec des likes et des dislikes.

## Fonctionnalités

**Visiteur (non connecté) :**
- Voir tous les posts et leurs commentaires.
- Voir le nombre de likes et de dislikes.
- Créer un compte et se connecter.

**Utilisateur connecté :**
- Créer un post et choisir une ou plusieurs catégories (SQL, Go, HTML).
- Commenter un post.
- Liker ou disliker un post ou un commentaire.
- Consulter son profil et se déconnecter.

## Prérequis

- Go version 1.26.2
- Git
- Un compilateur C (par exemple gcc, ou les « Xcode Command Line Tools » sur Mac), car le package `sqlite3` en a besoin
- Une connexion internet au premier lancement, pour télécharger les packages

## Lancer le projet

1. Récupérer le code :
   ```
   git clone https://github.com/PierreBasley/forum.git
   ```
2. Entrer dans le dossier du projet :
   ```
   cd forum
   ```
3. Créer le dossier de la base de données (étape nécessaire tant que le projet ne le crée pas automatiquement) :
   ```
   mkdir bdd
   ```
4. Lancer le serveur :
   ```
   go run .
   ```
5. Ouvrir le forum dans le navigateur : http://localhost:8080

Pour arrêter le serveur, appuyer sur `Ctrl + C` dans le terminal.

## Structure du projet

- `main.go` : point d'entrée. Il ouvre la base de données, exécute les migrations, puis démarre le serveur.
- `server/` : serveur HTTP, routes, handlers, gestion des erreurs et migrations.
- `login/` : comptes, mots de passe et cookies.
- `session/` : gestion des sessions des utilisateurs.
- `models/` : structures de données (utilisateur, post, commentaire).
- `migrations/` : fichiers SQL numérotés qui créent les tables de la base.
- `templates/` : pages HTML.
- `static/` : fichiers CSS.

## Base de données

Le projet utilise SQLite. Le fichier `bdd/db.sqlite` est créé au premier lancement, puis les tables sont créées automatiquement par les migrations : `roles`, `users`, `posts`, `comments`, `likes`, `comment_likes` et `password_reset_codes`.

## Sécurité

- Les mots de passe sont chiffrés avec bcrypt et ne sont jamais stockés en clair.
- Le cookie de session ne contient qu'un identifiant aléatoire (UUID). Les informations de la session restent côté serveur, et la session a une durée limitée.
- Le cookie est `HttpOnly` : le JavaScript du navigateur ne peut pas le lire.

## Packages utilisés

En plus de la bibliothèque standard de Go :

- `github.com/mattn/go-sqlite3` : accès à la base SQLite
- `golang.org/x/crypto/bcrypt` : chiffrement des mots de passe
- `github.com/gofrs/uuid` : génération d'identifiants uniques

## Gestion des erreurs

Le serveur renvoie une page d'erreur adaptée pour les codes HTTP 400, 401, 403, 404, 405 et 500.

## Auteurs

- Abdou Cherif
- Terapty-returns
- PierreBasley
- Hicham Hehe
