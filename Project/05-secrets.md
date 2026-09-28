# Gestion des secrets

## Objectifs

- Créer des secrets au scope organisation ou repository.
- Injecter une version précise sous forme de variable d'environnement dans un run autorisé.
- Ne jamais retourner une valeur après sa soumission.
- Permettre rotation, révocation, audit et suppression différée sûre.

## Modèle proposé

`SecretDefinition` contient : nom, description, scope, propriétaire, nom de variable cible, dates, expiration optionnelle et version active. `SecretVersion` contient une référence opaque vers le fournisseur de secrets, une date d'expiration optionnelle et jamais la valeur en clair dans les tables métier.

L'état exposé est calculé à partir de la version active : `ACTIVE`, `EXPIRING_SOON`, `EXPIRED` ou `REVOKED`. La durée définissant `EXPIRING_SOON` est configurable par l'exploitation. Un secret expiré reste visible dans l'historique et l'audit, mais ne peut plus être résolu pour un nouveau run.

Les noms sont validés comme variables POSIX (`[A-Z_][A-Z0-9_]*` par défaut). Les noms réservés par le runtime sont refusés.

## Résolution

La résolution n'injecte pas automatiquement tous les secrets visibles. Une `AgentDefinition` ou un repository déclare des **bindings** explicites :

`nom logique -> secret scoped -> version/active -> variable d'environnement`

Ordre proposé en cas de même variable :

1. binding explicite repository ;
2. binding explicite organisation ;
3. absence d'injection.

Une collision non explicitement résolue doit bloquer le run, pas choisir silencieusement une valeur. Les secrets globaux éventuels appartiennent exclusivement à la plateforme. Ils ne sont jamais listés, recherchables, adressables ni sélectionnables depuis une API ou une vue d'organisation. Le layout organisation ne manipule que les scopes `organisation` et `repository`.

## Stockage et injection

### Cible recommandée

Utiliser un fournisseur externe (Vault ou service KMS/secret manager disponible dans l'environnement). L'application conserve seulement les références et métadonnées.

### Option de démarrage

Chiffrement applicatif par enveloppe : clé de données par version, chiffrée par une clé maître hors base. Cette option exige rotation, sauvegarde et procédure de perte de clé documentées.

### Dans Kubernetes

- Le reconciler récupère les valeurs seulement après autorisation du run.
- Il crée un Secret Kubernetes temporaire nommé par identifiant non sensible, avec owner/reference ou tâche de nettoyage durable.
- Le Pod reçoit uniquement les clés bindées. La valeur devient une variable d'environnement du processus agent, conformément au besoin ; elle est donc lisible par ce processus.
- Le Secret et le jeton de run sont supprimés/révoqués immédiatement après l'état terminal.
- Les valeurs connues alimentent un filtre de redaction avant persistance et diffusion, sans considérer ce filtre comme une garantie unique.

## API et interface

- Écriture seule pour la valeur : réponse `201` avec métadonnées et fingerprint non réversible.
- Rotation créant une nouvelle version avec une expiration optionnelle ; les runs déjà démarrés gardent leur version.
- La liste affiche la date d'expiration et l'état calculé. Elle permet de filtrer les secrets expirés ou proches de l'expiration.
- La création d'un nouveau run est bloquée avec une erreur actionnable lorsqu'un binding résout une version expirée.
- Test de connectivité séparé pour les credentials Git/modèle, avec résultat sans écho de secret.
- Confirmation forte pour révocation/suppression, avec affichage des profils affectés.
- Audit : créateur, rotateur, révocateur, run ayant résolu la version, jamais la valeur.

## Masquage dynamique des conversations

Le masquage est un contrôle serveur, pas un effet CSS. La version brute d'un événement ne doit jamais être envoyée au navigateur ni écrite dans PostgreSQL ou le stockage objet par défaut.

Pour chaque run, le plan de contrôle construit un jeu de redaction éphémère à partir des valeurs réellement injectées et de variantes bornées : valeur exacte, préfixes d'authentification connus, encodage URL et Base64 lorsque pertinent. Le jeu est transmis au bridge par un canal temporaire et n'apparaît pas dans le snapshot, les logs ou l'audit.

Pipeline proposé :

1. privilégier des outils structurés qui marquent les champs sensibles et empêcher l'agent de lire les secrets qui ne lui sont pas nécessaires ;
2. filtrer dans le bridge avant la sortie du Pod ;
3. filtrer à nouveau dans l'Event ingestor avant persistance et SSE ;
4. remplacer chaque correspondance par `[SECRET_REDACTED]` et produire un audit ne contenant que l'identifiant du secret et le run ;
5. mettre en quarantaine l'événement plutôt que le diffuser si le filtre ne peut pas déterminer une sortie sûre.

Le filtre streaming conserve une fenêtre suffisante pour détecter une valeur répartie sur plusieurs chunks ; aucun chunk n'est diffusé avant d'être déclaré sûr. Une bibliothèque de recherche multi-motifs éprouvée est préférée à une suite de remplacements naïfs.

Les valeurs inconnues du coffre (par exemple un secret déjà présent dans le repository) nécessitent en complément des détecteurs de formats : clés privées, tokens usuels, chaînes d'entropie élevée et règles administrables. Cette détection reste probabiliste ; le produit doit afficher cette limite et proposer quarantaine/alerte plutôt que promettre un masquage absolu de toute donnée transformée, chiffrée ou reformulée.

## Tests de sécurité obligatoires

- Absence dans API, traces, erreurs, événements, manifests exportés et captures d'audit.
- Refus d'un secret d'une autre organisation, y compris via identifiant connu.
- Nettoyage après succès, échec, annulation et crash du reconciler.
- Rotation pendant un run sans mutation du snapshot courant.
- Expiration avant un nouveau run, passage à `EXPIRING_SOON` et refus de résolution après expiration.
- Redaction des variantes usuelles tout en documentant ses limites.
- Valeur répartie entre plusieurs événements/chunks, encodée en URL/Base64 ou précédée de `Bearer`.
- Garantie qu'aucune version brute n'est observable dans la réponse SSE, la base, le stockage objet ou les logs.
- Événement inconnu suspect mis en quarantaine et non seulement masqué dans le DOM.

