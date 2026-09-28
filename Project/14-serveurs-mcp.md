# Serveurs MCP

## Objectif et périmètre

KubeCoder fournit un catalogue de serveurs Model Context Protocol (MCP) utilisables par les agents. Le MVP accepte exactement deux transports :

- `STDIO` : KubeCoder démarre un processus approuvé et échange des messages MCP sur son entrée/sortie standard ;
- `SSE` : KubeCoder se connecte à un endpoint MCP distant autorisé via HTTPS et consomme son flux Server-Sent Events.

Le protocole SSE utilisé par un serveur MCP est distinct du SSE de l'interface KubeCoder qui diffuse les événements d'un run vers le navigateur.

## Portées et visibilité

Une `McpServerDefinition` possède l'une des portées suivantes :

- **globale** : créée par un administrateur plateforme et publiable à plusieurs organisations ;
- **organisation** : propriété d'une organisation et invisible des autres organisations.

Le catalogue d'administration permet de créer, modifier, tester, versionner et auditer les définitions. Il n'affiche pas d'état d'activation : l'existence d'une définition globale ne l'active dans aucune organisation.

L'espace organisation permet d'ajouter et modifier uniquement ses propres définitions, puis d'activer explicitement chaque serveur. Une définition globale peut être activée ou désactivée localement, mais elle reste non modifiable depuis l'organisation et porte l'indication « Géré globalement ». Tous les serveurs sont désactivés par défaut. La liste contient le nom, la source, le transport et l'activation locale, mais jamais les valeurs de credentials.

Un serveur visible dans une organisation n'est pas automatiquement utilisable par tous les agents. Chaque version publiée d'`AgentDefinition` possède une liste explicite de bindings MCP. Le snapshot du run fige les définitions MCP, leurs versions et leurs politiques d'outils.

## Configuration STDIO

Une définition STDIO contient :

- un nom et une description ;
- un exécutable issu d'une allowlist administrée ;
- une liste ordonnée d'arguments validés ;
- un environnement composé de valeurs non sensibles et de références `secret://...` ;
- une limite de démarrage, de durée, de CPU et de mémoire ;
- une politique des outils autorisés parmi ceux annoncés par le serveur.

La commande n'est jamais passée à un shell. Le worker construit directement `argv`, refuse les chemins, options ou exécutables non autorisés et ne monte que le workspace nécessaire en lecture seule ou lecture-écriture selon la politique.

## Configuration SSE

Une définition SSE contient :

- un endpoint HTTPS ;
- des headers non sensibles et/ou références `secret://...` ;
- un credential binding optionnel ;
- des timeouts de connexion, heartbeat et appel d'outil ;
- une allowlist d'hôtes, ports et outils.

Le backend résout et vérifie l'adresse après chaque redirection. Les adresses loopback, link-local, metadata cloud, réseaux internes non autorisés et changements DNS suspects sont refusés afin de limiter SSRF et DNS rebinding.

## Test de connexion dans un conteneur Docker dédié

Le bouton **Tester dans un conteneur** ne lance rien dans le navigateur ni dans le processus API. Il crée une demande de test confiée à un worker séparé. Pour chaque demande, le worker :

1. crée un conteneur Docker éphémère depuis une image de test MCP épinglée par digest ;
2. applique un utilisateur non-root, un filesystem racine en lecture seule, aucune capability, aucune socket Docker et des limites CPU/mémoire/PID ;
3. crée uniquement les fichiers temporaires et références de secrets nécessaires ;
4. applique une politique réseau dédiée : aucune sortie pour STDIO sauf destination déclarée, et uniquement l'hôte autorisé pour SSE ;
5. démarre le serveur ou client, négocie le handshake MCP, exécute `tools/list` et valide le schéma sans appeler d'outil métier ; la liste et son nombre ne sont jamais supposés avant cette découverte ;
6. filtre la sortie, produit un résultat structuré et borné puis arrête et supprime le conteneur, le réseau, les volumes et les secrets temporaires ;
7. lance un nettoyeur durable si l'arrêt nominal échoue.

Le test possède un timeout court et un identifiant de corrélation. L'API expose les états `QUEUED`, `STARTING`, `HANDSHAKING`, `SUCCEEDED`, `FAILED`, `TIMED_OUT` et `CLEANING`. Les journaux affichables contiennent seulement les étapes, codes d'erreur nettoyés, durée, version de protocole et noms d'outils. Les arguments, headers, variables, valeurs de secrets et résultats d'outils ne sont jamais journalisés.

L'accès au moteur Docker appartient exclusivement au worker isolé. L'API web ne reçoit pas la socket Docker et ne construit pas de commande libre. En production Kubernetes, ce worker reste un composant séparé et contrôlé ; le choix technique d'accès à un runtime compatible ne réduit pas les garanties décrites ci-dessus.

## Résolution pour un run

Avant le provisioning, le backend :

1. récupère la version publiée de l'agent ;
2. vérifie chaque binding MCP et sa portée ;
3. refuse une définition non activée dans l'organisation, inaccessible ou dont un secret est expiré ;
4. résout la politique d'outils et les références de secrets ;
5. ajoute au snapshot les identifiants, versions, transports, digest de configuration et outils autorisés, jamais les valeurs ;
6. provisionne STDIO dans la frontière du run ou prépare le client SSE avec les seules sorties réseau autorisées.

Une modification ultérieure d'un serveur MCP n'affecte pas un run existant. Une nouvelle version d'agent ou de serveur est nécessaire pour les runs futurs.

## Audit et observabilité

Sont audités : création, modification, test, activation/désactivation, binding à un agent et refus de résolution. Les métriques couvrent le taux de succès des tests, leur durée, les échecs de nettoyage, les handshakes de run, les timeouts et les appels par nom d'outil. Aucun argument ni résultat d'outil n'est un label de métrique ou une donnée de log.

## Activation et résultat de test

Une définition n'a pas d'état d'activation au niveau du catalogue d'administration. L'activation est un binding local à l'organisation, créé à `disabled` et modifiable uniquement par un administrateur de cette organisation.

Le test de connexion conserve son propre résultat et sa date sans devenir l'état du serveur. Une modification de transport, commande, endpoint, argument, header, secret ou politique d'outils rend le résultat précédent obsolète. Un agent ne peut sélectionner qu'un serveur activé ; la sélection reste vide par défaut.
