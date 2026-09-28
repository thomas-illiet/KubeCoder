# Modèle de domaine

## Agrégats principaux

| Entité | Rôle | Relations clés |
|---|---|---|
| `User` | Identité applicative issue du couple OIDC `issuer + subject` | membre de plusieurs organisations |
| `Organization` | Frontière de propriété et d'autorisation | membres, repositories, secrets, skills, sessions |
| `Membership` | Rôle local d'un utilisateur dans une organisation | `OWNER`, `ADMIN`, `MEMBER` |
| `Repository` | Référence vers un dépôt Git et sa branche par défaut | appartient à une organisation |
| `CredentialBinding` | Référence un moyen d'authentification Git sans exposer sa valeur | scope organisation ou repository |
| `SecretDefinition` | Métadonnées, scope, expiration optionnelle et version d'un secret | valeur stockée dans un coffre/chiffrement séparé |
| `Skill` | Contenu versionné utilisable par l'agent | créé au scope global, organisation ou repository |
| `SkillPolicy` | Activation ou désactivation à un niveau donné | résout l'état effectif |
| `McpServerDefinition` | Définition versionnée d'un serveur MCP STDIO ou SSE | scope global ou organisation, secrets référencés |
| `AgentMcpBinding` | Autorisation explicite d'un serveur MCP et de ses outils pour une version d'agent | lie agent, serveur et politique d'outils |
| `McpConnectionTest` | Exécution bornée d'un test dans un conteneur Docker éphémère | états, résultat nettoyé, corrélation et nettoyage |
| `AgentDefinition` | Agent publié par un administrateur : moteur, modèle, prompt, credentials, ressources, outils et politiques | versionnée, jamais modifiable par un membre standard |
| `RepositoryAgentBinding` | Association choisie par l'utilisateur entre son repository et un agent autorisé | contient uniquement les options non sensibles permises par l'agent |
| `Session` | Conversation durable liée à une organisation et un repository | contient messages et runs |
| `Run` | Une tentative d'exécution Kubernetes | utilise un snapshot immuable |
| `ConfigurationSnapshot` | Résultat figé de la résolution du profil, des skills et références de secrets | empreinte SHA-256 |
| `Event` | Flux ordonné d'un run | message, sortie outil, état, erreur, artefact |
| `Artifact` | Fichier ou résultat durable produit par un run | métadonnées SQL, contenu en stockage objet |
| `AuditEvent` | Trace d'une action sensible | acteur, organisation, cible, résultat, corrélation |

## Distinctions structurantes

### Session et run

- Une **session** survit aux Pods et contient la continuité fonctionnelle du chat.
- Un **run** commence à la soumission d'un message et se termine par succès, échec, annulation ou expiration.
- Une session peut contenir plusieurs runs séquentiels. Le parallélisme dans une même session est refusé au MVP pour éviter les conflits de workspace.
- Reprendre une session crée un nouveau run ; cela ne ressuscite pas un conteneur terminé.

### Agent, association repository et snapshot

- Une **définition d'agent** est modifiable et versionnée uniquement depuis l'administration.
- Une **association repository-agent** sélectionne cette définition et renseigne seulement les options que l'administrateur a déclarées configurables par l'utilisateur.
- Un **snapshot** est immuable et contient les versions/digests effectivement résolus, y compris les bindings MCP et les outils autorisés.
- Les valeurs de secrets ne sont jamais copiées dans le snapshot ; seules leurs références et versions y figurent.

### Message et événement

- Un **message** est une entrée lisible de la conversation (`user`, `assistant`, `system`).
- Un **événement** est une donnée technique ordonnée pouvant produire ou enrichir un message : token, appel d'outil, progression, fichier, erreur.
- Les événements techniques déjà filtrés ont une politique de rétention distincte des messages consolidés.

## Identifiants et invariants

- Utiliser des UUID/ULID non devinables ; ne pas incorporer de nom d'organisation ou de secret dans un identifiant public.
- Toute ressource tenant-aware porte `organization_id`, même lorsqu'il est déductible par jointure.
- Toute lecture ou mutation tenant-aware vérifie l'appartenance côté serveur.
- Un repository appartient à exactement une organisation au MVP.
- Un run référence exactement un snapshot et une version d'image du moteur.
- Un agent ne peut utiliser qu'un serveur MCP lié à sa version publiée et compatible avec la portée de l'organisation du run.
- Un test MCP ne partage ni conteneur, ni filesystem, ni credential avec un autre test et doit produire un état de nettoyage terminal.
- Un membre standard ne peut ni lire ni remplacer le modèle, le prompt système, les credentials, les outils privilégiés ou les ressources techniques maximales d'une définition d'agent.
- Les événements d'un run ont une séquence monotone et une clé d'idempotence.
- Une seule exécution non terminale est autorisée par session au MVP.

## États proposés

`Session.status` : `ACTIVE`, `ARCHIVED`, `DELETED`.

`Run.status` :

`QUEUED -> RESOLVING -> PROVISIONING -> STARTING -> RUNNING -> {SUCCEEDED | FAILED | CANCELED | TIMED_OUT}`

États supplémentaires contrôlés : `CANCELING`, `ORPHANED`, `RECOVERY_REQUIRED`.

Le backend conserve la cause structurée d'un échec séparément du texte destiné à l'utilisateur.

