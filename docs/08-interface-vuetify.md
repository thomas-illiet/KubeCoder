# Interface Vue/Vuetify

## Direction visuelle

- Vue 3 + Vuetify 3, thème sombre par défaut.
- L'ensemble des libellés, messages, exemples et états visibles de l'interface est rédigé en anglais ; les identifiants techniques et noms de protocoles restent inchangés.
- Palette retenue pour le mock : shell presque noir, trois niveaux de surfaces graphite nettement différenciés et bordures franches. L'indigo est l'unique accent structurel et les couleurs cyan, ambre, verte ou rouge restent réservées aux informations sémantiques.
- Les cartes de filtres, en-têtes de résultats, champs, lignes alternées et drawers se distinguent d'abord par leur luminosité et leur bordure, sans gradients décoratifs, lueurs ou alternance de couleurs entre les lignes.
- États exprimés par couleur **et** texte/icône afin de rester accessibles.
- Densité confortable pour le chat, compacte pour les tables d'administration.
- Contraste WCAG AA, focus clavier visible et réduction des animations respectant les préférences système.

## Structure de navigation

- Barre supérieure : organisation courante toujours visible ; le changement ouvre une recherche côté serveur paginée avec organisations récentes, adaptée à plusieurs milliers d'organisations, puis recherche globale, état du système et profil.
- Navigation latérale organisation regroupée par catégories : `General` pour la vue d'ensemble, `Workspace` pour Repositories et Sessions, `Configuration` pour Skills, Serveurs MCP et Secrets, puis `Organization` pour les paramètres. Le choix d'un agent publié s'effectue dans le contexte d'un repository ou lors de la création d'une session, sans page catalogue dédiée dans l'espace organisation.
- Administration séparée et regroupée par catégories : `General`, `Agent platform` pour les agents, skills et serveurs MCP, `Security & access` pour les secrets et membres, puis `Infrastructure` pour les images/adaptateurs. Les titres de catégories sont masqués lorsque la navigation est réduite et remplacés par des séparateurs discrets.
- Breadcrumb portant systématiquement organisation puis repository/session.

## Écrans MVP

### Authentification et profil

Le mock possède des pages dédiées `Login` et `Logout` hors des layouts métier. Le menu utilisateur ouvre une page `Profile` basique depuis les espaces organisation et administration. Elle présente l'identité OIDC, les informations personnelles et l'action de déconnexion, sans simuler de persistance réelle ni stocker de credential.

### Accueil organisation

Runs actifs, sessions récentes, erreurs nécessitant une action et raccourci “Nouvelle session”.

### Repositories

Liste filtrable, fournisseur, branche par défaut, agent sélectionné, état d'accès, dernière vérification et actions. La création et l'édition utilisent une modal opaque et un stepper en quatre étapes : choix du provider, identité du repository adaptée au provider, paramètres communs du repository, puis récapitulatif. Le type de déploiement et l'URL d'instance sont définis par le backend à partir de la configuration de l'organisation et ne sont ni affichés ni modifiables dans cette modal. Le provider est verrouillé après création.

L'utilisateur choisit un agent dans le catalogue publié et ne voit que les options repository déclarées configurables. Le formulaire d'organisation ne contient aucun champ de credential : l'accès au provider est résolu en dehors de cette interface et aucune référence sensible n'y est affichée. Avant la persistance, le workflow cible teste la connectivité et détecte la branche par défaut ; le mock ne simule que la progression visuelle.

### Administration des agents

Écran réservé aux administrateurs : runtime/image, modèle/provider, endpoint OpenAI ou compatible, clé API LLM en écriture seule, prompt système, outils, ressources, règles réseau, skills, bindings de credentials et schéma des options utilisateur. Une version doit être validée puis publiée avant d'apparaître dans le sélecteur utilisateur. L'interface affiche un aperçu exact de la projection publique sans endpoint privé ni clé API.

Dans le catalogue utilisateur des agents, les cartes utilisent exclusivement les surfaces graphite du thème. Les icônes, badges de capacités et actions restent neutres ; seules les informations d'état emploient une couleur sémantique.

### Skills

Vues distinctes global/organisation/repository. Chaque ligne affiche scope, version, état local, état effectif et source de l'héritage. Contrôle tri-state clair : hériter, activer, désactiver.

### Serveurs MCP

Le catalogue d'administration sépare la carte de filtres de la table. Il permet de basculer entre définitions globales et organisation, et affiche transport, cible expurgée, date de modification et une action **Modifier**. Il n'affiche ni état d'activation, ni nombre d'outils : les outils ne sont connus qu'après la découverte MCP dynamique. La création utilise une modal opaque avec deux transports seulement : `STDIO` affiche commande, arguments et références de variables ; `SSE` affiche endpoint, headers référencés et credential binding.

L'action principale **Tester dans un conteneur** affiche les étapes asynchrones du conteneur Docker dédié, le handshake, le résultat nettoyé et le nettoyage. L'espace organisation possède un bouton **Ajouter un serveur** et réserve l'action **Modifier** aux définitions de l'organisation ; les définitions globales affichent « Géré globalement » et ne sont pas éditables dans ce layout. Leur activation locale reste possible. Tous les serveurs sont désactivés par défaut et la table n'affiche pas de colonne Capacités. Le formulaire d'agent ne présélectionne aucun MCP et ne propose que les serveurs activés dans l'organisation.

### Secrets

La gestion des secrets existe dans le layout organisation pour `OWNER` et `ADMIN`. Elle couvre les portées organisation/repository, la création, les détails expurgés, les bindings et la rotation. Métadonnées seulement, scope visible, date de rotation, expiration optionnelle, état calculé et usages. La valeur n'est saisie qu'à la création/rotation et n'est jamais réaffichée. Les états `actif`, `expire bientôt`, `expiré` et `révoqué` sont visibles et filtrables. L'administration globale ne reçoit pas implicitement l'accès aux valeurs tenant.

Dans l'administration, la page ne gère que les secrets applicatifs et leurs bindings. L'accès Git repose sur une clé publique unique configurée globalement dans le backend : aucun credential Git n'est listé, créé, modifié ou sélectionné depuis cette page.

### Création de session

Assistant court : repository, profil, branche/révision, skills effectifs, ressources puis résumé. Le résumé pré-run signale collisions, secrets manquants ou expirés et configuration effective.

### Chat de session

- En-tête : repository/branche, profil, état, durée, actions arrêter/reprendre/archiver.
- Fil central : messages et cartes d'événements condensables.
- Composer fixé en bas, visible quelle que soit la hauteur du fil.
- Deux onglets structurent la session : `Chat` pour la conversation OpenCode et les événements d'outils, puis `Code changes` pour la liste des fichiers modifiés et leur diff unifié.
- Panneau latéral : fichiers/artefacts, skills effectifs, snapshot et détails du run.
- Toute valeur filtrée est rendue par un composant uniforme `[SECRET_REDACTED]` ; un événement mis en quarantaine est représenté sans contenu sensible.
- Reconnexion automatique avec indicateur explicite ; aucun message n'est envoyé deux fois.

## États à concevoir explicitement

- chargement initial et reconnexion ;
- toute data table sans résultat affiche par défaut l'état réutilisable « Aucun élément trouvé », une aide pour modifier la recherche ou les filtres et, lorsqu'elle existe, la prochaine action pertinente ;
- run en file, démarrage lent, annulation et timeout ;
- secret/credential manquant ;
- session reprise nativement ou reconstruite ;
- permission insuffisante ;
- moteur indisponible sans perte du message.

## Responsive

- Desktop : navigation + chat + panneau contextuel.
- Tablette : panneau contextuel en drawer.
- Mobile : navigation temporaire, chat plein écran, actions critiques regroupées sans masquer le composer.

