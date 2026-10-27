# Instructions pour les agents

Ce fichier s'applique à l'ensemble du dépôt. Il décrit les contraintes du projet,
les points d'entrée du code et les vérifications à adapter à chaque changement.

## Règles de travail

- Ne jamais créer de commit ni pousser de code, de tag ou d'image sans
  approbation explicite de l'utilisateur. Une demande de modification n'est pas
  une autorisation de commit ou de push.
- Les commandes de déploiement et de publication peuvent également modifier des
  releases GitHub et leurs fichiers. Ne pas les utiliser comme simple validation.
- Lire le code concerné avant de modifier son comportement. Préserver les
  changements existants de l'utilisateur et limiter le diff au travail demandé.
- Respecter le style en place. Éviter les refontes, dépendances et abstractions
  supplémentaires sans besoin concret.
- Distinguer dans le compte rendu ce qui a été vérifié, ce qui reste une exigence
  et ce que l'environnement n'a pas permis de tester.

## Destination et contraintes du produit

`container` est un superviseur et un init léger destiné exclusivement à
**l'intérieur d'un conteneur Linux**. Ce n'est pas un gestionnaire de services de
l'hôte. Il accompagne les images `osixia/baseimage`, en mode mono-processus ou
multi-processus, sur Debian, Ubuntu et Alpine, en variantes root et nonroot.

- Conserver la compatibilité avec gVisor / `runsc`. Ne pas rendre obligatoires
  des privilèges, capacités ou interfaces noyau spécifiques à un hôte classique.
- Le superviseur tourne généralement en PID 1, mais les sous-commandes et les
  exécutions imbriquées fonctionnent aussi avec un autre PID. Ne pas imposer
  PID 1 à toute la CLI. Garder la protection PID 1 du nettoyage global.
- `container` est lui-même l'init du conteneur : un init placé devant lui
  (`docker run --init`, `init: true` de Compose, tini, docker-init) n'est pas
  une configuration prise en charge. Le superviseur n'y est plus PID 1 : pas de
  reaper, pas de `kill-all`, pas de nettoyage de l'état au démarrage. Ne pas
  ajouter de code pour ce cas ni le signaler comme une régression.
- `container processes start`, `stop`, `status`, `check` doivent rester utilisables par
  n'importe quel utilisateur dans le conteneur, y compris un UID numérique
  arbitraire, différent de celui du superviseur ou de celui qui a créé l'état.
  Ne pas exiger root, un compte nommé, un groupe dédié, `sudo`, ou un propriétaire
  identique pour ces commandes.
- Les commandes `container services require/download/install/link/unlink`
  préparent normalement l'image dans le Dockerfile et restent soumises aux
  permissions Unix des scripts, marqueurs et répertoires dans `/container/services`.
  Au runtime, elles s'exécutent avec un utilisateur disposant des droits
  nécessaires ; un refus de permission avec un autre UID est un comportement
  attendu. Ne pas rendre ces fichiers ou répertoires inscriptibles par tous
  pour contourner ce refus.
- `container services status` ne fait que lire `/container/services` : elle
  doit rester utilisable par n'importe quel utilisateur qui peut lire ce
  répertoire, sans droit d'écriture.
- Les scripts de service et les opérations système conservent leurs besoins
  réels en droits :
  ne pas ajouter d'élévation de privilèges pour installer des paquets ou écrire
  une configuration système.
- Préserver le fonctionnement avec une racine de conteneur en lecture seule et
  une zone d'exécution accessible en écriture. Les opérations de préparation de
  l'image appartiennent normalement au build.
- Prendre en charge les volumes, bind mounts, `emptyDir` et permissions fournies
  par Kubernetes, notamment `fsGroup`. Ne pas supposer que les répertoires montés
  appartiennent à l'UID courant ou que celui-ci peut les `chmod` / `chown`.

Ces points sont des exigences à préserver ou à atteindre lors des évolutions ;
leur présence ici ne certifie pas que tous les scénarios sont déjà couverts.

## Modèle de confiance et sécurité des écritures

L'état d'exécution partagé dans `/run/container/processes` n'est pas une
frontière de sécurité entre les utilisateurs du conteneur. Un utilisateur peut
modifier les marqueurs de contrôle, fausser un statut, arrêter un processus ou
perturber cet état. **Cette possibilité est acceptée.** Ne pas la « corriger »
en réservant le contrôle à root ou au propriétaire des fichiers. Les scripts et
marqueurs dans `/container/services` conservent leurs permissions Unix normales.

En revanche, une entrée ou un fichier d'état manipulé ne doit pas permettre au
superviseur d'écrire, tronquer, supprimer ou changer les permissions d'une cible
arbitraire hors du chemin prévu, en particulier lorsqu'il tourne en root.

- Valider les noms de services, processus et tags comme des composants de chemin
  uniques avec `helpers.ValidateName` / `ValidateNames`. Rejeter les noms vides,
  `.` / `..`, séparateurs et octets NUL ; `filepath.Join` seul ne suffit pas.
- Pour l'état interne partagé, réutiliser les protections de `helpers/files.go`
  et `helpers/os.go` : parcours par descripteurs de répertoires, `openat`,
  `O_NOFOLLOW`, créations exclusives, validation du type et opérations relatives
  au descripteur. Une vérification suivie d'une ouverture par chemin peut subir
  une course entre les deux opérations.
- `WriteFileAtomic` écrit dans un fichier temporaire privé puis publie par
  `renameat`. Ne pas remplacer cela par une ouverture avec troncature du fichier
  existant : un lien symbolique ou physique pourrait faire modifier une autre
  cible. Refuser les liens dans les répertoires intermédiaires.
- `ReadFileSafe` refuse les liens symboliques et les fichiers non réguliers ;
  `O_NONBLOCK` évite de rester bloqué à l'ouverture d'une FIFO. `CreateMarker`
  conserve un fichier régulier existant et `Remove` utilise `unlinkat`.
- Ne pas ajouter de restrictions d'UID propriétaire sur les fichiers d'état
  d'exécution des processus. Les répertoires de contrôle doivent permettre la
  création et la suppression entre utilisateurs.
- Ne pas appliquer ces permissions partagées récursivement à `/container`, aux
  scripts ou au système de fichiers entier. La configuration des répertoires
  d'exécution passe par `ConfigureRuntimeDirectory`.
- `ConfigureRuntimeDirectory` ne règle les droits que des répertoires qu'il
  crée : `0777`, en gardant les bits hérités comme le setgid de `fsGroup`. Un
  répertoire existant (couche d'image, volume, `tmpfs`, `emptyDir`) est utilisé
  tel quel, sans `chmod` : ses droits relèvent de celui qui l'a créé ou monté.
  Ne pas réintroduire de `chmod` des répertoires existants.
- Distinguer l'état interne des destinations explicitement données par
  l'utilisateur à `container envsubst file` / `templates`. Ces dernières passent
  intentionnellement par `WriteFileInPlace`, qui conserve l'inode et permet les
  liens et fichiers bind-montés. Ne pas généraliser ce comportement aux PID,
  marqueurs ou sorties du générateur.
- Préserver la validation des identifiants et noms Unix avant interpolation
  dans les commandes de création d'utilisateurs et de groupes. Préférer les
  arguments séparés à une commande shell construite par concaténation ; conserver
  `EscapeShell` pour les exports d'environnement destinés au shell.
- Ne pas journaliser de valeurs de variables d'environnement au niveau info ou
  supérieur : elles peuvent contenir des secrets. Les dumps complets existants
  restent au niveau debug ; ne pas en ajouter d'autres.
- Les arguments des commandes lancées ne s'affichent qu'au niveau debug : au
  niveau info et dans les erreurs, `helpers.Exec` ne montre que le chemin de
  l'exécutable, car une commande utilisateur peut contenir un mot de passe.
  `WithLogArgs` est réservé aux commandes construites par le binaire dont les
  arguments ne sont jamais des secrets (paquets, utilisateurs, groupes).

## Carte du dépôt

| Emplacement | Responsabilité |
| --- | --- |
| `main.go`, `config/config.go` | Initialisation du logger et configuration du runtime, chemins et noms des fichiers. |
| `cmd/` | CLI Cobra : point d'entrée, flags et sous-commandes ; délègue au cœur et aux helpers. |
| `core/core.go` | Interfaces et initialisation paresseuse des composants avec `sync.Once`. |
| `core/entrypoint.go` | Préparation du système de fichiers et des services, environnement, lancement du cycle de vie. |
| `core/lifecycle.go` | Étapes startup/process/finish, concurrence, redémarrages, signaux et arrêt. |
| `core/processes.go` | État des processus, demandes start/stop, statuts, tags et surveillance. |
| `core/services.go` | Découverte et filtrage des services, priorités, scripts et marqueurs. |
| `core/filesystem.go`, `core/environment.go` | Chemins, fichiers `.env` et restitution des variables. |
| `core/distribution.go`, `alpine/`, `debian/` | Détection via `/etc/os-release`, paquets, utilisateurs, groupes et installation des assets. Debian couvre aussi Ubuntu. |
| `core/generator.go`, `core/assets/generator/templates/` | Génération des Dockerfiles, services et fichiers d'environnement. |
| `helpers/` | Exécution, groupes de processus, reaping, accès fichiers, surveillance, substitution et validations. |
| `errors/`, `log/` | Erreurs communes, codes de sortie et journalisation console/JSON. |
| `Dockerfile` | Compilation statique Linux puis cibles `base` et `nonroot` ; binaire `/usr/sbin/container`. |
| `build/` | Module Go indépendant : matrice et tests de baseimage ; utilise la bibliothèque externe `github.com/osixia/dagger-build`. |
| `.github/workflows/go.yml` | Lint des deux modules, ShellCheck, govulncheck, pipeline Dagger et publication sur tags. |
| `docs/` | Exemples mono/multi-processus et tutoriels Nginx / PHP-FPM. |

Les deux modules, le Dockerfile (`golang:1.26.10`) et la CI utilisent Go `1.26.10`.
Le module `build/` utilise le SDK Dagger `0.21.10`.
`build/main.go` conserve la matrice d'images et les personnalisations du projet.
Les packages partagés viennent de `github.com/osixia/dagger-build`, figé sur
un commit publié dans `build/go.mod`. Go récupère cette dépendance normalement,
sans dépôt voisin requis ni étape spécifique dans la CI.

## Contrats de fonctionnement à préserver

### Système de fichiers et services

- `/container/environment` contient les fichiers dont le nom commence par
  `.env`, recherchés récursivement. `/container/services/<nom>` contient les
  scripts et marqueurs de chaque service.
- `/container/generator/templates` contient les modèles installés ;
  `/run/container/generator` reçoit les fichiers générés.
- Les scripts d'un service sont facultatifs : `download.sh`, `install.sh`,
  `startup.sh`, `process.sh`, `finish.sh`. Les marqueurs sont `.optional`,
  `.downloaded`, `.installed`, `.entrypoint`, avec `.priority`, `.stop-signal`
  (signal d'arrêt du `process.sh`, SIGTERM par défaut) et `.tags/<tag>`.
  « Link » désigne le marqueur `.entrypoint`, pas un lien symbolique.
- Sans sélection explicite, l'entrypoint exécute les services liés. Les commandes
  de services ont leurs propres filtres par défaut : lire `cmd/services/` avant
  de changer le traitement des services facultatifs ou déjà préparés.
- La priorité par défaut est 500 ; les nombres croissants puis les noms
  déterminent l'ordre. Les hooks startup et finish sont séquentiels ; les
  `process.sh` sont lancés concurremment, sans garantie d'ordre de démarrage.
- L'entrypoint prépare les téléchargements et installations manquants avant de
  charger les fichiers d'environnement, puis exécute startup → process → finish,
  avec les pré-commandes configurées et les commandes de sortie.
- Les variables déjà présentes dans l'environnement, même vides, restent
  inchangées. Les fichiers sont réunis dans l'ordre du parcours lexical et
  analysés avec `gotenv.StrictParse` ; préserver ses règles d'interpolation.
  `container run` charge aussi les fichiers sauf option contraire.
- Les modèles passent par envsubst : préserver les `$${VAR}` destinés à devenir
  des `${VAR}` dans les scripts générés. Vérifier les directives `go:embed` lors
  d'ajouts d'assets, notamment pour les fichiers et répertoires cachés.

### Contrôle des processus

- L'état se trouve dans `/run/container/processes` : `<nom>.pid`, `<nom>.down`,
  `<nom>.starting` et `<nom>/.tags/<tag>`. Ne pas isoler cet état dans un
  répertoire par UID.
- `stop` crée `.down`, `start` le retire ; le superviseur observe l'état et
  arrête ou relance son propre enfant. La CLI attend le résultat avec un délai
  de 30 secondes par défaut ; `--timeout 0` signifie une attente sans limite.
  Ne pas remplacer ce mécanisme par l'envoi direct de signaux depuis la CLI,
  qui échouerait entre UID différents.
- L'état d'un processus ne se lit que dans les fichiers partagés : up = fichier
  `.pid` valide (entier strictement positif) sans `.down`, down = `.down` sans
  `.pid` (`Process.IsUp` / `IsDown`, utilisés par `start`, `stop`, `status` et
  `check`). Pendant le délai avant un redémarrage, `.starting` garde le
  processus listé sans `.pid`. Plusieurs conteneurs peuvent partager
  `/run/container`, chacun limité à une partie des services avec `--exec`, et
  piloter les processus des autres : un PID n'a de sens que dans l'espace de PID
  du superviseur qui l'a écrit. Ne jamais tester un PID lu dans un fichier avec
  `kill(pid, 0)`, sauf pour les fichiers de ses propres services au démarrage.
- Le watcher est installé avant le lancement des services. Les événements
  déclenchent une relecture de l'état courant ; une erreur du watcher provoque
  une réconciliation. Préserver les protections contre les événements périmés
  et les courses avec le remplacement d'un processus.
- Un arrêt volontaire doit permettre un futur `start`, même avec les
  redémarrages automatiques désactivés. Au démarrage du conteneur, les demandes
  d'arrêt précédentes sont supprimées et les fichiers PID périmés sont nettoyés,
  uniquement pour les services lancés par ce conteneur : ceux des autres
  conteneurs qui partagent le répertoire restent intacts.
- Le redémarrage automatique est désactivé par défaut pour un seul service et
  activé pour plusieurs services ayant un `process.sh`, sauf choix explicite.
  Le délai exponentiel va de 100 ms à 5 s et revient à 100 ms après 30 s
  d'exécution. `--no-restart-delay` enlève ce délai sans activer le redémarrage
  ni modifier l'attente d'un service volontairement arrêté.
- Les sélections par noms et `tag:<nom>`, les alias et les différences entre une
  liste absente et une liste explicitement vide font partie du comportement.

### Signaux, attente et gVisor

- `helpers.Exec` gère le lancement, l'attente et l'escalade des signaux. Les
  services utilisent des groupes de processus pour arrêter aussi les descendants.
  L'arrêt (SIGTERM puis SIGKILL) vise tout le groupe ; les autres signaux
  relayés (SIGHUP, SIGUSR1…) ne visent que le processus principal, qui gère ses
  enfants comme le maître nginx ou php-fpm. Ne pas les renvoyer au groupe : le
  fils VRRP de keepalived passe en FAULT s'il reçoit SIGHUP directement.
  Conserver le comportement du terminal pour Bash, Ctrl+C et les commandes
  interactives. La fin de Bash ou de la commande personnalisée termine l'étape
  process et arrête les services associés.
- SIGINT, SIGTERM et SIGQUIT pilotent l'arrêt du cycle de vie. SIGHUP reste pris
  en charge pendant l'exécution et est transmis pendant l'étape process.
  Préserver l'exécution des hooks finish lors de l'arrêt normal des processus.
- Conserver les erreurs de sortie réelles même lors d'une annulation concurrente,
  les délais d'arrêt bornés et l'escalade SIGKILL. `errors.ExitCode` traduit une
  terminaison par signal en `128 + signal`.
- Le reaper s'active uniquement en PID 1. Il découvre ses enfants directs via
  `/proc/<pid>/stat` et attend uniquement les enfants non gérés. Le registre
  partagé avec `helpers.Exec` empêche de voler le statut attendu par `exec.Cmd`.
- Ne pas dépendre de `/proc/<pid>/task/<tid>/children` : cette interface est
  précisément évitée pour la compatibilité gVisor. Ne pas introduire de
  `wait(-1)` qui pourrait consommer la sortie d'un enfant supervisé.
- Les opérations actuelles utilisent notamment procfs, les signaux, les groupes
  de processus, `wait4`, fsnotify et les appels fichiers relatifs aux descripteurs.
  Toute nouvelle dépendance à une interface noyau doit être vérifiée sous gVisor ;
  une réussite sous Docker classique ne suffit pas à démontrer cette compatibilité.

## Style et maintenance

- Suivre l'organisation Go existante : imports standard, dépendances tierces,
  imports du projet ; formatage avec `gofmt`, sections commentées et conventions
  locales de noms. Garder le code, les commentaires et l'aide CLI en anglais.
- Conserver les interfaces du cœur, les options fonctionnelles de filtrage, les
  contextes et les points de synchronisation existants. Les wrappers Cobra
  doivent rester simples et déléguer la logique métier.
- Utiliser le package `log`, les traces d'entrée quand elles suivent le style
  voisin, les erreurs contextualisées avec `%w` et les helpers `Must` aux
  frontières appropriées. Ne pas terminer le programme depuis une nouvelle
  fonction métier simplement pour éviter de remonter une erreur.
- Séparer les sections d'un fichier par une bannière :
  ```go
  // Processes config
  // =============================
  ```
- Un concept du cœur = une interface publique, une struct privée et un
  constructeur `newXxx(...) (Xxx, error)`. Les configurations exposent
  `Validate() (bool, error)` et signalent un champ manquant avec
  `fmt.Errorf("Field: %w", errors.ErrRequired)`. Les noms de fichiers, chemins
  et clés d'environnement restent centralisés dans `config/config.go`.
- Les filtres de liste sont des options fonctionnelles (`WithProcessesNames`,
  `WithoutServicesTags`, `HandleServicesTagPrefixInNames`) ; un filtre à trois
  états utilise un `*bool`.
- Dans `core/` et `cmd/`, les fonctions commencent par une ligne vide puis une
  trace : `log.Trace("type.Method called")` ou
  `log.Tracef("type.Method called with x: %v", x)`. Les helpers bas niveau de
  `helpers/files.go` n'en ont pas : suivre le fichier voisin.
- Journaux : `log.Infof("Starting %v ...", name)` (suffixe ` ...`) pour les
  étapes visibles, `Debugf` pour le détail, `Warning` pour une erreur
  récupérable. `fmt.Print*` est réservé à la sortie d'une commande.
- Les erreurs du projet viennent de `github.com/osixia/container-baseimage/errors` ;
  importer la bibliothèque standard sous l'alias `goerrors` quand les deux sont
  nécessaires.
- Commandes Cobra : une variable `cobra.Command` par fichier, `Run` avec
  `helpers.Must` / `MustVal` (sauf besoin réel de `RunE`, comme
  `cmd/log/level.go`), et dans `init()` : `Flags().SortFlags = false` puis
  `cmdlog.AddFlags(cmd.Flags())`. Les `Short` commencent par une majuscule, les
  aides de flags par une minuscule ; un `\n` final dans l'aide d'un flag sépare
  les groupes dans `--help`. Les arguments de tag utilisent le préfixe `tag:`
  et le `Long` précise « With no argument: ... » quand c'est pertinent.
- Le binaire est compilé avec `CGO_ENABLED=0` et doit rester statique : aucune
  hypothèse sur la libc (glibc pour Debian/Ubuntu, musl pour Alpine).
- Les scripts Bash suivent `set -euo pipefail`, ceux en `/bin/sh` `set -eu`.
  Respecter le shell déclaré, citer les expansions et utiliser `exec` pour le
  processus principal d'un service, en transmettant `"$@"` si nécessaire.
- Lors d'une évolution publique, maintenir ensemble aide CLI, README, modèles
  et exemples concernés. Ne pas reformater les autres fichiers au passage.
- Messages de commit (quand un commit est approuvé) : préfixe conventionnel
  (`fix:`, `feat:`, `docs:`), sujet en minuscules à l'impératif sans point
  final, ligne vide, puis des puces `- ` qui expliquent le changement de
  comportement et sa raison, en anglais, lignes d'environ 72 colonnes :
  ```
  fix: stop services when the custom command or bash exits

  - When the custom command or bash run alongside services exits, whatever
    its exit code, end the process step
  ```

## Validation adaptée au changement

Pour une modification documentaire seule, vérifier les chemins, commandes et
le diff ; il n'est pas nécessaire de reconstruire toutes les images.

Pour du code Go, exécuter les vérifications pertinentes dans chaque module
touché. `go test ./...` à la racine **ne traverse pas** le module `build/`.

```bash
# Depuis la racine
go test ./...
go vet ./...
go build -o /tmp/container-baseimage-check .

# Depuis build/ si concerné
go test ./...
go vet ./...
go build -o /tmp/container-baseimage-build-check .
```

La CI utilise aussi `golangci-lint run` dans chacun des modules, ShellCheck et
govulncheck. Exécuter les outils nécessaires selon les fichiers modifiés et leur
disponibilité. Ne pas ajouter de tests qui recopient seulement l'implémentation ;
privilégier une régression observable et les cas de concurrence ou de sécurité.

Les tests Go de `helpers/` et `core/` couvrent la sécurité des écritures
(liens symboliques et physiques, FIFO, répertoires d'exécution), `helpers.Exec`
(codes de sortie, groupes, signaux relayés, escalade SIGKILL, fichier PID), le
reaper, l'état et la sélection des processus, les services et les fichiers
`.env`, ainsi que la réconciliation du watcher et le filtrage des signaux.
La CI les lance avec `go test -race ./...`, avec l'UID courant, hors PID 1.
La bibliothèque externe `dagger-build/job` fournit le contrôle générique de
version ; `build/main.go` personnalise `job.Test` pour exécuter les scénarios
propres à baseimage dans `build/tests/`, après `job.DefaultTest`.
Les contrôles Dagger de `build/tests/runtime.go` et `runtime_control.go` vérifient
les hooks, codes de sortie, signaux, redémarrages, sorties de Bash et commandes
personnalisées, puis un superviseur PID 1 contrôlé depuis des conteneurs d'UID
différents partageant son état. `build/tests/examples.go` teste les vrais scripts
et fichiers d'environnement des exemples, ainsi que les réponses HTTP des
tutoriels Nginx/PHP-FPM sur les images root Debian/Ubuntu. Ces contrôles ne
reconstruisent pas les Dockerfiles des exemples et ne constituent pas une
couverture complète de PID 1, des UID croisés ou de gVisor.

Les tests d'images locaux passent par le job `test`, sans publication :

```bash
# Depuis build/ ; nécessite un moteur de conteneurs accessible pour Dagger
go run main.go job test ../Dockerfile --target base
go run main.go job test ../Dockerfile --target nonroot --tags-suffix=-nonroot
```

Pour une évolution de supervision ou des accès fichiers, compléter selon le
risque par des scénarios en conteneur : PID 1, root/nonroot et UID arbitraire ;
`stop` par un UID puis `start` par un autre ; statut d'un processus d'un autre
UID ; filtres par tags ; signaux et descendants ; racine en lecture seule avec
runtime inscriptible ; volumes avec propriétaire différent ; liens symboliques
intermédiaires et finaux, liens physiques et fichiers non réguliers. Vérifier
que les cibles extérieures restent intactes et que la protection n'empêche pas
le contrôle partagé normal.

Quand gVisor est concerné, exécuter aussi les scénarios pertinents avec `runsc`
si disponible et signaler explicitement son absence sinon. Effectuer les tests
d'intégration du superviseur dans des conteneurs jetables : ne pas lancer
`container install` ni les scripts d'installation des distributions sur l'hôte.
