# Skills et configuration d'agent

## Scopes et résolution

Les skills existent dans un catalogue avec un scope de propriété :

- **global** : créé par un administrateur plateforme, disponible selon sa politique globale ;
- **organisation** : visible uniquement dans l'organisation propriétaire ;
- **repository** : spécifique à un repository.

Chaque niveau peut définir une politique `INHERIT`, `ENABLED` ou `DISABLED`. L'état effectif est calculé du plus général au plus spécifique : global, organisation, repository, puis sélection de session si cette personnalisation est autorisée.

Règles proposées :

- l'override le plus spécifique gagne ;
- une politique peut interdire qu'un niveau inférieur réactive un skill sensible ;
- l'interface affiche la valeur effective, sa source et la chaîne d'héritage ;
- les modifications affectent uniquement les futurs runs ;
- les versions de contenu sont immuables et identifiées par digest.

## Contenu d'un skill

- identifiant, nom, description et scope ;
- version et digest ;
- instructions Markdown ;
- fichiers/ressources autorisés ;
- contraintes de runtime ou outils requis ;
- politique d'activation et possibilité d'override ;
- auteur, provenance et statut de validation.

Les identifiants techniques et fichiers ne sont pas interprétés comme des secrets. Un scanner refuse ou signale un contenu susceptible d'en contenir avant publication.

## Définition d'agent administrée

Une définition d'agent est créée, testée, versionnée et publiée dans l'espace d'administration. Un utilisateur standard peut la sélectionner mais ne peut pas voir ni modifier ses paramètres sensibles :

- adaptateur et version du moteur ;
- image OCI épinglée par digest ;
- fournisseur, modèle, endpoint OpenAI ou compatible OpenAI et paramètres ;
- clé API LLM saisie en écriture seule ou référence vers un credential existant ; seule la référence est conservée dans la définition et le snapshot ;
- prompt système de base ;
- ressources CPU, mémoire, stockage et durée dans les bornes techniques autorisées ;
- outils/capacités réseau ;
- bindings explicites de serveurs MCP STDIO/SSE et allowlist d'outils ;
- bindings de secrets ;
- politique de skills ;
- comportement de reprise et rétention.

Une définition peut être globale ou limitée à une organisation. Seuls `PLATFORM_ADMIN`, `OWNER` et `ADMIN` selon leur scope peuvent la modifier ou la publier. Les clés API, credentials, prompt système, endpoint privé, modèle/provider sensibles, outils privilégiés, ressources maximales et règles réseau restent absents de la projection utilisateur.

## Association configurée par l'utilisateur

Sur un repository qu'il est autorisé à gérer, l'utilisateur choisit un agent publié et renseigne uniquement un schéma d'options sûr défini par l'administrateur, par exemple branche de travail, sous-répertoire, stratégie de reprise ou skills laissés configurables. Il ne peut pas saisir/remplacer un credential, ajouter une variable d'environnement libre, changer l'image, le modèle, le prompt système, les outils ou les destinations réseau.

Le backend valide ces options contre le schéma de la version publiée. Si un agent est retiré ou qu'une nouvelle version impose une migration, l'association devient `ACTION_REQUIRED` au lieu de basculer silencieusement.

## Snapshot effectif du run

Avant provisioning, le backend génère un document canonique comprenant :

- versions de la définition d'agent, de l'association repository et de l'adaptateur ;
- image OCI par digest ;
- modèle et paramètres non secrets ;
- liste ordonnée des skills avec version, digest, état et provenance ;
- liste ordonnée des serveurs MCP avec définition/version, transport, digest, provenance et outils autorisés ;
- outils et règles réseau ;
- ressources et timeouts ;
- références/version des secrets, sans valeurs ;
- repository, révision de départ et stratégie de workspace.

Le JSON canonique reçoit une empreinte SHA-256. Le digest est affichable dans le détail du run et utilisé pour l'audit/reproduction.

## Validation avant run

- aucun skill activé ne manque ou n'est révoqué ;
- aucune dépendance d'outil n'est interdite par l'organisation ;
- chaque serveur MCP est activé dans l'organisation, compatible avec la portée du run et explicitement lié à la version d'agent ; aucun binding MCP n'est ajouté par défaut ;
- les endpoints SSE, exécutables STDIO, arguments et destinations réseau respectent leurs allowlists ;
- chaque binding de secret se résout sans collision ;
- les ressources demandées restent dans les bornes techniques autorisées ;
- l'image et l'adaptateur sont autorisés ;
- le profil de reprise est compatible avec le moteur.

